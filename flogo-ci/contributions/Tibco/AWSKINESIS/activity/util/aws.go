package util

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
)

// GetAWSSession is util function which returns with an aws session
// func GetAWSSession(context activity.Context) (*session.Session, error) {
// 	connectionInfo := context.GetInput(ivConnection)
// 	conn, err := awsconn.NewConnection(connectionInfo)
// 	if err != nil {
// 		return nil, activity.NewError("AWS connection is not configured", "KINESIS-UTIL-9001", err.Error())
// 	}
// 	return conn.NewSession(), nil
// }

//SetErrorObject sets the error complexObject from error
func SetErrorObject(err error, context activity.Context) {
	if awsErr, ok := err.(awserr.Error); ok {
		errSchema := make(map[string]interface{})
		errSchema["error_code"] = awsErr.Code()
		errSchema["error_message"] = awsErr.Message()
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			errSchema["statusCode"] = reqErr.StatusCode()
			errSchema["requestId"] = reqErr.RequestID()
		}
		output := &Output{}
		output.Error = errSchema
		_ = context.SetOutputObject(output)
	}
}

type Output struct {
	Error map[string]interface{} `md:"Error"`
}

// ToMap ...
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Error": o.Error,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Error, err = coerce.ToObject(values["Error"])
	if err != nil {
		return err
	}
	return nil
}
