package connection

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

const (
	sqlSelect = "SELECT "
	from      = "FROM"
	where     = "WHERE "
	and       = "AND "
	or        = "OR "
)

// Flags - borrowed from Gosharedconnection
/* const (
	_FLAG_NOT_NULL = 1 << iota
	_FLAG_PRI_KEY
	_FLAG_UNIQUE_KEY
	_FLAG_MULTIPLE_KEY
	_FLAG_BLOB
	_FLAG_UNSIGNED
	_FLAG_ZEROFILL
	_FLAG_BINARY
	_FLAG_ENUM
	_FLAG_AUTO_INCREMENT
	_FLAG_TIMESTAMP
	_FLAG_SET
	_FLAG_NO_DEFAULT_VALUE
)
*/

type (
	//Connection datastructure for storing sharedconnection connection details
	Connection struct {
		DatabaseURL string `json:"databaseURL"`
		Host        string `json:"host"`
		Port        string `json:"port"`
		User        string `json:"user"`
		Password    string `json:"password"`
		DbName      string `json:"databaseName"`
		TLSMode     string `json:"sslmode"`
		Cacert      []byte `json:"cacert"`
		Clientcert  []byte `json:"clientcert"`
		Clientkey   []byte `json:"clientkey"`
		db          *sql.DB
	}

	//Connector is a representation of connector.json metadata for the sharedconnection connection
	Connector struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Title       string `json:"title"`
		Version     string `json:"version"`
		Type        string `json:"type"`
		Ref         string `json:"ref"`
		Settings    []struct {
			Name  string      `json:"name"`
			Type  string      `json:"type"`
			Value interface{} `json:"value"`
		} `json:"settings"`
	}

	//record is the one row of the ResultSet retrieved from database after execution of SQL Query
	record map[string]interface{}

	//ResultSet is an aggregation of SQL Query data records fetched from the database
	ResultSet struct {
		Record []*record `json:"records"`
		//Record []map[string]interface{} `json:"records"`
	}

	//Input is a representation of acitivity's input parametres
	Input struct {
		Parameters map[string]interface{}   `json:"parameters,omitempty"`
		Values     []map[string]interface{} `json:"values,omitempty"`
	}

	//Query structure for SQL Queries
	Query struct {
		TableName string            `json:"tableName"`
		Cols      []string          `json:"columns"`
		Filters   map[string]string `json:"filters"`
	}

	//QueryActivity provides Activity metadata for Flogo
	QueryActivity struct {
		metadata *activity.Metadata
	}
)

var connectorCache map[string]*Connection
var connectorCacheMutex sync.Mutex

var preparedQueryCache map[string]*sql.Stmt
var preparedQueryCacheMutex sync.Mutex

func init() {
	connectorCache = make(map[string]*Connection, 100)
	preparedQueryCache = make(map[string]*sql.Stmt, 100)
	go keepalive()
}

// sortable type for values in parameters collections
type tableOrdinal struct {
	ordinal  int
	keyfield string
	value    interface{}
}
type sortableTableOrdinal []tableOrdinal

func (o sortableTableOrdinal) Len() int           { return len(o) }
func (o sortableTableOrdinal) Less(i, j int) bool { return o[i].ordinal < o[j].ordinal }
func (o sortableTableOrdinal) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

func newSortableTableOrdinal(inputData *Input, prepared string) (ordinalSlice sortableTableOrdinal) {
	ordinalSlice = make(sortableTableOrdinal, len(inputData.Parameters))
	i := 0
	for keyfield, value := range inputData.Parameters {
		offset := strings.Index(prepared, "?"+keyfield)
		os := tableOrdinal{offset, keyfield, value}
		ordinalSlice[i] = os
		i++
	}
	return
}

//QueryHelper is a simple query parser for extraction of values from an insert statement
// This was taken from the postgres connector verbatum.  Its a shame we can't share code...
// type QueryHelper struct {
// 	sqlString   string
// 	values      []string
// 	first       string
// 	last        string
// 	valuesToken string
// }

// // NewQueryHelper creates a new instance of QueryHelper
// func NewQueryHelper(sql string) *QueryHelper {
// 	qh := &QueryHelper{
// 		sqlString:   sql,
// 		valuesToken: "",
// 	}
// 	return qh
// }

