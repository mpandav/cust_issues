package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	dlog "log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
)

var logCache = log.ChildLogger(log.RootLogger(), "connection.kafka")
var factory = &KafkaFactory{}

type Settings struct {
	RefreshFrequency  int    `md:"refreshFrequency"`
	RetryMax          int    `md:"retryMax"`
	RetryBackoff      int    `md:"retryBackoff"`
	ConnectionTimeout int    `md:"connectionTimeout"`
	Name              string `md:"name,required"`
	Brokers           string `md:"brokers,required"`
	AuthMode          string `md:"authMode,allowed(None,SSL,SASL/PLAIN,SASL/SCRAM-SHA-256,SASL/SCRAM-SHA-512)"`
	UserName          string `md:"userName"`
	Password          string `md:"password"`
	SecurityProtocol  string `md:"securityProtocol,allowed(SASL_PLAINTEXT,SASL_SSL)"`
	ClientCert        string `md:"clientCert"`
	ClientKey         string `md:"clientKey"`
	CaCert            string `md:"caCert"`
	ClientID          string `md:"clientID"`
	ClientSecret      string `md:"clientSecret"`
	TokenURL          string `md:"tokenURL"`
	Scope             string `md:"scope"`
}

type KafkaClientConfig struct {
	Brokers      []string
	SaramaConfig sarama.Config
}

