package selectObjectContent

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	Connection connection.Manager `md:"connection,required"` // Select an AWS Connection
}

// Input schema for S3 select object activity
type Input struct {
	InputSerialization  string                 `md:"inputSerialization"`
	CompressionType     string                 `md:"compressionType,required"`
	OutputSerialization string                 `md:"outputSerialization"`
	Input               map[string]interface{} `md:"input"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"inputSerialization":  i.InputSerialization,
		"compressionType":     i.CompressionType,
		"outputSerialization": i.OutputSerialization,
		"input":               i.Input,
	}
}

// FromMap coerce values to params
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.InputSerialization, err = coerce.ToString(values["inputSerialization"])
	if err != nil {
		return err
	}
	i.CompressionType, err = coerce.ToString(values["compressionType"])
	if err != nil {
		return err
	}
	i.OutputSerialization, err = coerce.ToString(values["outputSerialization"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return nil
}

// Output for S3 select object activity
type Output struct {
	Output              string                 `md:"output"`
	OutputSerialization string                 `md:"outputSerialization"`
	Error               map[string]interface{} `md:"error"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output":              o.Output,
		"outputSerialization": o.OutputSerialization,
		"error":               o.Error,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToString(values["output"])
	if err != nil {
		return err
	}
	o.OutputSerialization, err = coerce.ToString(values["outputSerialization"])
	if err != nil {
		return err
	}
	o.Error, err = coerce.ToObject(values["error"])
	if err != nil {
		return err
	}
	return nil
}
