package flogomq

import (
	"bytes"
	"container/list"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ibm-messaging/mq-golang/v5/ibmmq"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
)

const (
	MQMT_REQUEST  = 1
	MQMT_REPLY    = 2
	MQMT_DATAGRAM = 8
	MQMT_REPORT   = 4

	ENCRYTIONMODE_NONE       = "None"
	ENCRYTIONMODE_SERVERAUTH = "TLS-ServerAuth"
	ENCRYTIONMODE_MUTUALAUTH = "TLS-MutualAuth"

	MAX_LENGTH_MESSAGE       = 10000
	DEFAULT_POLLING_INTERVAL = 1000

	CCDTURL_FILE    = "CCDT-File"
	CCDTURL_AUTH    = "CCDT-URL-Auth"
	CCDTURL_NONAUTH = "CCDT-URL-NonAuth"
)

type DestObjects struct {
	Object  ibmmq.MQObject
	Options int32
}

// FlogoMqm is the structure that carries all connection specific info
type FlogoMqm struct {
	Qmgr         ibmmq.MQQueueManager
	Ctlo         *ibmmq.MQCTLO
	PutDests     map[string]ibmmq.MQObject
	GetDests     map[string]ibmmq.MQObject
	PubDests     map[string]ibmmq.MQObject
	PubDestsList map[string]DestObjects
	PutDestsList map[string]DestObjects
	keystoreDir  string
	inUse        bool
	ccdtFileName string
}

var connectionPool = make(map[string]*list.List)
var qmgrMapMutex = &sync.Mutex{}
var bufLen int

const maxBufSize = 100 * 1024 * 1024 // 100 MB

func (t *FlogoMqm) GetOutputData(activityName string, valuetype string, getmqmd *ibmmq.MQMD, gmo *ibmmq.MQGMO, msgbuff []byte, logger log.Logger) (outputData map[string]interface{}, err error) {
	outputData = make(map[string]interface{})
	outputData["MessageProperties"] = t.getMessageProperties(activityName, gmo.MsgHandle, logger)
	outputData["MQMD"] = t.getMqmdProps(activityName, getmqmd, logger)
	outputString := make(map[string]interface{})
	if valuetype == "String" {
		outputString["String"] = string(msgbuff)
		outputData["Message"] = outputString
		logger.Debugf("%s activity got basic string message with content [%s]", activityName, headString(outputString["String"].(string), 128))
	} else {
		outputJSONMap := make(map[string]interface{})
		json.Unmarshal(msgbuff, &outputJSONMap)
		outputData["Message"] = outputJSONMap
		logger.Debugf("%s activity got complex message with content [%v]", activityName, outputJSONMap)
	}

	return
}
func (t *FlogoMqm) getMqmdProps(activityName string, mqmd *ibmmq.MQMD, logger log.Logger) map[string]interface{} {
	mqmdProps := make(map[string]interface{})
	switch mqmd.MsgType {
	case MQMT_REQUEST:
		mqmdProps["MsgType"] = "Request"
		break
	case MQMT_REPLY:
		mqmdProps["MsgType"] = "Reply"
		break
	case MQMT_DATAGRAM:
		mqmdProps["MsgType"] = "Datagram"
		break
	case MQMT_REPORT:
		mqmdProps["MsgType"] = "Report"
		break
	}
	mqmdProps["MsgId"] = base64.StdEncoding.EncodeToString(mqmd.MsgId)
	mqmdProps["CorrelId"] = base64.StdEncoding.EncodeToString(mqmd.CorrelId)
	mqmdProps["Encoding"] = int64(mqmd.Encoding)
	mqmdProps["CodedCharSetId"] = int64(mqmd.CodedCharSetId)
	mqmdProps["Format"] = mqmd.Format
	mqmdProps["Priority"] = int64(mqmd.Priority)
	mqmdProps["BackoutCount"] = int64(mqmd.BackoutCount)
	mqmdProps["ReplyToQ"] = mqmd.ReplyToQ
	mqmdProps["ReplyToQmgr"] = mqmd.ReplyToQMgr
	mqmdProps["UserIdentifier"] = mqmd.UserIdentifier
	mqmdProps["AccountingToken"] = base64.StdEncoding.EncodeToString(mqmd.AccountingToken)
	mqmdProps["ApplIdentityData"] = mqmd.ApplIdentityData
	mqmdProps["PutApplType"] = int64(mqmd.PutApplType)
	mqmdProps["PutApplName"] = mqmd.PutApplName
	mqmdProps["PutDate"] = mqmd.PutDate
	mqmdProps["PutTime"] = mqmd.PutTime
	mqmdProps["ApplOriginData"] = mqmd.ApplOriginData
	logger.Debugf("%s created MQMD output map: [%v]", activityName, mqmdProps)
	return mqmdProps
}
func (t *FlogoMqm) getMessageProperties(activityName string, msgHandle ibmmq.MQMessageHandle, logger log.Logger) map[string]interface{} {
	outputProps := make(map[string]interface{})
	impo := ibmmq.NewMQIMPO()
	pd := ibmmq.NewMQPD()

	// avoid leaking the properties handle
	defer msgHandle.DltMH(ibmmq.NewMQDMHO())

	impo.Options = ibmmq.MQIMPO_CONVERT_VALUE | ibmmq.MQIMPO_INQ_FIRST
	for propsToRead := true; propsToRead; {
		name, value, err := msgHandle.InqMP(impo, pd, "%")
		if err != nil {
			//got an error.. no more properties
			propsToRead = false
		} else {
			outputProps[name] = value
			logger.Debugf("%s getMessageProperties processed property [%s] with value [%s]", activityName, name, value)
		}
		impo.Options = ibmmq.MQIMPO_CONVERT_VALUE | ibmmq.MQIMPO_INQ_NEXT
	}
	return outputProps
}

