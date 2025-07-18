package put

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesisvideo"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	mkutil "github.com/tibco/wi-plugins/contributions/awsKinesis/src/app/AWSKINESIS/activity/util"
)

var activityLog = log.ChildLogger(log.RootLogger(), "aws-kinesis-putstream")

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

	activityLog.Debugf("Executing Kinesis Put activity\n")

	activityLog.Debugf("Executing Get activity")

	input := &Input{}
	err = context.GetInputObject(input)

	inputMap := input.Input

	if inputMap == nil {
		return false, activity.NewError(fmt.Sprintf("Input is required in query activity for"), "AWSKINESIS-CREATESTREAM-9001", nil)
	}

	streamType := input.StreamType
	recordType := input.RecordType

	session := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if session.Config.Endpoint != nil {
		endpoint := *session.Config.Endpoint
		session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}

	if streamType == "DataStream" {
		activityLog.Debugf("Put Data In Data Stream\n")
		kc := kinesis.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session\n")
		return ExecutePutDataStream(kc, inputMap, recordType, context)

	} else if streamType == "VideoStream" {
		activityLog.Debugf("Put Data Video Stream\n")
		vkc := kinesisvideo.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session\n")
		return ExecutePutVideoStream(vkc, inputMap, recordType, context)

	} else if streamType == "Firehose-DeliveryStream" {
		activityLog.Debugf("Put Data Into Delivery Stream\n")
		fkc := firehose.New(session, endpointConfig)
		activityLog.Debugf("Created AWS Session\n")
		return ExecutePutDeliveryStream(fkc, inputMap, recordType, context)
	}
	return true, nil
}

// ExecutePutDataStream function is specific to DataStream creation
func ExecutePutDataStream(kc *kinesis.Kinesis, input map[string]interface{}, recordType string, context activity.Context) (done bool, err error) {

	putstreamInput := input

	activityLog.Debugf("Data is ", input)

	output := &Output{}
	var outMsg map[string]interface{}

	if recordType == "SingleRecord" {
		activityLog.Debugf("Putting Single record in Data Stream\n")
		inputParams := &kinesis.PutRecordInput{}

		var recordData interface{}

		//Constructing AWS putRecord Input
		if putstreamInput != nil {
			if putstreamInput["Data"] != nil {
				recordData = putstreamInput["Data"].(interface{})
				data, err := getRecordData(recordData)
				if err != nil {
					return false, activity.NewError(fmt.Sprintf("Error in converting the record Data %s", err.Error()), "KINESIS-PUTSTREAM-9004", nil)
				}
				activityLog.Debugf("Data is ", data)
				inputParams.Data = data
			}
			if putstreamInput["ExplicitHashKey"] != nil {
				ExplicitHashKey := putstreamInput["ExplicitHashKey"].(string)
				inputParams.ExplicitHashKey = &ExplicitHashKey
			}
			if putstreamInput["PartitionKey"] != nil {
				PartitionKey := putstreamInput["PartitionKey"].(string)
				inputParams.PartitionKey = &PartitionKey
			}
			if putstreamInput["SequenceNumberForOrdering"] != nil {
				SequenceNumberForOrdering := putstreamInput["SequenceNumberForOrdering"].(string)
				inputParams.SequenceNumberForOrdering = &SequenceNumberForOrdering
			}
			if putstreamInput["StreamName"] != nil {
				StreamName := putstreamInput["StreamName"].(string)
				inputParams.StreamName = &StreamName
			}
		}

		//read the values from configuration
		activityLog.Debugf("InputParams is ", inputParams)

		out, err := kc.PutRecord(inputParams)
		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Put Data in Data Stream due to error:%s.", err.Error()), "AWSKINESIS-CREATESTREAM-9004", nil)
		}

		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output.Message = outMsg

	} else if recordType == "MultipleRecords" {

		activityLog.Debugf("Putting Multiple records in Data Stream\n")
		inputParams := &kinesis.PutRecordsInput{}
		putstreamInput := input

		//Read Input and construct Object for AWS PutRecordsInput
		if putstreamInput != nil {
			if putstreamInput["StreamName"] != nil {
				StreamName := putstreamInput["StreamName"].(string)
				inputParams.StreamName = &StreamName
			}
			if putstreamInput["Records"] != nil {
				recordData, _ := putstreamInput["Records"].([]interface{})
				if recordData != nil {
					entries := make([]*kinesis.PutRecordsRequestEntry, len(recordData))
					for i := 0; i < len(recordData); i++ {
						curRecord, _ := coerce.ToObject(recordData[i])
						var recordData interface{}
						recordData = curRecord["Data"].(interface{})
						if recordData != nil {
							data, err := getRecordData(recordData)
							if err != nil {
								return false, activity.NewError(fmt.Sprintf("Error in converting the record Data %s", err.Error()), "KINESIS-PUTSTREAM-9004", nil)
							}
							entries[i].Data = data
						}
						if curRecord["ExplicitHashKey"] != nil {
							ExplicitHashKey := curRecord["ExplicitHashKey"].(string)
							entries[i].ExplicitHashKey = &ExplicitHashKey
						}
						if curRecord["PartitionKey"] != nil {
							PartitionKey := curRecord["PartitionKey"].(string)
							entries[i].PartitionKey = &PartitionKey
						}
					}
					inputParams.Records = entries
				}
			}
		}
		out, err := kc.PutRecords(inputParams)

		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Put Records in Data Stream due to error:%s.", err.Error()), "AWSKINESIS-CREATESTREAM-9005", nil)
		}
		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output.Message = outMsg
	}
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	activityLog.Debugf("Data Successfully Put in DataStream\n")

	return true, nil
}

