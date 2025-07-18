package get

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Queue        string                 `md:"queue"`
	CorellId     string                 `md:"corellId"`
	GmoConvert   bool                   `md:"gmoConvert"`
	MsgId        string                 `md:"msgId"`
	MaxSize      int32                  `md:"maxSize"`
	WaitInterval int32                  `md:"waitInterval"`
	ValueType    string                 `md:"valueType"`
	Connection   map[string]interface{} `md:"Connection"`
}

type Input struct {
	Queue        string `md:"queue"`
	CorellId     string `md:"corellId"`
	MsgId        string `md:"msgId"`
	WaitInterval int32  `md:"waitInterval"`
}

type Output struct {
	MessageJson       map[string]interface{} `md:"MessageJson"`
	MessageProperties map[string]interface{} `md:"MessageProperties"`
	MQMD              map[string]interface{} `md:"MQMD"`
	Message           map[string]interface{} `md:"Message"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"queue":        i.Queue,
		"corellId":     i.CorellId,
		"msgId":        i.MsgId,
		"waitInterval": i.WaitInterval,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.Queue, err = coerce.ToString(values["queue"])
	if err != nil {
		return err
	}
	i.CorellId, err = coerce.ToString(values["corellId"])
	if err != nil {
		return err
	}
	i.MsgId, err = coerce.ToString(values["msgId"])
	if err != nil {
		return err
	}
	i.WaitInterval, err = coerce.ToInt32(values["waitInterval"])
	if err != nil {
		return err
	}
	return nil

}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Message":           o.Message,
		"MessageJson":       o.MessageJson,
		"MessageProperties": o.MessageProperties,
		"MQMD":              o.MQMD,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Message, err = coerce.ToObject(values["Message"])
	if err != nil {
		return err
	}
	o.MessageJson, err = coerce.ToObject(values["MessageJson"])
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
