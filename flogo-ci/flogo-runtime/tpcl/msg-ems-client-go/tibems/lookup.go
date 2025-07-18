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
go_tibemsLookupContext_Create(
    uintptr_t*  context,
    const char* brokerURL,
    const char* username,
    const char* password)
{
  return tibemsLookupContext_Create((tibemsLookupContext*)context, brokerURL, username, password);
}

static tibems_status
go_tibemsLookupContext_CreateSSL(
    uintptr_t*  context,
    const char* brokerURL,
    const char* username,
    const char* password,
    uintptr_t   SSLparams,
    const char* pk_password)
{
  return tibemsLookupContext_CreateSSL(
      (tibemsLookupContext*)context, brokerURL, username, password, (tibemsSSLParams)SSLparams, pk_password);
}

static tibems_status
go_tibemsLookupContext_Destroy(uintptr_t context)
{
  return tibemsLookupContext_Destroy((tibemsLookupContext)context);
}

static tibems_status
go_tibemsLookupContext_Lookup(
    uintptr_t   context,
    const char* name,
    void**      object)
{
  return tibemsLookupContext_Lookup((tibemsLookupContext)context, name, object);
}

static tibems_status
go_tibemsLookupContext_LookupDestination(
    uintptr_t   context,
    const char* name,
    uintptr_t*  destination)
{
  return tibemsLookupContext_LookupDestination((tibemsLookupContext)context, name, (tibemsDestination*)destination);
}

static tibems_status
go_tibemsLookupContext_LookupConnectionFactory(
    uintptr_t   context,
    const char* name,
    uintptr_t*  factory)
{
  return tibemsLookupContext_LookupConnectionFactory(
      (tibemsLookupContext)context, name, (tibemsConnectionFactory*)factory);
}

*/
import "C"
import (
	"github.com/cockroachdb/errors"
)

type LookupContext struct {
	cLookupContext C.uintptr_t
}

func LookupContextCreate(brokerURL string, username string, password string, sslParams *SSLParams, privateKeyPassword string) (*LookupContext, error) {
	var lookupContext = LookupContext{cLookupContext: 0}

	cBrokerURL := C.CString(brokerURL)
	defer C.free(unsafe.Pointer(cBrokerURL))
	var cUsername *C.char
	if username != "" {
		cUsername = C.CString(username)
		defer C.free(unsafe.Pointer(cUsername))
	}
	var cPassword *C.char
	if password != "" {
		cPassword = C.CString(password)
		defer C.free(unsafe.Pointer(cPassword))
	}
	var cSSLParams C.uintptr_t
	if sslParams != nil {
		cSSLParams = sslParams.cSSLParams
	}
	var cPrivateKeyPassword *C.char
	if privateKeyPassword != "" {
		cPrivateKeyPassword = C.CString(privateKeyPassword)
		defer C.free(unsafe.Pointer(cPrivateKeyPassword))
	}

	var status C.tibems_status
	if sslParams != nil {
		status = C.go_tibemsLookupContext_CreateSSL(&lookupContext.cLookupContext, cBrokerURL, cUsername, cPassword, cSSLParams, cPrivateKeyPassword)
	} else {
		status = C.go_tibemsLookupContext_Create(&lookupContext.cLookupContext, cBrokerURL, cUsername, cPassword)
	}
	if status != tibems_OK {
		switch status {
		case tibems_SERVER_NOT_CONNECTED:
			return nil, errors.Wrap(ErrNotConnected, "EMS server connection failed")
		case tibems_SECURITY_EXCEPTION:
			return nil, errors.Wrap(ErrNotAuthorized, "Invalid credentials or TLS configuration, or user unauthorized")
		case tibems_INVALID_CLIENT_ID:
			return nil, errors.Wrap(ErrInvalid, "Client ID is invalid or is already in use")
		case tibems_INVALID_HOSTNAME:
			return nil, errors.Wrap(ErrInvalid, "Server hostname is invalid or hostname lookup failed")
		case tibems_SSL_ERROR:
			return nil, errors.Wrap(ErrTls, "Unspecified TLS error")
		default:
			return nil, errors.Wrapf(ErrGeneric, "Unexpected connection error code %d", int(status))
		}
	}

	return &lookupContext, nil
}

func (ctx *LookupContext) LookupDestination(name string) (*Destination, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	var cDestination C.uintptr_t

	status := C.go_tibemsLookupContext_LookupDestination(ctx.cLookupContext, cName, &cDestination)
	if status != tibems_OK {
		return nil, statusToError(status)
	}

	destination := Destination{
		cDestination: cDestination,
	}

	return &destination, nil
}

func (ctx *LookupContext) LookupConnectionFactory(name string) (*ConnectionFactory, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	var cFactory C.uintptr_t

	status := C.go_tibemsLookupContext_LookupConnectionFactory(ctx.cLookupContext, cName, &cFactory)
	if status != tibems_OK {
		return nil, statusToError(status)
	}

	factory := ConnectionFactory{
		cConnectionFactory: cFactory,
	}

	return &factory, nil
}

func (ctx *LookupContext) Close() error {
	status := C.go_tibemsLookupContext_Destroy(ctx.cLookupContext)
	if status != tibems_OK {
		return statusToError(status)
	}
	return nil
}
