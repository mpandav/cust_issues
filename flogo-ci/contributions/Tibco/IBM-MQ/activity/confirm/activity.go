package confirm

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

var activityMd = activity.ToMetadata()

// var logger = log.GetLogger("flogo-ibmmq-confirm")
var versionPrinted = false

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
	var log log.Logger = context.Logger()
	//attrs := make(map[string]*data.Attribute)
	attrs := make(map[string]interface{})
	attrs["confirm"] = true
	context.ActivityHost().Reply(attrs, nil)
	log.Debugf("%s IBM-MQ Confirm processed", context.Name())
	return true, nil
}
