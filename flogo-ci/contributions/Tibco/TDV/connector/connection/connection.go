package connection

import (
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/alexbrainman/odbc"
	"github.com/alexbrainman/odbc/api"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	cgoConn "github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/connection"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/execute"
)

var logCache = log.ChildLogger(log.RootLogger(), "tdv.connection")

var factory = &TDVFactory{}

// var db *sql.DB

// Settings for tdv
type Settings struct {
	Name        string `md:"name,required"`
	Description string `md:"description"`
	DatabaseURL string `md:"databaseURL"`
	Server      string `md:"server,required"`
	Port        int    `md:"port,required"`
	User        string `md:"user,required"`
	Password    string `md:"password,required"`
	Onprem      bool   `md:"onprem,required"`
	DataSource  string `md:"datasource,required"`
	Domain      string `md:"domain,required"`
	SSLMode     string `md:"sslmode"`
	// Catalog     string `md:"catalog"`
	// MaxOpenConnections   int    `md:"maxopenconnection"`
	// MaxIdleConnections   int    `md:"maxidleconnection"`
	// MaxConnLifetime      string `md:"connmaxlifetime"`
	MaxConnRetryAttempts int    `md:"maxconnectattempts"`
	ConnRetryDelay       int    `md:"connectionretrydelay"`
	ConnectionTimeout    int    `md:"connectiontimeout"`
	SessionTimeout       int    `md:"sessionTimeout"`
	RequestTimeout       int    `md:"requestTimeout"`
	TLSConfig            bool   `md:"tlsconfig"`
	TLSMode              string `md:"tlsparam"`
	Cacert               string `md:"cacert"`
	//	Clientcert           string `md:"clientcert"`
	//	Clientkey            string `md:"clientkey"`

	// Value interface{} `md:"value"`

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

type TDVFactory struct {
}

// Type TDVFactory
func (*TDVFactory) Type() string {
	return "TDV"
}

// TDVSharedConfigManager details
type TDVSharedConfigManager struct {
	mu                   *sync.RWMutex
	name                 string
	db                   *sql.DB
	CACertFilePath       string
	cgoConnectionID      string
	CgoConnection        *cgoConn.Conn
	PingStatmentHandle   interface{} `json:"-"`
	maxConnRetryAttempts int
	dataSource           string
	conninfo             string
	connRetryDelay       int

	// ClientCertFilePath string
	// ClientKeyFilePath  string
}

func findDriver(v string) string {
	odbcDriverPath := filepath.Join(v, "libcomposite85_x64.so")
	_, err := os.Stat(odbcDriverPath)
	if err == nil {
		return odbcDriverPath
	}
	return ""
}

// NewManager TDVFactory
func (*TDVFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &TDVSharedConfigManager{}
	var err error

	s := &Settings{}
	err = metadata.MapToStruct(settings, s, false)

	if err != nil {
		return nil, err
	}
	cServer := s.Server
	if cServer == "" {
		return nil, errors.New("Required Parameter Server Name is missing")
	}
	cPort := s.Port
	if cPort == 0 {
		return nil, errors.New("Required Parameter Port is missing")
	}
	cDataSource := s.DataSource
	if cDataSource == "" {
		return nil, errors.New("Required Parameter DataSource name is missing")
	}
	cDomain := s.Domain
	if cDomain == "" {
		return nil, errors.New("Required Parameter Domain name is missing")
	}
	cUser := s.User
	if cUser == "" {
		return nil, errors.New("Required Parameter User is missing")
	}
	cPassword := s.Password
	if cPassword == "" {
		return nil, errors.New("Required Parameter Password is missing")
	}
	cOnprem := s.Onprem
	cOnpremstr := strconv.FormatBool(cOnprem)

	logCache.Debug("Onprem selected?", cOnpremstr)
	cMaxConnRetryAttempts := s.MaxConnRetryAttempts
	if cMaxConnRetryAttempts == 0 {
		logCache.Debug("Maximum connection retry attempt is 0, no retry attempts will be made")
	}
	if cMaxConnRetryAttempts < 0 {
		logCache.Debugf("Max connection retry attempts received is %d", cMaxConnRetryAttempts)
		return nil, fmt.Errorf("max connection retry attempts cannot be a negative number")
	}
	cConnRetryDelay := s.ConnRetryDelay
	if cConnRetryDelay == 0 {
		logCache.Debug("Connection Retry Delay is 0, no delays between the retry attempts")
	}
	if cConnRetryDelay < 0 {
		logCache.Debugf("Connection retry delay value received is %d", cConnRetryDelay)
		return nil, fmt.Errorf("Connection retry delay cannot be a negative number")
	}

	cConnTimeout := s.ConnectionTimeout
	if cConnTimeout < 0 {
		logCache.Debugf("Connection Timeout received is %d", cConnTimeout)
		return nil, fmt.Errorf("Connection timeout cannot be a negative number")
	}
	cSessionTimeout := s.SessionTimeout
	if cSessionTimeout < 0 {
		logCache.Debugf("Session Timeout received is %d", cSessionTimeout)
		return nil, fmt.Errorf("session timeout cannot be a negative number")
	}
	cRequestTimeout := s.RequestTimeout
	if cRequestTimeout < 0 {
		logCache.Debugf("Request Timeout received is %d", cRequestTimeout)
		return nil, fmt.Errorf("request timeout cannot be a negative number")
	}
	//basic Connection string without Driver
	conninfo := fmt.Sprintf("server=%s;port=%d;uid=%s;pwd=%s;datasource=%s;domain=%s;connectTimeout=%d;sessionTimeout=%d;requestTimeout=%d;", cServer, cPort, cUser, cPassword, cDataSource, cDomain, cConnTimeout, cSessionTimeout, cRequestTimeout)
	var odbcDriverPath string
	pwd := os.Getenv("PWD")

	odbcDriverPath = filepath.Join(pwd, "supplement", "TDV", "libcomposite85_x64.so")
	_, err = os.Stat(odbcDriverPath)
	if err != nil {
		odbcDriverPath = ""
		// ldPath, ok := os.LookupEnv("LD_LIBRARY_PATH")
		// if !ok {
		// 	return nil, fmt.Errorf("cannot find libcomposite85_x64.so. LD_LIBRARY_PATH not set")
		// }
		// ldPath = strings.Trim(ldPath, ":")
		// ldPathArr := strings.Split(ldPath, ":")
		// for _, v := range ldPathArr {
		// 	odbcDriverPath = findDriver(v)
		// 	if odbcDriverPath != "" {
		// 		break
		// 	}
		// }
		// if odbcDriverPath == "" {
		// 	return nil, fmt.Errorf("cannot find libcomposite85_x64.so in LD_LIBRARY_PATH")
		// }
		odbcDriverPath, err = getDriverName()
		if err != nil {
			logCache.Debug("Not found any driver name on system")
			return nil, err
		}
		logCache.Debug("Odbc Driver is at ", odbcDriverPath)
	}
	conninfo = conninfo + fmt.Sprintf("driver=%s;", odbcDriverPath)

	cTLSConfig := s.TLSConfig
	if cTLSConfig == false {
		logCache.Debugf("Login attempting plain connection")
		conninfo = conninfo + "encrypt=false;"

	} else {
		logCache.Debugf("Login attempting SSL connection")
		// 	cTLSMode := s.TLSMode
		// 	 extract byte array of cert data for file content string
		cCAcert, _ := getByteCertDataForPemFile(s.Cacert, "ca")
		// 	cClientcert, _ := getByteCertDataForPemFile(s.Clientcert, "client")
		// 	cClientKey, _ := getByteCertDataForPemFile(s.Clientkey, "clientkey")
		conninfo = conninfo + "encrypt=true;validateRemoteCert=true;"
		//conninfo = fmt.Sprintf("driver=%s;server=%s;port=%d;uid=%s;pwd=%s;datasource=%s;domain=%s;connectTimeout=%d;encrypt=true;validateRemoteCert=true;", odbcDriverPath, cServer, cPort, cUser, cPassword, cDataSource, cDomain, cConnTimeout)
		//create temp file
		if cCAcert != nil {
			cCAcertPath, err := copyCertToTempFile(cCAcert, s.Name+"_CA")
			if err == nil {
				// Currentpwd := os.Getenv("PWD")
				// logCache.Info("Present working directory is ", Currentpwd)
				// cCAcertPath = filepath.Join(temp, cCAcertPath)
				//To store The file Path in sharedConfManager to delete it on stop() method
				sharedConn.CACertFilePath = cCAcertPath
				conninfo = conninfo + fmt.Sprintf("sslCACert=%s;", cCAcertPath)
			} else {
				return nil, fmt.Errorf("Error wh: %s", err.Error())
			}
		}

	}
	sharedConn.CgoConnection = nil
	sharedConn.db = nil
	sharedConn.connRetryDelay = cConnRetryDelay
	sharedConn.dataSource = cDataSource
	sharedConn.maxConnRetryAttempts = cMaxConnRetryAttempts
	sharedConn.conninfo = conninfo

	sharedConn.mu = &sync.RWMutex{}

	return sharedConn, nil
}

// Type TDVSharedConfigManager details
func (p *TDVSharedConfigManager) Type() string {

	return "TDV"
}

// GetConnection TDVSharedConfigManager details
func (p *TDVSharedConfigManager) GetConnection() interface{} {
	return p
}

// GetSQLConnection TDVSharedConfigManager DB SQL connection
func (p *TDVSharedConfigManager) GetSQLConnection() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db
}

