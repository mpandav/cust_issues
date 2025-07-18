package tcmpublisher

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/tibco/flogo-messaging/src/app/Messaging/connector/tcm"
)

type Input struct {
	Connection  connection.Manager     `md:"tcmConnection,required"`
	Destination string                 `md:"destination"`
	Message     map[string]interface{} `md:"message"`
}

func (o *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"tcmConnection": o.Connection,
		"destination":   o.Destination,
		"message":       o.Message,
	}
}

func (o *Input) FromMap(values map[string]interface{}) error {
	var err error
	o.Destination, err = coerce.ToString(values["destination"])
	if err != nil {
		return err
	}

	o.Connection, err = tcm.GetSharedConfiguration(values["tcmConnection"])
	if err != nil {
		return err
	}

	o.Message, err = coerce.ToObject(values["message"])
	if err != nil {
		return err
	}

	return nil
}
