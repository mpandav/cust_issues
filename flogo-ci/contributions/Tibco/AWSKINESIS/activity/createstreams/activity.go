package createstreams

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/conversion"
	mkutil "github.com/tibco/wi-plugins/contributions/awsKinesis/src/app/AWSKINESIS/activity/util"
)

var activityLog = log.ChildLogger(log.RootLogger(), "activity.awskinesis.createstream")

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
	activityLog.Debugf("Executing Kinesis CreateStream activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		activityLog.Errorf("Error while parsing Input Object, %v", err)
		return false, err
	}

	inputMap := input.Input
	inputMap["DeliveryStreamConfiguration"] = input.DeliveryStreamConfiguration
	inputMap["KinesisStreamSourceConfiguration"] = input.KinesisStreamSourceConfiguration
	//activityLog.Info(inputMap)
	if inputMap == nil {
		return false, activity.NewError(fmt.Sprintf("Input is required in query activity for"), "AWSKINESIS-CREATESTREAM-9001", nil)
	}

	streamType := input.StreamType
	deliveryStreamType := input.DeliveryStreamType
	destinationType := input.DestinationType

	session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if session.Config.Endpoint != nil {
		endpoint := *session.Config.Endpoint
		session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Failed to connect to AWS <CausedBy>: %s. Check credentials configured in the connection ", err.Error()), "AWS-CONFIG-9100", err.Error())
	}

	if streamType == "DataStream" {
		activityLog.Debugf("Creating Data Stream \n")
		kc := kinesis.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session \n")
		return ExecuteCreateDataStream(kc, inputMap, context)
	} else if streamType == "VideoStream" {
		activityLog.Debugf("Creating Video Stream \n")
		vkc := kinesisvideo.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session \n")
		return ExecuteCreateVideoStream(vkc, inputMap, context)
	} else if streamType == "Firehose-DeliveryStream" {
		activityLog.Debugf("Creating Delivery Stream \n")
		dkc := firehose.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session \n")
		return ExecuteCreateDeliveryStream(dkc, deliveryStreamType, destinationType, inputMap, context)
	}
	return true, nil
}

// ExecuteCreateDataStream function is specific to DataStream creation
func ExecuteCreateDataStream(kc *kinesis.Kinesis, input map[string]interface{}, context activity.Context) (done bool, err error) {

	inputBytes, err := json.Marshal(input)
	inputParams := &kinesis.CreateStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Convert Input Data To Bytes Error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9003", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse Input Data Error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9004", nil)
	}
	//read the values from configuration
	activityLog.Debugf("Creating DataStream " + aws.StringValue(inputParams.StreamName) + " With shard count " + strconv.FormatInt(aws.Int64Value(inputParams.ShardCount), 10))

	_, err = kc.CreateStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to create stream due to error:%s.", err.Error()), "AWSKINESIS-CREATESTREAM-9005", nil)
	}
	outMsg := map[string]interface{}{
		"StreamName": aws.StringValue(inputParams.StreamName),
	}
	// o, _ := json.Marshal(out)
	// json.Unmarshal(o, &outMsg)
	// outMsg["StreamName"] = aws.StringValue(inputParams.StreamName)
	output := &Output{}
	output.Message = outMsg

	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	activityLog.Debugf("Created DataStream Successfully \n", aws.StringValue(inputParams.StreamName))

	return true, nil
}

