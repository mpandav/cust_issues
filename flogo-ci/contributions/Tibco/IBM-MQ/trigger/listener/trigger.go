package listener

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"github.com/tibco/wi-contrib/connection/generic"

	flogomq "github.com/tibco/wi-ibmmq/src/app/IBM-MQ"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

// var log = logger.GetLogger("flogo-ibmmq-listen")
var versionPrinted = false

// MqTriggerFactory My Trigger factory
// type MqTriggerFactory struct {
// 	metadata *trigger.Metadata
// }

//NewFactory create a new Trigger factory
// func NewFactory(md *trigger.Metadata) trigger.Factory {
// 	return &MqTriggerFactory{metadata: md}
// }

//New Creates a new trigger instance for a given id
// func (t *MqTriggerFactory) New(config *trigger.Config) trigger.Trigger {
// 	return &MqTrigger{metadata: t.metadata, config: config}
// }

type connectionError struct {
	name  string
	error *ibmmq.MQReturn
}

func init() {
	_ = trigger.Register(&MqTrigger{}, &Factory{})
}

// MqListener structure containing all handles for a running instance of a messge handler
type MqListener struct {
	activityName string
	connection   *generic.Connection
	destName     string
	valueType    string
	flogoMqm     flogomq.FlogoMqm
	dest         ibmmq.MQObject
	handler      trigger.Handler
	mqcbd        *ibmmq.MQCBD
	mqmd         *ibmmq.MQMD
	mqgmo        *ibmmq.MQGMO
	mqcmho       *ibmmq.MQCMHO
	mqsd         *ibmmq.MQSD
	inrecovery   bool
}

// MqTrigger is a stub for your Trigger implementation
type MqTrigger struct {
	ctx             trigger.InitContext
	Tmetadata       *trigger.Metadata
	Tconfig         *trigger.Config
	Log             log.Logger
	ctlo            *ibmmq.MQCTLO
	running         bool
	pollinginterval int32
	clientconfirm   bool
	handlers        map[string]*MqListener
	connError       chan *connectionError
	name            string
}

type Factory struct {
}

