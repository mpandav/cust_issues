package sendmessage

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input struct for activity input
type Settings struct {
	EMSConManager   connection.Manager `md:"connection"`
	DestinationType string             `md:"destinationType"`
	Destination     string             `md:"settingDestination"`
	DeliveryDelay   int64              `md:"deliveryDelay"`
}

type Input struct {
	MessageBody       string                 `md:"message"`
	Destination       string                 `md:"destination"`
	Headers           map[string]interface{} `md:"headers"`
	MessageProperties map[string]interface{} `md:"messageProperties"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"message":           i.MessageBody,
		"destination":       i.Destination,
		"headers":           i.Headers,
		"messageProperties": i.MessageProperties,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.MessageBody, err = coerce.ToString(values["message"])
	if err != nil {
		return err
	}
	i.Destination, err = coerce.ToString(values["destination"])
	if err != nil {
		return err
	}
	i.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}
	i.MessageProperties, err = coerce.ToObject(values["messageProperties"])
	if err != nil {
		return err
	}
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
