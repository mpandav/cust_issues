package tibems

/*
#include <inttypes.h>
#include <tibems/tibems.h>

#ifndef XID_T
struct xid_t
{
  long formatID;
  long gtrid_length;
  long bqual_length;
  char data[128];
};
#endif

#include <tibems/xares.h>

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
go_tibemsXAConnection_Get(
    uintptr_t*  connection,
    const char* brokerURL)
{
  return tibemsXAConnection_Get((tibemsConnection*)connection, brokerURL);
}

static tibems_status
go_tibemsXAConnection_GetXASession(
    uintptr_t  connection,
    uintptr_t* xaSession)
{
  return tibemsXAConnection_GetXASession((tibemsConnection)connection, (tibemsSession*)xaSession);
}

static tibems_status
go_tibemsXAResource_Commit(
    uintptr_t   xaResource,
    XID*        xid,
    tibems_bool onePhase)
{
  return tibemsXAResource_Commit((tibemsXAResource)xaResource, xid, onePhase);
}

static tibems_status
go_tibemsXAResource_End(
    uintptr_t xaResource,
    XID*      xid,
    int       flags)
{
  return tibemsXAResource_End((tibemsXAResource)xaResource, xid, flags);
}

static tibems_status
go_tibemsXAResource_GetTransactionTimeout(
    uintptr_t   xaResource,
    tibems_int* seconds)
{
  return tibemsXAResource_GetTransactionTimeout((tibemsXAResource)xaResource, seconds);
}

static tibems_status
go_tibemsXAResource_isSameRM(
    uintptr_t    xaResource,
    uintptr_t    xaResource2,
    tibems_bool* result)
{
  return tibemsXAResource_isSameRM((tibemsXAResource)xaResource, (tibemsXAResource)xaResource2, result);
}

static tibems_status
go_tibemsXAResource_Prepare(
    uintptr_t xaResource,
    XID*      xid)
{
  return tibemsXAResource_Prepare((tibemsXAResource)xaResource, xid);
}

static tibems_status
go_tibemsXAResource_Recover(
    uintptr_t   xaResource,
    XID*        xids,
    tibems_int  desiredCount,
    tibems_int* returnedCount,
    tibems_int  flag)
{
  return tibemsXAResource_Recover((tibemsXAResource)xaResource, xids, desiredCount, returnedCount, flag);
}

static tibems_status
go_tibemsXAResource_Rollback(
    uintptr_t xaResource,
    XID*      xid)
{
  return tibemsXAResource_Rollback((tibemsXAResource)xaResource, xid);
}

static tibems_status
go_tibemsXAResource_SetTransactionTimeout(
    uintptr_t  xaResource,
    tibems_int seconds)
{
  return tibemsXAResource_SetTransactionTimeout((tibemsXAResource)xaResource, seconds);
}

static tibems_status
go_tibemsXAResource_Start(
    uintptr_t  xaResource,
    XID*       xid,
    tibems_int flags)
{
  return tibemsXAResource_Start((tibemsXAResource)xaResource, xid, flags);
}

static tibems_status
go_tibemsXAResource_SetRMID(
    uintptr_t  xaResource,
    tibems_int rmid)
{
  return tibemsXAResource_SetRMID((tibemsXAResource)xaResource, rmid);
}

static tibems_status
go_tibemsXAResource_GetRMID(
    uintptr_t   xaResource,
    tibems_int* rmid)
{
  return tibemsXAResource_GetRMID((tibemsXAResource)xaResource, rmid);
}

static tibems_status
go_tibemsXASession_GetXAResource(
    uintptr_t  session,
    uintptr_t* xaResource)
{
  return tibemsXASession_GetXAResource((tibemsSession)session, (tibemsXAResource*)xaResource);
}

*/
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

type XAConnection struct {
	Connection
}

// Ensure XAConnection implements EMSConnection
var _ EMSConnection = (*XAConnection)(nil)

type XASession struct {
	Session
}

// Ensure XASession implements EMSSession
var _ EMSSession = (*XASession)(nil)

type XID struct {
	FormatID                  int32
	GlobalTransactionIDLength int32
	BranchQualifierLength     int32
	ID                        [128]byte
}

func (xid *XID) toCxid(cXID *C.XID) {
	cXID.formatID = C.long(xid.FormatID)
	cXID.gtrid_length = C.long(xid.GlobalTransactionIDLength)
	cXID.bqual_length = C.long(xid.BranchQualifierLength)
	C.memcpy(unsafe.Pointer(&cXID.data), unsafe.Pointer(&xid.ID[0]), 128)
}

func xidFromCxid(cXID *C.XID) *XID {
	var xid XID
	xid.FormatID = int32(cXID.formatID)
	xid.GlobalTransactionIDLength = int32(cXID.gtrid_length)
	xid.BranchQualifierLength = int32(cXID.bqual_length)
	C.memcpy(unsafe.Pointer(&(xid.ID[0])), unsafe.Pointer(&cXID.data), 128)
	return &xid
}

// Exported for testing.
//
//export xidHandleFromCxid
func xidHandleFromCxid(cXID *C.XID) uintptr {
	var xid XID
	xid.FormatID = int32(cXID.formatID)
	xid.GlobalTransactionIDLength = int32(cXID.gtrid_length)
	xid.BranchQualifierLength = int32(cXID.bqual_length)
	C.memcpy(unsafe.Pointer(&(xid.ID[0])), unsafe.Pointer(&cXID.data), 128)
	return uintptr(cgo.NewHandle(&xid))
}

