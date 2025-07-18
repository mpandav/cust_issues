package parsejson

import (
	"errors"

	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	Validate   bool   `md:"validate"`
	JsonString string `md:"jsonString"` //
}

const (
	ivJSONData   = "jsonString"
	ivValidate   = "validate"
	ovJSONObject = "jsonObject"
)

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		ivJSONData: i.JsonString,
		ivValidate: i.Validate,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {

	var err error
	i.JsonString, err = coerce.ToString(values[ivJSONData])
	if err != nil {
		return err
	}

	if i.JsonString == "" {
		return errors.New("JSON string empty")
	}

	i.Validate, _ = coerce.ToBool(values[ivValidate])

	return nil
}

type Output struct {
	JsonObject interface{} `md:"jsonObject"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		ovJSONObject: o.JsonObject,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.JsonObject, err = coerce.ToAny(values[ovJSONObject])
	if err != nil {
		return err
	}

	return nil
}
