package rediscommand

import (

	///connection import remaining
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	redisconnection "github.com/tibco/wi-redis/src/app/Redis/connector/connection"
)

type Input struct {
	RedisConnection connection.Manager     `md:"Connection"`
	Group           string                 `md:"Group"`
	Command         string                 `md:"Command"`
	Input           map[string]interface{} `md:"input"`
}

type Output struct {
	Output map[string]interface{} `md:"Output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Connection": i.RedisConnection,
		"Group":      i.Group,
		"Command":    i.Command,
		"input":      i.Input,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.RedisConnection, err = redisconnection.GetSharedConfiguration(values["Connection"])
	if err != nil {
		return err
	}

	i.Group, err = coerce.ToString(values["Group"])
	if err != nil {
		return err
	}
	i.Command, err = coerce.ToString(values["Command"])
	if err != nil {
		return err
	}
	i.Input, err = coerce.ToObject(values["input"])
	if err != nil {
		return err
	}
	return err
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Output": o.Output,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Output, err = coerce.ToObject(values["Output"])
	if err != nil {
		return err
	}
	return err
}
