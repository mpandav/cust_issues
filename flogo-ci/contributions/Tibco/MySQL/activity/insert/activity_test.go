package insert

/*

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"git.tibco.com/git/product/ipaas/wi-contrib.git/connection/generic"
	"git.tibco.com/git/product/ipaas/wi-mysql.git/src/app/MySQL/connector/connection/mysql"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata
var connector *mysql.Connection

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

var flexConnectionJSONBad = []byte(`{
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
      "value": 33096,
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
      "value": "widev999",
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
		var activityMap map[string]interface{}
		var activityMetadata *ActInput
		err = json.Unmarshal(jsonMetadataBytes, activityMap)
		activityMetadata.FromMap(activityMap)

	}
	return activityMetadata
}

func getConnector(t *testing.T) (connector map[string]interface{}, err error) {

	connector = make(map[string]interface{})
	err = json.Unmarshal([]byte(flexConnectionJSON), &connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	return
}

func getConnection(t *testing.T) (connection *mysql.Connection, err error) {
	connector, err := getConnector(t)
	assert.NotNil(t, connector)
	connectionObj, err := generic.NewConnection(connector)
	if err != nil {
		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
		t.Fail()
	}

	connection, err = mysql.GetConnection(connectionObj)
	if err != nil {
		t.Errorf("MySQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	return
}

func TestGetConnection(t *testing.T) {
	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)

	_, err = conn.Login(logger)
	if err != nil {
		t.Errorf("MySQL Login failed %s", err.Error())
	}
	conn.Logout(logger)
}

func TestLogin(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	conn := mysql.Connection{
		DatabaseURL: "",
		Host:        "flex-linux-gazelle.na.tibco.com",
		Port:        "3306",
		User:        "widev",
		Password:    "widev",
		DbName:      "northwind",
	}
	status, err := conn.Login(logger)
	if status == false {
		fmt.Println(err)
		t.Fail()
	}
	conn.Logout(logger)
}

func TestInsert(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	var prepareInsert = `INSERT INTO wcntest VALUES ('1','mary','popins','109');`
	var deleteStatement = `DELETE FROM wcntest where id = '1';`

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	inputParams := &mysql.Input{}
	resultInsert, err := conn.PreparedInsert(prepareInsert, inputParams, nil, logger)
	assert.Nil(t, err)

	resultDelete, err := conn.PreparedDelete(deleteStatement, inputParams, logger)
	assert.Nil(t, err)

	insertJSON, err := json.Marshal(resultInsert)
	deleteJSON, err := json.Marshal(resultDelete)
	fmt.Printf("\nResult  for insert:  %s \nResult for delete: %s", string(insertJSON), string(deleteJSON))

	conn.Logout(logger)
	assert.Nil(t, err)
}

func TestMixedCaseSql(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	query := `INSERT INTO pet (NAME,Owner,sPecies,SeX) VALUES (?NAME1,'Bill the Cat','dog','M'),('walter', ?NAME2,'dog','M'),('snark','Bill the Cat',?NAME3,'M');`

	var inputJSON = []byte(`{
		"parameters": {
			"NAME2": "Satan",
			"NAME1": "Mystopheles",
			"NAME3": "Bealzebub"
		}
	}`)
	var inputFields []interface{}
	inputParams := &mysql.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedInsert(query, inputParams, inputFields, logger)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Errr %s\n", err)
	}
	assert.NotNil(t, result)
	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s\n", string(resultJSON))
	conn.Logout(logger)
}

func TestEval(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//query := `INSERT INTO PET (NAME,Owner,sPecies,SeX) VALUES (?NAME,?Owner,?sPecies,?SeX),(?NAME,?Owner,?sPecies,?SeX);`
	//query := `INSERT INTO PET (NAME,Owner,sPecies,SeX) VALUES (?NAME,?Owner,?sPecies,?SeX);`
	query := `INSERT INTO petone  VALUES (?NAME,?Owner,?sPecies,?SeX,birth,death);`
	connector, err := getConnector(t)
	assert.Nil(t, err)
	var inputFields []interface{}
	var inputParams interface{}
	var inputJSON = []byte(`{
		"values": [
			{
				"NAME": "Bill the Cat",
				"Owner": "Shroeder",
				"sPecies": "Manx",
				"SeX": "female",
				"birth":"1991-06-01",
				"death":"1999-06-01"
			},
			{
				"NAME": "CatInTheHat",
				"Owner": "Schindler",
				"sPecies": "Fluffy",
				"SeX": "male",
				"birth":"1991-06-01",
				"death":"1999-06-01"
				}
			]
	}`)
	err = json.Unmarshal(inputJSON, &inputParams)
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput("Input", inputParams.(map[string]interface{}))
	tc.SetInput(FieldsProperty, inputFields)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	}

	output := tc.GetOutput(OutputProperty)
	assert.NotNil(t, output)
	assert.Nil(t, err)
	fmt.Printf("%v\n", output)

}

func TestEvalParms(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)
	settings := &Settings{}
	iCtx := test.NewActivityInitContext(settings, nil)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	query := `insert into university.petone select * from university.pet where NAME = ?NAME;`
	connector, err := getConnector(t)
	assert.Nil(t, err)
	var inputFields []interface{}
	var inputParams interface{}
	var inputJSON = []byte(`{
		"parameters":{
			"NAME":"stolie"
		},
		"values": []
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput("Input", inputParams.(map[string]interface{}))
	tc.SetInput(FieldsProperty, inputFields)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	}
	output := tc.GetOutput(OutputProperty)
	assert.NotNil(t, output)
	assert.Nil(t, err)
	fmt.Printf("%v\n", output)
}

var blobFields = []byte(`
    [
        {
            "FieldName": "id",
            "Type": "INT",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": true
        },
        {
            "FieldName": "fname",
            "Type": "VARCHAR",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": true
        },
        {
            "FieldName": "lname",
            "Type": "VARCHAR",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": true
        },
        {
            "FieldName": "age",
            "Type": "INT",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": true
        },
        {
            "FieldName": "photo",
            "Type": "BLOB",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": true
		},
		{
            "FieldName": "photoz",
            "Type": "BLOB",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": false
        }

    ]
`)

func TestInsertBlob(t *testing.T) {
	var prepareInsert = `insert into testblob (id,fname,lname,age,photo) values (?id,?fname,?lname,?age,?photo);`
	var fields interface{}
	log.SetLogLevel(logger, log.DebugLevel)

	err := json.Unmarshal(blobFields, &fields)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)

	inputParams := &mysql.Input{}
	//	inputParams.Values = append(inputParams.Values, "name")

	var inputJSON = []byte(`{
		"Values": [
			{
			"id": 200,
			"fname": "Mystopheles",
			"lname": "Bealzebub",
			"age": 2019
			}
		]
	}`)
	err = json.Unmarshal(inputJSON, inputParams)
	photoBytes, err := ioutil.ReadFile("data/GrizlyEating.jpg")
	photoString := base64.StdEncoding.EncodeToString(photoBytes)
	inputParams.Values[0]["photo"] = photoString
	resultInsert, err := conn.PreparedInsert(prepareInsert, inputParams, fields, logger)
	assert.Nil(t, err)

	// resultDelete, err := conn.PreparedDelete(deleteStatement, inputParams, logger)
	// assert.Nil(t, err)

	insertJSON, err := json.Marshal(resultInsert)
	// deleteJSON, err := json.Marshal(resultDelete)
	fmt.Printf("\nResult  for insert:  %s ", string(insertJSON))
	conn.Logout(logger)
	assert.Nil(t, err)
}

func TestInsertBlobParm(t *testing.T) {
	var prepareInsert = `insert into testblob (id,fname,lname,age,photo) values (?id,?fname,?lname,?age,?photoz);`
	var fields interface{}
	log.SetLogLevel(logger, log.DebugLevel)

	err := json.Unmarshal(blobFields, &fields)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)

	inputParams := &mysql.Input{}
	//	inputParams.Values = append(inputParams.Values, "name")

	var inputJSON = []byte(`{
		"Values": [
			{
				"id": 11,
				"fname": "Mystopheles",
				"lname": "Bealzebub",
				"age": 2019
				}
		],
		"Parameters": {}
	}`)
	err = json.Unmarshal(inputJSON, inputParams)
	photoBytes, err := ioutil.ReadFile("/home/wcn00/Pictures/LeahAndWillowDayOne.jpg")
	photoString := base64.StdEncoding.EncodeToString(photoBytes)
	inputParams.Parameters["photoz"] = photoString
	resultInsert, err := conn.PreparedInsert(prepareInsert, inputParams, fields, logger)
	assert.Nil(t, err)

	// resultDelete, err := conn.PreparedDelete(deleteStatement, inputParams, logger)
	// assert.Nil(t, err)

	insertJSON, err := json.Marshal(resultInsert)
	// deleteJSON, err := json.Marshal(resultDelete)
	fmt.Printf("\nResult  for insert:  %s ", string(insertJSON))
	conn.Logout(logger)
	assert.Nil(t, err)
}
*/
