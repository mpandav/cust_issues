package callprocedure

import (
	"bytes"
	"database/sql"
	"encoding/json"
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
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "OracleDatabase-callprocedure"), activityName: "callprocedure"}, nil
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
	query := input.CallProcedure

	if len(strings.TrimSpace(query)) == 0 {
		return false, helper.GetError(helper.SpecifySQL, activity.activityName)
	}
	activity.logger.Debugf(helper.GetMessage(helper.InputQuery, query))

	// if !strings.HasPrefix(strings.ToUpper(query), "CALLPROCUDURE") {
	// 	return false, helper.GetError(helper.FailedExecuteSQL, activity.activityName, errors.New("Not a valid SQL Query statement"))
	// }

	//Get data from input mapper
	inputData2 := input.FieldsInfo
	inputData := input.Input
	activity.logger.Debugf(helper.GetMessage(helper.FieldsInfo, inputData2))
	activity.logger.Debugf(helper.GetMessage(helper.InputData, inputData))

	inputParams, err := helper.GetInputDataCall(inputData, inputData2, activity.logger)
	if err != nil {
		return false, helper.GetError(helper.FailedInputProcess, activity.activityName, err.Error())
	}
	activity.logger.Debugf(helper.GetMessage(helper.InputParams, inputParams))

	//Prepare and Execute SQL
	activity.logger.Infof(helper.GetMessage(helper.ExecutionFlowInfo, "Executing Call Procedure"))
	outputMap, err := helper.PreparedQueryCALL(db, query, inputParams, activity.logger)
	if err != nil {
		return false, helper.GetError(helper.FailedExecuteSQL, activity.activityName, err.Error())
	}
	activity.logger.Infof(helper.GetMessage(helper.ExecutionFlowInfo, "Call Procedure execution successful"))

	if len(outputMap) != 0 {
		out, err := coerce.ToString(outputMap)
		if err != nil {
			return false, helper.GetError(helper.DefaultError, activity.activityName, err.Error())
		}
		activity.logger.Debugf(helper.GetMessage(helper.ActivityOutput, out))
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(outputMap)
		err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
		if err != nil {
			return false, helper.GetError(helper.DefaultError, activity.activityName, err.Error())
		}
		err = context.SetOutputObject(output)
		if err != nil {
			return false, helper.GetError(helper.DefaultError, activity.activityName, err.Error())
		}
	}

	return true, nil
}
