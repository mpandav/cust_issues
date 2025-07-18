package publisher

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input struct for activity input
type Settings struct {
	GooglePubSubConnection connection.Manager `md:"googleConnection,required"`
	Topic                  string             `md:"topicName,required"`
	MessageOrdering        bool               `md:"messageOrdering"`
	MessageDataFormat      string             `md:"messageDataFormat,required,allowed(String,JSON)"`
}

type Input struct {
	MessageData        interface{}       `json:"message"`
	MessageAttributes  map[string]string `json:"messageAttributes"`
	MessageOrderingKey string            `json:"messageOrderingKey"`
	TopicId            string            `json:"topicId"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"message":            i.MessageData,
		"messageAttributes":  i.MessageAttributes,
		"messageOrderingKey": i.MessageOrderingKey,
		"topicId":            i.TopicId,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {

	i.MessageAttributes, _ = coerce.ToParams(values["messageAttributes"])
	if i.MessageAttributes == nil || len(i.MessageAttributes) == 0 {
		i.MessageAttributes = make(map[string]string)
	}
	i.MessageOrderingKey, _ = coerce.ToString(values["messageOrderingKey"])
	i.MessageData, _ = values["message"]
	i.TopicId, _ = coerce.ToString(values["topicId"])
	return nil
}

// Output struct for activity output
type Output struct {
	MessageId string `md:"messageId"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"messageId": o.MessageId,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.MessageId, err = coerce.ToString(values["messageId"])
	if err != nil {
		return err
	}

	return nil
}
