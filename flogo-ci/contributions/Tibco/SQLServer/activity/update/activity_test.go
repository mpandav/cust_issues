// /*
//  * Copyright Â© 2017. TIBCO Software Inc.
//  * This file is subject to the license terms contained
//  * in the license file that is distributed with this file.
//  */
package update

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

// func TestUpdate(t *testing.T) {
// 	var prepareInsert = `update PETS set birth = '2011-03-08', death = '2019-09-22' 	where NAME = 'ceiledth';`
// 	var prepareQuery = `select * from PETS where Owner = 'Wendell';`

// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)

// 	inputParams := &sqlserver.Input{}
// 	resultInsert, err := conn.PreparedUpdate(prepareInsert, inputParams, nil, log)
// 	assert.Nil(t, err)

// 	resultDelete, err := conn.PreparedQuery(prepareQuery, inputParams, log)
// 	assert.Nil(t, err)

// 	insertJSON, err := json.Marshal(resultInsert)
// 	deleteJSON, err := json.Marshal(resultDelete)
// 	fmt.Printf("\nResult  for update:  %s \nResult for select: %s", string(insertJSON), string(deleteJSON))

// 	conn.Logout(log)
// 	assert.Nil(t, err)
// }

// func TestMixedCaseSql(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)

// 	conn, err := getConnection(t)
// 	assert.Nil(t, err)
// 	_, err = conn.Login(log)
// 	assert.Nil(t, err)
// 	update := `update PETS set NAME = ?NAME, Owner = ?Owner, sPecies = ?sPecies , SeX = ?SeX where sPecies = 'Angel';`
// 	var inputJSON = []byte(`{
// 		"parameters": {
// 			"NAME": "Satan",
// 			"Owner": "Mystopheles",
// 			"sPecies": "eurokyte",
// 			"SeX": "z"
// 		}
// 	}`)

// 	inputParams := &sqlserver.Input{}
// 	err = json.Unmarshal(inputJSON, inputParams)
// 	assert.Nil(t, err)
// 	result, err := conn.PreparedUpdate(update, inputParams, nil, log)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		fmt.Printf("Errr %s\n", err)
// 	}
// 	assert.NotNil(t, result)
// 	resultJSON, err := json.Marshal(result)
// 	assert.Nil(t, err)
// 	fmt.Printf("%s\n", string(resultJSON))

// 	query := `select * from PETS where sPecies = 'eurokyte';`
// 	var queryJSON = []byte(`{
// 		"parameters": {
// 		}
// 	}`)

// 	queryParams := &sqlserver.Input{}
// 	err = json.Unmarshal(queryJSON, queryParams)
// 	assert.Nil(t, err)
// 	resultQuery, err := conn.PreparedQuery(query, queryParams, log)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		fmt.Printf("Errr %s\n", err)
// 	}
// 	assert.NotNil(t, resultQuery)
// 	resultJSON, err = json.Marshal(resultQuery)
// 	assert.Nil(t, err)
// 	fmt.Printf("%s\n", string(resultJSON))

// 	conn.Logout(log)
// }

// func TestEval(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)
// 	act := NewActivity(getActivityMetadata())
// 	tc := test.NewTestActivityContext(act.Metadata())

// 	query := `update PETS set NAME = ?NAME, Owner = ?Owner, sPecies = ?sPecies , SeX = ?SeX where sPecies = 'eurokyte';`

// 	connector, err := getConnector(t)
// 	assert.Nil(t, err)

// 	tc.SetInput(connectionProp, connector)
// 	tc.SetInput(queryProperty, query)

// 	var inputParams interface{}

// 	var inputJSON = []byte(`{
// 			"parameters": {
// 				"NAME": "George",
// 				"Owner": "The Mad Hatter",
// 				"sPecies": "Monkey",
// 				"SeX": "F"
// 			}
// 		}`)

// 	err = json.Unmarshal(inputJSON, &inputParams)
// 	assert.Nil(t, err)
// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	assert.Nil(t, err)
// 	if err != nil {
// 		t.Errorf("Could not execute prepared query %s", query)
// 		t.Fail()
// 	}

// 	complexOutput := tc.GetOutput(outputProperty)
// 	assert.NotNil(t, complexOutput)
// 	outputData := complexOutput.(*data.ComplexObject).Value
// 	dataBytes, err := json.Marshal(outputData)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, dataBytes)
// 	fmt.Printf("%s\n", string(dataBytes))

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

// func TestUpdateBlob(t *testing.T) {
// 	log.SetLogLevel(logger.InfoLevel)

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
// 	update := `update testblob set fname=?fname,lname=?lname,photo=?photo;`
// 	var inputJSON = []byte(`{
// 		"parameters": {
// 			"fname": "cerino",
// 			"lname" : "debergirac"
// 		}
// 	}`)

// 	inputParams := &sqlserver.Input{}
// 	err = json.Unmarshal(inputJSON, inputParams)
// 	photoBytes, err := ioutil.ReadFile("../../../../tests/data/billthecat.jpeg")
// 	photoString := base64.StdEncoding.EncodeToString(photoBytes)
// 	inputParams.Parameters["photo"] = photoString
// 	assert.Nil(t, err)
// 	result, err := conn.PreparedUpdate(update, inputParams, fields, log)
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