// OpenTopicForPub Open a topic for publication
func (t *FlogoMqm) OpenTopicForPub(topicName string, topicDynamic string, activityName string, contextSupport string, logger log.Logger) (publicationObj ibmmq.MQObject, err error) {
	mqod := ibmmq.NewMQOD()
	topicOptions := ibmmq.MQOO_OUTPUT
	mqod.ObjectType = ibmmq.MQOT_TOPIC
	if len(topicName) > 0 {
		mqod.ObjectName = topicName
	}
	if len(topicDynamic) > 0 {
		mqod.ObjectString = topicDynamic
	}
	topic := mqod.ObjectName + "." + mqod.ObjectString
	if len(topic) == 0 {
		return publicationObj, fmt.Errorf("%s failed to open topic for pub because neither a predefined topic or dynamic topic name was provided on input", activityName)
	}
	// if topicObj, ok := t.PubDests[topic]; ok {
	// 	logger.Debugf("%s using cached topic object %s", activityName, topic)
	// 	return topicObj, nil
	// }
	contextSetting := contextSupport
	if contextSetting == "Identity" {
		topicOptions |= ibmmq.MQOO_SET_IDENTITY_CONTEXT
	} else if contextSetting == "All" {
		topicOptions |= ibmmq.MQOO_SET_ALL_CONTEXT
	}
	if topicObj, ok := t.PubDestsList[topic]; ok {
		if topicObj.Options == topicOptions {
			logger.Debugf("%s using cached topic object %s", activityName, topic)
			return topicObj.Object, nil
		}
	}
	publicationObj, err = t.Qmgr.Open(mqod, topicOptions)
	if err != nil {
		logger.Errorf("%s failed to open topic for pub [%s] err: %s", activityName, topic, err)
		return publicationObj, fmt.Errorf("%s failed to open topic for pub [%s - %s]",
			activityName, topicName, topicDynamic)
	}
	t.PubDests[topic] = publicationObj
	t.PubDestsList[topic] = DestObjects{
		Object:  publicationObj,
		Options: topicOptions,
	}
	logger.Debugf("%s opened topic for pub [%s] and cached the object", activityName, topic)
	return
}

func HaveMessagePropertiesInHandle(properties map[string]interface{}, logger log.Logger) bool {
	return properties != nil
}

