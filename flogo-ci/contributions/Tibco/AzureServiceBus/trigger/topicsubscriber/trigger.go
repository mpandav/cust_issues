package topicsubscriber

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

// OSS UPGRADE---
var logCache = log.ChildLogger(log.RootLogger(), "AzureServiceBus-trigger-topicsubscriber")

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&SBTopicSubscriberTrigger{}, &MyTriggerFactory{})
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
	return &SBTopicSubscriberTrigger{metadata: t.metadata, config: config, ascm: ascm}, nil
}

// Metadata implements trigger.Factory.Metadata
func (*MyTriggerFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

// SBTopicSubscriberTrigger is a stub for your Trigger implementation
type SBTopicSubscriberTrigger struct {
	metadata         *trigger.Metadata
	config           *trigger.Config
	topicSubscribers []*TopicSubscriber
	ascm             *connection.AzureServiceBusSharedConfigManager
}

// TopicSubscriber is structure of a single TopicSubscriber
type TopicSubscriber struct {
	handler             trigger.Handler
	topic               *azservicebus.Receiver
	topicSessionRcv     *azservicebus.SessionReceiver
	deadLetter          *azservicebus.Receiver
	ctx                 context.Context
	listenctxCancelFunc context.CancelFunc
	topicName           string
	subscriptionName    string
	sessionID           string
	connString          string
	valueType           string
	receiveMode         string
	isDeadLetter        bool
	timeOut             int
	retrycount          int
	retryInterval       int
}

// Initialize SBTopicSubscriberTrigger
func (t *SBTopicSubscriberTrigger) Initialize(ctx trigger.InitContext) error {

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
		trcvr := &TopicSubscriber{}
		trcvr.handler = handler
		trcvr.connString = connStr
		trcvr.topicName = handlerSettings.Topic
		trcvr.receiveMode = handlerSettings.ReceiveMode
		trcvr.subscriptionName = handlerSettings.SubscriptionName
		trcvr.valueType = handlerSettings.ValueType
		trcvr.sessionID = handlerSettings.SessionId
		trcvr.isDeadLetter = handlerSettings.DeadLetter
		trcvr.timeOut = handlerSettings.Timeout
		trcvr.retrycount = handlerSettings.Count
		trcvr.retryInterval = handlerSettings.Interval
		t.topicSubscribers = append(t.topicSubscribers, trcvr)
	}

	return nil
}

// Start implements trigger.Trigger.Start
func (t *SBTopicSubscriberTrigger) Start() (err error) {

	for _, trcvr := range t.topicSubscribers {
		client := t.ascm.ServiceBusClient
		if trcvr.retrycount > 0 {
			//create new client only if retry counts are different on trigger
			retryOptions := azservicebus.RetryOptions{
				MaxRetries:    int32(trcvr.retrycount),
				RetryDelay:    time.Duration(trcvr.retryInterval) * time.Millisecond,
				MaxRetryDelay: time.Duration(trcvr.retryInterval) * time.Millisecond,
			}
			client, err = azservicebus.NewClientFromConnectionString(trcvr.connString,
				&azservicebus.ClientOptions{
					RetryOptions: retryOptions,
				})
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
		}

		if trcvr.isDeadLetter {
			topicSub, err := getDeadLetter(client, trcvr)
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
			trcvr.deadLetter = topicSub
			go trcvr.listenDeadletter()
		} else if len(trcvr.sessionID) > 0 {
			topicSessionRcv, err := getTopicWithSession(client, trcvr)
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
			trcvr.topicSessionRcv = topicSessionRcv
			go trcvr.listen()
		} else {
			topicSub, err := getTopic(client, trcvr)
			if err != nil {
				logCache.Error(err.Error())
				return err
			}
			trcvr.topic = topicSub
			go trcvr.listen()
		}
	}
	//log.Infof("Trigger - %s  started", t.config.Name)
	return nil
}

func getTopic(client *azservicebus.Client, trcvr *TopicSubscriber) (*azservicebus.Receiver, error) {
	if trcvr.timeOut > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(trcvr.timeOut))
		trcvr.ctx = ctx
		trcvr.listenctxCancelFunc = cancel
	} else {
		trcvr.ctx = context.Background()
	}

	if trcvr.receiveMode == "ReceiveAndDelete" {
		logCache.Debugf("Using receiveMode ReceiveAndDelete on topic %s with subscription %s", trcvr.topicName, trcvr.subscriptionName)
		topicSub, err := client.NewReceiverForSubscription(
			trcvr.topicName,
			trcvr.subscriptionName,
			&azservicebus.ReceiverOptions{
				ReceiveMode: azservicebus.ReceiveModeReceiveAndDelete,
			})
		return topicSub, err
	}
	logCache.Debugf("Using receiveMode PeekLock on topic %s with subscription %s", trcvr.topicName, trcvr.subscriptionName)
	topicSub, err := client.NewReceiverForSubscription(
		trcvr.topicName,
		trcvr.subscriptionName,
		&azservicebus.ReceiverOptions{
			ReceiveMode: azservicebus.ReceiveModePeekLock,
		})
	return topicSub, err
}