func init() {
	if os.Getenv(log.EnvKeyLogLevel) == "DEBUG" {
		// Enable debug logs for sarama lib
		sarama.Logger = dlog.New(os.Stderr, "[flogo-kafka]", dlog.LstdFlags)
	}

	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type KafkaFactory struct {
}

func (*KafkaFactory) Type() string {
	return "Kafka"
}

func (*KafkaFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {

	sharedConn := &KafkaSharedConfigManager{
		producers: make(map[string]sarama.SyncProducer),
		consumers: make(map[string]sarama.Consumer),
		plock:     &sync.RWMutex{},
		slock:     &sync.RWMutex{},
	}
	var err error
	sharedConn.config, err = getKafkaClientConfig(settings)
	if err != nil {
		return nil, err
	}
	return sharedConn, nil
}

type KafkaSharedConfigManager struct {
	config    *KafkaClientConfig
	name      string
	producers map[string]sarama.SyncProducer
	consumers map[string]sarama.Consumer
	plock     *sync.RWMutex
	slock     *sync.RWMutex
}

func (k *KafkaSharedConfigManager) Type() string {
	return "Kafka"
}

func (k *KafkaSharedConfigManager) GetConnection() interface{} {
	return k
}

func (k *KafkaSharedConfigManager) GetClientConfiguration() *KafkaClientConfig {
	return k.config
}

func (k *KafkaSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func (k *KafkaSharedConfigManager) AddProducer(name string, producer sarama.SyncProducer) {
	logCache.Debugf("Adding producer client [%s] in cache", name)
	defer k.plock.Unlock()
	k.plock.Lock()
	k.producers[name] = producer
}

func (k *KafkaSharedConfigManager) AddConsumer(name string, consumer sarama.Consumer) {
	logCache.Debugf("Adding consumer client [%s] in cache", name)
	defer k.slock.Unlock()
	k.slock.Lock()
	k.consumers[name] = consumer
}

func (k *KafkaSharedConfigManager) GetProducer(name string) sarama.SyncProducer {
	defer k.plock.RUnlock()
	k.plock.RLock()
	return k.producers[name]
}

func (k *KafkaSharedConfigManager) GetConsumer(name string) sarama.Consumer {
	defer k.slock.RUnlock()
	k.slock.RLock()
	return k.consumers[name]
}

func (k *KafkaSharedConfigManager) Start() error {
	return nil
}

func (k *KafkaSharedConfigManager) Stop() error {
	logCache.Debug("Cleaning up client cache")

	for n, producer := range k.producers {
		logCache.Debugf("Closing producer client [%s]", n)
		producer.Close()
	}

	for n, consumer := range k.consumers {
		logCache.Debugf("Closing consumer client [%s]", n)
		consumer.Close()
	}

	return nil
}

func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {

	var cManager connection.Manager
	var err error
	_, ok := conn.(map[string]interface{})
	if ok {
		cManager, err = handleLegacyConnection(conn)
	} else {
		cManager, err = coerce.ToConnection(conn)
	}

	if err != nil {
		return nil, err
	}

	return cManager, nil
}

func handleLegacyConnection(conn interface{}) (connection.Manager, error) {

	connectionObject, _ := coerce.ToObject(conn)
	if connectionObject == nil {
		return nil, errors.New("Connection object is nil")
	}

	id := connectionObject["id"].(string)

	cManager := connection.GetManager(id)
	if cManager == nil {

		connObject, err := generic.NewConnection(connectionObject)
		if err != nil {
			return nil, err
		}

		cManager, err = factory.NewManager(connObject.Settings())
		if err != nil {
			return nil, err
		}

		err = connection.RegisterManager(id, cManager)
		if err != nil {
			return nil, err
		}
	}

	return cManager, nil

}

func getKafkaClientConfig(settings map[string]interface{}) (*KafkaClientConfig, error) {
	connectionConfig := &KafkaClientConfig{}

	s := &Settings{}

	err := metadata.MapToStruct(settings, s, false)

	if err != nil {
		return nil, err
	}

	cName := s.Name
	brokers := s.Brokers

	if brokers == "" {
		return nil, errors.New("Invalid broker configuration. At-least one Kafka broker must be provided.")
	}

	connectionConfig.Brokers = strings.Split(brokers, ",")

	connectionConfig.SaramaConfig = *sarama.NewConfig()

	connectionConfig.SaramaConfig.Net.DialTimeout = time.Duration(s.ConnectionTimeout) * time.Second
	authMode := s.AuthMode
	if authMode == "SASL/PLAIN" {
		connectionConfig.SaramaConfig.Net.SASL.User = s.UserName
		if connectionConfig.SaramaConfig.Net.SASL.User == "" {
			return nil, errors.New("User name must be set for SASL authentication")
		}
		connectionConfig.SaramaConfig.Net.SASL.Password = s.Password
		if connectionConfig.SaramaConfig.Net.SASL.Password == "" {
			return nil, errors.New("Password  must be set for SASL authentication")
		}
		connectionConfig.SaramaConfig.Net.SASL.Enable = true

		// TODO Better configuration to identify Azure Event Hub
		if connectionConfig.SaramaConfig.Net.SASL.User == "$ConnectionString" {
			//This is Azure Event Hub
			connectionConfig.SaramaConfig.Net.SASL.Version = sarama.SASLHandshakeV0
		} else {
			connectionConfig.SaramaConfig.Net.SASL.Version = sarama.SASLHandshakeV1
		}
		connectionConfig.SaramaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	} else if authMode == "None" {
		// Reset security protocol for no auth
		s.SecurityProtocol = ""
	}

	if s.SecurityProtocol == "SASL_SSL" || authMode == "SSL" || authMode == "SASL/SCRAM-SHA-256" || authMode == "SASL/SCRAM-SHA-512" || authMode == "SASL/OAUTHBEARER" {
		connectionConfig.SaramaConfig.Net.TLS.Enable = true
	}

	if authMode == "SSL" || authMode == "SASL/PLAIN" || authMode == "SASL/SCRAM-SHA-256" || authMode == "SASL/SCRAM-SHA-512" || authMode == "SASL/OAUTHBEARER" {
		clientCets, err := decodeCerts(s.ClientCert)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to load Client certificate configured on Connection[%s]", cName))
		}
		clientPrivateKey, err := decodeCerts(s.ClientKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to load Client private key configured on Connection[%s]", cName))
		}

		caCert, err := decodeCerts(s.CaCert)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to load CA/Server certificate configured on Connection[%s]", cName))
		}

		connectionConfig.SaramaConfig.Net.TLS.Config, err = getTLSConfig(clientCets, clientPrivateKey, caCert, s.SecurityProtocol)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to process certifcate for Connection[%s] due to error - %s", cName, err.Error()))
		}

		if authMode == "SSL" && connectionConfig.SaramaConfig.Net.TLS.Config != nil && connectionConfig.SaramaConfig.Net.TLS.Config.InsecureSkipVerify {
			logCache.Warn("Server certificate validation is disabled since server certificate is not configured. This is not recommended in production.")
		}
	}

	if authMode == "SASL/SCRAM-SHA-256" || authMode == "SASL/SCRAM-SHA-512" {
		connectionConfig.SaramaConfig.Net.SASL.Enable = true
		connectionConfig.SaramaConfig.Net.SASL.User = s.UserName

		connectionConfig.SaramaConfig.Net.SASL.Handshake = true
		if connectionConfig.SaramaConfig.Net.SASL.User == "" {
			return nil, errors.New("User name must be set for SASL authentication")
		}
		connectionConfig.SaramaConfig.Net.SASL.Password = s.Password
		if connectionConfig.SaramaConfig.Net.SASL.Password == "" {
			return nil, errors.New("Password  must be set for SASL authentication")
		}
		if authMode == "SASL/SCRAM-SHA-256" {
			connectionConfig.SaramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
			connectionConfig.SaramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256

		}
		if authMode == "SASL/SCRAM-SHA-512" {
			connectionConfig.SaramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
			connectionConfig.SaramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		}
	} else if authMode == "SASL/OAUTHBEARER" {
		connectionConfig.SaramaConfig.Net.SASL.Enable = true
		connectionConfig.SaramaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		scope := strings.Split(s.Scope, ",")
		connectionConfig.SaramaConfig.Net.SASL.TokenProvider = NewTokenProvider(s.ClientID, s.ClientSecret, s.TokenURL, scope)
	}

	connectionConfig.SaramaConfig.Metadata.Retry.Backoff = time.Duration(s.RetryBackoff) * time.Millisecond
	connectionConfig.SaramaConfig.Metadata.Retry.Max = s.RetryMax
	connectionConfig.SaramaConfig.Metadata.RefreshFrequency = time.Duration(s.RefreshFrequency) * time.Minute
	connectionConfig.SaramaConfig.Version = sarama.V1_0_0_0
	connectionConfig.SaramaConfig.ClientID = engine.GetAppName() + "-" + engine.GetAppVersion() + "-" + strconv.Itoa(time.Now().Nanosecond())
	return connectionConfig, nil
}