func GetMessagePropertiesInHandle(properties map[string]interface{}, activityName string, flogoQMgr FlogoMqm, logger log.Logger) (ibmmq.MQMessageHandle, error) {
	cmho := ibmmq.NewMQCMHO()
	putMsgHandle, err := flogoQMgr.Qmgr.CrtMH(cmho)
	if !(reflect.TypeOf(properties).String() == "string") {
		props := properties
		if err != nil {
			logger.Debugf("%s getMessagePropertiesInHandle failed to get an mqrfh2 header for reason [%s]", activityName, err)
			return putMsgHandle, err
		}
		smpo := ibmmq.NewMQSMPO()
		pd := ibmmq.NewMQPD()
		for propname, propvalue := range props {
			//TODO deal with non string props should they be needed
			logger.Debugf("%s getMessagePropertiesInHandle adding property [%s] with value [%s]", activityName, propname, propvalue)
			err = putMsgHandle.SetMP(smpo, propname, pd, propvalue)
			if err != nil {
				logger.Debugf("%s getMessagePropertiesInHandle failed while adding property [%s] with value [%s] for reason [%s]", activityName, propname, propvalue.(string), err)
				return putMsgHandle, fmt.Errorf("%s getMessagePropertiesInHandle failed while adding property [%s] with value [%s] for reason [%s]", activityName, propname, propvalue.(string), err)
			}
		}
	}
	return putMsgHandle, nil
}

func GetMqmdFromContext(inMqmd interface{}, activityName string, msgType string, contextSetting string, pmo *ibmmq.MQPMO, connectionUsername string, logger log.Logger) (*ibmmq.MQMD, error) {
	putMqmd := ibmmq.NewMQMD()
	if !(reflect.TypeOf(inMqmd).String() == "string") {

		inMqmd := inMqmd.(map[string]interface{})

		if value := inMqmd["ReplyToQ"]; value != nil {
			putMqmd.ReplyToQ = headString(value.(string), 48)
			logger.Debugf("%s getMqmdFromContext set ReplyToQ [%s]", activityName, putMqmd.ReplyToQ)
		}
		if value := inMqmd["ReplyToQmgr"]; value != nil {
			putMqmd.ReplyToQMgr = headString(value.(string), 48)
			logger.Debugf("%s getMqmdFromContext set ReplyToQmgr [%s]", activityName, putMqmd.ReplyToQMgr)
		}
		//CorrelatonID's should be base64 encoded before mapping, but if not, use the string.
		if value := inMqmd["CorrelId"]; value != nil {
			putMqmd.CorrelId = GetBytesFromString(value.(string), 24, true)
			logger.Debugf("%s getMqmdFromContext set CorrelId [%x]", activityName, putMqmd.CorrelId)
		}
		if value := inMqmd["MsgId"]; value != nil {
			putMqmd.MsgId = GetBytesFromString(value.(string), 24, true)
			logger.Debugf("%s getMqmdFromContext set MsgId [%x]", activityName, putMqmd.MsgId)
		} else {
			pmo.Options |= ibmmq.MQPMO_NEW_MSG_ID
		}
		if value := inMqmd["Format"]; value != nil {
			putMqmd.Format = headString(value.(string), 8)
			logger.Debugf("%s getMqmdFromContext set Format [%s]", activityName, putMqmd.Format)
		}
		if value := inMqmd["Priority"]; value != nil {
			putMqmd.Priority = int32(value.(float64))
			logger.Debugf("%s getMqmdFromContext set Priority [%d]", activityName, putMqmd.Priority)
		}
		if value := inMqmd["Expiry"]; value != nil {
			putMqmd.Expiry = int32(value.(float64))
			logger.Debugf("%s getMqmdFromContext set Expiry [%d]", activityName, putMqmd.Expiry)
		}
		if value := inMqmd["Encoding"]; value != nil {
			putMqmd.Encoding = int32(value.(float64))
			logger.Debugf("%s getMqmdFromContext set Encoding [%d]", activityName, putMqmd.Encoding)
		}
		if value := inMqmd["CodedCharSetId"]; value != nil {
			putMqmd.CodedCharSetId = int32(value.(float64))
			logger.Debugf("%s getMqmdFromContext set CodedCharSetId [%d]", activityName, putMqmd.CodedCharSetId)
		}

		if contextSetting == "Default" {
			if contextSetting == "Default" {
				pmo.Options |= ibmmq.MQPMO_DEFAULT_CONTEXT
				logger.Debugf("%s getMqmdFromContext set pmo option MQPMO_DEFAULT_CONTEXT", activityName)
			}
		}

		if contextSetting == "Identity" || contextSetting == "All" {
			if contextSetting == "Identity" {
				pmo.Options |= ibmmq.MQPMO_SET_IDENTITY_CONTEXT
				logger.Debugf("%s getMqmdFromContext set pmo option MQPMO_SET_IDENTITY_CONTEXT", activityName)
			}
			if value := inMqmd["UserIdentifier"]; value != nil {
				putMqmd.UserIdentifier = headString(value.(string), 12)
				logger.Debugf("%s getMqmdFromContext set UserIdentifier [%s]", activityName, putMqmd.UserIdentifier)
			}
			if value := inMqmd["UserIdentifier"]; value == nil {
				putMqmd.UserIdentifier = connectionUsername
				logger.Debugf("%s getMqmdFromContext set UserIdentifier [%s]", activityName, putMqmd.UserIdentifier)
			}
			if value := inMqmd["AccountingToken"]; value != nil {
				putMqmd.AccountingToken = GetBytesFromString(value.(string), 32, true)
				logger.Debugf("%s getMqmdFromContext set AccountingToken [%s]", activityName, value.(string))
			}
			if value := inMqmd["ApplIdentityData"]; value != nil {
				putMqmd.ApplIdentityData = headString(value.(string), 32)
				logger.Debugf("%s getMqmdFromContext set ApplIdentityData [%s]", activityName, putMqmd.ApplIdentityData)
			}
		}
		if contextSetting == "All" {
			pmo.Options |= ibmmq.MQPMO_SET_ALL_CONTEXT
			logger.Debugf("%s getMqmdFromContext set pmo option MQPMO_SET_ALL_CONTEXT", activityName)
			if value := inMqmd["PutApplType"]; value != nil {
				putMqmd.PutApplType = int32(value.(float64))
				logger.Debugf("%s getMqmdFromContext set PutApplType [%d]", activityName, putMqmd.PutApplType)
			}
			if value := inMqmd["PutApplName"]; value != nil {
				putMqmd.PutApplName = headString(value.(string), 28)
				logger.Debugf("%s getMqmdFromContext set PutApplName [%s]", activityName, putMqmd.PutApplName)
			}
			if value := inMqmd["ApplOriginData"]; value != nil {
				putMqmd.ApplOriginData = headString(value.(string), 4)
				logger.Debugf("%s getMqmdFromContext set ApplOriginData [%s]", activityName, putMqmd.ApplOriginData)
			}
		}
	}

	switch msgType {
	case "Datagram":
		putMqmd.MsgType = MQMT_DATAGRAM
		break
	case "Request":
		putMqmd.MsgType = MQMT_REQUEST
		break
	case "Reply":
		putMqmd.MsgType = MQMT_REPLY
		break
	default:
		putMqmd.MsgType = MQMT_DATAGRAM
	}
	logger.Debugf("getMqmdFromContext set MsgType [%d]", putMqmd.MsgType)
	putMqmd.PutDateTime = time.Now().UTC()
	return putMqmd, nil

}

