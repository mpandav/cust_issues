package tcmtrigger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/TIBCOSoftware/eftl"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"
	"github.com/tibco/flogo-messaging/src/app/Messaging/connector/tcm"
	//"github.com/tibco/flogo-messaging/Messaging/connector/tcm"
)

const (
	ivMessage           = "message"
	AckModeExplicit     = "Explicit"
	AckModeAuto         = "Auto"
	DurableTypeStandard = "Standard"
	DurableTypeShared   = "Shared"
	ProcessingModeAsync = "Async"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &Output{})
var EventBasedFlowControl bool

func init() {
	_ = trigger.Register(&MyTrigger{}, &MyTriggerFactory{})
}

// MyTriggerFactory My Trigger factory
type MyTriggerFactory struct {
	metadata *trigger.Metadata
}

//NewFactory create a new Trigger factory

// New Creates a new trigger instance for a given id
func (t *MyTriggerFactory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &MyTrigger{metadata: t.metadata, id: config.Id}, nil
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
	id        string
	consumers []*consumer
	logger    log.Logger
}

type consumer struct {
	durableSub, subscribed, asyncMode                    bool
	maxMsgCount, currentMsgCount                         int
	connectionManager                                    connection.Manager
	tcmConnection                                        tcm.Connection
	handler                                              trigger.Handler
	eftlSub                                              *eftl.Subscription
	ackMode                                              eftl.AcknowledgeMode
	durableName, durableType, destination, matcherString string
	stopChan                                             chan bool
	wg                                                   sync.WaitGroup
}

func (consumer *consumer) GetTCMConnection() (tcm.Connection, error) {
	return consumer.connectionManager.(*tcm.TCMSharedConfigManager).GetSubscribeConnection()
}

func (t *MyTrigger) Initialize(ctx trigger.InitContext) error {
	t.logger = ctx.Logger()
	EventBasedFlowControl = app.EnableFlowControl()

	for _, handler := range ctx.GetHandlers() {
		tConsumer := &consumer{}

		conn, err := tcm.GetSharedConfiguration(handler.Settings()["tcmConnection"])
		if err != nil {
			return err
		}

		tConsumer.connectionManager = conn
		s := &Settings{}
		err = metadata.MapToStruct(handler.Settings(), s, true)
		if err != nil {
			return err
		}

		matcher := make(map[string]interface{})
		if s.Matcher != "" {
			//Add attribute names
			if strings.HasPrefix(s.Matcher, "{") {
				// Matcher configured through app property in the format {"m1": "v1", "m2": v2}
				matcher, err = coerce.ToObject(s.Matcher)
				if err != nil {
					return fmt.Errorf(fmt.Sprintf("Invalid content matcher [%s]. It must be a valid JSON object {\"k1\":\"v1\"}.", s.Matcher))
				}
			} else if strings.HasPrefix(s.Matcher, "[") {
				attrsNames, _ := coerce.ToArray(s.Matcher)
				for _, v := range attrsNames {
					attrInfo := v.(map[string]interface{})
					attrType := attrInfo["Type"].(string)
					if attrType == "String" {
						matcher[attrInfo["Name"].(string)] = attrInfo["Value"].(string)
					} else if attrType == "Integer" {
						val, _ := coerce.ToInt(attrInfo["Value"])
						matcher[attrInfo["Name"].(string)] = val
					} else if attrType == "Boolean" {
						val, _ := coerce.ToBool(attrInfo["Value"])
						matcher[attrInfo["Name"].(string)] = val
					}
				}
			} else {
				return fmt.Errorf(fmt.Sprintf("Invalid content matcher [%s]. It must be a valid JSON object {\"k1\":\"v1\"}.", s.Matcher))
			}
		}

		dest := s.Destination
		if len(dest) > 0 {
			matcher["_dest"] = dest
		}

		matcherString, err := json.Marshal(matcher)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("[Handler:%s] -  Failed to convert matcher to string due to error - {%s}", handler.Name(), err.Error()))
		}

		tConsumer.matcherString = string(matcherString)
		tConsumer.durableSub = s.DurableSub
		if s.DurableSub {
			tConsumer.durableName = s.DurableName
			tConsumer.durableType = s.DurableType
			ctx.Logger().Debugf("[Handler:%s] - Durable type set to '%s'", handler.Name(), tConsumer.durableType)

			if s.AckMode == "" || s.AckMode == AckModeAuto {
				tConsumer.ackMode = eftl.AcknowledgeModeAuto
				ctx.Logger().Debugf("[Handler:%s] -  Message acknowledge mode set to '%s'", handler.Name(), AckModeAuto)
			} else if s.AckMode == AckModeExplicit {
				tConsumer.ackMode = eftl.AcknowledgeModeClient
				ctx.Logger().Debugf("[Handler:%s] -  Message acknowledge mode set to '%s'", handler.Name(), AckModeExplicit)
			}

			if s.DurableType == DurableTypeStandard {
				clientId := engine.GetAppName() + "-" + t.id + "-" + handler.Name()
				// Check client ID
				tcmConn, err := tConsumer.GetTCMConnection()
				if err != nil {
					return err
				}
				if clientId != tcmConn.GetEFTLConnection().Options.ClientID {
					// Set clientId to app name
					tcmConn.GetEFTLConnection().Options.ClientID = clientId
					err = tcmConn.Reconnect(handler.Name())
					if err != nil {
						return err
					}
					ctx.Logger().Debugf("[Handler:%s] - Connection is reset with a new ClientID '%s' for standard durable subscriber.", handler.Name(), tcmConn.GetEFTLConnection().Options.ClientID)
				}
			}
		}

		tConsumer.handler = handler
		tConsumer.asyncMode = s.ProcessingMode == ProcessingModeAsync
		tConsumer.maxMsgCount = getMaxMessageCount()
		tConsumer.wg = sync.WaitGroup{}
		t.consumers = append(t.consumers, tConsumer)

		ctx.Logger().Debugf("[Handler:%s] -  Matcher:%s", handler.Name(), matcherString)
	}

	return nil

}

