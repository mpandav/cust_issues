package subscribe

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Settings ...
type Settings struct {
	MqttConnection connection.Manager `md:"mqttConnection,required"`
}

// HandlerSettings ...
type HandlerSettings struct {
	Topic      string `md:"topic"`
	QoS        int    `md:"qos"`
	ValueType  string `md:"valueType"`
	ShowWill   bool   `md:"showwill"`
	Will       string `md:"will"`
	WillTopic  string `md:"willtopic"`
	WillQoS    int    `md:"willqos"`
	WillRetain bool   `md:"willretain"`
}

// Output ...
type Output struct {
	Topic       string                 `md:"topic"`
	Retained    bool                   `md:"retained"`
	QoS         int                    `md:"qos"`
	Duplicate   bool                   `md:"duplicate"`
	MessageID   int                    `md:"messageID"`
	StringValue string                 `md:"stringValue"`
	JSONValue   map[string]interface{} `md:"jsonValue"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"topic":       o.Topic,
		"retained":    o.Retained,
		"qos":         o.QoS,
		"duplicate":   o.Duplicate,
		"messageID":   o.MessageID,
		"stringValue": o.StringValue,
		"jsonValue":   o.JSONValue,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Topic, err = coerce.ToString(values["topic"])
	if err != nil {
		return err
	}
	o.Retained, err = coerce.ToBool(values["retained"])
	if err != nil {
		return err
	}
	o.QoS, err = coerce.ToInt(values["qos"])
	if err != nil {
		return err
	}
	o.Duplicate, err = coerce.ToBool(values["duplicate"])
	if err != nil {
		return err
	}
	o.MessageID, err = coerce.ToInt(values["messageID"])
	if err != nil {
		return err
	}
	o.StringValue, err = coerce.ToString(values["stringValue"])
	if err != nil {
		return err
	}
	o.JSONValue, err = coerce.ToObject(values["jsonValue"])
	if err != nil {
		return err
	}
	return nil
}
