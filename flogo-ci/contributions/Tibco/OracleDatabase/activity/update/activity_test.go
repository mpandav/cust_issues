package update

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/tibco/wi-oracledb/src/app/OracleDatabase/connector/oracledb"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

var oracleConnectionJSON = []byte(`{
	"name": "oracle",
	"description": "",
	"host": "10.102.137.151",
	"port": 1521,
	"user": "tibco",
	"password": "tibco",
	"database": "Service Name",
	"SID": "orcl.apac.tibco.com"
}`)

func TestRegister(t *testing.T) {
	ref := activity.GetRef(&MyActivity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func setup() {
	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(oracleConnectionJSON), &conn)
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

	db = connManager.GetConnection().(*sql.DB)

	dropFLOGOSTUDENTTable := `drop table FLOGOSTUDENT`
	createFLOGOSTUDENTTable := `create table FLOGOSTUDENT (ROLLNUM number, STDNAME varchar(20), MARKS number, CLASS number)`
	insertIntoFLOGOSTUDENT := `INSERT ALL
							    into FLOGOSTUDENT values (1,'Tom',23,1)
								into FLOGOSTUDENT values (1,'Tom',25,1)
                                into FLOGOSTUDENT values (2,'John',20,3)
                                into FLOGOSTUDENT values (2,'John',2,3)
								into FLOGOSTUDENT values (3,'Chris',100,1)
								SELECT * FROM dual`

	db.Exec(dropFLOGOSTUDENTTable)
	db.Exec(createFLOGOSTUDENTTable)
	db.Exec(insertIntoFLOGOSTUDENT)
}

func shutdown() {
	if db != nil {
		db.Close()
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestUpdateSQL(t *testing.T) {
	query := `UPDATE FLOGOSTUDENT SET STDNAME = ?STDNAME, MARKS = 12 WHERE ROLLNUM = ?ROLLNUM;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "STDNAME": "NewValue",
            "ROLLNUM": 1
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestUpdateSQLRepeatedParameters(t *testing.T) {
	query := `UPDATE FLOGOSTUDENT SET STDNAME = ?NAME WHERE ROLLNUM = ?input AND MARKS != ?input;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "NAME": "NewValue1",
            "input": 2
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestUpdateWithSelectSQL(t *testing.T) {
	query := `UPDATE FLOGOSTUDENT SET STDNAME = ?NAME WHERE ROLLNUM = (select max(ROLLNUM) from FLOGOSTUDENT);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "NAME": "NewValue2"
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func PrepareAndExecute(t *testing.T, query string, inputParams map[string]interface{}) {
	queryActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "OracleDatabase-update"), activityName: "update"}
	//set logging to debug level
	log.SetLogLevel(queryActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(queryActivity.Metadata())

	//connection
	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(oracleConnectionJSON), &conn)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	odb := &oracledb.OracleDatabaseFactory{}
	connManager, err := odb.NewManager(conn)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	aInput := &Input{Connection: connManager, Query: query, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := queryActivity.Eval(tc)
	assert.True(t, ok)
	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	}
}