// ExecutePutVideoStream function is specific to DataStream creation
func ExecutePutVideoStream(vkc *kinesisvideo.KinesisVideo, input map[string]interface{}, recordType string, context activity.Context) (done bool, err error) {

	//Not implemented yet
	return true, nil
}

// ExecutePutDeliveryStream function is specific to DataStream creation
func ExecutePutDeliveryStream(fkc *firehose.Firehose, input map[string]interface{}, recordType string, context activity.Context) (done bool, err error) {

	putdeliverystreamInput := input

	activityLog.Debugf("Data is ", input)
	output := &Output{}
	var outMsg map[string]interface{}

	if recordType == "SingleRecord" {
		activityLog.Debugf("Putting Single record in Delivery Stream\n")
		inputParams := &firehose.PutRecordInput{}

		if putdeliverystreamInput != nil {
			if putdeliverystreamInput["DeliveryStreamName"] != nil {
				DeliveryStreamName := putdeliverystreamInput["DeliveryStreamName"].(string)
				inputParams.DeliveryStreamName = &DeliveryStreamName
			}
			if putdeliverystreamInput["Record"] != nil {
				record := putdeliverystreamInput["Record"]
				recordInput, _ := coerce.ToObject(record)
				recordData := recordInput["Data"].(interface{})
				data, err := getRecordData(recordData)
				if err != nil {
					return false, activity.NewError(fmt.Sprintf("Error in converting the record Data %s", err.Error()), "KINESIS-PUTSTREAM-9004", nil)
				}

				entry := &firehose.Record{}
				entry.Data = data
				inputParams.Record = entry
			}
		}
		activityLog.Debugf("InputParams is ", inputParams)

		out, err := fkc.PutRecord(inputParams)
		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Put Data in Data Stream due to error:%s.", err.Error()), "AWSKINESIS-PUTSTREAM-9006", nil)
		}
		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output.Message = outMsg

	} else if recordType == "MultipleRecords" {

		activityLog.Debugf("Putting Multiple records in Delivery Stream\n")
		inputParams := &firehose.PutRecordBatchInput{}

		if putdeliverystreamInput != nil {
			if putdeliverystreamInput["DeliveryStreamName"] != nil {
				DeliveryStreamName := putdeliverystreamInput["DeliveryStreamName"].(string)
				inputParams.DeliveryStreamName = &DeliveryStreamName
			}
			if putdeliverystreamInput["Records"] != nil {
				records, _ := putdeliverystreamInput["Records"].([]interface{})
				if records != nil {
					entries := make([]*firehose.Record, len(records))
					for i := 0; i < len(records); i++ {
						curRecord, _ := coerce.ToObject(records[i])
						var recordData interface{}
						recordData = curRecord["Data"].(interface{})
						data, err := getRecordData(recordData)
						if err != nil {
							return false, activity.NewError(fmt.Sprintf("Error in converting the record Data %s", err.Error()), "KINESIS-PUTSTREAM-9004", nil)
						}
						entries[i].Data = data
					}
					inputParams.Records = entries
				}
			}
		}
		activityLog.Debugf("InputParams is ", inputParams)

		out, err := fkc.PutRecordBatch(inputParams)
		if err != nil {
			//set error if we have to and then exit
			mkutil.SetErrorObject(err, context)
			return false, activity.NewError(fmt.Sprintf("Failed to Put Data Records in Data Stream due to error:%s.", err.Error()), "AWSKINESIS-PUTSTREAM-9007", nil)
		}

		o, _ := json.Marshal(out)
		json.Unmarshal(o, &outMsg)
		output.Message = outMsg

	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	activityLog.Debugf("Put Delivery Stream Successfully executed\n")

	return true, nil
}

// getRecordData converts data from interface{} to []byte
func getRecordData(recordData interface{}) ([]byte, error) {

	var reqBody []byte
	switch recordData.(type) {
	case string:
		activityLog.Debugf("Request Body string type ", recordData.(string))
		reqBody = ([]byte(recordData.(string)))
	default:
		b, err := json.Marshal(recordData)
		if err != nil {
			return nil, err
		}
		activityLog.Debugf("Request Body interface type ", string(b))
		reqBody = ([]byte(b))
	}
	return reqBody, nil
}
