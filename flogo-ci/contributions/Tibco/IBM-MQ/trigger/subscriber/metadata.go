package subscriber

import "github.com/project-flogo/core/data/coerce"

type Settings struct {
	ClientConfirm   bool  `md:"clientconfirm"`
	PollingInterval int32 `md:"pollinginterval"`
}

type HandlerSettings struct {
	Topic        string                 `md:"topic"`
	Dynamictopic string                 `md:"dynamictopic"`
	Durable      bool                   `md:"durable"`
	Durablename  string                 `md:"durablename"`
	Newpubsonly  bool                   `md:"newpubsonly"`
	ValueType    string                 `md:"valueType"`
	Connection   map[string]interface{} `md:"Connection"`
}

type Output struct {
	MessageJson       map[string]interface{} `md:"MessageJson"`
	Message           map[string]interface{} `md:"Message"`
	MessageProperties map[string]interface{} `md:"MessageProperties"`
	MQMD              map[string]interface{} `md:"MQMD"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"MessageJson":       o.MessageJson,
		"Message":           o.Message,
		"MessageProperties": o.MessageProperties,
		"MQMD":              o.MQMD,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.MessageJson, err = coerce.ToObject(values["MessageJson"])
	if err != nil {
		return err
	}
	o.Message, err = coerce.ToObject(values["Message"])
	if err != nil {
		return err
	}
	o.MessageProperties, err = coerce.ToObject(values["MessageProperties"])
	if err != nil {
		return err
	}
	o.MQMD, err = coerce.ToObject(values["MQMD"])
	if err != nil {
		return err
	}
	return nil
}
