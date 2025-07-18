package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/linkedin/goavro"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"
	"github.com/tibco/wi-contrib/engine/jsonschema"
	"github.com/tibco/wi-plugins/contributions/kafka/src/app/Kafka/connector/kafka"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&MyTrigger{}, &MyTriggerFactory{})
}

var (
	codec *goavro.Codec
)

// MyTriggerFactory My Trigger factory
type MyTriggerFactory struct {
	metadata *trigger.Metadata
}

// NewFactory create a new Trigger factory
func NewFactory(md *trigger.Metadata) trigger.Factory {
	return &MyTriggerFactory{metadata: md}
}

// New Creates a new trigger instance for a given id
func (t *MyTriggerFactory) New(config *trigger.Config) (trigger.Trigger, error) {

	s := &Settings{}
	var err error
	s.Connection, err = kafka.GetSharedConfiguration(config.Settings["kafkaConnection"])

	if err != nil {
		return nil, err
	}

	return &MyTrigger{metadata: t.metadata, settings: s, id: config.Id}, nil
}

// Metadata implements trigger.Factory.Metadata
func (*MyTriggerFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

//var log = logger..GetLogger("kafka-trigger-consumer")

// MyTrigger is a stub for your Trigger implementation
type MyTrigger struct {
	metadata  *trigger.Metadata
	settings  *Settings
	consumers []*consumer
	logger    log.Logger
	id        string
}

type consumer struct {
	handler            trigger.Handler
	payloadType        string
	brokers            []string
	topic              []string
	groupId            string
	clientConfig       *sarama.Config
	kConsumer          sarama.ConsumerGroup
	done               chan bool
	logger             log.Logger
	schemaDefined      string
	responseSchemaChan chan responseSchema
	context            context.Context
	cancelFunc         context.CancelFunc
	flowControl        chan bool
	//PartitionConsumer Params
	partitionId       int32
	partitionConsumer sarama.PartitionConsumer
	consumer          sarama.Consumer
	customPartition   bool
	initialOffset     string
	offset            int64
	timestamp         string
}

type responseSchema struct {
	ready bool
	err   error
}

func (t *MyTrigger) Initialize(ctx trigger.InitContext) error {
	t.logger = ctx.Logger()
	ksc, _ := t.settings.Connection.(*kafka.KafkaSharedConfigManager)

	config := ksc.GetClientConfiguration()

	for _, handler := range ctx.GetHandlers() {

		s := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), s, true)
		if err != nil {
			return err
		}

		consumerConfig := config.CreateConsumerConfig()
		tags := make(map[string]string)
		if s.InitialOffset == "Newest" {
			consumerConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
			tags["offset_type"] = "Newest"
		} else {
			consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
			tags["offset_type"] = "Oldest"
		}

		commitInterval := s.CommitInterval
		if commitInterval > 0 {
			consumerConfig.Consumer.Offsets.CommitInterval = time.Duration(commitInterval) * time.Millisecond
		}

		fetchMinBytes := s.FetchMinBytes
		if fetchMinBytes > 0 {
			consumerConfig.Consumer.Fetch.Min = int32(fetchMinBytes)
		}
		consumerConfig.Consumer.Return.Errors = true

		fetchMaxWait := s.FetchMaxWait
		if fetchMaxWait > 0 {
			consumerConfig.Consumer.MaxWaitTime = time.Duration(fetchMaxWait) * time.Millisecond
		}
		heartbeatInterval := s.HeartbeatInterval
		if heartbeatInterval > 0 {
			consumerConfig.Consumer.Group.Heartbeat.Interval = time.Duration(heartbeatInterval) * time.Millisecond
		}

		sessionTimeout := s.SessionTimeout
		if sessionTimeout > 0 {
			consumerConfig.Consumer.Group.Session.Timeout = time.Duration(sessionTimeout) * time.Millisecond
		}

		tConsumer := &consumer{}
		tConsumer.logger = handler.Logger()
		tConsumer.handler = handler
		tConsumer.payloadType = s.ValueType
		tConsumer.partitionId = s.PartitionId
		tConsumer.customPartition = s.CustomPartition
		tConsumer.initialOffset = s.InitialOffset
		tConsumer.offset = s.Offset
		tConsumer.timestamp = s.TimeStamp
		handler.Logger().Debugf("PayloadType - %s", tConsumer.payloadType)
		tConsumer.brokers = config.Brokers
		if tConsumer.payloadType == "Avro" {
			hs := handler.Schemas()
			if hs != nil {
				aSchema := hs.Output["avroData"]
				if aSchema != nil {
					sObject, ok := aSchema.(schema.Schema)
					if ok {
						tConsumer.schemaDefined = sObject.Value()
					} else {
						schemaObj, err := coerce.ToObject(aSchema)
						if err != nil {
							handler.Logger().Debugf("error while typecasting")
							return fmt.Errorf("Avro Schema must be configured")
						}
						tConsumer.schemaDefined, _ = coerce.ToString(schemaObj["value"])
					}
				}
			} else {
				return fmt.Errorf("Failed to get Avro Schemas from the configuration")
			}
		}

		tags["payload_type"] = tConsumer.payloadType
		if s.Topic != "" {
			tConsumer.topic = strings.Split(s.Topic, ",")
			tags["topic_name"] = s.Topic
		} else {
			return fmt.Errorf("Topic must be configured kafka consumer.")
		}

		hc, ok := handler.(trigger.HandlerEventConfig)
		if ok {
			hc.SetDefaultEventData(tags)
		}
		tConsumer.groupId = s.ConsumerGroup
		tConsumer.clientConfig = consumerConfig
		tConsumer.done = make(chan bool)
		t.consumers = append(t.consumers, tConsumer)
	}

	return nil
}

