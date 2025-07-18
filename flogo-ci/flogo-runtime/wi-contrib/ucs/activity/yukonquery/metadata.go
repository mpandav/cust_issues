package yukonquery

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	YukonConnection         connection.Manager `md:"connection"`
	Action                  string             `md:"action"`
	RequiresLookupCondition bool               `md:"requiresLookupCondition"`
	RequiresInputData       bool               `md:"requiresInputData"`
}

type Input struct {
	DataObject string                 `md:"dataObject"`
	Action     string                 `md:"action"`
	Filter     map[string]interface{} `md:"filter"`
	Input      []InputData            `json:"input"`
}

type InputData struct {
	QueryInput      QueryInputDetails `json:"inputData"`
	LookupCondition interface{}       `json:"lookupCondition"`
}

type OutputDetails struct {
	Action     string                   `json:"action"`
	DataObject string                   `json:"dataObject"`
	Results    []map[string]interface{} `json:"results"`
}
type Output struct {
	Output         OutputDetails `md:"output"`
	FieldSelection []interface{} `md:"fieldSelection"`
}

type QueryInputDetails struct {
	Select    []string    `json:"select"`
	From      string      `json:"from"`
	Condition interface{} `json:"condition"`
}

func (i *Input) FromMap(values map[string]interface{}) error {
	// var err error
	dataObject, err := coerce.ToString(values["dataObject"])
	if err != nil {
		return err
	}
	i.DataObject = dataObject

	filter, err := coerce.ToObject(values["filter"])
	if err != nil {
		return err
	}
	lookupCondition := filter["lookupCondition"]

	queryInputDetails := QueryInputDetails{From: dataObject, Condition: lookupCondition}

	input := InputData{QueryInput: queryInputDetails, LookupCondition: lookupCondition}

	inputs := []InputData{input}
	s := make([]InputData, len(inputs))
	for i, v := range inputs {
		s[i] = v
	}
	i.Input = s
	return nil
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"input": i.Input,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	fieldSelection, err := coerce.ToArray(values["fieldSelection"])
	if err != nil {
		return err
	}

	o.FieldSelection = fieldSelection

	output, err := coerce.ToObject(values["output"])
	if err != nil {
		return err
	}

	o.Output = OutputDetails{}
	o.Output.Action, _ = coerce.ToString(output["action"])
	o.Output.DataObject, _ = coerce.ToString(output["dataObject"])
	rawResults, _ := coerce.ToArray(output["results"])
	results := make([]map[string]interface{}, len(rawResults))
	for index, value := range rawResults {
		val, ok := value.(map[string]interface{})
		if ok {
			results[index] = val
		}
	}
	o.Output.Results = results
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	output, _ := coerce.ToObject(o.Output)
	return map[string]interface{}{
		"output":         output,
		"fieldSelection": o.FieldSelection,
	}
}

func (s *Settings) FromMap(values map[string]interface{}) error {
	var err error
	s.YukonConnection, err = coerce.ToConnection(values["connection"])
	if err != nil {
		return err
	}
	s.Action, err = coerce.ToString(values["action"])
	if err != nil {
		return err
	}
	s.RequiresLookupCondition, err = coerce.ToBool(values["requiresLookupCondition"])
	if err != nil {
		return err
	}
	s.RequiresInputData, err = coerce.ToBool(values["requiresInputData"])
	if err != nil {
		return err
	}
	return nil
}

func (s *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"connection": s.YukonConnection,
		"action":     s.Action,
	}
}
