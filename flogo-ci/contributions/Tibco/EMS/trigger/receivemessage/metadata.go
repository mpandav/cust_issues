package receivemessage

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type Settings struct {
	Connection connection.Manager `md:"connection"`
}

type HandlerSettings struct {
	AckMode                 string `md:"ackMode"`
	Destination             string `md:"destination"`
	DestinationType         string `md:"destinationType"`
	ProcessingMode          string `md:"processingMode"`
	DurableSubscriber       bool   `md:"durableSubscriber"`
	SharedDurableSubscriber bool   `md:"sharedDurableSubscriber"`
	SubscriptionName        string `md:"subscriptionName"`
}

// Output struct for trigger output
type Output struct {
	Message           string                 `md:"message"`
	Headers           map[string]interface{} `md:"headers"`
	MessageProperties interface{}            `md:"messageProperties"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"message":           o.Message,
		"headers":           o.Headers,
		"messageProperties": o.MessageProperties,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.Message, err = coerce.ToString(values["message"])
	if err != nil {
		return err
	}
	o.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}
	o.MessageProperties, err = coerce.ToObject(values["messageProperties"])
	if err != nil {
		return err
	}

	return nil
}
