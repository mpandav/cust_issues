package producer

import (
	"encoding/json"
	"fmt"
	"strings"

	"time"

	"github.com/Shopify/sarama"
	"github.com/linkedin/goavro"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/support/trace"
	"github.com/tibco/wi-contrib/engine/jsonschema"
	"github.com/tibco/wi-plugins/contributions/kafka/src/app/Kafka/connector/kafka"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {

	_ = activity.Register(&MyActivity{}, New)
}

var (
	codec *goavro.Codec
)

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MyActivity{}, nil
}

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
}

func (*MyActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (producerAct *MyActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debugf("Executing Kafka Producer Activity")

	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	ksc, _ := input.Connection.(*kafka.KafkaSharedConfigManager)

	key := context.Name() + "#" + context.ActivityHost().Name()
	producer := ksc.GetProducer(key)

	if producer == nil {
		//Create and cache producer
		clientConnconfig := ksc.GetClientConfiguration()
		producerConfig := clientConnconfig.CreateProducerConfig()
		producerConfig.Producer.Return.Successes = true
		producerConfig.Producer.Return.Errors = true
		ackMode := input.AckMode
		if ackMode == "None" {
			producerConfig.Producer.RequiredAcks = sarama.NoResponse
		} else if ackMode == "Leader" {
			producerConfig.Producer.RequiredAcks = sarama.WaitForLocal
		} else {
			producerConfig.Producer.RequiredAcks = sarama.WaitForAll
		}

		if input.Partitioner == "Manual" {
			producerConfig.Producer.Partitioner = sarama.NewManualPartitioner
		}

		if producerConfig.Producer.RequiredAcks == sarama.WaitForAll {
			timeOut := input.AckTimeout
			if timeOut > 0 {
				producerConfig.Producer.Timeout = time.Duration(timeOut) * time.Millisecond
			} else {
				eMessage := fmt.Sprintf("Invalid 'Ack Timeout' configuration for Activity[%s] in Flow[%s]. It must be a valid positive and greater than 0 value.", context.Name(), context.ActivityHost().Name())
				context.Logger().Error(eMessage)
				return false, activity.NewError(eMessage, "", nil)
			}
		}

		compressionType := input.CompressionType
		if compressionType == "Snappy" {
			producerConfig.Producer.Compression = sarama.CompressionSnappy
		} else if ackMode == "GZIP" {
			producerConfig.Producer.Compression = sarama.CompressionGZIP
		} else if ackMode == "LZ4" {
			producerConfig.Producer.Compression = sarama.CompressionLZ4
		} else {
			producerConfig.Producer.Compression = sarama.CompressionNone
		}

		maxRequestSize := input.MaxRequestSize
		if maxRequestSize > 0 {
			producerConfig.Producer.Flush.Bytes = maxRequestSize
		} else {
			eMessage := fmt.Sprintf("Invalid 'Max Request Size' configuration for Activity[%s] in Flow[%s]. It must be a valid positive and greater than 0 value.", context.Name(), context.ActivityHost().Name())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		}

		maxMessages := input.MaxMessages
		if maxMessages >= 0 {
			producerConfig.Producer.Flush.MaxMessages = maxMessages
		} else {
			eMessage := fmt.Sprintf("Invalid 'Max Messages' configuration for Activity[%s] in Flow[%s]. It must be a valid positive value.", context.Name(), context.ActivityHost().Name())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		}

		frequency := input.Frequency
		if frequency > 0 {
			producerConfig.Producer.Flush.Frequency = time.Duration(frequency) * time.Millisecond
		} else {
			eMessage := fmt.Sprintf("Invalid 'Frequency' configuration for Activity[%s] in Flow[%s]. It must be a valid positive and greater than 0 value.", context.Name(), context.ActivityHost().Name())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		}

		producer, err = sarama.NewSyncProducer(clientConnconfig.Brokers, producerConfig)
		if err != nil {
			if strings.Contains(err.Error(), "Authentication failed") {
				eMessage := fmt.Sprintf("Authentication failed for Activity[%s] in Flow[%s]: %s", context.Name(), context.ActivityHost().Name(), err.Error())
				context.Logger().Error(eMessage)
				// Prevent retry on error for authentication errors
				return false, activity.NewError(eMessage, "", nil)
			} else {
				eMessage := fmt.Sprintf("Failed to create producer for Activity[%s] in Flow[%s] due to error - %s. Enable debug logs for more details.", context.Name(), context.ActivityHost().Name(), err.Error())
				context.Logger().Error(eMessage)
				// Allow retry for other errors
				return false, activity.NewRetriableError(eMessage, "", nil)
			}
		}
		ksc.AddProducer(key, producer)
	}

	message := &sarama.ProducerMessage{}

	recKey := input.Key
	if recKey != "" {
		message.Key = sarama.StringEncoder(recKey)
	}

	topic := input.Topic
	if topic == "" {
		eMessage := fmt.Sprintf("Invalid 'topic' configuration for Activity[%s] in Flow[%s]. It must be a non empty valid string value.", context.Name(), context.ActivityHost().Name())
		context.Logger().Error(eMessage)
		return false, activity.NewError(eMessage, "", nil)
	}

	message.Topic = topic

	partitionId := input.Partition
	if input.Partitioner == "Manual" && partitionId < 0 {
		eMessage := fmt.Sprintf("Invalid configuration. The 'partitionId' must be set for Activity[%s] in Flow[%s] when partitioner mode is set to 'Manual'.", context.Name(), context.ActivityHost().Name())
		context.Logger().Error(eMessage)
		return false, activity.NewError(eMessage, "", nil)
	} else {
		message.Partition = partitionId
	}

	valType := input.ValueType
	if valType == "String" {
		valueIntf := input.StringValue
		if valueIntf == "" {
			eMessage := fmt.Sprintf("A valid string value must be set for Activity[%s] in Flow[%s].", context.Name(), context.ActivityHost().Name())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		}
		message.Value = sarama.StringEncoder(valueIntf)
	} else if valType == "JSON" {
		var jsonData []byte

		if input.JsonValue == nil {
			eMessage := fmt.Sprintf("A valid JSON value must be set for Activity[%s] in Flow[%s].", context.Name(), context.ActivityHost().Name())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		}

		jsonData, err = json.Marshal(input.JsonValue)
		if err != nil {
			eMessage := fmt.Sprintf("Failed to read JSON value for Activity[%s] in Flow[%s] due to error - %s", context.Name(), context.ActivityHost().Name(), err.Error())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		} else {
			message.Value = sarama.ByteEncoder(jsonData)
		}
	} else if valType == "Avro" {

		if input.AvroData == nil {
			eMessage := fmt.Sprintf("A valid Avro Schema must be set for Activity[%s] in Flow[%s].", context.Name(), context.ActivityHost().Name())
			context.Logger().Error(eMessage)
			return false, activity.NewError(eMessage, "", nil)
		}
		schemaInput := " "
		if sIO, ok := context.(schema.HasSchemaIO); ok {
			outputSchema := sIO.GetInputSchema("avroData")
			schemaInput = outputSchema.Value()
		} else {
			context.Logger().Debugf("No schema")
			eMessage := fmt.Sprintf("A valid Avro Schema must be set for Activity[%s] in Flow[%s].", context.Name(), context.ActivityHost().Name())
			return false, activity.NewError(eMessage, "", nil)
		}

		schemaByte := []byte(schemaInput)
		var schemaJSON map[string]interface{}
		json.Unmarshal(schemaByte, &schemaJSON)

		rootSchemaName := schemaJSON["name"]
		var name string

		if rootSchemaName == nil {
			if !isPrimitiveType(schemaJSON) {
				if schemaJSON["type"] == "array" && !isPrimitiveType(schemaJSON["items"]) {
					items := schemaJSON["items"].(map[string]interface{})
					tempName := items["name"]
					if tempName != nil || len(tempName.(string)) != 0 {
						name = items["name"].(string)
					} else {
						name = "records"
					}
				} else {
					name = "records"
				}
			} else {
				name = "records"
			}
		} else {
			name = rootSchemaName.(string)
		}
		messageTemp := input.AvroData

		if messageTemp == nil {
			return false, fmt.Errorf("No Data specified")
		} else {
			jsonBytes, err := json.Marshal(messageTemp)
			jsonMessage := string(jsonBytes)
			err = jsonschema.ValidateFromObject(jsonMessage, messageTemp)
			if err != nil {
				return false, fmt.Errorf("Message validation error %s", err.Error())
			}
		}

		record := messageTemp[name]

		if record != nil {
			record = record.(interface{})
		} else {
			record = messageTemp
		}

		//json schema validation
		if schemaInput != "" {
			jsonBytes, err := json.Marshal(schemaInput)
			if err != nil {
				return false, fmt.Errorf(err.Error())
			}

			jsonSchema := string(jsonBytes)
			err = jsonschema.ValidateFromObject(jsonSchema, schemaInput)
			if err != nil {
				return false, fmt.Errorf("Schema validation error %s", err.Error())
			}
		}

		codec, _ = goavro.NewCodec(schemaInput)
		var binary []byte

		fieldMap := convertJSONToAvro(record)

		context.Logger().Debug(fieldMap)
		binary, _ = codec.BinaryFromNative(nil, fieldMap)

		if binary == nil || len(binary) <= 0 {
			return false, activity.NewError("Serialization error: Please check Avro schema or inputs, schema is wrong or some required inputs are missing", "", "")
		}
		message.Value = sarama.ByteEncoder(binary)

	}

	if input.Headers != nil {

		for k, v := range input.Headers {
			var recHeader sarama.RecordHeader
			recHeader.Key = []byte(k)
			recHeader.Value, err = getBytes(v)
			if err != nil {
				eMessage := fmt.Sprintf("Producer[%s] in Flow[%s] failed to send record due to header processing error - %s", context.Name(), context.ActivityHost().Name(), err.Error())
				context.Logger().Error(eMessage)
				return false, activity.NewError(eMessage, "", nil)
			}
			message.Headers = append(message.Headers, recHeader)
			context.Logger().Debugf("Header - '%s' set in the message", k)
		}

	}
	if trace.Enabled() {
		tracingHeader := make(map[string]string)
		_ = trace.GetTracer().Inject(context.GetTracingContext(), trace.TextMap, tracingHeader)
		for headerKey, headerValue := range tracingHeader {
			message.Headers = append(message.Headers, sarama.RecordHeader{
				Key:   []byte(headerKey),
				Value: []byte(headerValue),
			})
		}
	}

	pid, offset, err := producer.SendMessage(message)
	if err != nil {
		eMessage := fmt.Sprintf("Producer[%s] in Flow[%s] failed to send record due to error - %s", context.Name(), context.ActivityHost().Name(), err.Error())
		context.Logger().Error(eMessage)
		// Check for authentication error and prevent retry
		if strings.Contains(err.Error(), "Authentication failed") {
			return false, activity.NewError(eMessage, "", nil) // Prevent retry for authentication errors
		} else {
			return false, activity.NewRetriableError(eMessage, "", nil) // Allow retry for other errors
		}
	}

	output := &Output{}
	output.Partition = pid
	output.Offset = offset
	output.Topic = topic

	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	currentTime := time.Now().Format(time.RFC3339)
	context.Logger().Infof("Record successfully sent to Topic[%s], Partition[%d], Offset[%d], Timestamp[%s]", topic, pid, offset, currentTime)
	context.Logger().Debugf("Kafka Producer Activity successfully executed")
	return true, nil
}

func getBytes(v interface{}) ([]byte, error) {

	nv, err := coerce.ToString(v)
	if err != nil {
		return nil, err
	}
	return []byte(nv), nil
}
func isPrimitiveType(schema interface{}) bool {

	switch schema.(type) {
	case string:
		schemaType := schema.(string)
		if schemaType == "enum" || schemaType == "fixed" || schemaType == "string" || schemaType == "int" || schemaType == "float" || schemaType == "boolean" || schemaType == "long" || schemaType == "double" || schemaType == "bytes" {
			return true
		}
		break
	default:
		return false
	}

	return false
}
func convertJSONToAvro(record interface{}) interface{} {
	var parent interface{}
	switch record.(type) {
	case map[string]interface{}:
		var mapOfFields map[string]interface{}
		result := make(map[string]interface{})
		mapOfFields = record.(map[string]interface{})
		for k := range mapOfFields {
			if k == "map" {
				//mapResult := make(map[string]interface{})
				mapElements := mapOfFields[k].([]interface{})
				for _, element := range mapElements {
					record := element.(map[string]interface{})
					key := record["key"].(string)
					result[key] = convertJSONToAvro(record["value"])
				}
				//result[k] = mapResult
			} else if k == "arrayItem" {
				return mapOfFields[k]
			} else {
				result[k] = convertJSONToAvro(mapOfFields[k])
			}
		}
		parent = result
		break
	case []interface{}:
		recordElements := record.([]interface{})
		result := make([]interface{}, len(recordElements))
		for index, element := range recordElements {
			result[index] = convertJSONToAvro(element)
		}
		parent = result
		break
	default:
		parent = record
		// context.Logger().Errorf("No record specified")
		// return false, fmt.Errorf("No record specified")
	}
	return parent
}
