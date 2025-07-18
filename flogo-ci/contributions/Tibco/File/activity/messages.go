package flogofile

import (
	"fmt"

	"github.com/project-flogo/core/activity"
)

// Constants
const (
	CategoryName = "File"
	//Debug Message
	ActivityInput  = 1001
	ActivityOutput = 1002
	//Info Messages
	ActivityStart = 2001
	ActivityEnd   = 2002
	//Error Messages
	DefaultError        = 4001
	FailedInputObject   = 4002
	CreateFileError     = 4003
	CreateDirError      = 4004
	FailedToGetFileInfo = 4005
	FailedInputProcess  = 4006
	SpecifyInput        = 4007
	Deserialize         = 4008
	Unmarshall          = 4009
	RemoveFileError     = 4010
	ReadFileError       = 4011
	WriteFileError      = 4012
	RenameFileError     = 4013
	CopyFileError       = 4014
	ListFileError       = 4015
)

var messages = make(map[int]string)

func init() {
	//Debug
	messages[ActivityInput] = "Input data: %s"
	messages[ActivityOutput] = "Output data: %s"
	//Info
	messages[ActivityStart] = "Executing File %s activity"
	messages[ActivityEnd] = "Completed the execution of File %s activity"
	//Error
	messages[DefaultError] = "%s"
	messages[FailedInputObject] = "Failed to get Input object from context. Error: [%s]"
	messages[CreateFileError] = "Error in creating the file : %s"
	messages[CreateDirError] = "Error in creating the directory : %s"
	messages[FailedToGetFileInfo] = "Failed to get file information : %s"
	messages[FailedInputProcess] = "Failed to process input arguments : %s"
	messages[SpecifyInput] = "No input specified"
	messages[Deserialize] = "Cannot deserialize inputData : %s"
	messages[Unmarshall] = "Cannot unmarshall data into parameters : %s"
	messages[RemoveFileError] = "Error in removing the file : %s"
	messages[ReadFileError] = "Error in reading the file : %s"
	messages[WriteFileError] = "Error in writing to the file : %s"
	messages[RenameFileError] = "Error in renaming the file : %s"
	messages[CopyFileError] = "Error in copying the file : %s"
	messages[ListFileError] = "Error in listing the files : %s"
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