// IsSQLConnectionNil return if connection is nil
func (p *TDVSharedConfigManager) IsSQLConnectionNil() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db == nil
}

// GetCgoConnection TDVSharedConfigManager CGO connection pointer
func (p *TDVSharedConfigManager) GetCgoConnection() *cgoConn.Conn {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.CgoConnection
}

// SetSQLConnection Sets Connection using Db/SQL
func (p *TDVSharedConfigManager) SetSQLConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// add connection delay
	// check for bad err connecton and then only do retry, dont do it for invalid creds types of errors
	// move the code outside of the for loop for attempt=0 since we are treating 0 as no reattempts
	// do sql.open outside of loop and then check for the errors based on error type and if reattempts!=0 then do retry

	var db *sql.DB
	var err error
	dbConnected := 0
	if p.maxConnRetryAttempts == 0 {
		logCache.Info("No connection retry selected, connection attempt will be tried only once...")
		db, err = sql.Open("odbc", p.conninfo)
		if err != nil {
			return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
		} else {
			err = db.Ping()
			if err != nil {
				return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
			}
			dbConnected = 1
		}
	} else if p.maxConnRetryAttempts > 0 {
		logCache.Debugf("Maximum connection retry attempts allowed - %d", p.maxConnRetryAttempts)
		logCache.Debugf("Connection retry delay - %d", p.connRetryDelay)
		db, err = sql.Open("odbc", p.conninfo)
		if err != nil {
			// return nil, dont retry
			return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
		}
		// sql returned db handle in first attempt
		if db != nil && err == nil {
			logCache.Info("Trying to ping the database server...")
			err = db.Ping()
			logCache.Debug("Error after ping the database server : ", err)
			// retry attempt on ping only for conn refused and driver bad conn
			// When host is DNS for SSL conn and wifi is off, retry fails due to 'tcp: lookup' error. Works fine when host is IP Address
			if err != nil {
				if err == driver.ErrBadConn || strings.Contains(strings.ToLower(strings.ToLower(err.Error())), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
					strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") || strings.Contains(strings.ToLower(err.Error()), "dial tcp: lookup") ||
					strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "timedout") || strings.Contains(strings.ToLower(err.Error()), "connection is closed") ||
					strings.Contains(strings.ToLower(err.Error()), "request timed out") || strings.Contains(strings.ToLower(err.Error()), "timed out") || strings.Contains(strings.ToLower(err.Error()), "net.Error") || strings.Contains(strings.ToLower(err.Error()), "i/o timeout") {
					logCache.Info("Failed to ping the database server, trying again...")
					for i := 1; i <= p.maxConnRetryAttempts; i++ {
						logCache.Infof("Connecting to database server... Attempt-[%d]", i)
						// retry delay
						time.Sleep(time.Duration(p.connRetryDelay) * time.Second)
						logCache.Info("Trying to ping the database server...")
						err = db.Ping()
						logCache.Debug("Error after ping the database server : ", err)
						if err != nil {
							if err == driver.ErrBadConn || strings.Contains(strings.ToLower(strings.ToLower(err.Error())), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
								strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") || strings.Contains(strings.ToLower(err.Error()), "dial tcp: lookup") ||
								strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "timedout") || strings.Contains(strings.ToLower(err.Error()), "connection is closed") ||
								strings.Contains(strings.ToLower(err.Error()), "request timed out") || strings.Contains(strings.ToLower(err.Error()), "timed out") || strings.Contains(strings.ToLower(err.Error()), "net.Error") || strings.Contains(strings.ToLower(err.Error()), "i/o timeout") {
								continue
							} else {
								return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
							}
						} else {
							// ping succesful
							dbConnected = 1
							logCache.Infof("Successfully connected to database server in attempt-[%d]", i)
							break
						}
					}
					if dbConnected == 0 {
						logCache.Errorf("Could not connect to database server even after %d number of attempts", p.maxConnRetryAttempts)
						return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
					}
				} else {
					return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
				}
			}
			if dbConnected != 0 {
				logCache.Info("Successfully connected to database server")
			}
		}
	}

	p.db = db
	// sharedConn.db.SetConnMaxLifetime(lifetimeDuration)

	logCache.Debug("----------- DB Stats after setting extra connection configs -----------")
	logCache.Debug("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	logCache.Debug("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	logCache.Debug("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	logCache.Debug("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	logCache.Debug("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))

	err = p.db.Ping()
	if err != nil {
		return err
	}
	return nil
}

// Check If SQL Connection is alive
func (p *TDVSharedConfigManager) IsSQLConnectionAlive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.db.Ping()
	if err != nil {
		logCache.Info("Connection is not alive : ", err)
		p.db = nil
		return false
	}
	return true
}

// Check If SQL Connection is alive
func (p *TDVSharedConfigManager) IsCGOConnectionAlive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var err error
	err = cgoConn.IsAlive(p.CgoConnection.H)
	if err != nil {
		logCache.Info("Connection is not alive : ", err)
		return false
	}
	err = cgoConn.CustomPing(p.PingStatmentHandle.(api.SQLHSTMT), "SELECT 1;")
	if err != nil {
		logCache.Info("Connection is not alive : ", err)
		return false
	}
	return true
}

