package connection

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
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

type (
	//Connection datastructure for storing SQLServer connection details
	Connection struct {
		DatabaseURL  string `json:"databaseURL"`
		Host         string `json:"host"`
		Port         int    `json:"port"`
		User         string `json:"user"`
		Password     string `json:"password"`
		DatabaseName string `json:"databaseName"`
		SSLMode      string `json:"sslmode"`
		db           *sql.DB
		connKey      string
	}

	//Connector is a representation of connector.json metadata for the sqlserver connection
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

	//Query structure for SQL Queries
	Query struct {
		TableName string            `json:"tableName"`
		Cols      []string          `json:"columns"`
		Filters   map[string]string `json:"filters"`
	}

	// QueryActivity provides Activity metadata for Flogo
	QueryActivity struct {
		metadata *activity.Metadata
	}

	//record is the one row of the ResultSet retrieved from database after execution of SQL Query
	record map[string]interface{}

	//ResultSet is an aggregation of SQL Query data records fetched from the database
	ResultSet struct {
		Record []*record `json:"records"`
	}

	//Input is a representation of acitivity's input parametres
	Input struct {
		Parameters map[string]interface{}   `json:"parameters,omitempty"`
		Values     []map[string]interface{} `json:"values,omitempty"`
	}
)

var connectorCache map[string]*Connection
var connectorCacheMutex sync.Mutex

var preparedQueryCache map[string]*sql.Stmt
var preparedQueryCacheMutex sync.Mutex

// var versionPrinted = false

func init() {
	connectorCache = make(map[string]*Connection, 100)
	preparedQueryCache = make(map[string]*sql.Stmt, 100)
	go keepalive()
}

// func version() {
// 	if !versionPrinted {
// 		contribution.PrintVersion(logCache)
// 		versionPrinted = true
// 	}
// }

// func keepalive() {
// 	tick := time.Tick(20 * time.Minute)
// 	for {
// 		select {
// 		// Got a timeout! fail with a timeout error
// 		case <-tick:
// 			for key, conn := range connectorCache {
// 				err := conn.db.Ping()
// 				if err != nil {
// 					log.RootLogger().Warnf("SqlServer cached connection  keepalive Ping connection [%s] is dead: [%s]", key, err)
// 					conn.db = nil
// 					delete(preparedQueryCache, key)
// 					delete(connectorCache, key)
// 				} else {
// 					log.RootLogger().Debugf("SqlServer cached connection keepalive Ping connection: [%s]", key)
// 				}
// 			}
// 		}
// 	}
// }

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
// This was taken from the sqlserver connector verbatum.  Its a shame we can't share code...
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
// 		valuesToken: "VALUES",
// 	}
// 	return qh
// }

// // Compose reconstitutes the query and returns it
// func (qp *QueryHelper) Compose() string {
// 	return qp.first + qp.valuesToken + " " + strings.Join(qp.values, ", ") + " " + qp.last
// }

// // ComposeWithValues reconstitutes the query with external values
// func (qp *QueryHelper) ComposeWithValues(values []string) string {
// 	return qp.first + qp.valuesToken + " " + strings.Join(values, ", ") + " " + qp.last
// }

// // Decompose parses the SQL string to extract values from a SQL statement
// func (qp *QueryHelper) Decompose() []string {
// 	//sql := `INSERT INTO distributors (did, name) values (1, 'Cheese', 9.99), (2, 'Bread', 1.99), (3, 'Milk', 2.99) `
// 	parts := strings.Split(qp.sqlString, "VALUES") //what if nested statement has a values too, not supporting that at the moment
// 	if len(parts) == 1 {
// 		parts = strings.Split(qp.sqlString, "values")
// 		qp.valuesToken = "values"
// 	} //should contain the values clause, since we are doing validation in the UI
// 	if len(parts) == 1 {
// 		//Values provided by an expression.  Hopefully all on one line..
// 		return nil
// 	}

// 	qp.first = parts[0]
// 	spart := parts[1]
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

