package mapper

import (
	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&MapperActivity{}, New)
}

func New(actx activity.InitContext) (activity.Activity, error) {
	return &MapperActivity{}, nil
}

// MapperActivity is an Activity that is used to log a message to the console
type MapperActivity struct {
}

func (a *MapperActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *MapperActivity) Eval(context activity.Context) (done bool, err error) {

	actinput := &Input{}
	err = context.GetInputObject(actinput)
	if err != nil {
		return false, err
	}

	output := &Output{
		Output: actinput.Input,
	}
	context.SetOutputObject(output)
	return true, nil
}
