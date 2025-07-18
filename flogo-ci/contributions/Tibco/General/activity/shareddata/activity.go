package shareddata

import (
	"fmt"

	"github.com/project-flogo/core/app"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/flow/instance"
)

const (
	scopeApp string = "Application"
	opGet    int    = 1
	opSet    int    = 2
	opDel    int    = 3
)

func init() {
	_ = activity.Register(&SharedDataAct{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Output{})

// Activity is a Counter Activity implementation
type SharedDataAct struct {
	scope string
	op    int
	dt    data.Type
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	act := &SharedDataAct{}

	act.scope = s.Scope

	if s.Op == "SET" {
		act.op = opSet
	} else if s.Op == "GET" {
		act.op = opGet
	} else {
		act.op = opDel
	}

	if s.Type != "" {
		var t data.Type
		if s.Type == "number" {
			t = data.TypeFloat64
		} else {
			t, err = data.ToTypeEnum(s.Type)
			if err != nil {
				return nil, err
			}
		}

		act.dt = t
	}
	return act, nil
}

// Metadata implements activity.Activity.Metadata
func (a *SharedDataAct) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *SharedDataAct) Eval(ctx activity.Context) (bool, error) {
	input := &Input{}
	err := ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	if len(input.Key) <= 0 {
		return false, fmt.Errorf("input field key is required")
	}

	key := "TIB_" + a.scope + ":" + input.Key

	switch a.op {
	case opGet:
		ctx.Logger().Debugf("getting data by key [%s] for scope [%s]", input.Key, a.scope)
		scopeData, exist := getData(a.scope, key, ctx)
		ctx.Logger().Debugf("got data by key [%s] for scope [%s]", input.Key, a.scope)
		//Validation
		if a.dt == data.TypeObject {
			scopeData, err = coerce.ToType(scopeData, data.TypeAny)
		} else {
			scopeData, err = coerce.ToType(scopeData, a.dt)
		}
		if err != nil {
			return false, err
		}
		output := &Output{Output: map[string]interface{}{"exist": exist, "data": scopeData}}
		err = ctx.SetOutputObject(output)
		if err != nil {
			return false, err
		}
	case opSet:
		var inputData interface{}
		var ok bool
		inputData, ok = input.input["data"]
		if !ok {
			return false, fmt.Errorf("setting data is empty for key [%s]", input.Key)
		}
		ctx.Logger().Debugf("setting data for key [%s]", input.Key)
		if a.dt == data.TypeObject {
			inputData, err = coerce.ToType(inputData, data.TypeAny)
		} else {
			inputData, err = coerce.ToType(inputData, a.dt)
		}
		if err != nil {
			return false, err
		}
		err = setData(a.scope, key, inputData, ctx)
		if err != nil {
			return false, err
		}
		ctx.Logger().Debugf("set data for key [%s]", input.Key)
	case opDel:
		ctx.Logger().Debugf("deleting data for key [%s]", input.Key)
		_, exist := getData(a.scope, key, ctx)
		delData(a.scope, key)
		output := &Output{Output: map[string]interface{}{"exist": exist}}
		err = ctx.SetOutputObject(output)
		ctx.Logger().Debugf("deleted data for key [%s]", input.Key)
	}
	return true, nil
}

func (a *SharedDataAct) isObject() bool {
	return a.dt == data.TypeObject
}

func getData(scope string, name string, ctx activity.Context) (interface{}, bool) {
	if scope == scopeApp {
		return app.GetValue(name)
	} else {
		scopeInst := ctx.ActivityHost().Scope()
		inst, ok := scopeInst.(*instance.Instance)
		if ok {
			val, exist := inst.GetMasterScope().GetValue(name)
			if attr, ok := val.(*data.Attribute); ok {
				if attr != nil {
					return attr.Value(), exist
				} else {
					return nil, exist
				}
			}
			return val, exist
		}
	}
	return nil, false
}

func setData(scope string, name string, value interface{}, ctx activity.Context) error {
	if scope == scopeApp {
		app.SetValue(name, value)
	} else {
		scopeInst := ctx.ActivityHost().Scope()
		inst, ok := scopeInst.(*instance.Instance)
		if ok {
			return inst.GetMasterScope().SetValue(name, value)
		}
	}
	return nil
}

func delData(scope string, name string) error {
	if scope == scopeApp {
		app.DeleteValue(name)
	}
	return nil
}
