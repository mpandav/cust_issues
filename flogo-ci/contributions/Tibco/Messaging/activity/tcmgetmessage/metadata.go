package tcmgetmessage

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/tibco/flogo-messaging/src/app/Messaging/connector/tcm"
)

type Settings struct {
	ConnectionManager connection.Manager `md:"tcmConnection,required"`
	Matcher           string             `md:"matcher"`
	DurableName       string             `md:"durableName"`
	Destination       string             `md:"destination"`
}

func (o *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"tcmConnection": o.ConnectionManager,
		"matcher":       o.Matcher,
		"destination":   o.Destination,
		"durableName":   o.DurableName,
	}
}

func (o *Settings) FromMap(values map[string]interface{}) error {
	var err error
	o.ConnectionManager, err = tcm.GetSharedConfiguration(values["tcmConnection"])
	if err != nil {
		return err
	}

	o.Matcher, err = coerce.ToString(values["matcher"])
	if err != nil {
		return err
	}

	o.Destination, err = coerce.ToString(values["destination"])
	if err != nil {
		return err
	}

	o.DurableName, err = coerce.ToString(values["durableName"])
	if err != nil {
		return err
	}
	return err
}

type Input struct {
	Timeout int `md:"timeout"`
}

func (o *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"timeout": o.Timeout,
	}
}

func (o *Input) FromMap(values map[string]interface{}) error {

	var err error

	o.Timeout, err = coerce.ToInt(values["timeout"])
	if err != nil {
		return err
	}
	return nil
}

type MessageMetadata struct {
	Id            int64 `json:"messageId"`
	DeliveryCount int64 `json:"deliveryAttempt"`
}

type Output struct {
	Message  map[string]interface{} `md:"message,required"`
	Metadata MessageMetadata        `md:"metadata"`
}

func (o *Output) ToMap() map[string]interface{} {
	md, _ := coerce.ToObject(o.Metadata)
	return map[string]interface{}{
		"message":  o.Message,
		"metadata": md,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.Message, err = coerce.ToObject(values["message"])
	if err != nil {
		return err
	}

	md, err := coerce.ToObject(values["metadata"])
	if err != nil {
		return err
	}
	o.Metadata = MessageMetadata{}
	o.Metadata.Id, _ = coerce.ToInt64(md["messageId"])
	o.Metadata.DeliveryCount, _ = coerce.ToInt64(md["deliveryAttempt"])

	return nil
}
