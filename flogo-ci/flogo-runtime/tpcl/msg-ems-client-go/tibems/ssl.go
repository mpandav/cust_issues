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
go_tibemsSSLParams_SetExpectedHostName(
    uintptr_t   params,
    const char* expected_hostname)
{
  return tibemsSSLParams_SetExpectedHostName((tibemsSSLParams)params, expected_hostname);
}

static uintptr_t
go_tibemsSSLParams_Create(void)
{
  return (uintptr_t)tibemsSSLParams_Create();
}

static void
go_tibemsSSLParams_Destroy(uintptr_t params)
{
  return tibemsSSLParams_Destroy((tibemsSSLParams)params);
}

static tibems_status
go_tibemsSSLParams_SetCiphers(
    uintptr_t   params,
    const char* ciphers)
{
  return tibemsSSLParams_SetCiphers((tibemsSSLParams)params, ciphers);
}

static tibems_status
go_tibemsSSLParams_SetIdentity(
    uintptr_t   params,
    const void* data,
    tibems_int  size,
    tibems_int  encoding)
{
  return tibemsSSLParams_SetIdentity((tibemsSSLParams)params, data, size, encoding);
}

static tibems_status
go_tibemsSSLParams_GetIdentity(
    uintptr_t    params,
    const void** data,
    tibems_int*  size,
    tibems_int*  encoding)
{
  return tibemsSSLParams_GetIdentity((tibemsSSLParams)params, data, size, encoding);
}

static tibems_status
go_tibemsSSLParams_SetIdentityFile(
    uintptr_t   params,
    const char* filename,
    tibems_int  encoding)
{
  return tibemsSSLParams_SetIdentityFile((tibemsSSLParams)params, filename, encoding);
}

static tibems_status
go_tibemsSSLParams_AddIssuerCert(
    uintptr_t   params,
    const void* data,
    tibems_int  size,
    tibems_int  encoding)
{
  return tibemsSSLParams_AddIssuerCert((tibemsSSLParams)params, data, size, encoding);
}

static tibems_status
go_tibemsSSLParams_AddIssuerCertFile(
    uintptr_t   params,
    const char* filename,
    tibems_int  encoding)
{
  return tibemsSSLParams_AddIssuerCertFile((tibemsSSLParams)params, filename, encoding);
}

static tibems_status
go_tibemsSSLParams_SetPrivateKey(
    uintptr_t   params,
    const void* data,
    tibems_int  size,
    tibems_int  encoding)
{
  return tibemsSSLParams_SetPrivateKey((tibemsSSLParams)params, data, size, encoding);
}

static tibems_status
go_tibemsSSLParams_GetPrivateKey(
    uintptr_t    params,
    const void** data,
    tibems_int*  size,
    tibems_int*  encoding)
{
  return tibemsSSLParams_GetPrivateKey((tibemsSSLParams)params, data, size, encoding);
}

static tibems_status
go_tibemsSSLParams_SetPrivateKeyFile(
    uintptr_t   params,
    const char* filename,
    tibems_int  encoding)
{
  return tibemsSSLParams_SetPrivateKeyFile((tibemsSSLParams)params, filename, encoding);
}

static tibems_status
go_tibemsSSLParams_AddTrustedCert(
    uintptr_t   params,
    const void* data,
    tibems_int  size,
    tibems_int  encoding)
{
  return tibemsSSLParams_AddTrustedCert((tibemsSSLParams)params, data, size, encoding);
}

static tibems_status
go_tibemsSSLParams_AddTrustedCertFile(
    uintptr_t   params,
    const char* filename,
    tibems_int  encoding)
{
  return tibemsSSLParams_AddTrustedCertFile((tibemsSSLParams)params, filename, encoding);
}

static tibems_status
go_tibemsSSLParams_SetAuthOnly(
    uintptr_t   params,
    tibems_bool auth_only)
{
  return tibemsSSLParams_SetAuthOnly((tibemsSSLParams)params, auth_only);
}

static tibems_status
go_tibemsSSLParams_SetVerifyHost(
    uintptr_t   params,
    tibems_bool verify)
{
  return tibemsSSLParams_SetVerifyHost((tibemsSSLParams)params, verify);
}

static tibems_status
go_tibemsSSLParams_SetVerifyHostName(
    uintptr_t   params,
    tibems_bool verify)
{
  return tibemsSSLParams_SetVerifyHostName((tibemsSSLParams)params, verify);
}

extern tibems_status
goSSLHostNameVerifierCallback(
    char*,
    char*,
    char*,
    uintptr_t);