// // Compose reconstitutes the query and returns it
// func (qp *QueryHelper) Compose() string {
// 	return qp.first + qp.valuesToken + " " + strings.Join(qp.values, ", ") + " " + qp.last
// }

// // ComposeWithValues reconstitutes the query with external values
// func (qp *QueryHelper) ComposeWithValues(values []string) string {
// 	if qp.valuesToken != "" {
// 		return qp.first + qp.valuesToken + " " + strings.Join(values, ", ") + " " + qp.last
// 	} else {
// 		return qp.first + " " + strings.Join(values, ", ") + " " + qp.last
// 	}
// }

// // Decompose parses the SQL string to extract values from a SQL statement
// func (qp *QueryHelper) Decompose() []string {
// 	//sql := `INSERT INTO distributors (did, name) values (1, 'Cheese', 9.99), (2, 'Bread', 1.99), (3, 'Milk', 2.99) `
// 	// parts := strings.Split(qp.sqlString, "VALUES") //what if nested statement has a values too, not supporting that at the moment
// 	// if len(parts) == 1 {
// 	// 	parts = strings.Split(qp.sqlString, "values")
// 	// 	qp.valuesToken = "values"
// 	// } //should contain the values clause, since we are doing validation in the UI

// 	regex := regexp.MustCompile(`(?i) values *`)
// 	sqlpattern := SplitAfter(qp.sqlString, regex)
// 	if len(sqlpattern) == 1 {
// 		//Values provided by an expression.  Hopefully all on one line..
// 		return nil
// 	}

// 	// sqlpattern[0] = strings.ToLower(sqlpattern[0])
// 	// if sqlpattern[0].Contains("values") {
// 	// 	sqlpattern[0] = strings.Split(sqlpattern, "values")
// 	// }
// 	qp.first = sqlpattern[0]
// 	//qp.valuesToken = "values"
// 	spart := sqlpattern[1]
// 	spartLength := len(spart)

// 	i := 0
// 	braketCount := 0
// 	for i < spartLength {
// 		ch := spart[i]
// 		i++
// 		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
// 			continue
// 		}
// 		if ch == '(' {
// 			braketCount = braketCount + 1
// 			position := i
// 			for i < len(spart) {
// 				ch = spart[i]
// 				//fmt.Print(string(ch))
// 				i++
// 				if ch == '(' {
// 					braketCount = braketCount + 1
// 				}
// 				if ch == ')' || ch == 0 {
// 					braketCount = braketCount - 1
// 					if braketCount == 0 {
// 						break
// 					}
// 				}
// 			}
// 			value := "(" + spart[position:i-1] + ")"
// 			qp.values = append(qp.values, value)
// 			if i == spartLength {
// 				break
// 			}
// 			ch = spart[i]
// 			i++
// 			for ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
// 				ch = spart[i]
// 				i++
// 			}
// 			if ch != ',' {
// 				break
// 			}
// 		}
// 	}
// 	qp.last = spart[i-1:]
// 	return qp.values
// }

// SplitAfter function returns the values part of the SQL
// func SplitAfter(s string, re *regexp.Regexp) []string {
// 	var (
// 		r []string
// 		p int
// 	)
// 	is := re.FindAllStringIndex(s, -1)
// 	if is == nil {
// 		return append(r, s)
// 	}
// 	for _, i := range is {
// 		r = append(r, s[p:i[1]])
// 		p = i[1]
// 	}
// 	return append(r, s[p:])

// }

//certDecode extract certificate text from connection context
func certDecode(certIntf interface{}) (certBytes []byte, filename string, err error) {
	if certIntf == nil {
		return nil, "", fmt.Errorf("certDecode got nil certificate")
	}
	if cert, ok := certIntf.(map[string]interface{}); ok {
		if certStr, ok := cert["content"]; ok && len(strings.Split(certStr.(string), ",")) > 1 {
			filename = cert["filename"].(string)
			certStrEncoded := strings.Split(certStr.(string), ",")[1]
			certBytes, err = base64.StdEncoding.DecodeString(certStrEncoded)
			if err != nil {
				return nil, "", fmt.Errorf("sharedconnection::certDecode Failed to decode the base64 certificate from config: [%s]", err)
			}
			return
		}
	}
	return nil, "", fmt.Errorf("certDecode certificate from connection is not a map")
}

