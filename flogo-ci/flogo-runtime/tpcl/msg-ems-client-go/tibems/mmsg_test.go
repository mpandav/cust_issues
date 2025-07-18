package tibems

import (
	"reflect"
	"testing"
)

func Test_CreateMapMsg(t *testing.T) {
	msg, err := CreateMapMsg()
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

func TestMapMsg_SetInt(t *testing.T) {
	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf(err.Error())
	}
	if msg == nil {
		t.Errorf("msg was nil")
	}

	msg.SetInt("myInt", 42)
	myInt, err := msg.GetInt("myInt")
	if err != nil {
		t.Errorf(err.Error())
	}
	if myInt != 42 {
		t.Errorf("myInt was not 42")
	}

	err = msg.ClearBody()
	if err != nil {
		t.Errorf(err.Error())
	}
	myInt, err = msg.GetInt("myInt")
	if err == nil {
		t.Errorf("found myInt even after clearing message")
	}

	msg.SetInt("myInt", 42)
	err = msg.flushPending()
	if err != nil {
		t.Errorf(err.Error())
	}
	myInt, err = msg.GetInt("myInt")
	if err != nil {
		t.Errorf(err.Error())
	}
	if myInt != 42 {
		t.Errorf("myInt was not 42")
	}

	msg.clearFieldMaps()
	myInt, err = msg.GetInt("myInt")
	if err != nil {
		t.Errorf(err.Error())
	}
	if myInt != 42 {
		t.Errorf("myInt was not 42")
	}

	myInt, err = msg.GetInt("myInt")
	if err != nil {
		t.Errorf(err.Error())
	}
	if myInt != 42 {
		t.Errorf("myInt was not 42")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestMapMsg_UnmarshalPrimitives(t *testing.T) {

	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	msg.SetShort("int16", 123)
	msg.SetInt("intThirtyTwo", 456)
	msg.SetLong("int64", 789)
	msg.SetString("string", "myString")

	var TestStruct struct {
		Int16  int16  `msg:"int16"`
		Int32  int32  `msg:"intThirtyTwo"`
		Int64  int64  `msg:"int64"`
		String string `msg:"string"`
	}

	err = Unmarshal(msg, &TestStruct)
	if err != nil {
		t.Errorf("Error unmarshaling")
	}
	if TestStruct.Int16 != 123 {
		t.Errorf("Int16 was not 123")
	}
	if TestStruct.Int32 != 456 {
		t.Errorf("Int32 was not 456")
	}
	if TestStruct.Int64 != 789 {
		t.Errorf("Int64 was not 789")
	}
	if TestStruct.String != "myString" {
		t.Errorf("String was not myString")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestMapMsg_UnmarshalPtrToPrimitives(t *testing.T) {

	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	msg.SetShort("int16", 123)
	msg.SetInt("intThirtyTwo", 456)
	msg.SetLong("int64", 789)
	msg.SetString("string", "myString")

	var TestStruct struct {
		Int16  *int16  `msg:"int16"`
		Int32  *int32  `msg:"intThirtyTwo"`
		Int64  *int64  `msg:"int64"`
		String *string `msg:"string"`
	}

	err = Unmarshal(msg, &TestStruct)
	if err != nil {
		t.Errorf("Error unmarshaling")
	}
	if *TestStruct.Int16 != 123 {
		t.Errorf("Int16 was not 123")
	}
	if *TestStruct.Int32 != 456 {
		t.Errorf("Int32 was not 456")
	}
	if *TestStruct.Int64 != 789 {
		t.Errorf("Int64 was not 789")
	}
	if *TestStruct.String != "myString" {
		t.Errorf("String was not myString")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestMapMsg_UnmarshalPrimitiveArrays(t *testing.T) {
	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	msg.SetShortArray("int16array", []int16{1, 2, 3})
	msg.SetIntArray("intThirtyTwoArray", []int32{4, 5, 6})
	msg.SetLongArray("int64array", []int64{7, 8, 9})
	msg.SetBytes("bytesarray", []byte{10, 11, 12})

	var TestStruct struct {
		Int16Array []int16 `msg:"int16array"`
		Int32Array []int32 `msg:"intThirtyTwoArray"`
		Int64Array []int64 `msg:"int64array"`
		BytesArray []byte  `msg:"bytesarray"`
	}

	err = Unmarshal(msg, &TestStruct)
	if err != nil {
		t.Errorf("Error unmarshaling")
	}

	if !reflect.DeepEqual(TestStruct.Int16Array, []int16{1, 2, 3}) {
		t.Errorf("Int16Array was not [1,2,3]")
	}
	if !reflect.DeepEqual(TestStruct.Int32Array, []int32{4, 5, 6}) {
		t.Errorf("Int32Array was not [4,5,6]")
	}
	if !reflect.DeepEqual(TestStruct.Int64Array, []int64{7, 8, 9}) {
		t.Errorf("Int64Array was not [7,8,9]")
	}
	if !reflect.DeepEqual(TestStruct.BytesArray, []byte{10, 11, 12}) {
		t.Errorf("BytesArray was not [10,11,12]")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestMapMsg_SetAndGetLongArray(t *testing.T) {
	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	err = msg.SetLongArray("int64array", []int64{1, 2, 3})
	if err != nil {
		t.Errorf(err.Error())
	}
	bytes, err := msg.GetAsBytes()
	if err != nil {
		t.Errorf(err.Error())
	}
	deserializedMsg, err := CreateMsgFromBytes(bytes)
	if err != nil {
		t.Errorf(err.Error())
	}

	array, err := deserializedMsg.(*MapMsg).GetLongArray("int64array")
	if err != nil {
		t.Errorf(err.Error())
	}

	if !reflect.DeepEqual(array, []int64{1, 2, 3}) {
		t.Errorf("int64array was not [1,2,3]")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}

	err = deserializedMsg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestMapMsg_UnmarshalPtrToPrimitiveArrays(t *testing.T) {
	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	msg.SetShortArray("int16array", []int16{1, 2, 3})
	msg.SetIntArray("intThirtyTwoArray", []int32{4, 5, 6})
	msg.SetLongArray("int64array", []int64{7, 8, 9})
	msg.SetBytes("bytesarray", []byte{10, 11, 12})

	var TestStruct struct {
		Int16Array *[]int16 `msg:"int16array"`
		Int32Array *[]int32 `msg:"intThirtyTwoArray"`
		Int64Array *[]int64 `msg:"int64array"`
		BytesArray *[]byte  `msg:"bytesarray"`
	}

	err = Unmarshal(msg, &TestStruct)
	if err != nil {
		t.Errorf("Error unmarshaling")
	}

	if !reflect.DeepEqual(*TestStruct.Int16Array, []int16{1, 2, 3}) {
		t.Errorf("Int16Array was not [1,2,3]")
	}
	if !reflect.DeepEqual(*TestStruct.Int32Array, []int32{4, 5, 6}) {
		t.Errorf("Int32Array was not [4,5,6]")
	}
	if !reflect.DeepEqual(*TestStruct.Int64Array, []int64{7, 8, 9}) {
		t.Errorf("Int64Array was not [7,8,9]")
	}
	if !reflect.DeepEqual(*TestStruct.BytesArray, []byte{10, 11, 12}) {
		t.Errorf("BytesArray was not [10,11,12]")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestMapMsg_UnmarshalPtrToStruct(t *testing.T) {
	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	msg.SetShortArray("int16array", []int16{1, 2, 3})
	msg.SetIntArray("intThirtyTwoArray", []int32{4, 5, 6})
	msg.SetLongArray("int64array", []int64{7, 8, 9})
	msg.SetBytes("bytesarray", []byte{10, 11, 12})

	type TestStruct struct {
		Int16Array []int16 `msg:"int16array"`
		Int32Array []int32 `msg:"intThirtyTwoArray"`
		Int64Array []int64 `msg:"int64array"`
		BytesArray []byte  `msg:"bytesarray"`
	}

	var OuterStruct struct {
		Inner *TestStruct `msg:"inner"`
	}
	outerMsg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	outerMsg.SetMapMsg("inner", msg)

	err = Unmarshal(outerMsg, &OuterStruct)
	if err != nil {
		t.Errorf("Error unmarshaling")
	}

	if !reflect.DeepEqual(OuterStruct.Inner.Int16Array, []int16{1, 2, 3}) {
		t.Errorf("Int16Array was not [1,2,3]")
	}
	if !reflect.DeepEqual(OuterStruct.Inner.Int32Array, []int32{4, 5, 6}) {
		t.Errorf("Int32Array was not [4,5,6]")
	}
	if !reflect.DeepEqual(OuterStruct.Inner.Int64Array, []int64{7, 8, 9}) {
		t.Errorf("Int64Array was not [7,8,9]")
	}
	if !reflect.DeepEqual(OuterStruct.Inner.BytesArray, []byte{10, 11, 12}) {
		t.Errorf("BytesArray was not [10,11,12]")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
	err = outerMsg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}

}

func TestMapMsg_UnmarshalObjectArray(t *testing.T) {
	msg, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}

	type User struct {
		Name string `msg:"name"`
	}

	userMsg0, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	userMsg1, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	userMsg2, err := CreateMapMsg()
	if err != nil {
		t.Errorf("Message create failed")
	}
	userMsg0.SetString("name", "one")
	msg.SetMapMsg("u0", userMsg0)
	userMsg1.SetString("name", "two")
	msg.SetMapMsg("u1", userMsg1)
	userMsg2.SetString("name", "three")
	msg.SetMapMsg("u2", userMsg2)

	type UsersStruct struct {
		Users []User `msg:"u"`
	}

	var users UsersStruct

	err = Unmarshal(msg, &users)
	if err != nil {
		t.Errorf("Error unmarshaling")
	}

	if len(users.Users) != 3 {
		t.Errorf("Error unmarshaling")
	}
	if users.Users[0].Name != "one" {
		t.Errorf("User 0 not 'one'")
	}
	if users.Users[1].Name != "two" {
		t.Errorf("User 1 not 'two'")
	}
	if users.Users[2].Name != "three" {
		t.Errorf("User 2 not 'three'")
	}

	err = msg.Close()
	if err != nil {
		t.Errorf(err.Error())
	}
}