func getMaxMessageCount() int {
	if engine.GetRunnerType() == engine.ValueRunnerTypePooled {
		return engine.GetRunnerWorkers()
	}
	// For DIRECT mode
	return 200
}

func (t *MyTrigger) Start() error {

	for _, consumer := range t.consumers {
		consumer.stopChan = make(chan bool)
		tcmConn, err := consumer.GetTCMConnection()
		if err != nil {
			return fmt.Errorf("Error get or create connect to TCM: %s", err.Error())
		}
		consumer.tcmConnection = tcmConn
		// channel for receiving subscription response
		subChan := make(chan *eftl.Subscription, 1)
		// channel for receiving published messages
		msgChan := make(chan eftl.Message)

		err = subscribe(consumer, msgChan, subChan)
		if err != nil {
			consumer.handler.Logger().Errorf("Subscribe failed due to error - {%s}, Now trying to reconnect", err.Error())
			err = tcmConn.Reconnect(t.id)
			if err != nil {
				return fmt.Errorf("[Handler:%s] -  Failed to reconnect due to error - {%s}", consumer.handler.Name(), err.Error())
			}

			err = subscribe(consumer, msgChan, subChan)
			if err != nil {
				return fmt.Errorf("[Handler:%s] -  Subscription failed due to error - {%s}. Now trying to reconnect", consumer.handler.Name(), err.Error())
			}
			go consumer.handleSubscription(subChan, msgChan)
		} else {
			go consumer.handleSubscription(subChan, msgChan)
		}

	}
	t.logger.Infof("Triggered started")
	return nil
}

func (t *MyTrigger) Stop() error {

	for _, consumer := range t.consumers {
		// TODO We should unsubscribe on graceful shutdown. Unfortunately this will break backward compatibility. Hence, reverted.
		// TODO In future, we should expose UI configuration to support unsubscribe.
		/*err := consumer.tcmConnection.GetEFTLConnection().Unsubscribe(consumer.eftlSub)
		if err != nil {
			t.logger.Errorf("Failed to unsubscribe. Error:", err.Error())
		}*/
		consumer.subscribed = false
		consumer.eftlSub = nil
		close(consumer.stopChan)
	}

	t.logger.Infof("Triggered stopped")
	return nil
}

func (t *MyTrigger) Pause() error {
	for _, consumer := range t.consumers {
		err := consumer.tcmConnection.GetEFTLConnection().CloseSubscription(consumer.eftlSub)
		if err != nil {
			t.logger.Errorf("Failed to close subscription. Error:", err.Error())
		}
		consumer.subscribed = false
		consumer.eftlSub = nil
		close(consumer.stopChan)
	}
	t.logger.Infof("Trigger is paused")
	return nil
}

