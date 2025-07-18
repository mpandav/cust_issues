package tibems

/*
#include <inttypes.h>
#include <tibems/proxy.h>
#include <tibems/tibems.h>
#include <tibems/tibufo.h>

extern tibems_status
goOAuth2TokenFetchCallback(
    char**,
    uintptr_t);

static tibems_status
_cOAuth2TokenFetchCallback(
    char**           accessToken,
    tibemsConnection connection)
{
  return goOAuth2TokenFetchCallback(accessToken, (uintptr_t)connection);
}

static tibems_status
_setOAuth2TokenFetchCallback(uintptr_t factory)
{
  return tibemsConnectionFactory_SetOAuth2TokenFetchCallback(
      (tibemsConnectionFactory)factory, _cOAuth2TokenFetchCallback);
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
go_tibemsConnectionFactory_SetOAuth2TokenFetchCallback(
    uintptr_t factory,
    uintptr_t callback)
{
  return tibemsConnectionFactory_SetOAuth2TokenFetchCallback(
      (tibemsConnectionFactory)factory, (tibemsOAuth2TokenFetchCallback)callback);
}

static uintptr_t
go_tibemsConnectionFactory_Create()
{
  return (uintptr_t)tibemsConnectionFactory_Create();
}

static tibems_status
go_tibemsConnectionFactory_SetServerURL(
    uintptr_t   factory,
    const char* url)
{
  return tibemsConnectionFactory_SetServerURL((tibemsConnectionFactory)factory, url);
}

static tibems_status
go_tibemsConnectionFactory_Destroy(uintptr_t factory)
{
  return tibemsConnectionFactory_Destroy((tibemsConnectionFactory)factory);
}

static uintptr_t
go_tibemsUFOConnectionFactory_Create()
{
  return (uintptr_t)tibemsUFOConnectionFactory_Create();
}

static uintptr_t
go_tibemsUFOConnectionFactory_CreateFromConnectionFactory(uintptr_t emsFactory)
{
  return (uintptr_t)tibemsUFOConnectionFactory_CreateFromConnectionFactory((tibemsConnectionFactory)emsFactory);
}

static tibems_status
go_tibemsUFOConnectionFactory_RecoverConnection(
    uintptr_t factory,
    uintptr_t ufoConnection)
{
  return tibemsUFOConnectionFactory_RecoverConnection(
      (tibemsConnectionFactory)factory, (tibemsConnection)ufoConnection);
}

static tibems_status
go_tibemsConnectionFactory_CreateXAConnection(
    uintptr_t   factory,
    uintptr_t*  connection,
    const char* username,
    const char* password)
{
  return tibemsConnectionFactory_CreateXAConnection(
      (tibemsConnectionFactory)factory, (tibemsConnection*)connection, username, password);
}

static tibems_status
go_tibemsConnectionFactory_CreateConnection(
    uintptr_t   factory,
    uintptr_t*  connection,
    const char* username,
    const char* password)
{
  return tibemsConnectionFactory_CreateConnection(
      (tibemsConnectionFactory)factory, (tibemsConnection*)connection, username, password);
}

static tibems_status
go_tibemsConnectionFactory_SetOAuth2Params(
    uintptr_t factory,
    uintptr_t oauth2Params)
{
  return tibemsConnectionFactory_SetOAuth2Params((tibemsConnectionFactory)factory, (tibemsOAuth2Params)oauth2Params);
}

static tibems_status
go_tibemsConnectionFactory_SetSSLProxy(
    uintptr_t   factory,
    const char* proxy_host,
    tibems_int  proxy_port)
{
  return tibemsConnectionFactory_SetSSLProxy((tibemsConnectionFactory)factory, proxy_host, proxy_port);
}

static tibems_status
go_tibemsConnectionFactory_SetSSLProxyAuth(
    uintptr_t   factory,
    const char* proxy_user,
    const char* proxy_password)
{
  return tibemsConnectionFactory_SetSSLProxyAuth((tibemsConnectionFactory)factory, proxy_user, proxy_password);
}

static tibems_status
go_tibemsConnectionFactory_GetSSLProxyHost(
    uintptr_t    factory,
    const char** proxy_host)
{
  return tibemsConnectionFactory_GetSSLProxyHost((tibemsConnectionFactory)factory, proxy_host);
}

static tibems_status
go_tibemsConnectionFactory_GetSSLProxyPort(
    uintptr_t   factory,
    tibems_int* proxy_port)
{
  return tibemsConnectionFactory_GetSSLProxyPort((tibemsConnectionFactory)factory, proxy_port);
}

static tibems_status
go_tibemsConnectionFactory_GetSSLProxyUser(
    uintptr_t    factory,
    const char** proxy_user)
{
  return tibemsConnectionFactory_GetSSLProxyUser((tibemsConnectionFactory)factory, proxy_user);
}

static tibems_status
go_tibemsConnectionFactory_GetSSLProxyPassword(
    uintptr_t    factory,
    const char** proxy_password)
{
  return tibemsConnectionFactory_GetSSLProxyPassword((tibemsConnectionFactory)factory, proxy_password);
}

static tibems_status
go_tibemsConnectionFactory_PrintToBuffer(
    uintptr_t  factory,
    char*      buffer,
    tibems_int maxlen)
{
  return tibemsConnectionFactory_PrintToBuffer((tibemsConnectionFactory)factory, buffer, maxlen);
}

static tibems_status
go_tibemsConnectionFactory_SetPkPassword(
    uintptr_t   factory,
    const char* pk_password)
{
  return tibemsConnectionFactory_SetPkPassword((tibemsConnectionFactory)factory, pk_password);
}

static tibems_status
go_tibemsConnectionFactory_SetSSLParams(
    uintptr_t factory,
    uintptr_t sslparams)
{
  return tibemsConnectionFactory_SetSSLParams((tibemsConnectionFactory)factory, (tibemsSSLParams)sslparams);
}

static tibems_status
go_tibemsConnectionFactory_SetClientID(
    uintptr_t   factory,
    const char* cid)
{
  return tibemsConnectionFactory_SetClientID((tibemsConnectionFactory)factory, cid);
}

static tibems_status
go_tibemsConnectionFactory_SetMetric(
    uintptr_t                      factory,
    tibemsFactoryLoadBalanceMetric metric)
{
  return tibemsConnectionFactory_SetMetric((tibemsConnectionFactory)factory, metric);
}

static tibems_status
go_tibemsConnectionFactory_SetConnectAttemptCount(
    uintptr_t  factory,
    tibems_int connAttempts)
{
  return tibemsConnectionFactory_SetConnectAttemptCount((tibemsConnectionFactory)factory, connAttempts);
}

static tibems_status
go_tibemsConnectionFactory_SetConnectAttemptDelay(
    uintptr_t  factory,
    tibems_int delay)
{
  return tibemsConnectionFactory_SetConnectAttemptDelay((tibemsConnectionFactory)factory, delay);
}

static tibems_status
go_tibemsConnectionFactory_SetConnectAttemptTimeout(
    uintptr_t  factory,
    tibems_int connAttemptTimeout)
{
  return tibemsConnectionFactory_SetConnectAttemptTimeout((tibemsConnectionFactory)factory, connAttemptTimeout);
}

static tibems_status
go_tibemsConnectionFactory_SetReconnectAttemptCount(
    uintptr_t  factory,
    tibems_int connAttempts)
{
  return tibemsConnectionFactory_SetReconnectAttemptCount((tibemsConnectionFactory)factory, connAttempts);
}

static tibems_status
go_tibemsConnectionFactory_SetReconnectAttemptDelay(
    uintptr_t  factory,
    tibems_int delay)
{
  return tibemsConnectionFactory_SetReconnectAttemptDelay((tibemsConnectionFactory)factory, delay);
}

static tibems_status
go_tibemsConnectionFactory_SetReconnectAttemptTimeout(
    uintptr_t  factory,
    tibems_int reconnAttemptTimeout)
{
  return tibemsConnectionFactory_SetReconnectAttemptTimeout((tibemsConnectionFactory)factory, reconnAttemptTimeout);
}

static tibems_status
go_tibemsConnectionFactory_SetUserName(
    uintptr_t   factory,
    const char* username)
{
  return tibemsConnectionFactory_SetUserName((tibemsConnectionFactory)factory, username);
}

static tibems_status
go_tibemsConnectionFactory_SetUserPassword(
    uintptr_t   factory,
    const char* password)
{
  return tibemsConnectionFactory_SetUserPassword((tibemsConnectionFactory)factory, password);
}

*/
import "C"
import (
	"github.com/cockroachdb/errors"
	"sync"
	"unsafe"
)