// Metadata implements trigger.Trigger.Metadata
func (t *MyTrigger) Metadata() *trigger.Metadata {
	return t.metadata
}

// Start implements trigger.Trigger.Start
func (t *MyTrigger) Start() error {

	t.logger.Infof("Starting Trigger - %s", t.id)

	for _, consumerConfig := range t.consumers {
		consumerConfig.responseSchemaChan = make(chan responseSchema)
		consumerConfig.flowControl = make(chan bool)
		consumerConfig.context, consumerConfig.cancelFunc = context.WithCancel(context.Background())

		if consumerConfig.customPartition {
			client, err := sarama.NewClient(consumerConfig.brokers, consumerConfig.clientConfig)
			if err != nil {
				return fmt.Errorf("error creating kafka client %v", err.Error())
			}

			switch consumerConfig.initialOffset {
			case "Newest":
				consumerConfig.offset = sarama.OffsetNewest
			case "Oldest":
				consumerConfig.offset = sarama.OffsetOldest
			case "Seek By Offset":
				//already assigned offset in Initialize func
			case "Seek By Timestamp":
				if consumerConfig.timestamp == "" {
					return fmt.Errorf("timestamp cannot be empty for Seek By Timestamp")
				}
				// Validate the timestamp format
				if err := validateTimestamp(consumerConfig.timestamp); err != nil {
					return err
				}
				//for timestamp if its a consumer group model we take the offset for first topic and partitionId 0
				consumerConfig.offset, err = getOffsetByTimestamp(client, consumerConfig.topic[0], consumerConfig.partitionId, consumerConfig.timestamp)
				if err != nil {
					return fmt.Errorf("error getting offset by timestamp: %v", err)
				}
			}
			consumer, err := sarama.NewConsumerFromClient(client)
			if err != nil {
				return fmt.Errorf("error creating kafka consumer %v", err.Error())
			}
			consumerConfig.consumer = consumer
			// Consumer topic starts only the first topic
			if len(consumerConfig.topic) > 1 {
				t.logger.Warnf("Partition consumer will start on first topic [%s] only", consumerConfig.topic[0])
			}

			partitionConsumer, err := consumer.ConsumePartition(consumerConfig.topic[0], consumerConfig.partitionId, consumerConfig.offset)
			if err != nil {
				return fmt.Errorf("error creating partition consumer %v", err.Error())
			}

			consumerConfig.partitionConsumer = partitionConsumer
			// Receive messages from the specified partition
			go func() {
				msgChan := partitionConsumer.Messages()
				errChan := partitionConsumer.Errors()
				for {
					select {
					case message, open := <-msgChan:
						if !open {
							return
						}
						if app.EnableFlowControl() {
							go consumerConfig.processMessage(message, nil)
						} else {
							consumerConfig.processMessage(message, nil)
						}
					case error := <-errChan:
						consumerConfig.logger.Errorf(error.Err.Error())
					case <-consumerConfig.flowControl:
						// Waiting for runner queue to be empty and resume be called
						<-consumerConfig.flowControl
					}
				}
			}()

		} else {
			consumerGroup, err := sarama.NewConsumerGroup(consumerConfig.brokers, consumerConfig.groupId, consumerConfig.clientConfig)
			if err != nil {
				t.logger.Errorf("Failed to create consumer due to error - %s. Enable debug logs for more details.", err.Error())
				return err
			}

			consumerConfig.kConsumer = consumerGroup
			go func() {
				for {
					if err = consumerGroup.Consume(consumerConfig.context, consumerConfig.topic, consumerConfig); err != nil {
						t.logger.Errorf("Error from consumer: %v", err)
						consumerConfig.responseSchemaChan <- responseSchema{ready: false, err: err}
						return
					}
					if consumerConfig.context.Err() != nil {
						return
					}
				}
			}()

			//block execution till consumer group handler is not ready
			response := <-consumerConfig.responseSchemaChan
			if !response.ready {
				t.logger.Errorf(response.err.Error())
				return response.err
			}
		}
	}
	t.logger.Infof("Trigger - %s  started", t.id)
	return nil
}
func validateTimestamp(timestamp string) error {
	if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
		return fmt.Errorf("invalid timestamp format")
	}
	return nil
}
func getOffsetByTimestamp(client sarama.Client, topic string, partition int32, timestamp string) (int64, error) {
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return 0, fmt.Errorf("error parsing timestamp: %v", err)
	}

	// Convert time to milliseconds since epoch
	timestampMs := parsedTime.UnixNano() / 1e6
	offset, err := client.GetOffset(topic, partition, timestampMs)
	if err != nil {
		return 0, fmt.Errorf("error fetching offset: %v", err)
	}
	return offset, nil
}

