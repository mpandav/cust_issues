package httpresponse

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	Trigger      string                 `md:"trigger"`
	Responsecode string                 `md:"responsecode"`
	Input        map[string]interface{} `md:"input"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"trigger":      i.Trigger,
		"responsecode": i.Responsecode,
		"input":        i.Input,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.Trigger, err = coerce.ToString(values["trigger"])
	if err != nil {
		return err
	}

	i.Responsecode, err = coerce.ToString(values["responsecode"])
	if err != nil {
		return err
	}

	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return nil
}

type Output struct {
	Code     int64                  `md:"code"`
	Response map[string]interface{} `md:"response"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":     o.Code,
		"response": o.Response,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error

	o.Code, err = coerce.ToInt64(values["code"])
	if err != nil {
		return err
	}

	o.Response, err = coerce.ToObject(values["response"])
	if err != nil {
		return err
	}
	return nil
}
