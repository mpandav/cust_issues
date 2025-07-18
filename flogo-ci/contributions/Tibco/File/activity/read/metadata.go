package read

import (
	"github.com/project-flogo/core/data/coerce"
)

// Input corresponds to activity.json inputs
type Input struct {
	ReadAs   string                 `md:"readAs,required"`
	Compress string                 `md:"compress,required"`
	Input    map[string]interface{} `md:"input"`
}

// Output corresponds to activity.json outputs
type Output struct {
	Output map[string]interface{} `md:"output,required"`
}

// ToMap converts Input struct to map
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"readAs":   i.ReadAs,
		"compress": i.Compress,
		"input":    i.Input,
	}
}

// FromMap converts a map to Input struct
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.ReadAs, err = coerce.ToString(values["readAs"])
	if err != nil {
		return err
	}

	i.Compress, err = coerce.ToString(values["compress"])
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
