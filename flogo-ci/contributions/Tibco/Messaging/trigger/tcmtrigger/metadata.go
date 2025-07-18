package tcmtrigger

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	DurableSub     bool   `md:"durableSub"`
	DurableName    string `md:"durableName"`
	DurableType    string `md:"durableType"`
	Destination    string `md:"destination"`
	AckMode        string `md:"ackMode"`
	ProcessingMode string `md:"processingMode"`
	Matcher        string `md:"matcher"`
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
