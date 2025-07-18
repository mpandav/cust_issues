package callprocedure

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/tibco/wi-oracledb/src/app/OracleDatabase/connector/oracledb"
)

// Input corresponds to activity.json inputs
type Input struct {
	Connection    connection.Manager     `md:"Oracle Database Connection,required"`
	CallProcedure string                 `md:"CallProcedure"`
	Input         map[string]interface{} `md:"input"`
	FieldsInfo    string                 `md:"FieldsInfo"`
}

// Output corresponds to activity.json outputs
type Output struct {
	Output map[string]interface{} `md:"Output,required"`
}

// ToMap converts Input struct to map
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Oracle Database Connection": i.Connection,
		"CallProcedure":              i.CallProcedure,
		"input":                      i.Input,
		"FieldsInfo":                 i.FieldsInfo,
	}
}

// FromMap converts a map to Input struct
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Connection, err = oracledb.GetSharedConfiguration(values["Oracle Database Connection"])
	if err != nil {
		return err
	}

	i.CallProcedure, err = coerce.ToString(values["CallProcedure"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	i.FieldsInfo, err = coerce.ToString(values["FieldsInfo"])
	if err != nil {
		return err
	}

	return nil
}

// ToMap converts Output struct to map
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Output": o.Output,
	}
}

// FromMap converts a map to Output struct
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["Output"])
	if err != nil {
		return err
	}

	return nil
}
