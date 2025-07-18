package subscriber

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

// MqSubscriber structure containing all handles for a running instance of a messge handler
type MqSubscriber struct {
	activityName string
	connection   *generic.Connection
	destName     string
	dynamictopic string
	durable      bool
	durablename  string
	newpubsonly  bool
	valueType    string
	flogoMqm     flogomq.FlogoMqm
	dest         ibmmq.MQObject
	subscription ibmmq.MQObject
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
	handlers        map[string]*MqSubscriber
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
	var handler *MqSubscriber
	handlerFound := false

	if err != nil && err.MQCC != ibmmq.MQCC_OK {
		if err.MQCC == ibmmq.MQCC_WARNING {
			t.Log.Warnf("%s subscriber got warning in callback for queuemgr [%s] warning [%s]", t.Tconfig.Id, mqm.Name, err)
			return
		}
		if err.MQRC != ibmmq.MQRC_NO_MSG_AVAILABLE {
			t.Log.Errorf("%s subscriber got error in callback for queuemgr [%s] queue [%s] warning [%s]", t.Tconfig.Id, mqm.Name, queueObj.Name, err)
			if err.MQRC ==
				ibmmq.MQRC_CONNECTION_QUIESCING || err.MQRC == ibmmq.MQRC_CONNECTION_BROKEN || err.MQRC == ibmmq.MQRC_CONNECTION_STOPPED || err.MQRC == ibmmq.MQRC_CONNECTION_STOPPING || err.MQRC == ibmmq.MQRC_CONNECTION_ERROR || err.MQRC == ibmmq.MQRC_CONNECTION_NOT_AVAILABLE || err.MQRC == ibmmq.MQRC_CONNECTION_QUIESCING || err.MQRC == ibmmq.MQRC_Q_MGR_STOPPING || err.MQRC == ibmmq.MQRC_Q_MGR_QUIESCING {
				t.Log.Errorf("%s subscriber queuemgr [%s] connection recovery dispatched", t.Tconfig.Id, mqm.Name)
				connectionErr := new(connectionError)
				connectionErr.error = err
				connectionErr.name = mqm.Name
				t.connError <- connectionErr
			}
			return
		}
		t.Log.Debugf("%s on queuemgr [%s] subscriber timeout", t.Tconfig.Id, mqm.Name)
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
	t.Log.Debugf("%s subscriber starting flow with message from queue [%s] of length [%d] ", t.Tconfig.Id, queueObj.Name, len(msgbuff))
	outputData, errGetData := t.getOutputData(handler, md, gmo, msgbuff)
	if errGetData != nil {
		t.Log.Errorf("%s subscriber failed to create output for message [%x] for reason [%]; starting backout", t.Tconfig.Id, errGetData, md.MsgId)
		errBack := mqm.Back()
		if errBack != nil {
			t.Log.Errorf("%s subscriber failed to back out message [%x] for reason [%s] MESSAGE LOST!", t.Tconfig.Id, md.MsgId, errBack)
		} else {
			t.Log.Errorf("%s subscriber backed out message [%x] ", t.Tconfig.Id, md.MsgId)
		}
	}
	defer gmo.MsgHandle.DltMH(ibmmq.NewMQDMHO())

	attr, flowError := handler.handler.Handle(context.Background(), outputData)
	if t.clientconfirm {
		if flowError == nil {
			if attr["confirm"] != nil && attr["confirm"] == true {
				errcommit := mqm.Cmit()
				if errcommit != nil {
					t.Log.Errorf("%s subscriber failed to commit a unit of work on behalf of message [%x] for reason [%s]", t.Tconfig.Id, md.MsgId, errcommit)
					mqm.Back()
				}
				t.Log.Debugf("%s subscriber commited a unit of work on behalf of message [%x]", t.Tconfig.Id, md.MsgId)
			} else {
				errback := mqm.Back()
				if errback != nil {
					t.Log.Errorf("%s subscriber failed and backout a unit of work on behalf of message [%x] for reason [%s]", t.Tconfig.Id, md.MsgId, errback)
				}
				t.Log.Warnf("%s message is not confirmed and message [%x] backed out; flow err: [%s]", t.Tconfig.Id, md.MsgId, flowError)
			}
		} else {
			errback := mqm.Back()
			if errback != nil {
				t.Log.Errorf("%s subscriber failed and backout a unit of work on behalf of message [%x] for reason [%s]", t.Tconfig.Id, md.MsgId, errback)
			}
			t.Log.Debugf("%s flow failed and message [%x] backed out; flow err: %s", t.Tconfig.Id, md.MsgId, flowError)

		}
	}
}

func (t *MqTrigger) getOutputData(subscriber *MqSubscriber, getmqmd *ibmmq.MQMD, gmo *ibmmq.MQGMO, msgbuff []byte) (outputData map[string]interface{}, err error) {
	outputData = make(map[string]interface{})
	outputData["MessageProperties"] = t.getMessageProperties(gmo.MsgHandle)
	outputData["MQMD"] = t.getMqmdProps(getmqmd)
	outputString := make(map[string]interface{})
	if subscriber.valueType == "String" {
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

func (t *MqTrigger) recoverHandler(handlerID string, handler *MqSubscriber) {
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
func (t *MqTrigger) openTopicForSub(subscriber *MqSubscriber) error {
	subscriber.mqsd = ibmmq.NewMQSD()
	subscriber.mqsd.Options = ibmmq.MQSO_CREATE | ibmmq.MQSO_MANAGED
	subscriber.mqsd.ObjectString = subscriber.dynamictopic
	subscriber.mqsd.ObjectName = subscriber.destName

	if subscriber.newpubsonly {
		subscriber.mqsd.Options |= ibmmq.MQSO_NEW_PUBLICATIONS_ONLY
	}
	if subscriber.durable {
		subscriber.mqsd.Options |= ibmmq.MQSO_DURABLE | ibmmq.MQSO_RESUME
		subscriber.mqsd.SubName = subscriber.durablename
		subscriptionObj, err := subscriber.flogoMqm.Qmgr.Sub(subscriber.mqsd, &subscriber.dest)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("%s failed to open durable subscription [%s]",
				t.Tconfig.Id, subscriber.durablename)
		}
		t.Log.Debugf("%s opened durable subscription [%s]", t.Tconfig.Id, subscriber.durablename)
		subscriber.subscription = subscriptionObj
	} else {
		subscriber.mqsd.Options |= ibmmq.MQSO_NON_DURABLE
		subscriptionObj, err := subscriber.flogoMqm.Qmgr.Sub(subscriber.mqsd, &subscriber.dest)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("%s failed to open non-durable subscription [%s/%s]",
				t.Tconfig.Id, subscriber.destName, subscriber.dynamictopic)
		}
		t.Log.Debugf("%s opened non-durable subscription [%s/%s]",
			t.Tconfig.Id, subscriber.destName, subscriber.dynamictopic)
		subscriber.subscription = subscriptionObj
	}
	if subscriber.durable {
		subscriber.subscription.Close(0)
	}
	return nil
}

// InitHandler Initialize the handler from context.  Used in recovery to recreate connections and dests
func (t *MqTrigger) InitHandler(subscriber *MqSubscriber) (subscriberID string, err error) {
	t.Log.Debugf("%s subscriber [%s] opening connection", t.Tconfig.Id, subscriberID)
	if _, ok := t.handlers[subscriberID]; ok {
		t.Log.Errorf("%s trigger initialization for handler [%s] skipped because it is a duplicate qmgr.queue", t.Tconfig.Id, subscriberID)
		return "", nil
	}

	mqm, err := flogomq.GetQueueManager(t.Tconfig.Id, subscriber.connection, t.Log)
	if err != nil {
		return "", fmt.Errorf("%s trigger failed to initialize for reason [%s]", t.Tconfig.Id, err)
	}
	subscriber.flogoMqm = mqm
	err = t.openTopicForSub(subscriber)
	if err != nil {
		return "", fmt.Errorf("%s trigger initialization failed for reason [%s]", t.Tconfig.Id, err)
	}
	t.Log.Debugf("%s subscriber opened topic [%s/%s] as objectid [%p]", t.Tconfig.Id, subscriber.destName, subscriber.dynamictopic, &subscriber.dest)

	subscriberID = fmt.Sprintf("%s.%p", subscriber.connection.GetSetting("qmname").(string), &subscriber.destName)

	subscriber.mqcmho = ibmmq.NewMQCMHO()
	mh, err := subscriber.flogoMqm.Qmgr.CrtMH(subscriber.mqcmho)
	if err != nil {
		return "", fmt.Errorf("%s trigger initialization failed while creating a message handle for queue [%s] for reason [%s]", t.Tconfig.Id, subscriber.destName, err)
	}

	// Register the callback
	subscriber.mqmd = ibmmq.NewMQMD()
	subscriber.mqgmo = ibmmq.NewMQGMO()
	if t.clientconfirm == false {
		subscriber.mqgmo.Options = ibmmq.MQGMO_NO_SYNCPOINT
	} else {
		subscriber.mqgmo.Options = ibmmq.MQGMO_SYNCPOINT
	}
	subscriber.mqgmo.Options |= ibmmq.MQGMO_WAIT
	subscriber.mqgmo.WaitInterval = t.pollinginterval
	subscriber.mqgmo.Options |= ibmmq.MQGMO_PROPERTIES_AS_Q_DEF
	subscriber.mqgmo.Options |= ibmmq.MQGMO_PROPERTIES_IN_HANDLE
	subscriber.mqgmo.MsgHandle = mh

	subscriber.mqcbd = ibmmq.NewMQCBD()
	subscriber.mqcbd.CallbackFunction = t.msgCallback
	subscriber.inrecovery = false
	err = subscriber.dest.CB(ibmmq.MQOP_REGISTER, subscriber.mqcbd, subscriber.mqmd, subscriber.mqgmo)
	if err != nil {
		return "", fmt.Errorf("%s subscriber [%s] failed to set message callback for reason: %s", t.Tconfig.Id, subscriberID, err)
	}
	t.Log.Debugf("%s subscriber [%s] set message callback with queue manager [%p] and queue [%p]", t.Tconfig.Id, subscriberID, &subscriber.flogoMqm.Qmgr, &subscriber.dest)
	return subscriberID, nil
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
	t.handlers = make(map[string]*MqSubscriber)
	t.connError = make(chan *connectionError)
	go t.monitor()
	for _, handler := range ctx.GetHandlers() {
		subscriber := new(MqSubscriber)
		subscriber.activityName = t.Tconfig.Id
		subscriber.handler = handler
		connectionObj := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), connectionObj, true)

		if connectionObj == nil {
			return fmt.Errorf("%s trigger failed to access connection in config", t.Tconfig.Id)
		}
		subscriber.connection, err = generic.NewConnection(connectionObj.Connection)
		if err != nil {
			return fmt.Errorf("%s trigger initialization internal error - no connection settings in config", t.Tconfig.Id)
		}
		subscriber.durable = false
		if connectionObj.Durable == true {
			subscriber.durable = connectionObj.Durable
		}
		subscriber.newpubsonly = false
		if connectionObj.Newpubsonly == true {
			subscriber.newpubsonly = connectionObj.Newpubsonly
		}
		if connectionObj.Durable == true {
			if len(connectionObj.Durablename) > 0 {
				subscriber.durablename = connectionObj.Durablename
			}
		}
		if len(connectionObj.Topic) > 0 {
			subscriber.destName = connectionObj.Topic
		}
		if len(connectionObj.Dynamictopic) > 0 {
			subscriber.dynamictopic = connectionObj.Dynamictopic
		}

		if len(connectionObj.ValueType) > 0 {
			subscriber.valueType = connectionObj.ValueType
		} else {
			return fmt.Errorf("%s trigger initialization failed valueType not specified", t.Tconfig.Id)
		}

		subscriberID, err := t.InitHandler(subscriber)
		if err != nil {
			return err
		}
		t.handlers[subscriberID] = subscriber
		t.Log.Debugf("%s added handler [%s]", t.Tconfig.Id, subscriberID)

		t.Log.Debugf("%s initialized handler for connection [%s]", t.Tconfig.Id, handler.Name())
	}
	return nil
}

// Metadata implements trigger.Trigger.Metadata
func (t *MqTrigger) Metadata() *trigger.Metadata {
	return t.Tmetadata
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
