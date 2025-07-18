package put

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	s3util "github.com/tibco/wi-amazons3/src/app/AmazonS3/activity"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// Constants for activity
const (
	ActivityName = "Put"

	ConfPutType     = "putType"
	ConfInputType   = "inputType"
	ConfPreserveACL = "preserveACL"

	Bucket = "bucket"
	Object = "object"
	Upload = "upload"
	Copy   = "copy"
)

// Activity The S3 Put Activity
type Activity struct {
	metadata *activity.Metadata
}

func init() {
	_ = activity.Register(&Activity{}, New)
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	act := &Activity{}
	return act, nil
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
	logger.Info(s3util.GetMessage(s3util.ActivityStart, ActivityName, ctx.Name()))

	s3Session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if s3Session.Config.Endpoint != nil {
		endpoint := *s3Session.Config.Endpoint
		s3Session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}
	s3Svc := s3.New(s3Session, endpointConfig)

	serviceName := input.ServiceName

	var outputObject interface{}
	switch serviceName {
	case Bucket:
		result, err := s3util.DoCreateBucket(ctx, s3Svc, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Create Bucket", err.Error())
		}
		resultWithBucketName := map[string]string{
			"Location":   *result.Location,
			"BucketName": input.Input["Bucket"].(string),
		}
		resultForLog, err := json.MarshalIndent(resultWithBucketName, "", "  ")
		if err != nil {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Create Bucket", err.Error())
		}
		logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), string(resultForLog)))
		outputObject = resultWithBucketName

	case Object:
		putType := input.PutType
		switch putType {
		case Upload:
			inputType := input.InputType
			s3Uploader := s3manager.NewUploader(s3Session)
			result := &s3.PutObjectOutput{}
			if inputType == "text" {
				// text content
				result, err = s3util.DoUploadObject(ctx, s3Svc, s3Uploader, input.Input, true, logger)
			} else {
				// upload from local file
				result, err = s3util.DoUploadObject(ctx, s3Svc, s3Uploader, input.Input, false, logger)
			}
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Upload Object", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result

		case Copy:
			isPreserveACL := input.PreserveACL
			result, err := s3util.DoCopyObject(ctx, s3Svc, input.Input, isPreserveACL, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Copy Object", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result
		}
	}
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