func (t *Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New Creates a new trigger instance for a given id
func (t *Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &MqTrigger{name: config.Id, Tconfig: config}, nil
}

// The callback crashes when used with msg properties.  Its either user error, or a bug, but we will go for our own
// function to listen for messages.
func (t *MqTrigger) msgCallback(mqm *ibmmq.MQQueueManager, queueObj *ibmmq.MQObject, md *ibmmq.MQMD, gmo *ibmmq.MQGMO, msgbuff []byte, cbc *ibmmq.MQCBC, err *ibmmq.MQReturn) {
	var handler *MqListener
	handlerFound := false

	if err != nil && err.MQCC != ibmmq.MQCC_OK {
		if err.MQCC == ibmmq.MQCC_WARNING {
			t.Log.Warnf("%s listener got warning in callback for queuemgr [%s] warning [%s]", t.Tconfig.Id, mqm.Name, err)
			return
		}
		if err.MQRC != ibmmq.MQRC_NO_MSG_AVAILABLE {
			t.Log.Errorf("%s listener got error in callback for queuemgr [%s] queue [%s] warning [%s]", t.Tconfig.Id, mqm.Name, queueObj.Name, err)
			if err.MQRC ==
				ibmmq.MQRC_CONNECTION_QUIESCING || err.MQRC == ibmmq.MQRC_CONNECTION_BROKEN || err.MQRC == ibmmq.MQRC_CONNECTION_STOPPED || err.MQRC == ibmmq.MQRC_CONNECTION_STOPPING || err.MQRC == ibmmq.MQRC_CONNECTION_ERROR || err.MQRC == ibmmq.MQRC_CONNECTION_NOT_AVAILABLE || err.MQRC == ibmmq.MQRC_CONNECTION_QUIESCING || err.MQRC == ibmmq.MQRC_Q_MGR_STOPPING || err.MQRC == ibmmq.MQRC_Q_MGR_QUIESCING {
				t.Log.Errorf("%s listener queuemgr [%s] connection recovery dispatched", t.Tconfig.Id, mqm.Name)
				connectionErr := new(connectionError)
				connectionErr.error = err
				connectionErr.name = mqm.Name
				t.connError <- connectionErr
			}
			return
		}
		t.Log.Debugf("%s on queuemgr [%s] listener timeout", t.Tconfig.Id, mqm.Name)
		return

	}
	for key, handlerObj := range t.handlers {
		t.Log.Debugf("attempting to match handler [%s] queueMgr [%s] and queue [%s] for dest addr [%p] and queueObj addr [%p]", key, mqm.Name, queueObj.Name, &handlerObj.dest, queueObj)
		if *queueObj == handlerObj.dest {
			handler = handlerObj
			handlerFound = true
			break
		}
	}
	if !handlerFound {
		t.Log.Errorf("%s failed dereference queue handler for queueMgr [%s] and queue  [%s]", t.Tconfig.Id, mqm.Name, queueObj.Name)
		panic("Failed to dereference handler for message.  Fatal error!")
	}
	t.Log.Debugf("%s listener starting flow with message from queue [%s] of length [%d] ", t.Tconfig.Id, queueObj.Name, len(msgbuff))
	outputData, errGetData := t.getOutputData(handler, md, gmo, msgbuff)
	if errGetData != nil {
		t.Log.Errorf("%s listener failed to create output for message [%x] for reason [%]; starting backout", t.Tconfig.Id, errGetData, md.MsgId)
		errBack := mqm.Back()
		if errBack != nil {
			t.Log.Errorf("%s listener failed to back out message [%x] for reason [%s] MESSAGE LOST!", t.Tconfig.Id, md.MsgId, errBack)
		} else {
			t.Log.Errorf("%s listener backed out message [%x] ", t.Tconfig.Id, md.MsgId)
		}
	}
	defer gmo.MsgHandle.DltMH(ibmmq.NewMQDMHO())

	attr, flowError := handler.handler.Handle(context.Background(), outputData)
	if t.clientconfirm {
		if flowError == nil {
			if attr["confirm"] != nil && attr["confirm"] == true {
				errcommit := mqm.Cmit()
				if errcommit != nil {
					t.Log.Errorf("%s listener failed to commit a unit of work on behalf of message [%x] for reason [%s]", t.Tconfig.Id, md.MsgId, errcommit)
					mqm.Back()
				}
				t.Log.Debugf("%s listener commited a unit of work on behalf of message [%x]", t.Tconfig.Id, md.MsgId)
			} else {
				errback := mqm.Back()
				if errback != nil {
					t.Log.Errorf("%s listener failed and backout a unit of work on behalf of message [%x] for reason [%s]", t.Tconfig.Id, md.MsgId, errback)
				}
				t.Log.Warnf("%s message is not confirmed and message [%x] backed out; flow err: [%s]", t.Tconfig.Id, md.MsgId, flowError)
			}
		} else {
			errback := mqm.Back()
			if errback != nil {
				t.Log.Errorf("%s listener failed and backout a unit of work on behalf of message [%x] for reason [%s]", t.Tconfig.Id, md.MsgId, errback)
			}
			t.Log.Debugf("%s flow failed and message [%x] backed out; flow err: %s", t.Tconfig.Id, md.MsgId, flowError)

		}
	}
}

func (t *MqTrigger) getOutputData(listener *MqListener, getmqmd *ibmmq.MQMD, gmo *ibmmq.MQGMO, msgbuff []byte) (outputData map[string]interface{}, err error) {
	outputData = make(map[string]interface{})
	outputData["MessageProperties"] = t.getMessageProperties(gmo.MsgHandle)
	outputData["MQMD"] = t.getMqmdProps(getmqmd)
	outputString := make(map[string]interface{})
	if listener.valueType == "String" {
		outputString["String"] = string(msgbuff)
		outputData["Message"] = outputString
	} else {
		outputJSONMap := make(map[string]interface{})
		json.Unmarshal(msgbuff, &outputJSONMap)
		outputData["MessageJson"] = outputJSONMap
	}
	t.Log.Debugf("%s trigger output: %v", t.Tconfig.Id, outputData)
	return
}
func (t *MqTrigger) getMqmdProps(mqmd *ibmmq.MQMD) map[string]interface{} {
	mqmdProps := make(map[string]interface{})
	switch mqmd.MsgType {
	case flogomq.MQMT_REQUEST:
		mqmdProps["MsgType"] = "Request"
		break
	case flogomq.MQMT_REPLY:
		mqmdProps["MsgType"] = "Reply"
		break
	case flogomq.MQMT_DATAGRAM:
		mqmdProps["MsgType"] = "Datagram"
		break
	case flogomq.MQMT_REPORT:
		mqmdProps["MsgType"] = "Report"
		break
	}
	mqmdProps["MsgId"] = base64.StdEncoding.EncodeToString(mqmd.MsgId)
	mqmdProps["CorrelId"] = base64.StdEncoding.EncodeToString(mqmd.CorrelId)
	mqmdProps["Encoding"] = float64(mqmd.Encoding)
	mqmdProps["CodedCharSetId"] = float64(mqmd.CodedCharSetId)
	mqmdProps["Format"] = mqmd.Format
	mqmdProps["Priority"] = float64(mqmd.Priority)
	mqmdProps["BackoutCount"] = float64(mqmd.BackoutCount)
	mqmdProps["ReplyToQ"] = mqmd.ReplyToQ
	mqmdProps["ReplyToQmgr"] = mqmd.ReplyToQMgr
	mqmdProps["UserIdentifier"] = mqmd.UserIdentifier
	mqmdProps["AccountingToken"] = base64.StdEncoding.EncodeToString(mqmd.AccountingToken)
	mqmdProps["ApplIdentityData"] = mqmd.ApplIdentityData
	mqmdProps["PutApplType"] = float64(mqmd.PutApplType)
	mqmdProps["PutApplName"] = mqmd.PutApplName
	mqmdProps["PutDate"] = mqmd.PutDate
	mqmdProps["PutTime"] = mqmd.PutTime
	mqmdProps["ApplOriginData"] = mqmd.ApplOriginData

	return mqmdProps
}
func (t *MqTrigger) getMessageProperties(msgHandle ibmmq.MQMessageHandle) map[string]interface{} {
	outputProps := make(map[string]interface{})
	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()
	impo.Options = ibmmq.MQIMPO_CONVERT_VALUE | ibmmq.MQIMPO_INQ_FIRST
	for propsToRead := true; propsToRead; {
		name, value, err := msgHandle.InqMP(impo, pd, "%")
		if err != nil {
			//got an error.. no more properties
			propsToRead = false
		} else {
			outputProps[name] = value
			t.Log.Debugf("%s getMessageProperties processed property [%s] with value [%s]", t.Tconfig.Id, name, value)
		}
		impo.Options = ibmmq.MQIMPO_CONVERT_VALUE | ibmmq.MQIMPO_INQ_NEXT
	}
	return outputProps
}

func (t *MqTrigger) recoverHandler(handlerID string, handler *MqListener) {
	handler.flogoMqm.CloseAllDests(t.Log)
	handler.flogoMqm.Disconnect(t.Log)
	delete(t.handlers, handlerID)
	for true {
		handlerID, err := t.InitHandler(handler)
		if err != nil {
			t.Log.Infof("trigger handler [%s] in recovery and failed to rconnect for reason [%s]", handlerID, err)
			time.Sleep(time.Duration(t.pollinginterval) * time.Millisecond)
		} else {
			t.Log.Infof("trigger handler [%s] in recovery about to resume connection", handlerID)
			err := handler.flogoMqm.Qmgr.Ctl(ibmmq.MQOP_START, handler.flogoMqm.Ctlo)
			if err != nil {
				handler.flogoMqm.CloseAllDests(t.Log)
				handler.flogoMqm.Disconnect(t.Log)
				t.Log.Infof("trigger handler [%s] in recovery and failed to restart for reason [%s]", handlerID, err)
				time.Sleep(time.Duration(t.pollinginterval) * time.Millisecond)
			} else {
				t.handlers[handlerID] = handler
				t.Log.Infof("trigger handler [%s] recovery successful", handlerID)
				return
			}
		}
	}

}
func (t *MqTrigger) monitor() {
	for conErr := range t.connError {
		t.Log.Infof("monitor got connerror name: %s Err: %s ", conErr.name, conErr.error)
		for key, handler := range t.handlers {
			if !handler.inrecovery && conErr.name == handler.flogoMqm.Qmgr.Name {
				handler.inrecovery = true
				go t.recoverHandler(key, handler)
			}
		}
	}
}

// InitHandler Initialize the handler from context.  Used in recovery to recreate connections and dests
func (t *MqTrigger) InitHandler(listener *MqListener) (listenerID string, err error) {
	t.Log.Debugf("%s listener [%s] opening connection", t.Tconfig.Id, listenerID)
	if _, ok := t.handlers[listenerID]; ok {
		t.Log.Errorf("%s trigger initialization for handler [%s] skipped because it is a duplicate qmgr.queue", t.Tconfig.Id, listenerID)
		return "", nil
	}

	mqm, err := flogomq.GetQueueManager(t.Tconfig.Id, listener.connection, t.Log)
	if err != nil {
		return "", fmt.Errorf("%s trigger failed to initialize for reason [%s]", t.Tconfig.Id, err)
	}
	listener.flogoMqm = mqm
	listener.dest, err = listener.flogoMqm.GetGetQueue(listener.destName, t.Log)
	if err != nil {
		return "", fmt.Errorf("%s trigger initialization failed while opening queue [%s] for reason [%s]", t.Tconfig.Id, listener.destName, err)
	}
	t.Log.Debugf("%s listener [%s] opened queue [%s] as objectid [%p]", t.Tconfig.Id, listenerID, listener.destName, &listener.dest)

	listenerID = fmt.Sprintf("%s.%p", listener.connection.GetSetting("qmname").(string), &listener.destName)

	listener.mqcmho = ibmmq.NewMQCMHO()
	mh, err := listener.flogoMqm.Qmgr.CrtMH(listener.mqcmho)
	if err != nil {
		return "", fmt.Errorf("%s trigger initialization failed while creating a message handle for queue [%s] for reason [%s]", t.Tconfig.Id, listener.destName, err)
	}

	// Register the callback
	listener.mqmd = ibmmq.NewMQMD()
	listener.mqgmo = ibmmq.NewMQGMO()
	if t.clientconfirm == false {
		listener.mqgmo.Options = ibmmq.MQGMO_NO_SYNCPOINT
	} else {
		listener.mqgmo.Options = ibmmq.MQGMO_SYNCPOINT
	}
	listener.mqgmo.Options |= ibmmq.MQGMO_WAIT
	listener.mqgmo.WaitInterval = t.pollinginterval
	listener.mqgmo.Options |= ibmmq.MQGMO_PROPERTIES_AS_Q_DEF
	listener.mqgmo.Options |= ibmmq.MQGMO_PROPERTIES_IN_HANDLE
	listener.mqgmo.MsgHandle = mh

	listener.mqcbd = ibmmq.NewMQCBD()
	listener.mqcbd.CallbackFunction = t.msgCallback
	listener.inrecovery = false
	err = listener.dest.CB(ibmmq.MQOP_REGISTER, listener.mqcbd, listener.mqmd, listener.mqgmo)
	if err != nil {
		return "", fmt.Errorf("%s listener [%s] failed to set message callback for reason: %s", t.Tconfig.Id, listenerID, err)
	}
	t.Log.Debugf("%s listener [%s] set message callback with queue manager [%p] and queue [%p]", t.Tconfig.Id, listenerID, &listener.flogoMqm.Qmgr, &listener.dest)
	return listenerID, nil
}

// Initialize implements trigger.Trigger.Init
func (t *MqTrigger) Initialize(ctx trigger.InitContext) (err error) {
	t.Log = ctx.Logger()
	if !versionPrinted {
		flogomq.PrintVersion(t.Log)
	}
	t.ctx = ctx
	t.clientconfirm = t.Tconfig.Settings["clientconfirm"].(bool)
	t.pollinginterval = t.Tconfig.Settings["pollinginterval"].(int32)
	t.handlers = make(map[string]*MqListener)
	t.connError = make(chan *connectionError)
	go t.monitor()
	for _, handler := range ctx.GetHandlers() {
		listener := new(MqListener)
		listener.activityName = t.Tconfig.Id
		listener.handler = handler
		handlerSetting := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)

		if handlerSetting == nil {
			return fmt.Errorf("%s trigger failed to access connection in config", t.Tconfig.Id)
		}
		listener.connection, err = generic.NewConnection(handlerSetting.Connection)
		if err != nil {
			return fmt.Errorf("%s trigger initialization internal error - no connection settings in config", t.Tconfig.Id)
		}
		if len(handlerSetting.Queue) > 0 {
			listener.destName = handlerSetting.Queue
		} else {
			return fmt.Errorf("%s trigger initialization failed queue not specified", t.Tconfig.Id)
		}

		if len(handlerSetting.ValueType) > 0 {
			listener.valueType = handlerSetting.ValueType
		} else {
			return fmt.Errorf("%s trigger initialization failed valueType not specified", t.Tconfig.Id)
		}

		listenerID, err := t.InitHandler(listener)
		if err != nil {
			return err
		}
		t.handlers[listenerID] = listener
		t.Log.Debugf("%s added handler [%s]", t.Tconfig.Id, listenerID)

		t.Log.Debugf("%s initialized handler for connection [%s]", t.Tconfig.Id, handler.Name())
	}
	return nil
}

