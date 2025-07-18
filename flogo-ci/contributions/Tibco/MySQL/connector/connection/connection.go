package connection

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

var logCache = log.ChildLogger(log.RootLogger(), "mysql.connection")
var factory = &Factory{}

// Settings for MySQL connection
type Settings struct {
	Name                 string `md:"name,required"`
	Description          string `md:"description"`
	Host                 string `md:"host,required"`
	Port                 int    `md:"port,required"`
	DbName               string `md:"databaseName,required"`
	User                 string `md:"user,required"`
	Password             string `md:"password,required"`
	Onprem               bool   `md:"onprem,required"`
	MaxOpenConns         int    `md:"maxopenconnection"`
	MaxIdleConns         int    `md:"maxidleconnection"`
	MaxConnLifetime      string `md:"connmaxlifetime"`
	MaxConnRetryAttempts int    `md:"maxconnectattempts"`
	ConnRetryDelay       int    `md:"connectionretrydelay"`
	ConnectionTimeout    int    `md:"connectiontimeout"`
	TLSConfig            bool   `md:"tlsconfig"`
	TLSMode              string `md:"tlsparam"`
	Cacert               string `md:"cacert"`
	Clientcert           string `md:"clientcert"`
	Clientkey            string `md:"clientkey"`
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

// Factory for mysql connection
type Factory struct {
}

// Type MySQLFactory
func (*Factory) Type() string {
	return "MySQL"
}

// NewManager for MySQL
func (*Factory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &SharedConfigManager{}
	var err error

	// to maintain backward compatibility with FE 2.13.0 release wherein we hardcoded tcptimeout value in mysql to 150sec
	var tcpTimeout = 150

	tcpTimeoutInt, ok := settings["connectiontimeout"]
	if ok {
		tcpTimeout, _ = coerce.ToInt(tcpTimeoutInt)
	}

	setting := &Settings{}
	err = metadata.MapToStruct(settings, setting, false)

	setting.ConnectionTimeout = tcpTimeout

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
	cDbName := setting.DbName
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

	cMaxOpenConn := setting.MaxOpenConns
	if cMaxOpenConn == 0 {
		logCache.Debug("Default value of Max open connection chosen. Default is 0")
	}
	if cMaxOpenConn < 0 {
		logCache.Debugf("Max open connection value received is %d, it will be defaulted to 0", cMaxOpenConn)
	}
	cMaxIdleConn := setting.MaxIdleConns
	if cMaxIdleConn == 2 {
		logCache.Debug("Default value of Max idle connection chosen. Default is 2")
	}
	if cMaxIdleConn < 0 {
		logCache.Debugf("Max idle connection value received is %d, it will be defaulted to 0", cMaxIdleConn)
	}
	cMaxConnLifetime := setting.MaxConnLifetime
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
	} else {
		logCache.Debugf("Connection timeout selected: %d secs", cConnTimeout)
	}

	logCache.Debug("Onprem selected?", cOnpremstr)

	cTLSConfig := setting.TLSConfig
	var conninfo string
	if cTLSConfig == false {
		logCache.Debugf("Login attempting plain connection")
		conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cConnTimeout)
	} else {
		cTLSMode := setting.TLSMode
		//connection name is used to register and deregister tlsconfig for secure connections to DB
		sharedConn.name = cName

		// extract byte array of cert data fro file content string
		if cTLSMode != "" {
			logCache.Debugf("Login attempting SSL connection")

			Cacert, _ := getByteCertDataForPemFile(setting.Cacert, "ca")
			Clientcert, _ := getByteCertDataForPemFile(setting.Clientcert, "client")
			Clientkey, _ := getByteCertDataForPemFile(setting.Clientkey, "clientkey")
			if Clientcert == nil && Clientkey == nil && Cacert != nil {
				tlsconfig, err := getTLSConfigFromConfigOneWay(Cacert, (cTLSMode != "VerifyIdentity"))
				if err != nil {
					if cTLSMode == "Preferred" {
						logCache.Infof("Warning: %s", err.Error())
					} else {
						return nil, err
					}
				}
				//register tls config for cacert
				//fix for Preferred mode when server is in nonssl mode and fallback should happen without giving error even if certs are not valid
				if cTLSMode == "Preferred" {
					mysql.RegisterTLSConfig("preferred", tlsconfig)
					conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=preferred&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cConnTimeout)
				} else {
					mysql.RegisterTLSConfig(cName, tlsconfig)
					conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cName, cConnTimeout)
				}
			} else if Clientcert != nil && Clientkey != nil && Cacert == nil {
				tlsconfig, err := getTLSConfigFromConfigForCertAndKey(Clientcert, Clientkey, (cTLSMode != "VerifyIdentity"))
				if err != nil {
					if cTLSMode == "Preferred" {
						logCache.Infof("Warning: %s", err.Error())
					} else {
						return nil, err
					}
				}
				//register tls config for cacert
				//fix for Preferred mode when server is in nonssl mode and fallback should happen without giving error even if certs are not valid
				if cTLSMode == "Preferred" {
					mysql.RegisterTLSConfig("preferred", tlsconfig)
					conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=preferred&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cConnTimeout)
				} else {
					mysql.RegisterTLSConfig(cName, tlsconfig)
					conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cName, cConnTimeout)
				}

			} else if Clientcert != nil && Clientkey != nil && Cacert != nil {
				tlsconfig, err := getTLSConfigFromConfig(Cacert, Clientcert, Clientkey, (cTLSMode != "VerifyIdentity"))
				if err != nil {
					if cTLSMode == "Preferred" {
						logCache.Infof("Warning: %s", err.Error())
					} else {
						return nil, err
					}
				}
				//register tls config for cacert
				//fix for Preferred mode when server is in nonssl mode and fallback should happen without giving error even if certs are not valid
				if cTLSMode == "Preferred" {
					mysql.RegisterTLSConfig("preferred", tlsconfig)
					conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=preferred&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cConnTimeout)
				} else {
					mysql.RegisterTLSConfig(cName, tlsconfig)
					conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?tls=%s&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cName, cConnTimeout)
				}
			} else {
				cMode := decodeTLSParamIfNoCertsPassed(cTLSMode)
				conninfo = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s%s&timeout=%ds", cUser, cPassword, cHost, cPort, cDbName, cMode, cConnTimeout)
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
		logCache.Debug("No connection retry selected, connection attempt will be tried only once...")
		db, err = sql.Open("mysql", conninfo)
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
		db, err = sql.Open("mysql", conninfo)
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
				if err == driver.ErrBadConn || err == mysql.ErrInvalidConn || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "network is unreachable") ||
					strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "dial tcp: lookup") ||
					strings.Contains(err.Error(), "connection timed out") || strings.Contains(err.Error(), "timedout") || strings.Contains(err.Error(), "time out") ||
					strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "net.Error") || strings.Contains(err.Error(), "i/o timeout") {
					logCache.Info("Failed to ping the database server, trying again...")
					for i := 1; i <= cMaxConnRetryAttempts; i++ {
						logCache.Infof("Connecting to database server... Attempt-[%d]", i)
						// retry delay
						time.Sleep(time.Duration(cConnRetryDelay) * time.Second)
						err = db.Ping()
						if err != nil {
							if err == driver.ErrBadConn || err == mysql.ErrInvalidConn || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "network is unreachable") ||
								strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "dial tcp: lookup") ||
								strings.Contains(err.Error(), "connection timed out") || strings.Contains(err.Error(), "timedout") || strings.Contains(err.Error(), "time out") ||
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

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return sharedConn, nil
}

