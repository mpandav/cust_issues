/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package tcmpublisher

import (
	"fmt"
	"strings"
	"time"

	"github.com/TIBCOSoftware/eftl"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/tibco/flogo-messaging/src/app/Messaging/connector/tcm"
)

var activityMd = activity.ToMetadata(&Input{})

func init() {
	_ = activity.Register(&MyActivity{}, New)
}

func New(activity.InitContext) (activity.Activity, error) {
	return &MyActivity{}, nil
}

// MyActivity is a stub for your Activity implementation
type MyActivity struct {
}

func (*MyActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (producerAct *MyActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debugf("Executing TCM Publisher Activity")

	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	//Get Connection from single place where Subscriber and Publisher using same connection
	conn, err := input.Connection.(*tcm.TCMSharedConfigManager).GetPublisherConnection()
	if err != nil {
		return false, fmt.Errorf("error connecting to TIBCO Cloud Message server: %s", err.Error())
	}

	if input.Message == nil {
		return false, activity.NewError("Message must be configured", "TCM-SENDMESSAGE-4002", nil)
	}

	input.Message["_dest"] = input.Destination
	if trace.Enabled() {
		traceInfo := make(map[string]string)
		_ = trace.GetTracer().Inject(context.GetTracingContext(), trace.TextMap, traceInfo)
		if len(traceInfo) > 0 {
			for k, v := range traceInfo {
				// Inject returned keys as internal fields
				input.Message["_"+k] = v
			}
		}
	}

	//Encode value and publish
	id := context.ActivityHost().Name() + "/" + context.Name()
	return PublishMsg(conn, input.Message, id, context.Logger())
}

func PublishMsg(conn tcm.Connection, msgMap map[string]interface{}, name string, logger log.Logger) (bool, error) {
	conn.WaitingConnectionAvailable(name)
	msg := TCMMessage{}.Decode(msgMap)
	err := publish(conn.GetEFTLConnection(), msg.ToMap())
	logger.Debug(msg)
	if err != nil {
		// Don't disconnect and reconnect for below error
		if strings.HasPrefix(err.Error(), "json:") || err == eftl.ErrInvalidResponse || err == eftl.ErrMessageTooBig || err == eftl.ErrNotAuthenticated || err == eftl.ErrNotAuthorized {
			return false, activity.NewError(fmt.Sprintf("[%s] Failed to send message to TIBCO Cloud Messaging service due to error - {%s}.", name, err.Error()), "TCM-MESSAGEPUB-4005", nil)
		} else {
			// Reconnect and try for other error
			logger.Warnf("[%s] Failed to send message to TIBCO Cloud Messaging service due to error - {%s}, Trying to reconnect", name, err.Error())
			err = conn.Reconnect(name)
			if err != nil {
				return false, fmt.Errorf("[%s] Reconnect tcm connection error - {%s}", name, err.Error())
			}
			logger.Infof("[%s] Reconnected", name)
			err = publish(conn.GetEFTLConnection(), msg.ToMap())
			if err != nil {
				if strings.HasPrefix(err.Error(), "json:") || err == eftl.ErrInvalidResponse || err == eftl.ErrMessageTooBig || err == eftl.ErrNotAuthenticated || err == eftl.ErrNotAuthorized {
					return false, activity.NewError(fmt.Sprintf("[%s] Failed to send message to TIBCO Cloud Messaging service due to error - {%s}.", name, err.Error()), "TCM-MESSAGEPUB-4005", nil)
				}
				return false, activity.NewRetriableError(fmt.Sprintf("[%s] Failed to send message to TIBCO Cloud Messaging service due to error - {%s}.", name, err.Error()), "TCM-MESSAGEPUB-4005", nil)
			}
		}
	}
	logger.Info("Message published")
	return true, nil
}

func publish(conn *eftl.Connection, msg eftl.Message) error {
	completionChan := make(chan *eftl.Completion, 1)
	if err := conn.PublishAsync(msg, completionChan); err != nil {
		return err
	}
	select {
	case completion := <-completionChan:
		return completion.Error
	case <-time.After(60 * time.Second):
		return fmt.Errorf("publisher activity timeout after 60s")
	}
}

type TCMMessage map[string]interface{}

func (msg TCMMessage) ToMap() map[string]interface{} {
	return msg
}

func (msg TCMMessage) Decode(m map[string]interface{}) TCMMessage {
	return decodeObject(m)
}

func decodeObject(msg map[string]interface{}) map[string]interface{} {
	newMsg := make(map[string]interface{})
	for k, v := range msg {
		switch t := v.(type) {
		default:
			newMsg[k] = t
		case map[string]interface{}:
			newMsg[k] = decodeObject(t)
		case []interface{}:
			newMsg[k] = decodeArray(t)
		}
	}
	return newMsg
}

func decodeArray(array []interface{}) interface{} {
	if len(array) <= 0 {
		return []string{}
	}

	if len(array) > 0 {
		switch array[0].(type) {
		case int:
			s := make([]int, 0, len(array))
			for _, elem := range array {
				i, _ := coerce.ToInt(elem)
				s = append(s, i)
			}
			return s
		case int32:
			s := make([]int32, 0, len(array))
			for _, elem := range array {
				i, _ := coerce.ToInt32(elem)
				s = append(s, i)
			}
			return s
		case int64:
			s := make([]int64, 0, len(array))
			for _, elem := range array {
				i, _ := coerce.ToInt64(elem)
				s = append(s, i)
			}
			return s
		case float64:
			s := make([]float64, 0, len(array))
			for _, elem := range array {
				i, _ := coerce.ToFloat64(elem)
				s = append(s, i)
			}
			return s
		case float32:
			s := make([]float32, 0, len(array))
			for _, elem := range array {
				i, _ := coerce.ToFloat32(elem)
				s = append(s, i)
			}
			return s
		case string:
			s := make([]string, 0, len(array))
			for _, elem := range array {
				i, _ := coerce.ToString(elem)
				s = append(s, i)
			}
			return s
		case map[string]interface{}:
			s := make([]map[string]interface{}, 0, len(array))
			for _, elem := range array {
				obj, _ := coerce.ToObject(elem)
				s = append(s, decodeObject(obj))
			}
			return s
		case []interface{}:
			return array
			//This not support by TCM. just add to TCM and it will throw error

			//for _, elem := range array {
			//	array, _ := coerce.ToArray(elem)
			//	converetedArray := decodeArray(array)
			//}
			//return decodeArray(array)
		default:
			return array
		}
	}
	return array
}
