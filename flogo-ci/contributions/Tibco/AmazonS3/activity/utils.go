package s3util

import (
	// "github.com/TIBCOSoftware/flogo-lib/core/activity"
	// "github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/project-flogo/core/activity"
)

// GetAwsString returns pointer to string or nil if empty
func GetAwsString(v string) *string {
	if v != "" {
		return aws.String(v)
	}
	return nil
}

// ComplexValue returns value
// func ComplexValue(complexObject *data.ComplexObject) interface{} {
// 	if complexObject != nil {
// 		return complexObject.Value
// 	}
// 	return nil
// }

// func ComplexValue(complexObject *data.ComplexObject) interface{} {
// 	if complexObject != nil {
// 		return complexObject.Value
// 	}
// 	return nil
// }

// IsFatalError checks for AWS Error
func IsFatalError(context activity.Context, err error) bool {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			errMap := make(map[string]interface{})
			errMap["code"] = awsErr.Code()
			errMap["message"] = awsErr.Message()
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occured
				errMap["statusCode"] = reqErr.StatusCode()
				errMap["requestId"] = reqErr.RequestID()
			}
			// errorComplex := &data.ComplexObject{Metadata: "", Value: errMap}
			// context.SetOutput("error", errorComplex)
			context.SetOutput("error", errMap)
		}
		return true
	}
	return false
}

// SetErrorInfo - Sets Code and Message
func SetErrorInfo(context activity.Context, err error, code string, msg string) {
	errMap := make(map[string]interface{})
	errMap["code"] = code
	errMap["message"] = msg
	if reqErr, ok := err.(awserr.RequestFailure); ok {
		// A service error occured
		errMap["statusCode"] = reqErr.StatusCode()
		errMap["requestId"] = reqErr.RequestID()
	}
	// errorComplex := &data.ComplexObject{Metadata: "", Value: errMap}
	// context.SetOutput("error", errorComplex)
	context.SetOutput("error", errMap)
}
