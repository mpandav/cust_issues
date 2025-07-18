package tibems

import (
	"encoding/json"
	"testing"
)

func Test_CreateObjectMsg(t *testing.T) {
	msg, err := CreateObjectMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}
	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestObjectMsg_SetObjectBytes(t *testing.T) {
	msg, err := CreateObjectMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}

	err = msg.SetObjectBytes([]byte("It is a truth universally acknowledged..."))
	if err != nil {
		t.Errorf(err.Error())
	}

	bytes, err := msg.GetObjectBytes()
	if err != nil {
		t.Errorf(err.Error())
	}
	if string(bytes) != "It is a truth universally acknowledged..." {
		t.Errorf("bytes were incorrect")
	}

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf(string(jsonMessage))

	err = msg.ClearBody()
	if err != nil {
		t.Errorf(err.Error())
	}

	bytes, err = msg.GetObjectBytes()
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(bytes) > 0 {
		t.Errorf("bytes were not cleared")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}
