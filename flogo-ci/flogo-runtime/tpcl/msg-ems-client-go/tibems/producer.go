package tibems

/*
#include <inttypes.h>
#include <tibems/tibems.h>

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
go_tibemsMsgProducer_SendToDestinationEx(
    uintptr_t          msgProducer,
    uintptr_t          destination,
    tibemsMsg          msg,
    tibemsDeliveryMode deliveryMode,
    tibems_int         priority,
    tibems_long        timeToLive)
{
  return tibemsMsgProducer_SendToDestinationEx(
      (tibemsMsgProducer)msgProducer, (tibemsDestination)destination, msg, deliveryMode, priority, timeToLive);
}

static tibems_status
go_tibemsMsgProducer_Close(uintptr_t msgProducer)
{
  return tibemsMsgProducer_Close((tibemsMsgProducer)msgProducer);
}

static tibems_status
go_tibemsMsgProducer_SetDeliveryDelay(
    uintptr_t   msgProducer,
    tibems_long deliveryDelay)
{
  return tibemsMsgProducer_SetDeliveryDelay((tibemsMsgProducer)msgProducer, deliveryDelay);
}

static tibems_status
go_tibemsMsgProducer_SetDeliveryMode(
    uintptr_t  msgProducer,
    tibems_int deliveryMode)
{
  return tibemsMsgProducer_SetDeliveryMode((tibemsMsgProducer)msgProducer, deliveryMode);
}

static tibems_status
go_tibemsMsgProducer_SetDisableMessageID(
    uintptr_t   msgProducer,
    tibems_bool doDisableMessageID)
{
  return tibemsMsgProducer_SetDisableMessageID((tibemsMsgProducer)msgProducer, doDisableMessageID);
}

static tibems_status
go_tibemsMsgProducer_SetDisableMessageTimestamp(
    uintptr_t   msgProducer,
    tibems_bool doDisableMessageTimeStamp)
{
  return tibemsMsgProducer_SetDisableMessageTimestamp((tibemsMsgProducer)msgProducer, doDisableMessageTimeStamp);
}

static tibems_status
go_tibemsMsgProducer_SetPriority(
    uintptr_t  msgProducer,
    tibems_int priority)
{
  return tibemsMsgProducer_SetPriority((tibemsMsgProducer)msgProducer, priority);
}

static tibems_status
go_tibemsMsgProducer_SetTimeToLive(
    uintptr_t   msgProducer,
    tibems_long timeToLive)
{
  return tibemsMsgProducer_SetTimeToLive((tibemsMsgProducer)msgProducer, timeToLive);
}

static tibems_status
go_tibemsMsgProducer_Send(
    uintptr_t msgProducer,
    tibemsMsg msg)
{
  return tibemsMsgProducer_Send((tibemsMsgProducer)msgProducer, msg);
}

static tibems_status
go_tibemsMsgProducer_SetNPSendCheckMode(
    uintptr_t         producer,
    tibemsNpCheckMode mode)
{
  return tibemsMsgProducer_SetNPSendCheckMode((tibemsMsgProducer)producer, mode);
}

static tibems_status
go_tibemsMsgProducer_SendEx(
    uintptr_t   msgProducer,
    tibemsMsg   msg,
    tibems_int  deliveryMode,
    tibems_int  priority,
    tibems_long timeToLive)
{
  return tibemsMsgProducer_SendEx((tibemsMsgProducer)msgProducer, msg, deliveryMode, priority, timeToLive);
}

extern void
goMsgCompletionCallback(
    tibems_status,
    uintptr_t);

static void
_cMsgCompletionCallback(
    tibemsMsg     msg,
    tibems_status status,
    void*         closure)
{
  goMsgCompletionCallback(status, (uintptr_t)closure);
}

static tibems_status
_asyncSend(
    uintptr_t msgProducer,
    uintptr_t destination,
    tibemsMsg message,
    uintptr_t asyncSendClosure)
{
  if (destination == 0)
  {
    return tibemsMsgProducer_AsyncSend(
        (tibemsMsgProducer)msgProducer, message, _cMsgCompletionCallback, (void*)asyncSendClosure);
  }
  else
  {
    return tibemsMsgProducer_AsyncSendToDestination(
        (tibemsMsgProducer)msgProducer,
        (tibemsDestination)destination,
        message,
        _cMsgCompletionCallback,
        (void*)asyncSendClosure);
  }
}

static tibems_status
_asyncSendEx(
    uintptr_t   msgProducer,
    uintptr_t   destination,
    tibemsMsg   message,
    tibems_int  deliveryMode,
    tibems_int  priority,
    tibems_long timeToLive,
    uintptr_t   asyncSendClosure)
{
  if (destination == 0)
  {
    return tibemsMsgProducer_AsyncSendEx(
        (tibemsMsgProducer)msgProducer,
        message,
        deliveryMode,
        priority,
        timeToLive,
        _cMsgCompletionCallback,
        (void*)asyncSendClosure);
  }
  else
  {
    return tibemsMsgProducer_AsyncSendToDestinationEx(
        (tibemsMsgProducer)msgProducer,
        (tibemsDestination)destination,
        message,
        deliveryMode,
        priority,
        timeToLive,
        _cMsgCompletionCallback,
        (void*)asyncSendClosure);
  }
}

*/
import "C"
import "runtime/cgo"