// sarama consumer group life cycle methods
func (consumer *consumer) Setup(session sarama.ConsumerGroupSession) error {
	consumer.responseSchemaChan <- responseSchema{ready: true, err: nil}
	return nil
}

func (consumer *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for {
		select {
		case msg := <-claim.Messages():
			consumer.logger.Infof("Record received for Topic[%s]", msg.Topic)
			if app.EnableFlowControl() {
				go consumer.processMessage(msg, session)
			} else {
				consumer.processMessage(msg, session)
			}
		case err := <-consumer.kConsumer.Errors():
			consumer.logger.Errorf("Failed to poll records due to error - '%v'", err)
		case <-consumer.done:
			consumer.logger.Info("Polling is stopped")
			consumer.cancelFunc()
			return nil
		case <-session.Context().Done():
			return nil
		case <-consumer.flowControl:
			// Waiting for runner queue to be empty and resume be called
			<-consumer.flowControl
		}
	}
}

func (consumer *consumer) processMessage(msg *sarama.ConsumerMessage, session sarama.ConsumerGroupSession) {
	consumer.logger.Debugf("Processing record from Topic[%s], Partition[%d], Offset[%d]", msg.Topic, msg.Partition, msg.Offset)

	out := &Output{}
	out.Topic = msg.Topic
	out.Offset = msg.Offset
	out.Partition = int(msg.Partition)
	out.Key = string(msg.Key)

	if msg.Value != nil {
		deserVal := consumer.payloadType
		//deserVal := "Avro"
		if deserVal == "String" {
			out.StringValue = string(msg.Value)
		} else if deserVal == "JSON" {
			var mdata map[string]interface{}
			err := json.Unmarshal(msg.Value, &mdata)
			if err != nil {
				consumer.logger.Errorf("Failed to process record from Topic[%s], Partition[%d], Offset[%d] due to incompatible value type. Received value is not a valid JSON. Make sure your publisher is configured appropriately to send a valid JSON value.", msg.Topic, msg.Partition, msg.Offset)
				return
			}
			out.JsonValue = mdata
		} else if deserVal == "Avro" {

			schemaInput := consumer.schemaDefined
			if schemaInput != "" {
				jsonBytes, _ := json.Marshal(schemaInput)
				jsonSchema := string(jsonBytes)
				err := jsonschema.ValidateFromObject(jsonSchema, schemaInput)
				if err != nil {
					consumer.logger.Errorf("Schema validation error %s", err.Error())
					return
				}
			}

			schemaByte := []byte(schemaInput)
			var schemaJSON map[string]interface{}
			json.Unmarshal(schemaByte, &schemaJSON)
			rootSchemaName := schemaJSON["name"]
			var name string

			isRootArray := false
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
						isRootArray = true
					} else {
						name = "records"
					}
				} else {
					name = "records"
				}
			} else {
				name = rootSchemaName.(string)
			}

			message := msg.Value
			if message == nil {
				consumer.logger.Errorf("Error on receiving data on Topic[%s], Partition[%d], Offset[%d] due to incompatible value type. Received value is not a valid JSON. Make sure your publisher ", msg.Topic, msg.Partition, msg.Offset)
				return
			}

			codec, _ := goavro.NewCodec(schemaInput)
			native, _, err := codec.NativeFromBinary(msg.Value)

			if err != nil {
				consumer.logger.Errorf("Deserialization error %s, Check if you have set the same schema on producer and consumer", "", err.Error())
				return
			}

			tmpMap := make(map[string]interface{})
			var result interface{}

			if name == "records" || isRootArray || schemaJSON["type"] == "record" {
				tmpMap[name] = convertAvroToJSON(native, schemaJSON)
				result = tmpMap
			} else {
				consumer.logger.Debugf("Inside else %S", schemaJSON["type"])
				tmpMap[name] = native
				result = convertAvroToJSON(tmpMap, schemaJSON)
			}

			outputData, _ := json.Marshal(result)
			consumer.logger.Debugf("\n\n==>Received Value:\n%s\n\n", string(outputData))
			consumer.logger.Debugf("type of outputData is ")
			consumer.logger.Debugf(fmt.Sprintf("%T", outputData))
			//consumer.logger.Infof("type of avroValue is ")
			//consumer.logger.Infof(fmt.Sprintf("%T", triggerData["avroValue"]))
			//triggerData["avroValue"].(*data.ComplexObject).Value = outputData
			out.AvroData = tmpMap

		}
	}

	out.Headers = make(map[string]interface{})
	if len(msg.Headers) > 0 {
		for i := range msg.Headers {
			val, err := getInterface(msg.Headers[i].Value)
			if err != nil {
				consumer.logger.Errorf("Failed to process record from Topic[%s], Partition[%d], Offset[%d] due to header processing error - %s", msg.Topic, msg.Partition, msg.Offset, err.Error())
				return
			}

			hName := string(msg.Headers[i].Key)
			consumer.logger.Debugf("Header - '%s' received in the message", hName)
			out.Headers[hName] = val
		}
		consumer.logger.Infof("Headers - %v", out.Headers)
	}

	eventId := fmt.Sprintf("%s#%d#%d", msg.Topic, msg.Partition, msg.Offset)
	ctx := context.Background()
	if trace.Enabled() {
		tracingHeader := make(map[string]string)
		for key, value := range out.Headers {
			stringValue, ok := value.(string)
			if ok {
				tracingHeader[key] = stringValue
			}
		}
		tc, _ := trace.GetTracer().Extract(trace.TextMap, tracingHeader)
		if tc != nil {
			ctx = trace.AppendTracingContext(ctx, tc)
		}
	}

	_, err := consumer.handler.Handle(trigger.NewContextWithEventId(ctx, eventId), out)
	if err != nil {
		consumer.logger.Errorf("Failed to process record from Topic[%s], Partition[%d], Offset[%d] due to error - %s", msg.Topic, msg.Partition, msg.Offset, err.Error())
	} else {
		// record is successfully processed. Mark as processed.
		consumer.logger.Infof("Record from Topic[%s], Partition[%d], Offset[%d] is successfully processed", msg.Topic, msg.Partition, msg.Offset)
		// Mark message for non partition consumer
		if session != nil {
			session.MarkMessage(msg, "")
		}
	}
}

