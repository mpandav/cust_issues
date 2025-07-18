package presign

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input schema for S3 Get activity
type Input struct {
	Connection        connection.Manager     `md:"connection,required"`    // Select an AWS Connection
	OperationType     string                 `md:"operationType,required"` // Select one operation type
	ExpirationTimeSec int64                  `md:"expirationTimeSec,required"`
	Input             map[string]interface{} `md:"input"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"connection":        i.Connection,
		"operationType":     i.OperationType,
		"expirationTimeSec": i.ExpirationTimeSec,
		"input":             i.Input,
	}
}

// FromMap coerce values to params
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = coerce.ToConnection(values["connection"])
	if err != nil {
		return err
	}
	i.OperationType, err = coerce.ToString(values["operationType"])
	if err != nil {
		return err
	}
	i.ExpirationTimeSec, err = coerce.ToInt64(values["expirationTimeSec"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return nil
}

// Output for S3 Get activity
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