func (c *KafkaClientConfig) CreateProducerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Net.TLS = c.SaramaConfig.Net.TLS
	config.Net.SASL.User = c.SaramaConfig.Net.SASL.User
	config.Net.SASL.Password = c.SaramaConfig.Net.SASL.Password
	config.Net.SASL.Enable = c.SaramaConfig.Net.SASL.Enable
	config.Net.SASL.SCRAMClientGeneratorFunc = c.SaramaConfig.Net.SASL.SCRAMClientGeneratorFunc
	config.Net.SASL.TokenProvider = c.SaramaConfig.Net.SASL.TokenProvider
	config.Net.SASL.Handshake = c.SaramaConfig.Net.SASL.Handshake
	config.Net.SASL.Mechanism = c.SaramaConfig.Net.SASL.Mechanism
	config.Net.SASL.Version = c.SaramaConfig.Net.SASL.Version
	config.Net.DialTimeout = c.SaramaConfig.Net.DialTimeout
	config.Metadata.Retry.Backoff = c.SaramaConfig.Metadata.Retry.Backoff
	config.Metadata.Retry.Max = c.SaramaConfig.Metadata.Retry.Max
	config.Metadata.RefreshFrequency = c.SaramaConfig.Metadata.RefreshFrequency
	config.Version = c.SaramaConfig.Version
	config.ClientID = c.SaramaConfig.ClientID
	return config
}

func (c *KafkaClientConfig) CreateConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Net.TLS = c.SaramaConfig.Net.TLS
	config.Net.SASL.User = c.SaramaConfig.Net.SASL.User
	config.Net.SASL.Password = c.SaramaConfig.Net.SASL.Password
	config.Net.SASL.Enable = c.SaramaConfig.Net.SASL.Enable
	config.Net.SASL.SCRAMClientGeneratorFunc = c.SaramaConfig.Net.SASL.SCRAMClientGeneratorFunc
	config.Net.SASL.TokenProvider = c.SaramaConfig.Net.SASL.TokenProvider
	config.Net.SASL.Handshake = c.SaramaConfig.Net.SASL.Handshake
	config.Net.SASL.Mechanism = c.SaramaConfig.Net.SASL.Mechanism
	config.Net.SASL.Version = c.SaramaConfig.Net.SASL.Version
	config.Net.DialTimeout = c.SaramaConfig.Net.DialTimeout
	config.Metadata.Retry.Backoff = c.SaramaConfig.Metadata.Retry.Backoff
	config.Metadata.Retry.Max = c.SaramaConfig.Metadata.Retry.Max
	config.Metadata.RefreshFrequency = c.SaramaConfig.Metadata.RefreshFrequency
	config.Version = c.SaramaConfig.Version
	config.ClientID = c.SaramaConfig.ClientID
	return config
}

func decodeCerts(certVal interface{}) ([]byte, error) {
	if certVal == nil {
		return nil, nil
	}
	certValStr := fmt.Sprintf("%v", certVal)

	if strings.HasPrefix(certValStr, "file://") {
		certFile := certValStr[7:]
		return ioutil.ReadFile(certFile)
	} else {
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
}

func getTLSConfig(clientCert []byte, clientKey []byte, caCert []byte, protocol string) (*tls.Config, error) {

	tlsConfig := &tls.Config{}
	if clientCert == nil && clientKey == nil && caCert == nil {
		if protocol == "" || protocol == "SASL_PLAINTEXT" {
			return nil, nil
		}
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
