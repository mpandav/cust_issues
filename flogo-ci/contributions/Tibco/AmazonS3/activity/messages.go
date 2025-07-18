package s3util

import (
	"fmt"

	"github.com/project-flogo/core/activity"
)

//Constants
const (
	CategoryName = "AmazonS3"
	//Debug Message
	ActivityInput      = 1001
	ActivityOutput     = 1002
	Session            = 1003
	ObjectExists       = 1004
	DestinationPathMsg = 1005
	ExistingProperty   = 1006
	//Info Messages
	ActivityStart = 2001
	//Error Messages
	DefaultError                    = 4001
	ConnectionNotConfigured         = 4002
	FailedToConnectAWS              = 4003
	FailedInExecution               = 4004
	FailedToConvertInputToBytes     = 4005
	FailedToParseInputData          = 4006
	FailedToValidateInputData       = 4007
	FailedToConvertOutputToBytes    = 4008
	ObjectDoesNotExist              = 4009
	ErrorUpdatingObjectACL          = 4010
	InvalidWritePermissionForObject = 4011
	ErrorGettingBucketProperty      = 4012
	ErrorGettingObjectProperty      = 4013

	DownloadError = 4020
	MappingError  = 4021
	UploadError   = 4022
)

var messages = make(map[int]string)

func init() {
	// Info
	messages[ActivityStart] = "Executing AmazonS3 %s activity - [%s]"

	// Debug
	messages[ActivityInput] = "Input for activty [%s]: %s"
	messages[ActivityOutput] = "Output for activity [%s]: %s"
	messages[Session] = "Getting session for S3 service"
	messages[ObjectExists] = "Object exists, getting it now..."
	messages[DestinationPathMsg] = "Destination file path set as: %s"
	messages[ExistingProperty] = "Existing %s: %s"

	// Error
	messages[ConnectionNotConfigured] = "AWS connection is not configured: %s"
	messages[FailedToConnectAWS] = "Failed to connect to AWS <CausedBy> %s. Check credentials configured in the connection: [%s]"
	messages[FailedInExecution] = "Failed to %s <CausedBy> %s"
	messages[FailedToConvertInputToBytes] = "Error converting input data to bytes: %s"
	messages[FailedToConvertOutputToBytes] = "Error converting input data to bytes: %s"
	messages[FailedToParseInputData] = "Error parsing input data: %s"
	messages[FailedToValidateInputData] = "Error validating input data: %s"
	messages[DownloadError] = "Error downloading file <CausedBy> %s"
	messages[MappingError] = "Error Mapping Output for [%s] <CausedBy> %s"
	messages[UploadError] = "Error reading file <CausedBy> %s"
	messages[ObjectDoesNotExist] = "Failed to %s. <CausedBy> Object [%s] does not exist in bucket [%s]"
	messages[ErrorUpdatingObjectACL] = "Error updating object acl <CausedBy> %s"
	messages[InvalidWritePermissionForObject] = "WRITE permission is not applicable for an object. Please use WRITE_ACP instead."
	messages[ErrorGettingObjectProperty] = "Error getting the %s for the object [%s] in the bucket [%s] <CausedBy> %s"
	messages[ErrorGettingBucketProperty] = "Error getting the %s for the bucket [%s] <CausedBy> %s"
}

//GetError to create activity error
func GetError(errConst int, activityName string, parms ...interface{}) *activity.Error {
	errCode := CategoryName + "-" + activityName + "-" + string(errConst)
	return activity.NewError(GetMessage(errConst, parms...), errCode, nil)
}

//GetMessage to get message
func GetMessage(msgConst int, parms ...interface{}) string {
	if parms != nil {
		return fmt.Sprintf(messages[msgConst], parms...)
	}
	return messages[msgConst]
}