// SetCgoConnection Sets Connection using Direct CGO Layer and store it in connectionHandle
func (p *TDVSharedConfigManager) SetCgoConnection() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var err error
	var cgoConnection *cgoConn.Conn
	var cgoConnectionID string

	dbConnected := 0
	if p.maxConnRetryAttempts == 0 {
		logCache.Info("No connection retry selected, connection attempt will be tried only once...")
		cgoConnection, cgoConnectionID, err = cgoConn.ConnectDB(p.conninfo, p.name)
		if err != nil {
			return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
		}
		dbConnected = 1
	} else if p.maxConnRetryAttempts > 0 {
		logCache.Debugf("Maximum connection retry attempts allowed - %d", p.maxConnRetryAttempts)
		logCache.Debugf("Connection retry delay - %d", p.connRetryDelay)

		logCache.Info("Trying to Connect the database server...")
		cgoConnection, cgoConnectionID, err = cgoConn.ConnectDB(p.conninfo, p.name)
		logCache.Debug("Error after Trying to connect the database server : ", err)
		// retry attempt on ping only for conn refused and driver bad conn
		// When host is DNS for SSL conn and wifi is off, retry fails due to 'tcp: lookup' error. Works fine when host is IP Address
		if err != nil {
			if strings.Contains(strings.ToLower(strings.ToLower(err.Error())), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
				strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") || strings.Contains(strings.ToLower(err.Error()), "dial tcp: lookup") ||
				strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "timedout") ||
				strings.Contains(strings.ToLower(err.Error()), "timed out") || strings.Contains(strings.ToLower(err.Error()), "net.Error") || strings.Contains(strings.ToLower(err.Error()), "connection is closed") ||
				strings.Contains(strings.ToLower(err.Error()), "i/o timeout") {
				logCache.Info("Failed to Connect the database server, trying again...")
				for i := 1; i <= p.maxConnRetryAttempts; i++ {
					logCache.Infof("Connecting to database server... Attempt-[%d]", i)
					// retry delay
					time.Sleep(time.Duration(p.connRetryDelay) * time.Second)
					cgoConnection, cgoConnectionID, err = cgoConn.ConnectDB(p.conninfo, p.name)
					logCache.Debug("Error after Trying to connect the database server : ", err)
					if err != nil {
						if strings.Contains(strings.ToLower(strings.ToLower(err.Error())), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
							strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") || strings.Contains(strings.ToLower(err.Error()), "dial tcp: lookup") ||
							strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "timedout") ||
							strings.Contains(strings.ToLower(err.Error()), "timed out") || strings.Contains(strings.ToLower(err.Error()), "net.Error") || strings.Contains(strings.ToLower(err.Error()), "connection is closed") ||
							strings.Contains(strings.ToLower(err.Error()), "i/o timeout") {
							continue
						} else {
							return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
						}
					} else {
						// ping succesful
						dbConnected = 1
						logCache.Infof("Successfully connected to database server in attempt-[%d]", i)
						break
					}
				}
				if dbConnected == 0 {
					logCache.Errorf("Could not connect to database server even after %d number of attempts", p.maxConnRetryAttempts)
					return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
				}
			} else {
				return fmt.Errorf("could not open connection to datasource %s, %s", p.dataSource, err.Error())
			}
		} else {
			dbConnected = 1
		}
		if dbConnected != 0 {
			logCache.Info("Successfully connected to database server")
			p.PingStatmentHandle, err = execute.GetStatementHandle(cgoConnection.H)
			if err != nil {
				logCache.Debug("Error while Getting PingStatementHandle : ", err)
			}
		}
	}
	if err != nil {
		return err
	}
	p.CgoConnection = cgoConnection
	p.cgoConnectionID = cgoConnectionID

	return nil
}

