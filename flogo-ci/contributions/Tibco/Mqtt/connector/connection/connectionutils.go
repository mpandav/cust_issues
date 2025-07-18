package connection

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
)

// ConnectClient ...
func ConnectClient(settings *Settings) (client mqtt.Client, err error) {
	encryptionMode := settings.Encryptionmode
	if !(encryptionMode == "None" || encryptionMode == "TLS-Cert" || encryptionMode == "TLS-ClientAuth") {
		return nil, fmt.Errorf("Invalid encryption mode: %s", encryptionMode)
	}
	cacert := settings.Cacert
	clientcert := settings.Clientcert
	clientkey := settings.Clientkey

	opts := mqtt.NewClientOptions()
	opts.AddBroker(settings.Broker)
	opts.SetUsername(settings.User)
	opts.SetPassword(settings.Password)
	// if cleanSession is false use ResumeSubs
	// opts.ResumeSubs = true
	opts.AutoReconnect = true
	opts.OnConnect = onConnectHandler
	opts.OnConnectionLost = connectionLostHandler
	// opts.KeepAlive = 5
	// opts.PingTimeout = 5 * time.Second
	// opts.ConnectTimeout = 5 * time.Second
	if settings.ClientId != "" {
		opts.SetClientID(settings.ClientId)
	}
	opts.SetStore(mqtt.NewMemoryStore())
	if encryptionMode == "TLS-Cert" {
		tlsConfig, err := getTLSConfigFromIntf(cacert, nil, nil)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tlsConfig)
	} else if encryptionMode == "TLS-ClientAuth" {
		tlsConfig, err := getTLSConfigFromIntf(cacert, clientcert, clientkey)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tlsConfig)
	}
	if settings.ShowWill {
		willTopic := settings.WillTopic
		will := settings.Will
		qos := settings.WillQoS
		retain := settings.WillRetain
		logger.Debugf("Setting will related fields: Will Topic [%s] Will Message: [%s] Will QoS: [%d] Will Retain [%t]", willTopic, will, qos, retain)
		opts.SetBinaryWill(willTopic, []byte(will), byte(qos), retain)
	}
	logger.Infof("Creating new Mqtt connection with name: [%s]", settings.Name)
	logger.Debugf("Creating Mqtt client with parms: Broker [%s] User [%s] Encryption Mode [%s] CACert set? [%t] ClientCert set? [%t] ClientKey set? [%t]",
		settings.Broker, settings.User, settings.Encryptionmode, cacert != "", clientcert != "", clientkey != "")
	client = mqtt.NewClient(opts)
	logger.Debug("Establishing Mqtt connection with broker...")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, activity.NewError(fmt.Sprintf("Connection to mqtt broker failed %v", token.Error()), "MQTT-PUB-4002", nil)
	}
	return client, nil
}

func onConnectHandler(c mqtt.Client) {
	if !isFirstAttemptToConnect {
		logger.Info("Mqtt client reconnected successfully")
		OnConnectNotifier <- true
	} else {
		logger.Info("Mqtt client connected successfully")
	}
	isFirstAttemptToConnect = false
}

func connectionLostHandler(c mqtt.Client, err error) {
	logger.Infof("Mqtt client got disconnected unintentionally with error: [%s]...Reconnecting...", err.Error())
}

func getTLSConfigFromIntf(cacert interface{}, clientcert interface{}, clientkey interface{}) (*tls.Config, error) {
	certpool := x509.NewCertPool()
	certBytes, err := certDecode(cacert)
	if err != nil {
		return nil, err
	}
	if !certpool.AppendCertsFromPEM(certBytes) {
		return nil, fmt.Errorf("Failed to parse cacert PEM data from connection")
	}
	var cert tls.Certificate
	if clientcert != nil && clientkey != nil {
		clientCertBytes, err := certDecode(clientcert)
		if err != nil {
			return nil, err
		}
		clientKeyBytes, err := certDecode(clientkey)
		if err != nil {
			return nil, err
		}
		cert, err = tls.X509KeyPair(clientCertBytes, clientKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to create an internal keypair from the client cert and key provided on the connection for reason: %s", err)
		}
	}
	// Create tls.Config with desired tls properties
	return &tls.Config{
		RootCAs:            certpool,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

// decodeCertificate shamelessly swiped from graphql/server.go in the hope that this convention for encoding
// certs in app properties will eventually be documented.
func certDecode(cert interface{}) ([]byte, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate contains is nil")
	}
	if reflect.TypeOf(cert).String() == "map[string]interface {}" {
		logger.Debug("Mqtt Publisher configured file selector")
		cacert := cert.(map[string]interface{})
		var header = "base64,"
		value := cacert["content"].(string)
		if value == "" {
			return nil, fmt.Errorf("certificate contains no data")
		}
		if strings.Index(value, header) >= 0 {
			value = value[strings.Index(value, header)+len(header):]
			decodedLen := base64.StdEncoding.DecodedLen(len(value))
			destArray := make([]byte, decodedLen)
			_, err := base64.StdEncoding.Decode(destArray, []byte(value))
			if err != nil {
				return nil, fmt.Errorf("certificate not base64 encoded in config: [%s]", err)
			}
			return destArray, nil
		}
	} else if reflect.TypeOf(cert).String() == "string" {
		if strings.HasPrefix(cert.(string), "{") {
			logger.Debug("Mqtt Publisher configured from file selector")
			certObj, err := coerce.ToObject(cert)
			if err == nil {
				certValue, ok := certObj["content"].(string)
				if !ok || certValue == "" {
					return nil, fmt.Errorf("No content found for certificate")
				}
				return base64.StdEncoding.DecodeString(strings.Split(certValue, ",")[1])
			}
		}
		index := strings.IndexAny(cert.(string), ",")
		if index > -1 {
			//some encoding is there
			logger.Debug("Mqtt Publisher configured with a base64 mime string")
			encoding := cert.(string)[:index]
			certValue := cert.(string)[index+1:]
			if strings.EqualFold(encoding, "base64") {
				return base64.StdEncoding.DecodeString(certValue)
			}
			return nil, fmt.Errorf("error parsing the certificate or encoding [%s] may not be supported", encoding)
		} else if strings.HasPrefix(cert.(string), "file://") {
			// app property pointing to a file
			logger.Debug("Mqtt Publisher configured with file url [%s]", cert.(string))
			fileName := cert.(string)[7:]
			return ioutil.ReadFile(fileName)
		} else if strings.Contains(cert.(string), "/") || strings.Contains(cert.(string), "\\") {
			logger.Debug("Mqtt Publisher configured with fully qualified file name [%s]", cert.(string))
			_, err := os.Stat(cert.(string))
			if err != nil {
				return nil, fmt.Errorf("Mqtt Publisher cannot find certificate file: [%s] for reason: [%s]", cert.(string), err)
			}
			return ioutil.ReadFile(cert.(string))
		}
	}
	return nil, fmt.Errorf("certificate is not formatted correctly or is not available")
}
