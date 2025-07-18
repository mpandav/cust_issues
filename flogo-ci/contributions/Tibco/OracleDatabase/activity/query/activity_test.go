package query

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
								into FLOGOSTUDENT values (3,'Chris',100,1)
								SELECT * FROM dual`

	db.Exec(dropFLOGOSTUDENTTable)
	db.Exec(createFLOGOSTUDENTTable)
	db.Exec(insertIntoFLOGOSTUDENT)

	dropFLOGODATATYPESTable := `drop table FLOGO_DATATYPES`
	createFLOGODATATYPESTable := `create table FLOGO_DATATYPES 
                                (
                                    char1 CHAR, char2 CHAR(4), char3 CHAR(3 CHAR),
									var1 VARCHAR(30), var2 VARCHAR2(20 CHAR), nvar1 NCHAR(10), nvar2 NVARCHAR2(20), l1 LONG,
									num1 NUMBER, num2 NUMBER(6), num3 NUMBER(8,2), f1 FLOAT(5), bf BINARY_FLOAT, bd BINARY_DOUBLE,
									d1 DATE, t1 TIMESTAMP(3), t2 TIMESTAMP(8), t3 TIMESTAMP(8) WITH TIME ZONE,
									t4 TIMESTAMP(3) WITH LOCAL TIME ZONE,
									y1 INTERVAL YEAR TO MONTH, y2 INTERVAL YEAR TO MONTH, y3 INTERVAL YEAR TO MONTH, y4 INTERVAL YEAR(3) TO MONTH,
									i1 INTERVAL DAY TO SECOND, i2 INTERVAL DAY TO SECOND, i3 INTERVAL DAY TO SECOND, i4 INTERVAL DAY TO SECOND, i5 INTERVAL DAY TO SECOND,
									i6 INTERVAL DAY(3) TO SECOND(8), i7 INTERVAL DAY(3) TO SECOND(8)
								)`

	insertIntoFLOGODATATYPES := `INSERT into FLOGO_DATATYPES values
								(
									'c','char','cha',
									'var1','var2','nvar1','nvar2','abc',
									1,123456,3.33,4.4444,5.55555,6.666666,
									'09-JAN-10','10-FEB-2019 9:26:50.123456789','10-JAN-2019 9:26:50.123456789 PM', '31-JAN-2011 9:57:24.01800012 PM -07:00',
									'31-JAN-2011 9:57:24.018 PM',
									INTERVAL '10-2' YEAR TO MONTH, INTERVAL '9' MONTH, '+01-03', INTERVAL '120-2' YEAR(3) TO MONTH,
									INTERVAL '11 10:09' DAY TO MINUTE, INTERVAL '99 10' DAY TO HOUR, INTERVAL '09:08:07.66' HOUR TO SECOND, INTERVAL '09:30' HOUR TO MINUTE, INTERVAL '5' DAY,
									INTERVAL '250' HOUR(3), INTERVAL '09:08:07.123456789' HOUR TO SECOND(8)
								)`

	db.Exec(dropFLOGODATATYPESTable)
	db.Exec(createFLOGODATATYPESTable)
	db.Exec(insertIntoFLOGODATATYPES)
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

func TestDatatypes(t *testing.T) {
	query := `select * from FLOGO_DATATYPES;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestSelectStar(t *testing.T) {
	query := `select * from FLOGOSTUDENT;`
	inputParams := make(map[string]interface{})
	inputParams = nil

	PrepareAndExecute(t, query, inputParams)
}

func TestSelectWhereClauseWithParam(t *testing.T) {
	query := `select * from FLOGOSTUDENT where ROLLNUM = ?ROLLNUM;`

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

func TestSelectMultipleWhereClauseWithParam(t *testing.T) {
	query := `select * from FLOGOSTUDENT where ROLLNUM != ?a AND STDNAME = ?b AND CLASS = ?a;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
             "a": 1,
             "b":"Chris"
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestMultiSelect(t *testing.T) {
	query := `select * from FLOGOSTUDENT where ROLLNUM = (select min(ROLLNUM) from FLOGOSTUDENT);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
         }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestMaxFunc(t *testing.T) {
	query := `select max(MARKS) as marks_max from FLOGOSTUDENT;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestSelectDistinct(t *testing.T) {
	query := `select distinct(ROLLNUM) from FLOGOSTUDENT;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestSelectOrderBy(t *testing.T) {
	query := `select * from FLOGOSTUDENT order by MARKS desc;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestSelectSumFuncAndGroupBy(t *testing.T) {
	query := `SELECT ROLLNUM, SUM(MARKS) AS TotalMarks FROM FLOGOSTUDENT GROUP BY ROLLNUM;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func PrepareAndExecute(t *testing.T, query string, inputParams map[string]interface{}) {
	queryActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "OracleDatabase-query"), activityName: "query"}
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
