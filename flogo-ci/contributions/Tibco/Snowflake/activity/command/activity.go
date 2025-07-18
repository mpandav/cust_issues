package command

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	snowflake "github.com/tibco/wi-snowflake/src/app/Snowflake/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

// New creates a new activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "Snowflake-activity-command"), activityName: "command"}, nil
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
	activity.logger.Infof(snowflake.GetMessage(snowflake.ActivityStart, activity.activityName))

	input := &Input{}
	output := &Output{}

	//1. Get Input Object
	err = context.GetInputObject(input)
	if err != nil {
		return false, snowflake.GetError(snowflake.FailedInputObject, activity.activityName, err.Error())
	}

	//2. Get db from input connection
	db := input.Connection.GetConnection().(*sql.DB)

	// 3. Get data from input mapper
	inputData := input.Input
	inputParams, err := snowflake.GetCommandInput(inputData, activity.logger)
	if err != nil {
		return false, snowflake.GetError(snowflake.FailedInputProcess, activity.activityName, err.Error())
	}

	command := input.Command

	if command == "PUT" {
		commandOutput, err := snowflake.ExecutePutCommand(inputData, inputParams, db, activity.activityName, activity.logger)
		if err != nil {
			return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
		}
		//6. Logging result at debug level
		out, err := coerce.ToString(commandOutput)
		if err != nil {
			return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
		}
		activity.logger.Debugf(snowflake.GetMessage(snowflake.ActivityOutput, out))

		//7. Setting result to context output
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(commandOutput)
		err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
		if err != nil {
			return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
		}
	} else if command == "COPY INTO" {
		commandOutput, err := snowflake.ExecuteCopyIntoCommand(inputData, inputParams, db, activity.activityName, activity.logger)
		if err != nil {
			return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
		}
		//6. Logging result at debug level
		out, err := coerce.ToString(commandOutput)
		if err != nil {
			return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
		}
		activity.logger.Debugf(snowflake.GetMessage(snowflake.ActivityOutput, out))

		//7. Setting result to context output
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(commandOutput)
		err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
		if err != nil {
			return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
		}
	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
	}

	return true, nil
}
