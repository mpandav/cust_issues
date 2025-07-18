package publish

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	azureservicebusconnection "github.com/tibco/wi-azservicebus/src/app/AzureServiceBus/connector/connection"
)

type Input struct {
	AzureServiceBusConnection connection.Manager     `md:"Connection"`
	EntityType                string                 `md:"entityType"`
	EntityName                string                 `md:"entityName"`
	Input                     map[string]interface{} `md:"input"`
	Timeout                   int                    `md:"Timeout"`
}

type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection": i.AzureServiceBusConnection,
		"entityType": i.EntityType,
		"entityName": i.EntityName,
		"input":      i.Input,
		"Timeout":    i.Timeout,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.AzureServiceBusConnection, err = azureservicebusconnection.GetSharedConfiguration(values["Connection"])
	if err != nil {
		return err
	}
	i.EntityType, err = coerce.ToString(values["entityType"])
	if err != nil {
		return err
	}
	i.EntityName, err = coerce.ToString(values["entityName"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	i.Timeout, err = coerce.ToInt(values["Timeout"])
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
