package action

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Input struct {
	MailchimpConnection connection.Manager     `md:"Connection"`
	Resource            string                 `md:"Resource"`
	Action              string                 `md:"Action"`
	Input               map[string]interface{} `md:"input"`
}
type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection": i.MailchimpConnection,
		"Resource":   i.Resource,
		"Action":     i.Action,
		"input":      i.Input,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.MailchimpConnection, err = coerce.ToConnection(values["Connection"])
	if err != nil {
		return err
	}

	i.Resource, err = coerce.ToString(values["Resource"])
	if err != nil {
		return err
	}
	i.Action, err = coerce.ToString(values["Action"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return err
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
	return err
}