func (sharedConn *SharedConfigManager) getSchema(fields interface{}) map[string]string {
	schema := map[string]string{}
	for _, fieldObject := range fields.([]interface{}) {
		if fieldName, ok := fieldObject.(map[string]interface{})["FieldName"]; ok {
			schema[fieldName.(string)] = fieldObject.(map[string]interface{})["Type"].(string)
		}
	}
	return schema
}

// GetConnection returns a deserialized conneciton object it does not establish a
// connection with the database. The client needs to call Login to establish a
// connection
// func GetConnection(connector interface{}) (*SharedConfigManager, error) {
// 	// version(log)
// 	genericConn, err := genericN.NewConnection(connector)
// 	if err != nil {
// 		return nil, errors.New("Failed to load SQLServer connection configuration")
// 	}

// 	// conn, ok := connectorCache[genericConn.GetId()]
// 	// if ok
// 	{
// 		// 	return conn, nil
// 		// }
// 		conn := &Connection{}
// 		conn.Host, err = data.CoerceToString(genericConn.GetSetting("host"))
// 		if err != nil {
// 			return nil, fmt.Errorf("connection getter for host failed: %s", err)
// 		}
// 		logCache.Debugf("getconnection processed host: %s", conn.Host)
// 		conn.Port, err = data.CoerceToInteger(genericConn.GetSetting("port"))
// 		if err != nil {
// 			return nil, fmt.Errorf("connection getter for port failed: %s", err)
// 		}
// 		logCache.Debugf("getconnection processed port: %d", conn.Port)
// 		conn.User, err = data.CoerceToString(genericConn.GetSetting("user"))
// 		if err != nil {
// 			return nil, fmt.Errorf("connection getter for user failed: %s", err)
// 		}
// 		logCache.Debugf("getconnection processed user: %s", conn.User)
// 		conn.Password, err = data.CoerceToString(genericConn.GetSetting("password"))
// 		if err != nil {
// 			return nil, fmt.Errorf("connection getter for password failed: %s", err)
// 		}
// 		logCache.Debugf("getconnection processed databaseName: %s", conn.DatabaseName)
// 		conn.DatabaseName, err = data.CoerceToString(genericConn.GetSetting("databaseName"))
// 		if err != nil {
// 			return nil, fmt.Errorf("connection getter for databaseName failed: %s", err)
// 		}
// 		// err = conn.validate()
// 		// if err != nil {
// 		// 	return nil, fmt.Errorf("Connection validation error %s", err.Error())
// 		// }
// 		// conninfo := getConnectionSignature(conn)
// 		// connKey := fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(conninfo)))
// 		// cachedCon, ok := connectorCache[connKey]
// 		// if ok {
// 		// 	return cachedCon, nil
// 		// }
// 		return conn, nil
// 	}
// }

// func getConnectionSignature(con *Connection) string {
// 	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", con.Host, con.User, con.Password, con.Port, con.DbName)
// 	return connString
// }
// func getConnectionSignatureForDebug(con *Connection) string {
// 	connString := fmt.Sprintf("server=%s;user id=%s;password=xxxxxxxx;port=%d;database=%s", con.Host, con.User, con.Port, con.DbName)
// 	return connString
// }

//validate validates and  set config values to connection struct. It returns an error
//if required value is not provided or cannot be correctly converted
// func (con *SharedConfigManager) validate() (err error) {

// 	if con.Host == "" {
// 		return fmt.Errorf("Required parameter Host missing, %s", err.Error())
// 	}

// 	if con.Port == 0 {
// 		return fmt.Errorf("Required parameter Port missing, %s", err.Error())
// 	}

// 	if con.User == "" {
// 		return fmt.Errorf("Required parameter User missing, %s", err.Error())
// 	}

// 	if con.DbName == "" {
// 		return fmt.Errorf("Required parameter DbName missing, %s", err.Error())
// 	}

// 	if con.Password == "" {
// 		return fmt.Errorf("Required parameter Password missing, %s", err.Error())
// 	}
// 	return nil
// }

func convertFieldsToSchema(fields interface{}) map[string]string {
	if fields == nil {
		return nil
	}
	schema := map[string]string{}
	for _, fieldObject := range fields.([]interface{}) {
		if fieldName, ok := fieldObject.(map[string]interface{})["FieldName"]; ok {
			schema[fieldName.(string)] = fieldObject.(map[string]interface{})["Type"].(string)
		}
	}
	return schema
}