//GetConnection returns a deserialized conneciton object it does not establish a
//connection with the database. The client needs to call Login to establish a
//connection
/*
func GetConnection(connector *generic.Connection) (conn *Connection, err error) {
	connectorCacheMutex.Lock()
	defer connectorCacheMutex.Unlock()
	conn, ok := connectorCache[connector.GetId()]
	if ok {
		return conn, nil
	}

	conn = &Connection{}
	conn.Host = connector.GetSetting("host").(string)
	if port, ok := connector.GetSetting("port").(string); ok {
		conn.Port = port
	} else if port, ok := connector.GetSetting("port").(float64); ok {
		conn.Port = fmt.Sprintf("%.0f", port)
	} else {
		return nil, fmt.Errorf("GetConnection failed to decode port")
	}
	conn.User = connector.GetSetting("user").(string)
	conn.Password = connector.GetSetting("password").(string)
	conn.DbName = connector.GetSetting("databaseName").(string)
	if err := conn.validate(); err != nil {
		return nil, fmt.Errorf("Connection validation error %s", err.Error())
	}
	if connector.GetSetting("tlsparm") == nil {
		conn.TLSMode = "None"
	}
	if connector.GetSetting("tlsparm") != nil && connector.GetSetting("tlsparm").(string) != "None" {
		conn.TLSMode = connector.GetSetting("tlsparm").(string)
		conn.Cacert, _, err = certDecode(connector.GetSetting("cacert"))
		if err != nil {
			return nil, err
		}
		conn.Clientcert, _, err = certDecode(connector.GetSetting("clientcert"))
		if err != nil {
			return nil, err
		}
		conn.Clientkey, _, err = certDecode(connector.GetSetting("clientkey"))
		if err != nil {
			return nil, err
		}
	}
	connectorCache[connector.GetId()] = conn
	return conn, nil
}
*/

func keepalive() {
	tick := time.Tick(20 * time.Minute)
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-tick:
			for key, conn := range connectorCache {
				err := conn.db.Ping()
				if err != nil {
					log.RootLogger().Warnf("sharedconnection::keepalive Ping connection [%s] is dead: [%s]", key, err)
					conn.db = nil
				} else {
					log.RootLogger().Debugf("sharedconnection::keepalive Ping connection: [%s]", key)
				}
			}
		}
	}
}

//validate validates and  set config values to connection struct. It returns an error
//if required value is not provided or cannot be correctly converted
func (con *Connection) validate() (err error) {

	if con.Host == "" {
		return fmt.Errorf("Required parameter Host missing, %s", err.Error())
	}

	if con.Port == "" {
		return fmt.Errorf("Required parameter Port missing, %s", err.Error())
	}

	if con.User == "" {
		return fmt.Errorf("Required parameter User missing, %s", err.Error())
	}

	if con.DbName == "" {
		return fmt.Errorf("Required parameter DbName missing, %s", err.Error())
	}

	if con.Password == "" {
		return fmt.Errorf("Required parameter Password missing, %s", err.Error())
	}
	return nil
}

//Retrieve executes a dynamic SQL on the database and returns a result set
func (con *Connection) Retrieve(query Query, log log.Logger) (results *ResultSet, err error) {
	con.Login(log)

	var buffer bytes.Buffer

	buffer.WriteString(sqlSelect)
	buffer.WriteString(query.Cols[0])
	for j := 1; j < len(query.Cols); j = j + 1 {
		buffer.WriteString(", ")
		buffer.WriteString(query.Cols[j])
	}

	buffer.WriteString(" FROM " + query.TableName)
	buffer.WriteString(" WHERE ")
	remaining := len(query.Filters)
	for key, value := range query.Filters {
		buffer.WriteString(key)
		buffer.WriteString(" = ")
		buffer.WriteString(value)
		buffer.WriteString(" ")
		if remaining--; remaining > 0 {
			buffer.WriteString("AND ")
		}
	}

	queryString := string(buffer.Bytes())
	return con.SQLQuery(queryString, log)
}

func getTypeFromFields(fields interface{}, fieldname string) string {
	if fields == nil {
		return ""
	}
	for _, fldObj := range fields.([]interface{}) {
		if fldname, ok := fldObj.(map[string]interface{})["FieldName"]; ok {
			if fldname.(string) == fieldname {
				return fldObj.(map[string]interface{})["Type"].(string)
			}
		}
	}

	return ""
}

