package connection

import (
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Settings ...
type Settings struct {
	Name           string `md:"name"`
	Description    string `md:"description"`
	Broker         string `md:"broker"`
	User           string `md:"user"`
	Password       string `md:"password"`
	Encryptionmode string `md:"encryptionMode"`
	Cacert         string `md:"cacert"`
	Clientcert     string `md:"clientcert"`
	Clientkey      string `md:"clientkey"`
	ShowWill       bool   `md:"showwill"`
	Will           string `md:"will"`
	WillTopic      string `md:"willtopic"`
	WillQoS        int    `md:"willqos"`
	WillRetain     bool   `md:"willretain"`
	ClientId       string `md:"clientid"`
}

// MqttFactory ...
type MqttFactory struct {
}

// MqttConfigManager ...
type MqttConfigManager struct {
	sync.Mutex
	ClientConfig *Settings
	Client       mqtt.Client
}