// func (sharedConn *SharedConfigManager) getStatement(prepared string) (stmt *sql.Stmt, err error) {
// 	preparedQueryCacheMutex.Lock()
// 	defer preparedQueryCacheMutex.Unlock()
// 	stmt, ok := preparedQueryCache[prepared]
// 	if ok {
// 		delete(preparedQueryCache, prepared)
// 	} else {
// 		stmt, err = sharedConn.db.Prepare(prepared)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return
// }

// fix for WIMSSS-244
func (sharedConn *SharedConfigManager) getStatement(prepared string) (stmt *sql.Stmt, err error) {
	preparedQueryCacheMutex.Lock()
	defer preparedQueryCacheMutex.Unlock()
	stmt, ok := preparedQueryCache[prepared]
	if !ok {
		stmt, err = sharedConn.db.Prepare(prepared)
		if err != nil {
			return nil, err
		}
		preparedQueryCache[prepared] = stmt
	}
	return stmt, nil
}

//fix for WIMSSS-244
// func (sharedConn *SharedConfigManager) returnStatement(prepared string, stmt *sql.Stmt) {
// 	preparedQueryCacheMutex.Lock()
// 	defer preparedQueryCacheMutex.Unlock()
// 	preparedQueryCache[prepared] = stmt
// }

// PreparedInsert allows querying database with named parameters
func (sharedConn *SharedConfigManager) PreparedInsert(queryString string, inputData *Input, fields interface{}, logCache log.Logger, queryTimeout int) (results map[string]interface{}, err error) {
	// func (sharedConn *SharedConfigManager) PreparedInsert(queryString string, inputData *Input, fields interface{}) (results *ResultSet, err error) {
	var ctx context.Context
	var cancel context.CancelFunc
	var result sql.Result

	schema := sharedConn.getSchema(fields)
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
	// logCache.Debugf("Prepared insert about to exec statement [%s]", prepared)
	// logCache.Debugf("Prepared insert about to exec inputArgs [%v]", inputArgs)
	logCache.Debugf("Prepared insert statement: [%s] and  Parameters: [%v] ", prepared, inputArgs)

	stmt, err := sharedConn.getStatement(prepared)
	// stmt, err := sharedConn.db.Prepare(prepared)
	if err != nil {
		logCache.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}

	logCache.Debugf("----------- DB Stats in Insert activity -----------")
	sharedConn.printDBStats(sharedConn.db, logCache)
	if queryTimeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(queryTimeout)*time.Second)
		defer cancel()

		result, err = stmt.ExecContext(ctx, inputArgs...)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				err = errors.New("Query execution timed out...")
			}
			logCache.Errorf("Executing prepared query got error: %s", err)
			// stmt.Close()
			return nil, err
		}
	} else {
		// querytimeout=0; no context with timeout unlimited
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result, err = stmt.ExecContext(ctx, inputArgs...)
		if err != nil {
			logCache.Errorf("Executing prepared query got error: %s", err)
			return nil, err
		}
	}

	// sharedConn.returnStatement(prepared, stmt)

	output := make(map[string]interface{})
	output["rowsAffected"], _ = result.RowsAffected()
	output["lastInsertId"], _ = result.LastInsertId()
	return output, nil
}

