/*
 * Copyright Â© 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

package query
/***
import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"git.tibco.com/git/product/ipaas/wi-postgres.git/src/app/PostgreSQL/connector/connection/connection"
	"git.tibco.com/git/product/ipaas/wi-postgres.git/src/app/github.com/TIBCOSoftware/flogo-lib/core/activity"
	"git.tibco.com/git/product/ipaas/wi-postgres.git/src/app/github.com/TIBCOSoftware/flogo-lib/core/data"
	"git.tibco.com/git/product/ipaas/wi-postgres.git/src/app/github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/project-flogo/core/support/log"

	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata

var connectionJSON = `{
	 "id" : "PostgresTestConnection",
	 "name": "tibco-postgres",
	 "description" : "PostgreSQL Test Connection",
	 "title": "PostgreSQL Connector",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "ref": "https://git.tibco.com/git/product/ipaas/wi-postgres.git/activity/query",
	 "keyfield": "name",
	 "settings": [
		 {
		   "name": "name",
		   "value": "MyConnection",
		   "type": "string"
		 },
		 {
		   "name": "description",
		   "value": "PostgreSQL Connection",
		   "type": "string"
		 },
		 {
		   "name": "host",
		   "value": "hyperion.na.tibco.com",
		   "type": "string"
		 },
		 {
		   "name": "port",
		   "value": 5432,
		   "type": "int"
		 },
		 {
		   "name": "databaseName",
		   "value": "university",
		   "type": "string"

		 },
		 {
		   "name": "user",
		   "value": "admin",
		   "type": "string"

		 },
		 {
		   "name": "password",
		   "value": "admin",
		   "type": "string"

		 }
	   ]
 }`

var invalidConnectionJSON = []byte(`{
	 "id" : "PostgresTestConnection",
	 "name": "tibco-PostgreSQL",
	 "description" : "PostgreSQL Test Connection",
	 "title": "AWS PostgreSQL Connector",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "ref": "https://git.tibco.com/git/product/ipaas/wi-Postgres.git/activity/query",
	 "keyfield": "name",
	 "settings": [
		 {
		   "name": "name",
		   "value": "MyConnection",
		   "type": "string"
		 },
		 {
		   "name": "description",
		   "value": "My PostgreSQL Connection",
		   "type": "string"
		 },
		 {
		   "name": "host",
		   "value": "wasp-deva.na.tibco.com",
		   "type": "string"
		 },
		 {
		   "name": "port",
		   "value": 5432,
		   "type": "int"
		 },
		 {
		   "name": "databaseName",
		   "value": "wicluster",
		   "type": "string"

		 },
		 {
		   "name": "user",
		   "value": "doesnotexist",
		   "type": "string"

		 },
		 {
		   "name": "password",
		   "value": "nopassword",
		   "type": "string"

		 }
	   ]
 }`)

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}
	return activityMetadata
}

func getConnector(t *testing.T, valid bool) (connector map[string]interface{}, err error) {
	connector = make(map[string]interface{})
	if valid {
		err = json.Unmarshal([]byte(connectionJSON), &connector)
	} else {
		err = json.Unmarshal([]byte(invalidConnectionJSON), &connector)
	}
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	return
}

func getConnection(t *testing.T) (connection *postgres.Connection, err error) {
	connector, err := getConnector(t, true)
	assert.NotNil(t, connector)

	connection, err = postgres.GetConnection(connector)
	if err != nil {
		t.Errorf("PostgreSQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	return
}

func TestGetConnection(t *testing.T) {
	// log.SetLogLevel(logger.InfoLevel)
	connector, err := getConnector(t, true)
	assert.Nil(t, err)
	connection, err := postgres.GetConnection(connector)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("PostgreSQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	err = connection.Login()
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("PostgreSQL Login failed %s", err.Error())
	}
	connection.Logout()
}

func TestInvalidGetConnection(t *testing.T) {
	log.SetLogLevel(logger.InfoLevel)
	connector, err := getConnector(t, false)
	assert.Nil(t, err)
	connection, err := postgres.GetConnection(connector)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("PostgreSQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	err = connection.Login()
	if err != nil {
		fmt.Printf("PostgreSQL Login failed [%s] as expected \n", err.Error())
	}
	assert.Error(t, err)
	connection.Logout()
}

func TestActivityRegistration(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	if act == nil {
		t.Error("Activity Not Registered")
		t.Fail()
		return
	}
}

func TestLogin(t *testing.T) {
	conn := postgres.Connection{
		DatabaseURL: "",
		Host:        "wasp-deva.na.tibco.com",
		Port:        5432,
		User:        "admin",
		Password:    "admin",
		DbName:      "university",
	}
	err := conn.Login()
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}

func TestQueryString(t *testing.T) {

	//-- Find events in the 99.9 percentile in terms of all time gross sales.
	var testQuery = `select name as instructor_name, course_id from instructor, teaches where instructor.ID = teaches.ID limit 10;;`

	conn, err := getConnection(t)
	assert.Nil(t, err)
	err = conn.Login()

	inputParams := &postgres.Input{}
	result, err := conn.PreparedQuery(testQuery, inputParams)
	assert.Nil(t, err)

	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s", string(resultJSON))

	conn.Logout()
	assert.Nil(t, err)
}

var inputMap = map[string]interface{}{
	"state":       "CA",
	"likesports":  true,
	"liketheatre": true,
	"likejazz":    true,
}

func TestPreparedSQL(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	err = conn.Login()
	assert.Nil(t, err)
	query := `select distinct T.name from instructor as T,
	instructor as S where T.salary > S.salary
	and S.dept_name = ?dept limit 10;`

	var inputJSON = []byte(`{
		"parameters": {
			"dept": "Biology"
		}
	}`)

	inputParams := &postgres.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedQuery(query, inputParams)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	conn.Logout()
	//assert.Nil(t, err)
}

func TestPreparedSQLEval(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	query := `select id, name from instructor where dept_name = ?dept_name limit 10;`

	connector, err := getConnector(t, true)
	assert.Nil(t, err)

	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")

	var inputParams interface{}

	var inputJSON = []byte(`{
		"parameters": {
			"dept_name": "Statistics"
		}
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complex)

	_, err = act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(OutputProperty)
		assert.NotNil(t, complexOutput)
		fmt.Printf("\nQuery executed: %v\n", complexOutput)
		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)

		t.Logf("%s", complexOutput)
		assert.Nil(t, err)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}

func TestPreparedSQLNoArgs(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	err = conn.Login()
	assert.Nil(t, err)
	var query = "select * from instructor limit 1;"

	inputParams := &postgres.Input{}

	var inputJSON = []byte(`{
		 "parameters": {}
	 }`)
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedQuery(query, inputParams)
	assert.Nil(t, err)

	assert.Nil(t, err)
	assert.NotNil(t, result)

	conn.Logout()

}

func TestPreparedSQLComplexEval(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	var query = `with dept_total (dept_name, value) as (select dept_name, sum(salary)
	from instructor group by dept_name), dept_total_avg(value) as (select
		avg(value) from dept_total) select dept_name from dept_total,
		dept_total_avg where dept_total.value >= dept_total_avg.value;
	`
	connector, err := getConnector(t, true)
	assert.Nil(t, err)

	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")

	var inputParams interface{}

	var inputJSON = []byte(`{}`)
	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(InputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(OutputProperty)
		t.Logf("%s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSQLJoins(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	var tests = []struct {
		id       string
		query    string
		input    string
		expected string
		errorMsg string
	}{
		{`RightOuterJoin-WIPGRS-67`,
			`SELECT * FROM classroom RIGHT OUTER JOIN department ON (classroom.building = department.building);`,
			`{"parameters":{}}`,
			`{"records":[{"budget":80000,"building":"Packard","capacity":500,"dept_name":"Music","room_number":"101"},{"budget":50000,"building":"Painter","capacity":10,"dept_name":"History","room_number":"514"},{"budget":120000,"building":"Painter","capacity":10,"dept_name":"Finance","room_number":"514"},{"budget":85000,"building":"Taylor","capacity":70,"dept_name":"Elec. Eng.","room_number":"3128"},{"budget":100000,"building":"Taylor","capacity":70,"dept_name":"Comp. Sci.","room_number":"3128"},{"budget":70000,"building":"Watson","capacity":30,"dept_name":"Physics","room_number":"100"},{"budget":90000,"building":"Watson","capacity":30,"dept_name":"Biology","room_number":"100"},{"budget":70000,"building":"Watson","capacity":50,"dept_name":"Physics","room_number":"120"},{"budget":90000,"building":"Watson","capacity":50,"dept_name":"Biology","room_number":"120"},{"budget":123,"building":"test1","capacity":null,"dept_name":"Test","room_number":null}]}`,
			`no error`,
		},
	}

	conn, err := GetPostgresObject()
	if err != nil {
		t.Errorf("connection failed: %s", err.Error())
		t.FailNow()
		return
	}
	for _, test := range tests {
		t.Logf("Running test %s\n", test.id)
		testQuery(t, test.id, test.query, test.input, test.expected, test.errorMsg, conn)
	}
}
func TestBlobs(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	var tests = []struct {
		id        string
		query     string
		input     string
		fields    string
		errorMsg  string
		fieldName string
		fileName  string
	}{
		{id: `InsertPngBlob`,
			query:     `Select logo from flogo.connector where jiraid=?jiraid;`,
			input:     `{"parameters":{"jiraid":"WIPRGS"}}`,
			fields:    `[{"FieldName":"logo","Type":"BYTEA","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"product_id","Type":"INT4","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"version","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"jiraid","Type":"TEXT","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"description","Type":"VARCHAR","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"NUMERIC","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
			errorMsg:  `no error`,
			fieldName: "logo",
			fileName:  "../insert/icons/ic-postgres-insert@3x.png",
		},
	}
	conn, err := GetPostgresObject()
	if err != nil {
		t.Errorf("connection failed: %s", err.Error())
		t.FailNow()
		return
	}
	for _, test := range tests {
		t.Logf("Running test %s\n", test.id)
		testBlob(t, test.id, test.query, test.input, test.fields, test.fieldName, test.fileName, test.errorMsg, conn)
	}

}

func testBlob(t *testing.T, id string, query string, input string, fields string,
	fieldName string, fileName string, errorMsg string, conn interface{}) {

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	connector := conn.(map[string]interface{})["connector"]
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, id)
	tc.SetInput(FieldsProperty, fields)

	inputParams := postgres.Input{}
	err := json.Unmarshal([]byte(input), &inputParams)

	photoBytes, err := ioutil.ReadFile(fileName)
	photoString := base64.StdEncoding.EncodeToString(photoBytes)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complex)

	_, err = act.Eval(tc)
	if err != nil {
		if err.Error() == errorMsg {
			return
		}
		t.Errorf("%s", err.Error())
		return
	}
	complexOutput := tc.GetOutput(OutputProperty)
	outputData := complexOutput.(*data.ComplexObject).Value

	record, ok := outputData.(*postgres.ResultSet)
	if !ok {
		t.Errorf("return value not a result set")
		return
	}

	entry := record.Record[0]
	retrievedString := (*entry)["logo"]
	if photoString != retrievedString {
		t.Errorf("query response has wrong value, got:  %s -- expected: %s", retrievedString, photoString)
		return
	}
}

// GetPostgresObject for testing
func GetPostgresObject() (connector interface{}, err error) {
	cb, err := ioutil.ReadFile("data/connectionData.json")
	if err != nil {
		return connector, err
	}
	err = json.Unmarshal(cb, &connector)
	if err != nil {
		return connector, err
	}
	return
}

func testQuery(t *testing.T, id string, query string, input string, expected string, errorMsg string, conn interface{}) {

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	connector := conn.(map[string]interface{})["connector"]
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, id)

	var inputParams interface{}
	err := json.Unmarshal([]byte(input), &inputParams)
	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complex)

	_, err = act.Eval(tc)
	if err != nil {
		if err.Error() == errorMsg {
			return
		}
		t.Errorf("%s", err.Error())
		return
	}
	complexOutput := tc.GetOutput(OutputProperty)
	outputData := complexOutput.(*data.ComplexObject).Value
	dataBytes, err := json.Marshal(outputData)
	if err != nil {
		t.Errorf("invalid response format")
		return
	}
	value := string(dataBytes)
	if expected != value {
		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
		return
	}
}
***/