package tibems

/*
#include <inttypes.h>
#include <tibems/tibems.h>

extern void
_tibemsMem_Free_traced(
    void* ptr,
    char* file,
    int   line);

extern void
goExceptionCallback(
    tibems_status,
    uintptr_t);

static void
_emsFree(void* ptr)
{
  _tibemsMem_Free_traced(ptr, __FILE__, __LINE__);
}

static void
_cExceptionCallback(
    tibemsConnection connection,
    tibems_status    status,
    void*            closure)
{
  goExceptionCallback(status, (uintptr_t)closure);
}

static tibems_status
_cSetExceptionCallback(
    uintptr_t connection,
    uintptr_t handle)
{
  return tibemsConnection_SetExceptionListener((tibemsConnection)connection, _cExceptionCallback, (void*)handle);
}

static void
_callExceptionCallback(
    uintptr_t     cb,
    uintptr_t     connHandle,
    tibems_status status,
    uintptr_t     closure)
{
  ((tibemsExceptionCallback)cb)((tibemsConnection)connHandle, status, (void*)closure);
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
go_tibemsConnectionMetaData_GetEMSMajorVersion(
    uintptr_t   metaData,
    tibems_int* majorVersion)
{
	return tibemsConnectionMetaData_GetEMSMajorVersion((tibemsConnectionMetaData)metaData, majorVersion);
}

static tibems_status
go_tibemsConnectionMetaData_GetEMSMinorVersion(
    uintptr_t   metaData,
    tibems_int* minorVersion)
{
	return tibemsConnectionMetaData_GetEMSMinorVersion((tibemsConnectionMetaData)metaData, minorVersion);
}

static tibems_status
go_tibemsConnectionMetaData_GetEMSVersion(
    uintptr_t    metaData,
    const char** version)
{
	return tibemsConnectionMetaData_GetEMSVersion((tibemsConnectionMetaData)metaData, version);
}

static tibems_status
go_tibemsConnectionMetaData_GetEMSProviderName(
    uintptr_t    metaData,
    const char** providerName)
{
	return tibemsConnectionMetaData_GetEMSProviderName((tibemsConnectionMetaData)metaData, providerName);
}

static tibems_status
go_tibemsConnectionMetaData_GetProviderMajorVersion(
    uintptr_t   metaData,
    tibems_int* majorVersion)
{
	return tibemsConnectionMetaData_GetProviderMajorVersion((tibemsConnectionMetaData)metaData, majorVersion);
}

static tibems_status
go_tibemsConnectionMetaData_GetProviderMinorVersion(
    uintptr_t   metaData,
    tibems_int* minorVersion)
{
	return tibemsConnectionMetaData_GetProviderMinorVersion((tibemsConnectionMetaData)metaData, minorVersion);
}

static tibems_status
go_tibemsConnectionMetaData_GetProviderVersion(
    uintptr_t    metaData,
    const char** version)
{
	return tibemsConnectionMetaData_GetProviderVersion((tibemsConnectionMetaData)metaData, version);
}

static tibems_status
go_tibemsConnection_Create(
    uintptr_t*  connection,
    const char* brokerURL,
    const char* clientId,
    const char* username,
    const char* password)
{
  return tibemsConnection_Create((tibemsConnection*)connection, brokerURL, clientId, username, password);
}

static tibems_status
go_tibemsConnection_CreateSession(
    uintptr_t             connection,
    uintptr_t*            session,
    tibems_bool           transacted,
    tibemsAcknowledgeMode acknowledgeMode)
{
  return tibemsConnection_CreateSession(
      (tibemsConnection)connection, (tibemsSession*)session, transacted, acknowledgeMode);
}

static tibems_status
go_tibemsConnection_Close(uintptr_t conn)
{
  return tibemsConnection_Close((tibemsConnection)conn);
}

static tibems_status
go_tibemsConnection_Start(uintptr_t connection)
{
  return tibemsConnection_Start((tibemsConnection)connection);
}

static tibems_status
go_tibemsConnection_Stop(uintptr_t connection)
{
  return tibemsConnection_Stop((tibemsConnection)connection);
}

static tibems_status
go_tibemsConnection_GetClientId(
    uintptr_t    connection,
    const char** clientId)
{
  return tibemsConnection_GetClientId((tibemsConnection)connection, clientId);
}

static tibems_status
go_tibemsConnection_SetClientId(
    uintptr_t   connection,
    const char* clientId)
{
  return tibemsConnection_SetClientId((tibemsConnection)connection, clientId);
}

static tibems_status
go_tibemsConnection_GetMetaData(
    uintptr_t  connection,
    uintptr_t* metaData)
{
  return tibemsConnection_GetMetaData((tibemsConnection)connection, (tibemsConnectionMetaData*)metaData);
}

static tibems_status
go_tibemsConnection_GetActiveURL(
    uintptr_t connection,
    char**    serverURL)
{
  return tibemsConnection_GetActiveURL((tibemsConnection)connection, serverURL);
}

static tibems_status
go_tibemsConnection_IsDisconnected(
    uintptr_t    connection,
    tibems_bool* disconnected)
{
  return tibemsConnection_IsDisconnected((tibemsConnection)connection, disconnected);
}

static tibems_status
go_tibemsXAConnection_Close(uintptr_t connection)
{
  return tibemsXAConnection_Close((tibemsConnection)connection);
}

static tibems_status
go_tibemsConnection_CreateSSL(
    uintptr_t*  connection,
    const char* brokerURL,
    const char* clientId,
    const char* username,
    const char* password,
    uintptr_t   params,
    const char* pk_password)
{
  return tibemsConnection_CreateSSL(
      (tibemsConnection*)connection, brokerURL, clientId, username, password, (tibemsSSLParams)params, pk_password);
}

static tibems_status
go_tibemsXAConnection_Create(
    uintptr_t*  connection,
    const char* brokerURL,
    const char* clientId,
    const char* username,
    const char* password)
{
  return tibemsXAConnection_Create((tibemsConnection*)connection, brokerURL, clientId, username, password);
}

static tibems_status
go_tibemsXAConnection_CreateSSL(
    uintptr_t*  connection,
    const char* brokerURL,
    const char* clientId,
    const char* username,
    const char* password,
    uintptr_t   params,
    const char* pk_password)
{
  return tibemsXAConnection_CreateSSL(
      (tibemsConnection*)connection, brokerURL, clientId, username, password, (tibemsSSLParams)params, pk_password);
}

static tibems_status
go_tibemsXAConnection_CreateXASession(
    uintptr_t  connection,
    uintptr_t* session)
{
  return tibemsXAConnection_CreateXASession((tibemsConnection)connection, (tibemsSession*)session);
}

*/
import "C"
import (
	"github.com/cockroachdb/errors"
	"runtime/cgo"
	"unsafe"
)