// PreparedDelete allows querying database with named parameters
// func (con *SharedConfigManager) PreparedDelete(queryString string, inputData *Input, log logger.Logger) (results map[string]interface{}, err error) {
func (sharedConn *SharedConfigManager) PreparedDelete(queryString string, inputData *Input, logCache log.Logger, queryTimeout int) (results map[string]interface{}, err error) {
	var ctx context.Context
	var cancel context.CancelFunc
	var result sql.Result

	prepared, inputArgs, _, err := EvaluateQuery(queryString, *inputData)
	if err != nil {
		return nil, err
	}
	logCache.Debugf("Prepared delete statement: [%s]  and  Parameters: [%v] ", prepared, inputArgs)

	stmt, err := sharedConn.getStatement(prepared)
	// stmt, err := sharedConn.db.Prepare(prepared)
	if err != nil {
		logCache.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}

	logCache.Debugf("----------- DB Stats in Delete activity -----------")
	sharedConn.printDBStats(sharedConn.db, logCache)

	if queryTimeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(queryTimeout)*time.Second)
		defer cancel()
		result, err = stmt.ExecContext(ctx, inputArgs...)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				err = errors.New("Query execution timed out...")
			}
			logCache.Errorf("Executing prepared query got error: %s", err)
			// stmt.Close()
			return nil, err
		}
	} else {
		// querytimeout=0; no context with timeout unlimited
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result, err = stmt.ExecContext(ctx, inputArgs...)
		if err != nil {
			logCache.Errorf("Executing prepared query got error: %s", err)
			return nil, err
		}
	}

	// sharedConn.returnStatement(prepared, stmt)

	output := make(map[string]interface{})
	output["rowsAffected"], _ = result.RowsAffected()
	return output, nil
}

// PreparedQuery allows querying database with named parameters
// func (con *SharedConfigManager) PreparedQuery(queryString string, inputData *Input, log logger.Logger) (results *ResultSet, err error) {
func (sharedConn *SharedConfigManager) PreparedQuery(queryString string, inputData *Input, logCache log.Logger, queryTimeout int) (results *ResultSet, err error) {
	var ctx context.Context
	var cancel context.CancelFunc
	var rows *sql.Rows

	prepared, inputArgs, _, err := EvaluateQuery(queryString, *inputData)
	if err != nil {
		return nil, err
	}
	logCache.Debugf("Prepared query statement: [%s]  and  Parameters: [%v] ", prepared, inputArgs)

	stmt, err := sharedConn.getStatement(prepared)
	// stmt, err := sharedConn.db.Prepare(prepared)
	if err != nil {
		logCache.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}

	logCache.Debugf("----------- DB Stats in Query activity -----------")
	sharedConn.printDBStats(sharedConn.db, logCache)

	if queryTimeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(queryTimeout)*time.Second)
		defer cancel()
		rows, err = stmt.QueryContext(ctx, inputArgs...)
		//rows, err := stmt.Query(inputArgs...)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				err = errors.New("Query execution timed out...")
			}
			logCache.Errorf("Executing prepared query got error: %s", err)
			// stmt.Close()
			return nil, err
		}
	} else {
		// querytimeout=0; no context with timeout unlimited
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rows, err = stmt.QueryContext(ctx, inputArgs...)
		if err != nil {
			logCache.Errorf("Executing prepared query got error: %s", err)
			return nil, err
		}
	}

	// sharedConn.returnStatement(prepared, stmt)

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
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			// fmt.Printf("%s", err.Error())
			if strings.Contains(err.Error(), "context deadline exceeded") {
				err = errors.New("Query execution timed out...")
				logCache.Errorf("Error scanning rows: %s", err)
			}
			return nil, fmt.Errorf("Error scanning rows, %s", err)
		}

		m := make(record)
		for i, b := range cols {
			dbType := coltypes[i].DatabaseTypeName()
			//added check for duplicate columns in case of JOIN, WIMSSS-346
			if _, found := m[columns[i]]; found {
				columnCount[columns[i]]++
				columns[i] = fmt.Sprintf("%s_%d", columns[i], columnCount[columns[i]])
			}
			switch dbType {
			case "DECIMAL", "MONEY", "SMALLMONEY":
				if b == nil {
					m[columns[i]] = b
				} else {
					x := b.([]byte)
					if nx, err := strconv.ParseFloat(string(x), 64); err == nil {
						m[columns[i]] = nx
					}
				}
			case "BINARY", "IMAGE", "VARBINARY":
				if b == nil {
					m[columns[i]] = b
				} else {
					m[columns[i]] = base64.StdEncoding.EncodeToString(b.([]byte))
				}
			case "DATE":
				if b == nil {
					m[columns[i]] = b
				} else {
					x := b.(time.Time)
					d := DateOf(x)
					m[columns[i]] = fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
				}
			case "TIME":
				if b == nil {
					m[columns[i]] = b
				} else {
					x := b.(time.Time)
					t := TimeOf(x)
					s := fmt.Sprintf("%02d:%02d:%02d", t.Hour, t.Minute, t.Second)
					if t.Nanosecond == 0 {
						m[columns[i]] = s
					} else {
						m[columns[i]] = s + fmt.Sprintf(".%09d", t.Nanosecond)
					}
				}
			case "UNIQUEIDENTIFIER":
				// First four parts of guid are integers 32 16 16 16 bit respectively.
				// the reultset []byte incorrectly interprets data big-endian/little-endian (http://en.wikipedia.org/wiki/Endianness) mixup.
				// So swap order of bytes to get correct byte order.
				// ref: https://github.com/denisenkom/go-mssqldb/issues/56
				if b == nil {
					m[columns[i]] = b
				} else {
					x := b.([]byte)
					x[0], x[1], x[2], x[3] = x[3], x[2], x[1], x[0]
					x[4], x[5] = x[5], x[4]
					x[6], x[7] = x[7], x[6]

					// mssql.UNIQUEIDENTIFIER type which is of type [16]byte
					type UUID [16]byte
					if len(x) != 16 {
						return nil, fmt.Errorf("Given slice is not valid UUID sequence")
					}
					u := new(UUID)
					copy(u[:], x)
					idStr := fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
					m[columns[i]] = idStr
				}
			default:
				m[columns[i]] = b
			}
		}
		if len(m) > 0 {
			resultSet.Record = append(resultSet.Record, &m)
		}
	}

	return &resultSet, nil
}

