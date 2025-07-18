package tibems

/*
#include <tibems/tibems.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

// Ensure BytesMsg implements Message
var _ Message = (*BytesMsg)(nil)

// A BytesMsg contains a sequence of raw bytes.
type BytesMsg struct {
	Msg
}

// SetBytes sets the content of a BytesMsg.
func (msg *BytesMsg) SetBytes(bytes []byte) error {
	status := C.tibemsBytesMsg_SetBytes(msg.cMessage, unsafe.Pointer(&bytes[0]), C.tibems_uint(len(bytes)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

// GetBytes gets the content of a BytesMsg as a []byte array.
func (msg *BytesMsg) GetBytes() ([]byte, error) {
	var cValue unsafe.Pointer
	var cSize C.tibems_uint
	status := C.tibemsBytesMsg_GetBytes(msg.cMessage, &cValue, &cSize)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	value := C.GoBytes(cValue, C.int(cSize))
	return value, nil
}

// GetBodyLength returns the number of bytes in the body of a BytesMsg.
func (msg *BytesMsg) GetBodyLength() (int32, error) {
	var cLength C.tibems_int
	status := C.tibemsBytesMsg_GetBodyLength(msg.cMessage, &cLength)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}

	return int32(cLength), nil
}

func (msg *BytesMsg) ReadBoolean() (bool, error) {
	var cValue C.tibems_bool
	status := C.tibemsBytesMsg_ReadBoolean(msg.cMessage, &cValue)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (msg *BytesMsg) ReadBytes(requestedLength int32) ([]byte, error) {
	var cValue unsafe.Pointer
	var cReturnLength C.tibems_int
	status := C.tibemsBytesMsg_ReadBytes(msg.cMessage, &cValue, C.tibems_int(requestedLength), &cReturnLength)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	value := C.GoBytes(cValue, C.int(cReturnLength))
	return value, nil
}

func (msg *BytesMsg) ReadByte() (byte, error) {
	var cValue C.tibems_byte
	status := C.tibemsBytesMsg_ReadByte(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return byte(cValue), nil
}

func (msg *BytesMsg) ReadDouble() (float64, error) {
	var cValue C.tibems_double
	status := C.tibemsBytesMsg_ReadDouble(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return float64(cValue), nil
}

func (msg *BytesMsg) ReadFloat() (float32, error) {
	var cValue C.tibems_float
	status := C.tibemsBytesMsg_ReadFloat(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return float32(cValue), nil
}

func (msg *BytesMsg) ReadInt() (int32, error) {
	var cValue C.tibems_int
	status := C.tibemsBytesMsg_ReadInt(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return int32(cValue), nil
}

func (msg *BytesMsg) ReadLong() (int64, error) {
	var cValue C.tibems_long
	status := C.tibemsBytesMsg_ReadLong(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return int64(cValue), nil
}

func (msg *BytesMsg) ReadShort() (int16, error) {
	var cValue C.tibems_short
	status := C.tibemsBytesMsg_ReadShort(msg.cMessage, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return int16(cValue), nil
}

func (msg *BytesMsg) WriteBoolean(value bool) error {
	var cValue C.tibems_bool
	if value == true {
		cValue = 1
	} else {
		cValue = 0
	}
	status := C.tibemsBytesMsg_WriteBoolean(msg.cMessage, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteBytes(value []byte) error {
	status := C.tibemsBytesMsg_WriteBytes(msg.cMessage, unsafe.Pointer(&value[0]), C.tibems_uint(len(value)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteByte(value byte) error {
	status := C.tibemsBytesMsg_WriteByte(msg.cMessage, C.tibems_byte(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteDouble(value float64) error {
	status := C.tibemsBytesMsg_WriteDouble(msg.cMessage, C.tibems_double(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteFloat(value float32) error {
	status := C.tibemsBytesMsg_WriteFloat(msg.cMessage, C.tibems_float(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteInt(value int32) error {
	status := C.tibemsBytesMsg_WriteInt(msg.cMessage, C.tibems_int(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteLong(value int64) error {
	status := C.tibemsBytesMsg_WriteLong(msg.cMessage, C.tibems_long(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *BytesMsg) WriteShort(value int16) error {
	status := C.tibemsBytesMsg_WriteShort(msg.cMessage, C.tibems_short(value))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

// CreateBytesMsg creates a new, empty BytesMsg.
//
// Applications MUST call [BytesMsg.Close] when finished with a BytesMsg to avoid resource leaks.
func CreateBytesMsg() (*BytesMsg, error) {
	var message = BytesMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.tibemsBytesMsg_Create((*C.tibemsBytesMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

func (msg *BytesMsg) Reset() error {
	status := C.tibemsBytesMsg_Reset(msg.cMessage)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

// Close cleans up resources used by a BytesMsg. A BytesMsg MUST NOT be used again after Close has been called.
func (message *BytesMsg) Close() error {
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

type jsonBytesMsgBody struct {
	Type  string `json:"type"`
	Bytes []byte `json:"bytes"`
}

type jsonBytesMsg struct {
	jsonMsg
	Body jsonBytesMsgBody `json:"body"`
}

func (msg *BytesMsg) MarshalJSON() ([]byte, error) {
	headers, err := marshalHeader(msg)
	if err != nil {
		return nil, err
	}
	properties, err := marshalProperties(msg)
	if err != nil {
		return nil, err
	}
	bytes, err := msg.GetBytes()
	if err != nil {
		return nil, err
	}
	jsonMessage := jsonBytesMsg{
		jsonMsg: jsonMsg{
			Header:     headers,
			Properties: properties,
		},
		Body: jsonBytesMsgBody{
			Type:  "BYTES",
			Bytes: bytes,
		},
	}
	return json.Marshal(jsonMessage)
}