// Ensure Connection implements EMSConnection
var _ EMSConnection = (*Connection)(nil)

// A Connection represents a single connection to an EMS server.
type Connection struct {
	cConnection             C.uintptr_t
	exceptionListenerHandle *cgo.Handle
	isXA                    bool
	factory                 *ConnectionFactory
}

type EMSConnection interface {
	Close() error
	GetClientID() (string, error)
	SetClientID(clientId string) error
	IsDisconnected() (bool, error)
	GetMetaData() (*ConnectionMetaData, error)
	Start() error
	Stop() error
	CreateSession(transacted bool, ackMode AcknowledgeMode) (*Session, error)
	GetActiveURL() (string, error)
	SetExceptionListener(listenerCallback ExceptionCallback) error
	GetExceptionListener() ExceptionCallback
}

type ConnectionOptions struct {
	SSLParams          *SSLParams
	PrivateKeyPassword string
}

// Used for testing.
//
//export createConnectionOptions
func createConnectionOptions(sslParamsHandle uintptr, pkPass *C.char) C.uintptr_t {
	opts := &ConnectionOptions{
		SSLParams:          cgo.Handle(sslParamsHandle).Value().(*SSLParams),
		PrivateKeyPassword: C.GoString(pkPass),
	}
	return C.uintptr_t(cgo.NewHandle(opts))
}

