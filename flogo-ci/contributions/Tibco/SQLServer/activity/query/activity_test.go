// /*
//  * Copyright Â© 2017. TIBCO Software Inc.
//  * This file is subject to the license terms contained
//  * in the license file that is distributed with this file.
//  */
package query

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"testing"
// 	"time"

// 	// "git.tibco.com/git/product/ipaas/wi-mssql.git/src/app/SQLServer/connector/connection/sqlserver"
// 	sqlserver "git.tibco.com/git/product/ipaas/wi-mssql.git/src/app/SQLServer/connector/connection"

// 	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
// 	"github.com/TIBCOSoftware/flogo-lib/core/activity"
// 	"github.com/TIBCOSoftware/flogo-lib/core/data"
// 	"github.com/TIBCOSoftware/flogo-lib/logger"
// 	"github.com/stretchr/testify/assert"
// )

// var activityMetadata *activity.Metadata

// //var connector *map[string]interface{}

// var connectionJSON = `{
// 	 "id" : "SqlServerTestConnection",
// 	 "name": "tibco-sqlserver",
// 	 "description" : "SqlServer Test Connection",
// 	 "title": "SqlServer Connector",
// 	 "type": "flogo:connector",
// 	 "version": "1.0.0",
// 	 "ref": "https://git.tibco.com/git/product/ipaas/wi-mssql.git/activity/query",
// 	 "keyfield": "name",
// 	 "settings": [
// 		 {
// 		   "name": "name",
// 		   "value": "MyConnection",
// 		   "type": "string"
// 		 },
// 		 {
// 		   "name": "description",
// 		   "value": "SqlServer Connection",
// 		   "type": "string"
// 		 },
// 		 {
// 		   "name": "host",
// 		   "value": "flex-linux-gazelle",
// 		   "type": "string"
// 		 },
// 		 {
// 		   "name": "port",
// 		   "value": 1433,
// 		   "type": "int"
// 		 },
// 		 {
// 		   "name": "databaseName",
// 		   "value": "NORTHWND",
// 		   "type": "string"

// 		 },
// 		 {
// 		   "name": "user",
// 		   "value": "widev",
// 		   "type": "string"

// 		 },
// 		 {
// 		   "name": "password",
// 		   "value": "widev",
// 		   "type": "string"

// 		 }
// 	   ]
//  }`

// var invalidConnectionJSON = []byte(`{
// 	"id" : "MySQLTestConnection",
// 	"name": "tibco-sqlserver",
// 	"description" : "SqlServer Test Connection",
// 	"title": "SqlServer Connector",
// 	"type": "flogo:connector",
// 	"version": "1.0.0",
// 	"ref": "https://git.tibco.com/git/product/ipaas/wi-mssql.git/activity/query",
// 	"keyfield": "name",
// 	"settings": [
// 		{
// 			"name": "name",
// 			"value": "MyConnection",
// 			"type": "string"
// 		},
// 		{
// 			"name": "description",
// 			"value": "SqlServer Connection",
// 			"type": "string"
// 		},
// 		{
// 			"name": "host",
// 			"value": "flex-linux-gazelle",
// 			"type": "string"
// 		},
// 		{
// 			"name": "port",
// 			"value": 1433,
// 			"type": "int"
// 		},
// 		{
// 			"name": "databaseName",
// 			"value": "NORTHWND",
// 			"type": "string"

// 		},
// 		{
// 			"name": "user",
// 			"value": "dead",
// 			"type": "string"

// 		},
// 		{
// 			"name": "password",
// 			"value": "dead",
// 			"type": "string"

// 		}
// 		]
// }`)

// func getActivityMetadata() *activity.Metadata {
// 	if activityMetadata == nil {
// 		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
// 		if err != nil {
// 			panic("No Json Metadata found for activity.json path")
// 		}
// 		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
// 	}
// 	return activityMetadata
// }

// func getConnector(t *testing.T) (connector map[string]interface{}, err error) {
// 	connector = make(map[string]interface{})
// 	err = json.Unmarshal([]byte(connectionJSON), &connector)
// 	if err != nil {
// 		t.Errorf("Error: %s", err.Error())
// 	}
// 	return
// }
// func getInvalidConnector(t *testing.T) (connector map[string]interface{}, err error) {
// 	connector = make(map[string]interface{})
// 	err = json.Unmarshal([]byte(invalidConnectionJSON), &connector)
// 	if err != nil {
// 		t.Errorf("Error: %s", err.Error())
// 	}
// 	return
// }

// func getConnection(t *testing.T) (connection *sqlserver.Connection, err error) {
// 	connector, err := getConnector(t)
// 	assert.NotNil(t, connector)
// 	connection, err = sqlserver.GetConnection(&connector, log)
// 	if err != nil {
// 		t.Errorf("SqlServer get connection failed %s", err.Error())
// 		t.Fail()
// 	}
// 	assert.NotNil(t, connection)
// 	return
// }