// Start implements trigger.Trigger.Start
func (t *MqTrigger) Start() error {
	var errStr string
	for handlerName, handler := range t.handlers {
		// go t.listenWorker(handler)
		t.Log.Debugf("%s.%s trigger about to start connection", t.Tconfig.Id, handlerName)
		err := handler.flogoMqm.Qmgr.Ctl(ibmmq.MQOP_START, handler.flogoMqm.Ctlo)
		if err != nil {
			t.Log.Errorf("%s.%s trigger start failed for reason [%s]", t.Tconfig.Id, handlerName, err)
			errStr = fmt.Sprintf("%s.%s trigger start failed for reason [%s]", t.Tconfig.Id, handlerName, err)
		} else {
			t.Log.Debugf("%s.%s trigger handler started", t.Tconfig.Id, handlerName)
		}
	}
	if errStr != "" {
		return fmt.Errorf("%s", errStr)
	}
	t.running = true
	t.Log.Debugf("%s trigger started", t.Tconfig.Id)
	return nil
}

// Stop implements trigger.Trigger.Start
func (t *MqTrigger) Stop() error {
	for handlerName, handler := range t.handlers {
		t.Log.Debugf("%s.%s trigger about to stop connection", t.Tconfig.Id, handlerName)
		err := handler.flogoMqm.Qmgr.Ctl(ibmmq.MQOP_STOP, handler.flogoMqm.Ctlo)
		if err != nil {
			t.Log.Errorf("%s.%s trigger handler stop failed for reason [%s]", t.Tconfig.Id, handlerName, err)
		}
	}
	t.running = false
	t.Log.Debugf("%s trigger stopped", t.Tconfig.Id)
	return nil
}