func createConnection(brokerURL string, clientdId string, username string, password string, options *ConnectionOptions, isXA bool) (EMSConnection, error) {
	var connection EMSConnection

	cBrokerURL := C.CString(brokerURL)
	defer C.free(unsafe.Pointer(cBrokerURL))
	var cClientId *C.char
	if clientdId != "" {
		cClientId = C.CString(clientdId)
		defer C.free(unsafe.Pointer(cClientId))
	}
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
	var cPrivateKeyPassword *C.char
	if options != nil {
		if options.SSLParams != nil {
			cSSLParams = options.SSLParams.cSSLParams
		}
		if options.PrivateKeyPassword != "" {
			cPrivateKeyPassword = C.CString(options.PrivateKeyPassword)
			defer C.free(unsafe.Pointer(cPrivateKeyPassword))
		}
	}

	var status C.tibems_status
	if isXA {
		conn := XAConnection{Connection{
			cConnection:             0,
			exceptionListenerHandle: nil,
			isXA:                    true,
		}}
		connection = &conn
		if cSSLParams != 0 {
			status = C.go_tibemsXAConnection_CreateSSL(&conn.cConnection, cBrokerURL, cClientId, cUsername, cPassword, cSSLParams, cPrivateKeyPassword)
		} else {
			status = C.go_tibemsXAConnection_Create(&conn.cConnection, cBrokerURL, cClientId, cUsername, cPassword)
		}
	} else {
		conn := Connection{
			cConnection:             0,
			exceptionListenerHandle: nil,
			isXA:                    false,
		}
		connection = &conn
		if cSSLParams != 0 {
			status = C.go_tibemsConnection_CreateSSL(&conn.cConnection, cBrokerURL, cClientId, cUsername, cPassword, cSSLParams, cPrivateKeyPassword)
		} else {
			status = C.go_tibemsConnection_Create(&conn.cConnection, cBrokerURL, cClientId, cUsername, cPassword)
		}
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

	return connection, nil
}

// CreateConnection attempts to connect to the EMS server and returns a new Connection object if successful.
func CreateConnection(brokerURL string, clientdId string, username string, password string, options *ConnectionOptions) (*Connection, error) {
	connection, err := createConnection(brokerURL, clientdId, username, password, options, false)
	if err != nil {
		return nil, err
	}
	return connection.(*Connection), nil
}

// Close stops a Connection and releases resources associated with it. Applications MUST call Close when finished
// with a Connection object to avoid resource leaks. A Connection object MUST NOT be used after Close has been called.
func (connection *Connection) Close() error {
	if connection.factory != nil {
		cConnectionToConnectionMapLock.Lock()
	}

	var status C.tibems_status
	if connection.isXA {
		status = C.go_tibemsXAConnection_Close(connection.cConnection)
	} else {
		status = C.go_tibemsConnection_Close(connection.cConnection)
	}

	if connection.factory != nil {
		delete(cConnectionToConnectionMap, connection.cConnection)
		cConnectionToConnectionMapLock.Unlock()
	}

	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	if connection.exceptionListenerHandle != nil {
		connection.exceptionListenerHandle.Delete()
		connection.exceptionListenerHandle = nil
	}

	return nil
}

func (connection *Connection) GetClientID() (string, error) {
	var cValue *C.char
	status := C.go_tibemsConnection_GetClientId(connection.cConnection, &cValue)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}

	return C.GoString(cValue), nil
}

