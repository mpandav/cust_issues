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

static uintptr_t
go_tibemsOAuth2Params_Create(void)
{
  return (uintptr_t)tibemsOAuth2Params_Create();
}

static void
go_tibemsOAuth2Params_Destroy(uintptr_t params)
{
  return tibemsOAuth2Params_Destroy((tibemsOAuth2Params)params);
}

static tibems_status
go_tibemsOAuth2Params_SetAccessToken(
    uintptr_t   params,
    const char* accessToken)
{
  return tibemsOAuth2Params_SetAccessToken((tibemsOAuth2Params)params, accessToken);
}

static tibems_status
go_tibemsOAuth2Params_SetServerURL(
    uintptr_t   params,
    const char* url)
{
  return tibemsOAuth2Params_SetServerURL((tibemsOAuth2Params)params, url);
}

static tibems_status
go_tibemsOAuth2Params_SetClientID(
    uintptr_t   params,
    const char* clientId)
{
  return tibemsOAuth2Params_SetClientID((tibemsOAuth2Params)params, clientId);
}

static tibems_status
go_tibemsOAuth2Params_SetClientSecret(
    uintptr_t   params,
    const char* clientSecret)
{
  return tibemsOAuth2Params_SetClientSecret((tibemsOAuth2Params)params, clientSecret);
}

static tibems_status
go_tibemsOAuth2Params_SetServerTrustFile(
    uintptr_t   params,
    const char* path)
{
  return tibemsOAuth2Params_SetServerTrustFile((tibemsOAuth2Params)params, path);
}

static tibems_status
go_tibemsOAuth2Params_SetDisableVerifyHostname(
    uintptr_t   params,
    tibems_bool disableVerifyHostname)
{
  return tibemsOAuth2Params_SetDisableVerifyHostname((tibemsOAuth2Params)params, disableVerifyHostname);
}

static tibems_status
go_tibemsOAuth2Params_SetExpectedHostname(
    uintptr_t   params,
    const char* hostname)
{
  return tibemsOAuth2Params_SetExpectedHostname((tibemsOAuth2Params)params, hostname);
}

*/
import "C"
import "C"
import (
	"github.com/cockroachdb/errors"
	"runtime/cgo"
	"unsafe"
)

type OAuth2Params struct {
	cOAuth2Params            C.uintptr_t
	tokenFetchCallbackHandle cgo.Handle
}

func (params *OAuth2Params) Close() error {
	C.go_tibemsOAuth2Params_Destroy(params.cOAuth2Params)
	params.tokenFetchCallbackHandle.Delete()
	return nil
}

func CreateOAuth2Params() (*OAuth2Params, error) {
	var params = OAuth2Params{cOAuth2Params: 0}

	params.cOAuth2Params = C.go_tibemsOAuth2Params_Create()
	if params.cOAuth2Params == 0 {
		return nil, errors.New("could not create OAuth2 params; out of memory")
	}
	params.tokenFetchCallbackHandle = cgo.NewHandle(nil)

	return &params, nil
}

func (params *OAuth2Params) SetAccessToken(token string) error {
	cToken := C.CString(token)
	defer C.free(unsafe.Pointer(cToken))

	status := C.go_tibemsOAuth2Params_SetAccessToken(params.cOAuth2Params, cToken)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (params *OAuth2Params) SetClientID(clientID string) error {
	cClientID := C.CString(clientID)
	defer C.free(unsafe.Pointer(cClientID))

	status := C.go_tibemsOAuth2Params_SetClientID(params.cOAuth2Params, cClientID)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (params *OAuth2Params) SetClientSecret(secret string) error {
	cSecret := C.CString(secret)
	defer C.free(unsafe.Pointer(cSecret))

	status := C.go_tibemsOAuth2Params_SetClientSecret(params.cOAuth2Params, cSecret)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (params *OAuth2Params) SetVerifyHostName(verifyHostName bool) error {
	var cValue C.tibems_bool

	// These bool values are swapped, since this function is called
	// "SetVerifyHostName" for consistency with SSLParams.SetVerifyHostName.
	if verifyHostName {
		cValue = 0
	} else {
		cValue = 1
	}

	status := C.go_tibemsOAuth2Params_SetDisableVerifyHostname(params.cOAuth2Params, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (params *OAuth2Params) SetExpectedHostName(hostName string) error {
	cHostName := C.CString(hostName)
	defer C.free(unsafe.Pointer(cHostName))

	status := C.go_tibemsOAuth2Params_SetExpectedHostname(params.cOAuth2Params, cHostName)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (params *OAuth2Params) SetServerTrustFile(trustFile string) error {
	cFile := C.CString(trustFile)
	defer C.free(unsafe.Pointer(cFile))

	status := C.go_tibemsOAuth2Params_SetServerTrustFile(params.cOAuth2Params, cFile)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (params *OAuth2Params) SetServerURL(serverURL string) error {
	cURL := C.CString(serverURL)
	defer C.free(unsafe.Pointer(cURL))

	status := C.go_tibemsOAuth2Params_SetServerURL(params.cOAuth2Params, cURL)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func SetOAuth2Trace(trace bool) {
	var cTrace C.tibems_bool
	if trace {
		cTrace = 1
	} else {
		cTrace = 0
	}
	C.tibemsOAuth2_SetTrace(cTrace)
}

func SetOAuth2DebugTrace(trace bool) {
	var cTrace C.tibems_bool
	if trace {
		cTrace = 1
	} else {
		cTrace = 0
	}
	C.tibemsOAuth2_SetDebugTrace(cTrace)
}

func GetOAuth2Trace() bool {
	cTrace := C.tibemsOAuth2_GetTrace()
	if cTrace == 0 {
		return false
	} else {
		return true
	}
}

func GetOAuth2DebugTrace() bool {
	cTrace := C.tibemsOAuth2_GetDebugTrace()
	if cTrace == 0 {
		return false
	} else {
		return true
	}
}