func getTopicWithSession(client *azservicebus.Client, trcvr *TopicSubscriber) (*azservicebus.SessionReceiver, error) {
	if trcvr.timeOut > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(trcvr.timeOut))
		trcvr.ctx = ctx
		trcvr.listenctxCancelFunc = cancel
	} else {
		trcvr.ctx = context.Background()

	}

	if trcvr.receiveMode == "ReceiveAndDelete" {
		logCache.Debugf("Using receiveMode ReceiveAndDelete on topic %s with subscription %s", trcvr.topicName, trcvr.subscriptionName)

		topicSessionRcv, err := client.AcceptSessionForSubscription(
			trcvr.ctx,
			trcvr.topicName,
			trcvr.subscriptionName,
			trcvr.sessionID,
			&azservicebus.SessionReceiverOptions{
				ReceiveMode: azservicebus.ReceiveModeReceiveAndDelete,
			})
		return topicSessionRcv, err

	}
	logCache.Debugf("Using receiveMode PeekLock on topic %s with subscription %s", trcvr.topicName, trcvr.subscriptionName)
	topicSessionRcv, err := client.AcceptSessionForSubscription(
		trcvr.ctx,
		trcvr.topicName,
		trcvr.subscriptionName,
		trcvr.sessionID,
		&azservicebus.SessionReceiverOptions{
			ReceiveMode: azservicebus.ReceiveModePeekLock,
		})
	return topicSessionRcv, err
}

func getDeadLetter(client *azservicebus.Client, trcvr *TopicSubscriber) (*azservicebus.Receiver, error) {
	if trcvr.timeOut > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(trcvr.timeOut))
		trcvr.ctx = ctx
		trcvr.listenctxCancelFunc = cancel
	} else {
		trcvr.ctx = context.Background()

	}

	if trcvr.receiveMode == "ReceiveAndDelete" {
		logCache.Debugf("Using receiveMode ReceiveAndDelete on topic %s with deadLetter subscription %s", trcvr.topicName, trcvr.subscriptionName)
		topicSub, err := client.NewReceiverForSubscription(
			trcvr.topicName,
			trcvr.subscriptionName,
			&azservicebus.ReceiverOptions{
				ReceiveMode: azservicebus.ReceiveModeReceiveAndDelete,
				SubQueue:    azservicebus.SubQueueDeadLetter,
			})
		return topicSub, err
	}
	logCache.Debugf("Using receiveMode PeekLock on topic %s with deadLetter subscription %s", trcvr.topicName, trcvr.subscriptionName)
	topicSub, err := client.NewReceiverForSubscription(
		trcvr.topicName,
		trcvr.subscriptionName,
		&azservicebus.ReceiverOptions{
			ReceiveMode: azservicebus.ReceiveModePeekLock,
			SubQueue:    azservicebus.SubQueueDeadLetter,
		})
	return topicSub, err
}

