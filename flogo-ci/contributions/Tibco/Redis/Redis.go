package Redis

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	redisconnection "github.com/tibco/wi-redis/src/app/Redis/connector/connection"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	"github.com/redis/go-redis/v9"
)

var redisActivityLogger = log.ChildLogger(log.RootLogger(), "Redis")

// GetConnection returns a deserialized Zoho conneciton object it does not establish a
// connection with the zoho. If a connection with the same id as in the context is
// present in the cache, that connection from the cache is returned
func GetConnection(connection *redisconnection.RedisSharedConfigManager, inputParams map[string]interface{}) (con *redisconnection.RedisSharedConfigManager, err error) {

	var dbIndex float64

	if val, ok := inputParams["DatabaseIndex"]; ok {
		dbIndex = val.(float64)
		connection.DefaultDatabaseIndex = dbIndex

	}
	redisActivityLogger.Info("Database Partition selected: ", dbIndex)

	if connection.RedisClient == nil {
		redisActivityLogger.Info(connection.AuthMode)
		if connection.AuthMode == true {
			caCert, err := decodeCerts(connection.CaCert)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to load CA/Server certificate configured on Connection[%s]", connection.ConnectionName))
			}

			clientCert, err := decodeCerts(connection.ClientCert)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to load Client certificate configured on Connection[%s]", connection.ConnectionName))
			}

			clientkey, err := decodeCerts(connection.ClientKey)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to load Client key  configured on Connection[%s]", connection.ConnectionName))
			}

			tlsConfig, err := getTLSConfig(clientCert, clientkey, caCert, connection.RedisToken.Host)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to process certifcate due to error - %s", err.Error()))
			}

			connection.RedisClient = redis.NewClient(&redis.Options{
				Addr:      connection.RedisToken.Host + ":" + strconv.Itoa(connection.RedisToken.Port),
				Password:  connection.RedisToken.Password, // no password set
				DB:        int(dbIndex),                   // use default DB
				TLSConfig: tlsConfig,
			})

		} else {
			connection.RedisClient = redis.NewClient(&redis.Options{
				Addr:     connection.RedisToken.Host + ":" + strconv.Itoa(connection.RedisToken.Port),
				Password: connection.RedisToken.Password,
				DB:       int(dbIndex), // use default DB
			})

		}

	}

	//cachedConnection[id] = connection
	redisActivityLogger.Info("Returning New connection")
	return connection, nil

}

func decodeCerts(certVal interface{}) ([]byte, error) {
	if certVal == nil {
		return nil, nil
	}

	//TODO: Workaround
	certObj, err := coerce.ToObject(certVal)
	if err == nil {
		certVal = certObj["content"]
	}

	certStringVal, ok := certVal.(string)
	if !ok || certStringVal == "" {
		return nil, nil
	}

	index := strings.IndexAny(certStringVal, ",")
	if index > -1 {
		certStringVal = certStringVal[index+1:]
	}

	return base64.StdEncoding.DecodeString(certStringVal)
}

func getTLSConfig(clientCert []byte, clientKey []byte, caCert []byte, server string) (*tls.Config, error) {

	tlsConfig := &tls.Config{ServerName: server}
	if clientCert == nil && clientKey == nil && caCert == nil {

		// SSL/TLS is enabled but no certificates are configured
		tlsConfig.InsecureSkipVerify = true
		tlsConfig.ClientAuth = 0
	} else {
		if caCert != nil {
			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(caCert)
			if !ok {
				return nil, errors.New("Invalid CA/Server certificate. It must be a valid PEM certificate.")
			}
			tlsConfig.RootCAs = caCertPool
		} else {
			tlsConfig.InsecureSkipVerify = true
		}

		if clientCert != nil && clientKey != nil {
			//Mutual authentication enabled

			cert, err := tls.X509KeyPair(clientCert, clientKey)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
			tlsConfig.BuildNameToCertificate()
			tlsConfig.ClientAuth = 4
		}
	}

	return tlsConfig, nil
}
