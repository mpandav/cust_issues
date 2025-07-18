package connection

import (
	"errors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
)

var logger = log.ChildLogger(log.RootLogger(), "mqtt-connection")
var factory = &MqttFactory{}

// OnConnectNotifier will be used when client gets reconnected again
var OnConnectNotifier = make(chan bool)
var isFirstAttemptToConnect = true

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// Type ...
func (*MqttFactory) Type() string {
	return "Mqtt"
}

// NewManager ...
func (*MqttFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	configManager := &MqttConfigManager{}

	var err error
	s := &Settings{}
	err = metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}
	configManager.ClientConfig = s
	// MqttConn.client, err = ConnectClient(s)
	// if err != nil {
	// 	return nil, err
	// }
	return configManager, nil
}

// Type ...
func (k *MqttConfigManager) Type() string {
	return "Mqtt"
}

// GetConnection ...
func (k *MqttConfigManager) GetConnection() interface{} {
	return k
}

// GetMqttClient ...
func (k *MqttConfigManager) GetMqttClient() (mqtt.Client, error) {
	if k.Client == nil {
		k.Lock()
		defer k.Unlock()
		client, err := ConnectClient(k.ClientConfig)
		if err != nil {
			logger.Errorf("Unable to get Mqtt client [%s]", err.Error())
			return nil, err
		}
		k.Client = client
	}
	return k.Client, nil
}

// ReleaseConnection ...
func (k *MqttConfigManager) ReleaseConnection(connection interface{}) {
}

// Start ...
func (k *MqttConfigManager) Start() error {
	return nil
}

// Stop ...
func (k *MqttConfigManager) Stop() error {
	return nil
}

// GetSharedConfiguration ...
func GetSharedConfiguration(settings interface{}) (connection.Manager, error) {
	var cm connection.Manager
	var err error
	_, ok := settings.(map[string]interface{})
	if ok {
		logger.Info("Legacy connection detected")
		cm, err = handleLegacyConnection(settings)
	} else {
		cm, err = coerce.ToConnection(settings)
	}
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func handleLegacyConnection(settings interface{}) (connection.Manager, error) {
	connectionObject, _ := coerce.ToObject(settings)
	if connectionObject == nil {
		return nil, errors.New("Connection object is nil")
	}
	id := connectionObject["id"].(string)
	cm := connection.GetManager(id)
	if cm == nil {
		conn, err := generic.NewConnection(connectionObject)
		if err != nil {
			return nil, err
		}
		cm, err = factory.NewManager(conn.Settings())
		if err != nil {
			return nil, err
		}
		err = connection.RegisterManager(id, cm)
		if err != nil {
			return nil, err
		}
	}
	return cm, nil
}
