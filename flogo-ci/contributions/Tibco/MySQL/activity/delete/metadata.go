package delete

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Settings struct for activity input
type Settings struct {
}

// Input struct for activity input
type Input struct {
	Connection      connection.Manager     `md:"Connection,required"`
	DeleteStatement string                 `md:"DeleteStatement"`
	Input           map[string]interface{} `md:"input"`
	ManualMode      bool                   `md:"manualmode,required"`
	Fields          []interface{}          `md:"Fields"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection":      i.Connection,
		"DeleteStatement": i.DeleteStatement,
		"Input":           i.Input,
		"ManualMode":      i.ManualMode,
		"Fields":          i.Fields,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.Connection, err = coerce.ToConnection(values["Connection"])
	if err != nil {
		return err
	}

	i.DeleteStatement, err = coerce.ToString(values["DeleteStatement"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	i.ManualMode, err = coerce.ToBool(values["manualmode"])
	if err != nil {
		return err
	}

	i.Fields, err = coerce.ToArray(values["Fields"])
	if err != nil {
		return err
	}

	return nil
}

// Output struct for activity input
type Output struct {
	Output map[string]interface{} `md:"Output"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Output": o.Output,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values)
	if err != nil {
		return err
	}

	return nil
}
