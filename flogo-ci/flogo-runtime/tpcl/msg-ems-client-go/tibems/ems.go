// Package tibems implements a Go wrapper around the EMS C API.
package tibems

/*
#cgo pkg-config: ems
#include <tibems/tibems.h>
*/
import "C"
import (
	"fmt"
	"github.com/cockroachdb/errors"
	"strconv"
	"strings"
	"unsafe"
)

// Errors either returned by the EMS server or generated locally.
// These are the only errors client applications need to be able
// to handle.
var (
	ErrEMS            = errors.New("EMS error")                       // base error; other errors derive from this
	ErrNotAuthorized  = errors.Wrap(ErrEMS, "Not authorized")         // user credentials were incorrect or user is not authorized to perform the action attempted
	ErrTls            = errors.Wrap(ErrEMS, "TLS error")              // problem with TLS certificates, credentials, connection, etc.
	ErrNotConnected   = errors.Wrap(ErrEMS, "Not connected")          // application is not connected to the EMS server
	ErrReconnected    = errors.Wrap(ErrEMS, "Reconnected")            // application has reconnected to the EMS server
	ErrNotFound       = errors.Wrap(ErrEMS, "Not found")              // resource was not found or does not exist
	ErrAlreadyExists  = errors.Wrap(ErrEMS, "Already exists")         // an attempt to create a resource failed because that resource already exists
	ErrInvalid        = errors.Wrap(ErrEMS, "Invalid")                // a provided argument or object was not valid
	ErrNotAvailable   = errors.Wrap(ErrEMS, "Not available")          // a resource was not available; this may be a temporary condition
	ErrOutOfResources = errors.Wrap(ErrEMS, "Out of resources")       // a physical resource like memory or disk space has run out
	ErrExceededLimit  = errors.Wrap(ErrEMS, "Exceeded limit")         // an administratively-configured resource limit has already been reached
	ErrOS             = errors.Wrap(ErrEMS, "Operating system error") // a serious OS-level error (e.g. a disk write failure) has occurred
	ErrGeneric        = errors.Wrap(ErrEMS, "Generic error")          // unspecified error
	ErrNotImplemented = errors.Wrap(ErrEMS, "Not implemented")        // an attempt was made to use a feature or function that has not yet been implemented; it may be implemented in a later version
)