type MsgProducer struct {
	cProducer               C.uintptr_t
	deliveryDelay           int64
	deliveryMode            DeliveryMode
	destination             *Destination
	disableMessageID        bool
	disableMessageTimestamp bool
	priority                int32
	timeToLive              int64
	nonPersistentCheckMode  NPSendCheckMode
}

type msgCompletionCallbackInfo struct {
	callback MsgCompletionCallback
	msg      Message
}
type MsgCompletionCallback func(msg Message, err error)

//export goMsgCompletionCallback
func goMsgCompletionCallback(status C.tibems_status, callbackInfoHandle C.uintptr_t) {
	handle := cgo.Handle(callbackInfoHandle)
	callback := handle.Value().(msgCompletionCallbackInfo)
	var err error = nil
	if status != tibems_OK {
		err = statusToError(status)
	}
	callback.callback(callback.msg, err)
	handle.Delete()
}

// AsyncSendEx queues a message to be sent asynchronously and calls the given callback after the message has been sent
// (or has failed to be sent).
func (producer *MsgProducer) AsyncSend(msg Message, asyncSendCallback MsgCompletionCallback, options *SendOptions) error {

	err := msg.flushPending()
	if err != nil {
		return err
	}
	cbInfo := msgCompletionCallbackInfo{
		callback: asyncSendCallback,
		msg:      msg,
	}

	newHandle := cgo.NewHandle(cbInfo)

	var status C.tibems_status
	if options != nil {
		var cDestination C.uintptr_t = 0
		if options.Destination != nil {
			cDestination = options.Destination.cDestination
		}
		status = C._asyncSendEx(producer.cProducer, cDestination, msg.getCMessage(), C.tibems_int(options.DeliveryMode),
			C.tibems_int(options.Priority), C.tibems_long(options.TimeToLive), C.uintptr_t(newHandle))
	} else {
		status = C._asyncSend(producer.cProducer, 0, msg.getCMessage(), C.uintptr_t(newHandle))
	}

	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

// SendOptions are passed to AsyncSendEx or SendEx functions to modify the behavior of messages on a per-send basis.
type SendOptions struct {
	Destination  *Destination
	DeliveryMode DeliveryMode
	Priority     int32
	TimeToLive   int64
}

// Used for testing.
//
//export createCSendOptions
func createCSendOptions(producerHandle C.uintptr_t, destinationHandle C.uintptr_t, mode int32, priority int32, timeToLive int64) C.uintptr_t {
	producer := cgo.Handle(producerHandle).Value().(*MsgProducer)
	var destination *Destination
	if destinationHandle != 0 {
		destination = cgo.Handle(destinationHandle).Value().(*Destination)
	}
	if mode == -1 {
		deliveryMode, _ := producer.GetDeliveryMode()
		mode = int32(deliveryMode)
	}
	if priority == -1 {
		priority, _ = producer.GetPriority()
	}
	if timeToLive == -1 {
		timeToLive, _ = producer.GetTimeToLive()
	}

	return C.uintptr_t(cgo.NewHandle(&SendOptions{
		Destination:  destination,
		DeliveryMode: DeliveryMode(mode),
		Priority:     priority,
		TimeToLive:   timeToLive,
	}))
}

// Used for testing.
//
//export destroyCSendOptions
func destroyCSendOptions(optionsHandle C.uintptr_t) {
	cgo.Handle(optionsHandle).Delete()
}

func (producer *MsgProducer) Send(msg Message, options *SendOptions) error {

	err := msg.flushPending()
	if err != nil {
		return err
	}

	var status C.tibems_status
	if options != nil {
		if options.Destination != nil {
			status = C.go_tibemsMsgProducer_SendToDestinationEx(producer.cProducer, options.Destination.cDestination, msg.getCMessage(), C.tibemsDeliveryMode(options.DeliveryMode), C.tibems_int(options.Priority), C.tibems_long(options.TimeToLive))
		} else {
			status = C.go_tibemsMsgProducer_SendEx(producer.cProducer, msg.getCMessage(), C.int(options.DeliveryMode), C.tibems_int(options.Priority), C.tibems_long(options.TimeToLive))
		}
	} else {
		status = C.go_tibemsMsgProducer_Send(producer.cProducer, msg.getCMessage())
	}

	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (producer *MsgProducer) GetDestination() (*Destination, error) {
	return producer.destination, nil
}

func (producer *MsgProducer) GetDisableMessageID() (bool, error) {
	return producer.disableMessageID, nil
}

func (producer *MsgProducer) SetDisableMessageID(disable bool) error {
	var cValue C.tibems_bool
	if disable {
		cValue = 1
	} else {
		cValue = 0
	}
	status := C.go_tibemsMsgProducer_SetDisableMessageID(producer.cProducer, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.disableMessageID = disable
	return nil
}

func (producer *MsgProducer) GetDisableMessageTimestamp() (bool, error) {
	return producer.disableMessageTimestamp, nil
}

func (producer *MsgProducer) SetDisableMessageTimestamp(disable bool) error {
	var cValue C.tibems_bool
	if disable {
		cValue = 1
	} else {
		cValue = 0
	}
	status := C.go_tibemsMsgProducer_SetDisableMessageTimestamp(producer.cProducer, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.disableMessageTimestamp = disable
	return nil
}

func (producer *MsgProducer) GetDeliveryMode() (DeliveryMode, error) {
	return producer.deliveryMode, nil
}

func (producer *MsgProducer) SetDeliveryMode(mode DeliveryMode) error {
	status := C.go_tibemsMsgProducer_SetDeliveryMode(producer.cProducer, C.tibems_int(mode))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.deliveryMode = mode
	return nil
}

func (producer *MsgProducer) GetDeliveryDelay() (int64, error) {
	return producer.deliveryDelay, nil
}

func (producer *MsgProducer) SetDeliveryDelay(delay int64) error {
	status := C.go_tibemsMsgProducer_SetDeliveryDelay(producer.cProducer, C.tibems_long(delay))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.deliveryDelay = delay
	return nil
}

func (producer *MsgProducer) GetTimeToLive() (int64, error) {
	return producer.timeToLive, nil
}

func (producer *MsgProducer) SetTimeToLive(timeToLive int64) error {
	status := C.go_tibemsMsgProducer_SetTimeToLive(producer.cProducer, C.tibems_long(timeToLive))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.timeToLive = timeToLive
	return nil
}

func (producer *MsgProducer) GetPriority() (int32, error) {
	return producer.priority, nil
}

func (producer *MsgProducer) SetPriority(priority int32) error {
	status := C.go_tibemsMsgProducer_SetPriority(producer.cProducer, C.tibems_int(priority))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.priority = priority
	return nil
}

func (producer *MsgProducer) GetNPSendCheckMode() (NPSendCheckMode, error) {
	return producer.nonPersistentCheckMode, nil
}

func (producer *MsgProducer) SetNPSendCheckMode(mode NPSendCheckMode) error {
	status := C.go_tibemsMsgProducer_SetNPSendCheckMode(producer.cProducer, C.tibemsNpCheckMode(mode))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	producer.nonPersistentCheckMode = mode
	return nil
}

func (producer *MsgProducer) Close() error {
	status := C.go_tibemsMsgProducer_Close(producer.cProducer)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}