var cConnectionToConnectionMap = make(map[C.uintptr_t]any)
var cConnectionToConnectionMapLock = &sync.RWMutex{}
var globalFactoryConnectionCreateLock = &sync.Mutex{}
var globalFactoryConnectionCreateFactory *ConnectionFactory

type OAuth2TokenFetchCallback func(connection EMSConnection) (string, error)

//export goOAuth2TokenFetchCallback
func goOAuth2TokenFetchCallback(token **C.char, cConnection C.uintptr_t) C.tibems_status {

	cConnectionToConnectionMapLock.RLock()
	// If connection is not present in the map, then it's new and we're being called
	// inline while creating the connection via a factory (and currently also hold
	// the globalFactoryConnectionCreateLock lock).
	connection, found := cConnectionToConnectionMap[cConnection]
	if !found {
		cConnectionToConnectionMapLock.RUnlock()
		connection = globalFactoryConnectionCreateFactory.partiallyCreatedConnection
		switch conn := connection.(type) {
		case *Connection:
			conn.cConnection = cConnection
		case *XAConnection:
			conn.cConnection = cConnection
		}
		cConnectionToConnectionMapLock.Lock()
		cConnectionToConnectionMap[cConnection] = connection
		cConnectionToConnectionMapLock.Unlock()
	} else {
		cConnectionToConnectionMapLock.RUnlock()
	}

	var callback OAuth2TokenFetchCallback
	switch conn := connection.(type) {
	case *Connection:
		callback = conn.factory.oauth2TokenFetchCallback
	case *XAConnection:
		callback = conn.factory.oauth2TokenFetchCallback
	}

	goToken, err := callback(connection.(EMSConnection))
	if err != nil {
		return tibems_SECURITY_EXCEPTION
	}
	cToken := C.CString(goToken) // Don't free here, will be free'd by EMS.
	*token = cToken

	return tibems_OK
}