func (trcvr *TopicSubscriber) listen() {
	if len(trcvr.sessionID) > 0 {
		logCache.Infof("TopicSubscriber will now poll on Topic [%s] which has session support", trcvr.topicName)
		for {
			message, err := trcvr.topicSessionRcv.ReceiveMessages(trcvr.ctx, 1, nil)
			if err != nil {
				logCache.Error(err.Error())
				return
			}
			if len(message) < 1 { //context cancelled for receiver
				return
			}
			//handler call for flow execution
			resp := processMessage(message[0], trcvr)
			if trcvr.receiveMode == "ModePeekLock" {
				if resp.moveToDL {
					err = trcvr.topicSessionRcv.DeadLetterMessage(trcvr.ctx, message[0], &azservicebus.DeadLetterOptions{
						Reason:           &resp.deadLetter.deadLetterReason,
						ErrorDescription: &resp.deadLetter.deadLetterDescription,
					})
				} else {
					if resp.ack {
						err = trcvr.topicSessionRcv.CompleteMessage(trcvr.ctx, message[0], nil)
					} else {
						err = trcvr.topicSessionRcv.AbandonMessage(trcvr.ctx, message[0], nil)
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
		logCache.Infof("TopicSubscriber will now poll on Topic [%s] which does not have session support", trcvr.topicName)
		for {
			message, err := trcvr.topic.ReceiveMessages(trcvr.ctx, 1, nil)
			if err != nil {
				logCache.Error(err.Error())
				return
			}
			if len(message) < 1 { //context cancelled for receiver
				return
			}
			//handler call for flow execution
			resp := processMessage(message[0], trcvr)
			if trcvr.receiveMode == "ModePeekLock" {
				if resp.moveToDL {
					err = trcvr.topic.DeadLetterMessage(trcvr.ctx, message[0], &azservicebus.DeadLetterOptions{
						Reason:           &resp.deadLetter.deadLetterReason,
						ErrorDescription: &resp.deadLetter.deadLetterDescription,
					})
				} else {
					if resp.ack {
						err = trcvr.topic.CompleteMessage(trcvr.ctx, message[0], nil)
					} else {
						err = trcvr.topic.AbandonMessage(trcvr.ctx, message[0], nil)
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

func (trcvr *TopicSubscriber) listenDeadletter() {
	logCache.Infof("TopicSubscriber will now poll on Deadletter Topic [%s]", trcvr.topicName)
	for {
		message, err := trcvr.deadLetter.ReceiveMessages(trcvr.ctx, 1, nil)
		if err != nil {
			logCache.Error(err.Error())
			return
		}
		if len(message) < 1 { //context cancelled for receiver
			return
		}
		//handler call for flow execution
		resp := processMessage(message[0], trcvr)
		if trcvr.receiveMode == "ModePeekLock" {
			if resp.moveToDL {
				err = trcvr.deadLetter.DeadLetterMessage(trcvr.ctx, message[0], &azservicebus.DeadLetterOptions{
					Reason:           &resp.deadLetter.deadLetterReason,
					ErrorDescription: &resp.deadLetter.deadLetterDescription,
				})
			} else {
				if resp.ack {
					err = trcvr.deadLetter.CompleteMessage(trcvr.ctx, message[0], nil)
				} else {
					err = trcvr.deadLetter.AbandonMessage(trcvr.ctx, message[0], nil)
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

func processMessage(msg *azservicebus.ReceivedMessage, trcvr *TopicSubscriber) (rsp processResp) {
	var outputRoot = map[string]interface{}{}
	var brokerProperties = map[string]interface{}{}
	var deadLetter = map[string]string{}
	outputData := make(map[string]interface{})
	output := &Output{}

	deserVal := trcvr.valueType
	if deserVal == "String" {
		if msg.Body != nil {
			text := string(msg.Body)
			outputRoot["messageString"] = string(text)
		} else {
			logCache.Warnf("Azure Service bus Topic message content is nil")
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
				trcvr.ctx = trace.AppendTracingContext(trcvr.ctx, tc)
			}
		}
		outputRoot["customProperties"] = processCustomProperties(trcvr.handler.Schemas().Output, msg.ApplicationProperties)
		if trcvr.isDeadLetter {
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
		trcvr.ctx = trigger.NewContextWithEventId(trcvr.ctx, msg.MessageID)
	}
	attrs, err := trcvr.handler.Handle(trcvr.ctx, outputData)
	if err != nil {
		logCache.Errorf("Failed to process record from Topic [%s], due to error - %s", trcvr.topicName, err.Error())
		rsp.ack = false
		return
	} else if attrs["messageAck"] != nil && attrs["messageAck"] == false {
		rsp.ack = false
		return
	} else if attrs["deadletter"] != nil {
		dl, ok := attrs["deadletter"].(map[string]string)
		if !ok {
			logCache.Warnf("Failed to decode dead letter details, but still acknowledging the message")
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
	rsp.ack = true
	logCache.Infof("Record from Topic [%s] is successfully processed with subscription [%s]", trcvr.topicName, trcvr.subscriptionName)
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
func (t *SBTopicSubscriberTrigger) Stop() error {
	//logCache.Infof("Stopping Trigger - %s", t.config.Name)
	for _, trcvr := range t.topicSubscribers {
		logCache.Debugf("About to close ListenerHandler for Topic [%s]", trcvr.topicName)
		select {
		case <-time.After(2 * time.Second):
			if trcvr.isDeadLetter {
				trcvr.deadLetter.Close(trcvr.ctx)
			} else if len(trcvr.sessionID) > 0 {
				trcvr.topicSessionRcv.Close(trcvr.ctx)
			} else {
				trcvr.topic.Close(trcvr.ctx)
			}
			if trcvr.listenctxCancelFunc != nil {
				trcvr.listenctxCancelFunc()
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
