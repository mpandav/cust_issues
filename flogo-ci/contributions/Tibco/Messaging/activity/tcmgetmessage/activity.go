package tcmgetmessage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	goctx "context"

	"github.com/TIBCOSoftware/eftl"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/tibco/flogo-messaging/src/app/Messaging/connector/tcm"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	var err error
	s := &Settings{}

	err = metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	act := &MyActivity{}

	//Content Matcher
	matcher := make(map[string]interface{})

	if s.Matcher != "" {
		if strings.HasPrefix(s.Matcher, "{") {
			// Matcher configured through app property in the format {"m1": "v1", "m2": v2}
			matcher, err = coerce.ToObject(s.Matcher)
			if err != nil {
				return nil, fmt.Errorf(fmt.Sprintf("Invalid content matcher [%s]. It must be a valid JSON object {\"k1\":\"v1\"}.", s.Matcher))
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
			return nil, fmt.Errorf(fmt.Sprintf("Invalid content matcher [%s]. It must be a valid JSON object {\"k1\":\"v1\"}.", s.Matcher))
		}
	}

	dest := s.Destination
	if len(dest) > 0 {
		matcher["_dest"] = dest
	}
	var matcherJson []byte
	matcherJson, err = json.Marshal(matcher)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Failed to convert matcher to string due to error - {%s}", err.Error()))
	}
	matcherString := string(matcherJson)
	ctx.Logger().Debugf("Matcher:%s", matcherString)

	//connection
	act.tcmConnection, err = s.ConnectionManager.(*tcm.TCMSharedConfigManager).GetSubscribeConnection()
	if err != nil {
		if strings.Contains(err.Error(), eftl.ErrInvalidResponse.Error()) || strings.Contains(err.Error(), eftl.ErrMessageTooBig.Error()) || strings.Contains(err.Error(), eftl.ErrNotAuthenticated.Error()) || strings.Contains(err.Error(), eftl.ErrNotAuthorized.Error()) {
			return nil, fmt.Errorf("Error get or create connect to TCM: %s", err.Error())
		} else {
			return nil, activity.NewRetriableError(fmt.Sprintf("Failed to connect to TIBCO Cloud Messaging service due to error - {%s}.", err.Error()), "TCM-MESSAGEGET-4005", nil)
		}
	}
	ctx.Logger().Debugf("Connection established")

	//subscription
	subChan := make(chan *eftl.Subscription, 1)
	msgChan := make(chan eftl.Message)

	err = act.subscribe(msgChan, subChan, matcherString, s.DurableName)
	if err != nil {
		ctx.Logger().Errorf("Subscribe failed due to error - {%s}, Now trying to reconnect", err.Error())
		err = act.tcmConnection.Reconnect("GetMessageActivity")
		if err != nil {
			return nil, fmt.Errorf("Failed to reconnect due to error - {%s}", err.Error())
		}

		err = act.subscribe(msgChan, subChan, matcherString, s.DurableName)
		if err != nil {
			return nil, fmt.Errorf("Subscription failed due to error - {%s}.", err.Error())
		}
		err = act.handleSubscription(ctx, subChan, matcherString, s.DurableName)
	} else {
		err = act.handleSubscription(ctx, subChan, matcherString, s.DurableName)
	}

	return act, nil
}

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
	subscribed    bool
	tcmConnection tcm.Connection
	eftlSub       *eftl.Subscription
}

func (*MyActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (consumerAct *MyActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debugf("Executing TCM getMessage activity")

	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	context.Logger().Info("Waiting for message")

	ctx, _ := goctx.WithTimeout(goctx.Background(), time.Duration(input.Timeout*int(time.Millisecond)))

	select {
	case m, ok := <-consumerAct.eftlSub.MessageChan:
		if ok {
			consumerAct.callHandle(input, context, m)
		}
	case <-ctx.Done():
		return false, activity.NewRetriableError("Get message operation timed out", "TCM-MESSAGEGET-4005", nil)
	}

	return true, nil
}

func (consumerAct *MyActivity) Cleanup() error {
	var err error
	if consumerAct.eftlSub != nil {
		err = consumerAct.tcmConnection.GetEFTLConnection().Unsubscribe(consumerAct.eftlSub)
	}
	return err
}

func (consumerAct *MyActivity) subscribe(msgChan chan eftl.Message, subChan chan *eftl.Subscription, matcherString string, durableName string) (err error) {
	if durableName == "" {
		return fmt.Errorf("Durable name must be set for durable subscription")
	}
	subOpts := eftl.SubscriptionOptions{}
	subOpts.AcknowledgeMode = "auto"
	subOpts.DurableType = "shared"
	err = consumerAct.tcmConnection.GetEFTLConnection().SubscribeWithOptionsAsync(matcherString, durableName, subOpts, msgChan, subChan)
	return err
}

func (consumerAct *MyActivity) handleSubscription(context activity.InitContext, subChan chan *eftl.Subscription, matcherString string, durableName string) (err error) {
	sub := <-subChan
	{
		if sub.Error != nil {
			//Close old msg channel
			close(consumerAct.eftlSub.MessageChan)
			context.Logger().Errorf("Subscription [%s] failed due to error - {%s}, trying to re-subscribe", durableName, sub.Error.Error())

			reConSubChan := make(chan *eftl.Subscription, 1)
			reConnectMsgChan := make(chan eftl.Message)
			err = consumerAct.subscribe(reConnectMsgChan, reConSubChan, matcherString, durableName)
			if err != nil {
				panic(fmt.Errorf("Subscription [%s] failed to resubscribe with TCM server due to error - {%s}", durableName, err.Error()))
			}

			context.Logger().Infof("Subscription [%s] successfully resubscribed.", durableName)
			consumerAct.subscribed = false
			consumerAct.handleSubscription(context, reConSubChan, matcherString, durableName)
			return nil
		}

		if !consumerAct.subscribed {
			consumerAct.eftlSub = sub
			consumerAct.subscribed = true
			context.Logger().Debugf("Subscribed successfully")
			return nil
		}
	}
	return nil

}

func (consumerAct *MyActivity) callHandle(input *Input, contextAct activity.Context, m eftl.Message) {

	contextAct.Logger().Infof("Message(ID:%d) received", m.StoreMessageId())
	if m.DeliveryCount() > 1 {
		contextAct.Logger().Infof("Message(ID:%d) is redelivered. DeliveryAttempt:%d", m.StoreMessageId(), m.DeliveryCount())
	}

	outputData := &Output{}
	outputData.Metadata = MessageMetadata{Id: m.StoreMessageId(), DeliveryCount: m.DeliveryCount()}
	outputData.Message = messageToMap(m)

	err := contextAct.SetOutputObject(outputData)
	if err != nil {
		contextAct.Logger().Errorf("Invalid output object %s", err.Error())
	}
}

func messageToMap(msg eftl.Message) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range msg {
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

	// Remove internal fields
	delete(m, "_dest")
	delete(m, "_eftl:sequenceNumber")
	delete(m, "_eftl:subscriptionId")
	delete(m, "_eftl:deliveryCount")
	delete(m, "_eftl:storeMessageId")

	return m
}
