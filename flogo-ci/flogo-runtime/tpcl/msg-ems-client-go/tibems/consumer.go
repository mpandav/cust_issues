package tibems

/*
#include <inttypes.h>
#include <tibems/tibems.h>

extern void
goMsgCallback(
    tibemsMsg,
    uintptr_t);

static void
_cMsgCallback(
    tibemsMsgConsumer msgConsumer,
    tibemsMsg         msg,
    void*             closure)
{
  goMsgCallback(msg, (uintptr_t)closure);
}

// These one-line stub functions are needed because some EMS pointer types
// are not really pointers, but object IDs, and so can have values that are not
// seen by Go as valid. Go's runtime pointer checking will panic if it sees what
// it believes is an invalid pointer. We use uintptr_t's in place
// of EMS pointer types where needed in the arguments to these stub functions
// so Go doesn't check their values, and then cast them back to the required
// EMS pointer types to call the original EMS API function.
// Not all EMS pointer types are really object IDs; some really are pointers,
// so a stub function is not always needed for a given EMS API function.

static tibems_status
go_tibemsMsgConsumer_Close(uintptr_t consumer)
{
  return tibemsMsgConsumer_Close((tibemsMsgConsumer)consumer);
}

static tibems_status
go_tibemsMsgConsumer_Receive(
    uintptr_t  msgConsumer,
    tibemsMsg* msg)
{
  return tibemsMsgConsumer_Receive((tibemsMsgConsumer)msgConsumer, msg);
}

static tibems_status
go_tibemsMsgConsumer_ReceiveTimeout(
    uintptr_t   msgConsumer,
    tibemsMsg*  msg,
    tibems_long timeout)
{
  return tibemsMsgConsumer_ReceiveTimeout((tibemsMsgConsumer)msgConsumer, msg, timeout);
}

static tibems_status
go_tibemsMsgConsumer_ReceiveNoWait(
    uintptr_t  msgConsumer,
    tibemsMsg* msg)
{
  return tibemsMsgConsumer_ReceiveNoWait((tibemsMsgConsumer)msgConsumer, msg);
}

static tibems_status
_cSetMsgCallback(
    uintptr_t consumer,
    uintptr_t handle)
{
  return tibemsMsgConsumer_SetMsgListener((tibemsMsgConsumer)consumer, _cMsgCallback, (void*)handle);
}

static void
_callMsgCallback(
    uintptr_t consumerHandle,
    uintptr_t cb,
    uintptr_t msgHandle,
    uintptr_t closure)
{
  ((tibemsMsgCallback)cb)((tibemsMsgConsumer)consumerHandle, (tibemsMsg)msgHandle, (void*)closure);
}

*/
import "C"
import (
	"github.com/cockroachdb/errors"
	"runtime/cgo"
)

type MsgConsumer struct {
	cConsumer         C.uintptr_t
	destination       *Destination
	noLocal           bool
	selector          string
	msgCallbackHandle *cgo.Handle
	msgCallback       MsgCallback
}

type msgCallbackInfo struct {
	callback MsgCallback
}
type MsgCallback func(msg Message)

// Used for testing.
//
//export createCMsgCallback
func createCMsgCallback(consumerHandle C.uintptr_t, cFunc C.uintptr_t, cClosure C.uintptr_t) uintptr {
	var fn MsgCallback
	fn = func(msg Message) {
		msgHandle := cgo.NewHandle(msg)
		C._callMsgCallback(consumerHandle, cFunc, C.uintptr_t(msgHandle), cClosure)
		msgHandle.Delete()
	}
	return uintptr(cgo.NewHandle(fn))
}

//export goMsgCallback
func goMsgCallback(cMsg C.tibemsMsg, callbackInfoHandle C.uintptr_t) {
	handle := cgo.Handle(callbackInfoHandle)
	callback := handle.Value().(msgCallbackInfo)
	msg, err := instantiateSpecificMessageType(cMsg)
	if err != nil {
		return
	}
	callback.callback(msg)
}

func (consumer *MsgConsumer) GetMsgListener() (MsgCallback, error) {
	return consumer.msgCallback, nil
}

func (consumer *MsgConsumer) SetMsgListener(callback MsgCallback) error {
	cbInfo := msgCallbackInfo{
		callback: callback,
	}

	newHandle := cgo.NewHandle(cbInfo)
	status := C._cSetMsgCallback(consumer.cConsumer, C.uintptr_t(newHandle))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	if consumer.msgCallbackHandle != nil {
		consumer.msgCallbackHandle.Delete()
		consumer.msgCallbackHandle = nil
	}
	consumer.msgCallbackHandle = &newHandle
	consumer.msgCallback = callback

	return nil
}

func (consumer *MsgConsumer) GetDestination() (*Destination, error) {
	return consumer.destination, nil
}

func (consumer *MsgConsumer) GetNoLocal() (bool, error) {
	return consumer.noLocal, nil
}

func (consumer *MsgConsumer) GetMsgSelector() (string, error) {
	return consumer.selector, nil
}

func (consumer *MsgConsumer) Receive() (Message, error) {
	var cMessage C.tibemsMsg
	status := C.go_tibemsMsgConsumer_Receive(consumer.cConsumer, &cMessage)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return instantiateSpecificMessageType(cMessage)
}

func (consumer *MsgConsumer) ReceiveTimeout(timeout int64) (Message, error) {
	var cMessage C.tibemsMsg
	status := C.go_tibemsMsgConsumer_ReceiveTimeout(consumer.cConsumer, &cMessage, C.tibems_long(timeout))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return instantiateSpecificMessageType(cMessage)
}

func (consumer *MsgConsumer) ReceiveNoWait() (Message, error) {
	var cMessage C.tibemsMsg
	status := C.go_tibemsMsgConsumer_ReceiveNoWait(consumer.cConsumer, &cMessage)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return instantiateSpecificMessageType(cMessage)
}

func (consumer *MsgConsumer) Close() error {
	status := C.go_tibemsMsgConsumer_Close(consumer.cConsumer)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	if consumer.msgCallbackHandle != nil {
		consumer.msgCallbackHandle.Delete()
		consumer.msgCallbackHandle = nil
	}
	return nil
}

func instantiateSpecificMessageType(cMessage C.tibemsMsg) (Message, error) {
	var msgType C.tibemsMsgType
	status := C.tibemsMsg_GetBodyType(cMessage, &msgType)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	switch MsgType(msgType) {
	case MsgTypeMessage:
		return &Msg{cMessage: cMessage}, nil
	case MsgTypeMapMessage:
		return &MapMsg{Msg: Msg{cMessage: cMessage}}, nil
	case MsgTypeTextMessage:
		return &TextMsg{Msg: Msg{cMessage: cMessage}}, nil
	case MsgTypeBytesMessage:
		return &BytesMsg{Msg: Msg{cMessage: cMessage}}, nil
	case MsgTypeObjectMessage:
		return &ObjectMsg{Msg: Msg{cMessage: cMessage}}, nil
	case MsgTypeStreamMessage:
		return &StreamMsg{Msg: Msg{cMessage: cMessage}}, nil
	default:
		return nil, errors.Wrap(ErrNotImplemented, "Received unsupported message type")
	}
}
