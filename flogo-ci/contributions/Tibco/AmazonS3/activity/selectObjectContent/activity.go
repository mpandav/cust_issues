package selectObjectContent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	s3util "github.com/tibco/wi-amazons3/src/app/AmazonS3/activity"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// Constants for activity
const (
	ActivityName = "SelectObjectContent"

	// Mapping parameters
	paramBucket               = "Bucket"
	paramKey                  = "Key"
	paramExpectedBucketOwner  = "ExpectedBucketOwner"
	paramExpression           = "Expression"
	paramExpressionType       = "ExpressionType"
	paramInputSerialization   = "InputSerialization"
	paramOutputSerialization  = "OutputSerialization"
	paramRequestProgress      = "RequestProgress"
	paramSSECustomerAlgorithm = "SSECustomerAlgorithm"
	paramSSECustomerKey       = "SSECustomerKey"
	paramSSECustomerKeyMD5    = "SSECustomerKeyMD5"
	paramScanRange            = "ScanRange"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	act := &Activity{settings: s}
	return act, nil
}

// S3 select object activity
type Activity struct {
	metadata *activity.Metadata
	settings *Settings
}

// Metadata returns activity metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	var outputObject interface{}
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}

	// Create S3 service from session
	s3Session := a.settings.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if s3Session.Config.Endpoint != nil {
		endpoint := *s3Session.Config.Endpoint
		s3Session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}
	s3Svc := s3.New(s3Session, endpointConfig)

	key := input.Input[paramKey].(string)
	bucket := input.Input[paramBucket].(string)

	// var expectedBucketOwner string
	// if input.Input[paramExpectedBucketOwner] != nil {
	// 	expectedBucketOwner = input.Input[paramExpectedBucketOwner].(string)
	// }

	expression := input.Input[paramExpression].(string)
	expressionType := input.Input[paramExpressionType].(string)

	var sseCustomerAlgorithm string
	if input.Input[paramSSECustomerAlgorithm] != nil {
		sseCustomerAlgorithm = input.Input[paramSSECustomerAlgorithm].(string)
	}
	var sseCustomerKey string
	if input.Input[paramSSECustomerKey] != nil {
		sseCustomerKey = input.Input[paramSSECustomerKey].(string)
	}
	var sseCustomerKeyMD5 string
	if input.Input[paramSSECustomerKeyMD5] != nil {
		sseCustomerKeyMD5 = input.Input[paramSSECustomerKeyMD5].(string)
	}

	//RequestProgress
	RequestProgress := &s3.RequestProgress{}
	dataBytes, err := json.Marshal(input.Input[paramRequestProgress])
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}
	json.Unmarshal(dataBytes, &RequestProgress)

	//ScanRange
	scanRange := &s3.ScanRange{}
	dataBytes, err = json.Marshal(input.Input[paramScanRange])
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}
	json.Unmarshal(dataBytes, &scanRange)

	//reading output type and file path from output serilization
	var OutputType, DestinationFilePath string
	if input.Input[paramOutputSerialization] != nil {
		outputSerialization := input.Input[paramOutputSerialization].(map[string]interface{})
		if len(outputSerialization) > 0 && outputSerialization["OutputType"] != nil && outputSerialization["DestinationFilePath"] != nil {
			OutputType = outputSerialization["OutputType"].(string)
			DestinationFilePath = outputSerialization["DestinationFilePath"].(string)
		} else {
			OutputType = "Text"
		}
	}

	// check if file exists
	_, err = s3Svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, "Get Object", err.Error())
	} else {
		var request *s3.SelectObjectContentOutput

		//InputSerialization ->  CSV,JSON,Parquet
		if input.InputSerialization == "csv" {
			/*
				InputSerialization
			*/
			csvInput := &s3.CSVInput{}
			dataBytes, err := json.Marshal(input.Input[paramInputSerialization])
			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}
			json.Unmarshal(dataBytes, &csvInput)

			inputSerialization := &s3.InputSerialization{}
			//passing on default schema if inputSerialization is nil
			if csvInput != nil {
				inputSerialization.CSV = csvInput
			} else {
				inputSerialization.CSV = &s3.CSVInput{}
			}

			if input.CompressionType != "" {
				inputSerialization.CompressionType = &input.CompressionType
			}

			/*
				OutputSerialization
			*/
			outputSerialization, err := configureOutputSerializationSchema(input)
			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

			//creating a request for select object
			request, err = s3Svc.SelectObjectContent(&s3.SelectObjectContentInput{
				Bucket:         aws.String(bucket),
				Key:            aws.String(key),
				Expression:     aws.String(expression),
				ExpressionType: aws.String(expressionType),
				//ExpectedBucketOwner:  aws.String(expectedBucketOwner),
				InputSerialization:   inputSerialization,
				OutputSerialization:  outputSerialization,
				RequestProgress:      RequestProgress,
				SSECustomerAlgorithm: aws.String(sseCustomerAlgorithm),
				SSECustomerKey:       aws.String(sseCustomerKey),
				SSECustomerKeyMD5:    aws.String(sseCustomerKeyMD5),
				ScanRange:            scanRange,
			})
			logger.Debug(s3util.GetMessage(s3util.ActivityInput, ctx.Name(), request.GoString()))

			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

			outputObject, err = processResponse(request, input, ctx, OutputType, DestinationFilePath)

			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

		} else if input.InputSerialization == "json" {
			/*
				InputSerialization
			*/
			jsonInput := &s3.JSONInput{}
			dataBytes, err := json.Marshal(input.Input[paramInputSerialization])
			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}
			json.Unmarshal(dataBytes, &jsonInput)

			inputSerialization := &s3.InputSerialization{}
			//passing on default schema if inputSerialization is nil
			if jsonInput != nil {
				inputSerialization.JSON = jsonInput
			} else {
				inputSerialization.JSON = &s3.JSONInput{}
			}

			if input.CompressionType != "" {
				inputSerialization.CompressionType = &input.CompressionType
			}

			/*
				OutputSerialization
			*/
			outputSerialization, err := configureOutputSerializationSchema(input)
			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

			//creating a request for select object
			request, err = s3Svc.SelectObjectContent(&s3.SelectObjectContentInput{
				Bucket:         aws.String(bucket),
				Key:            aws.String(key),
				Expression:     aws.String(expression),
				ExpressionType: aws.String(expressionType),
				//ExpectedBucketOwner:  aws.String(expectedBucketOwner),
				InputSerialization:   inputSerialization,
				OutputSerialization:  outputSerialization,
				RequestProgress:      RequestProgress,
				SSECustomerAlgorithm: aws.String(sseCustomerAlgorithm),
				SSECustomerKey:       aws.String(sseCustomerKey),
				SSECustomerKeyMD5:    aws.String(sseCustomerKeyMD5),
				ScanRange:            scanRange,
			})
			logger.Debug(s3util.GetMessage(s3util.ActivityInput, ctx.Name(), request.GoString()))

			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

			outputObject, err = processResponse(request, input, ctx, OutputType, DestinationFilePath)

			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

		} else if input.InputSerialization == "parquet" {
			/*
				InputSerialization
			*/
			parquetInput := &s3.ParquetInput{}
			dataBytes, err := json.Marshal(input.Input[paramInputSerialization])
			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}
			json.Unmarshal(dataBytes, &parquetInput)

			inputSerialization := &s3.InputSerialization{}
			//passing on default schema if inputSerialization is nil
			if parquetInput != nil {
				inputSerialization.Parquet = parquetInput
			} else {
				inputSerialization.Parquet = &s3.ParquetInput{}
			}

			if input.CompressionType != "" {
				inputSerialization.CompressionType = &input.CompressionType
			}

			/*
				OutputSerialization
			*/
			outputSerialization, err := configureOutputSerializationSchema(input)
			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

			//creating a request for select object
			request, err = s3Svc.SelectObjectContent(&s3.SelectObjectContentInput{
				Bucket:         aws.String(bucket),
				Key:            aws.String(key),
				Expression:     aws.String(expression),
				ExpressionType: aws.String(expressionType),
				//ExpectedBucketOwner:  aws.String(expectedBucketOwner),
				InputSerialization:   inputSerialization,
				OutputSerialization:  outputSerialization,
				RequestProgress:      RequestProgress,
				SSECustomerAlgorithm: aws.String(sseCustomerAlgorithm),
				SSECustomerKey:       aws.String(sseCustomerKey),
				SSECustomerKeyMD5:    aws.String(sseCustomerKeyMD5),
				ScanRange:            scanRange,
			})
			logger.Debug(s3util.GetMessage(s3util.ActivityInput, ctx.Name(), request.GoString()))

			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

			outputObject, err = processResponse(request, input, ctx, OutputType, DestinationFilePath)

			if err != nil {
				return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
			}

		} else {
			return false, s3util.GetError(s3util.DefaultError, ctx.Name(), "Invalid input serialization")
		}
	}

	// Set output
	outputObjectCoerced, err := coerce.ToString(outputObject)
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}
	errObjectCoerced, err := coerce.ToObject(err)
	if err != nil {
		return false, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}

	output := &Output{}
	output.Error = errObjectCoerced
	output.OutputSerialization = input.OutputSerialization
	output.Output = outputObjectCoerced

	if strings.EqualFold(OutputType, "file") {
		output.Output = DestinationFilePath
	}

	ctx.SetOutputObject(output)
	return true, nil
}

