package queuereceiver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"
	connection "github.com/tibco/wi-azservicebus/src/app/AzureServiceBus/connector/connection"
)

// OssUpgrade
var logCache = log.ChildLogger(log.RootLogger(), "azureservicebus-trigger-queuereceiver")

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&SBQueueReceiverTrigger{}, &MyTriggerFactory{})
}

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
	err := metadata.MapToStruct(config.Settings, s, true)

	if err != nil {
		return nil, fmt.Errorf("Error occured in metadata.MapToStruct, error - [%s]", err.Error())
	}
	ascm, _ := s.Connection.(*connection.AzureServiceBusSharedConfigManager)
	return &SBQueueReceiverTrigger{metadata: t.metadata, config: config, ascm: ascm}, nil
}

// Metadata implements trigger.Factory.Metadata
func (*MyTriggerFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

// SBQueueReceiverTrigger is a stub for your Trigger implementation
type SBQueueReceiverTrigger struct {
	metadata       *trigger.Metadata
	config         *trigger.Config
	queueReceivers []*QueueReceiver
	ascm           *connection.AzureServiceBusSharedConfigManager
}

// QueueReceiver is structure of a single QueueReceiver
type QueueReceiver struct {
	handler             trigger.Handler
	queue               *azservicebus.Receiver
	queueSessionRcv     *azservicebus.SessionReceiver
	deadLetter          *azservicebus.Receiver
	ctx                 context.Context
	listenctxCancelFunc context.CancelFunc
	sessionID           string
	queueName           string
	connString          string
	valueType           string
	receiveMode         string
	isDeadLetter        bool
	timeOut             int
	retrycount          int
	retryInterval       int
}

// Initialize QueueReceiverTrigger
func (t *SBQueueReceiverTrigger) Initialize(ctx trigger.InitContext) error {

	ctx.Logger().Info("Initializing AzureServiceBus Trigger Context...")

	for _, handler := range ctx.GetHandlers() {

		handlerSettings := &HandlerSettings{}
		var err error
		err = metadata.MapToStruct(handler.Settings(), handlerSettings, true)
		if err != nil {
			return fmt.Errorf("Error occured in metadata.MapToStruct, error - [%s]", err.Error())
		}

		connStr := ""
		if strings.HasPrefix(t.ascm.AzureToken.ResourceURI, "https") {
			u, err := url.Parse(t.ascm.AzureToken.ResourceURI)
			if err != nil {
				return fmt.Errorf("Unable to parse namespace url %s", err.Error())
			}
			namespace := u.Host
			if u.Path != "" {
				connStr = "Endpoint=sb://" + namespace + "/;SharedAccessKeyName=" + t.ascm.AzureToken.AuthorizationRuleName + ";SharedAccessKey=" + t.ascm.AzureToken.PrimarysecondaryKey + ";EntityPath=" + u.Path
			} else {
				connStr = "Endpoint=sb://" + namespace + "/;SharedAccessKeyName=" + t.ascm.AzureToken.AuthorizationRuleName + ";SharedAccessKey=" + t.ascm.AzureToken.PrimarysecondaryKey
			}
		} else {
			connStr = "Endpoint=sb://" + t.ascm.AzureToken.ResourceURI + ".servicebus.windows.net/;SharedAccessKeyName=" + t.ascm.AzureToken.AuthorizationRuleName + ";SharedAccessKey=" + t.ascm.AzureToken.PrimarysecondaryKey
		}
		qrcvr := &QueueReceiver{}
		qrcvr.handler = handler
		qrcvr.connString = connStr
		qrcvr.queueName = handlerSettings.Queue
		qrcvr.receiveMode = handlerSettings.ReceiveMode
		qrcvr.valueType = handlerSettings.ValueType
		qrcvr.sessionID = handlerSettings.SessionId
		qrcvr.isDeadLetter = handlerSettings.Deadletter
		qrcvr.timeOut = handlerSettings.Timeout
		qrcvr.retrycount = handlerSettings.Count
		qrcvr.retryInterval = handlerSettings.Interval
		t.queueReceivers = append(t.queueReceivers, qrcvr)
	}

	return nil
}

// Start implements trigger.Trigger.Start
func (t *SBQueueReceiverTrigger) Start() (err error) {

	for _, qrcvr := range t.queueReceivers {
		client := t.ascm.ServiceBusClient
		if qrcvr.retrycount > 0 {
			//create new client only if retry counts are different on trigger
			retryOptions := azservicebus.RetryOptions{
				MaxRetries:    int32(qrcvr.retrycount),
				RetryDelay:    time.Duration(qrcvr.retryInterval) * time.Millisecond,
				MaxRetryDelay: time.Duration(qrcvr.retryInterval) * time.Millisecond,
			}
			client, err = azservicebus.NewClientFromConnectionString(qrcvr.connString,
				&azservicebus.ClientOptions{
					RetryOptions: retryOptions,
				})
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
		}

		if qrcvr.isDeadLetter {
			q, err := getDeadLetter(client, qrcvr)
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
			qrcvr.deadLetter = q
			go qrcvr.listenDeadletter()
		} else if len(qrcvr.sessionID) > 0 {
			queueSessionRcv, err := getQueueWithSession(client, qrcvr)
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
			qrcvr.queueSessionRcv = queueSessionRcv
			go qrcvr.listen()
		} else {
			q, err := getQueue(client, qrcvr)
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
			qrcvr.queue = q
			go qrcvr.listen()
		}
	}
	//log.Infof("Trigger - %s  started", t.config.Name)
	return nil
}

func getQueue(client *azservicebus.Client, qrcv *QueueReceiver) (*azservicebus.Receiver, error) {
	if qrcv.timeOut > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(qrcv.timeOut))
		qrcv.ctx = ctx
		qrcv.listenctxCancelFunc = cancel
	} else {
		qrcv.ctx = context.Background()
	}

	if qrcv.receiveMode == "ReceiveAndDelete" {
		logCache.Debugf("Using receiveMode ReceiveAndDelete on queue %s", qrcv.queueName)
		qe, err := client.NewReceiverForQueue(
			qrcv.queueName,
			&azservicebus.ReceiverOptions{
				ReceiveMode: azservicebus.ReceiveModeReceiveAndDelete,
			})
		return qe, err
	}
	logCache.Debugf("Using receiveMode PeekLock on queue %s", qrcv.queueName)
	qe, err := client.NewReceiverForQueue(
		qrcv.queueName,
		&azservicebus.ReceiverOptions{
			ReceiveMode: azservicebus.ReceiveModePeekLock,
		})
	return qe, err
}

