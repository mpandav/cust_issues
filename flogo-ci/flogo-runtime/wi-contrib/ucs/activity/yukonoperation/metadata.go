package yukonoperation

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
	Input      []InputDetails         `json:"input"`
}

type InputDetails struct {
	LookupCondition interface{}            `json:"lookupCondition"`
	InputData       map[string]interface{} `md:"inputData"`
}

type OutputDetails struct {
	Action     string                   `json:"action"`
	DataObject string                   `json:"dataObject"`
	Results    []map[string]interface{} `json:"results"`
}
type Output struct {
	Output OutputDetails `md:"output"`
}

func (i *Input) FromMap(values map[string]interface{}) error {
	// var err error
	dataObject, err := coerce.ToString(values["dataObject"])
	if err != nil {
		return err
	}
	i.DataObject = dataObject

	var lookupCondition interface{}
	if values["filter"] != nil {
		filter, err := coerce.ToObject(values["filter"])
		if err != nil {
			return err
		}
		i.Filter = filter
		lookupCondition = filter["lookupCondition"]
	} else {
		lookupCondition = nil
	}

	inputData, err := coerce.ToObject(values["inputData"])
	if err != nil {
		return err
	}

	input := InputDetails{LookupCondition: lookupCondition, InputData: inputData}
	inputs := []InputDetails{input}
	s := make([]InputDetails, len(inputs))
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
		"output": output,
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
