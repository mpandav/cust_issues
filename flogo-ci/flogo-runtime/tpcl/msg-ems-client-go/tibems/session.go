package tibems

/*
#include <stdint.h>
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
go_tibemsSession_CreateTemporaryQueue(
    uintptr_t  sess,
    uintptr_t* tmpQueue)
{
  return tibemsSession_CreateTemporaryQueue((tibemsSession)sess, (tibemsTemporaryQueue*)tmpQueue);
}

static tibems_status
go_tibemsSession_CreateTemporaryTopic(
    uintptr_t  sess,
    uintptr_t* tmpTopic)
{
  return tibemsSession_CreateTemporaryTopic((tibemsSession)sess, (tibemsTemporaryTopic*)tmpTopic);
}

static tibems_status
go_tibemsSession_Close(uintptr_t session)
{
  return tibemsSession_Close((tibemsSession)session);
}

static tibems_status
go_tibemsSession_Commit(uintptr_t session)
{
  return tibemsSession_Commit((tibemsSession)session);
}

static tibems_status
go_tibemsSession_CreateConsumer(
    uintptr_t   session,
    uintptr_t*  consumer,
    uintptr_t   destination,
    const char* optionalSelector,
    tibems_bool noLocal)
{
  return tibemsSession_CreateConsumer(
      (tibemsSession)session, (tibemsMsgConsumer*)consumer, (tibemsDestination)destination, optionalSelector, noLocal);
}

static tibems_status
go_tibemsSession_CreateSharedConsumer(
    uintptr_t   session,
    uintptr_t*  consumer,
    uintptr_t   topic,
    const char* sharedSubscriptionName,
    const char* optionalSelector)
{
  return tibemsSession_CreateSharedConsumer(
      (tibemsSession)session,
      (tibemsMsgConsumer*)consumer,
      (tibemsTopic)topic,
      sharedSubscriptionName,
      optionalSelector);
}

static tibems_status
go_tibemsSession_CreateDurableSubscriber(
    uintptr_t   session,
    uintptr_t*  topicSubscriber,
    uintptr_t   topic,
    const char* subscriberName,
    const char* optionalSelector,
    tibems_bool noLocal)
{
  return tibemsSession_CreateDurableSubscriber(
      (tibemsSession)session,
      (tibemsTopicSubscriber*)topicSubscriber,
      (tibemsTopic)topic,
      subscriberName,
      optionalSelector,
      noLocal);
}

static tibems_status
go_tibemsSession_CreateSharedDurableConsumer(
    uintptr_t   session,
    uintptr_t*  consumer,
    uintptr_t   topic,
    const char* durableName,
    const char* optionalSelector)
{
  return tibemsSession_CreateSharedDurableConsumer(
      (tibemsSession)session, (tibemsMsgConsumer*)consumer, (tibemsTopic)topic, durableName, optionalSelector);
}

static tibems_status
go_tibemsSession_CreateBrowser(
    uintptr_t   session,
    uintptr_t*  browser,
    uintptr_t   queue,
    const char* optionalSelector)
{
  return tibemsSession_CreateBrowser(
      (tibemsSession)session, (tibemsQueueBrowser*)browser, (tibemsQueue)queue, optionalSelector);
}

static tibems_status
go_tibemsSession_CreateProducer(
    uintptr_t  session,
    uintptr_t* producer,
    uintptr_t  destination)
{
  return tibemsSession_CreateProducer(
      (tibemsSession)session, (tibemsMsgProducer*)producer, (tibemsDestination)destination);
}

static tibems_status
go_tibemsSession_CreateBytesMessage(
    uintptr_t       session,
    tibemsBytesMsg* bytesMsg)
{
  return tibemsSession_CreateBytesMessage((tibemsSession)session, bytesMsg);
}

static tibems_status
go_tibemsSession_CreateMapMessage(
    uintptr_t     session,
    tibemsMapMsg* mapMsg)
{
  return tibemsSession_CreateMapMessage((tibemsSession)session, mapMsg);
}

static tibems_status
go_tibemsSession_CreateMessage(
    uintptr_t  session,
    tibemsMsg* msg)
{
  return tibemsSession_CreateMessage((tibemsSession)session, msg);
}

static tibems_status
go_tibemsSession_CreateStreamMessage(
    uintptr_t        session,
    tibemsStreamMsg* streamMsg)
{
  return tibemsSession_CreateStreamMessage((tibemsSession)session, streamMsg);
}

static tibems_status
go_tibemsSession_CreateTextMessage(
    uintptr_t      session,
    tibemsTextMsg* textMsg)
{
  return tibemsSession_CreateTextMessage((tibemsSession)session, textMsg);
}

static tibems_status
go_tibemsSession_Recover(uintptr_t session)
{
  return tibemsSession_Recover((tibemsSession)session);
}

static tibems_status
go_tibemsSession_Rollback(uintptr_t session)
{
  return tibemsSession_Rollback((tibemsSession)session);
}

static tibems_status
go_tibemsSession_Unsubscribe(
    uintptr_t   session,
    const char* subscriberName)
{
  return tibemsSession_Unsubscribe((tibemsSession)session, subscriberName);
}

static tibems_status
go_tibemsXASession_Close(uintptr_t session)
{
  return tibemsXASession_Close((tibemsSession)session);
}

static tibems_status
go_tibemsMsgRequestor_Create(
    uintptr_t  session,
    uintptr_t* msgRequestor,
    uintptr_t  destination)
{
  return tibemsMsgRequestor_Create(
      (tibemsSession)session, (tibemsMsgRequestor*)msgRequestor, (tibemsDestination)destination);
}

static tibems_status
go_tibemsMsgProducer_GetDeliveryDelay(
    uintptr_t    msgProducer,
    tibems_long* deliveryDelay)
{
  return tibemsMsgProducer_GetDeliveryDelay((tibemsMsgProducer)msgProducer, deliveryDelay);
}

static tibems_status
go_tibemsMsgProducer_GetDeliveryMode(
    uintptr_t   msgProducer,
    tibems_int* deliveryMode)
{
  return tibemsMsgProducer_GetDeliveryMode((tibemsMsgProducer)msgProducer, deliveryMode);
}

static tibems_status
go_tibemsMsgProducer_GetDisableMessageID(
    uintptr_t    msgProducer,
    tibems_bool* doDisableMessageID)
{
  return tibemsMsgProducer_GetDisableMessageID((tibemsMsgProducer)msgProducer, doDisableMessageID);
}

static tibems_status
go_tibemsMsgProducer_GetDisableMessageTimestamp(
    uintptr_t    msgProducer,
    tibems_bool* doDisableMessageTimeStamp)
{
  return tibemsMsgProducer_GetDisableMessageTimestamp((tibemsMsgProducer)msgProducer, doDisableMessageTimeStamp);
}

static tibems_status
go_tibemsMsgProducer_GetPriority(
    uintptr_t   msgProducer,
    tibems_int* priority)
{
  return tibemsMsgProducer_GetPriority((tibemsMsgProducer)msgProducer, priority);
}

static tibems_status
go_tibemsMsgProducer_GetTimeToLive(
    uintptr_t    msgProducer,
    tibems_long* timeToLive)
{
  return tibemsMsgProducer_GetTimeToLive((tibemsMsgProducer)msgProducer, timeToLive);
}

static tibems_status
go_tibemsMsgProducer_GetNPSendCheckMode(
    uintptr_t          producer,
    tibemsNpCheckMode* mode)
{
  return tibemsMsgProducer_GetNPSendCheckMode((tibemsMsgProducer)producer, mode);
}

*/
import "C"
import (
	"unsafe"
)

