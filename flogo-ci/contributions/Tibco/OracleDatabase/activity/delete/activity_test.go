package delete

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
                                into FLOGOSTUDENT values (4,'Mary',4,2)
                                into FLOGOSTUDENT values (4,'Mary',5,2)
                                into FLOGOSTUDENT values (5,'Kate',6,2)
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

func TestDeleteSQL(t *testing.T) {
	query := `DELETE FROM FLOGOSTUDENT WHERE ROLLNUM = ?ROLLNUM;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "ROLLNUM": 1
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestDeleteOnMultipleConditionsSQL(t *testing.T) {
	query := `DELETE FROM FLOGOSTUDENT WHERE STDNAME = ?STDNAME AND MARKS > ?MARKS;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "STDNAME": "John",
            "MARKS": 10
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestDeleteWithSelectSQL(t *testing.T) {
	query := `DELETE FROM FLOGOSTUDENT WHERE ROLLNUM = (select max(ROLLNUM) from FLOGOSTUDENT);`

	inputParams := make(map[string]interface{})
	inputParams = nil

	PrepareAndExecute(t, query, inputParams)
}

func TestDeleteSQLRepeatedParameters(t *testing.T) {
	query := `DELETE FROM FLOGOSTUDENT WHERE ROLLNUM = ?input AND MARKS != ?input;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "input": 4
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func PrepareAndExecute(t *testing.T, query string, inputParams map[string]interface{}) {
	queryActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "OracleDatabase-delete"), activityName: "delete"}
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
