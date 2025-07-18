package rename

import (
	"bytes"
	"encoding/json"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	flogofile "github.com/tibco/flogo-files/src/app/File/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

// New creates a new activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "File-activity-rename"), activityName: "rename"}, nil
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
	output := &Output{}

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

	renameFileOutput, err := flogofile.RenameFile(inputParams, createNonExistingDir, activity.activityName)
	if err != nil {
		return false, flogofile.GetError(flogofile.DefaultError, activity.activityName, err.Error())
	}

	//Logging output at debug level
	out, err := coerce.ToString(renameFileOutput)
	if err != nil {
		return false, flogofile.GetError(flogofile.DefaultError, activity.activityName, err.Error())
	}
	activity.logger.Debugf(flogofile.GetMessage(flogofile.ActivityOutput, out))

	//set the output
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(renameFileOutput)
	err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
	if err != nil {
		return false, flogofile.GetError(flogofile.DefaultError, activity.activityName, err.Error())
	}
	err = context.SetOutputObject(output)
	if err != nil {
		return false, flogofile.GetError(flogofile.DefaultError, activity.activityName, err.Error())
	}
	activity.logger.Infof(flogofile.GetMessage(flogofile.ActivityEnd, activity.activityName))

	return true, nil
}
