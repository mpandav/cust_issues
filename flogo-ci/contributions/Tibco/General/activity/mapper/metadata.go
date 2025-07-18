package mapper

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	Input interface{} `md:"input"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"input": i.Input,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	i.Input = values["input"]
	return nil
}

type Output struct {
	Output interface{} `md:"output"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error

	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return nil
}