func (connection *Connection) SetClientID(clientId string) error {
	cClientId := C.CString(clientId)
	defer C.free(unsafe.Pointer(cClientId))
	status := C.go_tibemsConnection_SetClientId(connection.cConnection, cClientId)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (connection *Connection) IsDisconnected() (bool, error) {
	var cDisconnected C.tibems_bool
	status := C.go_tibemsConnection_IsDisconnected(connection.cConnection, &cDisconnected)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cDisconnected == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

type ConnectionMetaData struct {
	EMSMajorVersion      int32
	EMSMinorVersion      int32
	EMSVersion           string
	ProviderMajorVersion int32
	ProviderMinorVersion int32
	ProviderVersion      string
	EMSProviderName      string
}

// Used for testing.
//
//export unpackMetaData
func unpackMetaData(metaDataHandle uintptr) (int32, int32, *C.char, int32, int32, *C.char, *C.char) {
	metaData := cgo.Handle(metaDataHandle).Value().(*ConnectionMetaData)
	return metaData.EMSMajorVersion, metaData.EMSMinorVersion, C.CString(metaData.EMSVersion), metaData.ProviderMajorVersion, metaData.ProviderMinorVersion, C.CString(metaData.ProviderVersion), C.CString(metaData.EMSProviderName)
}

// GetMetaData returns a ConnectionMetaData object with information about the current connection's JMS standard version.
func (connection *Connection) GetMetaData() (*ConnectionMetaData, error) {
	var cMetaData C.uintptr_t
	status := C.go_tibemsConnection_GetMetaData(connection.cConnection, &cMetaData)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cEMSMajorVersion C.tibems_int
	status = C.go_tibemsConnectionMetaData_GetEMSMajorVersion(cMetaData, &cEMSMajorVersion)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cEMSMinorVersion C.tibems_int
	status = C.go_tibemsConnectionMetaData_GetEMSMinorVersion(cMetaData, &cEMSMinorVersion)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cEMSVersion *C.char
	status = C.go_tibemsConnectionMetaData_GetEMSVersion(cMetaData, &cEMSVersion)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cProviderMajorVersion C.tibems_int
	status = C.go_tibemsConnectionMetaData_GetProviderMajorVersion(cMetaData, &cProviderMajorVersion)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cProviderMinorVersion C.tibems_int
	status = C.go_tibemsConnectionMetaData_GetProviderMinorVersion(cMetaData, &cProviderMinorVersion)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cProviderVersion *C.char
	status = C.go_tibemsConnectionMetaData_GetProviderVersion(cMetaData, &cProviderVersion)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	var cEMSProviderName *C.char
	status = C.go_tibemsConnectionMetaData_GetEMSProviderName(cMetaData, &cEMSProviderName)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &ConnectionMetaData{
		EMSMajorVersion:      int32(cEMSMajorVersion),
		EMSMinorVersion:      int32(cEMSMinorVersion),
		EMSVersion:           C.GoString(cEMSVersion),
		ProviderMajorVersion: int32(cProviderMajorVersion),
		ProviderMinorVersion: int32(cProviderMinorVersion),
		ProviderVersion:      C.GoString(cProviderVersion),
		EMSProviderName:      C.GoString(cEMSProviderName),
	}, nil
}

// Start causes a Connection to begin listening for messages.
func (connection *Connection) Start() error {
	status := C.go_tibemsConnection_Start(connection.cConnection)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

// Stop causes a connection to stop listening for messages.
func (connection *Connection) Stop() error {
	status := C.go_tibemsConnection_Stop(connection.cConnection)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	return nil
}

func (connection *Connection) createSession(transacted bool, ackMode AcknowledgeMode, isXA bool) (EMSSession, error) {
	var cTransacted C.tibems_bool
	if transacted {
		cTransacted = 1
	} else {
		cTransacted = 0
	}
	var session EMSSession
	var status C.tibems_status
	if isXA {
		sess := XASession{Session{
			cSession:   0,
			ackMode:    AckModeSessionTransacted,
			transacted: true,
			isXA:       true,
		}}
		session = &sess
		status = C.go_tibemsXAConnection_CreateXASession(connection.cConnection, &sess.cSession)
	} else {
		sess := Session{
			cSession:   0,
			ackMode:    ackMode,
			transacted: transacted,
			isXA:       false,
		}
		session = &sess
		status = C.go_tibemsConnection_CreateSession(connection.cConnection, &sess.cSession, cTransacted, C.tibemsAcknowledgeMode(ackMode))
	}
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return session, nil
}

// CreateSession creates a new session with the given transaction and message acknowledgment modes.
func (connection *Connection) CreateSession(transacted bool, ackMode AcknowledgeMode) (*Session, error) {
	session, err := connection.createSession(transacted, ackMode, false)
	if err != nil {
		return nil, err
	}
	return session.(*Session), nil
}

// GetActiveURL returns the URL of the EMS server the Connection is currently connected to.
func (connection *Connection) GetActiveURL() (string, error) {
	var cURL *C.char
	status := C.go_tibemsConnection_GetActiveURL(connection.cConnection, &cURL)
	if status != tibems_OK {
		return "", getErrorFromStatus(status)
	}
	defer C._emsFree(unsafe.Pointer(cURL))

	return C.GoString(cURL), nil
}

// Used for testing.
//
//export createCExceptionCallback
func createCExceptionCallback(cFunc C.uintptr_t, cClosure C.uintptr_t) uintptr {
	var fn ExceptionCallback
	fn = func(connection *Connection, err error) {
		connHandle := cgo.NewHandle(connection)
		C._callExceptionCallback(cFunc, C.uintptr_t(connHandle), getStatusFromError(err), cClosure)
		connHandle.Delete()
	}
	return uintptr(cgo.NewHandle(fn))
}

// An ExceptionCallback delivers connection error events like server disconnects or (if enabled) failover events.
type ExceptionCallback func(connection *Connection, err error)

// SetExceptionListener sets a callback to be called on the event of connection errors, disconnects,
// or (if enabled) failover events.
func (connection *Connection) SetExceptionListener(listenerCallback ExceptionCallback) error {

	cbInfo := exceptionCallbackInfo{
		callback:   listenerCallback,
		connection: connection,
	}

	newHandle := cgo.NewHandle(cbInfo)
	status := C._cSetExceptionCallback(connection.cConnection, C.uintptr_t(newHandle))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}

	if connection.exceptionListenerHandle != nil {
		connection.exceptionListenerHandle.Delete()
		connection.exceptionListenerHandle = nil
	}
	connection.exceptionListenerHandle = &newHandle

	return nil
}

// GetExceptionListener returns the currently set (if any) connection exception listener callback.
func (connection *Connection) GetExceptionListener() ExceptionCallback {
	if connection.exceptionListenerHandle != nil {
		return connection.exceptionListenerHandle.Value().(exceptionCallbackInfo).callback
	} else {
		return nil
	}
}

type exceptionCallbackInfo struct {
	callback   ExceptionCallback
	connection *Connection
}

//export goExceptionCallback
func goExceptionCallback(status C.tibems_status, callbackInfoHandle C.uintptr_t) {
	handle := cgo.Handle(callbackInfoHandle)
	callback := handle.Value().(exceptionCallbackInfo)
	callback.callback(callback.connection, statusToError(status))
}
