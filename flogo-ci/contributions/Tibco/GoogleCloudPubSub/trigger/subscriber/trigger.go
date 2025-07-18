package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

var permissions = []string{"pubsub.subscriptions.consume", "pubsub.subscriptions.get"}

const JSONMessageFormat = "JSON"

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

	t := &Trigger{name: config.Id}
	t.client = s.GooglePubSubConnection.GetConnection().(*pubsub.Client)
	return t, nil
}

type Trigger struct {
	subscriptions map[string]*SubscriptionHandler
	client        *pubsub.Client
	name          string
}

type SubscriptionHandler struct {
	cancelFunc    context.CancelFunc
	subscriber    *pubsub.Subscription
	handler       trigger.Handler
	logger        log.Logger
	messageFormat string
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {
	var err error
	t.subscriptions = make(map[string]*SubscriptionHandler)
	for _, handler := range ctx.GetHandlers() {
		handlerSetting := &HandlerSettings{}
		var err = metadata.MapToStruct(handler.Settings(), handlerSetting, true)
		if err != nil {
			return err
		}

		if handlerSetting.SubscriptionId == "" {
			handler.Logger().Error("Subscription Id must be configured")
			return errors.New("invalid subscription id")
		}

		if strings.Contains(handlerSetting.SubscriptionId, "/") && strings.HasPrefix(handlerSetting.SubscriptionId, "projects/") {
			ctx.Logger().Warnf("Subscription name '%s' found instead of Subscription Id", handlerSetting.SubscriptionId)
			segments := strings.Split(handlerSetting.SubscriptionId, "/")
			handlerSetting.SubscriptionId = segments[len(segments)-1]
		}
		ctx.Logger().Infof("Subscription Id set to '%s'", handlerSetting.SubscriptionId)

		subHandler := &SubscriptionHandler{}
		subHandler.subscriber = t.client.Subscription(handlerSetting.SubscriptionId)

		// Check required permissions
		perms, err := subHandler.subscriber.IAM().TestPermissions(context.Background(), permissions)
		if err == nil {
			if len(perms) != 2 {
				// Required permissions not configured
				handler.Logger().Error("The IAM role configured in the service account must have ['pubsub.subscriptions.get', 'pubsub.subscriptions.consume'] permissions.")
				return errors.New("insufficient IAM role permissions")
			}
		} else {
			if status.Code(err) == codes.NotFound {
				handler.Logger().Errorf("Subscription(Id:%s) not found in Google project. Check Subscription Id and Google Project Id.", handlerSetting.SubscriptionId)
				return errors.New("subscription not found")
			} else {
				handler.Logger().Errorf("Unable to validate required permissions due to error - %s", err.Error())
				return errors.New("failed to validate required permissions")
			}
		}

		if handlerSetting.FlowControlMode {
			subHandler.subscriber.ReceiveSettings.Synchronous = true
		}

		subHandler.subscriber.ReceiveSettings.MaxOutstandingMessages = handlerSetting.MaxOutstandingMessages
		handler.Logger().Infof("Message Format: %s, Flow Controlled: %v, Maximum Messages to be processed: %d", handlerSetting.MessageDataFormat, handlerSetting.FlowControlMode, handlerSetting.MaxOutstandingMessages)
		subHandler.handler = handler
		subHandler.logger = handler.Logger()
		subHandler.messageFormat = handlerSetting.MessageDataFormat
		t.subscriptions[handler.Name()] = subHandler
	}
	return err
}

func (t *Trigger) Start() error {
	for _, handler := range t.subscriptions {
		go handler.start()
	}
	return nil
}

func (t *Trigger) Stop() error {
	for _, handler := range t.subscriptions {
		handler.cancelFunc()
	}
	return nil
}

func (h *SubscriptionHandler) start() {
	var ctx context.Context
	ctx, h.cancelFunc = context.WithCancel(context.Background())
	h.logger.Info("Starting receiver goroutine")
	var counter = 1
	waitInterval := time.Second * 15
	for {
		err := h.subscriber.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			// reset the counter
			counter = 1
			h.logger.Infof("Message(Id:%s) received from the subscription(Id:%s) ", msg.ID, h.subscriber.ID())
			output := &Output{}
			if h.messageFormat == JSONMessageFormat {
				if json.Valid(msg.Data) {
					var data map[string]interface{}
					err := json.Unmarshal(msg.Data, &data)
					if err != nil {
						h.logger.Errorf("Failed to process JSON message data due to error - %s", err.Error())
						h.logger.Infof("Due to non recoverable error, message(Id:%s) is acknowledged by subscriber to avoid redelivery", msg.ID)
						msg.Ack()
						return
					}
					output.MessageData = data
				} else {
					h.logger.Errorf("Subscriber message format is set to JSON but payload received is not a valid JSON. Check subscriber [%s] configuration.", h.handler.Name())
					msg.Nack()
					return
				}
			} else {
				output.MessageData = string(msg.Data)
			}
			output.MessageMetaData = MessageMetadata{Attributes: make(map[string]string)}
			output.MessageMetaData.Id = msg.ID
			if msg.Attributes != nil && len(msg.Attributes) > 0 {
				output.MessageMetaData.Attributes = msg.Attributes
			}
			if msg.DeliveryAttempt != nil {
				output.MessageMetaData.DeliveryAttempt = *msg.DeliveryAttempt
			}
			_, err := h.handler.Handle(context.Background(), output)
			if err == nil {
				msg.Ack()
				h.logger.Infof("Message(Id:%s) acknowledged by subscriber", msg.ID)
			} else {
				h.logger.Errorf("Failed to trigger action due to error - %s. Message(Id:%s) is not acknowledged by subscriber", err.Error(), msg.ID)
				msg.Nack()
			}
		})
		if err == context.Canceled || status.Code(err) == codes.Canceled {
			h.logger.Debugf("Stopping receiver goroutine")
			return
		} else if status.Code(err) == codes.Unavailable || status.Code(err) == codes.Internal || status.Code(err) == codes.ResourceExhausted {
			if counter > 10 {
				h.logger.Errorf("All attempts to receive messages from the subscription(Id:%s) are exhausted. Terminating receiver goroutine. Check subscription in Google project", h.subscriber.String())
				return
			}
			sleepInterval := waitInterval * time.Duration(counter)
			h.logger.Errorf("Failed to receive messages for the subscription(Id:%s) due to transient error - %s. Attempting to resubscribe in %d seconds...", h.subscriber.String(), err.Error(), sleepInterval)
			time.Sleep(sleepInterval)
			counter++
		} else if err != nil {
			h.logger.Errorf("Failed to receive messages for the subscription(Id:%s) due to error - %s", h.subscriber.String(), err.Error())
			h.logger.Info("Stopping receiver goroutine")
			return
		}
	}
}