func getQueueWithSession(client *azservicebus.Client, qrcv *QueueReceiver) (*azservicebus.SessionReceiver, error) {
	if qrcv.timeOut > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(qrcv.timeOut))
		qrcv.ctx = ctx
		qrcv.listenctxCancelFunc = cancel
	} else {
		qrcv.ctx = context.Background()

	}

	if qrcv.receiveMode == "ReceiveAndDelete" {
		logCache.Debugf("Using receiveMode ReceiveAndDelete on queue %s", qrcv.queueName)

		queueSessionRcv, err := client.AcceptSessionForQueue(
			qrcv.ctx,
			qrcv.queueName,
			qrcv.sessionID,
			&azservicebus.SessionReceiverOptions{
				ReceiveMode: azservicebus.ReceiveModeReceiveAndDelete,
			})
		return queueSessionRcv, err

	}
	logCache.Debugf("Using receiveMode PeekLock on queue %s", qrcv.queueName)
	queueSessionRcv, err := client.AcceptSessionForQueue(
		qrcv.ctx,
		qrcv.queueName,
		qrcv.sessionID,
		&azservicebus.SessionReceiverOptions{
			ReceiveMode: azservicebus.ReceiveModePeekLock,
		})
	return queueSessionRcv, err
}

func getDeadLetter(client *azservicebus.Client, qrcv *QueueReceiver) (*azservicebus.Receiver, error) {
	if qrcv.timeOut > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(qrcv.timeOut))
		defer cancel()
		qrcv.ctx = ctx
		qrcv.listenctxCancelFunc = cancel
	} else {
		qrcv.ctx = context.Background()

	}

	if qrcv.receiveMode == "ReceiveAndDelete" {
		logCache.Debugf("Using receiveMode ReceiveAndDelete on deadLetter queue %s", qrcv.queueName)
		qe, err := client.NewReceiverForQueue(
			qrcv.queueName,
			&azservicebus.ReceiverOptions{
				ReceiveMode: azservicebus.ReceiveModeReceiveAndDelete,
				SubQueue:    azservicebus.SubQueueDeadLetter,
			})
		return qe, err
	}
	logCache.Debugf("Using receiveMode PeekLock on deadLetter queue %s", qrcv.queueName)
	qe, err := client.NewReceiverForQueue(
		qrcv.queueName,
		&azservicebus.ReceiverOptions{
			ReceiveMode: azservicebus.ReceiveModePeekLock,
			SubQueue:    azservicebus.SubQueueDeadLetter,
		})
	return qe, err
}

