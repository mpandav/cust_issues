package connection

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync"
	"time"

	"database/sql"

	"github.com/alexbrainman/odbc/api"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/execute"
)

type (
	//Connection datastructure for storing TDV connection details
	Connection struct {
		DatabaseURL string `json:"databaseURL"`
		Server      string `json:"server"`
		Port        int    `json:"port"`
		User        string `json:"user"`
		Password    string `json:"password"`
		DataSource  string `json:"datasource"`
		Domain      string `json:"domain"`
		SSLMode     string `json:"sslmode"`
		db          *sql.DB
	}

	//Connector is a representation of connector.json metadata for the tdv connection
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
	}
	//ProcResultSet is an aggregation of all results returned by procedure
	ProcResultSet map[string]interface{}

	//Input is a representation of acitivity's input parametres
	Input struct {
		Parameters map[string]interface{}   `json:"parameters,omitempty"`
		Values     []map[string]interface{} `json:"values,omitempty"`
	}
)

var connectorCache map[string]*Connection

var preparedQueryCache map[string]*sql.Stmt
var preparedQueryCacheMutex sync.Mutex

func init() {
	connectorCache = make(map[string]*Connection, 100)
	preparedQueryCache = make(map[string]*sql.Stmt, 100)
	go keepalive()

}

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