// decodeTLSParm translate tls parm to connection string
func decodeTLSParam(tlsparm string, connectionName string) string {
	switch tlsparm {
	case "Disabled":
		return "?tls=false"
	case "Required":
		return "?tls=skip-verify"
	case "Preferred":
		return "?tls=preferred"
	case "VerifyCA":
		return "?tls=" + connectionName
	case "VerifyIdentity":
		return "?tls=" + connectionName
	default:
		return ""
	}
}

// decodeTLSParamIfNoCertsPassed translate tls parm to connection string
func decodeTLSParamIfNoCertsPassed(tlsparm string) string {
	switch tlsparm {
	case "Disabled":
		return "?tls=false"
	case "Required":
		return "?tls=skip-verify"
	case "Preferred":
		return "?tls=preferred"
	case "VerifyCA":
		return "?tls=skip-verify"
	default:
		return ""
	}
}

//Get cacert file from connection
func getByteCertDataForPemFile(cert string, certType string) ([]byte, error) {
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
	// because user might send non base64 encoded data which should be considered as incorrect
	// and passed to the server
	certRealValue, err := base64.StdEncoding.DecodeString(cert)
	return certRealValue, err
}

func getTLSConfigFromConfig(Cacert []byte, Clientcert []byte, Clientkey []byte, insecureSkipVerify bool) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(Cacert) {
		return nil, fmt.Errorf("Failed to parse cacert PEM data from connection")
	}
	var cert tls.Certificate
	cert, err := tls.X509KeyPair(Clientcert, Clientkey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an internal keypair from the client cert and key provided on the connection for reason: %s", err)
	}

	// Create tls.Config with desired tls properties
	return &tls.Config{
		RootCAs:            certpool,
		InsecureSkipVerify: insecureSkipVerify,
		Certificates:       []tls.Certificate{cert},
	}, nil

}

// Require X509 user case where connection works only with cert and key too
func getTLSConfigFromConfigForCertAndKey(Clientcert []byte, Clientkey []byte, insecureSkipVerify bool) (*tls.Config, error) {
	var cert tls.Certificate
	cert, err := tls.X509KeyPair(Clientcert, Clientkey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create an internal keypair from the client cert and key provided on the connection for reason: %s", err)
	}
	// Create tls.Config with desired tls properties
	return &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		Certificates:       []tls.Certificate{cert},
	}, nil

}

func getTLSConfigFromConfigOneWay(Cacert []byte, insecureSkipVerify bool) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	if !certpool.AppendCertsFromPEM(Cacert) {
		return nil, fmt.Errorf("Failed to parse cacert PEM data from connection")
	}
	// Create tls.Config with desired tls properties
	// return &tls.Config{}, nil
	return &tls.Config{
		RootCAs:            certpool,
		InsecureSkipVerify: insecureSkipVerify,
	}, nil
}

// SharedConfigManager details
type SharedConfigManager struct {
	name string
	db   *sql.DB
}

// Type SharedConfigManager details
func (sharedconnection *SharedConfigManager) Type() string {
	return "MySQL"
}

// GetConnection SharedConfigManager details
func (sharedconnection *SharedConfigManager) GetConnection() interface{} {
	return sharedconnection.db
}

// ReleaseConnection SharedConfigManager details
func (sharedconnection *SharedConfigManager) ReleaseConnection(connection interface{}) {

}

// Start SharedConfigManager details
func (sharedconnection *SharedConfigManager) Start() error {
	return nil
}

// Stop SharedConfigManager details
func (sharedconnection *SharedConfigManager) Stop() error {
	logCache.Debugf("Closing consumer client [%s]", sharedconnection)

	//deregister the tlsconfig for ssl
	mysql.DeregisterTLSConfig(sharedconnection.name)
	mysql.DeregisterTLSConfig("preferred")

	for key := range preparedQueryCache {
		preparedQueryCache[key].Close()
	}

	sharedconnection.db.Close()

	return nil
}
