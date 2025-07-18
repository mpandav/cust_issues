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
go_tibemsQueueBrowser_GetNext(
    uintptr_t  queueBrowser,
    tibemsMsg* msg)
{
  return tibemsQueueBrowser_GetNext((tibemsQueueBrowser)queueBrowser, msg);
}

static tibems_status
go_tibemsQueueBrowser_Close(uintptr_t queueBrowser)
{
  return tibemsQueueBrowser_Close((tibemsQueueBrowser)queueBrowser);
}

*/
import "C"

// # Purpose
//
// A QueueBrowser views the messages in a queue without consuming them.
//
// # Remarks
//
// A browser is a dynamic enumerator of the queue (not a static snapshot). The queue is at the server, and its contents
// change as messages arrive and consumers remove them. The function [QueueBrowser.GetNext] gets the next message from
// the server.
//
// The browser can enumerate all messages in a queue or a subset of messages filtered by a message selector.
//
// Sessions serve as factories for queue browsers; see [Session.CreateBrowser].
type QueueBrowser struct {
	cQueueBrowser C.uintptr_t
	queue         *Queue
	selector      string
}

// GetNext returns the next EMS message, if any, for the given QueueBrowser.
func (browser *QueueBrowser) GetNext() (Message, error) {
	var cMessage C.tibemsMsg
	status := C.go_tibemsQueueBrowser_GetNext(browser.cQueueBrowser, &cMessage)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return instantiateSpecificMessageType(cMessage)
}

// GetMsgSelector returns the message selector string, if any, specified when the
// QueueBrowser was created.
func (browser *QueueBrowser) GetMsgSelector() (string, error) {
	return browser.selector, nil
}

// GetQueue returns the Queue used to create the QueueBrowser.
func (browser *QueueBrowser) GetQueue() (*Queue, error) {
	return browser.queue, nil
}

// Close releases the resources associated with a QueueBrowser. Applications MUST call Close
// when finished with a QueueBrowser object to avoid resource leaks.
func (browser *QueueBrowser) Close() error {
	status := C.go_tibemsQueueBrowser_Close(browser.cQueueBrowser)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}
