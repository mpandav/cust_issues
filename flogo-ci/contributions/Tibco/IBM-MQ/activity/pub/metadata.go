package pub

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	Connection     map[string]interface{} `md:"Connection"`
	Topic          string                 `md:"topic"`
	Topicdynamic   string                 `md:"topicdynamic"`
	Retain         bool                   `md:"retain"`
	ContextSupport string                 `md:"contextSupport"`
	MessageType    string                 `md:"messageType"`
	ValueType      string                 `md:"valueType"`
}

type Input struct {
	Topic         string                 `md:"topic"`
	Topicdynamic  string                 `md:"topicdynamic"`
	MessageJson   map[string]interface{} `md:"MessageJson"`
	MessageString string                 `md:"MessageString"`
	Properties    map[string]interface{} `md:"properties"`
	MQMD          map[string]interface{} `md:"MQMD"`
}

type Output struct {
	Output map[string]interface{} `md:"Output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"topic":         i.Topic,
		"topicdynamic":  i.Topicdynamic,
		"MessageString": i.MessageString,
		"MessageJson":   i.MessageJson,
		"properties":    i.Properties,
		"MQMD":          i.MQMD,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Topic, err = coerce.ToString(values["topic"])
	if err != nil {
		return err
	}
	i.Topicdynamic, err = coerce.ToString(values["topicdynamic"])
	if err != nil {
		return err
	}
	i.MessageString, err = coerce.ToString(values["MessageString"])
	if err != nil {
		return err
	}
	i.MessageJson, err = coerce.ToObject(values["MessageJson"])
	if err != nil {
		return err
	}
	i.Properties, err = coerce.ToObject(values["properties"])
	if err != nil {
		return err
	}
	i.MQMD, err = coerce.ToObject(values["MQMD"])
	if err != nil {
		return err
	}
	return nil

}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Output": o.Output,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["Output"])
	if err != nil {
		return err
	}
	return nil
}
