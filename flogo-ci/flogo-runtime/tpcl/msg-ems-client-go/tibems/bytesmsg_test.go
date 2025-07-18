package tibems

import "testing"

func Test_CreateBytesMsg(t *testing.T) {
	msg, err := CreateBytesMsg()
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

func TestBytesMsg_SetBytes(t *testing.T) {
	msg, err := CreateBytesMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}

	err = msg.SetBytes([]byte("It is a truth universally acknowledged..."))
	if err != nil {
		t.Errorf(err.Error())
	}

	bytes, err := msg.GetBytes()
	if err != nil {
		t.Errorf(err.Error())
	}
	if string(bytes) != "It is a truth universally acknowledged..." {
		t.Errorf("bytes were incorrect")
	}

	err = msg.ClearBody()
	if err != nil {
		t.Errorf(err.Error())
	}

	bytes, err = msg.GetBytes()
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
