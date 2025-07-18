package receivemessage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"
	"github.com/tibco/flogo-ems/src/app/EMS/connector/ems"
	"github.com/tibco/msg-ems-client-go/tibems"
)

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

type Factory struct {
}

func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}
	err := metadata.MapToStruct(config.Settings, s, true)

	if err != nil {
		return nil, err
	}

	emsTrigger := &Trigger{
		triggerName:      config.Id,
		emsConnectionMgr: s.Connection,
	}

	return emsTrigger, nil
}

type Trigger struct {
	connection       *tibems.Connection
	emsConnectionMgr connection.Manager
	triggerName      string
	EMSHandlers      []*EMSHandler
}

type EMSHandler struct {
	triggerName             string
	ackMode                 string
	processingMode          string
	destination             string
	destinationType         string
	destinationObj          *tibems.Destination
	durableSubscriber       bool
	sharedDurableSubscriber bool
	subscriptionName        string
	consumer                *tibems.MsgConsumer
	session                 *tibems.Session
	handler                 trigger.Handler
	revMsgChan              chan ReceiveMessage
	done                    chan bool
	logger                  log.Logger
	unAcknowledgeMsgs       bool
}

type ReceiveMessage struct {
	message tibems.Message
	err     error
}

type Property struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (emsTrigger *Trigger) Initialize(ctx trigger.InitContext) (err error) {

	for _, handler := range ctx.GetHandlers() {
		handlerSetting := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)
		if err != nil {
			return err
		}
		emsHandler := &EMSHandler{}
		emsHandler.triggerName = emsTrigger.triggerName
		emsHandler.destination = handlerSetting.Destination
		emsHandler.destinationType = handlerSetting.DestinationType
		emsHandler.ackMode = handlerSetting.AckMode
		emsHandler.processingMode = handlerSetting.ProcessingMode
		emsHandler.durableSubscriber = handlerSetting.DurableSubscriber
		emsHandler.subscriptionName = handlerSetting.SubscriptionName
		emsHandler.sharedDurableSubscriber = handlerSetting.SharedDurableSubscriber
		emsHandler.handler = handler
		emsHandler.revMsgChan = make(chan ReceiveMessage)
		emsHandler.done = make(chan bool)
		emsHandler.logger = ctx.Logger()
		emsHandler.unAcknowledgeMsgs = false
		emsTrigger.EMSHandlers = append(emsTrigger.EMSHandlers, emsHandler)
	}
	return err
}

func (emsTrigger *Trigger) Start() (err error) {
	emsTrigger.connection = emsTrigger.emsConnectionMgr.GetConnection().(*tibems.Connection)
	for _, handler := range emsTrigger.EMSHandlers {
		go handler.Start(emsTrigger)
	}
	return nil
}

func (emsTrigger *Trigger) Stop() error {
	for _, handler := range emsTrigger.EMSHandlers {
		handler.Stop()
	}
	return nil
}

