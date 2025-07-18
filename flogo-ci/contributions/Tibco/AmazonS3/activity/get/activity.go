package get

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	s3util "github.com/tibco/wi-amazons3/src/app/AmazonS3/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// Constants for activity
const (
	ActivityName = "Get"

	// Configuration constants
	ConfOperationType = "operationType"
	ConfOutputType    = "outputType"

	// Mapping parameters
	paramKey                 = "Key"
	paramBucket              = "Bucket"
	paramPrefix              = "Prefix"
	paramVersionID           = "VersionId"
	paramDestinationFilePath = "DestinationFilePath"

	// Miscellaneous
	Single      = "single"
	ListObjects = "listObjects"
	ListBuckets = "listBuckets"
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
	outputType := input.OutputType

	var outputObject interface{}

	switch operationType {
	case Single:
		downloader := s3manager.NewDownloader(s3Session)
		key := input.Input[paramKey].(string)
		bucket := input.Input[paramBucket].(string)
		var destinationPath, versionID string
		if input.Input[paramVersionID] != nil {
			versionID = input.Input[paramVersionID].(string)
		}

		// check if file exists
		_, err := s3Svc.HeadObject(&s3.HeadObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(key),
			VersionId: s3util.GetAwsString(versionID),
		})

		if err == nil {

			request := &s3.GetObjectInput{
				Bucket:    aws.String(bucket),
				Key:       aws.String(key),
				VersionId: s3util.GetAwsString(versionID),
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityInput, ctx.Name(), request.GoString()))

			// Get Object
			result, err := s3Svc.GetObject(request)
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Get Object", err.Error())
			}
			logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))

			var outputMap map[string]interface{}
			if outputType == "file" {
				// download file to disk
				destinationPath := input.Input[paramDestinationFilePath].(string)
				logger.Debug(s3util.GetMessage(s3util.DestinationPathMsg, destinationPath))
				_, err := s3util.DownloadFile(downloader, request, destinationPath)
				if s3util.IsFatalError(ctx, err) {
					return false, s3util.GetError(s3util.DownloadError, ctx.Name(), err.Error())
				}
				// get output
				outputMap, err = s3util.GetObjectOutput(request, result, destinationPath, false)
				if s3util.IsFatalError(ctx, err) {
					return false, s3util.GetError(s3util.MappingError, ctx.Name(), err.Error())
				}
			} else {
				// get text content
				outputMap, err = s3util.GetObjectOutput(request, result, destinationPath, true)
				if s3util.IsFatalError(ctx, err) {
					return false, s3util.GetError(s3util.MappingError, ctx.Name(), err.Error())
				}
			}
			// outputComplex = &data.ComplexObject{Metadata: "", Value: outputMap}
			outputObject = outputMap
		} else {
			if s3util.IsFatalError(ctx, err) {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Get Object", err.Error())
			}
		}

	case ListObjects:
		result, err := s3util.DoListObjects(ctx, s3Svc, input.Input, logger)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "List Objects", err.Error())
		}
		logger.Debug(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), result.GoString()))
		outputObject = result

	case ListBuckets:
		result, err := s3Svc.ListBuckets(nil)
		if s3util.IsFatalError(ctx, err) {
			return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "List Buckets", err.Error())
		}
		// filter results if prefix has been set
		var prefix string
		if input.Input[paramPrefix] != nil {
			prefix = input.Input[paramPrefix].(string)
		}
		filtered := []*s3.Bucket{}
		if prefix != "" {
			for _, b := range result.Buckets {
				if strings.HasPrefix(aws.StringValue(b.Name), prefix) {
					filtered = append(filtered, b)
				}
			}
			result.SetBuckets(filtered)
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
