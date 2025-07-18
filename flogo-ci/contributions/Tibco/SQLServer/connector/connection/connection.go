package connection

import (
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

var logCache = log.ChildLogger(log.RootLogger(), "sqlserver.connection")
var factory = &Factory{}

// Settings for sqlserver
type Settings struct {
	Name                  string `md:"name,required"`
	Description           string `md:"description"`
	Host                  string `md:"host,required"`
	Port                  int    `md:"port,required"`
	User                  string `md:"user,required"`
	Password              string `md:"password,required"`
	DatabaseName          string `md:"databaseName,required"`
	Onprem                bool   `md:"onprem,required"`
	MaxOpenConnection     int    `md:"maxOpenConnection"`
	MaxIdleConnection     int    `md:"maxIdleConnection"`
	ConnectionMaxLifetime string `md:"connectionMaxLifetime"`
	ConnectionTimeout     int    `md:"connectiontimeout"`
	MaxConnRetryAttempts  int    `md:"maxconnectattempts"`
	ConnRetryDelay        int    `md:"connectionretrydelay"`
	TLSParam              bool   `md:"tlsparam"`
	ValidateServerCert    bool   `md:"validatecert"`
	Cacert                string `md:"cacert"`
}

func init() {
	if os.Getenv(log.EnvKeyLogLevel) == "DEBUG" {
		logCache.DebugEnabled()
	}

	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// Factory for sqlserver connection
type Factory struct {
}

// Type Factory
func (*Factory) Type() string {
	return "SQLServer"
}

//decodeTLSParam translate tls parm to connection string
// func decodeTLSParam(tlsparm string) string {
// 	switch tlsparm {
// 	case "Client Initiated":
// 		return "encrypt=true;trustServerCertificate=false"
// 	case "Server Initiated":
// 		return "encrypt=false;trustServerCertificate=false"
// 	default:
// 		return ""
// 	}
// }

// copyCertToTempFile creates temp mssql.pem file for running app in container
// and sqlserver needs filepath for ssl cert so can not pass byte array which we get from connection tile
// TO DO: remove this file once used , yet to be decided
func copyCertToTempFile(certdata []byte, connctionName string) (string, error) {
	var path = connctionName + "_" + "mssql.pem"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return "", fmt.Errorf("Could not create file mssql.pem %s", err.Error())
		}
		defer file.Close()
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("Could not open file mssql.pem %s", err.Error())
	}
	_, err = file.Write(certdata)
	if err != nil {
		return "", fmt.Errorf("Could not write data to file mssql.pem %s", err.Error())
	}
	return path, nil
}

