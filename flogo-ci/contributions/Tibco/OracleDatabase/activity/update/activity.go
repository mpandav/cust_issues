package update

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-oracledb/src/app/OracleDatabase/activity/helper"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

// New creates a new activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "OracleDatabase-update"), activityName: "update"}, nil
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

	activity.logger.Infof(helper.GetMessage(helper.ActivityStart, activity.activityName))

	input := &Input{}
	output := &Output{}

	//Get Input Object
	err = context.GetInputObject(input)
	if err != nil {
		return false, helper.GetError(helper.FailedInputObject, activity.activityName, err.Error())
	}

	//Get db from input connection
	db := input.Connection.GetConnection().(*sql.DB)

	//Get query from input
	query := input.Query

	if len(strings.TrimSpace(query)) == 0 {
		return false, helper.GetError(helper.SpecifySQL, activity.activityName)
	}
	activity.logger.Infof(helper.GetMessage(helper.InputQuery, query))

	if !strings.HasPrefix(strings.ToUpper(query), "UPDATE") {
		return false, helper.GetError(helper.FailedExecuteSQL, activity.activityName, errors.New("Not a valid SQL Update statement"))
	}

	//Get data from input mapper
	inputData := input.Input
	inputParams, err := helper.GetInputData(inputData, activity.logger)
	if err != nil {
		return false, helper.GetError(helper.FailedInputProcess, activity.activityName, err.Error())
	}

	//Prepare and Execute SQL
	activity.logger.Infof(helper.GetMessage(helper.ExecutionFlowInfo, "Executing SQL Query"))
	result, err := helper.PreparedUpdateOrDelete(db, query, inputParams, activity.logger)
	if err != nil {
		return false, helper.GetError(helper.FailedExecuteSQL, activity.activityName, err.Error())
	}
	activity.logger.Infof(helper.GetMessage(helper.ExecutionFlowInfo, "SQL Query execution successful"))

	//Logging result at debug level
	out, err := coerce.ToString(result)
	if err != nil {
		return false, helper.GetError(helper.DefaultError, activity.activityName, err.Error())
	}
	activity.logger.Debugf(helper.GetMessage(helper.ActivityOutput, out))

	//Setting result to context output
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(result)
	err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
	if err != nil {
		return false, helper.GetError(helper.DefaultError, activity.activityName, err.Error())
	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, helper.GetError(helper.DefaultError, activity.activityName, err.Error())
	}

	return true, nil
}
