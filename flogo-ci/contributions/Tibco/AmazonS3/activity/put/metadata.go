package put

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input schema for S3 Put activity
type Input struct {
	Connection  connection.Manager     `md:"connection,required"` // Select an AWS Connection
	ServiceName string                 `md:"serviceName,required"`
	PutType     string                 `md:"putType"`
	InputType   string                 `md:"inputType"`
	PreserveACL bool                   `md:"preserveACL"`
	Input       map[string]interface{} `md:"input"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"connection":  i.Connection,
		"serviceName": i.ServiceName,
		"putType":     i.PutType,
		"inputType":   i.InputType,
		"preserveACL": i.PreserveACL,
		"input":       i.Input,
	}
}

// FromMap coerce values to params
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = coerce.ToConnection(values["connection"])
	if err != nil {
		return err
	}
	i.ServiceName, err = coerce.ToString(values["serviceName"])
	if err != nil {
		return err
	}
	i.PutType, err = coerce.ToString(values["putType"])
	if err != nil {
		return err
	}
	i.InputType, err = coerce.ToString(values["inputType"])
	if err != nil {
		return err
	}
	i.PreserveACL, err = coerce.ToBool(values["preserveACL"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return nil
}

// Output for S3 Put activity
type Output struct {
	Output map[string]interface{} `md:"output"`
	Error  map[string]interface{} `md:"error"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
		"error":  o.Error,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	o.Error, err = coerce.ToObject(values["error"])
	if err != nil {
		return err
	}
	return nil
}