// NewManager for SQLServer
func (*Factory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &SharedConfigManager{}
	var err error
	setting := &Settings{}
	err = metadata.MapToStruct(settings, setting, false)

	if err != nil {
		return nil, err
	}

	cName := setting.Name
	if cName == "" {
		return nil, errors.New("Required Parameter Name is missing")
	}
	cHost := setting.Host
	if cHost == "" {
		return nil, errors.New("Required Parameter Host Name is missing")
	}

	cPort := setting.Port
	if cPort == 0 {
		return nil, errors.New("Required Parameter Port is missing")
	}
	cDbName := setting.DatabaseName
	if cDbName == "" {
		return nil, errors.New("Required Parameter Database name is missing")
	}
	cUser := setting.User
	if cUser == "" {
		return nil, errors.New("Required Parameter User is missing")
	}
	cPassword := setting.Password
	if cPassword == "" {
		return nil, errors.New("Required Parameter Password is missing")
	}
	cOnprem := setting.Onprem
	cOnpremstr := strconv.FormatBool(cOnprem)

	logCache.Debug("Onprem selected?", cOnpremstr)

	cMaxOpenConnection := setting.MaxOpenConnection
	if cMaxOpenConnection == 0 {
		logCache.Debug("Default value of Max open connection chosen. Default is 0")
	}
	if cMaxOpenConnection < 0 {
		logCache.Debugf("Max open connection value received is %d, it will be defaulted to 0", cMaxOpenConnection)
	}
	cMaxIdleConnection := setting.MaxIdleConnection
	if cMaxIdleConnection == 2 {
		logCache.Debug("Default value of Max idle connection chosen. Default is 2")
	}
	if cMaxIdleConnection < 0 {
		logCache.Debugf("Max idle connection value received is %d, it will be defaulted to 0", cMaxIdleConnection)
	}
	cConnectionMaxLifetime := setting.ConnectionMaxLifetime
	if cConnectionMaxLifetime == "0" {
		logCache.Debug("Default value of Max lifetime of connection chosen. Default is 0")
	}
	if strings.HasPrefix(cConnectionMaxLifetime, "-") {
		logCache.Debugf("Max lifetime connection value received is %s, it will be defaulted to 0", cConnectionMaxLifetime)
	}

	var ConnectionMaxLifetimeDuration time.Duration
	if cConnectionMaxLifetime != "" {
		ConnectionMaxLifetimeDuration, err = time.ParseDuration(cConnectionMaxLifetime)
		if err != nil {
			return nil, fmt.Errorf("Could not parse connection lifetime duration")
		}
	}

	cMaxConnRetryAttempts := setting.MaxConnRetryAttempts
	if cMaxConnRetryAttempts == 0 {
		logCache.Debug("Maximum connection retry attempt is 0, no retry attempts will be made")
	}
	if cMaxConnRetryAttempts < 0 {
		logCache.Debugf("Max connection retry attempts received is %d", cMaxConnRetryAttempts)
		return nil, fmt.Errorf("Max connection retry attempts cannot be a negative number")
	}
	cConnRetryDelay := setting.ConnRetryDelay
	if cConnRetryDelay == 0 {
		logCache.Debug("Connection Retry Delay is 0, no delays between the retry attempts")
	}
	if cConnRetryDelay < 0 {
		logCache.Debugf("Connection retry delay value received is %d", cConnRetryDelay)
		return nil, fmt.Errorf("Connection retry delay cannot be a negative number")
	}

	cConnTimeout := setting.ConnectionTimeout
	if cConnTimeout < 0 {
		logCache.Debugf("Connection Timeout received is %d", cConnTimeout)
		return nil, fmt.Errorf("Connection timeout cannot be a negative number")
	}

	cTLSMode := setting.TLSParam
	var conninfo string
	if cTLSMode == false {
		logCache.Debugf("Login attempting plain connection")
		conninfo = fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;sslmode=disable;dial timeout=%d;",
			cHost, cPort, cUser, cPassword, cDbName, cConnTimeout)

	} else {
		logCache.Debugf("Login attempting SSL connection")
		cValidateServerCert := setting.ValidateServerCert
		// extract byte array of cert data for file content string
		cCAcert, _ := getByteCertDataForPemFile(setting.Cacert, "ca")
		//logCache.Info("CaCert as Bytes:", cCAcert)
		conninfo = fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;%s;dial timeout=%d;", cHost, cPort, cUser, cPassword, cDbName, validateServerCert(cValidateServerCert), cConnTimeout)

		//create temp file
		if cCAcert != nil {
			cCAcertPath, err := copyCertToTempFile(cCAcert, setting.Name)
			if err == nil {
				sharedConn.CertFilePath = cCAcertPath
				conninfo = conninfo + fmt.Sprintf("certificate=%s;", cCAcertPath)
			} else {
				return nil, fmt.Errorf("Error : %s", err.Error())
			}
		}

	}

	// add connection delay
	// check for bad err connecton and then only do retry, dont do it for invalid creds types of errors
	// move the code outside of the for loop for attempt=0 since we are treating 0 as no reattempts
	// do sql.open outside of loop and then check for the errors based on error type and if reattempts!=0 then do retry
	//
	var db *sql.DB
	dbConnected := 0
	if cMaxConnRetryAttempts == 0 {
		logCache.Info("No connection retry selected, connection attempt will be tried only once...")
		db, err = sql.Open("mssql", conninfo)
		if err != nil {
			return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
		} else {
			err = db.Ping()
			if err != nil {
				return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
			}
			dbConnected = 1
		}
	} else if cMaxConnRetryAttempts > 0 {
		logCache.Debugf("Maximum connection retry attempts allowed - %d", cMaxConnRetryAttempts)
		logCache.Debugf("Connection retry delay - %d", cConnRetryDelay)
		db, err = sql.Open("mssql", conninfo)
		if err != nil {
			// return nil, dont retry
			return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
		}
		// sql returned db handle in first attempt
		if db != nil && err == nil {
			logCache.Info("Trying to ping the database server...")
			err = db.Ping()
			// retry attempt on ping only for conn refused and driver bad conn
			if err != nil {
				lowerErr := strings.ToLower(err.Error())
				if err == driver.ErrBadConn || strings.Contains(lowerErr, "connection refused") || strings.Contains(lowerErr, "network is unreachable") ||
					strings.Contains(lowerErr, "connection reset by peer") || strings.Contains(lowerErr, "dial tcp: lookup") ||
					strings.Contains(lowerErr, "connection timed out") || strings.Contains(lowerErr, "timedout") || strings.Contains(lowerErr, "time out") ||
					strings.Contains(lowerErr, "timed out") || strings.Contains(lowerErr, "net.Error") || strings.Contains(lowerErr, "i/o timeout") ||
					strings.Contains(lowerErr, "no such host") || strings.Contains(lowerErr, fmt.Sprintf("dial tcp %s:%d: i/o timeout", cHost, cPort)) ||
					strings.Contains(lowerErr, "broken pipe") {

					logCache.Info("Failed to ping the database server, trying again...")
					for i := 1; i <= cMaxConnRetryAttempts; i++ {
						logCache.Infof("Connecting to database server... Attempt-[%d]", i)
						// retry delay
						time.Sleep(time.Duration(cConnRetryDelay) * time.Second)
						err = db.Ping()
						if err != nil {
							lowerErr := strings.ToLower(err.Error())
							if err == driver.ErrBadConn || strings.Contains(lowerErr, "connection refused") || strings.Contains(lowerErr, "network is unreachable") ||
								strings.Contains(lowerErr, "connection reset by peer") || strings.Contains(lowerErr, "dial tcp: lookup") ||
								strings.Contains(lowerErr, "connection timed out") || strings.Contains(lowerErr, "timedout") || strings.Contains(lowerErr, "time out") ||
								strings.Contains(lowerErr, "timed out") || strings.Contains(lowerErr, "net.Error") || strings.Contains(lowerErr, "i/o timeout") ||
								strings.Contains(lowerErr, "no such host") || strings.Contains(lowerErr, fmt.Sprintf("dial tcp %s:%d: i/o timeout", cHost, cPort)) ||
								strings.Contains(lowerErr, "broken pipe") {
								continue
							} else {
								return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
							}
						} else {
							// ping succesful
							dbConnected = 1
							logCache.Infof("Successfully connected to database server in attempt-[%d]", i)
							break
						}
					}
					if dbConnected == 0 {
						logCache.Errorf("Could not connect to database server even after %d number of attempts", cMaxConnRetryAttempts)
						return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
					}
				} else {
					return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
				}
			}
			if dbConnected != 0 {
				logCache.Info("Successfully connected to database server")
			}
		}
	}

	sharedConn.db = db
	sharedConn.db.SetMaxOpenConns(cMaxOpenConnection)
	sharedConn.db.SetMaxIdleConns(cMaxIdleConnection)
	sharedConn.db.SetConnMaxLifetime(ConnectionMaxLifetimeDuration)

	logCache.Debug("----------- DB Stats after setting extra connection configs -----------")
	logCache.Debug("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	logCache.Debug("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	logCache.Debug("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	logCache.Debug("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	logCache.Debug("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	logCache.Debug("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return sharedConn, nil
}

func validateServerCert(validateServerCert bool) string {
	switch validateServerCert {
	case false:
		return "encrypt=true;trustServerCertificate=true;"
	case true:
		return "encrypt=true;trustServerCertificate=false;"
	default:
		return ""
	}
}

// Get cacert file from connection
// Get cacert file from connection
func getByteCertDataForPemFile(cert string, certType string) ([]byte, error) {
	if cert == "" {
		return nil, fmt.Errorf("%s Certificate is empty", strings.ToUpper(certType))
	}
	//logCache.Info("Cert from Settings: ", cert)
	// if cert file is chosen in design time
	if strings.HasPrefix(cert, "{") {
		logCache.Debugf("%s Certificate received from fileselector", strings.ToUpper(certType))
		certObj, err := coerce.ToObject(cert)
		if err == nil {
			certRealValue, ok := certObj["content"].(string)
			logCache.Debugf("Fetched content from %s Certificate", strings.ToUpper(certType))
			if !ok || certRealValue == "" {
				return nil, fmt.Errorf("%s Certificate content not found", strings.ToUpper(certType))
			}

			index := strings.IndexAny(certRealValue, ",")
			if index > -1 {
				certRealValue = certRealValue[index+1:]
			}
			return base64.StdEncoding.DecodeString(certRealValue)
		}
		return nil, err
	}

	// if the certificate comes from application properties need to check whether it is base64 encoded
	certRealValue, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, fmt.Errorf("Error in parsing %s certificates or we may be not be supporting the given encoding", strings.ToUpper(certType))
	}
	logCache.Debugf("%s Certificate successfully decoded", strings.ToUpper(certType))
	return certRealValue, nil
}

// SharedConfigManager details
type SharedConfigManager struct {
	name         string
	db           *sql.DB
	connKey      string
	CertFilePath string
}

// Type SharedConfigManager details
func (sharedConn *SharedConfigManager) Type() string {
	return "SQLServer"
}

// GetConnection SharedConfigManager details
func (sharedConn *SharedConfigManager) GetConnection() interface{} {
	return sharedConn.db
}

// ReleaseConnection SharedConfigManager details
func (sharedConn *SharedConfigManager) ReleaseConnection(connection interface{}) {

}

// Start SharedConfigManager details
func (sharedConn *SharedConfigManager) Start() error {
	return nil
}

// Stop SharedConfigManager details
func (sharedConn *SharedConfigManager) Stop() error {
	logCache.Debugf("Cleaning up DB")
	sharedConn.db.Close()

	// clean up pem files on stop
	if sharedConn.CertFilePath != "" {
		err := os.Remove(sharedConn.CertFilePath)
		if err != nil {
			return fmt.Errorf("Error while removing pem file %s", err.Error())
		}
	}
	return nil
}