static tibems_status
_cSSLHostNameVerifierCallback(
    const char* connected_hostname,
    const char* expected_hostname,
    const char* certificate_name,
    void*       closure)
{
  return goSSLHostNameVerifierCallback(
      (char*)connected_hostname, (char*)expected_hostname, (char*)certificate_name, (uintptr_t)closure);
}

static tibems_status
_cSetHostNameVerifierCallback(
    uintptr_t params,
    uintptr_t handle)
{
  return tibemsSSLParams_SetHostNameVerifier((tibemsSSLParams)params, _cSSLHostNameVerifierCallback, (void*)handle);
}

static tibems_status
_callHostnameVerifierCallback(
    uintptr_t cb,
    char*     connected_hostname,
    char*     expected_hostname,
    char*     certificate_name,
    uintptr_t closure)
{
  return ((tibemsSSLHostNameVerifier)cb)(connected_hostname, expected_hostname, certificate_name, (void*)closure);
}

*/
import "C"
import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

type SSLParams struct {
	cSSLParams             C.uintptr_t
	hostNameVerifierHandle cgo.Handle
}

func (sslParams *SSLParams) Close() error {
	C.go_tibemsSSLParams_Destroy(sslParams.cSSLParams)
	sslParams.hostNameVerifierHandle.Delete()
	return nil
}

func CreateSSLParams() (*SSLParams, error) {
	var params = SSLParams{cSSLParams: 0}

	params.cSSLParams = C.go_tibemsSSLParams_Create()
	if params.cSSLParams == 0 {
		return nil, errors.New("could not create SSL params; out of memory")
	}
	params.hostNameVerifierHandle = cgo.NewHandle(nil)

	return &params, nil
}

