package delete

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strings"

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
	return &MyActivity{logger: log.ChildLogger(ctx.Logger(), "Snowflake-activity-delete"), activityName: "delete"}, nil
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

	//3. Get query from input
	query := input.Query
	if len(strings.TrimSpace(query)) == 0 {
		return false, snowflake.GetError(snowflake.NoSQLSpecified, activity.activityName)
	}
	activity.logger.Infof(snowflake.GetMessage(snowflake.InputQuery, query))

	//4. Get data from input mapper
	inputData := input.Input
	inputParams, err := snowflake.GetInputData(inputData, activity.logger, activity.activityName)
	if err != nil {
		return false, snowflake.GetError(snowflake.FailedInputProcess, activity.activityName, err.Error())
	}

	//5. Prepare and Execute SQL
	activity.logger.Infof(snowflake.GetMessage(snowflake.ExecutionFlowInfo, "Executing SQL Query"))
	result, err := snowflake.ExecutePreparedDelete(query, inputParams, db, activity.activityName, activity.logger)
	if err != nil {
		return false, snowflake.GetError(snowflake.FailedExecuteSQL, activity.activityName, err.Error())
	}
	activity.logger.Infof(snowflake.GetMessage(snowflake.ExecutionFlowInfo, "SQL Query execution successful"))

	//6. Logging result at debug level
	out, err := coerce.ToString(result)
	if err != nil {
		return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
	}
	activity.logger.Debugf(snowflake.GetMessage(snowflake.ActivityOutput, out))

	//7. Setting result to context output
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(result)
	err = json.Unmarshal(reqBodyBytes.Bytes(), &output.Output)
	if err != nil {
		return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, snowflake.GetError(snowflake.DefaultError, activity.activityName, err.Error())
	}

	return true, nil

	///////////////////

	//removeing the Logout method as we are maintaing the connector cache
	//defer snowflake.Logout(db, log, snowflakeConn.Name)
}