// # Purpose
//
// A Session provides an organizing context for message activity.
//
// # Remarks
//
// Sessions combine several roles:
//   - Create message producers and consumers
//   - Create message objects
//   - Create temporary destinations
//   - Create dynamic destinations
//   - Create queue browsers
//   - Serialize for inbound and outbound messages
//   - Serialize for asynchronous message events (or message listeners) of its consumer objects
//   - Cache inbound messages (until the program acknowledges them).
//   - Transaction support (when enabled).
//
// # Single Thread
//
// The Jakarta Messaging specification restricts programs to use each session within a single thread.
//
// # Associated Objects
//
// The same single-thread restriction applies to objects associated with a session—namely, messages, message consumers,
// durable subscribers, message producers, queue browsers, and temporary destinations (however, static and dynamic
// destinations are exempt from this restriction).
//
// # Corollary
//
// One consequence of this rule is that all the consumers of a session must deliver messages in the same mode—either
// synchronously or asynchronously.
//
// # Asynchronous
//
// In asynchronous delivery, the program registers message handler events or message listeners with the session’s
// consumer objects. An internal dispatcher thread delivers messages to those event handlers or listeners (in all the
// session’s consumer objects). No other thread may use the session (nor objects created by the session).
//
// # Synchronous
//
// In synchronous delivery, the program explicitly begins a thread for the session. That thread processes inbound
// messages and produces outbound messages, serializing this activity among the session’s producers and consumers.
// Functions that request the next message (such as [MsgConsumer.Receive]) can organize the thread’s activity.
//
// # Close
//
// The only exception to the rule restricting session calls to a single thread is the function [Session.Close];
// programs can call Close from any thread at any time.
//
// # Transactions
//
// A session has either transaction or non-transaction semantics. When a program specifies transaction semantics, the
// session object cooperates with the server, and all messages that flow through the session become part of a
// transaction.
//
//   - When the program calls [Session.Commit], the session acknowledges all inbound messages in the current
//     transaction, and the server delivers all outbound messages in the current transaction to their destinations.
//   - If the program calls [Session.Rollback], the session recovers all inbound messages in the current transaction
//     (so the program can consume them in a new transaction), and the server destroys all outbound messages in the
//     current transaction.
//
// After these actions, both Commit and Rollback immediately begin a new transaction.
type Session struct {
	cSession   C.uintptr_t
	ackMode    AcknowledgeMode
	transacted bool
	isXA       bool
}

