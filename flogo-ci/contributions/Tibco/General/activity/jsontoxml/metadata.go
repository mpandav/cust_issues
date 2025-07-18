package jsontoxml

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	JsonString string `md:"jsonString"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"jsonString": i.JsonString,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {

	var err error
	i.JsonString, err = coerce.ToString(values["jsonString"])
	if err != nil {
		return err
	}

	return nil
}

type Output struct {
	XmlString interface{} `md:"xmlString"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"xmlString": o.XmlString,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.XmlString, err = coerce.ToAny(values["xmlString"])
	if err != nil {
		return err
	}

	return nil
}
