package helper

import (
	"fmt"

	"github.com/project-flogo/core/activity"
)

// Constants
const (
	CategoryName = "OracleDatabase"
	//Debug Message
	ExecutingQuery    = 1001
	QueryWithParam    = 1002
	NoRowsFound       = 1003
	Login             = 1004
	Logout            = 1005
	ActivityOutput    = 1006
	InputDataContent  = 1007
	Preparedquery     = 1008
	OutputDataContent = 1009
	InputQuery        = 1010
	FieldsInfo        = 1011
	InputData         = 1012
	InputParams       = 1013
	//Info Messages
	ActivityStart     = 2001
	ExecutionFlowInfo = 2002
	//Error Messages
	DefaultError              = 4001
	ConnectionNotConfigured   = 4002
	ConnectionFailed          = 4003
	DBConnectionNotConfigured = 4004
	DBConnectionFailed        = 4005
	LoginError                = 4006
	SpecifySQL                = 4007
	FailedInputProcess        = 4008
	FailedExecuteSQL          = 4009
	OutputSchemaValidation    = 4010
	SpecifyInput              = 4012
	Unmarshall                = 4013
	Deserialize               = 4014
	QueryFailed               = 4017
	GettingInColumnInfo       = 4018
	GettingInColumnType       = 4019
	ScanningRow               = 4020
	DBConnectionIssue         = 4021
	FailedInputObject         = 4022
	TxcommitError             = 4023
)

var messages = make(map[int]string)

func init() {
	//Debug
	messages[ExecutingQuery] = "Executing prepared query : %s"
	messages[QueryWithParam] = "Prepared Query [%s] Parameters [%v]"
	messages[NoRowsFound] = "No rows returned for query : %s"
	messages[InputDataContent] = "Input data content : %s"
	messages[ActivityOutput] = "Output data content : %s"
	messages[Preparedquery] = "Prepared Query: [%s]"
	messages[OutputDataContent] = "OutputDataContent: %s"
	messages[FieldsInfo] = "FieldInfo are: %s"
	messages[InputQuery] = "Query Entered: %s"
	messages[FieldsInfo] = "Fields info: %s"
	messages[InputData] = "Input data: %s"
	messages[InputParams] = "Input Parameters: %s"
	//Info
	messages[ActivityStart] = "Executing OracleDatabase %s activity"
	messages[ExecutionFlowInfo] = "%s"
	//Error
	messages[DefaultError] = "Error is : %s"
	messages[DBConnectionNotConfigured] = "OracleDatabase connection not configured"
	messages[DBConnectionFailed] = "OracleDatabase Connection Failed : %s"
	messages[SpecifySQL] = "No SQL Statement specified"
	messages[FailedInputProcess] = "Failed to process input arguments : %s"
	messages[FailedExecuteSQL] = "Failed to execute : %s"
	messages[OutputSchemaValidation] = "Output Schema validation error : %s"
	messages[SpecifyInput] = "No input arguments for query specified"
	messages[Unmarshall] = "Cannot unmarshall data into parameters : %s"
	messages[Deserialize] = "Cannot deserialize inputData : %s"
	messages[QueryFailed] = "db.Query failed for reason : %s"
	messages[GettingInColumnInfo] = "Error getting column information : %s"
	messages[GettingInColumnType] = "Error determining column types : %s"
	messages[ScanningRow] = "Error getting query result set : %s"
	messages[DBConnectionIssue] = "Could not open connection to database %s, %s"
	messages[FailedInputObject] = "Failed to get Input object from context. Error: [%s]"
	messages[TxcommitError] = "Failed to commit Error: [%s]"
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