func (t *MyTrigger) Resume() error {
	t.logger.Infof("Trigger is resumed")
	return t.Start()
}

func subscribe(consumer *consumer, msgChan chan eftl.Message, subChan chan *eftl.Subscription) (err error) {
	if consumer.durableSub {
		durableType := consumer.durableType
		durableName := consumer.durableName
		if durableName == "" {
			return fmt.Errorf("[Handler:%s] -  Durable name must be set for durable subscription", consumer.handler.Name())
		}
		subOpts := eftl.SubscriptionOptions{}
		if durableType == DurableTypeShared {
			subOpts.DurableType = durableType
		}
		subOpts.AcknowledgeMode = consumer.ackMode
		err = consumer.tcmConnection.GetEFTLConnection().SubscribeWithOptionsAsync(consumer.matcherString, durableName, subOpts, msgChan, subChan)
	} else {
		err = consumer.tcmConnection.GetEFTLConnection().SubscribeAsync(consumer.matcherString, "", msgChan, subChan)
	}
	return err
}

func (consumer *consumer) handleSubscription(subChan chan *eftl.Subscription, msgChan chan eftl.Message) {
	for {
		select {
		case sub := <-subChan:
			if sub.Error != nil {
				//Close old msg channel
				close(msgChan)
				subscriptionName := consumer.handler.Name()
				if len(consumer.durableName) > 0 {
					subscriptionName = consumer.durableName
				}
				consumer.handler.Logger().Errorf("[Handler:%s] -  Subscription [%s] failed due to error - {%s}, trying to re-subscribe", consumer.handler.Name(), subscriptionName, sub.Error.Error())
				reConSubChan := make(chan *eftl.Subscription, 1)
				reConnectMsgChan := make(chan eftl.Message)
				err := subscribe(consumer, reConnectMsgChan, reConSubChan)
				if err != nil {
					panic(fmt.Errorf("[Handler:%s] -  Subscription [%s] failed to resubscribe with TCM server due to error - {%s}", consumer.handler.Name(), subscriptionName, err.Error()))
				}
				consumer.handler.Logger().Infof("[Handler:%s] -  Subscription [%s] successfully resubscribed.", consumer.handler.Name(), subscriptionName)
				consumer.subscribed = false
				go consumer.handleSubscription(reConSubChan, reConnectMsgChan)
				return
			}
			//Start Goroutine to receive messages from the message channel
			consumer.handler.Logger().Debugf("[Handler:%s] -  IsSubscribed: %v", consumer.handler.Name(), consumer.subscribed)
			if !consumer.subscribed {
				consumer.eftlSub = sub
				go consumer.handleMessage(msgChan)
				consumer.handler.Logger().Debugf("[Handler:%s] -  Subscribed successfully", consumer.handler.Name())
			}
		case <-consumer.stopChan:
			consumer.handler.Logger().Debug("Stop channel message received by handleSubscription")
			return
		}
	}
}

func (consumer *consumer) handleMessage(msgChan chan eftl.Message) {
	consumer.handler.Logger().Debugf("[Handler:%s] -  Started message handler", consumer.handler.Name())
	defer consumer.handler.Logger().Debugf("[Handler:%s] -  Stopping message handler", consumer.handler.Name())
	consumer.subscribed = true
	for {
		select {
		case m, ok := <-msgChan:
			if ok {
				//If flow control is enabled
				// use engine's flow limit control
				if EventBasedFlowControl {
					go consumer.callHandle(m)
				} else if consumer.asyncMode {
					consumer.wg.Add(1)
					consumer.currentMsgCount++
					go consumer.callHandle(m)
					if consumer.currentMsgCount >= consumer.maxMsgCount {
						consumer.handler.Logger().Infof("Total messages received are equal or more than maximum threshold [%d]. Blocking message handler.", consumer.maxMsgCount)
						consumer.wg.Wait()
						// reset count
						consumer.currentMsgCount = 0
						consumer.handler.Logger().Info("All received messages are processed. Unblocking message handler.")
					}
				} else {
					consumer.callHandle(m)
				}
			} else {
				//closed
				return
			}
		case <-consumer.stopChan:
			consumer.handler.Logger().Debug("Stop channel message received by handleMessage")
			return
		}
	}
}

