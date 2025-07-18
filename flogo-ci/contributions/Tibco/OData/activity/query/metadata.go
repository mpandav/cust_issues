package query

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	ODataConnection connection.Manager `md:"oDataConnection"`
}

// Input struct for activity input
type Input struct {
	ODataURI     string                 `md:"oDataURI"`
	QueryOptions map[string]interface{} `md:"queryOptions"`
	Parameters   map[string]interface{} `md:"parameters"`
	Headers      map[string]interface{} `md:"headers"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"oDataURI":     i.ODataURI,
		"queryOptions": i.QueryOptions,
		"parameters":   i.Parameters,
		"headers":      i.Headers,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.ODataURI, err = coerce.ToString(values["oDataURI"])
	if err != nil {
		return err
	}

	i.QueryOptions, err = coerce.ToObject(values["queryOptions"])
	if err != nil {
		return err
	}

	i.Parameters, err = coerce.ToObject(values["parameters"])
	if err != nil {
		return err
	}

	i.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}

	return nil
}

// Output struct for activity output
type Output struct {
	ResponseType    string                 `md:"responseType"`
	ResponseBody    interface{}            `md:"responseBody"`
	ResponseHeaders map[string]interface{} `md:"responseHeaders"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"responseType":    o.ResponseType,
		"responseBody":    o.ResponseBody,
		"responseHeaders": o.ResponseHeaders,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {

	var err error

	o.ResponseType, err = coerce.ToString(values["responseType"])
	if err != nil {
		return err
	}

	o.ResponseBody, err = coerce.ToObject(values["responseBody"])
	if err != nil {
		return err
	}

	o.ResponseHeaders, err = coerce.ToObject(values["responseHeaders"])
	if err != nil {
		return err
	}

	return nil
}
