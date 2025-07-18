package oracledb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/tibco/wi-oracledb/src/app/OracleDatabase/connector/oracledb"
)

var spchartestConnectionJSON = []byte(`{
	"name": "oracle",
	"description": "",
	"host": "10.102.137.151",
	"port": 1521,
	"user": "spchartest",
	"password": "T~i!b@c#o$i%s^te*s(t)i_n+g20&",
	"database": "Service Name",
	"SID": "orcl.apac.tibco.com"
}`)

var colontestConnectionJSON = []byte(`{
	"name": "oracle",
	"description": "",
	"host": "10.102.137.151",
	"port": 1521,
	"user": "orcl120",
	"password": "orcl:120",
	"database": "Service Name",
	"SID": "orcl.apac.tibco.com"
}`)

// Adding this test to see if a connection is created with special characters. No other tests required
func TestSpecialCharPasswordUser(t *testing.T) {
	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(spchartestConnectionJSON), &conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	odb := &oracledb.OracleDatabaseFactory{}
	connManager, err := odb.NewManager(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db := connManager.GetConnection().(*sql.DB)

	err = db.Ping()

	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
}