// A ConnectionFactory is an administered object for creating server connections.
//
// # Remarks
//
// Connection factories are administered objects. They support concurrent use.
// Administrators define connection factories in a repository. Each connection factory has administrative parameters
// that guide the creation of server connections. Usage follows either of two models:
//
// # EMS Server
//
// You can use the EMS server as a name service provider—one tibemsd process provides both the name repository and the
// message service. Administrators define factories in the name repository. Client programs create connection factory
// objects with the URL of the repository, and call [ConnectionFactory.CreateConnection]. This function automatically
// accesses the corresponding factory in the repository, and uses it to create a connection to the message service.
//
// # Administered Objects
//
// Administered objects let administrators configure EMS behavior at the enterprise level. Administrators define these
// objects, and client programs use them. This arrangement relieves program developers and end users of the
// responsibility for correct configuration.
type ConnectionFactory struct {
	cConnectionFactory           C.uintptr_t
	oauth2TokenFetchCallback     OAuth2TokenFetchCallback
	oauth2TokenFetchCallbackLock *sync.RWMutex
	partiallyCreatedConnection   any
}

func CreateConnectionFactory(serverURL string) (*ConnectionFactory, error) {
	factory := ConnectionFactory{cConnectionFactory: 0, oauth2TokenFetchCallbackLock: &sync.RWMutex{}}

	cFactory := C.go_tibemsConnectionFactory_Create()
	if cFactory == 0 {
		return nil, errors.Wrap(ErrOutOfResources, "not enough memory to create connection factory")
	}
	factory.cConnectionFactory = cFactory

	cServerURL := C.CString(serverURL)
	defer C.free(unsafe.Pointer(cServerURL))
	status := C.go_tibemsConnectionFactory_SetServerURL(cFactory, cServerURL)
	if status != tibems_OK {
		C.go_tibemsConnectionFactory_Destroy(factory.cConnectionFactory)
		return nil, statusToError(status)
	}
	return &factory, nil
}

func CreateUnsharedFailOverConnectionFactory(fromFactory *ConnectionFactory) (*ConnectionFactory, error) {
	factory := ConnectionFactory{cConnectionFactory: 0, oauth2TokenFetchCallbackLock: &sync.RWMutex{}}

	var cFactory C.uintptr_t
	if fromFactory == nil {
		cFactory = C.go_tibemsUFOConnectionFactory_Create()
	} else {
		cFactory = C.go_tibemsUFOConnectionFactory_CreateFromConnectionFactory(fromFactory.cConnectionFactory)
	}
	if cFactory == 0 {
		return nil, errors.Wrap(ErrOutOfResources, "not enough memory to create UFO connection factory")
	}
	factory.cConnectionFactory = cFactory
	return &factory, nil
}

func (ufoFactory *ConnectionFactory) RecoverUnsharedFailOverConnection(ufoConnection *Connection) error {
	status := C.go_tibemsUFOConnectionFactory_RecoverConnection(ufoFactory.cConnectionFactory, ufoConnection.cConnection)
	if status != tibems_OK {
		return statusToError(status)
	}
	return nil
}

func (factory *ConnectionFactory) Close() error {
	status := C.go_tibemsConnectionFactory_Destroy(factory.cConnectionFactory)
	if status != tibems_OK {
		return statusToError(status)
	}
	return nil
}