func (emsHandler *EMSHandler) Start(emsTrigger *Trigger) (err error) {
	// Get connection manager settings
	emsConnMgr, ok := emsTrigger.emsConnectionMgr.(*ems.EmsSharedConfigManager)
	if !ok {
		return fmt.Errorf("invalid connection manager")
	}

	// Validate ClientID requirements
	if emsHandler.destinationType == "Topic" && emsHandler.durableSubscriber {
		if emsHandler.subscriptionName == "" {
			return fmt.Errorf("subscription name is required for durable subscribers")
		}

		if emsHandler.sharedDurableSubscriber {
			if emsConnMgr.Settings.ClientID != "" {
				emsHandler.logger.Errorf("Error: Client ID should not be set for shared durable subscriber")
				return fmt.Errorf("Error: client ID should not be set for shared durable subscriber")
			}
		} else {
			if emsConnMgr.Settings.ClientID == "" {
				emsHandler.logger.Errorf("Error: Client ID is required for durable subscriber")
				return fmt.Errorf("Error: Client ID is required for durable subscriber")
			}
		}
	}
	// Create destination obj for consumer
	if emsHandler.destinationType == "Queue" {
		emsHandler.destinationObj, err = tibems.CreateDestination(tibems.DestTypeQueue, emsHandler.destination)
		if err != nil {
			emsHandler.logger.Error(err)
			return err
		}
	} else {
		emsHandler.destinationObj, err = tibems.CreateDestination(tibems.DestTypeTopic, emsHandler.destination)
		if err != nil {
			emsHandler.logger.Error(err)
			return err
		}
	}
	// Create Session
	if emsHandler.ackMode == "Explicit client" {
		emsHandler.session, err = emsTrigger.connection.CreateSession(false, tibems.AckModeExplicitClientAcknowledge)
	} else if emsHandler.ackMode == "Explicit Client Dups OK" {
		emsHandler.session, err = emsTrigger.connection.CreateSession(false, tibems.AckModeExplicitClientDupsOkAcknowledge)
	} else {
		emsHandler.session, err = emsTrigger.connection.CreateSession(false, tibems.AckModeAutoAcknowledge)
	}
	if err != nil {
		return err
	}
	// Create consumer
	if emsHandler.destinationType == "Topic" && emsHandler.durableSubscriber {
		topicDest := (*tibems.Topic)(emsHandler.destinationObj)
		if emsHandler.sharedDurableSubscriber {
			emsHandler.logger.Infof("Creating SHARED durable consumer '%s'", emsHandler.subscriptionName)
			emsHandler.consumer, err = emsHandler.session.CreateSharedDurableConsumer(topicDest,
				emsHandler.subscriptionName, "")
		} else {
			emsHandler.logger.Infof("Creating DURABLE subscriber '%s'", emsHandler.subscriptionName)
			emsHandler.consumer, err = emsHandler.session.CreateDurableSubscriber(topicDest,
				emsHandler.subscriptionName, "", false)
		}
	} else {
		// Handles non-durable topics and queues
		emsHandler.consumer, err = emsHandler.session.CreateConsumer(emsHandler.destinationObj, "", false)
	}
	if err != nil {
		emsHandler.logger.Errorf("Failed to start handler on trigger %s due : %s", emsHandler.triggerName, err.Error())
		return err
	}
	go func() {
		for {
			msg, err := emsHandler.consumer.Receive()
			emsHandler.revMsgChan <- ReceiveMessage{message: msg, err: err}
		}
	}()
	// session recover go routine to receive unacknowledge messages
	go func() {
		ticker := time.NewTicker(1 * time.Minute)

		for {
			select {
			case <-emsHandler.done:
				return
			case <-ticker.C:
				if emsHandler.unAcknowledgeMsgs {
					err = emsHandler.session.Recover()
					if err != nil {
						emsHandler.logger.Errorf("error while recovering session: ", err)
					}
					emsHandler.unAcknowledgeMsgs = false
				}
			}
		}
	}()

	for {
		select {
		case revMsg := <-emsHandler.revMsgChan:
			if revMsg.err != nil {
				emsHandler.logger.Error(revMsg.err)
				return revMsg.err
			}
			if emsHandler.processingMode == "Async" {
				// Handle messages concurrently on separate goroutine
				go emsHandler.handleMessage(revMsg)
			} else {
				emsHandler.handleMessage(revMsg)
			}
		case <-emsHandler.done:
			return
		}
	}
}