func (qrcvr *QueueReceiver) listen() {
	if len(qrcvr.sessionID) > 0 {
		logCache.Infof("QueueReceiver will now poll on Queue [%s] which has session support", qrcvr.queueName)
		for {
			message, err := qrcvr.queueSessionRcv.ReceiveMessages(qrcvr.ctx, 1, nil)
			if err != nil {
				logCache.Error(err.Error())
				return
			}
			if len(message) < 1 { //context cancelled for receiver
				return
			}
			//handler call for flow execution
			resp := processMessage(message[0], qrcvr)
			if qrcvr.receiveMode == "ModePeekLock" {
				if resp.moveToDL {
					err = qrcvr.queueSessionRcv.DeadLetterMessage(qrcvr.ctx, message[0], &azservicebus.DeadLetterOptions{
						Reason:           &resp.deadLetter.deadLetterReason,
						ErrorDescription: &resp.deadLetter.deadLetterDescription,
					})
				} else {
					if resp.ack {
						err = qrcvr.queueSessionRcv.CompleteMessage(qrcvr.ctx, message[0], nil)
					} else {
						err = qrcvr.queueSessionRcv.AbandonMessage(qrcvr.ctx, message[0], nil)
					}
				}
				if err != nil {
					var sbErr *azservicebus.Error
					if errors.As(err, &sbErr) {
						if sbErr.Code == azservicebus.CodeLockLost {
							logCache.Error("Message lock expired, extend the message lock duration as per flow time duration or Server reset connection")
						}
					} else {
						logCache.Error(err.Error())
						return
					}
				}
			}
		}
	} else {
		logCache.Infof("QueueReceiver will now poll on Queue [%s] which does not have session support", qrcvr.queueName)
		for {
			message, err := qrcvr.queue.ReceiveMessages(qrcvr.ctx, 1, nil)
			if err != nil {
				logCache.Error(err.Error())
				return
			}
			if len(message) < 1 { //context cancelled for receiver
				return
			}
			//handler call for flow execution
			resp := processMessage(message[0], qrcvr)
			if qrcvr.receiveMode == "ModePeekLock" {
				if resp.moveToDL {
					err = qrcvr.queue.DeadLetterMessage(qrcvr.ctx, message[0], &azservicebus.DeadLetterOptions{
						Reason:           &resp.deadLetter.deadLetterReason,
						ErrorDescription: &resp.deadLetter.deadLetterDescription,
					})
				} else {
					if resp.ack {
						err = qrcvr.queue.CompleteMessage(qrcvr.ctx, message[0], nil)
					} else {
						err = qrcvr.queue.AbandonMessage(qrcvr.ctx, message[0], nil)
					}
				}
				if err != nil {
					var sbErr *azservicebus.Error
					if errors.As(err, &sbErr) {
						if sbErr.Code == azservicebus.CodeLockLost {
							logCache.Error("Message lock expired, extend the message lock duration as per flow time duration or Server reset connection")
						}
					} else {
						logCache.Error(err.Error())
						return
					}
				}
			}
		}
	}
}

