// /*
//  * Copyright Â© 2017. TIBCO Software Inc.
//  * This file is subject to the license terms contained
//  * in the license file that is distributed with this file.
//  */
package insert

// import (
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"testing"

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
// 		   "value": "flex-linux-gazelle.na.tibco.com",
// 		   "type": "string"
// 		 },
// 		 {
// 		   "name": "port",
// 		   "value": 1433,
// 		   "type": "int"
// 		 },
// 		 {
// 		   "name": "databaseName",
// 		   "value": "university",
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
// 	logger.SetLogLevel(logger.InfoLevel)
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
// 	logger.SetLogLevel(logger.InfoLevel)
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
// 	logger.SetLogLevel(logger.DebugLevel)
// 	var testQuery = `INSERT INTO deletetest values ('901','richard','dawkins');`
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

// func TestPreparedSQLEval(t *testing.T) {
// 	logger.SetLogLevel(logger.DebugLevel)
// 	act := NewActivity(getActivityMetadata())
// 	assert.NotNil(t, act)
// 	tc := test.NewTestActivityContext(act.Metadata())
// 	assert.NotNil(t, tc)
// 	query := `INSERT INTO gotest values (?ID, ?fname, ?lname);`
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)

// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)
// 	tc.SetInput(queryNameProperty, "query")

// 	var inputParams interface{}
// 	var inputJSON = []byte(`{
// 		"values": [
// 		{
// 			"ID": "1001",
// 			"fname": "george",
// 			"lname": "ofthejungle"
// 		},
// 		{
// 			"ID": "1002",
// 			"fname": "george",
// 			"lname": "ofthejungle"
// 		}
// 	]}`)
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
// 		fmt.Printf("Output from insert: %v", outputData)
// 		dataBytes, err := json.Marshal(outputData)
// 		assert.Nil(t, err)
// 		assert.NotNil(t, dataBytes)
// 	}
// }
// func TestInsertHardCoded(t *testing.T) {
// 	logger.SetLogLevel(logger.DebugLevel)
// 	act := NewActivity(getActivityMetadata())
// 	assert.NotNil(t, act)
// 	tc := test.NewTestActivityContext(act.Metadata())
// 	assert.NotNil(t, tc)
// 	var query = `insert into gotest values (500,'ken','bard'),(600,'ken','bard'),(700,'ken','bard');`
// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)

// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)
// 	tc.SetInput(queryNameProperty, "query")

// 	var inputParams interface{}
// 	var inputJSON = []byte(`{
// 		"values": [
// 	]}`)
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
// 		fmt.Printf("Output from insert: %v", outputData)
// 		dataBytes, err := json.Marshal(outputData)
// 		assert.Nil(t, err)
// 		assert.NotNil(t, dataBytes)
// 	}
// }

// var blobFields = []byte(`
//     [
//         {
//             "FieldName": "id",
//             "Type": "INT",
//             "Selected": false,
//             "Parameter": true,
//             "isEditable": false,
//             "Value": true
//         },
//         {
//             "FieldName": "fname",
//             "Type": "VARCHAR",
//             "Selected": false,
//             "Parameter": true,
//             "isEditable": false,
//             "Value": true
//         },
//         {
//             "FieldName": "lname",
//             "Type": "VARCHAR",
//             "Selected": false,
//             "Parameter": true,
//             "isEditable": false,
//             "Value": true
//         },
//         {
//             "FieldName": "age",
//             "Type": "INT",
//             "Selected": false,
//             "Parameter": true,
//             "isEditable": false,
//             "Value": true
//         },
//         {
//             "FieldName": "photo",
//             "Type": "IMAGE",
//             "Selected": false,
//             "Parameter": true,
//             "isEditable": false,
//             "Value": true
// 		}
//     ]
// `)

// func TestInsertBlob(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)

// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)
// 	var fields interface{}
// 	err = json.Unmarshal(blobFields, &fields)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		fmt.Printf("Error: %s", err)
// 	}
// 	insert := `insert into testblob (id, fname,lname,photo ) values (?id,?fname,?lname,?photo);`
// 	var inputJSON = []byte(`{
// 		"values": [
// 		{
// 			"id": 99,
// 			"fname": "walter",
// 			"lname": "matheou"
// 		}
// 	]}`)

// 	inputParams := &sqlserver.Input{}
// 	err = json.Unmarshal(inputJSON, inputParams)
// 	photoBytes, err := ioutil.ReadFile("../../../../tests/data/billthecat.jpeg")
// 	photoString := base64.StdEncoding.EncodeToString(photoBytes)
// 	inputParams.Values[0]["photo"] = photoString
// 	assert.Nil(t, err)
// 	result, err := conn.PreparedInsert(insert, inputParams, fields, log)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		fmt.Printf("Errr %s\n", err)
// 	}
// 	assert.NotNil(t, result)
// 	resultJSON, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	fmt.Printf("%s\n", string(resultJSON))

// 	conn.Logout(log)
// }
