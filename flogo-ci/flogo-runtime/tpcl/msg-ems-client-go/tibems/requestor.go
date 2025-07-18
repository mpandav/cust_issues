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
go_tibemsMsgRequestor_Close(uintptr_t msgRequestor)
{
  return tibemsMsgRequestor_Close((tibemsMsgRequestor)msgRequestor);
}

static tibems_status
go_tibemsMsgRequestor_Request(
    uintptr_t  msgRequestor,
    tibemsMsg  msgSent,
    tibemsMsg* msgReply)
{
  return tibemsMsgRequestor_Request((tibemsMsgRequestor)msgRequestor, msgSent, msgReply);
}

*/
import "C"

type MsgRequestor struct {
	cMsgRequestor C.uintptr_t
}

// Request sends a message and waits (blocking indefinitely) for a reply. It returns
// the first reply received.
func (requestor *MsgRequestor) Request(message Message) (reply Message, err error) {
	err = message.flushPending()
	if err != nil {
		return nil, err
	}

	var cReply C.tibemsMsg

	status := C.go_tibemsMsgRequestor_Request(requestor.cMsgRequestor, message.getCMessage(), &cReply)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return instantiateSpecificMessageType(cReply)
}

func (requestor *MsgRequestor) Close() error {
	status := C.go_tibemsMsgRequestor_Close(requestor.cMsgRequestor)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}
