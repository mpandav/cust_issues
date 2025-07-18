package delete

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
	ActivityName = "Delete"

	Bucket = "bucket"
	Object = "object"

	paramKey    = "Key"
	paramBucket = "Bucket"

	headObjectErrorCode    = "ObjectDoesNotExist"
	headObjectErrorMessage = "The specified object does not exist in the given bucket."
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

// Eval implements activity.Activity.Eval
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

	// 4. Get input values
	serviceName := input.ServiceName

	var outputObject interface{}

	// 6. Execute Action
	switch serviceName {
	case Bucket:
		result, err := s3util.DoDeleteBucket(ctx, s3Svc, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Delete Bucket", err.Error())
		}
		logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
		outputObject = result
	case Object:
		key := input.Input[paramKey].(string)
		bucket := input.Input[paramBucket].(string)

		bucketExistsInput := &s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		}
		_, err := s3Svc.HeadBucket(bucketExistsInput)
		if s3util.IsFatalError(ctx, err) {
			// bucket does not exist
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Delete Object", err.Error())
		}
		objExistsInput := &s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}
		_, err = s3Svc.HeadObject(objExistsInput)
		if s3util.IsFatalError(ctx, err) {
			// object does not exist
			s3util.SetErrorInfo(ctx, err, headObjectErrorCode, headObjectErrorMessage)
			return false, s3util.GetError(s3util.ObjectDoesNotExist, ActivityName, "Delete Object", key, bucket)
		}
		result, err := s3util.DoDeleteObject(ctx, s3Svc, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Delete Object", err.Error())
		}
		logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
		outputObject = result
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