// ReleaseConnection TDVSharedConfigManager details
func (p *TDVSharedConfigManager) ReleaseConnection(connection interface{}) {

}

// ReleaseConnection TDVSharedConfigManager details
func (p *TDVSharedConfigManager) ReleaseCgoConnection() {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.CgoConnection.Disconnect()
	if err != nil {
		logCache.Debug("Error while Disconnection from Connection : ", err)
	}
	p.CgoConnection = nil
}

// Start TDVSharedConfigManager details
func (p *TDVSharedConfigManager) Start() error {
	return nil
}

// Stop TDVSharedConfigManager details
func (p *TDVSharedConfigManager) Stop() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	logCache.Debug("Cleaning up DB")
	//preparedQueryCache.Clear()

	for key := range preparedQueryCache {
		preparedQueryCache[key].Close()
	}
	if p.db != nil {
		p.db.Close()
	}

	// clean up CA cert pem files on stop

	_, err := os.Stat(p.CACertFilePath)
	if !os.IsNotExist(err) {
		err := os.Remove(p.CACertFilePath)
		if err != nil {
			return fmt.Errorf("Error while removing pem file %s", err.Error())
		}
		logCache.Debug("CA Cert File successfully deleted")
	}
	if p.CgoConnection != nil {
		p.ReleaseCgoConnection()
	}
	return nil
}

