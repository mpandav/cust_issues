package error

import (
	"github.com/project-flogo/core/data/coerce"
)

type Output struct {
	Activity string                 `md:"activity"`
	Message  string                 `md:"message"`
	Data     map[string]interface{} `md:"data"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"activity": o.Activity,
		"message":  o.Message,
		"data":     o.Data,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Activity, err = coerce.ToString(values["activity"])
	if err != nil {
		return err
	}
	o.Message, err = coerce.ToString(values["message"])
	if err != nil {
		return err
	}
	o.Data, err = coerce.ToObject(values["data"])
	if err != nil {
		return err
	}
	return nil
}