func headString(inStr string, length int) string {
	if len(inStr) > length {
		return inStr[:length]
	}
	return inStr

}
func GetBytesFromString(inStr string, length int, encoded bool) []byte {
	var decoded []byte
	var bytesOut = bytes.Repeat([]byte{0}, length)
	if encoded { // attempt to decode
		decodedBytes, err := base64.StdEncoding.DecodeString(inStr)
		if err != nil {
			decoded = []byte(inStr)
		} else {
			decoded = decodedBytes
		}
	} else {
		decoded = []byte(inStr)
	}
	for i := 0; i < length && i < len(decoded); i++ {
		bytesOut[i] = decoded[i]
	}
	return bytesOut
}

// GetGetQueue Get a queue object for input purposes
func (t *FlogoMqm) GetGetQueue(queueName string, logger log.Logger) (ibmmq.MQObject, error) {
	openOptions := ibmmq.MQOO_INPUT_AS_Q_DEF + ibmmq.MQOO_FAIL_IF_QUIESCING
	return t.getQueueGeneric(queueName, "", logger, openOptions, true)
}

// CloseDynamicQueue close the queue if it was dynamic.  specify delete
func (t *FlogoMqm) CloseDynamicQueue(queueName string, logger log.Logger) error {
	if queue, ok := t.PutDests[queueName]; ok && queueIsDynamic(queueName) {
		// todo not very dry
		logger.Debugf("closedynamicqueue about to close queue: [%s]", queueName)
		err := queue.Close(0) // ibmmq.MQCO_DELETE)
		if err != nil {
			logger.Debugf("closedynamicqueue for queue: [%s] failed for reason [%s]", queueName, err)
		}
		delete(t.PutDests, queueName)
		return err
	} else if queue, ok := t.GetDests[queueName]; ok && queueIsDynamic(queueName) {
		logger.Debugf("closedynamicqueue about to close queue: [%s]", queueName)
		err := queue.Close(0) // ibmmq.MQCO_DELETE)
		if err != nil {
			logger.Debugf("closedynamicqueue for queue: [%s] failed for reason [%s]", queueName, err)
		}
		delete(t.GetDests, queueName)
		return err
	}

	return nil
}