func (qrcvr *QueueReceiver) listenDeadletter() {
	logCache.Infof("QueueReceiver will now poll on DeadLetter Queue [%s]", qrcvr.queueName)
	for {
		message, err := qrcvr.deadLetter.ReceiveMessages(qrcvr.ctx, 1, nil)
		if err != nil {
			logCache.Error(err.Error())
			return
		}

		if len(message) < 1 { //context cancelled for receiver
			return
		}

		//handler call for flow execution
		resp := processMessage(message[0], qrcvr)
		if qrcvr.receiveMode == "ModePeekLock" {
			if resp.moveToDL {
				err = qrcvr.deadLetter.DeadLetterMessage(qrcvr.ctx, message[0], &azservicebus.DeadLetterOptions{
					Reason:           &resp.deadLetter.deadLetterReason,
					ErrorDescription: &resp.deadLetter.deadLetterDescription,
				})
			} else {
				if resp.ack {
					err = qrcvr.deadLetter.CompleteMessage(qrcvr.ctx, message[0], nil)
				} else {
					err = qrcvr.deadLetter.AbandonMessage(qrcvr.ctx, message[0], nil)
				}
			}
			if err != nil {
				var sbErr *azservicebus.Error
				if errors.As(err, &sbErr) {
					if sbErr.Code == azservicebus.CodeLockLost {
						logCache.Error("Message lock expired, extend the message lock duration as per flow time duration or Server reset connection")
					}
				} else {
					logCache.Error(err.Error())
					return
				}
			}
		}
	}
}

type processResp struct {
	ack        bool
	moveToDL   bool
	deadLetter deadLetterResp
}

type deadLetterResp struct {
	deadLetterReason      string
	deadLetterDescription string
}

func processMessage(msg *azservicebus.ReceivedMessage, qrcvr *QueueReceiver) (rsp processResp) {
	var outputRoot = map[string]interface{}{}
	var brokerProperties = map[string]interface{}{}
	var deadLetter = map[string]string{}
	outputData := make(map[string]interface{})
	output := &Output{}
	deserVal := qrcvr.valueType
	if deserVal == "String" {
		if msg.Body != nil {
			text := string(msg.Body)
			outputRoot["messageString"] = string(text)
		} else {
			logCache.Warnf("Azure Service bus Queue message content is nil")
		}

		brokerProperties["ContentType"] = copyValue(msg.ContentType)
		brokerProperties["CorrelationId"] = copyValue(msg.CorrelationID)
		brokerProperties["DeliveryCount"] = msg.DeliveryCount
		brokerProperties["Label"] = copyValue(msg.Subject)
		//brokerPropertiesResp["MessageId"] = msg.ID
		brokerProperties["PartitionKey"] = copyValue(msg.PartitionKey)
		brokerProperties["ReplyTo"] = copyValue(msg.ReplyTo)
		brokerProperties["SessionId"] = copyValue(msg.SessionID)
		ttl := msg.TimeToLive.String()
		ttlint, _ := strconv.Atoi(ttl)
		brokerProperties["TimeToLive"] = ttlint
		brokerProperties["To"] = msg.To

		outputRoot["brokerProperties"] = brokerProperties
		//pickng up tracing and application properties
		if trace.Enabled() && len(msg.ApplicationProperties) > 0 {
			tc, _ := trace.GetTracer().Extract(trace.TextMap, msg.ApplicationProperties)
			if tc != nil {
				qrcvr.ctx = trace.AppendTracingContext(qrcvr.ctx, tc)
			}
		}
		outputRoot["customProperties"] = processCustomProperties(qrcvr.handler.Schemas().Output, msg.ApplicationProperties)
		if qrcvr.isDeadLetter {
			if msg.DeadLetterReason != nil {
				deadLetter["reason"] = *msg.DeadLetterReason
			} else {
				deadLetter["reason"] = ""
			}
			if msg.DeadLetterErrorDescription != nil {
				deadLetter["description"] = *msg.DeadLetterErrorDescription
			} else {
				deadLetter["description"] = ""
			}
			outputRoot["deadLetter"] = deadLetter
		}

		outputData["output"] = outputRoot
		//logCache.Info("output at runtime %#v", output.Output)

		output.Output = outputData
	} else if deserVal == "JSON" {
		// future use
	}

	if msg.MessageID != "" {
		qrcvr.ctx = trigger.NewContextWithEventId(qrcvr.ctx, msg.MessageID)
	}
	attrs, err := qrcvr.handler.Handle(qrcvr.ctx, outputData)
	if err != nil {
		logCache.Errorf("Failed to process record from Queue [%s], due to error - %s", qrcvr.queueName, err.Error())
		rsp.ack = false
		return
	} else if attrs["messageAck"] != nil && attrs["messageAck"] == false {
		rsp.ack = false
		return
	} else if attrs["deadletter"] != nil {
		dl, ok := attrs["deadletter"].(map[string]string)
		if !ok {
			logCache.Warnf("Failed to decode dead letter details,not moving the message to dead letter queue but still acknowledging the message")
		} else {
			rsp.moveToDL = true
			if reason, ok := dl["deadLetterReason"]; ok {
				rsp.deadLetter.deadLetterReason = reason
			}
			if description, ok := dl["deadLetterDescription"]; ok {
				rsp.deadLetter.deadLetterDescription = description
			}
		}
	}
	rsp.ack = true // auto acknowledge messages
	logCache.Infof("Record from Queue [%s] is successfully processed", qrcvr.queueName)
	return
}

