/*Package connection ...
 */
package connection

import (
	"time"

	"github.com/alexbrainman/odbc/api"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/coredbutils"
)

// Conn holds the database connection handle
type Conn struct {
	H api.SQLHDBC `json:"-"`
}

// DbDrv directs requests to the environment handle
type DbDrv struct {
	*coredbutils.Driver
}

// ConnPool ...
type ConnPool struct {
	ConnID     string
	ConnName   string
	Timeout    time.Time
	Connection *Conn
	Files      []string
}

// Connection response
type Connection struct {
	ConnID string `json:"ConnectionId"`
}

var connpool = make(map[string]ConnPool)

var dbdrv DbDrv