// A Time represents a time with nanosecond precision.
//
// This type does not include location information, and therefore does not
// describe a unique moment in time.
//
// This type exists to represent the TIME type in storage-based APIs like BigQuery.
// Most operations on Times are unlikely to be meaningful. Prefer the DateTime type.
type Time struct {
	Hour       int // The hour of the day in 24-hour format; range [0-23]
	Minute     int // The minute of the hour; range [0-59]
	Second     int // The second of the minute; range [0-59]
	Nanosecond int // The nanosecond of the second; range [0-999999999]
}

// TimeOf returns the Time representing the time of day in which a time occurs
// in that time's location. It ignores the date.
func TimeOf(t time.Time) Time {
	var tm Time
	tm.Hour, tm.Minute, tm.Second = t.Clock()
	tm.Nanosecond = t.Nanosecond()
	return tm
}

// DateOf returns the Date in which a time occurs in that time's location.
func DateOf(t time.Time) Date {
	var d Date
	d.Year, d.Month, d.Day = t.Date()
	return d
}

// A Date represents a date (year, month, day).
//
// This type does not include location information, and therefore does not
// describe a unique 24-hour timespan
type Date struct {
	Year  int        // Year (e.g., 2014).
	Month time.Month // Month of the year (January = 1, ...).
	Day   int        // Day of the month, starting at 1.
}

func isBlob(typeName string) bool {
	if strings.Index(typeName, "BLOB") >= 0 {
		return true
	} else if strings.Index(typeName, "IMAGE") >= 0 {
		return true
	}
	return false
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

// PreparedUpdate allows querying database with named parameters
// func (con *SharedConfigManager) PreparedUpdate(queryString string, inputData *Input, fields interface{}, log logger.Logger) (results map[string]interface{}, err error) {
func (sharedConn *SharedConfigManager) PreparedUpdate(queryString string, inputData *Input, fields interface{}, logCache log.Logger, queryTimeout int) (results map[string]interface{}, err error) {
	var ctx context.Context
	var cancel context.CancelFunc
	var result sql.Result

	schema := sharedConn.getSchema(fields)
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
	logCache.Debugf("Prepared update statement: [%s] and Parameters: [%v] ", prepared, inputArgs)

	stmt, err := sharedConn.getStatement(prepared)
	// stmt, err := sharedConn.db.Prepare(prepared)
	if err != nil {
		logCache.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}

	logCache.Debugf("----------- DB Stats in Update activity -----------")
	sharedConn.printDBStats(sharedConn.db, logCache)

	if queryTimeout != 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(queryTimeout)*time.Second)
		defer cancel()
		result, err = stmt.ExecContext(ctx, inputArgs...)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				err = errors.New("Query execution timed out...")
			}
			logCache.Errorf("Executing prepared query got error: %s", err)
			// stmt.Close()
			return nil, err
		}
	} else {
		// querytimeout=0; no context with timeout unlimited
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		result, err = stmt.ExecContext(ctx, inputArgs...)
		if err != nil {
			logCache.Errorf("Executing prepared query got error: %s", err)
			return nil, err
		}
	}
	// sharedConn.returnStatement(prepared, stmt)

	output := make(map[string]interface{})
	output["rowsAffected"], _ = result.RowsAffected()
	return output, nil
}