// copyCertToTempFile creates temp mssql.pem file for running app in container
// and sqlserver needs filepath for ssl cert so can not pass byte array which we get from connection tile
func copyCertToTempFile(certdata []byte, name string) (string, error) {
	var path = name + "_" + "tdv.pem"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return "", fmt.Errorf("could not create file %s %s ", path, err.Error())
		}
		if err := os.Chmod(path, 0600); err != nil {
			return "", fmt.Errorf("could not give permissions file %s %s ", path, err.Error())
		}
		defer file.Close()
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return "", fmt.Errorf("could not open file %s %s", path, err.Error())
	}
	defer file.Close()
	_, err = file.Write(certdata)
	if err != nil {
		return "", fmt.Errorf("could not write data to file %s %s", path, err.Error())
	}
	return path, nil
}

// Get cacert file from connection
func getByteCertDataForPemFile(cert string, certType string) ([]byte, error) {
	//case when you provide file having less permission on Ui but as onUI there is no check as per now for cert and key
	// we need to handle it

	if cert == "" {
		return nil, fmt.Errorf("%s Certificate is empty", strings.ToUpper(certType))
	}

	// if cert file is chosen in design time
	if strings.HasPrefix(cert, "{") {
		logCache.Debugf("%s Certificate received from fileselector", strings.ToUpper(certType))
		certObj, err := coerce.ToObject(cert)
		if err == nil {
			certRealValue, ok := certObj["content"].(string)
			logCache.Debugf("Fetched content from %s Certificate", strings.ToUpper(certType))
			//if content is nil and filename is nil them it confirms that cert file is not passed
			//We should not only check for content because a cert file without cert data is considered as incorrect cert
			if (!ok || certRealValue == "") && certObj["file"] == nil {
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
	// if the certificate comes from application properties it is decoded and passed as it is
	//because user might send non base64 encoded data which should be considered as incorrect
	//and passed to the server
	certRealValue, err := base64.StdEncoding.DecodeString(cert)
	return certRealValue, err
}

func decodeTLSParam(tlsparm string) string {
	switch tlsparm {
	case "VerifyCA":
		return "verify-ca"
	case "VerifyFull":
		return "verify-full"
	default:
		return ""
	}
}
