package tibems

/*
#include <stdint.h>
#include <tibems/tibems.h>

#define MAX_STACK_FIELD_NAME_LENGTH 255
#define MAX_STACK_STRING_VALUE_LENGTH 16383

// These one-line stub functions are needed because some EMS pointer types
// are not really pointers, but object IDs, and so can have values that are not
// seen by Go as valid. Go's runtime pointer checking will panic if it sees what
// it believes is an invalid pointer. We use uintptr_t's in place
// of EMS pointer types where needed in the arguments to these stub functions
// so Go doesn't check their values, and then cast them back to the required
// EMS pointer types to call the original EMS API function.
// Not all EMS pointer types are really object IDs; some really are pointers,
// so a stub function is not always needed for a given EMS API function.

static tibems_status
go_tibemsMsg_GetDestination(
    tibemsMsg  message,
    uintptr_t* pDest)
{
  return tibemsMsg_GetDestination(message, (tibemsDestination*)pDest);
}

static tibems_status
go_tibemsMsg_SetDestination(
    tibemsMsg message,
    uintptr_t dest)
{
  return tibemsMsg_SetDestination(message, (tibemsDestination)dest);
}

static tibems_status
go_tibemsMsg_GetReplyTo(
    tibemsMsg  message,
    uintptr_t* pReply)
{
  return tibemsMsg_GetReplyTo(message, (tibemsDestination*)pReply);
}

static tibems_status
go_tibemsMsg_SetReplyTo(
    tibemsMsg message,
    uintptr_t reply)
{
  return tibemsMsg_SetReplyTo(message, (tibemsDestination)reply);
}

extern tibems_status
tibemsMsg_SetProperties(
    tibemsMsg          msg,
    tibems_uint        numFields,
    const char**       fieldNames,
    const tibems_int*  fieldNameLengths,
    const tibems_int*  fieldTypes,
    const int64_t*     fieldValues,
    const tibems_uint* fieldValueLengths)
{
  tibems_status status;
  tibems_uint   i;
  char          fieldName[MAX_STACK_FIELD_NAME_LENGTH + 1];

  for (i = 0; i < numFields; i++)
  {
    char* fieldNamePtr = &fieldName[0];
    if (fieldNameLengths[i] > MAX_STACK_FIELD_NAME_LENGTH)
    {
      fieldNamePtr = malloc(fieldNameLengths[i] + 1);
      if (fieldNamePtr == NULL)
      {
        return TIBEMS_NO_MEMORY;
      }
    }
    memcpy(fieldNamePtr, fieldNames[i], fieldNameLengths[i]);
    fieldNamePtr[fieldNameLengths[i]] = '\0';

    switch (fieldTypes[i])
    {
      case TIBEMS_BOOL:
        status = tibemsMsg_SetBooleanProperty(msg, fieldNamePtr, (tibems_bool)(fieldValues[i]));
        break;
      case TIBEMS_BYTE:
        status = tibemsMsg_SetByteProperty(msg, fieldNamePtr, (tibems_byte)(fieldValues[i]));
        break;
      case TIBEMS_SHORT:
        status = tibemsMsg_SetShortProperty(msg, fieldNamePtr, (tibems_short)(fieldValues[i]));
        break;
      case TIBEMS_INT:
        status = tibemsMsg_SetIntProperty(msg, fieldNamePtr, (tibems_int)(fieldValues[i]));
        break;
      case TIBEMS_LONG:
        status = tibemsMsg_SetLongProperty(msg, fieldNamePtr, (tibems_long)(fieldValues[i]));
        break;
      case TIBEMS_FLOAT:
        status = tibemsMsg_SetFloatProperty(msg, fieldNamePtr, (tibems_float)(fieldValues[i]));
        break;
      case TIBEMS_DOUBLE:
        status = tibemsMsg_SetDoubleProperty(msg, fieldNamePtr, (tibems_double)(fieldValues[i]));
        break;
      case TIBEMS_UTF8:
      {
        char stringValue[MAX_STACK_STRING_VALUE_LENGTH + 1], *stringValuePtr = &stringValue[0];
        if (fieldValueLengths[i] > MAX_STACK_STRING_VALUE_LENGTH)
        {
          stringValuePtr = malloc(fieldValueLengths[i] + 1);
          if (stringValuePtr == NULL)
          {
            return TIBEMS_NO_MEMORY;
          }
        }

        memcpy(stringValuePtr, (const void*)(fieldValues[i]), fieldValueLengths[i]);
        stringValuePtr[fieldValueLengths[i]] = '\0';
        status = tibemsMsg_SetStringProperty(msg, fieldNamePtr, stringValuePtr);

        if (stringValuePtr != &stringValue[0])
        {
          free(stringValuePtr);
        }
      }
      break;
      default:
        return TIBEMS_INVALID_FIELD;
    }

    if (fieldNamePtr != &fieldName[0])
    {
      free(fieldNamePtr);
    }

    if (status != TIBEMS_OK)
    {
      return status;
    }
  }

  return TIBEMS_OK;
}

extern tibems_status
tibemsMsg_GetPropertyType(
    tibemsMsg    msg,
    const char*  name,
    tibems_byte* type)
{
  tibemsMsgField field;
  memset(&field, 0, sizeof(tibemsMsgField));
  tibems_status status = tibemsMsg_GetProperty(msg, name, &field);
  if (status == TIBEMS_OK)
  {
    *type = field.type;
  }
  return status;
}

extern tibems_status
tibemsMsg_GetBytesProperty(
    tibemsMsg   msg,
    const char* name,
    char**      value,
    tibems_int* len)
{
  tibemsMsgField field;

  tibems_status status = tibemsMsg_GetProperty(msg, name, &field);
  if (status != TIBEMS_OK)
  {
    return status;
  }

  if (field.type != TIBEMS_BYTES)
  {
    return TIBEMS_CONVERSION_FAILED;
  }

  *value = (char*)(field.data.bytesValue);
  *len = field.size;

  return TIBEMS_OK;
}

*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"unsafe"
)

// Ensure Msg implements Message
var _ Message = (*Msg)(nil)

type Message interface {
	json.Marshaler

	GetBooleanProperty(name string) (bool, error)
	SetBooleanProperty(name string, value bool) error
	GetByteProperty(name string) (byte, error)
	SetByteProperty(name string, value byte) error
	GetShortProperty(name string) (int16, error)
	SetShortProperty(name string, value int16) error
	GetIntProperty(name string) (int32, error)
	SetIntProperty(name string, value int32) error
	GetLongProperty(name string) (int64, error)
	SetLongProperty(name string, value int64) error
	GetFloatProperty(name string) (float32, error)
	SetFloatProperty(name string, value float32) error
	GetDoubleProperty(name string) (float64, error)
	SetDoubleProperty(name string, value float64) error
	GetStringProperty(name string) (string, error)
	SetStringProperty(name string, value string) error

	PropertyExists(name string) (bool, error)
	GetPropertyType(name string) (FieldType, error)

	GetCorrelationID() (string, error)
	SetCorrelationID(value string) error

	GetDeliveryTime() (int64, error)

	GetDeliveryMode() (DeliveryMode, error)
	SetDeliveryMode(value DeliveryMode) error

	GetDestination() (*Destination, error)
	SetDestination(value *Destination) error

	GetExpiration() (int64, error)
	SetExpiration(value int64) error

	GetMessageID() (string, error)
	SetMessageID(value string) error

	GetPriority() (int32, error)
	SetPriority(value int32) error

	GetRedelivered() (bool, error)
	SetRedelivered(value bool) error

	SetReplyTo(value *Destination) error
	GetReplyTo() (*Destination, error)

	GetTimestamp() (int64, error)
	SetTimestamp(value int64) error

	GetType() (string, error)
	SetType(value string) error

	Print()
	fmt.Stringer

	GetBodyType() (MsgType, error)
	GetEncoding() (string, error)
	SetEncoding(string) error
	GetByteSize() (int32, error)
	GetAsBytes() ([]byte, error)

	CreateCopy() (Message, error)

	GetPropertyNames() (*MessageEnumeration, error)
	ClearProperties() error
	ClearBody() error
	MakeWriteable() error

	Acknowledge() error
	Close() error

	flushPending() error
	getCMessage() C.tibemsMsg
}

func (msg *Msg) Print() {
	C.tibemsMsg_Print(msg.cMessage)
}

func (msg *Msg) String() string {
	buf := make([]byte, 1024)
	status := C.tibemsMsg_PrintToBuffer(msg.cMessage, (*C.char)(unsafe.Pointer(&buf[0])), C.tibems_int(cap(buf)))
	for status == tibems_INSUFFICIENT_BUFFER {
		buf = make([]byte, cap(buf)*2)
		status = C.tibemsMsg_PrintToBuffer(msg.cMessage, (*C.char)(unsafe.Pointer(&buf[0])), C.tibems_int(cap(buf)))
	}
	if status != tibems_OK {
		return fmt.Sprintf("<error: %v>", getErrorFromStatus(status))
	}
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
}

func CreateMsgFromBytes(bytes []byte) (Message, error) {
	var cMessage C.tibemsMsg
	status := C.tibemsMsg_CreateFromBytes(&cMessage, unsafe.Pointer(&bytes[0]))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return instantiateSpecificMessageType(cMessage)
}

func (msg *Msg) MakeWriteable() error {
	status := C.tibemsMsg_MakeWriteable(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetAsBytes() ([]byte, error) {
	msg.flushPending()
	actualSize, err := msg.GetByteSize()
	if err != nil {
		return nil, err
	}
	bytes := make([]byte, int(actualSize))
	var cActualSize C.tibems_int
	status := C.tibemsMsg_GetAsBytesCopy(msg.cMessage, unsafe.Pointer(&bytes[0]), C.tibems_int(actualSize), &cActualSize)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return bytes, nil
}

func (msg *Msg) SetEncoding(encoding string) error {
	cValue := C.CString(encoding)
	defer C.free(unsafe.Pointer(cValue))

	status := C.tibemsMsg_SetEncoding(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) CreateCopy() (Message, error) {
	msg.flushPending()
	var cValue C.tibemsMsg
	status := C.tibemsMsg_CreateCopy(msg.cMessage, &cValue)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &Msg{
		cMessage: cValue,
	}, nil
}

func (msg *Msg) Acknowledge() error {
	status := C.tibemsMsg_Acknowledge(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetByteSize() (int32, error) {
	var cValue C.tibems_int
	status := C.tibemsMsg_GetByteSize(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return int32(cValue), nil
}

func (msg *Msg) GetEncoding() (string, error) {
	var cValue *C.char
	status := C.tibemsMsg_GetEncoding(msg.cMessage, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	return C.GoString(cValue), nil
}

func (msg *Msg) GetBodyType() (MsgType, error) {
	var cValue C.tibemsMsgType
	status := C.tibemsMsg_GetBodyType(msg.cMessage, &cValue)
	if status != tibems_OK {
		return MsgTypeUnknown, getErrorFromStatus(status)
	}
	return MsgType(cValue), nil
}

func (msg *Msg) GetDestination() (*Destination, error) {
	var cValue C.uintptr_t
	status := C.go_tibemsMsg_GetDestination(msg.cMessage, &cValue)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &Destination{cDestination: cValue}, nil
}

func (msg *Msg) SetDestination(destination *Destination) error {
	status := C.go_tibemsMsg_SetDestination(msg.cMessage, destination.cDestination)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetDeliveryMode() (DeliveryMode, error) {
	var cValue C.tibemsDeliveryMode
	status := C.tibemsMsg_GetDeliveryMode(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	value := DeliveryMode(cValue)
	return value, nil
}

func (msg *Msg) SetDeliveryMode(mode DeliveryMode) error {
	status := C.tibemsMsg_SetDeliveryMode(msg.cMessage, C.tibemsDeliveryMode(mode))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetDeliveryTime() (int64, error) {
	var cValue C.tibems_long
	status := C.tibemsMsg_GetDeliveryTime(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	value := int64(cValue)
	return value, nil
}

func (msg *Msg) GetType() (string, error) {
	var cValue *C.char
	status := C.tibemsMsg_GetType(msg.cMessage, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	value := C.GoString(cValue)
	return value, nil
}

func (msg *Msg) SetType(value string) error {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	status := C.tibemsMsg_SetType(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetRedelivered() (bool, error) {
	var cValue C.tibems_bool
	status := C.tibemsMsg_GetRedelivered(msg.cMessage, &cValue)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (msg *Msg) SetRedelivered(value bool) error {
	var cValue C.tibems_bool
	if value {
		cValue = 1
	} else {
		cValue = 0
	}
	status := C.tibemsMsg_SetRedelivered(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetCorrelationID() (string, error) {
	var cValue *C.char
	status := C.tibemsMsg_GetCorrelationID(msg.cMessage, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	value := C.GoString(cValue)
	return value, nil
}

func (msg *Msg) SetCorrelationID(value string) error {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	status := C.tibemsMsg_SetCorrelationID(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetMessageID() (string, error) {
	var cValue *C.char
	status := C.tibemsMsg_GetMessageID(msg.cMessage, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	value := C.GoString(cValue)
	return value, nil
}

func (msg *Msg) SetMessageID(value string) error {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	status := C.tibemsMsg_SetMessageID(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetTimestamp() (int64, error) {
	var cValue C.tibems_long
	status := C.tibemsMsg_GetTimestamp(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	value := int64(cValue)
	return value, nil
}

func (msg *Msg) SetTimestamp(timestamp int64) error {
	status := C.tibemsMsg_SetTimestamp(msg.cMessage, C.tibems_long(timestamp))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetPriority() (int32, error) {
	var cValue C.tibems_int
	status := C.tibemsMsg_GetPriority(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	value := int32(cValue)
	return value, nil
}

func (msg *Msg) SetPriority(priority int32) error {
	status := C.tibemsMsg_SetPriority(msg.cMessage, C.tibems_int(priority))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) GetExpiration() (int64, error) {
	var cValue C.tibems_long
	status := C.tibemsMsg_GetExpiration(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	value := int64(cValue)
	return value, nil
}

func (msg *Msg) SetExpiration(timestamp int64) error {
	status := C.tibemsMsg_SetExpiration(msg.cMessage, C.tibems_long(timestamp))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

type MessageEnumeration struct {
	cMsgEnum C.tibemsMsgEnum
}

func (msg *Msg) PropertyExists(name string) (bool, error) {
	err := msg.flushPending()
	if err != nil {
		return false, err
	}
	var cAnswer C.tibems_bool
	cPropName := C.CString(name)
	defer C.free(unsafe.Pointer(cPropName))
	status := C.tibemsMsg_PropertyExists(msg.cMessage, cPropName, &cAnswer)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cAnswer == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (msg *Msg) GetPropertyNames() (*MessageEnumeration, error) {
	err := msg.flushPending()
	if err != nil {
		return nil, err
	}
	var msgEnum = MessageEnumeration{cMsgEnum: nil}
	status := C.tibemsMsg_GetPropertyNames(msg.cMessage, &msgEnum.cMsgEnum)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &msgEnum, nil
}

func (msgEnum *MessageEnumeration) Close() error {
	status := C.tibemsMsgEnum_Destroy(msgEnum.cMsgEnum)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msgEnum *MessageEnumeration) GetNextName() (string, error) {
	var cValue *C.char
	status := C.tibemsMsgEnum_GetNextName(msgEnum.cMsgEnum, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	value := C.GoString(cValue)
	return value, nil
}

type Msg struct {
	cMessage C.tibemsMsg

	setBooleanProperties map[string]bool
	setByteProperties    map[string]byte
	setShortProperties   map[string]int16
	setIntProperties     map[string]int32
	setLongProperties    map[string]int64
	setFloatProperties   map[string]float32
	setDoubleProperties  map[string]float64
	setStringProperties  map[string]string

	pendingBooleanProperties map[string]bool
	pendingByteProperties    map[string]byte
	pendingShortProperties   map[string]int16
	pendingIntProperties     map[string]int32
	pendingLongProperties    map[string]int64
	pendingFloatProperties   map[string]float32
	pendingDoubleProperties  map[string]float64
	pendingStringProperties  map[string]string

	fieldNames        []*C.char
	fieldNameLengths  []int32
	fieldTypes        []FieldType
	fieldValues       []uintptr
	fieldValueLengths []uint32
}

type destinationHandle struct {
	DestinationName string `json:"name,omitempty"`
	DestinationType string `json:"type,omitempty"`
}

type messageHeader struct {
	MessageID     string             `json:"JMSMessageID,omitempty"`
	CorrelationID string             `json:"JMSCorrelationID,omitempty"`
	DeliveryMode  string             `json:"JMSDeliveryMode,omitempty"`
	DeliveryTime  int64              `json:"JMSDeliveryTime,omitempty"`
	Destination   *destinationHandle `json:"JMSDestination,omitempty"`
	Expiration    int64              `json:"JMSExpiration,omitempty"`
	Priority      int32              `json:"JMSPriority,omitempty"`
	Redelivered   bool               `json:"JMSRedelivered,omitempty"`
	ReplyTo       *destinationHandle `json:"JMSReplyTo,omitempty"`
	Timestamp     int64              `json:"JMSTimestamp,omitempty"`
	Type          string             `json:"JMSType,omitempty"`
}

func marshalProperties(message Message) ([]jsonMsgPropertyField, error) {
	properties := make([]jsonMsgPropertyField, 0)
	propNames, err := message.GetPropertyNames()
	if err != nil {
		return nil, err
	}
	defer func(propNames *MessageEnumeration) {
		err := propNames.Close()
		if err != nil {
			panic(err)
		}
	}(propNames)
	propName, _ := propNames.GetNextName()
	for propName != "" {
		fType, err := message.GetPropertyType(propName)
		if err != nil {
			return nil, err
		}
		switch fType {
		case FieldTypeNull:
		case FieldTypeBool:
			property, err := message.GetBooleanProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "bool",
				Value: property,
			})
		case FieldTypeByte:
			property, err := message.GetByteProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "byte",
				Value: property,
			})
		case FieldTypeShort:
			property, err := message.GetShortProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "short",
				Value: property,
			})
		case FieldTypeInt:
			property, err := message.GetIntProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "int",
				Value: property,
			})
		case FieldTypeLong:
			property, err := message.GetLongProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "long",
				Value: property,
			})
		case FieldTypeFloat:
			property, err := message.GetFloatProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "float",
				Value: property,
			})
		case FieldTypeDouble:
			property, err := message.GetDoubleProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "double",
				Value: property,
			})
		case FieldTypeUTF8:
			property, err := message.GetStringProperty(propName)
			if err != nil {
				return nil, err
			}
			properties = append(properties, jsonMsgPropertyField{
				Name:  propName,
				Type:  "string",
				Value: property,
			})
		default:
			log.Warn().Msgf("property '%s' type is unsupported; will not be used", propName)
		}
		propName, _ = propNames.GetNextName()
	}

	return properties, nil
}

func marshalHeader(message Message) (*messageHeader, error) {
	messageID, err := message.GetMessageID()
	if err != nil {
		return nil, err
	}
	correlationID, err := message.GetCorrelationID()
	if err != nil {
		return nil, err
	}
	delvMode, err := message.GetDeliveryMode()
	if err != nil {
		return nil, err
	}
	deliveryMode := ""
	switch delvMode {
	case DeliveryModeNonPersistent:
		deliveryMode = "NON_PERSISTENT"
	case DeliveryModePersistent:
		deliveryMode = "PERSISTENT"
	case DeliveryModeReliable:
		deliveryMode = "RELIABLE"
	}
	deliveryTime, err := message.GetDeliveryTime()
	if err != nil {
		return nil, err
	}
	destination, err := message.GetDestination()
	if err != nil {
		return nil, err
	}
	destinationType := ""
	destinationName := ""
	if destination != nil {
		destType, err := destination.GetType()
		if err == nil {
			switch destType {
			case DestTypeQueue:
				destinationType = "QUEUE"
			case DestTypeTopic:
				destinationType = "TOPIC"
			}
			destinationName, err = destination.GetName()
			if err != nil {
				return nil, err
			}
		} else if !errors.Is(err, ErrInvalid) {
			return nil, err
		}
	}
	expiration, err := message.GetExpiration()
	if err != nil {
		return nil, err
	}
	priority, err := message.GetPriority()
	if err != nil {
		return nil, err
	}
	redelivered, err := message.GetRedelivered()
	if err != nil {
		return nil, err
	}
	replyTo, err := message.GetReplyTo()
	if err != nil {
		return nil, err
	}
	var replyToDestination *destinationHandle
	if replyTo != nil {
		replyToName, err := replyTo.GetName()
		if err != nil {
			return nil, err
		}
		if replyToName != "" {
			rtType, err := replyTo.GetType()
			if err != nil {
				return nil, err
			}
			replyToType := ""
			switch rtType {
			case DestTypeQueue:
				replyToType = "QUEUE"
			case DestTypeTopic:
				replyToType = "TOPIC"
			}
			replyToDestination = &destinationHandle{
				DestinationName: replyToName,
				DestinationType: replyToType,
			}
		}
	}

	timestamp, err := message.GetTimestamp()
	if err != nil {
		return nil, err
	}
	messageType, err := message.GetType()
	if err != nil {
		return nil, err
	}

	var destHandle *destinationHandle
	if destinationName != "" {
		destHandle = &destinationHandle{
			DestinationName: destinationName,
			DestinationType: destinationType,
		}
	}

	return &messageHeader{
		MessageID:     messageID,
		CorrelationID: correlationID,
		DeliveryMode:  deliveryMode,
		DeliveryTime:  deliveryTime,
		Destination:   destHandle,
		Expiration:    expiration,
		Priority:      priority,
		Redelivered:   redelivered,
		ReplyTo:       replyToDestination,
		Timestamp:     timestamp,
		Type:          messageType,
	}, nil
}

type jsonMsgPropertyField struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type jsonMsg struct {
	Header     *messageHeader         `json:"header"`
	Properties []jsonMsgPropertyField `json:"properties"`
}

func (msg *Msg) MarshalJSON() ([]byte, error) {
	headers, err := marshalHeader(msg)
	if err != nil {
		return nil, err
	}
	properties, err := marshalProperties(msg)
	if err != nil {
		return nil, err
	}
	jsonMessage := jsonMsg{
		Header:     headers,
		Properties: properties,
	}
	return json.Marshal(jsonMessage)
}

func (msg *Msg) getCMessage() C.tibemsMsg {
	return msg.cMessage
}

func (msg *Msg) flushPending() error {

	numFields := len(msg.pendingBooleanProperties) +
		len(msg.pendingByteProperties) +
		len(msg.pendingShortProperties) +
		len(msg.pendingIntProperties) +
		len(msg.pendingLongProperties) +
		len(msg.pendingFloatProperties) +
		len(msg.pendingDoubleProperties) +
		len(msg.pendingStringProperties)

	if numFields == 0 {
		return nil
	}

	if len(msg.fieldNames) < numFields {
		msg.fieldNames = make([]*C.char, numFields)
		msg.fieldNameLengths = make([]int32, numFields)
		msg.fieldTypes = make([]FieldType, numFields)
		msg.fieldValues = make([]uintptr, numFields)
		msg.fieldValueLengths = make([]uint32, numFields)
	}

	fieldNum := 0

	for propName, value := range msg.pendingBooleanProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeBool
		if value {
			msg.fieldValues[fieldNum] = 1
		} else {
			msg.fieldValues[fieldNum] = 0
		}
		fieldNum++
		delete(msg.pendingBooleanProperties, propName)
		if msg.setBooleanProperties == nil {
			msg.setBooleanProperties = make(map[string]bool)
		}
		msg.setBooleanProperties[propName] = value
	}

	for propName, value := range msg.pendingByteProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeByte
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingByteProperties, propName)
		if msg.setByteProperties == nil {
			msg.setByteProperties = make(map[string]byte)
		}
		msg.setByteProperties[propName] = value
	}

	for propName, value := range msg.pendingShortProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeShort
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingShortProperties, propName)
		if msg.setShortProperties == nil {
			msg.setShortProperties = make(map[string]int16)
		}
		msg.setShortProperties[propName] = value
	}

	for propName, value := range msg.pendingIntProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeInt
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingIntProperties, propName)
		if msg.setIntProperties == nil {
			msg.setIntProperties = make(map[string]int32)
		}
		msg.setIntProperties[propName] = value
	}

	for propName, value := range msg.pendingLongProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeLong
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingLongProperties, propName)
		if msg.setLongProperties == nil {
			msg.setLongProperties = make(map[string]int64)
		}
		msg.setLongProperties[propName] = value
	}

	for propName, value := range msg.pendingFloatProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeFloat
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingFloatProperties, propName)
		if msg.setFloatProperties == nil {
			msg.setFloatProperties = make(map[string]float32)
		}
		msg.setFloatProperties[propName] = value
	}

	for propName, value := range msg.pendingDoubleProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeDouble
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingDoubleProperties, propName)
		if msg.setDoubleProperties == nil {
			msg.setDoubleProperties = make(map[string]float64)
		}
		msg.setDoubleProperties[propName] = value
	}

	for propName, value := range msg.pendingStringProperties {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeUTF8
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(C.CString(value)))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingStringProperties, propName)
		if msg.setStringProperties == nil {
			msg.setStringProperties = make(map[string]string)
		}
		msg.setStringProperties[propName] = value
	}

	status := C.tibemsMsg_SetProperties(msg.cMessage,
		C.tibems_uint(numFields),
		(**C.char)(unsafe.Pointer(&msg.fieldNames[0])),
		(*C.tibems_int)(unsafe.Pointer(&msg.fieldNameLengths[0])),
		(*C.tibems_int)(unsafe.Pointer(&msg.fieldTypes[0])),
		(*C.int64_t)(unsafe.Pointer(&msg.fieldValues[0])),
		(*C.tibems_uint)(unsafe.Pointer(&msg.fieldValueLengths[0])))

	for i := 0; i < numFields; i++ {
		C.free(unsafe.Pointer(msg.fieldNames[i]))
		if msg.fieldTypes[i] == FieldTypeUTF8 {
			C.free(unsafe.Pointer(msg.fieldValues[i]))
		}
	}

	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (msg *Msg) SetBooleanProperty(name string, value bool) error {
	if msg.pendingBooleanProperties == nil {
		msg.pendingBooleanProperties = make(map[string]bool)
	}
	msg.pendingBooleanProperties[name] = value
	return nil
}
func (msg *Msg) GetBooleanProperty(name string) (bool, error) {
	value, found := msg.pendingBooleanProperties[name]
	if !found {
		value, found = msg.setBooleanProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_bool
			status := C.tibemsMsg_GetBooleanProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return false, ErrNotFound
			}
			if status != tibems_OK {
				return false, getErrorFromStatus(status)
			}
			if cValue == 0 {
				value = false
			} else {
				value = true
			}
			if msg.setBooleanProperties == nil {
				msg.setBooleanProperties = make(map[string]bool)
			}
			msg.setBooleanProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *Msg) SetByteProperty(name string, value byte) error {
	if msg.pendingByteProperties == nil {
		msg.pendingByteProperties = make(map[string]byte)
	}
	msg.pendingByteProperties[name] = value
	return nil
}

func (msg *Msg) GetBytesProperty(name string) ([]byte, error) {
	cPropName := C.CString(name)
	defer C.free(unsafe.Pointer(cPropName))
	var cValue *C.char
	var cLen C.tibems_int
	status := C.tibemsMsg_GetBytesProperty(msg.cMessage, cPropName, &cValue, &cLen)
	if status == tibems_NOT_FOUND {
		return nil, ErrNotFound
	}
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return C.GoBytes(unsafe.Pointer(cValue), C.int(cLen)), nil
}

func (msg *Msg) GetByteProperty(name string) (byte, error) {
	value, found := msg.pendingByteProperties[name]
	if !found {
		value, found = msg.setByteProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_byte
			status := C.tibemsMsg_GetByteProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, getErrorFromStatus(status)
			}
			value = byte(cValue)
			if msg.setByteProperties == nil {
				msg.setByteProperties = make(map[string]byte)
			}
			msg.setByteProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *Msg) SetShortProperty(name string, value int16) error {
	if msg.pendingShortProperties == nil {
		msg.pendingShortProperties = make(map[string]int16)
	}
	msg.pendingShortProperties[name] = value
	return nil
}
func (msg *Msg) GetShortProperty(name string) (int16, error) {
	value, found := msg.pendingShortProperties[name]
	if !found {
		value, found = msg.setShortProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_short
			status := C.tibemsMsg_GetShortProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, getErrorFromStatus(status)
			}
			value = int16(cValue)
			if msg.setShortProperties == nil {
				msg.setShortProperties = make(map[string]int16)
			}
			msg.setShortProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *Msg) SetIntProperty(name string, value int32) error {
	if msg.pendingIntProperties == nil {
		msg.pendingIntProperties = make(map[string]int32)
	}
	msg.pendingIntProperties[name] = value
	return nil
}

func (msg *Msg) GetIntProperty(name string) (int32, error) {
	value, found := msg.pendingIntProperties[name]
	if !found {
		value, found = msg.setIntProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_int
			status := C.tibemsMsg_GetIntProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, getErrorFromStatus(status)
			}
			value = int32(cValue)
			if msg.setIntProperties == nil {
				msg.setIntProperties = make(map[string]int32)
			}
			msg.setIntProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *Msg) SetLongProperty(name string, value int64) error {
	if msg.pendingLongProperties == nil {
		msg.pendingLongProperties = make(map[string]int64)
	}
	msg.pendingLongProperties[name] = value
	return nil
}
func (msg *Msg) GetLongProperty(name string) (int64, error) {
	value, found := msg.pendingLongProperties[name]
	if !found {
		value, found = msg.setLongProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_long
			status := C.tibemsMsg_GetLongProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, getErrorFromStatus(status)
			}
			value = int64(cValue)
			if msg.setLongProperties == nil {
				msg.setLongProperties = make(map[string]int64)
			}
			msg.setLongProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *Msg) SetFloatProperty(name string, value float32) error {
	if msg.pendingFloatProperties == nil {
		msg.pendingFloatProperties = make(map[string]float32)
	}
	msg.pendingFloatProperties[name] = value
	return nil
}
func (msg *Msg) GetFloatProperty(name string) (float32, error) {
	value, found := msg.pendingFloatProperties[name]
	if !found {
		value, found = msg.setFloatProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_float
			status := C.tibemsMsg_GetFloatProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, getErrorFromStatus(status)
			}
			value = float32(cValue)
			if msg.setFloatProperties == nil {
				msg.setFloatProperties = make(map[string]float32)
			}
			msg.setFloatProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *Msg) SetDoubleProperty(name string, value float64) error {
	if msg.pendingDoubleProperties == nil {
		msg.pendingDoubleProperties = make(map[string]float64)
	}
	msg.pendingDoubleProperties[name] = value
	return nil
}
func (msg *Msg) GetDoubleProperty(name string) (float64, error) {
	value, found := msg.pendingDoubleProperties[name]
	if !found {
		value, found = msg.setDoubleProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_double
			status := C.tibemsMsg_GetDoubleProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, getErrorFromStatus(status)
			}
			value = float64(cValue)
			if msg.setDoubleProperties == nil {
				msg.setDoubleProperties = make(map[string]float64)
			}
			msg.setDoubleProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *Msg) SetStringProperty(name string, value string) error {
	if msg.pendingStringProperties == nil {
		msg.pendingStringProperties = make(map[string]string)
	}
	msg.pendingStringProperties[name] = value
	return nil
}
func (msg *Msg) GetStringProperty(name string) (string, error) {
	value, found := msg.pendingStringProperties[name]
	if !found {
		value, found = msg.setStringProperties[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.char
			status := C.tibemsMsg_GetStringProperty(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return "", ErrNotFound
			}
			if status != tibems_OK {
				return "", getErrorFromStatus(status)
			}
			value = C.GoString(cValue)
			if msg.setStringProperties == nil {
				msg.setStringProperties = make(map[string]string)
			}
			msg.setStringProperties[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

// CreateMsg creates a new, empty Msg.
//
// Applications MUST call [Msg.Close] when finished with a Msg to avoid resource leaks.
func CreateMsg() (*Msg, error) {
	var message = Msg{
		cMessage: nil,
	}
	status := C.tibemsMsg_Create(&message.cMessage)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &message, nil
}

// Close cleans up resources used by a Msg. A Msg MUST NOT be used again after Close has been called.
func (msg *Msg) Close() error {
	if msg == nil {
		return nil
	}
	msg.clearFieldMaps()
	status := C.tibemsMsg_Destroy(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (msg *Msg) SetReplyTo(reply *Destination) error {
	status := C.go_tibemsMsg_SetReplyTo(msg.cMessage, reply.cDestination)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (msg *Msg) GetReplyTo() (*Destination, error) {
	var cDestination C.uintptr_t
	status := C.go_tibemsMsg_GetReplyTo(msg.cMessage, &cDestination)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	if cDestination == 0 {
		return nil, nil
	}
	return &Destination{cDestination: cDestination}, nil
}

func (msg *Msg) ClearBody() error {
	status := C.tibemsMsg_ClearBody(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *Msg) clearFieldMaps() {
	msg.setBooleanProperties = nil
	msg.setByteProperties = nil
	msg.setShortProperties = nil
	msg.setIntProperties = nil
	msg.setLongProperties = nil
	msg.setFloatProperties = nil
	msg.setDoubleProperties = nil
	msg.setStringProperties = nil

	msg.pendingBooleanProperties = nil
	msg.pendingByteProperties = nil
	msg.pendingShortProperties = nil
	msg.pendingIntProperties = nil
	msg.pendingLongProperties = nil
	msg.pendingFloatProperties = nil
	msg.pendingDoubleProperties = nil
	msg.pendingStringProperties = nil
}

func (msg *Msg) ClearProperties() error {
	msg.clearFieldMaps()
	status := C.tibemsMsg_ClearProperties(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (msg *Msg) MarshalProperties(s any) error {
	_, err := marshalPropertiesInternal(msg, s)
	return err
}

func MarshalProperties(s any) (*Msg, error) {
	msg, err := CreateMsg()
	if err != nil {
		return nil, err
	}
	return msg, msg.MarshalProperties(s)
}

func marshalPropertiesInternal(msg *Msg, s any) (*Msg, error) {
	if reflect.TypeOf(s).Kind() == reflect.Struct {
		fields := reflect.VisibleFields(reflect.TypeOf(s))
		for i := 0; i < len(fields); i++ {
			field := fields[i]

			fieldName := field.Tag.Get("msg")
			if fieldName == "" {
				fieldName = field.Name
			}
			fieldValue := reflect.ValueOf(s).Field(i).Interface()
			err := marshalMsgPropertyField(msg, fieldName, fieldValue)
			if err != nil {
				return nil, err
			}
		}
		return msg, nil
	} else if reflect.TypeOf(s).Kind() == reflect.Map {
		fieldMap := s.(map[string]any)
		for fieldName, fieldValue := range fieldMap {
			err := marshalMsgPropertyField(msg, fieldName, fieldValue)
			if err != nil {
				return nil, err
			}
		}
	}

	return msg, nil
}

func (msg *Msg) GetPropertyType(name string) (FieldType, error) {
	cPropName := C.CString(name)
	defer C.free(unsafe.Pointer(cPropName))
	var fieldType C.tibems_byte
	status := C.tibemsMsg_GetPropertyType(msg.cMessage, cPropName, &fieldType)
	if status != tibems_OK {
		return FieldTypeNull, getErrorFromStatus(status)
	}
	return FieldType(fieldType), nil
}

func marshalMsgPropertyField(msg Message, fieldName string, fieldValue any) error {
	var err error
	kind := reflect.TypeOf(fieldValue).Kind()
	switch kind {
	case reflect.Int8:
		err = msg.SetByteProperty(fieldName, byte(fieldValue.(int8)))
	case reflect.Int16:
		err = msg.SetShortProperty(fieldName, fieldValue.(int16))
	case reflect.Int32:
		err = msg.SetIntProperty(fieldName, fieldValue.(int32))
	case reflect.Int64:
		err = msg.SetLongProperty(fieldName, fieldValue.(int64))
	case reflect.String:
		err = msg.SetStringProperty(fieldName, fieldValue.(string))
	case reflect.Bool:
		err = msg.SetBooleanProperty(fieldName, fieldValue.(bool))
	case reflect.Float32:
		err = msg.SetFloatProperty(fieldName, fieldValue.(float32))
	case reflect.Float64:
		err = msg.SetDoubleProperty(fieldName, fieldValue.(float64))
	default:
		return errors.New("struct field '" + fieldName + "' type is unsupported")
	}
	return err
}

func UnmarshalProperties(msg *Msg, s any) error {
	if reflect.TypeOf(s).Kind() != reflect.Pointer {
		return errors.New("Unmarshal requires a pointer to a struct")
	}

	fields := reflect.VisibleFields(reflect.TypeOf(s).Elem())
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		kind := fields[i].Type.Kind()
		fieldName := field.Tag.Get("msg")
		if fieldName == "" {
			fieldName = field.Name
		}
		fieldExists, err := msg.PropertyExists(fieldName)
		if err != nil {
			return err
		}
		if !fieldExists {
			continue
		}

		msgFieldType, err := msg.GetPropertyType(fieldName)
		if err != nil {
			return err
		}

		fieldValue := reflect.ValueOf(s).Elem().Field(i)
		switch kind {
		case reflect.Int8:
			if msgFieldType != FieldTypeByte {
				return errors.New("message field '" + fieldName + "' was not a byte; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetByteProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.Int16:
			if msgFieldType != FieldTypeShort {
				return errors.New("message field '" + fieldName + "' was not a short; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetShortProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.Int32:
			if msgFieldType != FieldTypeInt {
				return errors.New("message field '" + fieldName + "' was not an int; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetIntProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.Int64:
			if msgFieldType != FieldTypeLong {
				return errors.New("message field '" + fieldName + "' was not a long; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetLongProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.String:
			if msgFieldType != FieldTypeUTF8 {
				return errors.New("message field '" + fieldName + "' was not a string; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetStringProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.Bool:
			if msgFieldType != FieldTypeBool {
				return errors.New("message field '" + fieldName + "' was not a bool; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetBooleanProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.Float32:
			if msgFieldType != FieldTypeFloat {
				return errors.New("message field '" + fieldName + "' was not a float; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetFloatProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		case reflect.Float64:
			if msgFieldType != FieldTypeDouble {
				return errors.New("message field '" + fieldName + "' was not a double; cannot be assigned to struct field '" + field.Name + "'")
			}
			value, err := msg.GetDoubleProperty(fieldName)
			if err != nil {
				return err
			}
			fieldValue.Set(reflect.ValueOf(value))
		default:
			log.Debug().Msgf("struct field '%s' is not a supported type and will not be unmarshalled to", field.Name)
		}
	}

	return nil
}