func (sharedConn *SharedConfigManager) printDBStats(db *sql.DB, logCache log.Logger) {
	logCache.Debugf("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	logCache.Debugf("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	logCache.Debugf("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	logCache.Debugf("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	logCache.Debugf("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	logCache.Debugf("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))
}

//Login connects to the the sqlserver database cluster using the connection
//details provided in Connection configuration
// func (con *Connection) Login(log logger.Logger) (success bool, err error) {
// func (con *Connection) Login() (success bool, err error) {

// 	success, err = con.login()
// 	if err != nil {
// 		success, err = con.login()
// 	}
// 	return
// }

//login connects to the the sqlserver database cluster using the connection
//details provided in Connection configuration
// func (con *Connection) login(log logger.Logger) (success bool, err error) {
// func (con *Connection) login() (success bool, err error) {
// 	// conninfo := getConnectionSignature(con)
// 	// connKey := fmt.Sprintf("%d", crc32.ChecksumIEEE([]byte(conninfo)))
// 	// coninfoForDebug := getConnectionSignatureForDebug(con)
// 	conninfo := fmt.Sprintf("host=%s port=%d user=%s "+
// 		"password=%s dbname=%s sslmode=disable",
// 		con.Host, con.Port, con.User, con.Password, con.DatabaseName)
// 	logCache.Infof("SQLServer.Login() using connection string: %s", conninfo)
// 	// if con.db != nil {
// 	// return true, nil
// 	// }
// 	// db, err := sql.Open("mssql", conninfo)
// 	db, err := sql.Open("mssql", conninfo)

// 	if err != nil {
// 		return false, fmt.Errorf("Could not open connection to database %s, %s", con.DatabaseName, err.Error())
// 	}
// 	err = db.Ping()
// 	if err != nil {
// 		logCache.Infof("SqlServer.go:Login error on ping of new connection: [%s]", err)
// 		return false, err
// 	}
// 	con.db = db
// 	// con.connKey = connKey
// 	// connectorCache[connKey] = con
// 	// preparedQueryCache[connKey] = make(map[string]*sql.Stmt)

// 	logCache.Infof("Logged in %s to %s", con.User, con.DatabaseName)
// 	return true, nil
// }

// Login connects to the the postgres database cluster using the connection
// details provided in Connection configuration
// func (con *Connection) Login() (err error) {
// 	if con.db != nil {
// 		return nil
// 	}
// 	logCache.Infof("Sqlserver.go login")

// 	conninfo := fmt.Sprintf("host=%s;port=%d;user id=%s;password=%s;database=%s;sslmode=disable;",
// 		con.Host, con.Port, con.User, con.Password, con.DatabaseName)
// 	logCache.Infof("sqlserver.go conninfo %s", conninfo)
// 	db, err := sql.Open("mssql", conninfo)
// 	if err != nil {
// 		return fmt.Errorf("Could not open connection to database %s, %s", con.DatabaseName, err.Error())
// 	}
// 	con.db = db

// 	err = db.Ping()
// 	if err != nil {
// 		return err
// 	}
// 	logCache.Infof("Logged in %s to %s", con.User, con.DatabaseName)
// 	// log.Infof("login successful")
// 	return nil
// }

// Logout the database connection
// func (con *Connection) Logout(log logger.Logger) (err error) {
// func (con *Connection) Logout() (err error) {
// 	if con.db == nil {
// 		return nil
// 	}
// 	logCache.Debugf("Logout, connection is cached for %s to %s", con.User, con.DatabaseName)
// 	return
// }
