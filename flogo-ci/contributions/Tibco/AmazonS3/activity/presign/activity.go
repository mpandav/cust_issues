package presign

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	s3util "github.com/tibco/wi-amazons3/src/app/AmazonS3/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// Constants for activity
const (
	ActivityName = "Presign"

	// Configuration constants
	ConfOperationType = "operationType"

	// Mapping parameters
	paramKey    = "Key"
	paramBucket = "Bucket"

	// Miscellaneous
	Get    = "get"
	Put    = "put"
	Delete = "delete"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	act := &Activity{}
	return act, nil
}

// Activity The S3 Get Activity
type Activity struct {
	metadata *activity.Metadata
}

// Metadata returns activity metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	// Create S3 service from session
	s3Session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if s3Session.Config.Endpoint != nil {
		endpoint := *s3Session.Config.Endpoint
		s3Session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}
	s3Svc := s3.New(s3Session, endpointConfig)

	operationType := input.OperationType
	expirationTimeSec := input.ExpirationTimeSec

	var outputObject interface{}
	var urlStr string
	switch operationType {
	case Get:
		urlStr, err = s3util.DoGeneratePresignedURLGET(ctx, s3Svc, expirationTimeSec, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Presign GET", err.Error())
		}
	case Put:
		urlStr, err = s3util.DoGeneratePresignedURLPUT(ctx, s3Svc, expirationTimeSec, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Presign PUT", err.Error())
		}
	case Delete:
		urlStr, err = s3util.DoGeneratePresignedURLDELETE(ctx, s3Svc, expirationTimeSec, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Presign DELETE", err.Error())
		}
	}
	outputObject = map[string]string{
		"PresignedURL": urlStr,
	}
	logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), outputObject))
	// 7. Set output
	outputObjectCoerced, err := coerce.ToObject(outputObject)
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}
	errObjectCoerced, err := coerce.ToObject(err)
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}
	output := &Output{
		Output: outputObjectCoerced,
		Error:  errObjectCoerced,
	}
	ctx.SetOutputObject(output)
	return true, nil
}
