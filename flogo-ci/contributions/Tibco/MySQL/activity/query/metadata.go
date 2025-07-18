package query

import (
	"encoding/json"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	mysqlconnection "github.com/tibco/wi-mysql/src/app/MySQL/connector/connection"
)

// Settings struct for activity input
type Settings struct {
}

// Input struct for activity input
type Input struct {
	Connection connection.Manager     `md:"Connection,required"`
	Query      string                 `md:"Query"`
	Input      map[string]interface{} `md:"input"`
	ManaulMode bool                   `md:"manualmode,required"`
	Fields     []interface{}          `md:"Fields"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection": i.Connection,
		"Query":      i.Query,
		"Input":      i.Input,
		"ManualMoe":  i.ManaulMode,
		"Fields":     i.Fields,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.Connection, err = coerce.ToConnection(values["Connection"])
	if err != nil {
		return err
	}

	i.Query, err = coerce.ToString(values["Query"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	i.ManaulMode, err = coerce.ToBool(values["manualmode"])
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
	Output *mysqlconnection.ResultSet `md:"Output"`
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

	jsonoutput, err := json.Marshal(values["Output"])
	if err != nil {
		return err
	}
	var outputrecord *mysqlconnection.ResultSet
	err = json.Unmarshal(jsonoutput, outputrecord)
	if err != nil {
		return err
	}

	o.Output = outputrecord

	return nil
}
