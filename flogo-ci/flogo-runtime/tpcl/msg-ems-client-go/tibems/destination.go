package tibems

import "C"
import "unsafe"

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
go_tibemsSession_DeleteTemporaryQueue(
    uintptr_t session,
    uintptr_t tmpQueue)
{
  return tibemsSession_DeleteTemporaryQueue((tibemsSession)session, (tibemsTemporaryQueue)tmpQueue);
}

static tibems_status
go_tibemsSession_DeleteTemporaryTopic(
    uintptr_t session,
    uintptr_t tmpTopic)
{
  return tibemsSession_DeleteTemporaryTopic((tibemsSession)session, (tibemsTemporaryTopic)tmpTopic);
}

static tibems_status
go_tibemsDestination_Create(
    uintptr_t*            destination,
    tibemsDestinationType type,
    const char*           name)
{
  return tibemsDestination_Create((tibemsDestination*)destination, type, name);
}

static tibems_status
go_tibemsDestination_Destroy(uintptr_t destination)
{
  return tibemsDestination_Destroy((tibemsDestination)destination);
}

static tibems_status
go_tibemsDestination_GetName(
    uintptr_t  destination,
    char*      name,
    tibems_int name_len)
{
  return tibemsDestination_GetName((tibemsDestination)destination, name, name_len);
}

static tibems_status
go_tibemsDestination_GetType(
    uintptr_t              destination,
    tibemsDestinationType* type)
{
  return tibemsDestination_GetType((tibemsDestination)destination, type);
}

*/
import "C"

const (
	DestinationNameMax = 255 // Length of the longest allowable destination (topic or queue) name
)

type Destination struct {
	cDestination    C.uintptr_t
	cSession        C.uintptr_t // only used for temporary queues/topics linked to a session
	destinationType DestinationType
}

func CreateDestination(destinationType DestinationType, name string) (*Destination, error) {
	var destination = Destination{cDestination: 0, cSession: 0, destinationType: destinationType}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	status := C.go_tibemsDestination_Create(&destination.cDestination, C.tibemsDestinationType(destinationType), cName)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &destination, nil
}

func (destination *Destination) Close() error {
	var status C.tibems_status = tibems_INVALID_TYPE
	if destination.cSession != 0 {
		if destination.destinationType == DestTypeQueue {
			status = C.go_tibemsSession_DeleteTemporaryQueue(destination.cSession, destination.cDestination)
		} else if destination.destinationType == DestTypeTopic {
			status = C.go_tibemsSession_DeleteTemporaryTopic(destination.cSession, destination.cDestination)
		}
	} else {
		status = C.go_tibemsDestination_Destroy(destination.cDestination)
	}
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (destination *Destination) GetName() (string, error) {
	cName := C.malloc(DestinationNameMax + 1)
	defer C.free(unsafe.Pointer(cName))
	status := C.go_tibemsDestination_GetName(destination.cDestination, (*C.char)(cName), DestinationNameMax+1)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	return C.GoString((*C.char)(cName)), nil
}

func (destination *Destination) GetType() (DestinationType, error) {
	var cType C.tibemsDestinationType
	status := C.go_tibemsDestination_GetType(destination.cDestination, &cType)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return DestinationType(cType), nil
}

type Queue Destination
type Topic Destination
