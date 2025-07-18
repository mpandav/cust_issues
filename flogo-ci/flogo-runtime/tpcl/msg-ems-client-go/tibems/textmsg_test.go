package tibems

import (
	"encoding/json"
	"testing"
)

func Test_CreateTextMsg(t *testing.T) {
	msg, err := CreateTextMsg()
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

func TestTextMsg_SetText(t *testing.T) {
	msg, err := CreateTextMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}

	err = msg.SetText("It is a truth universally acknowledged...")
	if err != nil {
		t.Errorf(err.Error())
	}

	text, err := msg.GetText()
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "It is a truth universally acknowledged..." {
		t.Errorf("text was incorrect")
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

	text, err = msg.GetText()
	if err != nil {
		t.Errorf(err.Error())
	}
	if text != "" {
		t.Errorf("text was not cleared")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}
