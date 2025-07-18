package put

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

// var log = logger.GetLogger("flogo-ibmmq-put")
var versionPrinted = false

func init() {

	_ = activity.Register(&PutActivity{}, New)
}

// PutActivity is a stub for your Activity implementation
type PutActivity struct {
	settings  *Settings
	queueName string
	queueMgr  string
}

// Metadata implements activity.Activity.Metadata
func (a *PutActivity) Metadata() *activity.Metadata {
	return activityMd
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	act := &PutActivity{settings: s}
	return act, nil
}

// Eval implements activity.Activity.Eval
func (a *PutActivity) Eval(context activity.Context) (done bool, err error) {
	var log log.Logger = context.Logger()
	if !versionPrinted {
		flogomq.PrintVersion(log)
		versionPrinted = true
	}
	log.Debugf("%s IBM MQ put eval start", context.Name())
	// do eval
	var input Input
	err = context.GetInputObject(&input)
	msgid, corrid, err := a.PutMessage(input, context.Name(), log)
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

	log.Debugf("%s IBM MQ put eval end", context.Name())
	return true, nil
}

// PutMessage Take the put activity's input configs and
// mappings and send a message based on that.  Return
// the correlation and message id's
func (a *PutActivity) PutMessage(input Input, activityName string, logger log.Logger) (MsgID string, CorrelID string, err error) {
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

	a.queueName = a.settings.Queue
	if len(strings.TrimSpace(input.Queue)) != 0 {
		a.queueName = input.Queue
	}
	if len(a.queueName) == 0 {
		//which should be impossible because is required
		logger.Debugf("%s found no queue name in config", activityName)
		return "", "", fmt.Errorf("%s found no queue name in config", activityName)
	}
	a.queueMgr = a.settings.QueueMgr
	if len(strings.TrimSpace(input.QueueMgr)) != 0 {
		a.queueMgr = input.QueueMgr
	}
	queue, err := qMgr.GetPutQueue(a.queueName, a.queueMgr, a.settings.ContextSupport, logger)
	if err != nil {
		return "", "", err
	}
	var buffer []byte
	pmo := ibmmq.NewMQPMO()
	pmo.Options = ibmmq.MQPMO_NO_SYNCPOINT

	if a.settings.GenCorrelationID == true {
		pmo.Options |= ibmmq.MQPMO_NEW_CORREL_ID
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
				logger.Debugf("%s failed to decode map input for Put on queue [%s] for reason [%s]", activityName, queue.Name, err)
				return "", "", fmt.Errorf("%s failed to decode map input for Put on queue [%s] for reason [%s]", activityName, queue.Name, err)
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

	err = queue.Put(putMqmd, pmo, buffer)
	if err != nil && err.(*ibmmq.MQReturn).MQCC != ibmmq.MQCC_WARNING {
		logger.Debugf("%s failed to put to queue [%s] for reason [%s]", activityName, queue.Name, err)
		return "", "", fmt.Errorf("%s putmessage failed on the call to Put on queue [%s] for reason [%s]", activityName, queue.Name, err)
	} else if err != nil && err.(*ibmmq.MQReturn).MQCC == ibmmq.MQCC_WARNING {
		logger.Warnf("%s activity got this warning on the put api call [%s]", activityName, err)
	}
	err = qMgr.CloseDynamicQueue(a.queueName, logger)
	if err != nil {
		logger.Warnf("%s closing dynamic queue failed for reason [%s]", activityName, err)
	}
	logger.Debugf("%s put successfull; returning MsgId [%x] CorrId [%x]", activityName, putMqmd.MsgId, putMqmd.CorrelId)
	return base64.StdEncoding.EncodeToString(putMqmd.MsgId), base64.StdEncoding.EncodeToString(putMqmd.CorrelId), nil
}
