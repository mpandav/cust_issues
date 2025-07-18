package salesforce

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

type HandlerSettings struct {
	Connection          connection.Manager `"Connection Name"`
	ObjectName          string             `md:"Object Name"`
	SubscriberType      string             `md:"subscriberType"`
	AutoCreatePushTopic bool               `md:"autoCreatePushTopic"`
	ChannelName         string             `md:"channelName"`
	ReplayID            int                `md:"replayID"`
	Query               string             `md:"Query"`
}

type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"output": o.Output,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["output"])
	if err != nil {
		return err
	}
	return nil
}