func decodeBlob(blob string) []byte {
	decodedBlob, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return []byte(blob)
	}
	return decodedBlob
}

func getSchemaFromFields(fields interface{}) map[string]string {
	schema := map[string]string{}
	for _, fieldObject := range fields.([]interface{}) {
		if fieldName, ok := fieldObject.(map[string]interface{})["FieldName"]; ok {
			schema[fieldName.(string)] = fieldObject.(map[string]interface{})["Type"].(string)
		}
	}
	return schema
}

func (sharedconnection *SharedConfigManager) getStatement(prepared string) (stmt *sql.Stmt, err error) {
	preparedQueryCacheMutex.Lock()
	defer preparedQueryCacheMutex.Unlock()
	stmt, ok := preparedQueryCache[prepared]
	if !ok {
		stmt, err = sharedconnection.db.Prepare(prepared)
		if err != nil {
			return nil, err
		}
		preparedQueryCache[prepared] = stmt
	}
	return stmt, nil
}

// func (sharedconnection *SharedConfigManager) returnStatement(prepared string, stmt *sql.Stmt) {
// 	preparedQueryCacheMutex.Lock()
// 	defer preparedQueryCacheMutex.Unlock()
// 	preparedQueryCache[prepared] = stmt
// }

//PreparedInsert allows querying database with named parameters
func (sharedconnection *SharedConfigManager) PreparedInsert(queryString string, inputData *Input, fields interface{}, log log.Logger) (results map[string]interface{}, err error) {
	// log.Debugf("Executing prepared query %s", queryString)
	// log.Debugf("inputParms: %v", inputData)

	schema := getSchemaFromFields(fields)
	prepared, inputArgs, paramsArray, err := EvaluateQuery(queryString, *inputData)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(paramsArray); i++ {
		parameterType, ok := schema[paramsArray[i]]
		if ok && parameterType == "BLOB" || parameterType == "IMAGE" {
			inputArgs[i] = decodeBlob(inputArgs[i].(string))
		}
	}
	// log.Debugf("Prepared insert about to exec statement [%s]", prepared)
	// log.Debugf("Prepared insert about to exec inputArgs [%v]", inputArgs)
	log.Debugf("Prepared insert statement: [%s] and Parameters: [%v] ", prepared, inputArgs)
	stmt, err := sharedconnection.getStatement(prepared)
	if err != nil {
		log.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}
	log.Debug("----------- DB Stats in Insert activity -----------")
	sharedconnection.printDBStats(sharedconnection.db, log)
	// defer stmt.Close()
	result, err := stmt.Exec(inputArgs...)
	if err != nil {
		log.Errorf("Executing prepared query got error: %s", err)
		// stmt.Close()
		return nil, err
	}
	// sharedconnection.returnStatement(prepared, stmt)
	output := make(map[string]interface{})
	output["rowsAffected"], _ = result.RowsAffected()
	output["lastInsertId"], _ = result.LastInsertId()
	return output, nil
}

//PreparedUpdate allows querying database with named parameters
func (sharedconnection *SharedConfigManager) PreparedUpdate(queryString string, inputData *Input, fields interface{}, log log.Logger) (results map[string]interface{}, err error) {
	// log.Debugf("Executing prepared query %s", queryString)
	// log.Debugf("Query parameters: %v", inputData)

	schema := getSchemaFromFields(fields)
	prepared, inputArgs, paramsArray, err := EvaluateQuery(queryString, *inputData)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(paramsArray); i++ {
		parameterType, ok := schema[paramsArray[i]]
		if ok && parameterType == "BLOB" || parameterType == "IMAGE" {
			inputArgs[i] = decodeBlob(inputArgs[i].(string))
		}
	}
	log.Debugf("Prepared update statement: [%s] and Parameters: [%v] ", prepared, inputArgs)
	stmt, err := sharedconnection.getStatement(prepared)
	if err != nil {
		log.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}
	log.Debug("----------- DB Stats in Update activity -----------")
	sharedconnection.printDBStats(sharedconnection.db, log)
	// defer stmt.Close()
	result, err := stmt.Exec(inputArgs...)
	if err != nil {
		log.Errorf("Executing prepared query got error: %s", err)
		// stmt.Close()
		return nil, err
	}
	// sharedconnection.returnStatement(prepared, stmt)
	output := make(map[string]interface{})
	output["rowsAffected"], _ = result.RowsAffected()
	return output, nil
}

