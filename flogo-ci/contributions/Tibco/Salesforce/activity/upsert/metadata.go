package upsert

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

type Input struct {
	SalesforceConnection connection.Manager     `md:"connectionName"`
	ObjectName           string                 `md:"objectName"`
	Input                map[string]interface{} `md:"input"`
	ExternalIdFieldName  string                 `md:"externalIdFieldName"`
	AllOrNone            bool                   `md:"allOrNone"`
}
type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"connectionName":      i.SalesforceConnection,
		"objectName":          i.ObjectName,
		"input":               i.Input,
		"externalIdFieldName": i.ExternalIdFieldName,
		"allOrNone":           i.AllOrNone,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.SalesforceConnection, err = sfconnection.GetSharedConfiguration(values["connectionName"])
	if err != nil {
		return err
	}
	i.ObjectName, err = coerce.ToString(values["objectName"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	i.ExternalIdFieldName, err = coerce.ToString(values["externalIdFieldName"])
	if err != nil {
		return err
	}
	i.AllOrNone, err = coerce.ToBool(values["allOrNone"])
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
