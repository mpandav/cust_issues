package subscriber

import (
	"encoding/json"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	GooglePubSubConnection connection.Manager `md:"googleConnection,required"`
}

type HandlerSettings struct {
	SubscriptionId         string `md:"subscriptionId,required"`
	FlowControlMode        bool   `md:"flowControlMode"`
	MaxOutstandingMessages int    `md:"maxOutstandingMessages"`
	MessageDataFormat      string `md:"messageDataFormat,required,allowed(String,JSON)"`
}

type MessageMetadata struct {
	Id              string            `json:"messageId"`
	DeliveryAttempt int               `json:"deliveryAttempt"`
	Attributes      map[string]string `json:"messageAttributes"`
}

// Output struct for trigger output
type Output struct {
	MessageData     interface{}     `json:"message"`
	MessageMetaData MessageMetadata `json:"metadata"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	var msgData map[string]interface{}
	msg, _ := json.Marshal(o.MessageMetaData)
	_ = json.Unmarshal(msg, &msgData)
	return map[string]interface{}{
		"message":  o.MessageData,
		"metadata": msgData,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {
	o.MessageMetaData = MessageMetadata{}
	msgObject, err := coerce.ToObject(values["metadata"])
	if err != nil {
		return err
	}
	o.MessageMetaData.Attributes, _ = coerce.ToParams(msgObject["messageAttributes"])
	o.MessageMetaData.Id, _ = coerce.ToString(msgObject["messageId"])
	o.MessageMetaData.DeliveryAttempt, _ = coerce.ToInt(msgObject["deliveryAttempt"])
	o.MessageData = values["message"]
	return nil
}