// CloseAllDests close the queue if it was dynamic.  specify delete
func (t *FlogoMqm) CloseAllDests(logger log.Logger) error {
	for destName, dest := range t.GetDests {
		logger.Debugf("%s closing dest [%s]", t.Qmgr.Name, destName)
		err := dest.Close(0) // ibmmq.MQCO_DELETE)
		if err != nil {
			logger.Warnf("%s closing dest [%s] failed for reason %s", t.Qmgr.Name, destName, err)
		}
		delete(t.GetDests, destName)
	}
	for destName, dest := range t.PutDests {
		logger.Debugf("%s closing dest [%s]", t.Qmgr.Name, destName)
		err := dest.Close(0) // ibmmq.MQCO_DELETE)
		if err != nil {
			logger.Warnf("%s closing dest [%s] failed for reason %s", t.Qmgr.Name, destName, err)
		}
		delete(t.PutDests, destName)
	}
	return nil
}

// Disconnect disconnect and close the queue manager.
func (t *FlogoMqm) Disconnect(logger log.Logger) {
	err := t.Qmgr.Disc()
	if err != nil {
		logger.Warnf("%s error closing for reason %s", t.Qmgr.Name, err)
	}
}

func queueIsDynamic(queueName string) bool {
	return strings.HasPrefix(queueName, "AMQ.")
}

// GetPutQueue get a queue for output purposes
func (t *FlogoMqm) GetPutQueue(queueName string, queueMgr string, contextSupport string, logger log.Logger) (ibmmq.MQObject, error) {
	openOptions := ibmmq.MQOO_OUTPUT + ibmmq.MQOO_FAIL_IF_QUIESCING
	contextSetting := contextSupport
	if contextSetting == "Identity" {
		openOptions |= ibmmq.MQOO_SET_IDENTITY_CONTEXT
	} else if contextSetting == "All" {
		openOptions |= ibmmq.MQOO_SET_ALL_CONTEXT
	}
	logger.Debugf("put open options options [%d]", openOptions)
	return t.getQueueGeneric(queueName, queueMgr, logger, openOptions, false)
}

func (t *FlogoMqm) getQueueGeneric(queue string, queueMgr string, logger log.Logger, openOptions int32, forGet bool) (ibmmq.MQObject, error) {
	// if forGet {
	// 	if queueObj, ok := t.GetDests[queue]; ok {
	// 		return queueObj, nil
	// 	}
	// } else {
	// 	if queueObj, ok := t.PutDests[queue]; ok {
	// 		return queueObj, nil
	// 	}
	// }
	mqod := ibmmq.NewMQOD()
	var qObject ibmmq.MQObject
	mqod.ObjectType = ibmmq.MQOT_Q
	mqod.ObjectName = queue
	mqod.ObjectQMgrName = queueMgr
	if forGet {
		qObject, ok := t.GetDests[mqod.ObjectName]
		if ok {
			logger.Debugf("reusing existing queue obj for [%s]", qObject.Name)
			return qObject, nil
		}
	} else {
		if qObject, ok := t.PutDestsList[mqod.ObjectName]; ok {
			if qObject.Options == openOptions {
				logger.Debugf("reusing existing queue obj for [%s]", mqod.ObjectName)
				return qObject.Object, nil
			}
		}
		// qObject, ok := t.PutDests[mqod.ObjectName]
		// if ok {
		// 	logger.Debugf("reusing existing queue obj for [%s]", qObject.Name)
		// 	return qObject, nil
		// }
	}
	qObject, err := t.Qmgr.Open(mqod, openOptions)
	if err != nil {
		logger.Debugf("getQueue [%s] failed for reason [%s]", mqod.ObjectName, err)
		return qObject, err
	}
	if forGet {
		t.GetDests[mqod.ObjectName] = qObject
	} else {
		t.PutDests[mqod.ObjectName] = qObject
		t.PutDestsList[mqod.ObjectName] = DestObjects{
			Object:  qObject,
			Options: openOptions,
		}
	}
	logger.Debugf("opened queue [%s] with open options [%d] ", qObject.Name, openOptions)
	return qObject, nil
}