func UnmarshalRows(rows *sql.Rows) (results *ResultSet, err error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting column information, %s", err.Error())
	}

	count := len(columns)
	cols := make([]interface{}, count)
	args := make([]interface{}, count)
	coltypes, err := rows.ColumnTypes()

	if err != nil {
		return nil, fmt.Errorf("error determining column types, %s", err.Error())
	}

	logCache.Debug("Column types returned by driver: ")
	for i := range cols {
		logCache.Debugf("%v: %v", columns[i], coltypes[i].DatabaseTypeName())
		args[i] = &cols[i]
	}

	var resultSet ResultSet
	columnCount := make(map[string]int)
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return nil, fmt.Errorf("error scanning rows, %s", err.Error())
		}
		m := make(record)
		for i, b := range cols {
			dbType := coltypes[i].DatabaseTypeName()
			//added check for duplicate columns in case of JOIN, refer WIPGRS-452
			if _, found := m[columns[i]]; found {
				columnCount[columns[i]]++
				columns[i] = fmt.Sprintf("%s_%d", columns[i], columnCount[columns[i]])
			}
			if b == nil {
				m[columns[i]] = nil
				continue
			}
			//TODO check types
			switch dbType {
			case "NUMERIC", "DECIMAL":
				x := b.([]uint8)
				if nx, ok := strconv.ParseFloat(string(x), 64); ok == nil {
					m[columns[i]] = nx
				}
			case "CHAR", "VARCHAR", "BIT", "INTERVAL_DAY", "INTERVAL_MONTH", "INTERVAL_YEAR", "XML":
				x, ok := b.([]byte)
				if !ok {
					return nil, fmt.Errorf("error converting %s to []byte, %s", dbType, err.Error())
				}
				m[columns[i]] = string(bytes.Trim(x, "\x00"))
			case "SMALLINT", "TINYINT", "INTEGER":
				m[columns[i]] = b.(int32)
			case "BIGINT":
				m[columns[i]] = b.(int64)
			case "DOUBLE", "REAL", "FLOAT":
				m[columns[i]] = b.(float64)
			case "BYTEA":
				m[columns[i]] = base64.StdEncoding.EncodeToString(b.([]byte))
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

// GetStatement
func (p *TDVSharedConfigManager) getStatement(prepared string) (stmt *sql.Stmt, err error) {
	preparedQueryCacheMutex.Lock()
	defer preparedQueryCacheMutex.Unlock()
	stmt, ok := preparedQueryCache[prepared]
	if !ok {
		stmt, err = p.db.Prepare(prepared)
		if err != nil {
			return nil, err
		}
		preparedQueryCache[prepared] = stmt
	}
	return stmt, nil
}

func checkCount(rows *sql.Rows) (count int, err error) {
	logCache.Debugf("Inside check count rows for update")
	// var counter int
	// defer rows.Close()
	for rows.Next() {
		logCache.Debugf("Row found")
		if err := rows.Scan(&count); err != nil {
			//log.Fatal(err)
		}
		logCache.Debugf("Counter: %d", count)
	}
	return count, err
}

// PreparedQuery allows querying database with named parameters
func (p *TDVSharedConfigManager) PreparedQuery(queryString string, inputData *Input, paramTypes []string, logCache log.Logger) (results *ResultSet, err error) {
	// logCache.Debugf("Executing prepared query %s", queryString)
	logCache.Debugf("Input from activity : %v", inputData)
	logCache.Debugf("Prepared Parameter Types: %v", paramTypes)
	prepared, err := EvaluateQuery(queryString, *inputData, paramTypes)
	if err != nil {
		return nil, err
	}
	stmt, err := p.getStatement(prepared)
	if err != nil {
		logCache.Errorf("Failed to prepare statement: %s", err)
		return nil, err
	}

	logCache.Debug("----------- DB Stats in Query activity -----------")
	p.printDBStats(p.db, logCache)

	// defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		logCache.Errorf("Executing prepared query got error: %s", err)
		// stmt.Close()
		return nil, err
	}

	if rows == nil {
		logCache.Debugf("no rows returned for query %s", prepared)
		return nil, nil
	}
	// p.returnStatement(prepared, stmt)
	defer rows.Close()
	// logCache.Debug("Return from PreparedQuery")
	return UnmarshalRows(rows)
}

// Login connects to the the tdv database cluster using the connection
// details provided in Connection configuration
func (con *Connection) Login() (err error) {
	if con.db != nil {
		return nil
	}

	conninfo := fmt.Sprintf("server=%s;port=%d;user=%s;"+
		"password=%s;dbname=%s;domain=%s;",
		con.Server, con.Port, con.User, con.Password, con.DataSource, con.Domain)

	db, err := sql.Open("odbc", conninfo)
	if err != nil {
		return fmt.Errorf("Could not open connection to database %s, %s", con.DataSource, err.Error())
	}
	con.db = db

	err = db.Ping()
	if err != nil {
		return err
	}

	// log.Infof("login successful")
	return nil
}

// Logout the database connection
func (con *Connection) Logout() (err error) {
	if con.db == nil {
		return nil
	}
	err = con.db.Close()
	// log.Infof("Logged out %s to %s", con.User, con.DataSource)
	return
}

func (p *TDVSharedConfigManager) printDBStats(db *sql.DB, log log.Logger) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// log.Debug("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	log.Debug("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	log.Debug("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	log.Debug("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	log.Debug("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	log.Debug("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))
}
func (p *TDVSharedConfigManager) GetPreparedStatement(queryString string) (api.SQLHSTMT, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	stmt, err := execute.GetStatementHandle(p.CgoConnection.H)
	if err != nil {
		return stmt, err
	}
	stmt, err = execute.PrepareStatment(stmt, queryString)
	if err != nil {
		return stmt, err
	}
	return stmt, nil
}

func (p *TDVSharedConfigManager) ExecuteProcedure(stmt api.SQLHSTMT, queryString string, inputData *Input, inputParamMetadata []execute.InputParamMetadata, inputParamPositions []int, NumberOfBindableParams int, cursors []string, ps []execute.Parameter, logCache log.Logger) (*ProcResultSet, error) {
	var err error

	//Bind input parameters
	for i, param := range inputParamMetadata {
		if param.ParamType == "INTEGER" || param.ParamType == "SMALLINT" || param.ParamType == "BIGINT" || param.ParamType == "TINYINT" {
			floatData := inputData.Parameters[param.ParamName].(float64)
			intData := int64(floatData)
			ps[i].BindValue(stmt, inputParamPositions[i], intData)
		} else {
			ps[i].BindValue(stmt, inputParamPositions[i], inputData.Parameters[param.ParamName])
		}

	}
	stmt, err = execute.ExecuteStatement(stmt)
	if err != nil {
		return nil, err
	}
	var result ProcResultSet
	result, err = execute.FetchAllCursorResults(stmt, cursors, logCache)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func ReleaseStatmentHandle(stmt api.SQLHSTMT) {
	coredbutils.ReleaseHandle(stmt)
}
