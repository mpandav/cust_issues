package noacknowledge

import (
	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata()

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
	attrs := make(map[string]interface{})
	attrs["messageAck"] = false
	context.ActivityHost().Reply(attrs, nil)
	context.Logger().Info("Notifying service bus subscriber to not mark message complete.")
	return true, nil
}