// func TestGetConnection(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)
// 	connection, err := sqlserver.GetConnection(connector, log)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, connection)
// 	_, err = connection.Login(log)
// 	if err != nil {
// 		t.Errorf("SqlServer Login failed %s", err.Error())
// 	}
// 	connection.Logout(log)
// }

// func TestInvalidGetConnection(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	connector, err := getInvalidConnector(t)
// 	assert.Nil(t, err)
// 	connection, err := sqlserver.GetConnection(connector, log)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, connection)
// 	_, err = connection.Login(log)
// 	if err != nil {
// 		fmt.Printf("SqlServer Login failed %s as expected \n", err.Error())
// 	}
// 	assert.NotNil(t, err)
// 	err = connection.Logout(log)
// }

// func TestActivityRegistration(t *testing.T) {
// 	act := NewActivity(getActivityMetadata())
// 	if act == nil {
// 		t.Error("Activity Not Registered")
// 		t.Fail()
// 		return
// 	}
// }

// func TestQueryString(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	var testQuery = `select EmployeeID, FirstName, LastName from employees where LastName = 'Buchanan';`
// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)

// 	inputParams := &sqlserver.Input{}
// 	result, err := conn.PreparedQuery(testQuery, inputParams, log)
// 	assert.Nil(t, err)

// 	resultJSON, err := json.Marshal(result)
// 	assert.Nil(t, err)

// 	fmt.Printf("%s\n", string(resultJSON))

// 	conn.Logout(log)
// 	assert.Nil(t, err)
// }

// func TestPreparedSQL(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)

// 	query := `select employeeid,lastname,firstname,extension from employees where
// 	firstname = ?firstname;`
// 	var inputJSON = []byte(`{
// 		"parameters": {
// 			"firstname": "Nancy"
// 		}
// 	}`)

// 	inputParams := &sqlserver.Input{}
// 	err = json.Unmarshal(inputJSON, inputParams)
// 	assert.Nil(t, err)
// 	result, err := conn.PreparedQuery(query, inputParams, log)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, result)

// 	resultJSON, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	fmt.Printf("%s\n", string(resultJSON))
// 	conn.Logout(log)
// }

// // func TestPreparedSQLMultiWordCols(t *testing.T) {
// // 	log.SetLogLevel(logger.InfoLevel)
// // 	conn, err := getConnection(t)
// // 	assert.Nil(t, err)
// // 	_, err = conn.Login()
// // 	assert.Nil(t, err)

// // 	query := `select ID,"first name", "last name" from dbo.gotest where "first name" = ?first name;`
// // 	var inputJSON = []byte(`{
// // 		"parameters": {
// // 			"first name": "mary"
// // 		}
// // 	}`)

// // 	inputParams := &sqlserver.Input{}
// // 	err = json.Unmarshal(inputJSON, inputParams)
// // 	assert.Nil(t, err)
// // 	result, err := conn.PreparedQuery(query, inputParams)
// // 	assert.Nil(t, err)
// // 	assert.NotNil(t, result)

// // 	resultJSON, err := json.Marshal(result)
// // 	assert.Nil(t, err)
// // 	fmt.Printf("%s\n", string(resultJSON))
// // 	conn.Logout()
// // }

// func TestPreparedSQLTypes(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)

// 	query := `select * from wstypes;`
// 	var inputJSON = []byte(`{
// 		"parameters": {

// 		}
// 	}`)

// 	inputParams := &sqlserver.Input{}
// 	err = json.Unmarshal(inputJSON, inputParams)
// 	assert.Nil(t, err)
// 	result, err := conn.PreparedQuery(query, inputParams, log)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, result)

// 	resultJSON, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	fmt.Printf("%s\n", string(resultJSON))
// 	conn.Logout(log)
// }

// func TestPreparedSQLEval(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	act := NewActivity(getActivityMetadata())
// 	assert.NotNil(t, act)
// 	tc := test.NewTestActivityContext(act.Metadata())
// 	assert.NotNil(t, tc)
// 	query := `select * from employees where firstname=?firstname;`
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)

// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)
// 	tc.SetInput(queryNameProperty, "query")

// 	var inputParams interface{}
// 	var inputJSON = []byte(`{
// 		"parameters": {
// 			"firstname": "Nancy"
// 		}
// 	}`)
// 	err = json.Unmarshal(inputJSON, &inputParams)
// 	assert.Nil(t, err)

// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		t.Errorf("Could not execute prepared query %s", query)
// 		t.Fail()
// 	} else {
// 		complexOutput := tc.GetOutput(outputProperty)
// 		assert.NotNil(t, complexOutput)
// 		outputData := complexOutput.(*data.ComplexObject).Value
// 		dataBytes, err := json.Marshal(outputData)
// 		assert.Nil(t, err)
// 		assert.NotNil(t, dataBytes)

