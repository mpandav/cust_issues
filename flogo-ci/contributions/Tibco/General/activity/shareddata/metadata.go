package shareddata

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Scope string `md:"scope"`
	Op    string `md:"operation,allowed(GET,SET,DELETE)"` // The operation (get or set), 'get' is the default
	Type  string `md:"type"`                              // The data type of the shared value, default is 'any'
}

type Input struct {
	Key   string                 `md:"key""`
	input map[string]interface{} `md:"input"` // The value of the shared attribute
}

type Output struct {
	Output interface{} `md:"output"` // The value of the shared attribute
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"key":   i.Key,
		"input": i.input,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Key, err = coerce.ToString(values["key"])
	if err != nil {
		return err
	}
	i.input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return nil
}
