/*Package connection ...
 */
package connection

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"

	"github.com/alexbrainman/odbc/api"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
)

// MYworkdir ... important variable
var MYworkdir string

func init() {
	dbdrv = DbDrv{&coredbutils.Drv}
}

//IsAlive will take COnnection handle Input and check if connection is alive or not returns boolean
func IsAlive(ConnHandle api.SQLHDBC) error {
	var outvalue api.SQLINTEGER
	var outlen api.SQLINTEGER = 0
	api.SQLGetConnectAttr(ConnHandle, api.SQL_ATTR_CONNECTION_DEAD, api.SQLPOINTER(&outvalue), 0, &outlen)
	if outvalue == api.SQL_CD_FALSE {
		logCache.Debug("Connection is alive SQL_ATTR_CONNECTION_DEAD is False")
		return nil
	} else {
		logCache.Debug("Connection is dead SQL_ATTR_CONNECTION_DEAD is True")
		return fmt.Errorf("[ SQL_ATTR_CONNECTION_DEAD = True ]")
	}
}

//CustomPing will hit Select 1 Query on Provided Statement Handle
func CustomPing(stmt api.SQLHSTMT, query string) (err error) {
	queryText, err := syscall.ByteSliceFromString(query)
	if err != nil {
		return fmt.Errorf("could Not convert query in byte slice format : %v", err)
	}
	StatementLength := api.SQL_NTS
	ret := api.SQLExecDirect(stmt, (*api.SQLCHAR)(&queryText[0]), api.SQLINTEGER(StatementLength))
	if coredbutils.IsError(ret) {
		defer coredbutils.ReleaseHandle(stmt)
		return coredbutils.NewError("SQLExecDirect", stmt)
	}
	return nil
}

// CreateUniqueConnectionID ...
func CreateUniqueConnectionID(cdetails string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%v", cdetails)))
	return hex.EncodeToString(hash[:])
}

// GetConnection returns the connection
func GetConnection(cid string) (ConnPool, bool) {
	conn, ok := connpool[cid]
	if ok {
		if time.Since(conn.Timeout) > time.Second*60 {
			logCache.Info("Connection timed out dropping connection...")
			DropConnectionFromPool(cid)
			ok = false
		} else {
			logCache.Debug("Resetting the timeout for connection")
			conn, ok = connpool[cid]
			if ok {
				conn.Timeout = time.Now()
			}
		}
	}
	return conn, ok
}

// CleanUp ...
func CleanUp(cleanUpDataFiles []string) {
	logCache.Info("Cleaning up temp files...", cleanUpDataFiles)
	for _, value := range cleanUpDataFiles {
		os.Remove(value)
	}
}

// DropConnectionFromPool ...
func DropConnectionFromPool(cid string) {
	conn, ok := connpool[cid]
	if ok {
		delete(connpool, cid)
		CleanUp(conn.Files)
		err := conn.Connection.Disconnect()
		if err != nil {
			logCache.Info("Getting Error Here {DropConnectionFromPool}")
		}
		// delete(connpool, cid)
	}
}

// Connect ...
func (d *DbDrv) Connect(dsn string, connectionName string) (*Conn, string, error) {
	// connID := CreateConnID(dsn, &cd)
	//--myNote uniqueConnId gives same string if data is same
	connID := CreateUniqueConnectionID(dsn)

	// connection, ok := GetConnection(connID)
	// if !ok && cd.SSLMode != "" {
	// 	newdsn, err := FormatConnectionString(&cd)
	// 	if err != nil {
	// 		logCache .Info(err)
	// 		return nil, "", err
	// 	}
	// 	dsn = newdsn
	// }
	// if ok {
	// 	logCache.Debug("sending pooled connection from cgo")
	// 	return connection.Connection, connection.ConnID, nil
	// }
	var out api.SQLHANDLE
	ret := api.SQLAllocHandle(api.SQL_HANDLE_DBC, api.SQLHANDLE(d.H), &out)
	if coredbutils.IsError(ret) {
		//On Any error while Allocation Handle Trying to allocate handle again
		coredbutils.ReleaseHandle(api.SQLHDBC(out))
		ret = api.SQLAllocHandle(api.SQL_HANDLE_DBC, api.SQLHANDLE(d.H), &out)
	}
	h := api.SQLHDBC(out)
	err := coredbutils.Drv.Stats.UpdateHandleCount(api.SQL_HANDLE_DBC, 1)
	if err != nil {
		return nil, "", err
	}
	b := syscall.StringByteSlice(dsn)
	ret = api.SQLDriverConnect(h, 0,
		(*api.SQLCHAR)(unsafe.Pointer(&b[0])), api.SQL_NTS,
		nil, 0, nil, api.SQL_DRIVER_NOPROMPT)
	if coredbutils.IsError(ret) {
		defer coredbutils.ReleaseHandle(h)
		return nil, "", coredbutils.NewError("SQLDriverConnect", h)
	}
	// if len(cd.SSLFiles) < 1 {
	connpool[connID] = ConnPool{
		ConnID:     connID,
		ConnName:   connectionName,
		Timeout:    time.Now(),
		Connection: &Conn{H: h},
		//Files:      s["SSLFiles"],
	}
	// }
	return &Conn{H: h}, connID, nil
}

// Disconnect ...
func (c *Conn) Disconnect() (err error) {
	h := c.H
	defer func() {
		c.H = api.SQLHDBC(api.SQL_NULL_HDBC)
		coredbutils.ReleaseHandle(h)
	}()
	ret := api.SQLDisconnect(c.H)
	if coredbutils.IsError(ret) {
		return c.NewError("SQLDisconnect", h)
	}
	return err
}

// NewError ...
func (c *Conn) NewError(apiName string, handle interface{}) error {
	err := coredbutils.NewError(apiName, handle)
	return err
}
