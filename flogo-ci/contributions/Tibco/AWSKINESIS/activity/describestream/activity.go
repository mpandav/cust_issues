package describestream

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	mkutil "github.com/tibco/wi-plugins/contributions/awsKinesis/src/app/AWSKINESIS/activity/util"
)

var activityLog = log.ChildLogger(log.RootLogger(), "aws-kinesis-describestream")

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

	activityLog.Debugf("Executing Kinesis Describe Stream activity")

	input := &Input{}
	err = context.GetInputObject(input)

	inputMap := input.Input

	if inputMap == nil {
		return false, activity.NewError(fmt.Sprintf("Input is required in query activity for"), "AWSKINESIS-CREATESTREAM-9001", nil)
	}

	streamType := input.StreamType
	describeType := input.DescribeType

	session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if session.Config.Endpoint != nil {
		endpoint := *session.Config.Endpoint
		session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}

	if streamType == "DataStream" {
		activityLog.Debugf("Describing Data Stream")
		kc := kinesis.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session")
		return ExecuteDescribeDataStream(kc, inputMap, describeType, context)

	} else if streamType == "VideoStream" {
		activityLog.Debugf("Describing Video Stream")
		vkc := kinesisvideo.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session")
		return ExecuteDescribeVideoStream(vkc, inputMap, context)

	} else if streamType == "Firehose-DeliveryStream" {
		activityLog.Debugf("Describing Delivery Stream")
		fkc := firehose.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session")
		return ExecuteDescribeDeliveryStream(fkc, inputMap, context)
	}
	return true, nil
}

// ExecuteDescribeDataStream function is specific to DataStream creation
func ExecuteDescribeDataStream(kc *kinesis.Kinesis, input map[string]interface{}, describeType string, context activity.Context) (done bool, err error) {

	if describeType == "Limits" {
		activityLog.Debugf("Describing Limits \n")
		inputParams := &kinesis.DescribeLimitsInput{}

		out, err := kc.DescribeLimits(inputParams)

		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Describe stream due to error:%s.", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9003", nil)
		}

		var outMsg map[string]interface{}
		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output := &Output{}
		output.Message = outMsg
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}
		activityLog.Debugf("Describe Limits Successfully executed")

		return true, nil
	} else {
		activityLog.Debugf("Converting Input to Bytes \n")
		inputBytes, err := json.Marshal(input)
		inputParams := &kinesis.DescribeStreamInput{}

		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Convert Input Data to Bytes error %s", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9004", nil)
		}
		err = json.Unmarshal(inputBytes, inputParams)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse Input Data error %s", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9005", nil)
		}
		//read the values from configuration
		activityLog.Debugf("Describing DataStream " + aws.StringValue(inputParams.StreamName))

		out, err := kc.DescribeStream(inputParams)

		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Describe stream due to error:%s.", err.Error()), "AWSKINESIS-CREATESTREAM-9006", nil)
		}

		var outMsg map[string]interface{}
		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output := &Output{}
		output.Message = outMsg
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}
		activityLog.Debugf("Describe DataStream Successfully executed\n")

		return true, nil
	}
}

// ExecuteDescribeVideoStream function is specific to DataStream creation
func ExecuteDescribeVideoStream(vkc *kinesisvideo.KinesisVideo, input map[string]interface{}, context activity.Context) (done bool, err error) {

	activityLog.Debugf("Converting input to Bytes !!")
	inputBytes, err := json.Marshal(input)
	inputParams := &kinesisvideo.DescribeStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Convert Input Data to Bytes error %s", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9004", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse Input Data error %s", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9005", nil)
	}
	//read the values from configuration
	activityLog.Debugf("Describing Video Stream " + aws.StringValue(inputParams.StreamName))

	out, err := vkc.DescribeStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to Describe Video Stream due to error:%s.", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9007", nil)
	}

	var outMsg map[string]interface{}
	o, _ := json.Marshal(out)
	json.Unmarshal(o, &outMsg)
	output := &Output{}
	output.Message = outMsg
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	activityLog.Debugf("Describe Video Stream Successfully executed \n")

	return true, nil
}

// ExecuteDescribeDeliveryStream function is specific to DataStream creation
func ExecuteDescribeDeliveryStream(fkc *firehose.Firehose, input map[string]interface{}, context activity.Context) (done bool, err error) {

	activityLog.Debugf("converting input to bytes !!")
	inputBytes, err := json.Marshal(input)
	inputParams := &firehose.DescribeDeliveryStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Convert Input Data to Bytes error %s", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9004", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse Input Data error %s", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9005", nil)
	}
	//read the values from configuration
	activityLog.Debugf("Describing Delivery Stream " + aws.StringValue(inputParams.DeliveryStreamName))

	out, err := fkc.DescribeDeliveryStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to Describe Delivery Stream due to error:%s.", err.Error()), "AWSKINESIS-DESCRIBESTREAM-9008", nil)
	}

	var outMsg map[string]interface{}
	o, _ := json.Marshal(out)
	json.Unmarshal(o, &outMsg)
	output := &Output{}
	output.Message = outMsg
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	activityLog.Debugf("Describe Delivery Stream Successfully executed\n")

	return true, nil
}
