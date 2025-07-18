package deletestream

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

// const (
// 	ivInput      = "input"
// 	ivStreamType = "streamType"
// 	ovMessage    = "Message"
// )

var activityLog = log.ChildLogger(log.RootLogger(), "aws-kinesis-deletestream")

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

	activityLog.Debugf("Executing Kinesis DeleteStream activity \n")

	input := &Input{}
	err = context.GetInputObject(input)

	if err != nil {
		activityLog.Errorf("Error while parsing Input Object, %v", err)
		return false, err
	}

	inputMap := input.Input

	if inputMap == nil {
		return false, activity.NewError(fmt.Sprintf("Input is required in query activity for"), "AWSKINESIS-CREATESTREAM-9001", nil)
	}

	streamType := input.StreamType
	if streamType == "" {
		return false, activity.NewError("Select Type of Kinesis Stream to be Deleted", "KINESIS-DELETESTREAM-9002", nil)
	}

	session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if session.Config.Endpoint != nil {
		endpoint := *session.Config.Endpoint
		session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}

	if streamType == "DataStream" {
		activityLog.Debugf("Deleting Data Stream")
		kc := kinesis.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session \n")
		return ExecuteDeleteDataStream(kc, inputMap, context)
	} else if streamType == "VideoStream" {
		activityLog.Debugf("Deleting Video Stream")
		vkc := kinesisvideo.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session \n")
		return ExecuteDeleteVideoStream(vkc, inputMap, context)
	} else if streamType == "Firehose-DeliveryStream" {
		activityLog.Debugf("Deleting Delivery Stream")
		svc := firehose.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session \n")
		return ExecuteDeleteDeliveryStream(svc, inputMap, context)

	}
	return true, nil
}

// ExecuteDeleteDataStream function Deletes an existing Data Stream
func ExecuteDeleteDataStream(kc *kinesis.Kinesis, input map[string]interface{}, context activity.Context) (done bool, err error) {

	activityLog.Debugf("Converting Input to Bytes \n")
	inputBytes, err := json.Marshal(input)
	inputParams := &kinesis.DeleteStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Convert input data to bytes error %s", err.Error()), "AWSKINESIS-DELETESTREAM-9003", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse input data error : %s", err.Error()), "AWSKINESIS-DELETESTREAM-9004", nil)
	}

	out, err := kc.DeleteStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to Delete stream due to error:%s.", err.Error()), "AWSKINESIS-DELETESTREAM-9005", nil)
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
	activityLog.Debugf("Deleted DataStream Successfully ", aws.StringValue(inputParams.StreamName))

	return true, nil

}

// ExecuteDeleteVideoStream function Deletes an existing Video Stream
func ExecuteDeleteVideoStream(vkc *kinesisvideo.KinesisVideo, input map[string]interface{}, context activity.Context) (done bool, err error) {

	activityLog.Debugf("Converting input to bytes ")
	inputBytes, err := json.Marshal(input)
	inputParams := &kinesisvideo.DeleteStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse input data error : %s", err.Error()), "AWSKINESIS-DELETESTREAM-9004", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse input data error : %s", err.Error()), "AWSKINESIS-DELETESTREAM-9004", nil)
	}

	out, err := vkc.DeleteStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to Delete stream due to error:%s.", err.Error()), "AWSKINESIS-DELETESTREAM-9005", nil)
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
	activityLog.Debugf("Deleted Video Stream Successfully with ARN ", aws.StringValue(inputParams.StreamARN))

	return true, nil
}

// ExecuteDeleteDeliveryStream function Deletes an existing Delivery Stream
func ExecuteDeleteDeliveryStream(svc *firehose.Firehose, input map[string]interface{}, context activity.Context) (done bool, err error) {

	activityLog.Debugf("Converting input to bytes ")
	inputBytes, err := json.Marshal(input)
	inputParams := &firehose.DeleteDeliveryStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse input data error : %s", err.Error()), "AWSKINESIS-DELETESTREAM-9004", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse input data error : %s", err.Error()), "AWSKINESIS-DELETESTREAM-9004", nil)
	}

	out, err := svc.DeleteDeliveryStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to Delete stream due to error:%s.", err.Error()), "AWSKINESIS-DELETESTREAM-9003", nil)
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
	activityLog.Debugf("Deleted Delivery Stream Successfully\n", aws.StringValue(inputParams.DeliveryStreamName))

	return true, nil
}
