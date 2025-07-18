package callprocedure

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
}

func TestRegister(t *testing.T) {
	ref := activity.GetRef(&MyActivity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
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

func TestCallProcedureWithcursor(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"IN","FieldName":"E_ID","Type":"NUMBER","isEditable":true},{"Direction":"OUT","FieldName":"EMP_NAME","Type":"VARCHAR","isEditable":true},{"Direction":"OUT","FieldName":"H_Date","Type":"TIMESTAMP","isEditable":true},{"Direction":"OUT","FieldName":"p_recordset1","Type":"REFCURSOR","isEditable":true}],"ok":true,"query":""}`
	query := `CALL singleCursor(?E_ID,?EMP_NAME,?H_Date,?p_recordset1);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"E_ID": 3
		}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func TestCallProcedureWithMultipleCursor(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"IN","FieldName":"E_ID","Type":"NUMBER","isEditable":true},{"Direction":"OUT","FieldName":"EMP_NAME","Type":"VARCHAR","isEditable":true},{"Direction":"OUT","FieldName":"p_recordset1","Type":"REFCURSOR","isEditable":true}, {"Direction":"OUT","FieldName":"p_recordset2","Type":"REFCURSOR","isEditable":true}],"ok":true,"query":""}`
	query := `CALL multipleCursors(?E_ID,?EMP_NAME,?p_recordset1,?p_recordset2);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"E_ID": 1
		}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func TestCallProcedureWithSimpleOutParam(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"IN","FieldName":"E_ID","Type":"NUMBER","isEditable":true},{"Direction":"OUT","FieldName":"EMP_NAME","Type":"VARCHAR","isEditable":true},{"Direction":"OUT","FieldName":"HIRE_DATE","Type":"TIMESTAMP","isEditable":true}],"ok":true,"query":""}`
	query := `CALL queryDBUSER(?E_ID,?EMP_NAME,?HIRE_DATE);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"E_ID": 1
		}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func TestCallProcedureWithNoOutParamInsert(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"IN","FieldName":"EMP_ID","Type":"NUMBER","isEditable":true},{"Direction":"IN","FieldName":"EMP_NAME","Type":"VARCHAR","isEditable":true},{"Direction":"IN","FieldName":"HIRE_DATE","Type":"TIMESTAMP","isEditable":true}],"ok":true,"query":""}`
	query := `CALL insertDBUSER(?EMP_ID,?EMP_NAME,?HIRE_DATE);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"EMP_ID": 11,
			"EMP_NAME": "Tomm",
			"HIRE_DATE": "12-JUL-22 05.16.05.551000 PM"
		}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func TestCallProcedureWithNoOutParamUpdate(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"IN","FieldName":"EMP_ID","Type":"NUMBER","isEditable":true},{"Direction":"IN","FieldName":"EMP_NAME","Type":"VARCHAR","isEditable":true}],"ok":true,"query":""}`
	query := `CALL updateDBUSER(?EMP_ID,?EMP_NAME);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"EMP_ID": 3,
			"EMP_NAME": "Bella"
			}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func TestCallProcedureWithNoOutParamDelete(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"IN","FieldName":"EMP_ID","Type":"NUMBER","isEditable":true}],"ok":true,"query":""}`
	query := `CALL deleteDBUSER(?EMP_ID);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"EMP_ID": 11
			}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func TestCallProcedureWithInOutParam(t *testing.T) {
	fieldsInfo := `{"fields":[{"Direction":"INOUT","FieldName":"P_VAL","Type":"VARCHAR","isEditable":true}],"ok":true,"query":""}`
	query := `CALL testinout(?P_VAL);`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"P_VAL": "hello world"
		}
	}`)
	err1 := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err1)
	PrepareAndExecute(t, query, inputParams, fieldsInfo)
}

func PrepareAndExecute(t *testing.T, query string, inputParams map[string]interface{}, fieldsInfo string) {
	callProcActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "OracleDatabase-callprocedure"), activityName: "callprocedure"}
	//set logging to debug level
	log.SetLogLevel(callProcActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(callProcActivity.Metadata())

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

	aInput := &Input{Connection: connManager, CallProcedure: query, Input: inputParams, FieldsInfo: fieldsInfo}
	tc.SetInputObject(aInput)
	fmt.Println("Input is: ", aInput)
	ok, err := callProcActivity.Eval(tc)
	assert.True(t, ok)
	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	fmt.Println("Output is: ", aOutput)
	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	}
}
