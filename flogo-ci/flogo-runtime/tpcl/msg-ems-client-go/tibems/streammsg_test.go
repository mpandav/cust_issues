package tibems

import (
	"encoding/json"
	"reflect"
	"testing"
)

func Test_CreateStreamMsg(t *testing.T) {
	msg, err := CreateStreamMsg()
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

func TestStreamMsg_WriteAndRead(t *testing.T) {
	msg, err := CreateStreamMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}

	err = msg.WriteBoolean(true)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteByte('z')
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteBytes([]byte{1, 2, 3})
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteDouble(1.23)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteFloat(4.56)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteInt(123)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteLong(456)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteShort(789)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteString("All done!")
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.Reset()
	if err != nil {
		t.Errorf(err.Error())
	}

	boolValue, err := msg.ReadBoolean()
	if err != nil {
		t.Errorf(err.Error())
	}
	if !boolValue {
		t.Errorf("bool wrong")
	}

	byteValue, err := msg.ReadByte()
	if err != nil {
		t.Errorf(err.Error())
	}
	if byteValue != 'z' {
		t.Errorf("byte wrong")
	}

	bytesValue, err := msg.ReadBytes()
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(bytesValue, []byte{1, 2, 3}) {
		t.Errorf("bytes wrong")
	}

	doubleValue, err := msg.ReadDouble()
	if err != nil {
		t.Errorf(err.Error())
	}
	if doubleValue != 1.23 {
		t.Errorf("double wrong")
	}

	floatValue, err := msg.ReadFloat()
	if err != nil {
		t.Errorf(err.Error())
	}
	if floatValue != 4.56 {
		t.Errorf("float wrong")
	}

	intValue, err := msg.ReadInt()
	if err != nil {
		t.Errorf(err.Error())
	}
	if intValue != 123 {
		t.Errorf("int wrong")
	}

	longValue, err := msg.ReadLong()
	if err != nil {
		t.Errorf(err.Error())
	}
	if longValue != 456 {
		t.Errorf("long wrong")
	}

	shortValue, err := msg.ReadShort()
	if err != nil {
		t.Errorf(err.Error())
	}
	if shortValue != 789 {
		t.Errorf("short wrong")
	}

	stringValue, err := msg.ReadString()
	if err != nil {
		t.Errorf(err.Error())
	}
	if stringValue != "All done!" {
		t.Errorf("string wrong")
	}

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf(string(jsonMessage))

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestStreamMsg_NestStreamMsg(t *testing.T) {
	msg, err := CreateStreamMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}

	nestedMsg, err := CreateStreamMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if nestedMsg == nil {
		t.Errorf("nestedMsg was nil")
	}

	err = msg.WriteBoolean(true)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteLongArray([]int64{1, 2, 3})
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.WriteIntArray([]int32{1, 2, 3})
	if err != nil {
		t.Errorf(err.Error())
	}

	err = nestedMsg.WriteString("It worked!")
	if err != nil {
		t.Errorf(err.Error())
	}
	err = nestedMsg.Reset()
	if err != nil {
		t.Errorf(err.Error())
	}
	err = msg.WriteStreamMsg(nestedMsg)
	if err != nil {
		t.Errorf(err.Error())
	}

	err = msg.Reset()
	if err != nil {
		t.Errorf(err.Error())
	}

	jsonMessage, err := json.Marshal(msg)
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf(string(jsonMessage))

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}