// Ensure Session implements EMSSession
var _ EMSSession = (*Session)(nil)

type EMSSession interface {
	GetAcknowledgeMode() (AcknowledgeMode, error)
	GetTransacted() (bool, error)
	Commit() error
	Recover() error
	Rollback() error
	CreateTemporaryQueue() (*Destination, error)
	CreateTemporaryTopic() (*Destination, error)
	CreateBrowser(queue *Queue, selector string) (*QueueBrowser, error)
	CreateConsumer(destination *Destination, selector string, noLocal bool) (*MsgConsumer, error)
	CreateSharedConsumer(topic *Topic, sharedSubscriptionName string, selector string) (*MsgConsumer, error)
	CreateDurableSubscriber(topic *Topic, durableName string, selector string, noLocal bool) (*MsgConsumer, error)
	CreateSharedDurableConsumer(topic *Topic, durableName string, selector string) (*MsgConsumer, error)
	CreateRequestor(destination *Destination) (*MsgRequestor, error)
	CreateProducer(destination *Destination) (*MsgProducer, error)
	CreateBytesMessage() (*BytesMsg, error)
	CreateMapMessage() (*MapMsg, error)
	CreateMessage() (*Msg, error)
	CreateStreamMessage() (*StreamMsg, error)
	CreateTextMessage() (*TextMsg, error)
	Unsubscribe(durableName string) error
	Close() error
}

func (session *Session) GetAcknowledgeMode() (AcknowledgeMode, error) {
	return session.ackMode, nil
}

func (session *Session) GetTransacted() (bool, error) {
	return session.transacted, nil
}

