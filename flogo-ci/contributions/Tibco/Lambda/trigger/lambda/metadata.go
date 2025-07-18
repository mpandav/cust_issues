package lambda

import (
	"fmt"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/trigger"
	"github.com/tibco/wi-contrib/connection/generic"
)

// MyTrigger AWS Lambda trigger struct
type MyTrigger struct {
	metadata *trigger.Metadata
	config   *trigger.Config
	// TODO: Ask Tracy if this is normal: previously it was []*trigger.Handler
	handlers []trigger.Handler
}

// MyFactory AWS Lambda Trigger factory
type MyFactory struct {
	metadata *trigger.Metadata
}

// Settings ...
type Settings struct {
	Connection        connection.Manager `md:"ConnectionName,required"`
	ExecutionRoleName string             `md:"ExecutionRoleName"`
}

// ToMap ...
func (s *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"ConnectionName":    s.Connection,
		"ExecutionRoleName": s.ExecutionRoleName,
	}
}

// FromMap ...
func (s *Settings) FromMap(values map[string]interface{}) error {
	var err error
	var cManager connection.Manager
	_, ok := values["ConnectionName"].(map[string]interface{})
	if ok {
		cManager, err = handleLegacyConnection(values["ConnectionName"])
	} else {
		cManager, err = coerce.ToConnection(values["ConnectionName"])
	}
	s.Connection = cManager
	if err != nil {
		return err
	}
	s.ExecutionRoleName, err = coerce.ToString(values["ExecutionRoleName"])
	if err != nil {
		return err
	}
	return nil
}

// Output ...
type Output struct {
	Function     map[string]interface{} `md:"Function"`
	Context      map[string]interface{} `md:"Context"`
	Identity     map[string]interface{} `md:"Identity"`
	ClientApp    map[string]interface{} `md:"ClientApp"`
	EventPayload map[string]interface{} `md:"EventPayload"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Function":     o.Function,
		"Context":      o.Context,
		"Identity":     o.Identity,
		"ClientApp":    o.ClientApp,
		"EventPayload": o.EventPayload,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Function, err = coerce.ToObject(values["Function"])
	if err != nil {
		return err
	}
	o.Context, err = coerce.ToObject(values["Context"])
	if err != nil {
		return err
	}
	o.Identity, err = coerce.ToObject(values["Identity"])
	if err != nil {
		return err
	}
	o.ClientApp, err = coerce.ToObject(values["ClientApp"])
	if err != nil {
		return err
	}
	o.EventPayload, err = coerce.ToObject(values["EventPayload"])
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