// PreparedDelete allows deleting rows from a table with named parameters
func (sharedconnection *SharedConfigManager) PreparedDelete(queryString string, inputData *Input, log log.Logger) (results map[string]interface{}, err error) {
	// log.Debugf("Executing prepared query %s", queryString)

	prepared, inputArgs, paramsArray, err := EvaluateQuery(queryString, *inputData)
	if err != nil {
		return nil, err
	}
	logCache.Debugf("Prepared delete statement: [%s]  and  Parameters: [%v], Parameter Values : [%v] ", prepared, paramsArray, inputArgs)

	stmt, err := sharedconnection.getStatement(prepared)
	if err != nil {
		log.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}
	log.Debug("----------- DB Stats in Delete activity -----------")
	sharedconnection.printDBStats(sharedconnection.db, log)
	// defer stmt.Close()
	result, err := stmt.Exec(inputArgs...)
	if err != nil {
		log.Errorf("Executing prepared query got error: %s", err)
		// stmt.Close()
		return nil, err
	}
	// sharedconnection.returnStatement(prepared, stmt)
	output := make(map[string]interface{})
	output["rowsAffected"], _ = result.RowsAffected()
	return output, nil
}

//PreparedQuery allows querying database with named parameters
func (sharedconnection *SharedConfigManager) PreparedQuery(queryString string, inputData *Input, log log.Logger) (results *ResultSet, err error) {
	// log.Debugf("Executing prepared query %s", queryString)
	prepared, inputArgs, paramsArray, err := EvaluateQuery(queryString, *inputData)
	if err != nil {
		return nil, err
	}
	logCache.Debugf("Prepared query statement: [%s]  and  Parameters: [%v], Parameter Values : [%v] ", prepared, paramsArray, inputArgs)
	stmt, err := sharedconnection.getStatement(prepared)
	if err != nil {
		log.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}

	log.Debug("----------- DB Stats in Query activity -----------")
	sharedconnection.printDBStats(sharedconnection.db, log)

	// defer stmt.Close()
	rows, err := stmt.Query(inputArgs...)
	if err != nil {
		log.Errorf("Executing prepared query got error: %s", err)
		// stmt.Close()
		return nil, err
	}
	// sharedconnection.returnStatement(prepared, stmt)
	if rows == nil {
		log.Debugf("No rows returned for query %s", prepared)
		return nil, nil
	}
	defer rows.Close()
	return resultSetFromQueryObjs(rows, log)
}

//SQLQuery executes the SQL query provided in the string argument
func (con *Connection) SQLQuery(queryString string, log log.Logger) (results *ResultSet, err error) {
	con.Login(log)
	log.Debugf("%s", queryString)

	rows, err := con.db.Query(queryString, nil)
	if err != nil {
		return nil, fmt.Errorf("Error executing Query, %s", err.Error())
	}
	defer rows.Close()
	return resultSetFromQueryObjs(rows, log)
}

func resultSetFromQueryObjs(rows *sql.Rows, log log.Logger) (results *ResultSet, err error) {

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("Error getting column information, %s", err.Error())
	}

	count := len(columns)
	cols := make([]interface{}, count)
	args := make([]interface{}, count)
	coltypes, err := rows.ColumnTypes()
	if err != nil {
		fmt.Printf("%s", err.Error())
		return nil, fmt.Errorf("Error determining column types, %s", err.Error())
	}
	for i := range cols {
		args[i] = &cols[i]
	}
	var resultSet ResultSet
	columnCount := make(map[string]int)
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return nil, fmt.Errorf("Error scanning rows, %s", err.Error())
		}
		//m := make(map[string]interface{})
		m := make(record)
		for i, b := range cols {
			dbType := coltypes[i].DatabaseTypeName()
			//log.Debugf("dbtype is for column %s is %s", coltypes[i].Name(), dbType)

			//added check for duplicate columns in case of JOIN, WIMYSQ-492
			if _, found := m[columns[i]]; found {
				columnCount[columns[i]]++
				columns[i] = fmt.Sprintf("%s_%d", columns[i], columnCount[columns[i]])
			}
			if b == nil {
				m[columns[i]] = nil
				continue
			}
			switch dbType {
			case "VARCHAR", "CHAR", "TEXT", "TINYTEXT", "LONGTEXT", "MEDIUMTEXT", "DATE", "DATETIME", "TIME", "TIMESTAMP", "YEAR":
				m[columns[i]] = string(b.([]uint8))
			case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "BIGINT":
				m[columns[i]] = b.(int64)

			case "BIT":
				m[columns[i]] = b.([]uint8)[0]
			case "FLOAT":
				m[columns[i]] = b.(float32)
			case "DECIMAL", "NUMERIC":
				m[columns[i]], err = strconv.ParseFloat(string(b.([]uint8)), 64)
				if err != nil {
					return nil, err
				}
			case "DOUBLE":
				m[columns[i]] = b.(float64)
			case "TINYBLOB", "BLOB", "MEDIUMBLOB", "LONGBLOB":
				//fmt.Printf("%v", string(b.([]byte)))
				m[columns[i]] = base64.StdEncoding.EncodeToString(b.([]byte))
			default:
				log.Debugf("resultSetFromQueryObjs found uncategorized type: %s", dbType)
				m[columns[i]] = string(b.([]byte))
			}
		}
		if len(m) > 0 {
			resultSet.Record = append(resultSet.Record, &m)
		}
	}
	return &resultSet, nil
}

