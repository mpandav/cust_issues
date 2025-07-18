package invokesync

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input ...
type Input struct {
	Connection connection.Manager `md:"ConnectionName,required"`
	ARN        string             `md:"arn"`
	Payload    interface{}        `md:"payload"`
	LambdaARN  string             `md:"LambdaARN"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ConnectionName": i.Connection,
		"arn":            i.ARN,
		"payload":        i.Payload,
		"LambdaARN":      i.LambdaARN,
	}
}

// FromMap coerce values to params
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = coerce.ToConnection(values["ConnectionName"])
	if err != nil {
		return err
	}
	i.ARN, err = coerce.ToString(values["arn"])
	if err != nil {
		return err
	}

	i.Payload = values["payload"]

	i.LambdaARN, err = coerce.ToString(values["LambdaARN"])
	if err != nil {
		return err
	}
	return nil
}

// Output ...
type Output struct {
	Status int64                  `md:"status"`
	Result map[string]interface{} `md:"result"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"status": o.Status,
		"result": o.Result,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Status, err = coerce.ToInt64(values["status"])
	if err != nil {
		return err
	}
	o.Result, err = coerce.ToObject(values["result"])
	if err != nil {
		return err
	}
	return nil
}