func (consumer *consumer) callHandle(m eftl.Message) {
	defer func() {
		if !EventBasedFlowControl && consumer.asyncMode {
			consumer.wg.Done()
			consumer.currentMsgCount--
		}
	}()
	consumer.handler.Logger().Infof("[Handler:%s] -  Message(ID:%d) received", consumer.handler.Name(), m.StoreMessageId())
	if m.DeliveryCount() > 1 {
		consumer.handler.Logger().Infof("Message(ID:%d) is redelivered. DeliveryAttempt:%d", m.StoreMessageId(), m.DeliveryCount())
	}

	outputData := &Output{}
	outputData.Metadata = MessageMetadata{Id: m.StoreMessageId(), DeliveryCount: m.DeliveryCount()}
	ctx := context.Background()
	if trace.Enabled() {
		msgProp := extractTraceContext(m)
		tc, _ := trace.GetTracer().Extract(trace.TextMap, msgProp)
		if tc != nil {
			ctx = trace.AppendTracingContext(ctx, tc)
		}
	}
	outputData.Message = messageToMap(m)

	msgId, _ := coerce.ToString(m.StoreMessageId())
	if msgId != "" {
		ctx = trigger.NewContextWithEventId(ctx, msgId)
	}

	data, err := (*consumer).handler.Handle(ctx, outputData)
	if err != nil {
		consumer.handler.Logger().Errorf("[Handler:%s] -  Failed to trigger action due to error - {%s}.", consumer.handler.Name(), err.Error())
	}

	if consumer.durableSub {
		if consumer.ackMode == eftl.AcknowledgeModeClient {
			if data != nil && data["ack"] != nil {
				// Explicit ack through TCM Message Ack
				// Spawning separate goroutine to avoid deadlock in connection as ack call would need a lock on the connection which is already locked by the message receiver.
				go consumer.ackMessage(m)

			} else {
				consumer.handler.Logger().Warnf("[Handler:%s] -  Message is not acknowledged. Subscriber is configured with 'Explicit' ack mode but 'TCM Message Ack' activity is not configured in the corresponding flow. Either configure 'TCM Message Ack' activity in the flow or change ack mode to 'Auto'", consumer.handler.Name())
			}
		}

		if consumer.ackMode != eftl.AcknowledgeModeClient && data != nil && data["ack"] != nil {
			consumer.handler.Logger().Warnf("[Handler:%s] -  Explicit message acknowledgement is not supported for Auto mode. Either remove 'TCM Message Ack' activity from the flow or change ack mode to 'Explicit'.", consumer.handler.Name())
		}
	} else {
		if data != nil && data["ack"] != nil {
			consumer.handler.Logger().Warnf("[Handler:%s] -  Explicit message acknowledgement is not supported for non durable subscriber. Either remove 'TCM Message Ack' activity from the flow or change subscriber to durable subscriber with 'Explicit' ack mode.", consumer.handler.Name())
		}
	}
}

func extractTraceContext(m eftl.Message) map[string]string {
	msgProp := make(map[string]string)
	traceparent, ok := m["_traceparent"].(string)
	if ok {
		// Fix for open telemetry
		msgProp["traceparent"] = traceparent
	}
	return msgProp
}

func (consumer *consumer) ackMessage(m eftl.Message) {
	// Explicit ack through TCM Message Ack
	err := consumer.tcmConnection.GetEFTLConnection().Acknowledge(m)
	if err != nil {
		consumer.handler.Logger().Errorf("[Handler:%s] -  Failed to acknowledge message due to error - {%s}.", consumer.handler.Name(), err.Error())
	} else {
		consumer.handler.Logger().Debugf("[Handler:%s] -  Message is successfully acknowledged", consumer.handler.Name())
	}
}

func messageToMap(msg eftl.Message) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range msg {
		if strings.HasPrefix(k, "_") {
			//exclude internal properties start with _
			continue
		}
		if v != nil {
			switch t := (v).(type) {
			case eftl.Message:
				m[k] = messageToMap(t)
			case []eftl.Message:
				var msgs []interface{}
				for _, v := range t {
					msgs = append(msgs, messageToMap(v))
				}
				m[k] = msgs
			default:
				m[k] = v
			}
		} else {
			m[k] = v
		}
	}

	return m
}