// 	}
// }

// func TestPerformanceEval(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	act := NewActivity(getActivityMetadata())
// 	assert.NotNil(t, act)
// 	tc := test.NewTestActivityContext(act.Metadata())
// 	assert.NotNil(t, tc)
// 	query := `select * from employees where firstname=?firstname;`
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)

// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)
// 	tc.SetInput(queryNameProperty, "MyQuery")

// 	var inputParams interface{}

// 	var inputJSON = []byte(`{
// 		"parameters": {
// 			"firstname": "Nancy"
// 		}
// 	}`)

// 	err = json.Unmarshal(inputJSON, &inputParams)
// 	assert.Nil(t, err)

// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	start := time.Now()
// 	for i := 0; i < 10; i++ {
// 		iterStart := time.Now()
// 		_, err = act.Eval(tc)
// 		assert.Nil(t, err)

// 		if err != nil {
// 			t.Errorf("Could not execute prepared query %s", query)
// 			t.Fail()
// 		}
// 		iterEnd := time.Now()
// 		log.Infof("Iteration %d ran in %s \n", i, iterEnd.Sub(iterStart))
// 	}
// 	end := time.Now()
// 	log.Infof("Test performance ran 30 queries in: %s \n", end.Sub(start).String())
// }

// func TestPreparedSQLNoArgs(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)

// 	inputParams := &sqlserver.Input{}
// 	var inputJSON = []byte(`{
// 		 "parameters": {}
// 	 }`)
// 	err = json.Unmarshal(inputJSON, inputParams)
// 	assert.Nil(t, err)

// 	var query = "SELECT  * FROM customers ORDER BY companyname OFFSET  1 ROWS FETCH NEXT 5 ROWS ONLY;"
// 	result, err := conn.PreparedQuery(query, inputParams, log)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, result)
// 	conn.Logout(log)

// }

// func TestPreparedSQLCTEEval(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)
// 	act := NewActivity(getActivityMetadata())
// 	assert.NotNil(t, act)
// 	tc := test.NewTestActivityContext(act.Metadata())
// 	assert.NotNil(t, tc)

// 	var query = `
// 		WITH Managers AS
// 		(
// 		--initialization
// 		SELECT EmployeeID, LastName, ReportsTo
// 		FROM Employees
// 		WHERE ReportsTo IS NULL
// 		UNION ALL
// 		--recursive execution
// 		SELECT e.employeeID,e.LastName, e.ReportsTo
// 		FROM Employees e INNER JOIN Managers m
// 		ON e.ReportsTo = m.employeeID
// 		)
// 		SELECT * FROM Managers;
// 	`
// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)
// 	tc.SetInput(queryNameProperty, "MyQuery")

// 	var inputParams interface{}

// 	var inputJSON = []byte(`{}`)
// 	err = json.Unmarshal(inputJSON, &inputParams)
// 	assert.Nil(t, err)

// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput(inputProp, complex)

// 	ok, err := act.Eval(tc)
// 	assert.Nil(t, err)

// 	if err != nil {
// 		t.Errorf("Could not execute prepared query %s", query)
// 		t.Fail()
// 	} else {
// 		complexOutput := tc.GetOutput(outputProperty)
// 		complexBytes, _ := json.Marshal(complexOutput)
// 		t.Logf("TestPreparedSQLCTEEval:  \n%s\n", string(complexBytes))
// 		assert.True(t, ok)
// 		assert.Nil(t, err)
// 	}

// }

// func TestBlobEval(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)
// 	act := NewActivity(getActivityMetadata())
// 	assert.NotNil(t, act)
// 	tc := test.NewTestActivityContext(act.Metadata())
// 	assert.NotNil(t, tc)
// 	query := `select id,fname,lname,photo from testblob where id = ?id;`
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)

// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)
// 	tc.SetInput(queryNameProperty, "query")

// 	var inputParams interface{}
// 	var inputJSON = []byte(`{
// 		"parameters": {
// 			"id": 11
// 		}
// 	}`)
// 	err = json.Unmarshal(inputJSON, &inputParams)
// 	assert.Nil(t, err)

// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		t.Errorf("Could not execute prepared query %s", query)
// 		t.Fail()
// 	} else {
// 		complexOutput := tc.GetOutput(outputProperty)
// 		assert.NotNil(t, complexOutput)
// 		outputData := complexOutput.(*data.ComplexObject).Value
// 		dataBytes, err := json.Marshal(outputData)
// 		assert.Nil(t, err)
// 		assert.NotNil(t, dataBytes)
// 		fmt.Printf("%s\n", string(dataBytes))
// 	}
// }
