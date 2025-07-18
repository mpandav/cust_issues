package tibems

/*
#include <tibems/tibems.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

// Ensure StreamMsg implements Message
var _ Message = (*StreamMsg)(nil)

type StreamMsg struct {
	Msg
}

func (msg *StreamMsg) ReadBoolean() (bool, error) {
	var cValue C.tibems_bool

	status := C.tibemsStreamMsg_ReadBoolean(msg.cMessage, &cValue)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (msg *StreamMsg) WriteBoolean(value bool) error {
	var cValue C.tibems_bool
	if value {
		cValue = 1
	} else {
		cValue = 0
	}
	status := C.tibemsStreamMsg_WriteBoolean(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteStreamMsg(value *StreamMsg) error {
	value.flushPending()
	status := C.tibemsStreamMsg_WriteStreamMsg(msg.cMessage, value.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteMapMsg(value *MapMsg) error {
	value.flushPending()
	status := C.tibemsStreamMsg_WriteMapMsg(msg.cMessage, value.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadByte() (byte, error) {
	var cValue C.tibems_byte

	status := C.tibemsStreamMsg_ReadByte(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return byte(cValue), nil
}

func (msg *StreamMsg) WriteByte(value byte) error {
	status := C.tibemsStreamMsg_WriteByte(msg.cMessage, C.tibems_byte(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadDouble() (float64, error) {
	var cValue C.tibems_double

	status := C.tibemsStreamMsg_ReadDouble(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return float64(cValue), nil
}

func (msg *StreamMsg) WriteDouble(value float64) error {
	status := C.tibemsStreamMsg_WriteDouble(msg.cMessage, C.tibems_double(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteDoubleArray(value []float64) error {
	cArray := (*C.tibems_double)(unsafe.Pointer(&value[0]))
	status := C.tibemsStreamMsg_WriteDoubleArray(msg.cMessage, cArray, C.tibems_int(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadFloat() (float32, error) {
	var cValue C.tibems_float

	status := C.tibemsStreamMsg_ReadFloat(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return float32(cValue), nil
}

func (msg *StreamMsg) WriteFloat(value float32) error {
	status := C.tibemsStreamMsg_WriteFloat(msg.cMessage, C.tibems_float(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteFloatArray(value []float32) error {
	cArray := (*C.tibems_float)(unsafe.Pointer(&value[0]))
	status := C.tibemsStreamMsg_WriteFloatArray(msg.cMessage, cArray, C.tibems_int(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadInt() (int32, error) {
	var cValue C.tibems_int

	status := C.tibemsStreamMsg_ReadInt(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return int32(cValue), nil
}

func (msg *StreamMsg) WriteInt(value int32) error {
	status := C.tibemsStreamMsg_WriteInt(msg.cMessage, C.tibems_int(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteIntArray(value []int32) error {
	cArray := (*C.tibems_int)(unsafe.Pointer(&value[0]))
	status := C.tibemsStreamMsg_WriteIntArray(msg.cMessage, cArray, C.tibems_int(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadLong() (int64, error) {
	var cValue C.tibems_long

	status := C.tibemsStreamMsg_ReadLong(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return int64(cValue), nil
}

func (msg *StreamMsg) WriteLong(value int64) error {
	status := C.tibemsStreamMsg_WriteLong(msg.cMessage, C.tibems_long(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteLongArray(value []int64) error {
	cArray := (*C.tibems_long)(unsafe.Pointer(&value[0]))
	status := C.tibemsStreamMsg_WriteLongArray(msg.cMessage, cArray, C.tibems_int(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadShort() (int16, error) {
	var cValue C.tibems_short

	status := C.tibemsStreamMsg_ReadShort(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return int16(cValue), nil
}

func (msg *StreamMsg) WriteShort(value int16) error {
	status := C.tibemsStreamMsg_WriteShort(msg.cMessage, C.tibems_short(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) WriteShortArray(value []int16) error {
	cArray := (*C.tibems_short)(unsafe.Pointer(&value[0]))
	status := C.tibemsStreamMsg_WriteShortArray(msg.cMessage, cArray, C.tibems_int(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadString() (string, error) {
	var cValue *C.char

	status := C.tibemsStreamMsg_ReadString(msg.cMessage, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}

	return C.GoString(cValue), nil
}

func (msg *StreamMsg) WriteString(value string) error {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	status := C.tibemsStreamMsg_WriteString(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) ReadBytes() ([]byte, error) {
	var cValue unsafe.Pointer
	var cSize C.tibems_uint
	status := C.tibemsStreamMsg_ReadBytes(msg.cMessage, &cValue, &cSize)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	value := C.GoBytes(cValue, C.int(cSize))
	return value, nil
}

func (msg *StreamMsg) WriteBytes(value []byte) error {
	status := C.tibemsStreamMsg_WriteBytes(msg.cMessage, unsafe.Pointer(&value[0]), C.tibems_uint(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *StreamMsg) Reset() error {
	status := C.tibemsStreamMsg_Reset(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

// CreateStreamMsg creates a new, empty StreamMsg.
//
// Applications MUST call [StreamMsg.Close] when finished with a StreamMsg to avoid resource leaks.
func CreateStreamMsg() (*StreamMsg, error) {
	var message = StreamMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.tibemsStreamMsg_Create((*C.tibemsStreamMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// Close cleans up resources used by a StreamMsg. A StreamMsg MUST NOT be used again after Close has been called.
func (message *StreamMsg) Close() error {
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

type jsonStreamMsgField struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type jsonStreamMsgBody struct {
	Type   string               `json:"type"`
	Fields []jsonStreamMsgField `json:"fields"`
}

type jsonStreamMsg struct {
	jsonMsg
	Body *jsonStreamMsgBody `json:"body"`
}

func (msg *StreamMsg) MarshalJSON() ([]byte, error) {
	headers, err := marshalHeader(msg)
	if err != nil {
		return nil, err
	}
	properties, err := marshalProperties(msg)
	if err != nil {
		return nil, err
	}
	err = msg.Reset()
	if err != nil {
		return nil, err
	}
	fields := make([]jsonStreamMsgField, 0)
	for {
		var cField C.tibemsMsgField
		status := C.tibemsStreamMsg_ReadField(msg.cMessage, &cField)
		if status != tibems_OK {
			break
		}
		switch FieldType(cField._type) {
		case FieldTypeBool:
			if cField.data[0] == 0 {
				fields = append(fields, jsonStreamMsgField{
					Type:  "bool",
					Value: false,
				})
			} else {
				fields = append(fields, jsonStreamMsgField{
					Type:  "bool",
					Value: true,
				})
			}
		case FieldTypeByte:
			fields = append(fields, jsonStreamMsgField{
				Type:  "byte",
				Value: cField.data[0],
			})
		case FieldTypeUTF8:
			fields = append(fields, jsonStreamMsgField{
				Type:  "string",
				Value: C.GoString(*(**C.char)(unsafe.Pointer(&cField.data[0]))),
			})
		case FieldTypeBytes:
			fields = append(fields, jsonStreamMsgField{
				Type:  "bytes",
				Value: C.GoBytes(unsafe.Pointer(*(**C.char)(unsafe.Pointer(&cField.data[0]))), cField.size),
			})
		case FieldTypeFloat:
			fields = append(fields, jsonStreamMsgField{
				Type:  "float",
				Value: float32(*(*C.tibems_float)(unsafe.Pointer(&cField.data[0]))),
			})
		case FieldTypeFloatArray:
			cArray := unsafe.Slice((*float32)(unsafe.Pointer(&cField.data[0])), int(cField.count))
			value := make([]float32, len(cArray))
			copy(value, cArray)
			fields = append(fields, jsonStreamMsgField{
				Type:  "float[]",
				Value: value,
			})
		case FieldTypeDouble:
			fields = append(fields, jsonStreamMsgField{
				Type:  "double",
				Value: float64(*(*C.tibems_double)(unsafe.Pointer(&cField.data[0]))),
			})
		case FieldTypeDoubleArray:
			cArray := unsafe.Slice((*float64)(unsafe.Pointer(&cField.data[0])), int(cField.count))
			value := make([]float64, len(cArray))
			copy(value, cArray)
			fields = append(fields, jsonStreamMsgField{
				Type:  "double[]",
				Value: value,
			})
		case FieldTypeShort:
			fields = append(fields, jsonStreamMsgField{
				Type:  "short",
				Value: int16(*(*C.tibems_short)(unsafe.Pointer(&cField.data[0]))),
			})
		case FieldTypeShortArray:
			cArray := unsafe.Slice((*int16)(unsafe.Pointer(&cField.data[0])), int(cField.count))
			value := make([]int16, len(cArray))
			copy(value, cArray)
			fields = append(fields, jsonStreamMsgField{
				Type:  "short[]",
				Value: value,
			})
		case FieldTypeInt:
			fields = append(fields, jsonStreamMsgField{
				Type:  "int",
				Value: int32(*(*C.tibems_int)(unsafe.Pointer(&cField.data[0]))),
			})
		case FieldTypeIntArray:
			cArray := unsafe.Slice((*int32)(unsafe.Pointer(&cField.data[0])), int(cField.count))
			value := make([]int32, len(cArray))
			copy(value, cArray)
			fields = append(fields, jsonStreamMsgField{
				Type:  "int[]",
				Value: value,
			})
		case FieldTypeLong:
			fields = append(fields, jsonStreamMsgField{
				Type:  "long",
				Value: int64(*(*C.tibems_long)(unsafe.Pointer(&cField.data[0]))),
			})
		case FieldTypeLongArray:
			cArray := unsafe.Slice((*int64)(unsafe.Pointer(&cField.data[0])), int(cField.count))
			value := make([]int64, len(cArray))
			copy(value, cArray)
			fields = append(fields, jsonStreamMsgField{
				Type:  "long[]",
				Value: value,
			})
		case FieldTypeStreamMsg:
			msg, err := instantiateSpecificMessageType(*(*C.tibemsMsg)(unsafe.Pointer(&cField.data[0])))
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonStreamMsgField{
				Type:  "stream_message",
				Value: msg,
			})
		case FieldTypeMapMsg:
			msg, err := instantiateSpecificMessageType(*(*C.tibemsMsg)(unsafe.Pointer(&cField.data[0])))
			if err != nil {
				return nil, err
			}
			fields = append(fields, jsonStreamMsgField{
				Type:  "map_message",
				Value: msg,
			})
		}
	}

	err = msg.Reset()
	if err != nil {
		return nil, err
	}

	jsonMessage := jsonStreamMsg{
		jsonMsg: jsonMsg{
			Header:     headers,
			Properties: properties,
		},
		Body: &jsonStreamMsgBody{
			Type:   "STREAM",
			Fields: fields,
		},
	}
	return json.Marshal(jsonMessage)
}
