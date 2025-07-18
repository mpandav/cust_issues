package protobuf2json

import (
	"github.com/project-flogo/core/data/coerce"
)

// Settings ...
type Settings struct {
	ProtoFile            map[string]interface{} `md:"protoFile,required"`
	MessageTypeName      string                 `md:"messageTypeName,required"`
	IncludeDefaultValues bool                   `md:"includeDefaultValues,required"`
}

// ToMap ...
func (s *Settings) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"protoFile":            s.ProtoFile,
		"messageTypeName":      s.MessageTypeName,
		"includeDefaultValues": s.IncludeDefaultValues,
	}
}

// FromMap ...
func (s *Settings) FromMap(values map[string]interface{}) error {
	var err error
	s.ProtoFile, err = coerce.ToObject(values["protoFile"])
	if err != nil {
		return err
	}
	s.MessageTypeName, err = coerce.ToString(values["messageTypeName"])
	if err != nil {
		return err
	}
	s.IncludeDefaultValues, err = coerce.ToBool(values["includeDefaultValues"])
	if err != nil {
		return err
	}
	return nil
}

// Input ...
type Input struct {
	ProtoMessage string `md:"protoMessage,required"`
}

// ToMap ...
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"protoMessage": i.ProtoMessage,
	}
}

// FromMap ...
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.ProtoMessage, err = coerce.ToString(values["protoMessage"])
	if err != nil {
		return err
	}
	return nil
}

// Output ...
type Output struct {
	JSONMessage map[string]interface{} `md:"jsonMessage"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"jsonMessage": o.JSONMessage,
	}
}

// FromMap ...
func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.JSONMessage, err = coerce.ToObject(values["jsonMessage"])
	if err != nil {
		return err
	}
	return nil
}
