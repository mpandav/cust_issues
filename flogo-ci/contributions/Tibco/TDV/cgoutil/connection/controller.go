package connection

import (
	"bytes"
	"fmt"

	"github.com/project-flogo/core/support/log"
)

var logCache = log.ChildLogger(log.RootLogger(), "tdv.connection")

func MapToString(m map[string]interface{}) (string, error) {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
	}
	return b.String(), nil
}

// ConnectDB ...
func ConnectDB(dsn string, connectionName string) (*Conn, string, error) {

	conn, Cachedconnid, err := dbdrv.Connect(dsn, connectionName)
	if err != nil {
		// CleanUp(cd.SSLFiles)
		return nil, "", err
	}
	//logCache.Debug(dbdrv.Stats)
	return conn, Cachedconnid, nil
}

// DisconnectDB ...
func DisconnectDB(connectionID string) {

	discon, ok := GetConnection(connectionID)
	if !ok {
		logCache.Debug(connectionID, " No Such Connection...")
		return
	}
	// Deleting Connection from cached map
	delete(connpool, connectionID)
	err := discon.Connection.Disconnect()
	if err != nil {
		logCache.Debug(err, "Disconnect Failed...")
		return
	}
	logCache.Debug("Disconnected SuccesFully")
	logCache.Info(&dbdrv.Stats)
	return
}
