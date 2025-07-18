package publish

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	mqttconnection "github.com/tibco/wi-mqtt/src/app/Mqtt/connector/connection"
)

// Input ...
type Input struct {
	Connection  connection.Manager     `md:"Connection,required"`
	Topic       string                 `md:"topic"`
	Retain      bool                   `md:"retain"`
	QOS         int                    `md:"qos,allowed(0,1,2)"`
	ValueType   string                 `md:"valueType,required"`
	StringValue string                 `md:"stringValue"`
	JSONValue   map[string]interface{} `md:"jsonValue"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection":  i.Connection,
		"topic":       i.Topic,
		"retain":      i.Retain,
		"qos":         i.QOS,
		"valueType":   i.ValueType,
		"stringValue": i.StringValue,
		"jsonValue":   i.JSONValue,
	}
}

// FromMap ...
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = mqttconnection.GetSharedConfiguration(values["Connection"])
	if err != nil {
		return err
	}
	i.Topic, err = coerce.ToString(values["topic"])
	if err != nil {
		return err
	}
	i.Retain, err = coerce.ToBool(values["retain"])
	if err != nil {
		return err
	}
	i.QOS, err = coerce.ToInt(values["qos"])
	if err != nil {
		return err
	}
	i.ValueType, err = coerce.ToString(values["valueType"])
	if err != nil {
		return err
	}
	i.StringValue, err = coerce.ToString(values["stringValue"])
	if err != nil {
		return err
	}
	i.JSONValue, err = coerce.ToObject(values["jsonValue"])
	if err != nil {
		return err
	}
	return nil
}
