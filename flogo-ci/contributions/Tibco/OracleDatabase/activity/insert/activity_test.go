package insert

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

	dropFLOGODATATYPESTable := `drop table FLOGO_DATATYPES`
	createFLOGODATATYPESTable := `create table FLOGO_DATATYPES 
                                (
                                    char1 CHAR, char2 CHAR(4), char3 CHAR(3 CHAR),
                                    var1 VARCHAR(30), var2 VARCHAR2(20 CHAR), nvar1 NCHAR(10), nvar2 NVARCHAR2(20), l1 LONG,
                                    num1 NUMBER, num2 NUMBER(6), num3 NUMBER(8,2), f1 FLOAT(5), bf BINARY_FLOAT, bd BINARY_DOUBLE,
                                    d1 DATE, t1 TIMESTAMP(3), t2 TIMESTAMP(8), t3 TIMESTAMP(8) WITH TIME ZONE
                                )`

	/*createFLOGODATATYPESTable := `create table FLOGO_DATATYPES
	                                (
	                                    char1 CHAR, char2 CHAR(4), char3 CHAR(3 CHAR),
										var1 VARCHAR(30), var2 VARCHAR2(20 CHAR), nvar1 NCHAR(10), nvar2 NVARCHAR2(20), l1 LONG,
										num1 NUMBER, num2 NUMBER(6), num3 NUMBER(8,2), f1 FLOAT(5), bf BINARY_FLOAT, bd BINARY_DOUBLE,
										d1 DATE, t1 TIMESTAMP(3), t2 TIMESTAMP(8), t3 TIMESTAMP(8) WITH TIME ZONE,
										t4 TIMESTAMP(3) WITH LOCAL TIME ZONE,
										y1 INTERVAL YEAR TO MONTH, y2 INTERVAL YEAR TO MONTH, y3 INTERVAL YEAR TO MONTH, y4 INTERVAL YEAR(3) TO MONTH,
										i1 INTERVAL DAY TO SECOND, i2 INTERVAL DAY TO SECOND, i3 INTERVAL DAY TO SECOND, i4 INTERVAL DAY TO SECOND, i5 INTERVAL DAY TO SECOND,
										i6 INTERVAL DAY(3) TO SECOND(8), i7 INTERVAL DAY(3) TO SECOND(8)
									)`*/

	_, err = db.Exec(dropFLOGODATATYPESTable)
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec(createFLOGODATATYPESTable)
	if err != nil {
		fmt.Println(err)
	}

	dropFLOGOSTUDENTTable := `drop table FLOGOSTUDENT`
	createFLOGOSTUDENTTable := `create table FLOGOSTUDENT (ROLLNUM number, STDNAME varchar(20), MARKS number, CLASS number)`

	db.Exec(dropFLOGOSTUDENTTable)
	db.Exec(createFLOGOSTUDENTTable)

	dropSTUDENTtable := `drop table STUDENT`
	createSTUDENTtable := `create table STUDENT (ROLLNUM number, STDNAME varchar(20), MARKS number, CLASS number)`

	db.Exec(dropSTUDENTtable)
	db.Exec(createSTUDENTtable)

	insertIntoFLOGOSTUDENTtable := `INSERT ALL
							    into FLOGOSTUDENT values (1,'Tom',23,1)
								into FLOGOSTUDENT values (2,'Tom',25,1)
								into FLOGOSTUDENT values (3,'John',20,3)
								into FLOGOSTUDENT values (4,'Chris',100,1)
								into FLOGOSTUDENT values (5,'Student',100,1)
								SELECT * FROM dual`

	db.Exec(insertIntoFLOGOSTUDENTtable)
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

