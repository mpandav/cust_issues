package createstreams

import (
	"fmt"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/tibco/wi-contrib/connection/generic"
)

// Input schema for S3 Get activity
type Input struct {
	Connection                       connection.Manager     `md:"awsConnection,required"` // Select an AWS Connection
	StreamType                       string                 `md:"streamType,required"`
	DeliveryStreamType               string                 `md:"DeliveryStreamType,required"`
	DestinationType                  string                 `md:"DestinationType,required"`
	Input                            map[string]interface{} `md:"input"`
	DeliveryStreamConfiguration      map[string]interface{} `md:"DeliveryStreamConfiguration"`
	KinesisStreamSourceConfiguration map[string]interface{} `md:"KinesisStreamSourceConfiguration"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"awsConnection":                    i.Connection,
		"streamType":                       i.StreamType,
		"DeliveryStreamType":               i.DeliveryStreamType,
		"DestinationType":                  i.DestinationType,
		"input":                            i.Input,
		"DeliveryStreamConfiguration":      i.DeliveryStreamConfiguration,
		"KinesisStreamSourceConfiguration": i.KinesisStreamSourceConfiguration,
	}
}

// FromMap coerce values to params
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	var cManager connection.Manager
	_, ok := values["awsConnection"].(map[string]interface{})
	if ok {
		cManager, err = handleLegacyConnection(values["awsConnection"])
	} else {
		cManager, err = coerce.ToConnection(values["awsConnection"])
	}
	i.Connection = cManager
	if err != nil {
		return err
	}
	i.StreamType, err = coerce.ToString(values["streamType"])
	if err != nil {
		return err
	}
	i.DeliveryStreamType, err = coerce.ToString(values["DeliveryStreamType"])
	if err != nil {
		return err
	}
	i.DestinationType, err = coerce.ToString(values["DestinationType"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	i.DeliveryStreamConfiguration, err = coerce.ToObject(values["DeliveryStreamConfiguration"])
	if err != nil {
		return err
	}
	i.KinesisStreamSourceConfiguration, err = coerce.ToObject(values["KinesisStreamSourceConfiguration"])
	if err != nil {
		return err
	}
	return nil
}

type Output struct {
	Message map[string]interface{} `md:"Message"`
	Error   map[string]interface{} `md:"Error"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Message": o.Message,
		"Error":   o.Error,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Message, err = coerce.ToObject(values["Message"])
	if err != nil {
		return err
	}
	o.Error, err = coerce.ToObject(values["Error"])
	if err != nil {
		return err
	}
	return nil
}

func handleLegacyConnection(conn interface{}) (connection.Manager, error) {
	connectionObject, _ := coerce.ToObject(conn)
	if connectionObject == nil {
		return nil, fmt.Errorf("Connection object [%+v] invalid", conn)
	}
	id := connectionObject["id"].(string)
	cManager := connection.GetManager(id)
	if cManager == nil {
		connObject, err := generic.NewConnection(connectionObject)
		if err != nil {
			return nil, err
		}
		cManager, err = factory.NewManager(connObject.Settings())
		if err != nil {
			return nil, err
		}
		err = connection.RegisterManager(id, cManager)
		if err != nil {
			return nil, err
		}
	}
	return cManager, nil
}
