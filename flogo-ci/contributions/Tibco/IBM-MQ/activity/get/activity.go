package get

import (
	"fmt"
	"strings"
	"time"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"

	flogomq "github.com/tibco/wi-ibmmq/src/app/IBM-MQ"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// var logger = log.GetLogger("flogo-ibmmq-get")
var versionPrinted = false

const maxBufSize = 100 * 1024 * 1024 // 100 MB
var bufLen int

func init() {

	_ = activity.Register(&GetActivity{}, New)
}

// GetActivity is a stub for your Activity implementation
type GetActivity struct {
	settings     *Settings
	queueName    string
	msgID        string
	corellID     string
	waitInterval int32
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	act := &GetActivity{settings: s}
	return act, nil
}

// Metadata implements activity.Activity.Metadata
func (a *GetActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *GetActivity) Eval(context activity.Context) (done bool, err error) {
	var log log.Logger = context.Logger()
	log.Debugf("%s IBM MQ Get eval start", context.Name())
	if !versionPrinted {
		flogomq.PrintVersion(log)
		versionPrinted = true
	}

	// do eval
	var input Input
	err = context.GetInputObject(&input)
	getOutputData, err := a.GetMessage(input, context.Name(), log)
	if err != nil {
		fmt.Print(err)
		return false, err
	}
	outputMqmd := getOutputData["MQMD"]
	outputMessage := getOutputData["Message"]
	outputProps := getOutputData["MessageProperties"]
	context.SetOutput("MQMD", outputMqmd)
	if a.settings.ValueType == "String" {
		context.SetOutput("Message", outputMessage)
	} else {
		context.SetOutput("MessageJson", outputMessage)
	}
	context.SetOutput("MessageProperties", outputProps)

	log.Debugf("%s IBM MQ Get eval end", context.Name())
	return true, nil
}

// GetMessage return a message or a timeout error from the queue.
func (a *GetActivity) GetMessage(input Input, activityName string, logger log.Logger) (outputData map[string]interface{}, err error) {
	var buffer []byte
	if bufferVar := a.settings.MaxSize; int32(bufferVar) > 0 {
		buffer = make([]byte, int32(bufferVar))
		logger.Debugf("%s allocated get buffer of size[%d]", activityName, int32(bufferVar))
	} else {
		logger.Debugf("%s get buffer size not specified or 0, defaulting to 50000")
		buffer = make([]byte, int32(50000))
	}
	connectionObj := a.settings.Connection
	if connectionObj == nil {
		return nil, fmt.Errorf("%s internal error - no connection name specified in input", activityName)
	}

	connection, err := generic.NewConnection(connectionObj)
	if err != nil {
		return nil, fmt.Errorf("%s internal error - no connection settings in config", activityName)
	}

	qMgr, err := flogomq.GetQueueManager(activityName, connection, logger)
	if err != nil {
		return nil, err
	}
	defer flogomq.ReturnQueueManager(a.settings.Connection, activityName, qMgr, logger)

	a.queueName = a.settings.Queue
	if len(strings.TrimSpace(input.Queue)) != 0 {
		a.queueName = input.Queue
	}
	if len(a.queueName) == 0 {
		//which should be impossible because is required
		logger.Debugf("%s found no queue name in config", activityName)
		return nil, fmt.Errorf("%s found no queue name in config", activityName)
	}
	queue, err := qMgr.GetGetQueue(a.queueName, logger)
	if err != nil {
		return nil, err
	}

	bufLen = 0
	//if message size greater than maxSize given by user then message should not get truncated
	for trunc := true; trunc; {
		gmo := ibmmq.NewMQGMO()
		gmo.Options = ibmmq.MQGMO_NO_SYNCPOINT
		gmo.Options |= ibmmq.MQGMO_PROPERTIES_AS_Q_DEF

		if a.settings.GmoConvert == true {
			gmo.Options |= ibmmq.MQGMO_CONVERT
		}

		a.waitInterval = a.settings.WaitInterval
		if len(string(input.WaitInterval)) > 0 {
			a.waitInterval = input.WaitInterval
		}
		if a.waitInterval > 0 {
			gmo.Options |= ibmmq.MQGMO_WAIT
			gmo.WaitInterval = a.waitInterval
			logger.Debugf("%s setting wait interval to [%d]", activityName, gmo.WaitInterval)
		} else {
			gmo.Options |= ibmmq.MQGMO_NO_WAIT
			logger.Debugf("%s no wait interval specified", activityName)
		}

		getmqmd := ibmmq.NewMQMD()
		a.corellID = a.settings.CorellId
		if len(strings.TrimSpace(input.CorellId)) != 0 {
			a.corellID = input.CorellId
		}
		if len(a.corellID) > 0 {
			gmo.MatchOptions |= ibmmq.MQMO_MATCH_CORREL_ID
			getmqmd.CorrelId = flogomq.GetBytesFromString(a.corellID, 24, true)
			logger.Debugf("%s filtering on correlation ID [%x] with match optins [%d]", activityName, getmqmd.CorrelId, gmo.MatchOptions)
		}

		a.msgID = a.settings.MsgId
		if len(strings.TrimSpace(input.MsgId)) != 0 {
			a.msgID = input.MsgId
		}
		if len(a.msgID) > 0 {
			gmo.MatchOptions |= ibmmq.MQMO_MATCH_MSG_ID
			getmqmd.MsgId = flogomq.GetBytesFromString(a.msgID, 24, true)
			logger.Debugf("%s filtering on message ID [%x] with match optins [%d]", activityName, getmqmd.MsgId, gmo.MatchOptions)
		}
		cmho := ibmmq.NewMQCMHO()
		getMsgHandle, err := qMgr.Qmgr.CrtMH(cmho)
		if err != nil {
			return nil, fmt.Errorf("%s failed to allocate a get message handle, probably out of memory: [%s]", activityName, err)
		}
		defer getMsgHandle.DltMH(ibmmq.NewMQDMHO())

		gmo.MsgHandle = getMsgHandle
		start := time.Now()
		bufLen, err := queue.Get(getmqmd, gmo, buffer)
		if err != nil && err.(*ibmmq.MQReturn).MQCC != ibmmq.MQCC_WARNING {
			mqerr := err.(*ibmmq.MQReturn)
			if mqerr.MQRC != ibmmq.MQRC_NO_MSG_AVAILABLE {
				logger.Errorf("%s activity encountered error [%s]", activityName, err.(*ibmmq.MQReturn))
				trunc = false
			} else {
				elapsed := time.Since(start)
				logger.Debugf("%s timed out in %s", activityName, elapsed)
				trunc = false
			}
			return nil, err
		} else if err != nil && err.(*ibmmq.MQReturn).MQCC == ibmmq.MQCC_WARNING {
			mqreturn := err.(*ibmmq.MQReturn)
			if mqreturn.MQCC != ibmmq.MQCC_OK && mqreturn.MQRC == ibmmq.MQRC_TRUNCATED_MSG_FAILED && len(buffer) < maxBufSize {
				buffer = make([]byte, maxBufSize)
			} else {
				logger.Warnf("%s activity got this warning on the get api call [%s]", activityName, err)
				trunc = false
			}
		} else {
			trunc = false
		}
		if bufLen != 0 {
			logger.Debugf("%s activity got [%d] bytes from queue [%s] ", activityName, bufLen, queue.Name)
			msgbuff := make([]byte, bufLen)
			copy(msgbuff, buffer)
			outputData, err = qMgr.GetOutputData(activityName, a.settings.ValueType, getmqmd, gmo, msgbuff, logger)
		}
	}
	return
}