func configureOutputSerializationSchema(input *Input) (*s3.OutputSerialization, error) {
	outputSerializationSchema := &s3.OutputSerialization{}
	if input.OutputSerialization == "csv" {
		csvOutput := &s3.CSVOutput{}
		dataBytes, err := json.Marshal(input.Input[paramOutputSerialization])
		if err != nil {
			return nil, err
		}
		json.Unmarshal(dataBytes, &csvOutput)

		//passing on default schema if outputSerialization is nil
		if csvOutput != nil {
			outputSerializationSchema.CSV = csvOutput
		} else {
			outputSerializationSchema.CSV = &s3.CSVOutput{}
		}
	} else if input.OutputSerialization == "json" {
		JSONOutput := &s3.JSONOutput{}
		dataBytes, err := json.Marshal(input.Input[paramOutputSerialization])
		if err != nil {
			return nil, err
		}
		json.Unmarshal(dataBytes, &JSONOutput)

		//passing on default schema if outputSerialization is nil
		if JSONOutput != nil {
			outputSerializationSchema.JSON = JSONOutput
		} else {
			outputSerializationSchema.JSON = &s3.JSONOutput{}
		}
	}
	return outputSerializationSchema, nil
}

func processResponse(request *s3.SelectObjectContentOutput, input *Input, ctx activity.Context, OutputType string, DestinationFilePath string) (_ []byte, err error) {

	defer request.EventStream.Close()

	var records []byte
	for event := range request.EventStream.Events() {
		switch e := event.(type) {
		case *s3.RecordsEvent:
			records = append(records, e.Payload...)
		case *s3.StatsEvent:
			defer ctx.Logger().Info(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), fmt.Sprintf("Processed %d bytes\n", *e.Details.BytesProcessed)))
		case *s3.ProgressEvent:
			ctx.Logger().Info(s3util.GetMessage(s3util.ActivityOutput, ctx.Name(), fmt.Sprintf("ProgressEvent: %s \n", e.Details.String())))
		}
	}

	//checking for event stream error
	if err := request.EventStream.Err(); err != nil {
		ctx.Logger().Info(s3util.GetMessage(s3util.DefaultError, ctx.Name(), err.Error()))
		return nil, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
	}

	// configuring the output
	if strings.EqualFold(OutputType, "file") { //writing records to file
		file, err := os.OpenFile(DestinationFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
		}
		defer file.Close()
		_, err = file.Write(records)
		if err != nil {
			return nil, s3util.GetError(s3util.FailedInExecution, ActivityName, err.Error())
		}
		return nil, nil
	}

	return records, nil

}
