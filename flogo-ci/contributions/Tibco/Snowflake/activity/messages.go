package snowflake

import (
	"fmt"

	"github.com/project-flogo/core/activity"
)

// Constants
const (
	CategoryName = "Snowflake"
	//Debug Messages
	ActivityInput       = 1001
	ActivityOutput      = 1002
	ExecuteSQLWithParam = 1003

	//Info Messages
	ActivityStart     = 2001
	InputQuery        = 2002
	NoRowsFound       = 2003
	ExecutionFlowInfo = 2004

	//Error Messages
	DefaultError            = 4001
	ConnectionNotConfigured = 4002
	NoSQLSpecified          = 4003
	QueryPreparationFailed  = 4004
	SQLExecutionFailed      = 4005
	FailedGettingColumnInfo = 4006
	ScanningRowFailed       = 4007
	FailedGettingColumnType = 4008
	SpecifyInput            = 4009
	FailedInputObject       = 4010
	FailedInputProcess      = 4011
	FailedExecuteSQL        = 4012
)

var messages = make(map[int]string)

func init() {
	//Debug Messages
	messages[ActivityInput] = "Activity Input: %s"
	messages[ActivityOutput] = "Activity Output: %s"
	messages[ExecuteSQLWithParam] = "Query [%s] Parameters [%v]"

	//Info Messages
	messages[ActivityStart] = "Executing Snowflake %s activity"
	messages[InputQuery] = "Query Entered: %s"
	messages[NoRowsFound] = "No rows found"
	messages[ExecutionFlowInfo] = "%s"

	//Error Messages
	messages[DefaultError] = "%s"
	messages[ConnectionNotConfigured] = "Snowflake connection not configured"
	messages[NoSQLSpecified] = "No SQL Stetement specified"
	messages[QueryPreparationFailed] = "Query Preparation failed. Query: %s, Error: %s"
	messages[SQLExecutionFailed] = "SQL execution failed of %s activity. Error: %s"
	messages[FailedGettingColumnInfo] = "Error getting column information : %s"
	messages[ScanningRowFailed] = "Error scanning rows : %s"
	messages[FailedGettingColumnType] = "Error getting column types : %s"
	messages[SpecifyInput] = "No input arguments for query specified"
	messages[FailedInputObject] = "Failed to get Input object from context. Error: [%s]"
	messages[FailedInputProcess] = "Failed to process input arguments : %s"
	messages[FailedExecuteSQL] = "Failed to execute SQL Query : %s"
}

// GetError to create activity error
func GetError(errConst int, activityName string, parms ...interface{}) *activity.Error {
	errCode := CategoryName + "-" + activityName + "-" + string(errConst)
	return activity.NewError(GetMessage(errConst, parms...), errCode, nil)
}

// GetMessage to get message
func GetMessage(msgConst int, parms ...interface{}) string {
	if parms != nil {
		return fmt.Sprintf(messages[msgConst], parms...)
	}
	return messages[msgConst]
}
