package update

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
	ActivityName = "Update"

	ConfUpdateType = "updateType"

	Bucket     = "bucket"
	Object     = "object"
	ACL        = "acl"
	CORS       = "cors"
	Policy     = "policy"
	Versioning = "versioning"
	Website    = "website"
	Tagging    = "tagging"
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

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	logger.Info(s3util.GetMessage(s3util.ActivityStart, ActivityName, ctx.Name()))
	// 3. Create S3 service from session
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
	updateType := input.UpdateType

	var outputObject interface{}

	// 6. Get output
	switch serviceName {
	case Bucket:
		switch updateType {
		case ACL:
			result, err := s3util.DoUpdateBucketACL(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Bucket ACL", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result
		case CORS:
			result, err := s3util.DoUpdateBucketCORS(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Bucket CORS", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result
		case Policy:
			result, err := s3util.DoUpdateBucketPolicy(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Bucket Policy", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result
		case Versioning:
			result, err := s3util.DoUpdateBucketVersioning(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Bucket Versioning", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result
		case Website:
			result, err := s3util.DoUpdateBucketWebsite(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Bucket Website", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result

		}
	case Object:
		switch updateType {
		case ACL:
			result, err := s3util.DoUpdateObjectACL(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Object ACL", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
			outputObject = result
		case Tagging:
			result, err := s3util.DoUpdateObjectTagging(ctx, s3Svc, input.Input, logger)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Update Object Tagging", err.Error())
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
