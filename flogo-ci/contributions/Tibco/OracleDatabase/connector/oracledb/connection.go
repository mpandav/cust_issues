package oracledb

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"

	//register oracle dialer
	_ "github.com/godror/godror"
)

var logCache = log.ChildLogger(log.RootLogger(), "OracleDatabase-connection")
var factory = &OracleDatabaseFactory{}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// Settings corresponds to connection.json settings
type Settings struct {
	Name               string `md:"name,required"`
	Host               string `md:"host,required"`
	Port               int    `md:"port,required"`
	User               string `md:"user,required"`
	Password           string `md:"password,required"`
	Database           string `md:"database,required"`
	SID                string `md:"SID,required"`
	SetMaxOpenConns    int    `md:"SetMaxOpenConns"`
	SetMaxIdleConns    int    `md:"SetMaxIdleConns"`
	SetConnMaxLifetime string `md:"SetConnMaxLifetime"`
}

// OracleDatabaseFactory structure
type OracleDatabaseFactory struct {
}

// Type method of connection.ManagerFactory must be implemented by OracleDatabaseFactory
func (*OracleDatabaseFactory) Type() string {
	return "OracleDatabase"
}

// NewManager method of connection.ManagerFactory must be implemented by OracleDatabaseFactory
func (*OracleDatabaseFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &OracleDatabaseSharedConfigManager{}
	var err error
	s := &Settings{}

	err = metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}

	logCache.Debug("Connecting to Oracle Database")
	//loginInfo := fmt.Sprintf("%s/%s@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=%s)(PORT=%v))(CONNECT_DATA=(SID=%s)))", s.User, s.Password, s.Host, s.Port, s.SID)
	//loginInfo := fmt.Sprintf("%s/%s@%s:%d/%s", s.User, s.Password, s.Host, s.Port, s.SID)

	/*
		Modifying the connection creation dsn string according to https://godror.github.io/godror/doc/connection.html
		This is runtime fix for WIORA-152
	*/
	var connectString string
	if s.Database == "Service Name" {
		connectString = fmt.Sprintf(`(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=%s)(PORT=%v))(CONNECT_DATA=(SERVICE_NAME=%s)))`, s.Host, s.Port, s.SID)
	} else {
		connectString = fmt.Sprintf(`(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=%s)(PORT=%v))(CONNECT_DATA=(SID=%s)))`, s.Host, s.Port, s.SID)
	}

	// loginInfo := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, s.User, s.Password, connectString)
	loginInfo := fmt.Sprintf(`user="%s" password="%s" connectString="%s" standaloneConnection=1`, s.User, s.Password, connectString) //go connection pooling
	db, err := sql.Open("godror", loginInfo)
	if err != nil {
		return nil, fmt.Errorf("OracleDatabase Connection Failed due to error - [%s]", err.Error())
	}
	// Parse SetConnMaxLifetime with default value of 180s
	var connMaxLifetime time.Duration
	if s.SetConnMaxLifetime != "" {
		parsedDuration, err := time.ParseDuration(s.SetConnMaxLifetime)
		if err != nil {
			return nil, fmt.Errorf("Could not parse connection lifetime duration")
		}
		connMaxLifetime = parsedDuration
	} else {
		// Set default value if SetConnMaxLifetime is empty
		connMaxLifetime = 180 * time.Second
	}

	// Apply connection pool settings
	db.SetMaxOpenConns(s.SetMaxOpenConns)
	db.SetMaxIdleConns(s.SetMaxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Oracle Database due to error - [%s]", err.Error())
	}

	sharedConn.db = db
	return sharedConn, nil
}

// OracleDatabaseSharedConfigManager structure
type OracleDatabaseSharedConfigManager struct {
	db *sql.DB
}

// Type method of connection.Manager must be implemented by OracleDatabaseSharedConfigManager
func (o *OracleDatabaseSharedConfigManager) Type() string {
	return "OracleDatabase"
}

// GetConnection method of connection.Manager must be implemented by OracleDatabaseSharedConfigManager
func (o *OracleDatabaseSharedConfigManager) GetConnection() interface{} {
	return o.db
}

// ReleaseConnection method of connection.Manager must be implemented by OracleDatabaseSharedConfigManager
func (o *OracleDatabaseSharedConfigManager) ReleaseConnection(connection interface{}) {
}

// GetSharedConfiguration returns connection.Manager based on connection selected
func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {
	cManager, err := coerce.ToConnection(conn)
	if err != nil {
		return nil, err
	}

	return cManager, nil
}

// Start method would have business logic to start the shared resource. Since db is already initialized returning nil.
func (o *OracleDatabaseSharedConfigManager) Start() error {
	return nil
}

// Stop method would do business logic to stop the the shared resource. Closing db connection in this method.
func (o *OracleDatabaseSharedConfigManager) Stop() error {
	logCache.Info("Closing Oracle Database Connection")
	if o.db != nil {
		err := o.db.Close()
		return err
	}
	return nil
}