// ReturnQueueManager mark the queue manager as available
func ReturnQueueManager(Connection map[string]interface{}, activityName string, flogoMqm FlogoMqm, logger log.Logger) {
	qmgrMapMutex.Lock()
	defer qmgrMapMutex.Unlock()
	conname, _ := getConnameFromConfig(Connection, activityName)
	qmList := connectionPool[conname]
	if qmList == nil {
		qmList = list.New()
		qmList.Init()
		connectionPool[conname] = qmList
	}
	qmList.PushBack(flogoMqm)
	logger.Debugf("%s pushed existing queue manager for connection name [%s]", activityName, conname)
}

func getQueueManagerFromPool(conname string, logger log.Logger) (FlogoMqm, bool) {
	var flogoMqm FlogoMqm
	qmgrMapMutex.Lock()
	defer qmgrMapMutex.Unlock()
	qmList := connectionPool[conname]
	if qmList == nil {
		qmList = list.New()
		qmList.Init()
		connectionPool[conname] = qmList
		return flogoMqm, false
	}
	if qmList.Len() == 0 {
		return flogoMqm, false
	}
	flogoMqmElem := qmList.Front()
	if flogoMqmElem == nil {
		return flogoMqm, false
	}
	flogoMqm = qmList.Remove(flogoMqmElem).(FlogoMqm)
	return flogoMqm, true
}

func getConnameFromConfig(Connection map[string]interface{}, activityName string) (string, error) {
	connectionObj := Connection
	if connectionObj == nil {
		return "", fmt.Errorf("%s internal error - no connection name specified in input", activityName)
	}

	connectionSettings := connectionObj["settings"].([]interface{})
	if connectionSettings == nil {
		return "", fmt.Errorf("%s internal error - no connection settings in config", activityName)
	}
	for _, val := range connectionSettings {
		setting := val.(map[string]interface{})
		v := setting["value"]
		name := setting["name"].(string)

		switch name {
		case "name":
			return v.(string), nil
		}
	}
	return "", fmt.Errorf("%s get connection name from config failed", activityName)
}