func (factory *ConnectionFactory) CreateXAConnection(username string, password string) (*XAConnection, error) {
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

	connection := XAConnection{Connection{
		cConnection:             0,
		exceptionListenerHandle: nil,
		isXA:                    true,
		factory:                 factory,
	}}

	// About to call the C CreateConnection, which might call the OAuth2TokenFetch
	// callback inline. But we haven't finished creating our Go Connection object
	// yet (it's missing the cConnection pointer that we're about to create...), so
	// set a global reference (a thread-local reference would be better, but Golang
	// doesn't have thread-local storage) to this connection factory. If the
	// OAuth2TokenFetch does get called, it will first set the
	// factory.partiallyCreatedConnection.cConnection pointer before calling the
	// user's callback.
	globalFactoryConnectionCreateLock.Lock()
	factory.partiallyCreatedConnection = &connection
	globalFactoryConnectionCreateFactory = factory

	var cConnection C.uintptr_t
	status := C.go_tibemsConnectionFactory_CreateXAConnection(factory.cConnectionFactory, &cConnection, cUsername, cPassword)
	if status != tibems_OK {
		factory.partiallyCreatedConnection = nil
		globalFactoryConnectionCreateFactory = nil
		globalFactoryConnectionCreateLock.Unlock()
		return nil, statusToError(status)
	}

	factory.partiallyCreatedConnection = nil
	globalFactoryConnectionCreateFactory = nil
	globalFactoryConnectionCreateLock.Unlock()

	connection.cConnection = cConnection
	return &connection, nil
}

func (factory *ConnectionFactory) CreateConnection(username string, password string) (*Connection, error) {
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

	connection := Connection{
		cConnection:             0,
		exceptionListenerHandle: nil,
		isXA:                    false,
		factory:                 factory,
	}

	var cConnection C.uintptr_t

	// About to call the C CreateConnection, which might call the OAuth2TokenFetch
	// callback inline. But we haven't finished creating our Go Connection object
	// yet (it's missing the cConnection pointer that we're about to create...), so
	// set a global reference (a thread-local reference would be better, but Golang
	// doesn't have thread-local storage) to this connection factory. If the
	// OAuth2TokenFetch does get called, it will first set the
	// factory.partiallyCreatedConnection.cConnection pointer before calling the
	// user's callback.
	globalFactoryConnectionCreateLock.Lock()
	factory.partiallyCreatedConnection = &connection
	globalFactoryConnectionCreateFactory = factory

	status := C.go_tibemsConnectionFactory_CreateConnection(factory.cConnectionFactory, &cConnection, cUsername, cPassword)
	if status != tibems_OK {
		factory.partiallyCreatedConnection = nil
		globalFactoryConnectionCreateFactory = nil
		globalFactoryConnectionCreateLock.Unlock()
		return nil, statusToError(status)
	}

	factory.partiallyCreatedConnection = nil
	globalFactoryConnectionCreateFactory = nil
	globalFactoryConnectionCreateLock.Unlock()

	// Might've already been set inline with an OAuth2TokenFetch callback if it was set,
	// but if so, we're just setting it to the same value again here.
	connection.cConnection = cConnection

	return &connection, nil
}

