package createjob

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

type Input struct {
	SalesforceConnection connection.Manager     `md:"Connection Name"`
	Input                map[string]interface{} `md:"input"`
	Operation            string                 `md:"operation"`
}
type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection Name": i.SalesforceConnection,
		"input":           i.Input,
		"operation":       i.Operation,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.SalesforceConnection, err = sfconnection.GetSharedConfiguration(values["Connection Name"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	i.Operation, err = coerce.ToString(values["operation"])
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