func getInterface(bts []byte) (interface{}, error) {
	reflect.ValueOf(bts).Interface()
	return string(bts), nil
}

// Stop implements trigger.Trigger.Start
func (t *MyTrigger) Stop() error {
	t.logger.Infof("Stopping Trigger - %s", t.id)
	for _, consumer := range t.consumers {
		// Stop polling for Consumer group model
		if !consumer.customPartition {
			consumer.done <- true
		}
		if consumer.kConsumer != nil {
			consumer.kConsumer.Close()
			consumer.kConsumer = nil
		}

		if consumer.partitionConsumer != nil {
			consumer.partitionConsumer.Close()
		}
		// close partition consumer
		if consumer.consumer != nil {
			consumer.consumer.Close()
		}
	}

	t.logger.Infof("Trigger - %s  stopped", t.id)
	return nil
}

func (t *MyTrigger) Resume() error {
	t.logger.Infof("Resuming Trigger - %s", t.id)
	for _, consumer := range t.consumers {
		// Start polling
		consumer.flowControl <- true
	}
	return nil
}

func (t *MyTrigger) Pause() error {
	t.logger.Infof("Pausing Trigger - %s", t.id)
	for _, consumer := range t.consumers {
		// Stop polling
		consumer.flowControl <- true
	}

	t.logger.Infof("Trigger - %s  paused", t.id)
	return nil
}

