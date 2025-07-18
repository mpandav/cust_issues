package query

/*
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"git.tibco.com/git/product/ipaas/wi-contrib.git/connection/generic"
	"git.tibco.com/git/product/ipaas/wi-mysql.git/src/app/mysql/connector/mysql"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

const (
	ConnectionProp    = "Connection"
	DatabaseURL       = "datbaseURL"
	Host              = "Host"
	Port              = "Port"
	User              = "username"
	Password          = "Password"
	DatabaseName      = "DatabaseName"
	InputProp         = "input"
	ActivityOutput    = "output"
	QueryProperty     = "Query"
	QueryNameProperty = "QueryName"
	FieldsProperty    = "Fields"
	OutputProperty    = "Output"
	RecordsProperty   = "records"
)

var connector *mysql.Connection

var connectionJSON = []byte(`{
	 "id" : "MySQLTestConnection",
	 "name": "tibco-mysql",
	 "description" : "MySQL Test Connection",
	 "title": "MySQL Connector",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "ref": "https://git.tibco.com/git/product/ipaas/wi-mysql.git/activity/query",
	 "keyfield": "name",
	 "settings": [
		 {
		   "name": "name",
		   "value": "MyConnection",
		   "type": "string"
		 },
		 {
		   "name": "description",
		   "value": "MySQL Connection",
		   "type": "string"
		 },
		 {
		   "name": "host",
		   "value": "localhost",
		   "type": "string"
		 },
		 {
		   "name": "port",
		   "value": 3306,
		   "type": "int"
		 },
		 {
		   "name": "databaseName",
		   "value": "university",
		   "type": "string"

		 },
		 {
		   "name": "user",
		   "value": "widev",
		   "type": "string"

		 },
		 {
		   "name": "password",
		   "value": "widev",
		   "type": "string"

		 }
	   ]
 }`)
var connectionGilJSON = []byte(`{
	"id" : "MySQLTestConnection",
	"name": "tibco-mysql",
	"description" : "MySQL Test Connection",
	"title": "MySQL Connector",
	"type": "flogo:connector",
	"version": "1.0.0",
	"ref": "https://git.tibco.com/git/product/ipaas/wi-mysql.git/activity/query",
	"keyfield": "name",
	"settings": [
		{
		  "name": "name",
		  "value": "MyConnection",
		  "type": "string"
		},
		{
		  "name": "description",
		  "value": "MySQL Connection",
		  "type": "string"
		},
		{
		  "name": "host",
		  "value": "gil",
		  "type": "string"
		},
		{
		  "name": "port",
		  "value": 3306,
		  "type": "int"
		},
		{
		  "name": "databaseName",
		  "value": "widevtls",
		  "type": "string"

		},
		{
		  "name": "user",
		  "value": "widevtls",
		  "type": "string"

		},
		{
		  "name": "password",
		  "value": "widevtls",
		  "type": "string"

		}
	  ]
}`)

var invalidConnectionJSON = []byte(`{
	 "id" : "MySQLTestConnection",
	 "name": "tibco-MySQL",
	 "description" : "MySQL Test Connection",
	 "title": "AWS MySQL Connector",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "ref": "https://git.tibco.com/git/product/ipaas/wi-MySQL.git/activity/query",
	 "keyfield": "name",
	 "settings": [
		 {
		   "name": "name",
		   "value": "MyConnection",
		   "type": "string"
		 },
		 {
		   "name": "description",
		   "value": "My MySQL Connection",
		   "type": "string"
		 },
		 {
		   "name": "host",
		   "value": "flex-linux-gazelle.na.tibco.com",
		   "type": "string"
		 },
		 {
		   "name": "port",
		   "value": 3306,
		   "type": "int"
		 },
		 {
		   "name": "databaseName",
		   "value": "university",
		   "type": "string"

		 },
		 {
		   "name": "user",
		   "value": "root",
		   "type": "string"

		 },
		 {
		   "name": "password",
		   "value": "wrongpassword",
		   "type": "string"

		 }
	   ]
 }`)

var flexConnectionJSON = []byte(`{
  "id" : "MySQLTestConnection",
  "name": "tibco-mysql",
  "description" : "MySQL Test Connection",
  "title": "MySQL Connector",
  "type": "flogo:connector",
  "version": "1.0.0",
  "ref": "https://git.tibco.com/git/product/ipaas/wi-mysql.git/activity/query",
  "keyfield": "name",
  "settings": [
    {
      "name": "name",
      "value": "MyConnection",
      "type": "string"
    },
    {
      "name": "description",
      "value": "MySQL Connection",
      "type": "string"
    },
    {
      "name": "host",
      "value": "flex-linux-gazelle.na.tibco.com",
      "type": "string"
    },
    {
      "name": "port",
      "value": 3306,
      "type": "int"
    },
    {
      "name": "databaseName",
      "value": "northwind",
      "type": "string"

    },
    {
      "name": "user",
      "value": "widev",
      "type": "string"

    },
    {
      "name": "password",
      "value": "widev",
      "type": "string"

    }
    ]
}`)

var activityMetadata *activity.Metadata

// getTLSConnector yaya
func getTLSConnector(t *testing.T) (connector map[string]interface{}, err error) {
	connectorBytes, err := ioutil.ReadFile("/local/home/wcn00/go/src/git.tibco.com/git/product/ipaas/wi-mysql.git/src/tests/connectionTLS.json")
	if err != nil {
		t.Errorf("Failed to read tls connector config:  %s", err)
	}
	connector = make(map[string]interface{})
	err = json.Unmarshal(connectorBytes, &connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	return
}

// GetActivityMetadata is used only with tests
func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		var activityMap map[string]interface{}
		var activityMetadata *ActInput
		err = json.Unmarshal(jsonMetadataBytes, activityMap)
		activityMetadata.FromMap(activityMap)

	}
	return activityMetadata
}
func getConnector(t *testing.T) (connector map[string]interface{}, err error) {

	connector = make(map[string]interface{})
	err = json.Unmarshal([]byte(connectionJSON), &connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	return
}

const settingsConfig string = `{
	"connection": {
		"ref":"git.tibco.com/git/product/ipaas/wi-mysql.git/src/app/MySQL/connector/mysql/connection",
		"settings":{
			"brokerUrls":"localhost:9092"
		}
	}
	}`

func (mysql *SharedConfigManager) getConnection(t *testing.T) (connection *mysql.Connection, err error) {
	connector, err := getConnector(t)
	assert.NotNil(t, connector)

	actinputs := &ActInput{}

	connector.NewManager(s["connection"])

	sharedmanager := actinputs.Connection.(*mysql.SharedConfigManager)

	return
}
func TestQueries(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	var tests = []struct {
		id       string
		query    string
		input    string
		expected string
		errorMsg string
	}{
		{`MultipleSameParams`,
			`select count(*) from emr_patient WHERE practice_id=?abc and emr_practice_id=?abcxyx and practice_id=?abc;`,
			`{"parameters":{"abcxyx": "1959031", "abc": 105701}}`,
			`{"records":[{"count(*)":6}]}`,
			``,
		},
		// {`SubstringParams`,
		// 	`select count(*) from emr_patient WHERE practice_id=?abc and emr_practice_id=?abcxyx;`,
		// 	`{"parameters":{"abcxyx": "1959031", "abc": 105701}}`,
		// 	`{"records":[{"count(*)":6}]}`,
		// 	``,
		// },
		// {`WIMYSQ-144`,
		// 	`select count(*) from emr_patient WHERE practice_id=?z_emr_practice_id and emr_practice_id=?emr_practice_id;`,
		// 	`{"parameters":{"emr_practice_id": "1959031", "z_emr_practice_id": 105701}}`,
		// 	`{"records":[{"count(*)":6}]}`,
		// 	``,
		// },
		// {`SimpleQuery`,
		// 	`select count(*) from emr_patient WHERE practice_id=105701 and emr_practice_id="1959031";`,
		// 	`{"parameters":{}}`,
		// 	`{"records":[{"count(*)":6}]}`,
		// 	``,
		// },
	}

	conn, err := getConnector(t)
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

func testQuery(t *testing.T, id string, query string, input string, expected string, errorMsg string, conn interface{}) {

	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	var fields []interface{}
	tc.SetInput(ConnectionProp, conn)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, id)
	tc.SetInput(FieldsProperty, fields)

	var inputParams interface{}
	err = json.Unmarshal([]byte(input), &inputParams)
	tc.SetInput("input", inputParams.(map[string]interface{}))

	_, err = act.Eval(tc)
	if err != nil {
		if err.Error() == errorMsg {
			return
		}
		t.Errorf("%s", err.Error())
		return
	}

	output := tc.GetOutput(OutputProperty)
	assert.NotNil(t, output)
	assert.Nil(t, err)
	fmt.Printf("%v\n", output)
	dataBytes, err := json.Marshal(output)
	if err != nil {
		t.Errorf("invalid response format")
		return
	}
	if expected != string(dataBytes) {
		t.Errorf("query response has wrong value, got:  %s -- expected: %s", string(dataBytes), expected)
		return
	}
}

func TestGetConnection(t *testing.T) {

	connector := &mysql.Connector{}
	err := json.Unmarshal([]byte(connectionJSON), connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	connectionObj, err := generic.NewConnection(connector)
	if err != nil {
		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
		t.Fail()
	}
	connection, err := mysql.GetConnection(connectionObj)
	if err != nil {
		t.Errorf("MySQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	_, err = connection.Login(logger)
	if err != nil {
		t.Errorf("MySQL Login failed %s", err.Error())
	}
	connection.Logout(logger)
}

func TestGetTLSConnection(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	connectionObj, err := getTLSConnector(t)
	flogoConnection, err := generic.NewConnection(connectionObj["connector"])
	if err != nil {
		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
		t.Fail()
	}
	MysqlConnection, err := mysql.GetConnection(flogoConnection)
	if err != nil {
		t.Errorf("MySQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, MysqlConnection)
	_, err = MysqlConnection.Login(logger)
	if err != nil {
		t.Errorf("MySQL Login failed %s", err.Error())
	}
	MysqlConnection.Logout(logger)
}
func TestInvalidGetConnection(t *testing.T) {

	connector := &mysql.Connector{}
	err := json.Unmarshal([]byte(invalidConnectionJSON), connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	connectionObj, err := generic.NewConnection(connector)
	if err != nil {
		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
		t.Fail()
	}
	connection, err := mysql.GetConnection(connectionObj)
	if err != nil {
		t.Errorf("MySQL get connection failed %s", err.Error())
		t.Fail()
	}

	assert.NotNil(t, connection)
	_, err = connection.Login(logger)

	if err != nil {
		fmt.Printf("MySQL Login failed %s as expected \n", err.Error())
	}
	assert.Error(t, err)
	connection.Logout(logger)
}

func TestQueryString(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	var testQuery = `select name as instructor_name, course_id from instructor, teaches where instructor.ID = teaches.ID limit 10;;`

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Failed to login: %s\n", err)
	}

	inputParams := &mysql.Input{}
	result, err := conn.PreparedQuery(testQuery, inputParams, logger)
	assert.Nil(t, err)

	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s\n", string(resultJSON))

	conn.Logout(logger)
	assert.Nil(t, err)
}

var inputMap = map[string]interface{}{
	"state":       "CA",
	"likesports":  true,
	"liketheatre": true,
	"likejazz":    true,
}

func TestPreparedSQL(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	query := `select distinct T.name from instructor as T,
	instructor as S where T.salary > S.salary
	and S.dept_name = ?dept limit 10;`

	var inputJSON = []byte(`{
		"parameters": {
			"dept": "Biology"
		}
	}`)

	inputParams := &mysql.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedQuery(query, inputParams, logger)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s\n", string(resultJSON))
	conn.Logout(logger)
}

func TestPreparedSQL_IN(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	query := `SELECT * from course where cast(credits AS CHAR) in(?C1,?C2,?C3);`

	var inputJSON = []byte(`{
		"parameters": {
			"C1": "3",
			"C2": "4",
			"C3": "5"
		}
	}`)

	inputParams := &mysql.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedQuery(query, inputParams, logger)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s\n", string(resultJSON))
	conn.Logout(logger)
}

func TestPreparedSQL_Like(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	query := `SELECT * from course where cast(credits AS CHAR) like(%1%);`

	var inputJSON = []byte(`{
		"parameters": {

		}
	}`)

	inputParams := &mysql.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedQuery(query, inputParams, logger)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s\n", string(resultJSON))
	conn.Logout(logger)
}
func TestPreparedSQLEval(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	query := `select id, name from instructor where dept_name = ?dept_name limit 10;`
	connector, err := getConnector(t)
	assert.Nil(t, err)
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"parameters": {
			"dept_name": "Physics"
		}
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams.(map[string]interface{}))

	_, err = act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	} else {
		output := tc.GetOutput(OutputProperty)
		assert.NotNil(t, output)
		fmt.Printf("\nQuery executed: %v\n", output)
	}
}

func TestPerformanceEval(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	query := `select id, name from instructor where dept_name = ?dept_name limit 100;`
	connector, err := getConnector(t)
	assert.Nil(t, err)
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"parameters": {
			"dept_name": "Biology"
		}
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams.(map[string]interface{}))

	start := time.Now()
	for i := 0; i < 20000; i++ {
		iterStart := time.Now()
		_, err = act.Eval(tc)
		assert.Nil(t, err)

		if err != nil {
			t.Errorf("Could not execute prepared query %s", query)
			t.Fail()
		}
		iterEnd := time.Now()
		logger.Infof("Iteration %d ran in %s \n", i, iterEnd.Sub(iterStart))
	}
	end := time.Now()
	logger.Infof("Test performance ran 20000 queries in: %s \n", end.Sub(start).String())
}

func TestPreparedSQLNoArgs(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	var query = "select * from instructor limit 1;"

	inputParams := &mysql.Input{}

	var inputJSON = []byte(`{
		 "parameters": {}
	 }`)
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedQuery(query, inputParams, logger)
	assert.Nil(t, err)

	assert.Nil(t, err)
	assert.NotNil(t, result)

	conn.Logout(logger)

}

func TestFlexEmployeesEval(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())

	query := `select * from instructor where name=?first_name;`
	connector, err := getConnector(t)
	assert.Nil(t, err)
	var inputFields []interface{}
	var inputParams interface{}
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")
	tc.SetInput(FieldsProperty, inputFields)

	var inputJSON = []byte(`{
		"parameters": {
      "first_name":"Wu"
		}
	}`)
	err = json.Unmarshal(inputJSON, &inputParams)
	tc.SetInput("input", inputParams.(map[string]interface{}))
	_, err = act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(OutputProperty)
		assert.NotNil(t, complexOutput)
		fmt.Printf("\nQuery executed: %v\n", complexOutput)
		dataBytes, err := json.Marshal(complexOutput)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
		fmt.Printf("%s\n", string(dataBytes))

	}
}

/*
  * the common table expressions (with syntax) is not
  * supported in mysql till version 8.
  * Our test server is 5.7 so we'll leave this test commented out
  * for the time being

func TestPreparedSQLComplexEval(t *testing.T) {
	logger.SetLogLevel(logger.DebugLevel)

	act := NewActivity(GetActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	var query = `with dept_total (dept_name, value) as (select dept_name, sum(salary)
	from instructor group by dept_name), dept_total_avg(value) as (select
		avg(value) from dept_total) select dept_name from dept_total,
		dept_total_avg where dept_total.value >= dept_total_avg.value;
	`
	connector, err := getConnector(t)
	assert.Nil(t, err)

	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")

	var inputParams interface{}

	var inputJSON = []byte(`{}`)
	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

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
*/
/*
func TestNumericDatatypesEval(t *testing.T) {
	query := `select * from student limit 1;`
	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	connector, err := getConnector(t)
	assert.Nil(t, err)
	var inputFields []interface{}
	var inputParams interface{}

	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, "MyQuery")
	tc.SetInput(FieldsProperty, inputFields)
	var inputJSON = []byte(`{
	"parameters": {
		}
	}`)
	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams.(map[string]interface{}))

	_, err = act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(OutputProperty)
		assert.NotNil(t, complexOutput)
		fmt.Printf("\nQuery executed: %v\n", complexOutput)
		dataBytes, err := json.Marshal(complexOutput)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
		fmt.Printf("%s\n", string(dataBytes))
	}
}
*/
