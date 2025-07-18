package tibems

/*
#include <stdint.h>
#include <tibems/tibems.h>

#define MAX_STACK_FIELD_NAME_LENGTH 255
#define MAX_STACK_STRING_VALUE_LENGTH 16383

extern tibems_status
tibemsMapMsg_SetFields(
        tibemsMsg			msg,
        tibems_uint  		numFields,
        const char 			**fieldNames,
        const tibems_int	*fieldNameLengths,
        const tibems_int 	*fieldTypes,
        const int64_t		*fieldValues,
        const tibems_uint   *fieldValueLengths)
{
    tibems_status status;
    tibems_uint i;
    char fieldName[MAX_STACK_FIELD_NAME_LENGTH + 1];

    for (i = 0; i < numFields; i++) {
        char *fieldNamePtr = &fieldName[0];
        if (fieldNameLengths[i] > MAX_STACK_FIELD_NAME_LENGTH) {
            fieldNamePtr = malloc(fieldNameLengths[i] + 1);
            if (fieldNamePtr == NULL) {
                return TIBEMS_NO_MEMORY;
            }
        }
        memcpy(fieldNamePtr, fieldNames[i], fieldNameLengths[i]);
        fieldNamePtr[fieldNameLengths[i]] = '\0';

 		switch (fieldTypes[i]) {
            case TIBEMS_BOOL:
                status = tibemsMapMsg_SetBoolean(msg, fieldNamePtr, (tibems_bool)(fieldValues[i]));
                break;
            case TIBEMS_BYTE:
                status = tibemsMapMsg_SetByte(msg, fieldNamePtr, (tibems_byte)(fieldValues[i]));
                break;
            case TIBEMS_SHORT:
                status = tibemsMapMsg_SetShort(msg, fieldNamePtr, (tibems_short)(fieldValues[i]));
                break;
            case TIBEMS_INT:
                status = tibemsMapMsg_SetInt(msg, fieldNamePtr, (tibems_int)(fieldValues[i]));
                break;
            case TIBEMS_LONG:
                status = tibemsMapMsg_SetLong(msg, fieldNamePtr, (tibems_long)(fieldValues[i]));
                break;
            case TIBEMS_FLOAT:
                status = tibemsMapMsg_SetFloat(msg, fieldNamePtr, (tibems_float)(fieldValues[i]));
                break;
            case TIBEMS_DOUBLE:
                status = tibemsMapMsg_SetDouble(msg, fieldNamePtr, (tibems_double)(fieldValues[i]));
                break;
            case TIBEMS_UTF8: {
                char stringValue[MAX_STACK_STRING_VALUE_LENGTH + 1], *stringValuePtr = &stringValue[0];
                if (fieldValueLengths[i] > MAX_STACK_STRING_VALUE_LENGTH) {
                    stringValuePtr = malloc(fieldValueLengths[i] + 1);
                    if (stringValuePtr == NULL) {
                        return TIBEMS_NO_MEMORY;
                    }
                }

                memcpy(stringValuePtr, (const void *) (fieldValues[i]), fieldValueLengths[i]);
                stringValuePtr[fieldValueLengths[i]] = '\0';
                status = tibemsMapMsg_SetString(msg, fieldNamePtr, stringValuePtr);

                if (stringValuePtr != &stringValue[0]) {
                    free(stringValuePtr);
                }
            }
                break;
            case TIBEMS_BYTES:
                status = tibemsMapMsg_SetBytes(msg, fieldNamePtr, (void*)(fieldValues[i]), fieldValueLengths[i]);
                break;
            case TIBEMS_MAP_MSG:
                status = tibemsMapMsg_SetMapMsg(msg, fieldNamePtr, (tibemsMapMsg)(fieldValues[i]), 0);
                break;
            case TIBEMS_SHORT_ARRAY:
                status = tibemsMapMsg_SetShortArray(msg, fieldNamePtr, (tibems_short *)(fieldValues[i]), fieldValueLengths[i]);
                break;
            case TIBEMS_INT_ARRAY:
                status = tibemsMapMsg_SetIntArray(msg, fieldNamePtr, (tibems_int *)(fieldValues[i]), fieldValueLengths[i]);
                break;
            case TIBEMS_LONG_ARRAY:
                status = tibemsMapMsg_SetLongArray(msg, fieldNamePtr, (tibems_long *)(fieldValues[i]), fieldValueLengths[i]);
                break;
            case TIBEMS_FLOAT_ARRAY:
                status = tibemsMapMsg_SetFloatArray(msg, fieldNamePtr, (tibems_float *)(fieldValues[i]), fieldValueLengths[i]);
                break;
            case TIBEMS_DOUBLE_ARRAY:
                status = tibemsMapMsg_SetDoubleArray(msg, fieldNamePtr, (tibems_double *)(fieldValues[i]), fieldValueLengths[i]);
                break;
            default:
                return TIBEMS_INVALID_FIELD;
        }

        if (fieldNamePtr != &fieldName[0]) {
            free(fieldNamePtr);
        }

        if (status != TIBEMS_OK) {
            return status;
        }
    }

    return TIBEMS_OK;
}

extern tibems_status
tibemsMapMsg_GetShortArray(
    tibemsMsg           	msg,
    const char*         	name,
    const tibems_short**	array,
    tibems_uint*        	arrayCount)
{
	tibemsMsgField field;
	memset(&field, 0, sizeof(tibemsMsgField));
	tibems_status status = tibemsMapMsg_GetField(msg, name, &field);
	if (status == TIBEMS_OK) {
		if (field.type != TIBEMS_SHORT_ARRAY) {
			return TIBEMS_INVALID_TYPE;
		}
		*array = (tibems_short*)field.data.arrayValue;
		*arrayCount = (tibems_uint)field.count;
	}

	return status;
}

extern tibems_status
tibemsMapMsg_GetIntArray(
    tibemsMsg           msg,
    const char*         name,
    const tibems_int**	array,
    tibems_uint*        arrayCount)
{
	tibemsMsgField field;
	memset(&field, 0, sizeof(tibemsMsgField));
	tibems_status status = tibemsMapMsg_GetField(msg, name, &field);
	if (status == TIBEMS_OK) {
		if (field.type != TIBEMS_INT_ARRAY) {
			return TIBEMS_INVALID_TYPE;
		}
		*array = (tibems_int*)field.data.arrayValue;
		*arrayCount = (tibems_uint)field.count;
	}

	return status;
}

extern tibems_status
tibemsMapMsg_GetLongArray(
    tibemsMsg           msg,
    const char*         name,
    const tibems_long**	array,
    tibems_uint*        arrayCount)
{
	tibemsMsgField field;
	memset(&field, 0, sizeof(tibemsMsgField));
	tibems_status status = tibemsMapMsg_GetField(msg, name, &field);
	if (status == TIBEMS_OK) {
		if (field.type != TIBEMS_LONG_ARRAY) {
			return TIBEMS_INVALID_TYPE;
		}
		*array = (tibems_long*)field.data.arrayValue;
		*arrayCount = (tibems_uint)field.count;
	}

	return status;
}

extern tibems_status
tibemsMapMsg_GetFloatArray(
    tibemsMsg            msg,
    const char*          name,
    const tibems_float** array,
    tibems_uint*         arrayCount)
{
	tibemsMsgField field;
	memset(&field, 0, sizeof(tibemsMsgField));
	tibems_status status = tibemsMapMsg_GetField(msg, name, &field);
	if (status == TIBEMS_OK) {
		if (field.type != TIBEMS_FLOAT_ARRAY) {
			return TIBEMS_INVALID_TYPE;
		}
		*array = (tibems_float*)field.data.arrayValue;
		*arrayCount = (tibems_uint)field.count;
	}

	return status;
}

extern tibems_status
tibemsMapMsg_GetDoubleArray(
    tibemsMsg             msg,
    const char*           name,
    const tibems_double** array,
    tibems_uint*          arrayCount)
{
	tibemsMsgField field;
	memset(&field, 0, sizeof(tibemsMsgField));
	tibems_status status = tibemsMapMsg_GetField(msg, name, &field);
	if (status == TIBEMS_OK) {
		if (field.type != TIBEMS_DOUBLE_ARRAY) {
			return TIBEMS_INVALID_TYPE;
		}
		*array = (tibems_double*)field.data.arrayValue;
		*arrayCount = (tibems_uint)field.count;
	}

	return status;
}

extern tibems_status
tibemsMapMsg_GetFieldType(
    tibemsMsg  		msg,
    const char*		name,
    tibems_byte*	type)
{
	tibemsMsgField field;
	memset(&field, 0, sizeof(tibemsMsgField));
	tibems_status status = tibemsMapMsg_GetField(msg, name, &field);
	if (status == TIBEMS_OK) {
		*type = field.type;
	}
	return status;
}

*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Ensure MapMsg implements Message
var _ Message = (*MapMsg)(nil)

// Also make sure it implements MapMessage
var _ MapMessage = (*MapMsg)(nil)

type MapMessage interface {
	SetBoolean(name string, value bool) error
	GetBoolean(name string) (bool, error)
	SetByte(name string, value byte) error
	GetByte(name string) (byte, error)
	SetBytes(name string, value []byte) error
	GetBytes(name string) ([]byte, error)
	SetShort(name string, value int16) error
	GetShort(name string) (int16, error)
	SetInt(name string, value int32) error
	GetInt(name string) (int32, error)
	SetLong(name string, value int64) error
	GetLong(name string) (int64, error)
	SetFloat(name string, value float32) error
	GetFloat(name string) (float32, error)
	SetDouble(name string, value float64) error
	GetDouble(name string) (float64, error)
	SetString(name string, value string) error
	GetString(name string) (string, error)
	SetMapMsg(name string, value *MapMsg) error
	GetMapMsg(name string) (*MapMsg, error)
	SetShortArray(name string, value []int16) error
	GetShortArray(name string) ([]int16, error)
	SetIntArray(name string, value []int32) error
	GetIntArray(name string) ([]int32, error)
	SetLongArray(name string, value []int64) error
	GetLongArray(name string) ([]int64, error)
	SetFloatArray(name string, value []float32) error
	GetFloatArray(name string) ([]float32, error)
	SetDoubleArray(name string, value []float64) error
	GetDoubleArray(name string) ([]float64, error)

	ItemExists(name string) (bool, error)
	GetMapNames() (*MessageEnumeration, error)
}

type jsonMapMsgField struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type jsonMapMsgBody struct {
	Type   string            `json:"type"`
	Fields []jsonMapMsgField `json:"fields"`
}

type jsonMapMsg struct {
	jsonMsg
	Body *jsonMapMsgBody `json:"body"`
}

func (msg *MapMsg) MarshalJSON() ([]byte, error) {
	headers, err := marshalHeader(msg)
	if err != nil {
		return nil, err
	}
	properties, err := marshalProperties(msg)
	if err != nil {
		return nil, err
	}
	fields, err := marshalFields(msg)
	if err != nil {
		return nil, err
	}
	jsonMessage := jsonMapMsg{
		jsonMsg: jsonMsg{
			Header:     headers,
			Properties: properties,
		},
		Body: &jsonMapMsgBody{
			Type:   "MAP",
			Fields: fields,
		},
	}
	return json.Marshal(jsonMessage)
}

func marshalFields(message *MapMsg) ([]jsonMapMsgField, error) {
	fields := make([]jsonMapMsgField, 0)
	fieldNames, err := message.GetMapNames()
	if err != nil {
		return nil, err
	}
	defer func(fieldNames *MessageEnumeration) {
		err := fieldNames.Close()
		if err != nil {
			panic(err)
		}
	}(fieldNames)
	fieldName, _ := fieldNames.GetNextName()
	for fieldName != "" {
		fType, err := message.GetFieldType(fieldName)
		if err != nil {
			return nil, err
		}
		switch fType {
		case FieldTypeNull:
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "null",
				Value: nil,
			})
		case FieldTypeBool:
			value, err := message.GetBoolean(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "bool",
				Value: value,
			})
		case FieldTypeByte:
			value, err := message.GetByte(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "byte",
				Value: value,
			})
		case FieldTypeShort:
			value, err := message.GetShort(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "short",
				Value: value,
			})
		case FieldTypeShortArray:
			value, err := message.GetShortArray(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "short[]",
				Value: value,
			})
		case FieldTypeInt:
			value, err := message.GetInt(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "int",
				Value: value,
			})
		case FieldTypeIntArray:
			value, err := message.GetIntArray(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "int[]",
				Value: value,
			})
		case FieldTypeLong:
			value, err := message.GetLong(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "long",
				Value: value,
			})
		case FieldTypeLongArray:
			value, err := message.GetLongArray(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "long[]",
				Value: value,
			})
		case FieldTypeFloat:
			value, err := message.GetFloat(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "float",
				Value: value,
			})
		case FieldTypeFloatArray:
			value, err := message.GetFloatArray(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "float[]",
				Value: value,
			})
		case FieldTypeDouble:
			value, err := message.GetDouble(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "double",
				Value: value,
			})
		case FieldTypeDoubleArray:
			value, err := message.GetDoubleArray(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "double[]",
				Value: value,
			})
		case FieldTypeUTF8:
			value, err := message.GetString(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "string",
				Value: value,
			})
		case FieldTypeBytes:
			value, err := message.GetBytes(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "bytes",
				Value: value,
			})
		case FieldTypeMapMsg:
			value, err := message.GetMapMsg(fieldName)
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonMapMsgField{
				Name:  fieldName,
				Type:  "map_message",
				Value: value,
			})
		default:
			log.Warn().Msgf("field '%s' type is unsupported; will not be used", fieldName)
		}
		fieldName, _ = fieldNames.GetNextName()
	}

	return fields, nil
}

// A MapMsg is a message containing a set of name-value pairs.
type MapMsg struct {
	Msg

	setBooleanFields     map[string]bool
	setByteFields        map[string]byte
	setBytesFields       map[string][]byte
	setDoubleFields      map[string]float64
	setFloatFields       map[string]float32
	setIntFields         map[string]int32
	setLongFields        map[string]int64
	setShortFields       map[string]int16
	setStringFields      map[string]string
	setMapMsgFields      map[string]*MapMsg
	setShortArrayFields  map[string][]int16
	setIntArrayFields    map[string][]int32
	setLongArrayFields   map[string][]int64
	setFloatArrayFields  map[string][]float32
	setDoubleArrayFields map[string][]float64

	pendingBooleanFields     map[string]bool
	pendingByteFields        map[string]byte
	pendingBytesFields       map[string][]byte
	pendingDoubleFields      map[string]float64
	pendingFloatFields       map[string]float32
	pendingIntFields         map[string]int32
	pendingLongFields        map[string]int64
	pendingShortFields       map[string]int16
	pendingStringFields      map[string]string
	pendingMapMsgFields      map[string]*MapMsg
	pendingShortArrayFields  map[string][]int16
	pendingIntArrayFields    map[string][]int32
	pendingLongArrayFields   map[string][]int64
	pendingFloatArrayFields  map[string][]float32
	pendingDoubleArrayFields map[string][]float64

	fieldNames        []*C.char
	fieldNameLengths  []int32
	fieldTypes        []FieldType
	fieldValues       []uintptr
	fieldValueLengths []uint32
}

func (msg *MapMsg) GetMapNames() (*MessageEnumeration, error) {
	err := msg.flushPending()
	if err != nil {
		return nil, err
	}
	var msgEnum = MessageEnumeration{cMsgEnum: nil}
	status := C.tibemsMapMsg_GetMapNames(msg.cMessage, &msgEnum.cMsgEnum)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &msgEnum, nil
}

func (msg *MapMsg) ItemExists(name string) (bool, error) {
	err := msg.flushPending()
	if err != nil {
		return false, err
	}
	cPropName := C.CString(name)
	defer C.free(unsafe.Pointer(cPropName))
	var cValue C.tibems_bool
	status := C.tibemsMapMsg_ItemExists(msg.cMessage, cPropName, &cValue)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (msg *MapMsg) GetAsBytes() ([]byte, error) {
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

func (msg *MapMsg) CreateCopy() (Message, error) {
	msg.flushPending()
	var cValue C.tibemsMsg
	status := C.tibemsMsg_CreateCopy(msg.cMessage, &cValue)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &MapMsg{
		Msg: Msg{
			cMessage: cValue,
		}}, nil
}

func (msg *MapMsg) flushPending() error {

	err := msg.Msg.flushPending()
	if err != nil {
		return err
	}

	numFields := len(msg.pendingBooleanFields) +
		len(msg.pendingByteFields) +
		len(msg.pendingBytesFields) +
		len(msg.pendingDoubleFields) +
		len(msg.pendingFloatFields) +
		len(msg.pendingIntFields) +
		len(msg.pendingLongFields) +
		len(msg.pendingShortFields) +
		len(msg.pendingStringFields) +
		len(msg.pendingMapMsgFields) +
		len(msg.pendingShortArrayFields) +
		len(msg.pendingIntArrayFields) +
		len(msg.pendingLongArrayFields) +
		len(msg.pendingFloatArrayFields) +
		len(msg.pendingDoubleArrayFields)

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

	for propName, value := range msg.pendingBooleanFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeBool
		if value {
			msg.fieldValues[fieldNum] = 1
		} else {
			msg.fieldValues[fieldNum] = 0
		}
		fieldNum++
		delete(msg.pendingBooleanFields, propName)
		if msg.setBooleanFields == nil {
			msg.setBooleanFields = make(map[string]bool)
		}
		msg.setBooleanFields[propName] = value
	}

	for propName, value := range msg.pendingByteFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeByte
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingByteFields, propName)
		if msg.setByteFields == nil {
			msg.setByteFields = make(map[string]byte)
		}
		msg.setByteFields[propName] = value
	}

	for propName, value := range msg.pendingBytesFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeBytes
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(&value[0]))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingBytesFields, propName)
		if msg.setBytesFields == nil {
			msg.setBytesFields = make(map[string][]byte)
		}
		msg.setBytesFields[propName] = value
	}

	for propName, value := range msg.pendingDoubleFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeDouble
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingDoubleFields, propName)
		if msg.setDoubleFields == nil {
			msg.setDoubleFields = make(map[string]float64)
		}
		msg.setDoubleFields[propName] = value
	}

	for propName, value := range msg.pendingFloatFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeFloat
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingFloatFields, propName)
		if msg.setFloatFields == nil {
			msg.setFloatFields = make(map[string]float32)
		}
		msg.setFloatFields[propName] = value
	}

	for propName, value := range msg.pendingIntFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeInt
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingIntFields, propName)
		if msg.setIntFields == nil {
			msg.setIntFields = make(map[string]int32)
		}
		msg.setIntFields[propName] = value
	}

	for propName, value := range msg.pendingLongFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeLong
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingLongFields, propName)
		if msg.setLongFields == nil {
			msg.setLongFields = make(map[string]int64)
		}
		msg.setLongFields[propName] = value
	}

	for propName, value := range msg.pendingShortFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeShort
		msg.fieldValues[fieldNum] = uintptr(value)
		fieldNum++
		delete(msg.pendingShortFields, propName)
		if msg.setShortFields == nil {
			msg.setShortFields = make(map[string]int16)
		}
		msg.setShortFields[propName] = value
	}

	for propName, value := range msg.pendingStringFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeUTF8
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(C.CString(value)))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingStringFields, propName)
		if msg.setStringFields == nil {
			msg.setStringFields = make(map[string]string)
		}
		msg.setStringFields[propName] = value
	}

	for propName, value := range msg.pendingMapMsgFields {
		value.flushPending()
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeMapMsg
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(value.cMessage))
		fieldNum++
		delete(msg.pendingMapMsgFields, propName)
		if msg.setMapMsgFields == nil {
			msg.setMapMsgFields = make(map[string]*MapMsg)
		}
		msg.setMapMsgFields[propName] = value
	}

	for propName, value := range msg.pendingShortArrayFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeShortArray
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(&value[0]))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingShortArrayFields, propName)
		if msg.setShortArrayFields == nil {
			msg.setShortArrayFields = make(map[string][]int16)
		}
		msg.setShortArrayFields[propName] = value
	}

	for propName, value := range msg.pendingIntArrayFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeIntArray
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(&value[0]))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingIntArrayFields, propName)
		if msg.setIntArrayFields == nil {
			msg.setIntArrayFields = make(map[string][]int32)
		}
		msg.setIntArrayFields[propName] = value
	}

	for propName, value := range msg.pendingLongArrayFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeLongArray
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(&value[0]))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingLongArrayFields, propName)
		if msg.setLongArrayFields == nil {
			msg.setLongArrayFields = make(map[string][]int64)
		}
		msg.setLongArrayFields[propName] = value
	}

	for propName, value := range msg.pendingFloatArrayFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeFloatArray
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(&value[0]))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingFloatArrayFields, propName)
		if msg.setFloatArrayFields == nil {
			msg.setFloatArrayFields = make(map[string][]float32)
		}
		msg.setFloatArrayFields[propName] = value
	}

	for propName, value := range msg.pendingDoubleArrayFields {
		msg.fieldNames[fieldNum] = C.CString(propName)
		msg.fieldNameLengths[fieldNum] = int32(len(propName))
		msg.fieldTypes[fieldNum] = FieldTypeDoubleArray
		msg.fieldValues[fieldNum] = uintptr(unsafe.Pointer(&value[0]))
		msg.fieldValueLengths[fieldNum] = uint32(len(value))
		fieldNum++
		delete(msg.pendingDoubleArrayFields, propName)
		if msg.setDoubleArrayFields == nil {
			msg.setDoubleArrayFields = make(map[string][]float64)
		}
		msg.setDoubleArrayFields[propName] = value
	}

	status := C.tibemsMapMsg_SetFields(msg.cMessage,
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

func (msg *MapMsg) GetFieldType(name string) (FieldType, error) {
	cPropName := C.CString(name)
	defer C.free(unsafe.Pointer(cPropName))
	var fieldType C.tibems_byte
	status := C.tibemsMapMsg_GetFieldType(msg.cMessage, cPropName, &fieldType)
	if status != tibems_OK {
		return FieldTypeNull, getErrorFromStatus(status)
	}
	return FieldType(fieldType), nil
}

func (msg *MapMsg) SetBoolean(name string, value bool) error {
	if msg.pendingBooleanFields == nil {
		msg.pendingBooleanFields = make(map[string]bool)
	}
	msg.pendingBooleanFields[name] = value
	return nil
}
func (msg *MapMsg) GetBoolean(name string) (bool, error) {
	value, found := msg.pendingBooleanFields[name]
	if !found {
		value, found = msg.setBooleanFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_bool

			status := C.tibemsMapMsg_GetBoolean(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return false, ErrNotFound
			}
			if status != tibems_OK {
				return false, ErrNotFound
			}
			if cValue == 0 {
				value = false
			} else {
				value = true
			}
			if msg.setBooleanFields == nil {
				msg.setBooleanFields = make(map[string]bool)
			}
			msg.setBooleanFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetByte(name string, value byte) error {
	if msg.pendingByteFields == nil {
		msg.pendingByteFields = make(map[string]byte)
	}
	msg.pendingByteFields[name] = value
	return nil
}
func (msg *MapMsg) GetByte(name string) (byte, error) {
	value, found := msg.pendingByteFields[name]
	if !found {
		value, found = msg.setByteFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_byte
			status := C.tibemsMapMsg_GetByte(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, ErrNotFound
			}
			value = byte(cValue)
			if msg.setByteFields == nil {
				msg.setByteFields = make(map[string]byte)
			}
			msg.setByteFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetBytes(name string, value []byte) error {
	if msg.pendingBytesFields == nil {
		msg.pendingBytesFields = make(map[string][]byte)
	}
	msg.pendingBytesFields[name] = value
	return nil
}
func (msg *MapMsg) GetBytes(name string) ([]byte, error) {
	value, found := msg.pendingBytesFields[name]
	if !found {
		value, found = msg.setBytesFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue unsafe.Pointer
			var cSize C.tibems_uint
			status := C.tibemsMapMsg_GetBytes(msg.cMessage, cPropName, &cValue, &cSize)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			value = C.GoBytes(cValue, C.int(cSize))
			if msg.setBytesFields == nil {
				msg.setBytesFields = make(map[string][]byte)
			}
			msg.setBytesFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetShort(name string, value int16) error {
	if msg.pendingShortFields == nil {
		msg.pendingShortFields = make(map[string]int16)
	}
	msg.pendingShortFields[name] = value
	return nil
}
func (msg *MapMsg) GetShort(name string) (int16, error) {
	value, found := msg.pendingShortFields[name]
	if !found {
		value, found = msg.setShortFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_short
			status := C.tibemsMapMsg_GetShort(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, ErrNotFound
			}
			value = int16(cValue)
			if msg.setShortFields == nil {
				msg.setShortFields = make(map[string]int16)
			}
			msg.setShortFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetInt(name string, value int32) error {
	if msg.pendingIntFields == nil {
		msg.pendingIntFields = make(map[string]int32)
	}
	msg.pendingIntFields[name] = value
	return nil
}
func (msg *MapMsg) GetInt(name string) (int32, error) {
	value, found := msg.pendingIntFields[name]
	if !found {
		value, found = msg.setIntFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_int
			status := C.tibemsMapMsg_GetInt(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, ErrNotFound
			}
			value = int32(cValue)
			if msg.setIntFields == nil {
				msg.setIntFields = make(map[string]int32)
			}
			msg.setIntFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetLong(name string, value int64) error {
	if msg.pendingLongFields == nil {
		msg.pendingLongFields = make(map[string]int64)
	}
	msg.pendingLongFields[name] = value
	return nil
}
func (msg *MapMsg) GetLong(name string) (int64, error) {
	value, found := msg.pendingLongFields[name]
	if !found {
		value, found = msg.setLongFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_long
			status := C.tibemsMapMsg_GetLong(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, ErrNotFound
			}
			value = int64(cValue)
			if msg.setLongFields == nil {
				msg.setLongFields = make(map[string]int64)
			}
			msg.setLongFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetFloat(name string, value float32) error {
	if msg.pendingFloatFields == nil {
		msg.pendingFloatFields = make(map[string]float32)
	}
	msg.pendingFloatFields[name] = value
	return nil
}
func (msg *MapMsg) GetFloat(name string) (float32, error) {
	value, found := msg.pendingFloatFields[name]
	if !found {
		value, found = msg.setFloatFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_float
			status := C.tibemsMapMsg_GetFloat(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, ErrNotFound
			}
			value = float32(cValue)
			if msg.setFloatFields == nil {
				msg.setFloatFields = make(map[string]float32)
			}
			msg.setFloatFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetDouble(name string, value float64) error {
	if msg.pendingDoubleFields == nil {
		msg.pendingDoubleFields = make(map[string]float64)
	}
	msg.pendingDoubleFields[name] = value
	return nil
}
func (msg *MapMsg) GetDouble(name string) (float64, error) {
	value, found := msg.pendingDoubleFields[name]
	if !found {
		value, found = msg.setDoubleFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibems_double
			status := C.tibemsMapMsg_GetDouble(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return 0, ErrNotFound
			}
			if status != tibems_OK {
				return 0, ErrNotFound
			}
			value = float64(cValue)
			if msg.setDoubleFields == nil {
				msg.setDoubleFields = make(map[string]float64)
			}
			msg.setDoubleFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetString(name string, value string) error {
	if msg.pendingStringFields == nil {
		msg.pendingStringFields = make(map[string]string)
	}
	msg.pendingStringFields[name] = value
	return nil
}
func (msg *MapMsg) GetString(name string) (string, error) {
	value, found := msg.pendingStringFields[name]
	if !found {
		value, found = msg.setStringFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.char
			status := C.tibemsMapMsg_GetString(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return "", ErrNotFound
			}
			if status != tibems_OK {
				return "", ErrNotFound
			}
			value = C.GoString(cValue)
			if msg.setStringFields == nil {
				msg.setStringFields = make(map[string]string)
			}
			msg.setStringFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}
func (msg *MapMsg) SetMapMsg(name string, value *MapMsg) error {
	if msg.pendingMapMsgFields == nil {
		msg.pendingMapMsgFields = make(map[string]*MapMsg)
	}
	msg.pendingMapMsgFields[name] = value
	return nil
}
func (msg *MapMsg) GetMapMsg(name string) (*MapMsg, error) {
	value, found := msg.pendingMapMsgFields[name]
	if !found {
		value, found = msg.setMapMsgFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue C.tibemsMapMsg
			status := C.tibemsMapMsg_GetMapMsg(msg.cMessage, cPropName, &cValue)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			value = &MapMsg{Msg: Msg{cMessage: C.tibemsMsg(cValue)}}
			if msg.setMapMsgFields == nil {
				msg.setMapMsgFields = make(map[string]*MapMsg)
			}
			msg.setMapMsgFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetShortArray(name string, value []int16) error {
	if msg.pendingShortArrayFields == nil {
		msg.pendingShortArrayFields = make(map[string][]int16)
	}
	msg.pendingShortArrayFields[name] = value
	return nil
}
func (msg *MapMsg) GetShortArray(name string) ([]int16, error) {
	value, found := msg.pendingShortArrayFields[name]
	if !found {
		value, found = msg.setShortArrayFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.tibems_short
			var cCount C.tibems_uint
			status := C.tibemsMapMsg_GetShortArray(msg.cMessage, cPropName, &cValue, &cCount)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			if msg.setShortArrayFields == nil {
				msg.setShortArrayFields = make(map[string][]int16)
			}
			cArray := unsafe.Slice((*int16)(unsafe.Pointer(cValue)), (int)(cCount))
			value := make([]int16, len(cArray))
			copy(value, cArray)
			msg.setShortArrayFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetIntArray(name string, value []int32) error {
	if msg.pendingIntArrayFields == nil {
		msg.pendingIntArrayFields = make(map[string][]int32)
	}
	msg.pendingIntArrayFields[name] = value
	return nil
}
func (msg *MapMsg) GetIntArray(name string) ([]int32, error) {
	value, found := msg.pendingIntArrayFields[name]
	if !found {
		value, found = msg.setIntArrayFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.tibems_int
			var cCount C.tibems_uint
			status := C.tibemsMapMsg_GetIntArray(msg.cMessage, cPropName, &cValue, &cCount)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			if msg.setIntArrayFields == nil {
				msg.setIntArrayFields = make(map[string][]int32)
			}
			cArray := unsafe.Slice((*int32)(unsafe.Pointer(cValue)), (int)(cCount))
			value := make([]int32, len(cArray))
			copy(value, cArray)
			msg.setIntArrayFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetLongArray(name string, value []int64) error {
	if msg.pendingLongArrayFields == nil {
		msg.pendingLongArrayFields = make(map[string][]int64)
	}
	msg.pendingLongArrayFields[name] = value
	return nil
}
func (msg *MapMsg) GetLongArray(name string) ([]int64, error) {
	value, found := msg.pendingLongArrayFields[name]
	if !found {
		value, found = msg.setLongArrayFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.tibems_long
			var cCount C.tibems_uint
			status := C.tibemsMapMsg_GetLongArray(msg.cMessage, cPropName, &cValue, &cCount)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			if msg.setLongArrayFields == nil {
				msg.setLongArrayFields = make(map[string][]int64)
			}
			cArray := unsafe.Slice((*int64)(unsafe.Pointer(cValue)), (int)(cCount))
			value := make([]int64, len(cArray))
			copy(value, cArray)
			msg.setLongArrayFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetFloatArray(name string, value []float32) error {
	if msg.pendingFloatArrayFields == nil {
		msg.pendingFloatArrayFields = make(map[string][]float32)
	}
	msg.pendingFloatArrayFields[name] = value
	return nil
}
func (msg *MapMsg) GetFloatArray(name string) ([]float32, error) {
	value, found := msg.pendingFloatArrayFields[name]
	if !found {
		value, found = msg.setFloatArrayFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.tibems_float
			var cCount C.tibems_uint
			status := C.tibemsMapMsg_GetFloatArray(msg.cMessage, cPropName, &cValue, &cCount)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			if msg.setFloatArrayFields == nil {
				msg.setFloatArrayFields = make(map[string][]float32)
			}
			cArray := unsafe.Slice((*float32)(unsafe.Pointer(cValue)), (int)(cCount))
			value := make([]float32, len(cArray))
			copy(value, cArray)
			msg.setFloatArrayFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) SetDoubleArray(name string, value []float64) error {
	if msg.pendingDoubleArrayFields == nil {
		msg.pendingDoubleArrayFields = make(map[string][]float64)
	}
	msg.pendingDoubleArrayFields[name] = value
	return nil
}
func (msg *MapMsg) GetDoubleArray(name string) ([]float64, error) {
	value, found := msg.pendingDoubleArrayFields[name]
	if !found {
		value, found = msg.setDoubleArrayFields[name]
		if !found {
			cPropName := C.CString(name)
			defer C.free(unsafe.Pointer(cPropName))
			var cValue *C.tibems_double
			var cCount C.tibems_uint
			status := C.tibemsMapMsg_GetDoubleArray(msg.cMessage, cPropName, &cValue, &cCount)
			if status == tibems_NOT_FOUND {
				return nil, ErrNotFound
			}
			if status != tibems_OK {
				return nil, ErrNotFound
			}
			if msg.setDoubleArrayFields == nil {
				msg.setDoubleArrayFields = make(map[string][]float64)
			}
			cArray := unsafe.Slice((*float64)(unsafe.Pointer(cValue)), (int)(cCount))
			value := make([]float64, len(cArray))
			copy(value, cArray)
			msg.setDoubleArrayFields[name] = value
			return value, nil
		} else {
			return value, nil
		}
	} else {
		return value, nil
	}
}

func (msg *MapMsg) clearFieldMaps() {
	msg.Msg.clearFieldMaps()
	msg.setBooleanFields = nil
	msg.setByteFields = nil
	msg.setBytesFields = nil
	msg.setDoubleFields = nil
	msg.setFloatFields = nil
	msg.setIntFields = nil
	msg.setLongFields = nil
	msg.setShortFields = nil
	msg.setStringFields = nil
	msg.setMapMsgFields = nil
	msg.setShortArrayFields = nil
	msg.setIntArrayFields = nil
	msg.setLongArrayFields = nil
	msg.setFloatArrayFields = nil
	msg.setDoubleArrayFields = nil
	msg.pendingBooleanFields = nil
	msg.pendingByteFields = nil
	msg.pendingBytesFields = nil
	msg.pendingDoubleFields = nil
	msg.pendingFloatFields = nil
	msg.pendingIntFields = nil
	msg.pendingLongFields = nil
	msg.pendingShortFields = nil
	msg.pendingStringFields = nil
	msg.pendingMapMsgFields = nil
	msg.pendingShortArrayFields = nil
	msg.pendingIntArrayFields = nil
	msg.pendingLongArrayFields = nil
	msg.pendingFloatArrayFields = nil
	msg.pendingDoubleArrayFields = nil
}

// CreateMapMsg creates a new, empty MapMsg.
//
// Applications MUST call [MapMsg.Close] when finished with a MapMsg to avoid resource leaks.
func CreateMapMsg() (*MapMsg, error) {
	var message = MapMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.tibemsMapMsg_Create((*C.tibemsMapMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// Close cleans up resources used by a MapMsg. A MapMsg MUST NOT be used again after Close has been called.
func (message *MapMsg) Close() error {
	if message == nil {
		return nil
	}
	message.clearFieldMaps()
	status := C.tibemsMsg_Destroy(message.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (msg *MapMsg) ClearBody() error {
	msg.clearFieldMaps()
	status := C.tibemsMsg_ClearBody(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *MapMsg) ClearProperties() error {
	msg.Msg.clearFieldMaps()
	status := C.tibemsMsg_ClearProperties(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (msg *MapMsg) Marshal(s any) error {
	_, err := marshalInternal(msg, s)
	return err
}

func Marshal(s any) (*MapMsg, error) {
	msg, err := CreateMapMsg()
	if err != nil {
		return nil, err
	}
	return msg, msg.Marshal(s)
}

func marshalMapMsgField(msg *MapMsg, fieldName string, fieldValue any) error {
	kind := reflect.TypeOf(fieldValue).Kind()
	if kind == reflect.Pointer {
		if reflect.ValueOf(fieldValue).IsNil() {
			return nil
		}
		fieldValue = reflect.ValueOf(fieldValue).Elem().Interface()
		kind = reflect.TypeOf(fieldValue).Kind()
	}

	switch kind {
	case reflect.Array, reflect.Slice:
		subKind := reflect.TypeOf(fieldValue).Elem().Kind()
		switch subKind {
		case reflect.Uint8:
			msg.SetBytes(fieldName, fieldValue.([]uint8))
		case reflect.Int16:
			msg.SetShortArray(fieldName, fieldValue.([]int16))
		case reflect.Int32:
			msg.SetIntArray(fieldName, fieldValue.([]int32))
		case reflect.Int64:
			msg.SetLongArray(fieldName, fieldValue.([]int64))
		case reflect.Float32:
			msg.SetFloatArray(fieldName, fieldValue.([]float32))
		case reflect.Float64:
			msg.SetDoubleArray(fieldName, fieldValue.([]float64))
		case reflect.String:
			strArray := fieldValue.([]string)
			for i := 0; i < len(strArray); i++ {
				msg.SetString(fieldName+strconv.Itoa(i), strArray[i])
			}
			msg.SetInt("n"+fieldName, int32(len(strArray)))
		default:
			return errors.New("struct field '" + fieldName + "' type is unsupported")
		}
	case reflect.Uint8:
		msg.SetByte(fieldName, fieldValue.(uint8))
	case reflect.Int16:
		msg.SetShort(fieldName, fieldValue.(int16))
	case reflect.Int32:
		msg.SetInt(fieldName, fieldValue.(int32))
	case reflect.Int64:
		msg.SetLong(fieldName, fieldValue.(int64))
	case reflect.String:
		msg.SetString(fieldName, fieldValue.(string))
	case reflect.Bool:
		msg.SetBoolean(fieldName, fieldValue.(bool))
	case reflect.Float32:
		msg.SetFloat(fieldName, fieldValue.(float32))
	case reflect.Float64:
		msg.SetDouble(fieldName, fieldValue.(float64))
	case reflect.Struct:
		nestedMsg, err := Marshal(fieldValue)
		if err != nil {
			return err
		}
		msg.SetMapMsg(fieldName, nestedMsg)
	case reflect.Map:
		nestedMsg, err := Marshal(fieldValue)
		if err != nil {
			return err
		}
		msg.SetMapMsg(fieldName, nestedMsg)
	default:
		return errors.New("struct field '" + fieldName + "' type is unsupported")
	}
	return nil
}

type Unmarshaler interface {
	UnmarshalMapMessage(msg *MapMsg) error
}

type Marshaler interface {
	MarshalMapMessage(msg *MapMsg) error
}

func getFieldOpts(field reflect.StructField) (string, bool, bool, bool, bool) {
	fieldName := field.Name
	isProperty := false
	omit := false
	omitEmpty := false
	required := false
	opts := field.Tag.Get("msg")
	if opts != "" {
		options := strings.Split(opts, ",")
		if options[0] != "" {
			fieldName = options[0]
		}
		if len(options) > 1 {
			options = options[1:]
			for _, option := range options {
				switch option {
				case "prop":
					isProperty = true
				case "omitempty":
					omitEmpty = true
				case "required":
					required = true
				case "-":
					omit = true
				}
			}
		}
	}

	return fieldName, isProperty, omit, omitEmpty, required
}

func marshalInternal(msg *MapMsg, s any) (*MapMsg, error) {
	if reflect.TypeOf(s).Kind() == reflect.Pointer {
		s = reflect.ValueOf(s).Elem().Interface()
	}

	if reflect.TypeOf(s).Implements(reflect.TypeOf((*Marshaler)(nil)).Elem()) {
		iface := reflect.ValueOf(s).Interface().(Marshaler)
		return msg, iface.MarshalMapMessage(msg)
	}

	if reflect.TypeOf(s).Kind() == reflect.Struct {
		fields := reflect.VisibleFields(reflect.TypeOf(s))
		for i := 0; i < len(fields); i++ {
			field := fields[i]
			if field.Anonymous {
				continue
			}

			fieldName, isProperty, omit, omitEmpty, required := getFieldOpts(field)
			if omit {
				continue
			}
			fieldValue, err := reflect.ValueOf(s).FieldByIndexErr(field.Index)
			if err != nil {
				continue
			}
			if omitEmpty && fieldValue.IsZero() {
				continue
			}
			if (fieldValue.Kind() == reflect.Pointer) && fieldValue.IsNil() {
				if required {
					return nil, errors.New(fmt.Sprintf("field '%s' is required but was not set", fieldName))
				}
			}
			fieldInterface := fieldValue.Interface()
			if isProperty {
				err := marshalMsgPropertyField(msg, fieldName, fieldInterface)
				if err != nil {
					return nil, err
				}
			} else {
				err := marshalMapMsgField(msg, fieldName, fieldInterface)
				if err != nil {
					return nil, err
				}
			}
		}
	} else if reflect.TypeOf(s).Kind() == reflect.Map {
		fieldMap := s.(map[string]any)
		for fieldName, fieldValue := range fieldMap {
			err := marshalMapMsgField(msg, fieldName, fieldValue)
			if err != nil {
				return nil, err
			}
		}
	}
	return msg, nil
}

func Unmarshal(msg *MapMsg, s any) error {

	// Make sure we were given a pointer to a struct.
	t := reflect.TypeOf(s)
	if t.Kind() != reflect.Pointer {
		return errors.New("Unmarshal requires a pointer to a struct")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return errors.New("Unmarshal requires a pointer to a struct")
	}

	fields := reflect.VisibleFields(t)
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		if field.Anonymous {
			continue
		}
		fieldName, isProperty, omit, _, required := getFieldOpts(field)
		if omit {
			continue
		}
		fieldValue, err := reflect.ValueOf(s).Elem().FieldByIndexErr(field.Index)
		if err != nil {
			continue
		}
		fieldType := field.Type
		fieldKindWasPointer := false
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
			fieldKindWasPointer = true
		}
		msgFieldExists := false
		if isProperty {
			msgFieldExists, err = msg.PropertyExists(fieldName)
		} else {
			msgFieldExists, err = msg.ItemExists(fieldName)
		}
		if err != nil {
			return err
		}
		if !msgFieldExists {
			if (fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array) && (fieldType.Elem().Kind() == reflect.Struct || fieldType.Elem().Kind() == reflect.String) {
				// For arrays/slices of structs or strings, EMS doesn't support a native array type field, so it'll be "faked" with
				// several independent message fields named something like 'basename0', 'basename1', 'basename2', etc.
				if isProperty {
					msgFieldExists, err = msg.PropertyExists(fieldName + "0")
				} else {
					msgFieldExists, err = msg.ItemExists(fieldName + "0")
				}
				if err != nil {
					return err
				}
			}
		}

		if field.Type.Implements(reflect.TypeOf((*Unmarshaler)(nil)).Elem()) {
			iface := fieldValue.Interface().(Unmarshaler)
			err := iface.UnmarshalMapMessage(msg)
			if err != nil {
				return err
			}
			continue
		} else if !msgFieldExists {
			if required {
				return errors.New(fmt.Sprintf("field '%s' required but not present in map message", fieldName))
			}
			continue
		}

		switch fieldType.Kind() {
		case reflect.Struct:
			// If a pointer to another struct, then look for a nested MapMsg and
			// call Unmarshal recursively on it if found.
			nestedMsg, err := msg.GetMapMsg(fieldName)
			if err != nil {
				return err
			}
			// If struct doesn't exist yet, create it.
			if fieldKindWasPointer {
				if fieldValue.IsNil() {
					newValue := reflect.New(fieldType)
					fieldValue.Set(newValue)
				}
				structPtr := fieldValue.Interface()
				err = Unmarshal(nestedMsg, structPtr)
			} else {
				structPtr := fieldValue.Addr().Interface()
				err = Unmarshal(nestedMsg, structPtr)
			}

			if err != nil {
				return err
			}

		case reflect.Array, reflect.Slice:

			switch fieldType.Elem().Kind() {
			case reflect.Struct:
				var i = 0
				for msgFieldExists {
					nestedMsg, err := msg.GetMapMsg(fieldName + strconv.Itoa(i))
					if err != nil {
						return err
					}

					newValue := reflect.New(fieldType.Elem())
					structPtr := newValue.Interface()
					err = Unmarshal(nestedMsg, structPtr)
					if err != nil {
						return err
					}
					fieldValue.Set(reflect.Append(fieldValue, newValue.Elem()))
					i++
					msgFieldExists, err = msg.ItemExists(fieldName + strconv.Itoa(i))
					if err != nil {
						return err
					}
				}

			case reflect.String:
				// MapMsg doesn't have a string array type natively... it's frequently faked by either
				// creating a sub-message with a bunch of bool fields named after each string that
				// would be in the array... or by the usual array technique with an int field giving the
				// length of the array.  We'll try to intelligently guess which one this might be.
				exists, err := msg.ItemExists(fieldName)
				if err != nil {
					return err
				}
				if exists {
					msgFieldType, err := msg.GetFieldType(fieldName)
					if err != nil {
						return err
					}
					if msgFieldType == FieldTypeMapMsg {
						if value, err := msg.GetMapMsg(fieldName); err == nil {
							var strArray []string

							// Check for a length field.
							lengthFieldPresent, err := value.ItemExists("n")
							if err != nil {
								return err
							}
							if lengthFieldPresent {
								lengthFieldType, err := value.GetFieldType("n")
								if err != nil {
									return err
								}
								if lengthFieldType == FieldTypeInt {
									mapnames, err := value.GetMapNames()
									if err != nil {
										return err
									}
									str, _ := mapnames.GetNextName()
									for str != "" {
										strArray = append(strArray, str)
										str, _ = mapnames.GetNextName()
									}
									err = mapnames.Close()
									if err != nil {
										panic(err)
									}
									if fieldKindWasPointer {
										fieldValue.Set(reflect.ValueOf(&strArray).Convert(fieldValue.Type()))
									} else {
										fieldValue.Set(reflect.ValueOf(strArray).Convert(fieldValue.Type()))
									}
								}
							} else {
								// Just assume all fields in the message are string names as long as they're all bool fields.
								mapnames, err := value.GetMapNames()
								if err != nil {
									return err
								}
								str, _ := mapnames.GetNextName()
								for str != "" {
									fType, err := value.GetFieldType(str)
									if err != nil {
										closeErr := mapnames.Close()
										if closeErr != nil {
											panic(closeErr)
										}
										return err
									}
									if fType != FieldTypeBool {
										err := mapnames.Close()
										if err != nil {
											panic(err)
										}
										return errors.New("unknown string array type")
									}
									strArray = append(strArray, str)
									str, _ = mapnames.GetNextName()
								}
								err = mapnames.Close()
								if err != nil {
									panic(err)
								}
								if fieldKindWasPointer {
									fieldValue.Set(reflect.ValueOf(&strArray).Convert(fieldValue.Type()))
								} else {
									fieldValue.Set(reflect.ValueOf(strArray).Convert(fieldValue.Type()))
								}
							}
						}
					}
				} else {
					var strArray []string
					var i = 0
					exists, err = msg.ItemExists(fieldName + strconv.Itoa(i))
					if err != nil {
						return err
					}
					for exists {
						msgFieldType, err := msg.GetFieldType(fieldName + strconv.Itoa(i))
						if err != nil {
							return err
						}
						if msgFieldType == FieldTypeUTF8 {
							str, err := msg.GetString(fieldName + strconv.Itoa(i))
							if err != nil {
								return err
							}
							strArray = append(strArray, str)
						}
						i++
						exists, err = msg.ItemExists(fieldName + strconv.Itoa(i))
						if err != nil {
							return err
						}
					}
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&strArray).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(strArray).Convert(fieldValue.Type()))
					}
				}

			case reflect.Uint8:
				if value, err := msg.GetBytes(fieldName); err == nil {
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
					}
				} else {
					return err
				}
			case reflect.Int16:
				if value, err := msg.GetShortArray(fieldName); err == nil {
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
					}
				} else {
					return err
				}
			case reflect.Int32:
				if value, err := msg.GetIntArray(fieldName); err == nil {
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
					}
				} else {
					return err
				}
			case reflect.Int64:
				if value, err := msg.GetLongArray(fieldName); err == nil {
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
					}
				} else {
					return err
				}
			case reflect.Float32:
				if value, err := msg.GetFloatArray(fieldName); err == nil {
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
					}
				} else {
					return err
				}
			case reflect.Float64:
				if value, err := msg.GetDoubleArray(fieldName); err == nil {
					if fieldKindWasPointer {
						fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
					} else {
						fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
					}
				} else {
					return err
				}
			default:
				log.Debug().Msgf("struct field '%s' is not a supported type and will not be unmarshalled to", field.Name)
			}
		case reflect.Uint8:
			var value byte
			if isProperty {
				value, err = msg.GetByteProperty(fieldName)
			} else {
				value, err = msg.GetByte(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.Int16:
			var value int16
			if isProperty {
				value, err = msg.GetShortProperty(fieldName)
			} else {
				value, err = msg.GetShort(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.Int32:
			var value int32
			if isProperty {
				value, err = msg.GetIntProperty(fieldName)
			} else {
				value, err = msg.GetInt(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.Int64:
			var value int64
			if isProperty {
				value, err = msg.GetLongProperty(fieldName)
			} else {
				value, err = msg.GetLong(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.String:
			var value string
			if isProperty {
				value, err = msg.GetStringProperty(fieldName)
			} else {
				value, err = msg.GetString(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.Bool:
			var value bool
			if isProperty {
				value, err = msg.GetBooleanProperty(fieldName)
			} else {
				value, err = msg.GetBoolean(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.Float32:
			var value float32
			if isProperty {
				value, err = msg.GetFloatProperty(fieldName)
			} else {
				value, err = msg.GetFloat(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		case reflect.Float64:
			var value float64
			if isProperty {
				value, err = msg.GetDoubleProperty(fieldName)
			} else {
				value, err = msg.GetDouble(fieldName)
			}
			if err == nil {
				if fieldKindWasPointer {
					fieldValue.Set(reflect.ValueOf(&value).Convert(fieldValue.Type()))
				} else {
					fieldValue.Set(reflect.ValueOf(value).Convert(fieldValue.Type()))
				}
			} else {
				return err
			}
		default:
			log.Debug().Msgf("struct field '%s' is not a supported type and will not be unmarshalled to", field.Name)
		}
	}

	return nil
}
