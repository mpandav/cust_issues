package ems

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/msg-ems-client-go/tibems"
)

var logCache = log.ChildLogger(log.RootLogger(), "flogo-ems-connection")
var factory = &EMSFactory{}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type EMSFactory struct {
}

func (*EMSFactory) Type() string {
	return "EMS"
}

type Settings struct {
	Name                 string `md:"name"`
	ClientID             string `md:"clientID"`
	AuthMode             string `md:"authenticationMode"`
	Password             string `md:"password"`
	UserName             string `md:"userName"`
	ServerUrl            string `md:"serverUrl"`
	ClientCert           string `md:"clientCert"`
	ClientKey            string `md:"clientKey"`
	CaCert               string `md:"caCert"`
	PrivateKeyPwd        string `md:"privateKeyPassword"`
	HostnameVerification bool   `md:"noVerifyHostname"`
	ReconnectCount       int32  `md:"reconnectCount"`
	ReconnectDelay       int32  `md:"reconnectDelay"`
	ReconnectTimeout     int32  `md:"reconnectTimeout"`
	EnableMTLS           bool   `md:"enablemTLS"`
}

func (*EMSFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {

	s := &Settings{}
	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}
	cmanager := &EmsSharedConfigManager{
		Settings: s,
	}

	// append server url
	serverUrls := strings.Split(s.ServerUrl, ",")
	if len(serverUrls) == 1 {
		s.ServerUrl = s.ServerUrl + "," + s.ServerUrl
	}

	cmanager.ConnectionFactoy, err = tibems.CreateConnectionFactory(s.ServerUrl)
	if err != nil {
		return nil, err
	}

	if s.ClientID != "" {
		cmanager.ConnectionFactoy.SetClientID(s.ClientID)
	}
	// this will retry connection very first try
	cmanager.ConnectionFactoy.SetConnectAttemptCount(s.ReconnectCount)
	cmanager.ConnectionFactoy.SetConnectAttemptDelay(s.ReconnectDelay)
	cmanager.ConnectionFactoy.SetConnectAttemptTimeout(s.ReconnectTimeout)

	// set reconnection parameters in case network issues
	cmanager.ConnectionFactoy.SetReconnectAttemptCount(s.ReconnectCount)
	cmanager.ConnectionFactoy.SetReconnectAttemptDelay(s.ReconnectDelay)
	cmanager.ConnectionFactoy.SetReconnectAttemptTimeout(s.ReconnectTimeout)

	if s.AuthMode == "SSL" {
		sslParams, err := tibems.CreateSSLParams()
		if err != nil {
			return nil, err
		}
		// server certificate
		caCert, err := decodeCerts(s.CaCert)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to load Server certificate"))
		}
		err = sslParams.AddTrustedCert(caCert, tibems.SslEncodingPem)
		if err != nil {
			return nil, err
		}

		// Set hostname verification
		err = sslParams.SetVerifyHostName(s.HostnameVerification)
		if err != nil {
			return nil, err
		}

		if s.EnableMTLS {
			// client certificate
			clientCert, err := decodeCerts(s.ClientCert)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to load Client certificate"))
			}
			err = sslParams.SetIdentity(clientCert, tibems.SslEncodingPem)
			if err != nil {
				return nil, err
			}

			// Client key
			clientKey, err := decodeCerts(s.ClientKey)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Failed to load Client key"))
			}
			err = sslParams.SetPrivateKey(clientKey, tibems.SslEncodingPem)
			if err != nil {
				return nil, err
			}
		}

		//set ssl debug logs
		tibems.SetSSLDebugTrace(logCache.DebugEnabled())
		tibems.SetSSLTrace(logCache.DebugEnabled())

		err = cmanager.ConnectionFactoy.SetSSLParams(sslParams)
		if err != nil {
			return nil, err
		}

		//set private key password for client key
		if s.PrivateKeyPwd != "" {
			err = cmanager.ConnectionFactoy.SetPkPassword(s.PrivateKeyPwd)
			if err != nil {
				return nil, err
			}
		}
	}
	cmanager.Connection, err = cmanager.ConnectionFactoy.CreateConnection(s.UserName, s.Password)
	if err != nil {
		return nil, err
	}

	// Log retry attempts
	err = cmanager.Connection.SetExceptionListener(func(connection *tibems.Connection, err error) {
		if err != nil && strings.Contains(err.Error(), "Reconnected") {
			logCache.Infof("Reconnected successfully to EMS server..!")
			return
		}
		logCache.Infof("Connection Exception: %s, retrying connection ...", err.Error())
	})
	if err != nil {
		return nil, err
	}
	tibems.SetExceptionOnFTEvents(true)

	// set client Id incase not set by EMS server
	clientId, err := cmanager.Connection.GetClientID()
	if clientId != "" {
		cmanager.Connection.SetClientID(s.Name)
	}
	return cmanager, nil
}

type EmsSharedConfigManager struct {
	Connection       *tibems.Connection
	ConnectionFactoy *tibems.ConnectionFactory
	Settings         *Settings
}

func (k *EmsSharedConfigManager) Type() string {
	return "EMS"
}

func (k *EmsSharedConfigManager) GetConnection() interface{} {
	return k.Connection
}

func (k *EmsSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func (k *EmsSharedConfigManager) Start() (err error) {
	err = k.Connection.Start()
	if err != nil {
		return err
	}
	return nil
}

func (k *EmsSharedConfigManager) Stop() (err error) {
	if k.Connection != nil {
		err = k.Connection.Stop()
		if err != nil {
			return err
		}

		err = k.Connection.Close()
		if err != nil {
			return err
		}
	}

	if k.ConnectionFactoy != nil {
		err = k.ConnectionFactoy.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeCerts(certVal interface{}) ([]byte, error) {
	if certVal == nil {
		return nil, nil
	}
	certValStr := fmt.Sprintf("%v", certVal)

	if strings.HasPrefix(certValStr, "file://") {
		certFile := certValStr[7:]
		return os.ReadFile(certFile)
	} else {
		certObj, err := coerce.ToObject(certVal)
		if err == nil {
			certVal = certObj["content"]
		}

		certStringVal, ok := certVal.(string)
		if !ok || certStringVal == "" {
			return nil, fmt.Errorf("Failed to read the SSL certificates")
		}

		index := strings.IndexAny(certStringVal, ",")
		if index > -1 {
			certStringVal = certStringVal[index+1:]
		}

		return base64.StdEncoding.DecodeString(certStringVal)
	}
}