func (session *Session) Commit() error {
	status := C.go_tibemsSession_Commit(session.cSession)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (session *Session) Recover() error {
	status := C.go_tibemsSession_Recover(session.cSession)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (session *Session) Rollback() error {
	status := C.go_tibemsSession_Rollback(session.cSession)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (session *Session) CreateTemporaryQueue() (*Destination, error) {

	var temporaryQueue = Destination{
		cDestination:    0,
		cSession:        session.cSession,
		destinationType: DestTypeQueue,
	}
	status := C.go_tibemsSession_CreateTemporaryQueue(session.cSession, &temporaryQueue.cDestination)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &temporaryQueue, nil
}

func (session *Session) CreateTemporaryTopic() (*Destination, error) {

	var temporaryTopic = Destination{
		cDestination:    0,
		cSession:        session.cSession,
		destinationType: DestTypeTopic,
	}
	status := C.go_tibemsSession_CreateTemporaryTopic(session.cSession, &temporaryTopic.cDestination)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &temporaryTopic, nil
}

// CreateBrowser creates a new QueueBrowser for the given Queue using the given (optional) message selector string.
// An application MUST call [QueueBrowser.Close] when it is finished with a QueueBrowser object to avoid leaking
// resources.
func (session *Session) CreateBrowser(queue *Queue, selector string) (*QueueBrowser, error) {
	var browser = QueueBrowser{
		cQueueBrowser: 0,
		queue:         queue,
		selector:      selector,
	}
	var cSelector *C.char
	if selector != "" {
		cSelector = C.CString(selector)
		defer C.free(unsafe.Pointer(cSelector))
	} else {
		cSelector = nil
	}
	status := C.go_tibemsSession_CreateBrowser(session.cSession, &browser.cQueueBrowser, queue.cDestination, cSelector)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &browser, nil
}

// CreateConsumer creates a message consumer on the given destination, which can be a [Topic] or a [Queue].
//
// selector is an optional message selector string.
//
// If noLocal is true, the consumer will not receive any messages sent by producers sharing the same [Connection].
//
// Applications MUST call [MsgConsumer.Close] when they are finished with a MsgConsumer to avoid resource leaks.
func (session *Session) CreateConsumer(destination *Destination, selector string, noLocal bool) (*MsgConsumer, error) {
	var consumer = MsgConsumer{
		cConsumer:         0,
		destination:       destination,
		noLocal:           noLocal,
		selector:          selector,
		msgCallbackHandle: nil,
		msgCallback:       nil,
	}

	var cNoLocal C.tibems_bool
	if noLocal {
		cNoLocal = 1
	} else {
		cNoLocal = 0
	}
	var cSelector *C.char
	if selector != "" {
		cSelector = C.CString(selector)
		defer C.free(unsafe.Pointer(cSelector))
	} else {
		cSelector = nil
	}
	var cConsumer C.uintptr_t
	status := C.go_tibemsSession_CreateConsumer(session.cSession, &cConsumer, destination.cDestination, cSelector, cNoLocal)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	consumer.cConsumer = cConsumer

	return &consumer, nil
}

// CreateSharedConsumer creates a shared message consumer on the given [Topic].
//
// sharedSubscriptionName is the name used to identify the shared non-durable subscription.
//
// selector is an optional message selector string.
//
// Applications MUST call [MsgConsumer.Close] when they are finished with a MsgConsumer to avoid resource leaks.
func (session *Session) CreateSharedConsumer(topic *Topic, sharedSubscriptionName string, selector string) (*MsgConsumer, error) {
	destination := Destination{cDestination: topic.cDestination}

	var consumer = MsgConsumer{
		cConsumer:         0,
		destination:       &destination,
		noLocal:           false,
		selector:          selector,
		msgCallbackHandle: nil,
		msgCallback:       nil,
	}

	var cSubscriptionName *C.char
	if sharedSubscriptionName != "" {
		cSubscriptionName = C.CString(sharedSubscriptionName)
		defer C.free(unsafe.Pointer(cSubscriptionName))
	} else {
		cSubscriptionName = nil
	}

	var cSelector *C.char
	if selector != "" {
		cSelector = C.CString(selector)
		defer C.free(unsafe.Pointer(cSelector))
	} else {
		cSelector = nil
	}
	status := C.go_tibemsSession_CreateSharedConsumer(session.cSession, &consumer.cConsumer, topic.cDestination, cSubscriptionName, cSelector)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &consumer, nil
}

// CreateDurableSubscriber creates a durable message consumer on the given [Topic].
//
// durableName is the name used to identify the shared durable subscription.
//
// selector is an optional message selector string.
//
// Applications MUST call [MsgConsumer.Close] when they are finished with a MsgConsumer to avoid resource leaks.
func (session *Session) CreateDurableSubscriber(topic *Topic, durableName string, selector string, noLocal bool) (*MsgConsumer, error) {
	destination := Destination{cDestination: topic.cDestination}

	var consumer = MsgConsumer{
		cConsumer:         0,
		destination:       &destination,
		noLocal:           noLocal,
		selector:          selector,
		msgCallbackHandle: nil,
		msgCallback:       nil,
	}

	var cNoLocal C.tibems_bool
	if noLocal {
		cNoLocal = 1
	} else {
		cNoLocal = 0
	}
	var cDurableName *C.char
	if durableName != "" {
		cDurableName = C.CString(durableName)
		defer C.free(unsafe.Pointer(cDurableName))
	} else {
		cDurableName = nil
	}

	var cSelector *C.char
	if selector != "" {
		cSelector = C.CString(selector)
		defer C.free(unsafe.Pointer(cSelector))
	} else {
		cSelector = nil
	}
	status := C.go_tibemsSession_CreateDurableSubscriber(session.cSession, &consumer.cConsumer, topic.cDestination, cDurableName, cSelector, cNoLocal)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &consumer, nil
}

// CreateSharedDurableConsumer creates a shared durable message consumer on the given [Topic].
//
// durableName is the name used to identify the shared durable subscription.
//
// selector is an optional message selector string.
//
// Applications MUST call [MsgConsumer.Close] when they are finished with a MsgConsumer to avoid resource leaks.
func (session *Session) CreateSharedDurableConsumer(topic *Topic, durableName string, selector string) (*MsgConsumer, error) {
	destination := Destination{cDestination: topic.cDestination}

	var consumer = MsgConsumer{
		cConsumer:         0,
		destination:       &destination,
		noLocal:           false,
		selector:          selector,
		msgCallbackHandle: nil,
		msgCallback:       nil,
	}

	var cDurableName *C.char
	if durableName != "" {
		cDurableName = C.CString(durableName)
		defer C.free(unsafe.Pointer(cDurableName))
	} else {
		cDurableName = nil
	}

	var cSelector *C.char
	if selector != "" {
		cSelector = C.CString(selector)
		defer C.free(unsafe.Pointer(cSelector))
	} else {
		cSelector = nil
	}
	status := C.go_tibemsSession_CreateSharedDurableConsumer(session.cSession, &consumer.cConsumer, topic.cDestination, cDurableName, cSelector)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &consumer, nil
}

func (session *Session) CreateRequestor(destination *Destination) (*MsgRequestor, error) {
	var requestor = MsgRequestor{cMsgRequestor: 0}

	status := C.go_tibemsMsgRequestor_Create(session.cSession, &requestor.cMsgRequestor, destination.cDestination)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &requestor, nil
}

// CreateProducer creates a new [MsgProducer] object to send messages to the given [Destination].
//
// destination is used as the default [Destination] to send messages to if an individual [Message] to be sent
// does not have a [Destination] set.  destination is optional and can be nil, however in that case attempting to
// send a [Message] without a [Destination] having been set via [Msg.SetDestination] using this MsgProducer
// will result in an error.
//
// Applications MUST call [MsgProducer.Close] when they are finished with a MsgProducer to avoid resource leaks.
func (session *Session) CreateProducer(destination *Destination) (*MsgProducer, error) {

	var producer = MsgProducer{cProducer: 0, destination: destination}
	var cDestination C.uintptr_t = 0
	if destination != nil {
		cDestination = destination.cDestination
	}

	var cProducer C.uintptr_t
	status := C.go_tibemsSession_CreateProducer(session.cSession, &cProducer, cDestination)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	producer.cProducer = cProducer

	var cDelay C.tibems_long
	status = C.go_tibemsMsgProducer_GetDeliveryDelay(producer.cProducer, &cDelay)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	producer.deliveryDelay = int64(cDelay)

	var cDeliveryMode C.tibems_int
	status = C.go_tibemsMsgProducer_GetDeliveryMode(producer.cProducer, &cDeliveryMode)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	producer.deliveryMode = DeliveryMode(cDeliveryMode)

	var cDisableMessageID C.tibems_bool
	status = C.go_tibemsMsgProducer_GetDisableMessageID(producer.cProducer, &cDisableMessageID)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var disableMessageID bool
	if cDisableMessageID == 0 {
		disableMessageID = false
	} else {
		disableMessageID = true
	}
	producer.disableMessageID = disableMessageID

	var cDisableMessageTimestamp C.tibems_bool
	status = C.go_tibemsMsgProducer_GetDisableMessageTimestamp(producer.cProducer, &cDisableMessageTimestamp)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	if cDisableMessageTimestamp == 0 {
		producer.disableMessageTimestamp = false
	} else {
		producer.disableMessageTimestamp = true
	}

	var cPriority C.tibems_int
	status = C.go_tibemsMsgProducer_GetPriority(producer.cProducer, &cPriority)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	producer.priority = int32(cPriority)

	var cTTL C.tibems_long
	status = C.go_tibemsMsgProducer_GetTimeToLive(producer.cProducer, &cTTL)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	producer.timeToLive = int64(cTTL)

	var cNpCheckMode C.tibemsNpCheckMode
	status = C.go_tibemsMsgProducer_GetNPSendCheckMode(producer.cProducer, &cNpCheckMode)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	producer.nonPersistentCheckMode = NPSendCheckMode(cNpCheckMode)

	return &producer, nil
}

// CreateBytesMessage creates a new [BytesMsg] on this Session.
//
// Applications MUST call [BytesMsg.Close] when they are finished with a BytesMsg to avoid resource leaks.
func (session *Session) CreateBytesMessage() (*BytesMsg, error) {
	var message = BytesMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.go_tibemsSession_CreateBytesMessage(session.cSession, (*C.tibemsBytesMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// CreateMapMessage creates a new [MapMsg] on this Session.
//
// Applications MUST call [MapMsg.Close] when they are finished with a MapMsg to avoid resource leaks.
func (session *Session) CreateMapMessage() (*MapMsg, error) {
	var message = MapMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.go_tibemsSession_CreateMapMessage(session.cSession, (*C.tibemsMapMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// CreateMessage creates a new [Msg] on this Session.
//
// Applications MUST call [Msg.Close] when they are finished with a Msg to avoid resource leaks.
func (session *Session) CreateMessage() (*Msg, error) {
	var message = Msg{
		cMessage: nil,
	}
	status := C.go_tibemsSession_CreateMessage(session.cSession, &message.cMessage)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &message, nil
}

// CreateStreamMessage creates a new [StreamMsg] on this Session.
//
// Applications MUST call [StreamMsg.Close] when they are finished with a StreamMsg to avoid resource leaks.
func (session *Session) CreateStreamMessage() (*StreamMsg, error) {
	var message = StreamMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.go_tibemsSession_CreateStreamMessage(session.cSession, (*C.tibemsStreamMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

// CreateTextMessage creates a new [TextMsg] on this Session.
//
// Applications MUST call [TextMsg.Close] when they are finished with a TextMsg to avoid resource leaks.
func (session *Session) CreateTextMessage() (*TextMsg, error) {
	var message = TextMsg{
		Msg: Msg{cMessage: nil},
	}

	status := C.go_tibemsSession_CreateTextMessage(session.cSession, (*C.tibemsTextMsg)(unsafe.Pointer(&message.cMessage)))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &message, nil
}

func (session *Session) Unsubscribe(durableName string) error {
	cDurableName := C.CString(durableName)
	defer C.free(unsafe.Pointer(cDurableName))
	status := C.go_tibemsSession_Unsubscribe(session.cSession, cDurableName)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (session *Session) Close() error {
	var status C.tibems_status
	if session.isXA {
		status = C.go_tibemsXASession_Close(session.cSession)
	} else {
		status = C.go_tibemsSession_Close(session.cSession)
	}
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}