// GetQueueManager Get and cache a queue manager object.
func GetQueueManager(activityName string, connectionSettings *generic.Connection, logger log.Logger) (FlogoMqm, error) {
	var flogomqm FlogoMqm
	var conname, qmname, host, portString, chname, username, password, encryptionMode, certLabel, cipherspec, connectionType string
	var port float64
	var ccdtUrlType, ccdturl, ccdtUrlUsername, ccdtUrlPassword string
	var keystoreObj, stashfileObj, ccdtfileObj interface{}
	conname = connectionSettings.GetSetting("name").(string)
	qmname = connectionSettings.GetSetting("qmname").(string)
	username = connectionSettings.GetSetting("username").(string)
	password = connectionSettings.GetSetting("password").(string)
	host = connectionSettings.GetSetting("host").(string)
	chname = connectionSettings.GetSetting("chname").(string)
	port = connectionSettings.GetSetting("port").(float64)
	connectionType = connectionSettings.GetSetting("connectionType").(string)
	encryptionMode = connectionSettings.GetSetting("encryptionMode").(string)
	keystoreObj = connectionSettings.GetSetting("keystore").(interface{})
	stashfileObj = connectionSettings.GetSetting("keystorestash").(interface{})
	ccdtfileObj = connectionSettings.GetSetting("ccdtfile").(interface{})
	ccdtUrlType = connectionSettings.GetSetting("ccdtUrlType").(string)
	ccdturl = connectionSettings.GetSetting("ccdturl").(string)
	ccdtUrlUsername = connectionSettings.GetSetting("ccdtUrlUsername").(string)
	ccdtUrlPassword = connectionSettings.GetSetting("ccdtUrlPassword").(string)

	if encryptionMode != "None" {
		cipherspec = connectionSettings.GetSetting("cipherspec").(string)
		certLabel = connectionSettings.GetSetting("keystoreLabel").(string)
	}
	logger.Debugf("%s about to get queue manager handle for conname: %s", activityName, conname)
	if flogomqm, ok := getQueueManagerFromPool(conname, logger); ok {
		return flogomqm, nil
	}

	mqcno := ibmmq.NewMQCNO()
	mqcd := ibmmq.NewMQCD()
	mqcd.MaxMsgLength = 104857600 //100MB
	mqsco := ibmmq.NewMQSCO()
	mqcno.SSLConfig = mqsco
	mqcno.ApplName = "Tibco Flogo - IBM MQ"
	mqcno.Options = ibmmq.MQCNO_CLIENT_BINDING | ibmmq.MQCNO_HANDLE_SHARE_BLOCK | ibmmq.MQCNO_RECONNECT_Q_MGR | ibmmq.MQCNO_ALL_CONVS_SHARE
	if username != "" {
		csp := ibmmq.NewMQCSP()
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_USER_ID_AND_PWD
		csp.UserId = username
		csp.Password = password
		mqcno.SecurityParms = csp
	} else {
		csp := ibmmq.NewMQCSP()
		csp.AuthenticationType = ibmmq.MQCSP_AUTH_NONE
		mqcno.SecurityParms = csp
	}

	if connectionType == "Remote" {
		mqcd.ChannelName = chname
		portString = fmt.Sprintf("%f", port)
		mqcd.ConnectionName = host + "(" + portString + ")"
		mqcno.ClientConn = mqcd
	} else if connectionType == "CCDT" {
		if ccdtUrlType == CCDTURL_FILE {
			if ccdtfileObj != "" {
				ccdtFileTemp, err := createTempCcdtFile(activityName, ccdtfileObj, qmname, logger)
				if err != nil {
					return flogomqm, err
				}
				mqcno.CCDTUrl = ccdtFileTemp
				defer func(tempFile string) {
					os.RemoveAll(tempFile)
					logger.Debugf("%s getQueueManager cleaned up file [%s]", activityName, tempFile)
				}(ccdtFileTemp)
			}
		} else if ccdtUrlType == CCDTURL_AUTH {
			if ccdturl != "" {
				if ccdtUrlUsername != "" && ccdtUrlPassword != "" {
					var urlAuthField string
					ccdturlSplit := strings.SplitAfter(ccdturl, "://")
					urlAuthField = ccdtUrlUsername + ":" + ccdtUrlPassword + "@"
					var ccdtAuthUrl = ccdturlSplit[0] + urlAuthField + ccdturlSplit[1]
					mqcno.CCDTUrl = ccdtAuthUrl
				}
			}
		} else if ccdtUrlType == CCDTURL_NONAUTH {
			if ccdturl != "" {
				mqcno.CCDTUrl = ccdturl
			}
		} else {
			logger.Infof("Please provide valid CCDT file/url")
		}
	}

	if encryptionMode != ENCRYTIONMODE_NONE {
		mqcd.SSLCipherSpec = cipherspec
		repo, err := createTempKeystoreFile(activityName, keystoreObj, stashfileObj, qmname, logger)

		if err != nil {
			return flogomqm, err
		}
		mqsco.KeyRepository = repo + string(os.PathSeparator) + qmname

		if encryptionMode == ENCRYTIONMODE_SERVERAUTH {
			mqcd.SSLClientAuth = ibmmq.MQSCA_OPTIONAL
		} else if encryptionMode == ENCRYTIONMODE_MUTUALAUTH {
			mqsco.CertificateLabel = certLabel
			mqcd.SSLClientAuth = ibmmq.MQSCA_REQUIRED
		}
		defer func(dir string) {
			os.RemoveAll(dir)
			logger.Debugf("%s getQueueManager cleaned up folder [%s]", activityName, dir)

		}(repo)

	}

	qMgr, err := ibmmq.Connx(qmname, mqcno)
	if err == nil || (err != nil && err.(*ibmmq.MQReturn).MQCC == ibmmq.MQCC_WARNING) {
		if err != nil && err.(*ibmmq.MQReturn).MQCC == ibmmq.MQCC_WARNING {
			logger.Warnf("%s got warning while opening an MQ connection [%s]", activityName, err)
		}
		logger.Debugf("getQueueManager created a new queue manager for MQ connection [%s]", conname)
		flogomqm.Qmgr = qMgr
		flogomqm.PutDests = make(map[string]ibmmq.MQObject)
		flogomqm.GetDests = make(map[string]ibmmq.MQObject)
		flogomqm.PubDests = make(map[string]ibmmq.MQObject)
		flogomqm.PubDestsList = make(map[string]DestObjects)
		flogomqm.PutDestsList = make(map[string]DestObjects)
		flogomqm.inUse = true
		flogomqm.Ctlo = ibmmq.NewMQCTLO()
		return flogomqm, nil
	}
	logger.Debugf("%s getQueueManager %s failed for reason [%s]", activityName, conname, err)
	return flogomqm, fmt.Errorf("%s failed to create queue manager [%s] for MQ connection [%s] for reason [%s", activityName, qmname, conname, err)
}

