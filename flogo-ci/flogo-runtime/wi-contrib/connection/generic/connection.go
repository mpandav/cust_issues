package generic

import (
	"errors"
	"fmt"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression"
	"github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/data/resolve"
)

var resolver = resolve.NewCompositeResolver(map[string]resolve.Resolver{
	".":        &resolve.ScopeResolver{},
	"env":      &resolve.EnvResolver{},
	"property": &property.Resolver{},
})

type Connection struct {
	name    string
	id      string
	configs map[string]interface{}
}

func (conn *Connection) GetName() string {
	return conn.configs["name"].(string)
}

func (conn *Connection) GetId() string {
	return conn.id
}

func (conn *Connection) GetSetting(name string) interface{} {
	return conn.configs[name]
}

func (conn *Connection) Settings() map[string]interface{} {
	return conn.configs
}


func init() {
	expression.SetScriptFactoryCreator(script.NewExprFactory)
}

func NewConnection(connection interface{}) (*Connection, error) {

	connectionObject, _ := coerce.ToObject(connection)
	if connectionObject == nil {
		return nil, errors.New("Connection object is nil")
	}

	conn := &Connection{}
	cName, ok := connectionObject["title"].(string)
	if !ok {
		cName, _ = connectionObject["name"].(string)
	}
	conn.name = cName

	conn.id = connectionObject["id"].(string)

	settings, err := coerce.ToArray(connectionObject["settings"])
	if err == nil {
		conn.configs = make(map[string]interface{}, len(settings))
		for _, v := range settings {
			val, ok := v.(map[string]interface{})
			if ok {
				name := val["name"]
				value := val["value"]
				finaVal, err := resolveValue(value)
				if err != nil {
					return nil, fmt.Errorf("Failed to resolve [%s]", value)
				}
				conn.configs[name.(string)] = finaVal
			}

		}
	} else {
		newSettings, _ := coerce.ToObject(connectionObject["settings"])
		if newSettings != nil {
			conn.configs = make(map[string]interface{}, len(newSettings))
			for k, v := range newSettings {
				finaVal, err := resolveValue(v)
				if err != nil {
					return nil, fmt.Errorf("Failed to resolve [%s]", v)
				}
				conn.configs[k] = finaVal
			}
		} else {
			conn.configs = make(map[string]interface{}, len(connectionObject))
			for k, v := range connectionObject {
				finaVal, err := resolveValue(v)
				if err != nil {
					return nil, fmt.Errorf("Failed to resolve [%s]", v)
				}
				conn.configs[k] = finaVal
			}
		}
	}
	return conn, nil
}

func resolveValue(val interface{}) (interface{}, error) {
	strVal, ok := val.(string)
	if ok && len(strVal) > 0 && strVal[0] == '=' {
		return resolver.Resolve(strVal[1:], nil)

	}

	return val, nil

}
