package pub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"

	flogomq "github.com/tibco/wi-ibmmq/src/app/IBM-MQ"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// var log = logger.GetLogger("flogo-ibmmq-pub")
var versionPrinted = false

func init() {

	_ = activity.Register(&PubActivity{}, New)
}

// PubActivity is a stub for your Activity implementation
type PubActivity struct {
	settings     *Settings
	topicName    string
	topicDynamic string
	retain       bool
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	act := &PubActivity{settings: s}
	return act, nil
}

// Metadata implements activity.Activity.Metadata
func (a *PubActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *PubActivity) Eval(context activity.Context) (done bool, err error) {
	var log log.Logger = context.Logger()
	if !versionPrinted {
		flogomq.PrintVersion(log)
		versionPrinted = true
	}
	log.Debugf("%s IBM MQ pub eval start", context.Name())
	// do eval
	var input Input
	err = context.GetInputObject(&input)

	msgid, corrid, err := a.PubMessage(input, context.Name(), log)
	if err != nil {
		fmt.Print(err)
		return false, err
	}
	output := &Output{}
	result := make(map[string]interface{})
	result["CorrelId"] = corrid
	result["MsgId"] = msgid

	output.Output = result
	context.SetOutputObject(output)

	log.Debugf("%s IBM MQ pub eval end", context.Name())
	return true, nil
}

// PubMessage Take the pub activity's input configs and
// mappings and publish a message based on that.  Return
// the correlation and message id's
func (a *PubActivity) PubMessage(input Input, activityName string, logger log.Logger) (MsgID string, CorrelID string, err error) {
	connectionObj := a.settings.Connection
	if connectionObj == nil {
		return "", "", fmt.Errorf("%s internal error - no connection name specified in input", activityName)
	}
	connection, err := generic.NewConnection(connectionObj)
	if err != nil {
		return "", "", fmt.Errorf("%s internal error - no connection settings in config", activityName)
	}

	qMgr, err := flogomq.GetQueueManager(activityName, connection, logger)
	if err != nil {
		return "", "", err
	}
	defer flogomq.ReturnQueueManager(a.settings.Connection, activityName, qMgr, logger)

	a.topicName = a.settings.Topic
	if len(strings.TrimSpace(input.Topic)) != 0 {
		a.topicName = input.Topic
	}
	a.topicDynamic = a.settings.Topicdynamic
	if len(strings.TrimSpace(input.Topicdynamic)) != 0 {
		a.topicDynamic = input.Topicdynamic
	}

	publicationObj, err := qMgr.OpenTopicForPub(a.topicName, a.topicDynamic, activityName, a.settings.ContextSupport, logger)
	if err != nil {
		return "", "", err
	}

	var buffer []byte
	pmo := ibmmq.NewMQPMO()
	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT

	//Put to Topic only for publish activity Retain Publication is true
	a.retain = a.settings.Retain
	if a.retain == true {
		pmo.Options |= ibmmq.MQPMO_RETAIN
	}

	if flogomq.HaveMessagePropertiesInHandle(input.Properties, logger) {
		propMessageHandle, err := flogomq.GetMessagePropertiesInHandle(input.Properties, activityName, qMgr, logger)
		if err != nil {
			logger.Debugf("%s got an error processing properties [%s]", activityName, err)
		}
		pmo.OriginalMsgHandle = propMessageHandle
		defer pmo.OriginalMsgHandle.DltMH(ibmmq.NewMQDMHO())
	}

	if a.settings.ValueType == "String" {
		msgData := input.MessageString
		buffer = []byte(msgData)

	} else {
		if input.MessageJson != nil {
			messageJSONObj := input.MessageJson
			buffer, err = json.Marshal(messageJSONObj)
			if err != nil {
				logger.Debugf("%s failed to decode map input for Put on queue [%s] for reason [%s]", activityName, publicationObj.Name, err)
				return "", "", fmt.Errorf("%s failed to decode map input for Put on queue [%s] for reason [%s]", activityName, publicationObj.Name, err)
			}
		} else {
			buffer = make([]byte, 0)
		}
	}
	connUsername := connection.GetSetting("username").(string)
	putMqmd, err := flogomq.GetMqmdFromContext(input.MQMD, activityName, a.settings.MessageType, a.settings.ContextSupport, pmo, connUsername, logger)
	if err != nil {
		return "", "", err
	}

	err = publicationObj.Put(putMqmd, pmo, buffer)
	if err != nil && err.(*ibmmq.MQReturn).MQCC != ibmmq.MQCC_WARNING {
		logger.Debugf("%s failed to put to queue [%s] for reason [%s]", activityName, publicationObj.Name, err)
		return "", "", fmt.Errorf("%s putmessage failed on the call to Put on queue [%s] for reason [%s]", activityName, publicationObj.Name, err)
	} else if err != nil && err.(*ibmmq.MQReturn).MQCC == ibmmq.MQCC_WARNING {
		logger.Warnf("%s activity got this warning on the put api call [%s]", activityName, err)
	}
	logger.Debugf("%s pub successfull; returning MsgId [%x] CorrId [%x]", activityName, putMqmd.MsgId, putMqmd.CorrelId)
	return base64.StdEncoding.EncodeToString(putMqmd.MsgId), base64.StdEncoding.EncodeToString(putMqmd.CorrelId), nil
}
