package callprocedure

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/execute"
	tdvconnection "github.com/tibco/flogo-tdv/src/app/TDV/connector/connection"
)

// Input Structure
type Input struct {
	QueryName   string                 `md:"QueryName"`
	Catalog     string                 `md:"Catalog"`
	Schema      string                 `md:"Schema"`
	Procedure   string                 `md:"Procedure"`
	InputParams map[string]interface{} `md:"input"`
	Fields      []interface{}          `md:"Fields"`
}

// ToMap Input interface
func (o *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"QueryName":   o.QueryName,
		"Catalog":     o.Catalog,
		"Schema":      o.Schema,
		"Procedure":   o.Procedure,
		"InputParams": o.InputParams,
		"Fields":      o.Fields,
	}
}

// FromMap Input interface
func (o *Input) FromMap(values map[string]interface{}) error {
	var err error
	o.QueryName, err = coerce.ToString(values["Queryname"])
	if err != nil {
		return err
	}
	o.Catalog, err = coerce.ToString(values["Catalog"])
	if err != nil {
		return err
	}
	o.Schema, err = coerce.ToString(values["Schema"])
	if err != nil {
		return err
	}
	o.Procedure, err = coerce.ToString(values["Procedure"])
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
	return nil

}

// GenerateQuery generates Query string to envoke stored procedure and returns necessary metadata
func (input *Input) GenerateQuery() (string, []execute.InputParamMetadata, []int, []string, int, error) {
	var err error
	CallStatement := "{CALL "
	if input.Catalog == "" {
		return "", nil, nil, nil, 0, fmt.Errorf("received empty Catalog")
	}
	if input.Catalog != "All" {
		CallStatement = CallStatement + input.Catalog + "."
	}
	if input.Schema == "" {
		return "", nil, nil, nil, 0, fmt.Errorf("received empty Schema")
	}
	if input.Schema != "Public" {
		CallStatement = CallStatement + input.Schema + "."
	}
	if input.Procedure == "" {
		return "", nil, nil, nil, 0, fmt.Errorf("missing Procedure name")
	}
	CallStatement = CallStatement + input.Procedure + "("

	var inputParamMetadata []execute.InputParamMetadata
	paramPosition := 1
	inputParamPositions := []int{}
	outParamPositions := []int{}
	var cursors []string
	for _, v := range input.Fields {
		m, ok := v.(map[string]interface{})
		if !ok {
			return "", nil, nil, nil, 0, fmt.Errorf("error fetching metadata: %v", err)
		}
		if m["Type"] != "STRUCT" {
			switch dir := m["Direction"]; dir {
			case "IN":
				CallStatement = CallStatement + "?,"
				inputParamMetadata = append(inputParamMetadata, execute.InputParamMetadata{ParamName: fmt.Sprint(m["FieldName"]), ParamType: m["Type"].(string)})
				inputParamPositions = append(inputParamPositions, paramPosition)
				paramPosition = paramPosition + 1

			case "INOUT":
				CallStatement = CallStatement + "?,"
				inputParamPositions = append(inputParamPositions, paramPosition)
				inputParamMetadata = append(inputParamMetadata, execute.InputParamMetadata{ParamName: fmt.Sprint(m["FieldName"]), ParamType: m["Type"].(string)})
				paramPosition = paramPosition + 1

			case "RETURNVALUE":
				CallStatement = CallStatement + "?,"
				outParamPositions = append(outParamPositions, paramPosition)
				paramPosition = paramPosition + 1

			default:
				//we are getting OUT for params inside Cursor Do nothing for them
			}
		}
		if m["Type"] == "STRUCT" && m["Direction"] == "RETURNVALUE" {
			cursors = append(cursors, m["FieldName"].(string))
		}
	}

	if paramPosition > 1 {
		CallStatement = strings.TrimSuffix(CallStatement, ",")
	}

	CallStatement = CallStatement + ")};"
	return CallStatement, inputParamMetadata, inputParamPositions, cursors, paramPosition, nil
}

//Output struct
type Output struct {
	Output *tdvconnection.ProcResultSet `md:"Output"`
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
	var outputrecord *tdvconnection.ProcResultSet
	err = json.Unmarshal(jsonoutput, outputrecord)
	if err != nil {
		return err
	}

	o.Output = outputrecord

	return nil
}

type Settings struct {
	Connection connection.Manager `md:"Connection,required"`
}
