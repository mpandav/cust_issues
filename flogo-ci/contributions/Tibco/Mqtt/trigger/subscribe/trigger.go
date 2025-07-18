package subscribe

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	mqttconnection "github.com/tibco/wi-mqtt/src/app/Mqtt/connector/connection"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})
var logger = log.ChildLogger(log.RootLogger(), "trigger-mqtt-subscriber")

func init() {
	_ = trigger.Register(&MqttTrigger{}, &MqttFactory{})
}

// MqttFactory ...
type MqttFactory struct {
}

// Metadata ...
func (f *MqttFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New ...
func (*MqttFactory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}
	err := metadata.MapToStruct(config.Settings, s, true)
	if err != nil {
		return nil, err
	}
	return &MqttTrigger{id: config.Id, config: config, settings: s}, nil
}

// MqttTrigger ...
type MqttTrigger struct {
	id       string
	config   *trigger.Config
	settings *Settings
	handlers []trigger.Handler
	clients  []*MqttHandler
}

// MqttHandler ...
type MqttHandler struct {
	client   *mqtt.Client
	handler  trigger.Handler
	settings *HandlerSettings
}

// Initialize ...
func (t *MqttTrigger) Initialize(ctx trigger.InitContext) error {
	t.handlers = ctx.GetHandlers()

	configManager, err := mqttconnection.GetSharedConfiguration(t.settings.MqttConnection)
	if err != nil {
		return err
	}
	mqttConfigManager, ok := configManager.(*mqttconnection.MqttConfigManager)
	if !ok {
		return errors.New("Unable to get config manager for mqtt trigger")
	}
	// Init handlers
	for _, handler := range t.handlers {
		var mqttHandler = MqttHandler{}
		mqttHandler.handler = handler
		handlerSettingsJSON, err := json.Marshal(handler.Settings())
		if err != nil {
			return errors.New("Unable to marshal handlerSettings")
		}
		if err := json.Unmarshal(handlerSettingsJSON, &mqttHandler.settings); err != nil {
			return errors.New("Unable to unmarshal handlerSettings")
		}
		if mqttHandler.settings.ShowWill {
			logger.Info("Will fields are set...Creating new mqtt connection")
			config := mqttConfigManager.ClientConfig
			config.ShowWill = true
			config.Will = mqttHandler.settings.Will
			config.WillTopic = mqttHandler.settings.WillTopic
			config.WillQoS = mqttHandler.settings.WillQoS
			config.WillRetain = mqttHandler.settings.WillRetain
			client, err := mqttconnection.ConnectClient(config)
			if err != nil {
				return err
			}
			mqttHandler.client = &client
		} else {
			_, err = mqttConfigManager.GetMqttClient()
			if err != nil {
				return err
			}
			mqttHandler.client = &mqttConfigManager.Client
		}
		t.clients = append(t.clients, &mqttHandler)
	}
	logger.Debug("Mqtt Subscriber was successfully initialized")
	return nil
}

// Start ...
func (t *MqttTrigger) Start() error {
	for _, mqttHandler := range t.clients {
		handlerSettings := mqttHandler.settings
		mqttClient := *mqttHandler.client
		if !mqttClient.IsConnected() {
			if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
				msg := fmt.Sprintf("Mqtt Subscriber failed to subscribe for reason: %s", token.Error())
				logger.Debug(msg)
				return fmt.Errorf(msg)
			}
		}
		token := mqttClient.Subscribe(mqttHandler.settings.Topic, byte(mqttHandler.settings.QoS), func(client mqtt.Client, msg mqtt.Message) {
			// defer func() {
			// 	if err := recover(); err != nil {
			// 		logger.Errorf("Mqtt handler paniced with error: %s", err)
			// 		logger.Error(string(debug.Stack()))
			// 		// panic(err)
			// 	}
			// }()
			output := &Output{}
			retain := msg.Retained()
			topic := msg.Topic()
			qos := int(msg.Qos())
			messageID := int(msg.MessageID())
			output.Topic = topic
			output.Retained = retain
			output.QoS = qos
			output.Duplicate = msg.Duplicate()
			output.MessageID = messageID

			switch handlerSettings.ValueType {
			case "String":
				output.StringValue = string(msg.Payload())
				logger.Debugf("Mqtt Subscriber received a string message on topic [%s] retain [%s] qos [%s] messageID [%d] string value [%s]",
					topic, strconv.FormatBool(retain), strconv.FormatInt(int64(qos), 10), messageID, output.StringValue)
				break
			case "Base64":
				encoded := base64.StdEncoding.EncodeToString([]byte(msg.Payload()))
				output.StringValue = encoded
				logger.Debugf("Mqtt Subscriber received a base64 message on topic [%s] retain [%s] qos [%s] messageID [%s] string value [%s]",
					topic, strconv.FormatBool(retain), strconv.FormatInt(int64(qos), 10), messageID, output.StringValue)
				break
			default:
				var d map[string]interface{}
				err := json.Unmarshal(msg.Payload(), &d)
				if err != nil {
					logger.Errorf(fmt.Sprintf("Mqtt message contains an invalid json text: %s", err.Error()))
				}
				output.JSONValue = d
				logger.Debugf("Mqtt Subscriber received a JSON message on topic [%s] retain [%s] qos [%s] messageID [%s] string value [%s]",
					topic, strconv.FormatBool(retain), strconv.FormatInt(int64(qos), 10), messageID, string(msg.Payload()))
			}
			logger.Debug("Mqtt Subscriber about to trigger event")
			_, err := mqttHandler.handler.Handle(context.Background(), output)
			if err != nil {
				logger.Errorf("Mqtt Subscriber got error triggering event [%s]", err)
			}
		})
		if token.Error() != nil {
			// check the error types...if its connection level we can retry
			logger.Errorf("Mqtt subscriber was unable to subscribe due to [%s]", token.Error().Error())
		}
	}
	go t.OnReconnect()
	return nil
}

// Stop implements util.Managed.Stop
func (t *MqttTrigger) Stop() error {
	//unsubscribe from topic
	for _, mqttHandler := range t.clients {
		mqttClient := *mqttHandler.client
		// mqttClient.Disconnect(0)
		mqttClient.Disconnect(1000)
		logger.Infof("Mqtt Subscriber for topic [%s] stopped", mqttHandler.settings.Topic)
	}
	close(mqttconnection.OnConnectNotifier)
	return nil
}

// OnReconnect ...
func (t *MqttTrigger) OnReconnect() {
	if <-mqttconnection.OnConnectNotifier {
		logger.Debug("Mqtt subscriber is resubscribing")
		t.Start()
	}
	// for range mqttconnection.OnConnectNotifier {
	// 	fmt.Println("##########   OnReconnect   ########")
	// 	t.Start()
	// }
}
