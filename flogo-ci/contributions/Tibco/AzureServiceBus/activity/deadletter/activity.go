package deadletter

import (
	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
}

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{}, nil
}

// Metadata implements activity.Activity.Metadata
func (a *MyActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *MyActivity) Eval(context activity.Context) (done bool, err error) {
	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	attrs := make(map[string]interface{})
	attrs["deadletter"] = map[string]string{
		"deadLetterReason":      input.DeadLetterReason,
		"deadLetterDescription": input.DeadLetterDescription,
	}
	context.ActivityHost().Reply(attrs, nil)
	context.Logger().Info("Notifying service bus subscriber to move the message to deadletter.")
	return true, nil
}
