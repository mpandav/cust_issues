package salesforce

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"

	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

var logCache = log.ChildLogger(log.RootLogger(), "salesforce.trigger")

var triggerMd = trigger.NewMetadata(&HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&SalesforceTrigger{}, &MyTriggerFactory{})
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
	return &SalesforceTrigger{metadata: t.metadata, id: config.Id}, nil
}

// Metadata implements trigger.Factory.Metadata
func (*MyTriggerFactory) Metadata() *trigger.Metadata {
	return triggerMd
}

// SalesforceTrigger is a stub for your Trigger implementation
type SalesforceTrigger struct {
	metadata *trigger.Metadata
	//settings    *Settings
	id          string
	subscribers []*Subscriber
}

// Initialize creates the trigger context on startup.
func (t *SalesforceTrigger) Initialize(ctx trigger.InitContext) error {
	ctx.Logger().Debugf("Initializing trigger - %s", t.id)

	for _, handler := range ctx.GetHandlers() {
		handlerSettings := &HandlerSettings{}
		var err error
		err = metadata.MapToStruct(handler.Settings(), handlerSettings, true)
		if err != nil {
			return fmt.Errorf("Error occured in metadata.MapToStruct, error - [%s]", err.Error())
		}
		handlerSettings.Connection, err = sfconnection.GetSharedConfiguration(handler.Settings()["Connection Name"])
		if err != nil {
			return err
		}

		sscm, _ := handlerSettings.Connection.(*sfconnection.SalesforceSharedConfigManager)

		query := handlerSettings.Query
		channelName := handlerSettings.ChannelName
		subscriberType := handlerSettings.SubscriberType
		autoCreatePushTopic := handlerSettings.AutoCreatePushTopic
		if query == "" && channelName == "" {
			return fmt.Errorf("Channel name or Query field is not configured")
		}
		replayID := handlerSettings.ReplayID
		ctx.Logger().Debug("replayID : ", replayID)

		apiVersion := sscm.APIVersion
		if apiVersion == "" {
			apiVersion = sfconnection.DEFAULT_APIVERSION
		}

		versionNum, _ := strconv.ParseFloat(strings.TrimPrefix(apiVersion, "v"), 32)
		objectName := handlerSettings.ObjectName
		if objectName == "" {
			ctx.Logger().Errorf("missing object name")
			return fmt.Errorf("missing object name")
		}

		sub := &Subscriber{}
		sub.salesforceSharedConfigManager = sscm
		sub.subscriberType = subscriberType
		sub.autoCreatePushTopic = autoCreatePushTopic
		sub.replayID = replayID

		ctx.Logger().Debug("Subscriber type : ", sub.subscriberType)
		ctx.Logger().Debug("Auto Create PushTopic : ", sub.autoCreatePushTopic)
		if sub.subscriberType == "" || (sub.subscriberType == "PushTopic" && sub.autoCreatePushTopic == true) {
			ctx.Logger().Debug("Creating PushTopic...")
			tn := time.Now()
			st := fmt.Sprintf("%d%02d%02d%02d%02d%02d", tn.Year(), tn.Month(), tn.Day(), tn.Hour(), tn.Minute(), tn.Second())
			topicName := objectName + "-" + st

			if (utf8.RuneCountInString(topicName)) > 25 {
				topicName = topicName[0:25]
			}

			topicStr := fmt.Sprintf(`{
			"Name": "%s",
			"Query": "%s",
			"ApiVersion": %.1f,
			"NotifyForOperationCreate": true,
			"NotifyForOperationUpdate" : true,
			"NotifyForOperationUndelete" : true,
			"NotifyForOperationDelete" : true,
			"NotifyForFields" : "All"}`, topicName, query, versionNum)
			tp, err := sub.CreateTopic(topicStr)
			if err != nil {
				ctx.Logger().Errorf("Trigger not started, due to can't create salesforc push topic %s", err)
				panic(fmt.Sprintf("Trigger not started, due to can't create salesforc push topic %s", err))
			}
			ctx.Logger().Debugf("PushTopic created successfully")
			sub.topic = tp
		} else if sub.subscriberType == "PushTopic" && sub.autoCreatePushTopic == false {
			if strings.HasPrefix(channelName, "/topic/") {
				tp := PushTopic{Name: channelName[7:]}
				sub.topic = tp
			} else {
				return fmt.Errorf("Invalid channel name input")
			}
		} else if sub.subscriberType == "Change Data Capture" {
			if strings.HasPrefix(channelName, "/data/") {
				cdc := ChangeDataCapture{Name: channelName[6:]}
				sub.changeDataCapture = cdc
			} else {
				return fmt.Errorf("Invalid channel name input")
			}
		} else if sub.subscriberType == "Platform Event" {
			if strings.HasPrefix(channelName, "/event/") {
				platformEvent := PlatformEvent{Name: channelName[7:]}
				sub.platformEvent = platformEvent
			} else {
				return fmt.Errorf("Invalid channel name input")
			}
		}
		sub.handler = handler
		t.subscribers = append(t.subscribers, sub)
	}
	return nil
}

// Start implements trigger.Trigger.Start
func (t *SalesforceTrigger) Start() error {
	logCache.Debugf("Starting trigger - %s", t.id)

	fn := func(handler trigger.Handler, eventData interface{}) {

		out := &Output{}
		outputData := make(map[string]interface{})
		eData, err := json.Marshal(eventData)
		if err != nil {
			logCache.Errorf("Failed to Marshal eventData, error - %s", err.Error())
			return
		}
		err = json.Unmarshal(eData, &outputData)
		if err != nil {
			logCache.Errorf("Failed to Unmarshal eData, error - %s", err.Error())
			return
		}
		out.Output = outputData
		_, err = handler.Handle(context.Background(), out)
		if err != nil {
			logCache.Error("message handled failed: ", err.Error())
		}
	}

	for _, sub := range t.subscribers {
		if sub.subscriberType == "PushTopic" {
			logCache.Infof("Starting %s trigger to subscribe message on PushTopic: "+sub.topic.Name, t.id)
			go sub.ListenToPushTopic(fn)
		} else if sub.subscriberType == "Change Data Capture" {
			logCache.Info("Starting %s trigger to subscribe message on Change Data Capture: "+sub.changeDataCapture.Name, t.id)
			go sub.ListenToChangeDataCapture(fn)
		} else if sub.subscriberType == "Platform Event" {
			logCache.Info("Starting %s trigger to subscribe message on Platform Event: "+sub.platformEvent.Name, t.id)
			go sub.ListenToPlatformEvent(fn)
		}
	}
	return nil
}

// Stop implements ext.Trigger.Stop
func (t *SalesforceTrigger) Stop() error {
	logCache.Debugf("Stopping %s", t.id)

	for _, v := range t.subscribers {
		v.Stop()
	}
	return nil
}

func (t *SalesforceTrigger) Pause() error {
	return nil
}

func (t *SalesforceTrigger) Resume() error {
	return nil
}
