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

var logCache = log.ChildLogger(log.RootLogger(), "postgres.connection")
var factory = &PgFactory{}

// var db *sql.DB

// Settings for postgres
type Settings struct {
	DatabaseType         string `md:"databaseType,required"`
	Name                 string `md:"name,required"`
	Description          string `md:"description"`
	DatabaseURL          string `md:"databaseURL"`
	Host                 string `md:"host,required"`
	Port                 int    `md:"port,required"`
	User                 string `md:"user,required"`
	Password             string `md:"password,required"`
	Onprem               bool   `md:"onprem,required"`
	DbName               string `md:"databaseName,required"`
	SSLMode              string `md:"sslmode"`
	MaxOpenConnections   int    `md:"maxopenconnection"`
	MaxIdleConnections   int    `md:"maxidleconnection"`
	MaxConnLifetime      string `md:"connmaxlifetime"`
	MaxConnRetryAttempts int    `md:"maxconnectattempts"`
	ConnRetryDelay       int    `md:"connectionretrydelay"`
	ConnectionTimeout    int    `md:"connectiontimeout"`
	TLSConfig            bool   `md:"tlsconfig"`
	TLSMode              string `md:"tlsparam"`
	Cacert               string `md:"cacert"`
	Clientcert           string `md:"clientcert"`
	Clientkey            string `md:"clientkey"`
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

// PgFactory for postgres connection
type PgFactory struct {
}

// Type PgFactory
func (*PgFactory) Type() string {
	return "Postgres"
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

// NewManager PgFactory
func (*PgFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &PgSharedConfigManager{}
	var err error

	s := &Settings{}
	err = metadata.MapToStruct(settings, s, false)

	if err != nil {
		return nil, err
	}
	cHost := s.Host
	if cHost == "" {
		return nil, errors.New("Required Parameter Host Name is missing")
	}
	cPort := s.Port
	if cPort == 0 {
		return nil, errors.New("Required Parameter Port is missing")
	}
	cDbName := s.DbName
	if cDbName == "" {
		return nil, errors.New("Required Parameter Database name is missing")
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

	cMaxOpenConn := s.MaxOpenConnections
	if cMaxOpenConn == 0 {
		logCache.Debug("Default value of Max open connection chosen. Default is 0")
	}
	if cMaxOpenConn < 0 {
		logCache.Debugf("Max open connection value received is %d, it will be defaulted to 0", cMaxOpenConn)
	}
	cMaxIdleConn := s.MaxIdleConnections
	if cMaxIdleConn == 2 {
		logCache.Debug("Default value of Max idle connection chosen. Default is 2")
	}
	if cMaxIdleConn < 0 {
		logCache.Debugf("Max idle connection value received is %d, it will be defaulted to 0", cMaxIdleConn)
	}
	cMaxConnLifetime := s.MaxConnLifetime
	if cMaxConnLifetime == "0" {
		logCache.Debug("Default value of Max lifetime of connection chosen. Default is 0")
	}
	if strings.HasPrefix(cMaxConnLifetime, "-") {
		logCache.Debugf("Max lifetime connection value received is %s, it will be defaulted to 0", cMaxConnLifetime)
	}

	var lifetimeDuration time.Duration
	if cMaxConnLifetime != "" {
		lifetimeDuration, err = time.ParseDuration(cMaxConnLifetime)
		if err != nil {
			return nil, fmt.Errorf("Could not parse connection lifetime duration")
		}
	}

	cMaxConnRetryAttempts := s.MaxConnRetryAttempts
	if cMaxConnRetryAttempts == 0 {
		logCache.Debug("Maximum connection retry attempt is 0, no retry attempts will be made")
	}
	if cMaxConnRetryAttempts < 0 {
		logCache.Debugf("Max connection retry attempts received is %d", cMaxConnRetryAttempts)
		return nil, fmt.Errorf("Max connection retry attempts cannot be a negative number")
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

	cTLSConfig := s.TLSConfig
	var conninfo string
	if cTLSConfig == false {
		logCache.Debugf("Login attempting plain connection")
		conninfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d ", cHost, cPort, cUser, cPassword, cDbName, cConnTimeout)
	} else {
		logCache.Debugf("Login attempting SSL connection")
		cTLSMode := s.TLSMode
		// extract byte array of cert data for file content string
		cCAcert, _ := getByteCertDataForPemFile(s.Cacert, "ca")
		cClientcert, _ := getByteCertDataForPemFile(s.Clientcert, "client")
		cClientKey, _ := getByteCertDataForPemFile(s.Clientkey, "clientkey")

		conninfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d ", cHost, cPort, cUser, cPassword, cDbName, decodeTLSParam(cTLSMode), cConnTimeout)
		//create temp file
		if cCAcert != nil {
			cCAcertPath, err := copyCertToTempFile(cCAcert, s.Name+"_CA")
			if err == nil {
				sharedConn.CACertFilePath = cCAcertPath
				conninfo = conninfo + fmt.Sprintf("sslrootcert=%s ", cCAcertPath)
			} else {
				return nil, fmt.Errorf("Error wh: %s", err.Error())
			}
		}
		if cClientcert != nil {
			cclientCertPath, err1 := copyCertToTempFile(cClientcert, s.Name+"_CERT")
			if err1 == nil {
				sharedConn.ClientCertFilePath = cclientCertPath
				conninfo = conninfo + fmt.Sprintf("sslcert=%s ", cclientCertPath)
			} else {
				return nil, fmt.Errorf("Error : %s", err1.Error())
			}
		}
		if cClientKey != nil {
			cClientKeyPath, err2 := copyCertToTempFile(cClientKey, s.Name+"_Key")
			if err2 == nil {
				sharedConn.ClientKeyFilePath = cClientKeyPath
				conninfo = conninfo + fmt.Sprintf("sslkey=%s ", cClientKeyPath)
			} else {
				return nil, fmt.Errorf("Error : %s", err2.Error())
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
		db, err = sql.Open("postgres", conninfo)
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
		db, err = sql.Open("postgres", conninfo)
		if err != nil {
			// return nil, dont retry
			return nil, fmt.Errorf("Could not open connection to database %s, %s", cDbName, err.Error())
		}
		// sql returned db handle in first attempt
		if db != nil && err == nil {
			logCache.Info("Trying to ping the database server...")
			err = db.Ping()
			// retry attempt on ping only for conn refused and driver bad conn
			// When host is DNS for SSL conn and wifi is off, retry fails due to 'tcp: lookup' error. Works fine when host is IP Address
			if err != nil {
				if err == driver.ErrBadConn || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "network is unreachable") ||
					strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "dial tcp: lookup") ||
					strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "timedout") ||
					strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "net.Error") || strings.Contains(err.Error(), "i/o timeout") {
					logCache.Info("Failed to ping the database server, trying again...")
					for i := 1; i <= cMaxConnRetryAttempts; i++ {
						logCache.Infof("Connecting to database server... Attempt-[%d]", i)
						// retry delay
						time.Sleep(time.Duration(cConnRetryDelay) * time.Second)
						err = db.Ping()
						if err != nil {
							if err == driver.ErrBadConn || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "network is unreachable") ||
								strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "dial tcp: lookup") ||
								strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "timedout") ||
								strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "net.Error") || strings.Contains(err.Error(), "i/o timeout") {
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
	sharedConn.DatabaseType = s.DatabaseType
	sharedConn.db.SetMaxOpenConns(cMaxOpenConn)
	sharedConn.db.SetMaxIdleConns(cMaxIdleConn)
	sharedConn.db.SetConnMaxLifetime(lifetimeDuration)

	logCache.Debug("----------- DB Stats after setting extra connection configs -----------")
	logCache.Debug("Max Open Connections: " + strconv.Itoa(db.Stats().MaxOpenConnections))
	logCache.Debug("Number of Open Connections: " + strconv.Itoa(db.Stats().OpenConnections))
	logCache.Debug("In Use Connections: " + strconv.Itoa(db.Stats().InUse))
	logCache.Debug("Free Connections: " + strconv.Itoa(db.Stats().Idle))
	logCache.Debug("Max idle connection closed: " + strconv.FormatInt(db.Stats().MaxIdleClosed, 10))
	logCache.Debug("Max Lifetime connection closed: " + strconv.FormatInt(db.Stats().MaxLifetimeClosed, 10))

	err = sharedConn.db.Ping()
	if err != nil {
		return nil, err
	}
	return sharedConn, nil
}

// copyCertToTempFile creates temp mssql.pem file for running app in container
// and sqlserver needs filepath for ssl cert so can not pass byte array which we get from connection tile
func copyCertToTempFile(certdata []byte, name string) (string, error) {
	var path = name + "_" + "postgresql.pem"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return "", fmt.Errorf("Could not create file %s %s ", path, err.Error())
		}
		if err := os.Chmod(path, 0600); err != nil {
			return "", fmt.Errorf("Could not give permissions file %s %s ", path, err.Error())
		}
		defer file.Close()
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		return "", fmt.Errorf("Could not open file %s %s", path, err.Error())
	}
	_, err = file.Write(certdata)
	if err != nil {
		return "", fmt.Errorf("Could not write data to file %s %s", path, err.Error())
	}
	return path, nil
}

//Get cacert file from connection
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

// PgSharedConfigManager details
type PgSharedConfigManager struct {
	name               string
	db                 *sql.DB
	DatabaseType       string
	CACertFilePath     string
	ClientCertFilePath string
	ClientKeyFilePath  string
}

// Type PgSharedConfigManager details
func (p *PgSharedConfigManager) Type() string {
	return "Postgres"
}

// GetConnection PgSharedConfigManager details
func (p *PgSharedConfigManager) GetConnection() interface{} {
	return p.db
}

// ReleaseConnection PgSharedConfigManager details
func (p *PgSharedConfigManager) ReleaseConnection(connection interface{}) {

}

// Start PgSharedConfigManager details
func (p *PgSharedConfigManager) Start() error {
	return nil
}

// Stop PgSharedConfigManager details
func (p *PgSharedConfigManager) Stop() error {
	logCache.Debug("Cleaning up DB")
	//preparedQueryCache.Clear()

	for key := range preparedQueryCache {
		preparedQueryCache[key].Close()
	}

	p.db.Close()

	// clean up CA cert pem files on stop
	_, err := os.Stat(p.CACertFilePath)
	if !os.IsNotExist(err) {
		err := os.Remove(p.CACertFilePath)
		if err != nil {
			return fmt.Errorf("Error while removing pem file %s", err.Error())
		}
		logCache.Debug("CA Cert File successfully deleted")
	}

	// clean up client cert pem files on stop
	_, err = os.Stat(p.ClientCertFilePath)
	if !os.IsNotExist(err) {
		err := os.Remove(p.ClientCertFilePath)
		if err != nil {
			return fmt.Errorf("Error while removing pem file %s", err.Error())
		}
		logCache.Debug("Client Cert File successfully deleted")
	}
	// clean up client key pem files on stop
	_, err = os.Stat(p.ClientKeyFilePath)
	if !os.IsNotExist(err) {
		err := os.Remove(p.ClientKeyFilePath)
		if err != nil {
			return fmt.Errorf("Error while removing pem file %s", err.Error())
		}
		logCache.Debug("Client Key File successfully deleted")
	}
	return nil
}
