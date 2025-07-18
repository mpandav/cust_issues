package tibems

/*
#include <tibems/tibems.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

// Ensure TextMsg implements Message
var _ Message = (*TextMsg)(nil)

type TextMsg struct {
	Msg
}

type jsonTextMsgBody struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type jsonTextMsg struct {
	jsonMsg
	Body jsonTextMsgBody `json:"body"`
}

func (msg *TextMsg) MarshalJSON() ([]byte, error) {
	headers, err := marshalHeader(msg)
	if err != nil {
		return nil, err
	}
	properties, err := marshalProperties(msg)
	if err != nil {
		return nil, err
	}
	text, err := msg.GetText()
	if err != nil {
		return nil, err
	}
	jsonMessage := jsonTextMsg{
		jsonMsg: jsonMsg{
			Header:     headers,
			Properties: properties,
		},
		Body: jsonTextMsgBody{
			Type: "TEXT",
			Text: text,
		},
	}
	return json.Marshal(jsonMessage)
}

// SetText sets the string body of a TextMsg.
func (msg *TextMsg) SetText(text string) error {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	status := C.tibemsTextMsg_SetText(msg.cMessage, cText)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

// GetText returns the body of a TextMsg as a string.
func (msg *TextMsg) GetText() (string, error) {
	var cValue *C.char
	status := C.tibemsTextMsg_GetText(msg.cMessage, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	value := C.GoString(cValue)
	return value, nil
}

// CreateTextMsg creates a new, empty TextMsg.
//
// Applications MUST call [TextMsg.Close] when finished with a TextMsg to avoid resource leaks.
func CreateTextMsg() (*TextMsg, error) {
	var message = TextMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.tibemsTextMsg_Create((*C.tibemsTextMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// Close cleans up resources used by a TextMsg. A TextMsg MUST NOT be used again after Close has been called.
func (message *TextMsg) Close() error {
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
