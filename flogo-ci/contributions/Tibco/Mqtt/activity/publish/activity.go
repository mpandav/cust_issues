package publish

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/activity"
	mqttconnection "github.com/tibco/wi-mqtt/src/app/Mqtt/connector/connection"
)

var activityMd = activity.ToMetadata(&Input{})

// MqttActivity is a stub for your Activity implementation
type MqttActivity struct {
	metadata *activity.Metadata
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MqttActivity{}, nil
}

func init() {
	_ = activity.Register(&MqttActivity{}, New)
}

// Metadata implements activity.Activity.Metadata
func (a *MqttActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *MqttActivity) Eval(context activity.Context) (done bool, err error) {
	logger := context.Logger()
	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	logger.Infof("Executing Mqtt Publish activity - [%s]", context.Name())

	mqttConfigManager, ok := input.Connection.GetConnection().(*mqttconnection.MqttConfigManager)
	if !ok {
		return false, activity.NewError("Unable to get Mqtt config manager", "", nil)
	}
	_, err = mqttConfigManager.GetMqttClient()
	if err != nil {
		return false, activity.NewError("Mqtt publisher failed to get client", "", err)
	}

	var content interface{}
	switch input.ValueType {
	case "String":
		content = input.StringValue
		break
	case "Base64":
		content, err = base64.StdEncoding.DecodeString(input.StringValue)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Invalid Base64 string %v", err), "MQTT-PUB-4003", nil)
		}
		break
	default:
		data, err := json.Marshal(input.JSONValue)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Invalid json content %v", err), "MQTT-PUB-4004", nil)
		}
		content = string(data)
	}
	topic := input.Topic
	var qos byte
	qos = byte(input.QOS)
	retain := input.Retain
	logger.Debugf("Mqtt Publisher publishing message with parms: Topic [%s] Retain [%t] QoS [%d] ValueType [%s] Message [%+v]\n", topic, retain, input.QOS, input.ValueType, content)
	token := mqttConfigManager.Client.Publish(topic, qos, retain, content)
	token.Wait()
	err = token.Error()
	if err != nil {
		logger.Debugf("Error occurred while publishing Mqtt message: Error [%+v]", err)
		return false, activity.NewError(fmt.Sprintf("Error occurred while publishing Mqtt message: Error [%s]", err.Error()), "MQTT-PUB-4005", nil)
	}
	return true, nil
}
