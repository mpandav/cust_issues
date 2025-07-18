package snspublish

import (
	"encoding/json"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	snsutil "github.com/tibco/flogo-aws-sns/src/app/AmazonSNS/activity"
)

// Input ...
type Input struct {
	Connection            connection.Manager            `md:"connection,required"`
	MessageType           string                        `md:"messageType,required"`
	MessageAttributeNames []snsutil.MessageAttributeMap `md:"messageAttributeNames"`
	Input                 map[string]interface{}        `md:"input"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"connection":            i.Connection,
		"messageType":           i.MessageType,
		"messageAttributeNames": i.MessageAttributeNames,
		"input":                 i.Input,
	}
}

// FromMap coerce values to params
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = coerce.ToConnection(values["connection"])
	if err != nil {
		return err
	}
	i.MessageType, err = coerce.ToString(values["messageType"])
	if err != nil {
		return err
	}
	if values["messageAttributeNames"] == "" {
		messageAttributeNames := make([]snsutil.MessageAttributeMap, 0)
		i.MessageAttributeNames = messageAttributeNames
	} else {
		messageAttributeBytes, err := json.Marshal(values["messageAttributeNames"])
		if err != nil {
			return err
		}
		var messageAttributeNames []snsutil.MessageAttributeMap
		err = json.Unmarshal(messageAttributeBytes, &messageAttributeNames)
		if err != nil {
			return err
		}
		i.MessageAttributeNames = messageAttributeNames
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return nil
}

// Output ...
type Output struct {
	Output map[string]interface{} `md:"output"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return nil
}