func (sslParams *SSLParams) SetCiphers(ciphers string) error {
	cCiphers := C.CString(ciphers)
	defer C.free(unsafe.Pointer(cCiphers))

	status := C.go_tibemsSSLParams_SetCiphers(sslParams.cSSLParams, cCiphers)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) SetIdentity(data []byte, encoding SSLEncodingType) error {
	status := C.go_tibemsSSLParams_SetIdentity(sslParams.cSSLParams, unsafe.Pointer(&data[0]), C.tibems_int(len(data)), C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) GetIdentity() ([]byte, SSLEncodingType, error) {
	var cData unsafe.Pointer
	var cSize C.tibems_int
	var cEncoding C.tibems_int

	status := C.go_tibemsSSLParams_GetIdentity(sslParams.cSSLParams, &cData, &cSize, &cEncoding)
	if status != tibems_OK {
		return nil, 0, getErrorFromStatus(status)
	}

	return C.GoBytes(cData, C.int(cSize)), SSLEncodingType(cEncoding), nil
}

func (sslParams *SSLParams) SetIdentityFile(filename string, encoding SSLEncodingType) error {
	cFile := C.CString(filename)
	defer C.free(unsafe.Pointer(cFile))

	status := C.go_tibemsSSLParams_SetIdentityFile(sslParams.cSSLParams, cFile, C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) AddIssuerCert(data []byte, encoding SSLEncodingType) error {

	status := C.go_tibemsSSLParams_AddIssuerCert(sslParams.cSSLParams, unsafe.Pointer(&data[0]), C.tibems_int(len(data)), C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) AddIssuerCertFile(filename string, encoding SSLEncodingType) error {
	cFile := C.CString(filename)
	defer C.free(unsafe.Pointer(cFile))

	status := C.go_tibemsSSLParams_AddIssuerCertFile(sslParams.cSSLParams, cFile, C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) SetPrivateKey(data []byte, encoding SSLEncodingType) error {

	status := C.go_tibemsSSLParams_SetPrivateKey(sslParams.cSSLParams, unsafe.Pointer(&data[0]), C.tibems_int(len(data)), C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) GetPrivateKey() ([]byte, SSLEncodingType, error) {
	var cData unsafe.Pointer
	var cSize C.tibems_int
	var cEncoding C.tibems_int

	status := C.go_tibemsSSLParams_GetPrivateKey(sslParams.cSSLParams, &cData, &cSize, &cEncoding)
	if status != tibems_OK {
		return nil, 0, getErrorFromStatus(status)
	}

	return C.GoBytes(cData, C.int(cSize)), SSLEncodingType(cEncoding), nil
}

func (sslParams *SSLParams) SetPrivateKeyFile(filename string, encoding SSLEncodingType) error {
	cFile := C.CString(filename)
	defer C.free(unsafe.Pointer(cFile))

	status := C.go_tibemsSSLParams_SetPrivateKeyFile(sslParams.cSSLParams, cFile, C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) AddTrustedCert(data []byte, encoding SSLEncodingType) error {

	status := C.go_tibemsSSLParams_AddTrustedCert(sslParams.cSSLParams, unsafe.Pointer(&data[0]), C.tibems_int(len(data)), C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) AddTrustedCertFile(filename string, encoding SSLEncodingType) error {
	cFile := C.CString(filename)
	defer C.free(unsafe.Pointer(cFile))

	status := C.go_tibemsSSLParams_AddTrustedCertFile(sslParams.cSSLParams, cFile, C.tibems_int(encoding))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) SetAuthOnly(authOnly bool) error {

	var cValue C.tibems_bool
	if authOnly {
		cValue = 1
	} else {
		cValue = 0
	}

	status := C.go_tibemsSSLParams_SetAuthOnly(sslParams.cSSLParams, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) SetVerifyHost(verifyHost bool) error {

	var cValue C.tibems_bool
	if verifyHost {
		cValue = 1
	} else {
		cValue = 0
	}

	status := C.go_tibemsSSLParams_SetVerifyHost(sslParams.cSSLParams, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) SetVerifyHostName(verifyHostName bool) error {

	var cValue C.tibems_bool
	if verifyHostName {
		cValue = 1
	} else {
		cValue = 0
	}

	status := C.go_tibemsSSLParams_SetVerifyHostName(sslParams.cSSLParams, cValue)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

type sslHostNameVerifierCallback struct {
	callback SSLHostNameVerifier
}
type SSLHostNameVerifier func(connectedHostname string, expectedHostname string, certificateName string) error

// Used for testing.
//
//export createCHostnameVerifierCallback
func createCHostnameVerifierCallback(cFunc C.uintptr_t, cClosure C.uintptr_t) uintptr {
	var fn SSLHostNameVerifier
	fn = func(connectedHostname string, expectedHostname string, certificateName string) error {
		cConnectedHostname := C.CString(connectedHostname)
		defer C.free(unsafe.Pointer(cConnectedHostname))
		cExpectedHostname := C.CString(expectedHostname)
		defer C.free(unsafe.Pointer(cExpectedHostname))
		cCertificateName := C.CString(certificateName)
		defer C.free(unsafe.Pointer(cCertificateName))

		status := C._callHostnameVerifierCallback(cFunc, cConnectedHostname, cExpectedHostname, cCertificateName, cClosure)
		if status != tibems_OK {
			return getErrorFromStatus(status)
		}
		return nil
	}
	return uintptr(cgo.NewHandle(fn))
}

//export goSSLHostNameVerifierCallback
func goSSLHostNameVerifierCallback(connectedHostname *C.char, expectedHostname *C.char, certificateName *C.char, callbackInfoHandle C.uintptr_t) C.tibems_status {
	handle := cgo.Handle(callbackInfoHandle)
	callback := handle.Value().(sslHostNameVerifierCallback)
	err := callback.callback(C.GoString(connectedHostname), C.GoString(expectedHostname), C.GoString(certificateName))
	if err == nil {
		return tibems_OK
	} else {
		return tibems_SECURITY_EXCEPTION
	}
}

func (sslParams *SSLParams) SetHostNameVerifier(hostnameVerifierCallback SSLHostNameVerifier) error {

	cbInfo := sslHostNameVerifierCallback{
		callback: hostnameVerifierCallback,
	}

	sslParams.hostNameVerifierHandle.Delete()
	sslParams.hostNameVerifierHandle = cgo.NewHandle(cbInfo)

	status := C._cSetHostNameVerifierCallback(sslParams.cSSLParams, C.uintptr_t(sslParams.hostNameVerifierHandle))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (sslParams *SSLParams) SetExpectedHostName(expectedHostname string) error {
	cExpectedHostname := C.CString(expectedHostname)
	defer C.free(unsafe.Pointer(cExpectedHostname))

	status := C.go_tibemsSSLParams_SetExpectedHostName(sslParams.cSSLParams, cExpectedHostname)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func SetSSLTrace(trace bool) {
	var cTrace C.tibems_bool
	if trace {
		cTrace = 1
	} else {
		cTrace = 0
	}
	C.tibemsSSL_SetTrace(cTrace)
}

func SetSSLDebugTrace(trace bool) {
	var cTrace C.tibems_bool
	if trace {
		cTrace = 1
	} else {
		cTrace = 0
	}
	C.tibemsSSL_SetDebugTrace(cTrace)
}

func GetSSLTrace() bool {
	cTrace := C.tibemsSSL_GetTrace()
	if cTrace == 0 {
		return false
	} else {
		return true
	}
}

func GetSSLDebugTrace() bool {
	cTrace := C.tibemsSSL_GetDebugTrace()
	if cTrace == 0 {
		return false
	} else {
		return true
	}
}