func TestInsertValuesAllDatatypes(t *testing.T) {
	query := `INSERT into FLOGO_DATATYPES values
    (
        ?char1, ?char2, ?char3,
        ?var1, ?var2, ?nvar1, ?nvar2, ?l1,
        ?num1, ?num2, ?num3, ?f1, ?bf, ?bd,
        ?d1, ?t1, ?t2, ?t3
    );`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "values": [{
             "char1" : "c",
             "char2" : "char",
             "char3" : "cha",
             "var1" : "var1",
             "var2" : "var2",
             "nvar1" : "nvar1",
             "nvar2" : "nvar2",
             "l1" : "abc",
             "num1" : 1,
             "num2" : 123456,
             "num3" : 3.33,
             "f1" : 4.4444,
             "bf" : 5.55555,
             "bd" : 6.666666,
             "d1" : "09-JAN-10",
             "t1" : "10-FEB-2019 9:26:50.123456789",
             "t2" : "10-JAN-2019 9:26:50.123456789 PM",
             "t3" : "31-JAN-2011 9:57:24.01800012 PM -07:00"
         },{
            "char1" : "c",
            "char2" : "char",
            "char3" : "cha",
            "var1" : "var1",
            "var2" : "var2",
            "nvar1" : "nvar1",
            "nvar2" : "nvar2",
            "l1" : "abc",
            "num1" : 1,
            "num2" : 123456,
            "num3" : 3.33,
            "f1" : 4.4444,
            "bf" : 5.55555,
            "bd" : 6.666666,
            "d1" : "2019-10-10+00:00",
            "t1" : "2019-10-11T16:34:04+00:00",
            "t2" : "10-JAN-2019 9:26:50.123456789 PM",
            "t3" : "31-JAN-2011 9:57:24.01800012 PM -07:00"
        }]
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertNoParams(t *testing.T) {
	query := `insert into FLOGOSTUDENT values (12, 'TestInsert', 12, 12);`

	inputParams := make(map[string]interface{})
	inputParams = nil

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertParams(t *testing.T) {
	query := `insert into FLOGOSTUDENT values (?a, ?b, ?c, ?d);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
             "a" : 1,
             "b" : "Tom",
             "c" : 50,
             "d" : 1 
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertValues(t *testing.T) {
	query := `insert into FLOGOSTUDENT Values (?ROLLNUM, ?STDNAME, ?MARKS, ?CLASS);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "values": [{
             "ROLLNUM" : 1,
             "STDNAME" : "Tom",
             "MARKS" : 50,
             "CLASS" : 1
         }, {
            "ROLLNUM" : 2,
            "STDNAME" : "John",
            "MARKS" : 51,
            "CLASS" : 2 
         }]
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertValuesAndParams(t *testing.T) {
	query := `insert into FLOGOSTUDENT values (?ROLLNUM, ?STDNAME, ?c, ?d);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "values": [{
             "ROLLNUM" : 3,
             "STDNAME" : "Kate"
         },{
            "ROLLNUM" : 4,
            "STDNAME" : "Jack"
         }],
         "parameters": {
            "c" : 52,
            "d" : 3
         }
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertIncorrectValuesAndParams(t *testing.T) {
	query := `insert into FLOGOSTUDENT values (?a, ?STDNAME, ?c, ?d);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "values": [{
             "ROLLNUM" : 4,
             "STDNAME" : "Mary"
         }],
         "parameters": {
            "a" : 10,
            "c" : 55,
            "d" : 10
         }
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertIncorrectParamsAndValues(t *testing.T) {
	query := `insert into FLOGOSTUDENT values (?ROLLNUM, ?STDNAME, ?c, ?d);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "values": [{
             "ROLLNUM" : 5,
             "STDNAME" : "Sansa"
         }],
         "parameters": {
            "a" : 11,
            "c" : 56,
            "d" : 5
         }
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertWithSelectNoParams(t *testing.T) {
	query := `insert into STUDENT select * from FLOGOSTUDENT;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertWithSelectAndParams(t *testing.T) {
	query := `insert into STUDENT select * from FLOGOSTUDENT where rollnum > ?rollnum;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
            "rollnum" : 11
         }
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertALL(t *testing.T) {
	query := `insert all
		into FLOGOSTUDENT values (1,'Tom',23,1)
		into FLOGOSTUDENT values (2,'Tom',25,1)
		select * from dual;	
	`
	inputParams := make(map[string]interface{})
	inputParams = nil

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertALLwithParams(t *testing.T) {
	query := `insert all
		into FLOGOSTUDENT values (1,'Tom',23,?p1)
		into FLOGOSTUDENT values (2,?p2,25,1)
		select * from dual;	
	`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"p1" : 11,
			"p2" : "JERRY"
         }
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func TestInsertALLMultipleTableswithParams(t *testing.T) {
	query := `insert all
	into flogostudent (ROLLNUM, STDNAME, MARKS, CLASS) values (?p1, 'TOM', 23, 1)
	into flogostudent (ROLLNUM, STDNAME, MARKS) values (2, 'JERRY', 21)
	into student values (32, ?p2, 40, 3)
	select * from dual;
	`
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"parameters": {
			"p1" : 11,
			"p2" : "TEST"
         }
     }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PrepareAndExecute(t, query, inputParams)
}

func PrepareAndExecute(t *testing.T, query string, inputParams map[string]interface{}) {
	queryActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "OracleDatabase-insert"), activityName: "insert"}
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
