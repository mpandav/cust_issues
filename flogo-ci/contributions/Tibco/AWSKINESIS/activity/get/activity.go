package get

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	mkutil "github.com/tibco/wi-plugins/contributions/awsKinesis/src/app/AWSKINESIS/activity/util"
)

var activityLog = log.ChildLogger(log.RootLogger(), "aws-kinesis-getactivity")

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&Activity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	act := &Activity{}
	return act, nil
}

// Acvitity is the structure for Activity Metadata
type Activity struct {
	metadata *activity.Metadata
}

// Metadata returns activity metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(context activity.Context) (done bool, err error) {

	activityLog.Debugf("Executing Get activity")

	input := &Input{}
	err = context.GetInputObject(input)

	inputMap := input.Input

	if inputMap == nil {
		return false, activity.NewError(fmt.Sprintf("Input is required in query activity for"), "AWSKINESIS-CREATESTREAM-9001", nil)
	}

	streamType := input.StreamType
	getType := input.GetType

	session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if session.Config.Endpoint != nil {
		endpoint := *session.Config.Endpoint
		session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}

	if streamType == "DataStream" {
		activityLog.Debugf("Get From Data Stream")
		kc := kinesis.New(session, endpointConfig)
		activityLog.Debugf("Created aws session")
		return ExecuteGetDataStream(kc, inputMap, getType, context)

	} else if streamType == "VideoStream" {
		//Not Implemented yet

	}
	return true, nil

}

// ExecuteGetDataStream function is specific to DataStream creation
func ExecuteGetDataStream(kc *kinesis.Kinesis, input map[string]interface{}, getType string, context activity.Context) (done bool, err error) {

	activityLog.Debugf("Converting input to bytes\n")
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return false, activity.NewError("Convert input data to bytes error %s "+err.Error(), "KINESIS-GET-9004", nil)
	}
	output := &Output{}

	if getType == "Records" {
		inputParams := &kinesis.GetRecordsInput{}
		err = json.Unmarshal(inputBytes, inputParams)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse input data error %s", err.Error()), "KINESIS-GET-9005", nil)
		}
		out, err := kc.GetRecords(inputParams)
		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Put Data in Data Stream due to error:%s.", err.Error()), "AWSKINESIS-GET-9006", nil)
		}
		var outMsg map[string]interface{}
		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output.Message = outMsg

	} else if getType == "ShardIterator" {
		inputParams := &kinesis.GetShardIteratorInput{}
		err = json.Unmarshal(inputBytes, inputParams)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse input data error %s", err.Error()), "AWSKINESIS-GET-9007", nil)
		}
		out, err := kc.GetShardIterator(inputParams)
		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Put Data in Data Stream due to error:%s.", err.Error()), "AWSKINESIS-GET-9008", nil)
		}

		var outMsg map[string]interface{}
		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output.Message = outMsg

	}

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to Delete stream due to error:%s.", err.Error()), "AWSKINESIS-GET-9009", nil)
	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	return true, nil
}