// SetExceptionOnFTSwitch enables or disables the exception listener callback being called (if set)
// for fault-tolerant failover switch events.
func SetExceptionOnFTSwitch(callExceptionListener bool) error {

	var cExceptionOnFTSwitch C.tibems_bool
	if callExceptionListener {
		cExceptionOnFTSwitch = 1
	} else {
		cExceptionOnFTSwitch = 0
	}
	status := C.tibems_setExceptionOnFTSwitch(cExceptionOnFTSwitch)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

// GetExceptionOnFTSwitch returns true if the exception listener callback will be called (if set) for
// fault-tolerant failover switch events.
func GetExceptionOnFTSwitch() (bool, error) {
	var cExceptionOnFTSwitch C.tibems_bool
	status := C.tibems_getExceptionOnFTSwitch(&cExceptionOnFTSwitch)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cExceptionOnFTSwitch == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// SetExceptionOnFTEvents enables or disables the exception listener callback being called (if set)
// for fault-tolerant failover events (disconnected, reconnect attempt, reconnected).
func SetExceptionOnFTEvents(callExceptionListener bool) error {

	var cExceptionOnFTEvents C.tibems_bool
	if callExceptionListener {
		cExceptionOnFTEvents = 1
	} else {
		cExceptionOnFTEvents = 0
	}
	status := C.tibems_SetExceptionOnFTEvents(cExceptionOnFTEvents)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

// GetExceptionOnFTEvents returns true if the exception listener callback will be called (if set) for
// fault-tolerant failover events (disconnected, reconnect attempt, reconnected).
func GetExceptionOnFTEvents() (bool, error) {
	var cExceptionOnFTEvents C.tibems_bool
	status := C.tibems_GetExceptionOnFTEvents(&cExceptionOnFTEvents)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cExceptionOnFTEvents == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// statusToError groups a few common EMS errors into a simpler set of error objects.
func statusToError(status C.tibems_status) error {
	var err error

	switch status {
	case tibems_ILLEGAL_STATE:
		err = ErrNotAvailable
	case tibems_INVALID_CLIENT_ID:
		err = ErrInvalid
	case tibems_INVALID_DESTINATION:
		err = ErrInvalid
	case tibems_INVALID_SELECTOR:
		err = ErrInvalid
	case tibems_EXCEPTION:
		err = ErrInvalid
	case tibems_SECURITY_EXCEPTION:
		err = ErrNotAuthorized
	case tibems_MSG_EOF:
	case tibems_MSG_NOT_READABLE:
	case tibems_MSG_NOT_WRITEABLE:
	case tibems_SERVER_NOT_CONNECTED:
		err = ErrNotConnected
	case tibems_SUBJECT_COLLISION:
		err = ErrNotAuthorized
	case tibems_INVALID_PROTOCOL:
		err = ErrInvalid
	case tibems_INVALID_HOSTNAME:
		err = ErrInvalid
	case tibems_INVALID_PORT:
		err = ErrInvalid
	case tibems_NO_MEMORY:
		err = ErrOutOfResources
	case tibems_INVALID_ARG:
		err = ErrInvalid
	case tibems_SERVER_LIMIT:
		err = ErrExceededLimit
	case tibems_NOT_PERMITTED:
		err = ErrNotAuthorized
	case tibems_SERVER_RECONNECTED:
		err = ErrReconnected
	case tibems_INVALID_NAME:
		err = ErrInvalid
	case tibems_INVALID_SIZE:
		err = ErrInvalid
	case tibems_NOT_FOUND:
		err = ErrNotFound
	case tibems_CONVERSION_FAILED:
	case tibems_INVALID_MSG:
		err = ErrInvalid
	case tibems_INVALID_FIELD:
		err = ErrInvalid
	case tibems_CORRUPT_MSG:
	case tibems_TIMEOUT:
	case tibems_INTR:
		err = ErrNotConnected
	case tibems_DESTINATION_LIMIT_EXCEEDED:
		err = ErrExceededLimit
	case tibems_MEM_LIMIT_EXCEEDED:
		err = ErrExceededLimit
	case tibems_USER_INTR:
	case tibems_INVALID_IO_SOURCE:
	case tibems_OS_ERROR:
		err = ErrOS
	case tibems_INSUFFICIENT_BUFFER:
		err = ErrOutOfResources
	case tibems_EOF:
	case tibems_INVALID_FILE:
		err = ErrInvalid
	case tibems_FILE_NOT_FOUND:
		err = ErrNotFound
	case tibems_IO_FAILED:
	case tibems_ALREADY_EXISTS:
		err = ErrAlreadyExists
	case tibems_INVALID_CONNECTION:
		err = ErrInvalid
	case tibems_INVALID_SESSION:
		err = ErrInvalid
	case tibems_INVALID_CONSUMER:
		err = ErrInvalid
	case tibems_INVALID_PRODUCER:
		err = ErrInvalid
	case tibems_INVALID_USER:
		err = ErrInvalid
	case tibems_TRANSACTION_FAILED:
	case tibems_TRANSACTION_ROLLBACK:
	case tibems_TRANSACTION_RETRY:
	case tibems_INVALID_XARESOURCE:
	case tibems_FT_SERVER_LACKS_TRANSACTION:
	case tibems_NOT_INITIALIZED:
	// TLS errors
	case tibems_INVALID_CERT:
		err = ErrTls
	case tibems_INVALID_CERT_NOT_YET:
		err = ErrTls
	case tibems_INVALID_CERT_EXPIRED:
		err = ErrTls
	case tibems_INVALID_CERT_DATA:
		err = ErrTls
	case tibems_ALGORITHM_ERROR:
		err = ErrTls
	case tibems_SSL_ERROR:
		err = ErrTls
	case tibems_INVALID_PRIVATE_KEY:
		err = ErrTls
	case tibems_INVALID_ENCODING:
		err = ErrTls
	case tibems_NOT_ENOUGH_RANDOM:
		err = ErrTls
	case tibems_NOT_IMPLEMENTED:
		err = ErrNotImplemented
	}

	if err == nil {
		err = ErrGeneric
	}

	return errors.WithTelemetry(err, fmt.Sprintf("tibems_status=%d", int(status)))
}

//export getStatusFromError
func getStatusFromError(err error) C.tibems_status {
	if err == nil {
		return tibems_OK
	}
	keys := errors.GetTelemetryKeys(err)
	for i := 0; i < len(keys); i++ {
		keyPair := strings.SplitN(keys[i], "tibems_status=", 2)
		if len(keyPair) == 2 {
			if status, err := strconv.Atoi(keyPair[1]); err == nil {
				return C.tibems_status(status)
			}
		}
	}
	return tibems_NOT_IMPLEMENTED
}

func getErrorString() string {
	var errStr *C.char
	status := C.tibemsErrorContext_GetLastErrorString(nil, &errStr)
	if status == tibems_OK && errStr != nil {
		return C.GoString(errStr)
	} else {
		return ""
	}
}

func getErrorFromStatus(status C.tibems_status) error {
	baseStr := C.GoString(C.tibemsStatus_GetText(status))
	baseErr := statusToError(status)
	errStr := getErrorString()
	if errStr != "" {
		return errors.Wrapf(baseErr, "%s: %s", baseStr, errStr)
	} else {
		return errors.Wrapf(baseErr, "%s", baseStr)
	}
}

func GetConnectAttemptCount() (int32, error) {
	value := C.tibems_GetConnectAttemptCount()
	return int32(value), nil
}

func SetConnectAttemptCount(count int32) error {
	status := C.tibems_SetConnectAttemptCount(C.tibems_int(count))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetConnectAttemptDelay() (int32, error) {
	value := C.tibems_GetConnectAttemptDelay()
	return int32(value), nil
}

func SetConnectAttemptDelay(delay int32) error {
	status := C.tibems_SetConnectAttemptDelay(C.tibems_int(delay))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetConnectAttemptTimeout() (int32, error) {
	value := C.tibems_GetConnectAttemptTimeout()
	return int32(value), nil
}

func SetConnectAttemptTimeout(timeout int32) error {
	status := C.tibems_SetConnectAttemptTimeout(C.tibems_int(timeout))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetReconnectAttemptCount() (int32, error) {
	value := C.tibems_GetReconnectAttemptCount()
	return int32(value), nil
}

func SetReconnectAttemptCount(count int32) error {
	status := C.tibems_SetReconnectAttemptCount(C.tibems_int(count))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetReconnectAttemptDelay() (int32, error) {
	value := C.tibems_GetReconnectAttemptDelay()
	return int32(value), nil
}

func SetReconnectAttemptDelay(delay int32) error {
	status := C.tibems_SetReconnectAttemptDelay(C.tibems_int(delay))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetReconnectAttemptTimeout() (int32, error) {
	value := C.tibems_GetReconnectAttemptTimeout()
	return int32(value), nil
}

func SetReconnectAttemptTimeout(timeout int32) error {
	status := C.tibems_SetReconnectAttemptTimeout(C.tibems_int(timeout))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetSocketReceiveBufferSize() (int32, error) {
	value := C.tibems_GetSocketReceiveBufferSize()
	return int32(value), nil
}

func SetSocketReceiveBufferSize(size int32) error {
	status := C.tibems_SetSocketReceiveBufferSize(C.tibems_int(size))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func GetSocketSendBufferSize() (int32, error) {
	value := C.tibems_GetSocketSendBufferSize()
	return int32(value), nil
}

func SetSocketSendBufferSize(size int32) error {
	status := C.tibems_SetSocketSendBufferSize(C.tibems_int(size))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func SetTraceFile(fileName string) error {
	var cFileName *C.char = nil

	if fileName != "" {
		cFileName = C.CString(fileName)
		defer C.free(unsafe.Pointer(cFileName))
	}
	status := C.tibems_SetTraceFile(cFileName)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

// Version returns a string representing the EMS library version number.
func Version() string {
	return C.GoString(C.tibems_Version())
}