func convertAvroToJSON(record interface{}, schema interface{}) interface{} {
	var parent interface{}
	if isPrimitiveType(schema) {
		return record
	}
	schemaMap := schema.(map[string]interface{})
	switch schemaMap["type"] {
	case "array":
		newSchema := schemaMap["items"]

		var mapOfFields []interface{}
		mapOfFields = record.([]interface{})
		result := make([]interface{}, len(mapOfFields))
		for k := range mapOfFields {
			if isPrimitiveType(newSchema) {
				arrayItem := make(map[string]interface{})
				arrayItem["arrayItem"] = mapOfFields[k]
				result[k] = arrayItem
			} else {
				result[k] = convertAvroToJSON(mapOfFields[k], newSchema)
			}
		}
		parent = result
		break
	case "record":
		//recordRoot := make(map[string]interface{})
		newSchema := schemaMap["fields"].([]interface{})
		recordMap := record.(map[string]interface{})
		//recordElements := recordMap[recordName].(map[string]interface{})
		result := make(map[string]interface{})
		//name := schemaMap["name"].(string)
		//fields := recordMap[name].(map[string]interface{})
		//i := 0
		// for element := range fields {
		// 	fieldSchema := newSchema[i].(map[string]interface{})
		// 	if !isPrimitiveType(fieldSchema["type"]) {
		// 		result[element] = convertAvroToJSON(fields[element], fieldSchema["type"])
		// 	} else {
		// 		result[element] = fields[element]
		// 	}
		// 	i++
		// }
		for schemaElement := range newSchema {
			fieldSchema := newSchema[schemaElement].(map[string]interface{})
			fieldName := fieldSchema["name"].(string)
			if !isPrimitiveType(fieldSchema["type"]) {
				innerType := fieldSchema["type"].(map[string]interface{})
				result[fieldName] = convertAvroToJSON(recordMap[fieldName], innerType)
			} else {
				result[fieldName] = recordMap[fieldName]
			}
		}
		//recordRoot[name] = result
		parent = result
		break
	case "map":
		newSchema := schemaMap["values"].(interface{})
		recordElements := record.(map[string]interface{})
		rootMap := make(map[string]interface{})
		result := make([]interface{}, len(recordElements))
		i := 0
		for key := range recordElements {
			mapElement := make(map[string]interface{})
			mapElement["key"] = key
			if !isPrimitiveType(newSchema) {
				mapElement["value"] = convertAvroToJSON(recordElements[key], newSchema)
			} else {
				mapElement["value"] = recordElements[key]
			}
			result[i] = mapElement
			i++
		}
		rootMap["map"] = result
		parent = rootMap
		break
	default:
		if isPrimitiveType(schemaMap["type"]) {
			parent = record
		} else {

			//schemaName := schemaMap["name"].(string)
			//recordElements := record.(map[string]interface{})
			//parent = convertAvroToJSON(recordElements[schemaName], schemaMap["type"])

			//result := make([]interface{}, len(recordElements))

			newSchema := schemaMap["type"].(map[string]interface{})
			newSchemaType := newSchema["type"]

			if newSchemaType == "record" {
				//schemaName := schemaMap["name"].(string)
				//recordElements := record.(map[string]interface{})
				parent = convertAvroToJSON(record, newSchema)
			} else {
				// This works for Map but fails for Record of Records
				recordElements := record.([]interface{})
				parent = convertAvroToJSON(recordElements, schemaMap["type"])
			}
		}
	}
	return parent
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
