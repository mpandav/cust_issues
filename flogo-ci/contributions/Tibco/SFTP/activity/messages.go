package flogosftp

import (
	"fmt"

	"github.com/project-flogo/core/activity"
)

// Constants
const (
	CategoryName = "SFTP"
	//Debug Message
	ActivityInput  = 1001
	ActivityOutput = 1002
	//Info Messages
	ActivityStart = 2001
	ActivityEnd   = 2002
	//Error Messages
	DefaultError         = 4001
	FailedInputObject    = 4002
	FailedInputProcess   = 4003
	SpecifyInput         = 4004
	RemoteFileOpenError  = 4005
	LocalFileOpenError   = 4006
	GetOperationError    = 4007
	PutOperationError    = 4008
	Deserialize          = 4008
	Unmarshall           = 4009
	DecodeError          = 4010
	DeleteOperationError = 4011
	ListOperationError   = 4012
	MkdirOperationError  = 4013
	RenameOperationError = 4014
)

var messages = make(map[int]string)

func init() {
	//Debug
	messages[ActivityInput] = "Input data: %s"
	messages[ActivityOutput] = "Output data: %s"
	//Info
	messages[ActivityStart] = "Executing SFTP %s activity"
	messages[ActivityEnd] = "Completed the execution of SFTP %s activity"
	//Error
	messages[DefaultError] = "Error is : %s"
	messages[FailedInputObject] = "Failed to get Input object from context. Error: [%s]"
	messages[FailedInputProcess] = "Failed to process input arguments : %s"
	messages[SpecifyInput] = "No input specified"
	messages[RemoteFileOpenError] = "Unable to open remote file : %s"
	messages[LocalFileOpenError] = "Unable to open local file : %s"
	messages[GetOperationError] = "Get Operation failed. Unable to download remote file: %s"
	messages[PutOperationError] = "Put Operation failed. Unable to upload file: %s"
	messages[Deserialize] = "Cannot deserialize inputData : %s"
	messages[Unmarshall] = "Cannot unmarshall data into parameters : %s"
	messages[DecodeError] = "Error while decoding : %s"
	messages[DeleteOperationError] = "Failed to delete file. %s"
	messages[ListOperationError] = "Failed to Read directory : %s. Error : %s"
	messages[MkdirOperationError] = "Failed to create remote directory : %s. Error : %s"
	messages[RenameOperationError] = "Failed to rename remote file : %s. Error : %s"
}

// GetError to create activity error
func GetError(errConst int, activityName string, parms ...interface{}) *activity.Error {
	errCode := CategoryName + "-" + activityName + "-" + string(errConst)
	return activity.NewError(GetMessage(errConst, parms...), errCode, nil)
}

// GetError to create activity error
func GetRetriableError(errConst int, activityName string, parms ...interface{}) *activity.Error {
	errCode := CategoryName + "-" + activityName + "-" + string(errConst)
	return activity.NewRetriableError(GetMessage(errConst, parms...), errCode, nil)
}

// GetMessage to get message
func GetMessage(msgConst int, parms ...interface{}) string {
	if parms != nil {
		return fmt.Sprintf(messages[msgConst], parms...)
	}
	return messages[msgConst]
}
