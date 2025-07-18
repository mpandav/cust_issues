package httpresponse

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&HTTPResponseActivity{}, New)
}

func New(actx activity.InitContext) (activity.Activity, error) {
	return &HTTPResponseActivity{}, nil
}

// HTTPResponseActivity is an Activity that is used to log a message to the console
type HTTPResponseActivity struct {
}

func (a *HTTPResponseActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *HTTPResponseActivity) Eval(context activity.Context) (done bool, err error) {

	actinput := &Input{}
	err = context.GetInputObject(actinput)
	if err != nil {
		return false, err
	}
	codeValue, _ := coerce.ToInt64(actinput.Responsecode)

	output := &Output{
		Response: actinput.Input,
		Code:     codeValue,
	}
	context.SetOutputObject(output)
	return true, nil
}