func (con *Connection) getTLSConfigFromConfig() (*tls.Config, error) {
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(con.Cacert) {
		return nil, fmt.Errorf("Failed to parse cacert PEM data from connection")
	}
	var cert tls.Certificate
	cert, err := tls.X509KeyPair(con.Clientcert, con.Clientkey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an internal keypair from the client cert and key provided on the connection for reason: %s", err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		RootCAs:            certpool,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

//decodeTLSParm translate tls parm to connection string
func decodeTLSParm(tlsparm string) string {
	switch tlsparm {
	case "Verify":
		return "?tls=true&timeout=2m30s"
	case "SkipVerify":
		return "?tls=skip-verify&timeout=2m30s"
	case "Preferred":
		return "?tls=optional&timeout=2m30s"
	default:
		return ""
	}
}

//Login connects to the the sharedconnection database cluster using the connection
//details provided in Connection configuration
func (con *Connection) Login(log log.Logger) (success bool, err error) {
	if con.db != nil {
		if con.db.Ping() != nil {
			log.Warn("sharedconnection.go:Login error on ping of existing connection: [%s] reconnecting", err)
			con.db = nil
		} else {
			log.Debugf("Reused connection for %s to %s", con.User, con.DbName)
			return true, nil
		}
	}
	if con.TLSMode == "" || con.TLSMode == "None" {
		log.Debugf("Login attempting plain connection: %s:********@tcp(%s:%d)/%s", con.User, con.Host, con.Port, con.DbName)
		con.db, err = sql.Open("sharedconnection", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", con.User, con.Password, con.Host, con.Port, con.DbName, "?timeout=2m30s"))
	} else {
		tlsconfig, err := con.getTLSConfigFromConfig()
		if err != nil {
			return false, err
		}
		mysql.RegisterTLSConfig("custom", tlsconfig)
		log.Debugf("Login attempting TLS connection: %s:********@tcp(%s:%d)/%s%s", con.User, con.Host, con.Port, con.DbName, decodeTLSParm(con.TLSMode))
		con.db, err = sql.Open("sharedconnection", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", con.User, con.Password, con.Host, con.Port, con.DbName, decodeTLSParm(con.TLSMode)))
	}
	if err != nil {
		return false, fmt.Errorf("Could not open connection to database %s, %s", con.DbName, err.Error())
	}
	err = con.db.Ping()
	if err != nil {
		con.db = nil
		return false, fmt.Errorf("Could not ping connection to database %s, %s", con.DbName, err.Error())
	}
	log.Debugf("Logged in %s to %s", con.User, con.DbName)
	return true, nil
}

//Logout the database connection
func (con *Connection) Logout(log log.Logger) (err error) {
	if con.db == nil {
		return nil
	}
	err = con.db.Close()
	if err != nil {
		log.Debugf("Failed to close connection for reason: %s", err)
	}
	log.Debugf("Logged out %s to %s", con.User, con.DbName)
	return
}

func (sharedconnection *SharedConfigManager) printDBStats(db *sql.DB, log log.Logger) {
	log.Debug("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	log.Debug("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	log.Debug("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	log.Debug("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	log.Debug("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	log.Debug("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))
}