// process/typecast custom properties of a msg
func processCustomProperties(outputSchema map[string]interface{}, customProperties map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	customPropertiesSchema := outputSchema["customProperties"]
	if customPropertiesSchema != nil {
		valueStr := customPropertiesSchema.(map[string]interface{})["value"].(string)
		var value map[string]interface{}
		err := json.Unmarshal([]byte(valueStr), &value)
		if err != nil {
			logCache.Errorf("Unable to unmarshal custom properties schema  %s", err.Error())
			return res
		}
		propertiesSchema := value["properties"].(map[string]interface{})

		for k, v := range customProperties {
			if propertiesSchema[k] != nil {
				switch propertiesSchema[k].(map[string]interface{})["type"] {
				case "number":
					data, err := strconv.ParseFloat(fmt.Sprint(v), 64)
					if err != nil {
						logCache.Warnf(k+" expected type is number but received value %s", fmt.Sprint(v))
						res[k] = fmt.Sprint(v)
						continue
					}
					res[k] = data
				case "integer":
					data, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
					if err != nil {
						logCache.Warnf(k+" expected type is number but received value %s", fmt.Sprint(v))
						res[k] = fmt.Sprint(v)
						continue
					}
					res[k] = data
				case "boolean":
					data, err := strconv.ParseBool(fmt.Sprint(v))
					if err != nil {
						logCache.Warnf(k+" expected type is number but received value %s", fmt.Sprint(v))
						res[k] = fmt.Sprint(v)
						continue
					}
					res[k] = data
				default:
					res[k] = fmt.Sprint(v)
				}
			} else {
				res[k] = fmt.Sprint(v)
			}
		}

		//logCache.Info("custom obj %#v", res)
	}
	return res
}

// Stop implements trigger.Trigger.Stop
func (t *SBQueueReceiverTrigger) Stop() error {
	//logCache.Infof("Stopping Trigger - %s", t.config.Name)
	for _, qrcvr := range t.queueReceivers {
		logCache.Debugf("About to close ListenerHandler for Queue [%s]", qrcvr.queueName)
		select {
		case <-time.After(2 * time.Second):
			if qrcvr.isDeadLetter {
				qrcvr.deadLetter.Close(qrcvr.ctx)
			} else if len(qrcvr.sessionID) > 0 {
				qrcvr.queueSessionRcv.Close(qrcvr.ctx)
			} else {
				qrcvr.queue.Close(qrcvr.ctx)
			}
			if qrcvr.listenctxCancelFunc != nil {
				qrcvr.listenctxCancelFunc()
			}
		}
	}

	//logCache.Infof("Trigger - %s  stopped", t.config.Name)
	return nil
}

func copyValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}