// Used for testing.
//
//export deleteHandle
func deleteHandle(handle uintptr) {
	cgo.Handle(handle).Delete()
}

type XAResource struct {
	cXAResource C.uintptr_t
}

// CreateXAConnection attempts to connect to the EMS server and returns a new Connection object if successful.
func CreateXAConnection(brokerURL string, clientdId string, username string, password string, options *ConnectionOptions) (*XAConnection, error) {
	connection, err := createConnection(brokerURL, clientdId, username, password, options, true)
	if err != nil {
		return nil, err
	}
	return connection.(*XAConnection), nil
}

// CreateXASession creates a new XA session within the given XA connection.
func (connection *XAConnection) CreateXASession() (*XASession, error) {
	session, err := connection.createSession(true, AckModeNoAcknowledge, true)
	if err != nil {
		return nil, err
	}
	return session.(*XASession), nil
}

func (connection *XAConnection) GetXASession() (*XASession, error) {
	session := XASession{Session{
		cSession:   0,
		ackMode:    0,
		transacted: true,
		isXA:       true,
	}}
	status := C.go_tibemsXAConnection_GetXASession(connection.cConnection, &(session.cSession))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &session, nil
}

func (session *XASession) GetXAResource() (*XAResource, error) {
	res := XAResource{cXAResource: 0}
	status := C.go_tibemsXASession_GetXAResource(session.cSession, &(res.cXAResource))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}

	return &res, nil
}

func (session *XASession) GetSession() (*Session, error) {
	return &Session{
		cSession:   session.cSession,
		ackMode:    session.ackMode,
		transacted: session.transacted,
		isXA:       false,
	}, nil
}

func XAConnectionGet(brokerURL string) (*XAConnection, error) {
	connection := XAConnection{Connection{
		cConnection:             0,
		exceptionListenerHandle: nil,
		isXA:                    true,
	}}
	cBrokerURL := C.CString(brokerURL)
	defer C.free(unsafe.Pointer(cBrokerURL))
	status := C.go_tibemsXAConnection_Get(&(connection.cConnection), cBrokerURL)
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	return &connection, nil
}

func (resource *XAResource) Commit(xid *XID, onePhase bool) error {
	var cOnePhase C.tibems_bool
	if onePhase {
		cOnePhase = 1
	} else {
		cOnePhase = 0
	}
	var cXID C.XID
	xid.toCxid(&cXID)
	status := C.go_tibemsXAResource_Commit(resource.cXAResource, &cXID, cOnePhase)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (resource *XAResource) End(xid *XID, flags int) error {
	var cXID C.XID
	xid.toCxid(&cXID)
	status := C.go_tibemsXAResource_End(resource.cXAResource, &cXID, C.int(flags))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (resource *XAResource) GetRMID() (int32, error) {
	var cRMID C.tibems_int
	status := C.go_tibemsXAResource_GetRMID(resource.cXAResource, &cRMID)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return int32(cRMID), nil
}
func (resource *XAResource) GetTransactionTimeout() (int32, error) {
	var cValue C.tibems_int
	status := C.go_tibemsXAResource_GetTransactionTimeout(resource.cXAResource, &cValue)
	if status != tibems_OK {
		return 0, getErrorFromStatus(status)
	}
	return int32(cValue), nil
}

func (resource *XAResource) IsSameRM(resource2 *XAResource) (bool, error) {
	var cValue C.tibems_bool
	status := C.go_tibemsXAResource_isSameRM(resource.cXAResource, resource2.cXAResource, &cValue)
	if status != tibems_OK {
		return false, getErrorFromStatus(status)
	}
	if cValue == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (resource *XAResource) Prepare(xid *XID) error {
	var cXID C.XID
	xid.toCxid(&cXID)
	status := C.go_tibemsXAResource_Prepare(resource.cXAResource, &cXID)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (resource *XAResource) Recover(desiredCount int32, flag int32) ([]*XID, error) {
	var cReturnedCount C.tibems_int
	cXidArray := C.malloc(C.sizeof_XID * C.size_t(desiredCount))
	defer C.free(cXidArray)
	status := C.go_tibemsXAResource_Recover(resource.cXAResource, (*C.XID)(cXidArray), C.tibems_int(desiredCount), &cReturnedCount, C.tibems_int(flag))
	if status != tibems_OK {
		return nil, getErrorFromStatus(status)
	}
	xidArray := make([]*XID, int(cReturnedCount))
	for i := C.tibems_int(0); i < cReturnedCount; i++ {
		xidArray[i] = xidFromCxid((*C.XID)(unsafe.Pointer(uintptr(cXidArray) + uintptr(i*C.sizeof_XID))))
	}

	return xidArray, nil
}

func (resource *XAResource) Rollback(xid *XID) error {
	var cXID C.XID
	xid.toCxid(&cXID)
	status := C.go_tibemsXAResource_Rollback(resource.cXAResource, &cXID)
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (resource *XAResource) SetRMID(rmid int32) error {
	status := C.go_tibemsXAResource_SetRMID(resource.cXAResource, C.tibems_int(rmid))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (resource *XAResource) SetTransactionTimeout(seconds int32) error {
	status := C.go_tibemsXAResource_SetTransactionTimeout(resource.cXAResource, C.tibems_int(seconds))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}

func (resource *XAResource) Start(xid *XID, flags int) error {
	var cXID C.XID
	xid.toCxid(&cXID)
	status := C.go_tibemsXAResource_Start(resource.cXAResource, &cXID, C.int(flags))
	if status != tibems_OK {
		return getErrorFromStatus(status)
	}
	return nil
}
