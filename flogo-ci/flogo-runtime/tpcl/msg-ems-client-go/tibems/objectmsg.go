package tibems

/*
#include <tibems/tibems.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

// Ensure ObjectMsg implements Message
var _ Message = (*ObjectMsg)(nil)

type ObjectMsg struct {
	Msg
}

func (msg *ObjectMsg) SetObjectBytes(bytes []byte) error {
	status := C.tibemsObjectMsg_SetObjectBytes(msg.cMessage, unsafe.Pointer(&bytes[0]), C.tibems_uint(len(bytes)))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (msg *ObjectMsg) GetObjectBytes() ([]byte, error) {
	var cValue unsafe.Pointer
	var cSize C.tibems_uint
	status := C.tibemsObjectMsg_GetObjectBytes(msg.cMessage, &cValue, &cSize)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	value := C.GoBytes(cValue, C.int(cSize))
	return value, nil
}

// CreateObjectMsg creates a new, empty ObjectMsg.
//
// Applications MUST call [ObjectMsg.Close] when finished with an ObjectMsg to avoid resource leaks.
func CreateObjectMsg() (*ObjectMsg, error) {
	var message = ObjectMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.tibemsObjectMsg_Create((*C.tibemsObjectMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// Close cleans up resources used by an ObjectMsg. An ObjectMsg MUST NOT be used again after Close has been called.
func (message *ObjectMsg) Close() error {
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

type jsonObjectMsgBody struct {
	Type  string `json:"type"`
	Bytes []byte `json:"bytes"`
}

type jsonObjectMsg struct {
	jsonMsg
	Body jsonObjectMsgBody `json:"body"`
}

func (msg *ObjectMsg) MarshalJSON() ([]byte, error) {
	headers, err := marshalHeader(msg)
	if err != nil {
		return nil, err
	}
	properties, err := marshalProperties(msg)
	if err != nil {
		return nil, err
	}
	bytes, err := msg.GetObjectBytes()
	if err != nil {
		return nil, err
	}
	jsonMessage := jsonObjectMsg{
		jsonMsg: jsonMsg{
			Header:     headers,
			Properties: properties,
		},
		Body: jsonObjectMsgBody{
			Type:  "OBJECT",
			Bytes: bytes,
		},
	}
	return json.Marshal(jsonMessage)
}
