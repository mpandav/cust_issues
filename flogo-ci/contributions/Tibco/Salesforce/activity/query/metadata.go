package query

import (
	"fmt"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

type Input struct {
	SalesforceConnection connection.Manager `md:"Connection Name"`
	ObjectName           string             `md:"Object Name"`
	QueryType            string             `md:"queryType"`
	Query                string             `md:"query"`
}
type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (i *Input) ToMap() map[string]interface{} {
	fmt.Println("In ip ToMap")
	return map[string]interface{}{
		"Connection Name": i.SalesforceConnection,
		"Object Name":     i.ObjectName,
		"queryType":       i.QueryType,
		"query":           i.Query,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.SalesforceConnection, err = sfconnection.GetSharedConfiguration(values["Connection Name"])
	if err != nil {
		return err
	}
	i.ObjectName, err = coerce.ToString(values["Object Name"])
	if err != nil {
		return err
	}
	i.QueryType, err = coerce.ToString(values["queryType"])
	if err != nil {
		return err
	}
	i.Query, err = coerce.ToString(values["query"])
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
