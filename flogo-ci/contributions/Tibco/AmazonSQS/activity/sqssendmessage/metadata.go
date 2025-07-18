package sqssendmessage

import (
	"github.com/project-flogo/core/data/coerce"
)

// Input struct for activity input
type Settings struct {
	AWSConnection       	string     			   `md:"awsConnection"`
	QueueURL            	string                 `md:"queueUrl"`
	Delay             		int                    `md:"DelaySeconds"`
}

type Input struct {
	MessageBody         	string                 `md:"MessageBody"`
	MessageAttributeNames   []interface{} 		   `md:"MessageAttributeNames"`
	MessageAttributes       map[string]interface{} `md:"MessageAttributes"`
	MessageDeduplicationId	string				   `md:"MessageDeduplicationId"`
	MessageGroupId			string				   `md:"MessageGroupId"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"MessageBody":                  i.MessageBody,
		"MessageAttributeNames": 		i.MessageAttributeNames,
		"MessageAttributes":            i.MessageAttributes,
		"MessageDeduplicationId":		i.MessageDeduplicationId,
		"MessageGroupId":				i.MessageGroupId,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.MessageBody, err = coerce.ToString(values["MessageBody"])
	if err != nil {
		return err
	}

	i.MessageDeduplicationId, err = coerce.ToString(values["MessageDeduplicationId"])
	if err != nil {
		return err
	}

	i.MessageGroupId, err = coerce.ToString(values["MessageGroupId"])
	if err != nil {
		return err
	}

	i.MessageAttributeNames, err = coerce.ToArray(values["MessageAttributeNames"])
	if err != nil {
		return err
	}

	i.MessageAttributes, err = coerce.ToObject(values["MessageAttributes"])
	if err != nil {
		return err
	}

	return nil
}

// Output struct for activity output
type Output struct {
	Output map[string]interface{}		`md:"output"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return nil
}