func createTempKeystoreFile(activityName string, keystoreObj interface{}, stashfileObj interface{}, queuemgrname string, logger log.Logger) (keystoreDir string, err error) {
	keystoreBytes, err := getBytesFromFileSetting(activityName, keystoreObj, logger)
	if err != nil {
		return
	}
	stashfileBytes, err := getBytesFromFileSetting(activityName, stashfileObj, logger)
	if err != nil {
		return
	}
	keystoreDir, err = ioutil.TempDir(os.TempDir(), queuemgrname)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(keystoreDir+string(os.PathSeparator)+queuemgrname+".kdb", keystoreBytes, 0644)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(keystoreDir+string(os.PathSeparator)+queuemgrname+".sth", stashfileBytes, 0644)
	if err != nil {
		return
	}
	logger.Debugf("%s createTempKeystoreFile created a keystore of len [%d] and stash of len [%d] at [%s]", activityName, len(keystoreBytes), len(stashfileBytes), keystoreDir)
	return
}
func createTempCcdtFile(activityName string, ccdtfileObj interface{}, queuemgrname string, logger log.Logger) (ccdtFileName string, err error) {
	ccdtFileBytes, err := getBytesFromFileSetting(activityName, ccdtfileObj, logger)
	if err != nil {
		return
	}
	var ccdtFileDir string
	ccdtFileDir, err = ioutil.TempDir(os.TempDir(), queuemgrname)
	if err != nil {
		return
	}
	ccdtFileName = ccdtFileDir + string(os.PathSeparator) + queuemgrname
	err = ioutil.WriteFile(ccdtFileName, ccdtFileBytes, 0644)
	if err != nil {
		return
	}
	logger.Debugf("%s createTempCcdtFile created a ccdt file of len [%d] at [%s]", activityName, len(ccdtFileBytes), ccdtFileName)
	return
}
func getBytesFromFileSetting(activityName string, cert interface{}, logger log.Logger) ([]byte, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate contains is nil")
	}
	if reflect.TypeOf(cert).String() == "map[string]interface {}" {
		logger.Debug("IBM-MQ configured file selector")
		cacert := cert.(map[string]interface{})
		var header = "base64,"
		value := cacert["content"].(string)
		if value == "" {
			return nil, fmt.Errorf("%s file based setting contains no data", activityName)
		}
		if strings.Index(value, header) >= 0 {
			value = value[strings.Index(value, header)+len(header):]
			decodedLen := base64.StdEncoding.DecodedLen(len(value))
			destArray := make([]byte, decodedLen)
			actualen, err := base64.StdEncoding.Decode(destArray, []byte(value))
			if err != nil {
				return nil, fmt.Errorf("%s file based setting not base64 encoded: [%s]", activityName, err)
			}
			if decodedLen != actualen {
				newDestArray := make([]byte, actualen)
				copy(newDestArray, destArray)
				return newDestArray, nil
			}
			return destArray, nil
		}
	}
	return nil, fmt.Errorf("%s internal error; file based setting not formatted correctly", activityName)
}