func (emsHandler *EMSHandler) handleMessage(revMsg ReceiveMessage) {
	emsHandler.handler.Logger().Infof("Message received")
	var err error
	out := &Output{
		Headers:           make(map[string]interface{}),
		MessageProperties: make(map[string]interface{}),
	}
	message := revMsg.message.(*tibems.TextMsg)
	out.Message, err = message.GetText()
	if err != nil {
		emsHandler.logger.Error(err)
		return
	}

	// message headers
	msgDestination, err := message.GetDestination()
	if err == nil {
		out.Headers["destination"], _ = msgDestination.GetName()
	}

	replytodest, err := message.GetReplyTo()
	if err == nil && replytodest != nil {
		out.Headers["replyTo"], _ = replytodest.GetName()
	}

	deliveryMode, err := message.GetDeliveryMode()
	if err == nil {
		if deliveryMode == tibems.DeliveryModeNonPersistent {
			out.Headers["deliveryMode"] = "nonpersistent"
		} else if deliveryMode == tibems.DeliveryModePersistent {
			out.Headers["deliveryMode"] = "persistent"
		} else {
			out.Headers["deliveryMode"] = "reliable"
		}
	}

	out.Headers["messageID"], _ = message.GetMessageID()
	out.Headers["timestamp"], _ = message.GetTimestamp()
	out.Headers["expiration"], _ = message.GetExpiration()
	out.Headers["reDelivered"], _ = message.GetRedelivered()
	out.Headers["priority"], _ = message.GetPriority()
	out.Headers["correlationID"], _ = message.GetCorrelationID()

	ctx := context.Background()
	if trace.Enabled() {
		tracingHeader := make(map[string]string)
		messageEnumeration, err := message.GetPropertyNames()
		if err != nil {
			emsHandler.logger.Info("error while extracting opentracing header from message")
			emsHandler.logger.Error(err)
		}
		for {
			propertyName, err := messageEnumeration.GetNextName()
			if err != nil {
				break
			}
			tracingHeader[propertyName], _ = message.GetStringProperty(propertyName)
		}

		tc, _ := trace.GetTracer().Extract(trace.TextMap, tracingHeader)
		if tc != nil {
			ctx = trace.AppendTracingContext(ctx, tc)
		}
	}
	if out.Headers["messageID"] != "" {
		ctx = trigger.NewContextWithEventId(ctx, out.Headers["messageID"].(string))
	}

	//message properties
	getMessageProperties(emsHandler.logger, message, out)

	attrs, err := emsHandler.handler.Handle(ctx, out)
	if emsHandler.ackMode == "Explicit client" || emsHandler.ackMode == "Explicit Client Dups OK" {
		if err != nil {
			emsHandler.logger.Errorf("Error while processing message %s", err.Error())
			emsHandler.unAcknowledgeMsgs = true
		} else {
			if attrs["messageAck"] != nil && attrs["messageAck"] == true {
				err = revMsg.message.Acknowledge()
				if err != nil {
					emsHandler.logger.Errorf("Error while acknowledging message %s", err)
				}
			} else {
				emsHandler.unAcknowledgeMsgs = true
			}
		}
	}
}

func getMessageProperties(logger log.Logger, message tibems.Message, out *Output) {
	messageEnumeration, err := message.GetPropertyNames()
	if err == nil {
		var properties []Property
		for {
			propertyName, err := messageEnumeration.GetNextName()
			if err != nil {
				break
			}

			propertyValue, err := message.GetStringProperty(propertyName)
			if err != nil {
				logger.Errorf("Error getting property value for %s: %v", propertyName, err)
				continue
			}

			// Get property type, default to STRING if not found
			propertyType, err := message.GetPropertyType(propertyName)
			if err != nil {
				propertyType = 9 // 9 corresponds to STRING
			}

			var propertyTypeString string
			switch propertyType {
			case 1:
				propertyTypeString = "boolean"
			case 4:
				propertyTypeString = "short"
			case 5:
				propertyTypeString = "integer"
			case 6:
				propertyTypeString = "long"
			case 7:
				propertyTypeString = "float"
			case 8:
				propertyTypeString = "double"
			case 9:
				propertyTypeString = "string"
			default:
				propertyTypeString = "string"
			}
			properties = append(properties, Property{
				Name:  propertyName,
				Type:  propertyTypeString,
				Value: propertyValue,
			})
		}
		propMap := make(map[string]interface{})
		propMap["property"] = properties
		var msgProps interface{}
		tmp, _ := json.Marshal(propMap)
		json.Unmarshal(tmp, &msgProps)
		out.MessageProperties = msgProps
	} else {
		logger.Errorf("Error getting property names: %v", err)
	}
}

func (emsHandler *EMSHandler) Stop() (err error) {
	close(emsHandler.done)
	err = emsHandler.consumer.Close()
	if err != nil {
		return err
	}
	err = emsHandler.destinationObj.Close()
	if err != nil {
		return err
	}
	err = emsHandler.session.Close()
	if err != nil {
		return err
	}
	return nil
}