func (factory *ConnectionFactory) SetOAuth2Params(params *OAuth2Params) error {
	status := C.go_tibemsConnectionFactory_SetOAuth2Params(factory.cConnectionFactory, params.cOAuth2Params)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetOAuth2TokenFetchCallback(tokenFetchCallback OAuth2TokenFetchCallback) error {
	globalFactoryConnectionCreateLock.Lock()
	if tokenFetchCallback != nil {
		if factory.oauth2TokenFetchCallback == nil {
			// Set C callback to our generic one.
			C._setOAuth2TokenFetchCallback(factory.cConnectionFactory)
		}
	} else {
		C.go_tibemsConnectionFactory_SetOAuth2TokenFetchCallback(factory.cConnectionFactory, 0)
	}

	factory.oauth2TokenFetchCallback = tokenFetchCallback
	globalFactoryConnectionCreateLock.Unlock()
	return nil
}

func (factory *ConnectionFactory) SetConnectAttemptCount(count int32) error {
	status := C.go_tibemsConnectionFactory_SetConnectAttemptCount(factory.cConnectionFactory, C.tibems_int(count))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetConnectAttemptDelay(delay int32) error {
	status := C.go_tibemsConnectionFactory_SetConnectAttemptDelay(factory.cConnectionFactory, C.tibems_int(delay))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetConnectAttemptTimeout(timeout int32) error {
	status := C.go_tibemsConnectionFactory_SetConnectAttemptTimeout(factory.cConnectionFactory, C.tibems_int(timeout))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetReconnectAttemptCount(count int32) error {
	status := C.go_tibemsConnectionFactory_SetReconnectAttemptCount(factory.cConnectionFactory, C.tibems_int(count))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetReconnectAttemptDelay(delay int32) error {
	status := C.go_tibemsConnectionFactory_SetReconnectAttemptDelay(factory.cConnectionFactory, C.tibems_int(delay))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetReconnectAttemptTimeout(timeout int32) error {
	status := C.go_tibemsConnectionFactory_SetReconnectAttemptTimeout(factory.cConnectionFactory, C.tibems_int(timeout))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetClientID(clientId string) error {
	cClientId := C.CString(clientId)
	defer C.free(unsafe.Pointer(cClientId))
	status := C.go_tibemsConnectionFactory_SetClientID(factory.cConnectionFactory, cClientId)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetUserName(username string) error {
	cUsername := C.CString(username)
	defer C.free(unsafe.Pointer(cUsername))
	status := C.go_tibemsConnectionFactory_SetUserName(factory.cConnectionFactory, cUsername)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetUserPassword(password string) error {
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	status := C.go_tibemsConnectionFactory_SetUserPassword(factory.cConnectionFactory, cPassword)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetPkPassword(password string) error {
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	status := C.go_tibemsConnectionFactory_SetPkPassword(factory.cConnectionFactory, cPassword)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetMetric(metric FactoryLoadBalanceMetric) error {
	status := C.go_tibemsConnectionFactory_SetMetric(factory.cConnectionFactory, C.tibemsFactoryLoadBalanceMetric(metric))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetSSLParams(sslParams *SSLParams) error {
	status := C.go_tibemsConnectionFactory_SetSSLParams(factory.cConnectionFactory, sslParams.cSSLParams)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetSSLProxy(proxyHost string, proxyPort uint16) error {
	cHost := C.CString(proxyHost)
	defer C.free(unsafe.Pointer(cHost))
	status := C.go_tibemsConnectionFactory_SetSSLProxy(factory.cConnectionFactory, cHost, C.tibems_int(proxyPort))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) SetSSLProxyAuth(proxyUser string, proxyPassword string) error {
	cUser := C.CString(proxyUser)
	defer C.free(unsafe.Pointer(cUser))
	cPassword := C.CString(proxyPassword)
	defer C.free(unsafe.Pointer(cPassword))
	status := C.go_tibemsConnectionFactory_SetSSLProxyAuth(factory.cConnectionFactory, cUser, cPassword)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (factory *ConnectionFactory) GetSSLProxy() (string, uint16, error) {
	var cValue *C.char
	status := C.go_tibemsConnectionFactory_GetSSLProxyHost(factory.cConnectionFactory, &cValue)
	if status != tibems_OK {
		return "", 0, getErrorFromStatus(status)
	}
	var cPort C.tibems_int
	status = C.go_tibemsConnectionFactory_GetSSLProxyPort(factory.cConnectionFactory, &cPort)
	if status != tibems_OK {
		return "", 0, getErrorFromStatus(status)
	}
	return C.GoString(cValue), uint16(cPort), nil
}

func (factory *ConnectionFactory) GetSSLProxyAuth() (string, string, error) {
	var cUser *C.char
	status := C.go_tibemsConnectionFactory_GetSSLProxyUser(factory.cConnectionFactory, &cUser)
	if status != tibems_OK {
		return "", "", getErrorFromStatus(status)
	}
	var cPassword *C.char
	status = C.go_tibemsConnectionFactory_GetSSLProxyPassword(factory.cConnectionFactory, &cPassword)
	if status != tibems_OK {
		return "", "", getErrorFromStatus(status)
	}
	return C.GoString(cUser), C.GoString(cPassword), nil
}

func (factory *ConnectionFactory) String() string {
	bufSize := C.size_t(16384)
	buf := C.malloc(bufSize)
	status := C.go_tibemsConnectionFactory_PrintToBuffer(factory.cConnectionFactory, (*C.char)(buf), C.tibems_int(bufSize-1))
	for status == tibems_INSUFFICIENT_BUFFER {
		C.free(buf)
		bufSize *= 2
		buf = C.malloc(bufSize)
		status = C.go_tibemsConnectionFactory_PrintToBuffer(factory.cConnectionFactory, (*C.char)(buf), C.tibems_int(bufSize-1))
	}
	if status != tibems_OK {
		C.free(buf)
		return "<error>"
	}
	str := C.GoString((*C.char)(buf))
	C.free(buf)
	return str
}
