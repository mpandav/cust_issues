package command

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	snowflakedb "github.com/tibco/wi-snowflake/src/app/Snowflake/connector/connection"
)

// Input corresponds to activity.json inputs
type Input struct {
	Connection connection.Manager     `md:"Snowflake Connection,required"`
	Command    string                 `md:"command,required"`
	Input      map[string]interface{} `md:"input"`
}

// Output corresponds to activity.json outputs
type Output struct {
	Output map[string]interface{} `md:"output,required"`
}

// ToMap converts Input struct to map
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Snowflake Connection": i.Connection,
		"command":              i.Command,
		"input":                i.Input,
	}
}

// FromMap converts a map to Input struct
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = snowflakedb.GetSharedConfiguration(values["Snowflake Connection"])
	if err != nil {
		return err
	}

	i.Command, err = coerce.ToString(values["command"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	return nil
}

// ToMap converts Output struct to map
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

// FromMap converts a map to Output struct
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}

	return nil
}
