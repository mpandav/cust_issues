package copy

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	flogofile "github.com/tibco/flogo-files/src/app/File/activity"
)

var activityMd = activity.ToMetadata(&Input{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

// New creates a new activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "File-activity-copy"), activityName: "copy"}, nil
}

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	logger       log.Logger
	activityName string
}

// Metadata implements activity.Activity.Metadata
func (*MyActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (activity *MyActivity) Eval(context activity.Context) (done bool, err error) {
	activity.logger.Infof(flogofile.GetMessage(flogofile.ActivityStart, activity.activityName))

	input := &Input{}

	//Get Input Object
	err = context.GetInputObject(input)
	if err != nil {
		return false, flogofile.GetError(flogofile.FailedInputObject, activity.activityName, err.Error())
	}

	//Get data from input mapper
	inputData := input.Input
	inputParams, err := flogofile.GetInputData(inputData, activity.logger)
	if err != nil {
		return false, flogofile.GetError(flogofile.FailedInputProcess, activity.activityName, err.Error())
	}

	createNonExistingDir := input.CreateNonExistingDir
	includeSubDirectories := input.IncludeSubDirectories

	err = flogofile.CopyFile(inputParams, createNonExistingDir, includeSubDirectories, activity.activityName, activity.logger)
	if err != nil {
		return false, flogofile.GetError(flogofile.DefaultError, activity.activityName, err.Error())
	}

	activity.logger.Infof(flogofile.GetMessage(flogofile.ActivityEnd, activity.activityName))

	return true, nil
}
