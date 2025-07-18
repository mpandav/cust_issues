package xmltojson

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	XmlString string `md:"xmlString"`
	Ordered   bool   `md:"ordered"`
	TypeCast  bool   `md:"typeCast"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"xmlString": i.XmlString,
		"ordered":   i.Ordered,
		"typeCast":  i.TypeCast,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {

	var err error
	i.XmlString, err = coerce.ToString(values["xmlString"])
	if err != nil {
		return err
	}

	i.Ordered, _ = coerce.ToBool(values["ordered"])
	if err != nil {
		return err
	}
	i.TypeCast, _ = coerce.ToBool(values["typeCast"])
	if err != nil {
		return err
	}
	return nil
}

type Output struct {
	JsonObject interface{} `md:"jsonObject"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"jsonObject": o.JsonObject,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.JsonObject, err = coerce.ToAny(values["jsonObject"])
	if err != nil {
		return err
	}

	return nil
}
