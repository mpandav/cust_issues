package insert

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input Structure
type Input struct {
	Connection      connection.Manager `md:"Connection,required"`
	InsertStatement string             `md:"InsertStatement"`
	QueryTimeout    int                `md:"querytimeout"`
	Manualmode      bool               `md:"manualmode"`
	// Query           string             `md:"Query"`
	// RuntimeQuery string                 `md:"RuntimeQuery"`
	InputParams map[string]interface{} `md:"input"`
	Fields      []interface{}          `md:"Fields"`
}

// ToMap Input interface
func (o *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection":      o.Connection,
		"InsertStatement": o.InsertStatement,
		"querytimeout":    o.QueryTimeout,
		"manualmode":      o.Manualmode,
		// "Query":           o.Query,
		// "RuntimeQuery": o.RuntimeQuery,
		"InputParams": o.InputParams,
		"Fields":      o.Fields,
	}
}

// FromMap Input interface
func (o *Input) FromMap(values map[string]interface{}) error {
	var err error

	o.InsertStatement, err = coerce.ToString(values["InsertStatement"])
	if err != nil {
		return err
	}

	o.QueryTimeout, err = coerce.ToInt(values["querytimeout"])
	if err != nil {
		return err
	}

	o.Manualmode, err = coerce.ToBool(values["manualmode"])
	if err != nil {
		return err
	}

	o.InputParams, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}

	o.Fields, err = coerce.ToArray(values["Fields"])
	if err != nil {
		return err
	}

	o.Connection, err = coerce.ToConnection(values["Connection"])
	if err != nil {
		return err
	}

	return nil

}

//Output struct
// type Output struct {
// 	Output *sqlserver.ResultSet `md:"Output"`
// }
//Output struct
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
// func (o *Output) FromMap(values map[string]interface{}) error {
// 	var err error

// 	jsonoutput, err := json.Marshal(values["Output"])
// 	if err != nil {
// 		return err
// 	}
// 	var outputrecord *sqlserver.ResultSet
// 	err = json.Unmarshal(jsonoutput, outputrecord)
// 	if err != nil {
// 		return err
// 	}

// 	o.Output = outputrecord

// 	return nil
// }

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values)
	if err != nil {
		return err
	}

	return nil
}