// ExecuteCreateVideoStream function is to create Kinesis Video Stream
func ExecuteCreateVideoStream(vkc *kinesisvideo.KinesisVideo, input map[string]interface{}, context activity.Context) (done bool, err error) {

	//read the values from configuration
	inputBytes, err := json.Marshal(input)
	inputParams := &kinesisvideo.CreateStreamInput{}

	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Convert Input Data To Bytes Error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9003", nil)
	}
	err = json.Unmarshal(inputBytes, inputParams)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Parse Input Data Error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9004", nil)
	}

	activityLog.Debugf("Creating Video Stream with the following input \n", inputParams)

	out, err := vkc.CreateStream(inputParams)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to create stream due to error:%s.", err.Error()), "AWSKINESIS-CREATESTREAM-9005", nil)
	}

	activityLog.Debugf("Created Video Stream Successfully with following config %v", input)
	var outMsg map[string]interface{}
	o, _ := json.Marshal(out)
	json.Unmarshal(o, &outMsg)
	output := &Output{}
	output.Message = outMsg
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ExecuteCreateDeliveryStream function is to create Kinesis Delivery Stream
func ExecuteCreateDeliveryStream(dkc *firehose.Firehose, deliveryStreamType string, destinationType string, input map[string]interface{}, context activity.Context) (done bool, err error) {

	inputParam := &firehose.CreateDeliveryStreamInput{}

	// deliverystreamInput, _ := coerce.ToObject(input.Input)
	deliverystreamInput := input
	var deliverystreamName string

	activityLog.Debugf("Delivery Stream Input ", deliverystreamInput)
	if deliverystreamInput != nil {
		deliverystreamName = deliverystreamInput["DeliveryStreamName"].(string)
		activityLog.Debugf("Delivery Stream Name ", deliverystreamName)
	}

	inputParam.DeliveryStreamName = &deliverystreamName

	activityLog.Debugf("Getting Input to Create Delivery Stream \n")
	inputDSC := input["DeliveryStreamConfiguration"]
	activityLog.Debugf("Converting Input to Bytes \n")

	inputBytes, err := conversion.ConvertToBytes(inputDSC)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Convert Input Data To Bytes Error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9003", nil)
	}

	if destinationType == "ExtendedS3DestinationConfiguration" {
		extendedS3DestinationConfiguration := &firehose.ExtendedS3DestinationConfiguration{}
		err = json.Unmarshal(inputBytes, extendedS3DestinationConfiguration)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse ExtendedS3DestinationConfiguration input data error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9007", nil)
		}
		inputParam.ExtendedS3DestinationConfiguration = extendedS3DestinationConfiguration

	} else if destinationType == "S3DestinationConfiguration" {
		S3DestinationConfiguration := &firehose.S3DestinationConfiguration{}
		err = json.Unmarshal(inputBytes, S3DestinationConfiguration)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse S3DestinationConfiguration input data error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9008", nil)
		}
		inputParam.S3DestinationConfiguration = S3DestinationConfiguration

	} else if destinationType == "ElasticsearchDestinationConfiguration" {
		elasticsearchDestinationConfiguration := &firehose.ElasticsearchDestinationConfiguration{}
		err = json.Unmarshal(inputBytes, elasticsearchDestinationConfiguration)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse ElasticsearchDestinationConfiguration input data error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9009", nil)
		}
		inputParam.ElasticsearchDestinationConfiguration = elasticsearchDestinationConfiguration

	} else if destinationType == "RedshiftDestinationConfiguration" {
		redshiftDestinationConfiguration := &firehose.RedshiftDestinationConfiguration{}
		err = json.Unmarshal(inputBytes, redshiftDestinationConfiguration)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse RedshiftDestinationConfiguration input data error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9010", nil)
		}
		inputParam.RedshiftDestinationConfiguration = redshiftDestinationConfiguration

	} else if destinationType == "SplunkDestinationConfiguration" {
		splunkDestinationConfiguration := &firehose.SplunkDestinationConfiguration{}
		err = json.Unmarshal(inputBytes, splunkDestinationConfiguration)
		if err != nil {
			return false, activity.NewError(fmt.Sprintf("Parse SplunkDestinationConfiguration input data error %s", err.Error()), "AWSKINESIS-CREATESTREAM-9011", nil)
		}
		inputParam.SplunkDestinationConfiguration = splunkDestinationConfiguration

	} else {
		return false, activity.NewError("Atleast one Destination type is expected for creating Kinesis Delivery stream", "KINESIS-CREATESTREAM-9012", nil)
	}

	inputParam.DeliveryStreamType = &deliveryStreamType
	if deliveryStreamType == "KinesisStreamAsSource" {
		activityLog.Debugf("Delivery Stream Type Set to KinesisStreamAsSource \n")
		deliveryStream := input["KinesisStreamSourceConfiguration"]
		deliveryStreamBytes, err := conversion.ConvertToBytes(deliveryStream)
		kinesisStreamAsSourceConfiguration := &firehose.KinesisStreamSourceConfiguration{}
		if err != nil {
			return false, activity.NewError("Convert KinesisStreamAsSource DeliveryStream Data to bytes error "+err.Error(), "KINESIS-CREATESTREAM-9013", nil)
		}
		err = json.Unmarshal(deliveryStreamBytes, kinesisStreamAsSourceConfiguration)
		if err != nil {
			return false, activity.NewError("Parse KinesisStreamAsSource input data error "+err.Error(), "KINESIS-CREATESTREAM-9014", nil)
		}
		inputParam.KinesisStreamSourceConfiguration = kinesisStreamAsSourceConfiguration
	}

	activityLog.Debugf("Creating Delivery Stream with the following Input ", inputParam)

	out, err := dkc.CreateDeliveryStream(inputParam)

	if err != nil {
		//set error if we have to and then exit
		mkutil.SetErrorObject(err, context)
		return false, activity.NewError(fmt.Sprintf("Failed to create Delivery Stream due to error:%s.", err.Error()), "AWSKINESIS-CREATESTREAM-9015", nil)
	}

	activityLog.Debugf("Created Delivery Stream Successfully \n")
	var outMsg map[string]interface{}
	o, _ := json.Marshal(out)
	json.Unmarshal(o, &outMsg)
	output := &Output{}
	output.Message = outMsg
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	return true, nil

}
