package update

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

	// conn := mysql.Connection{
	// 	DatabaseURL: "",
	// 	Host:        "wasp-deva.na.tibco.com",
	// 	Port:        3306,
	// 	User:        "root",
	// 	Password:    "admin",
	// 	DbName:      "university",
	// }
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

func TestUpdate(t *testing.T) {
	var prepareInsert = `update petone set birth = '2011-03-08', death = '2019-09-22' 	where NAME = 'Ceilidth';`
	var prepareQuery = `select * from petone where Owner = 'Shroeder';`

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	var fields []interface{}
	inputParams := &mysql.Input{}
	resultInsert, err := conn.PreparedUpdate(prepareInsert, inputParams, fields, logger)
	assert.Nil(t, err)

	resultDelete, err := conn.PreparedQuery(prepareQuery, inputParams, logger)
	assert.Nil(t, err)

	insertJSON, err := json.Marshal(resultInsert)
	deleteJSON, err := json.Marshal(resultDelete)
	fmt.Printf("\nResult  for update:  %s \nResult for select: %s", string(insertJSON), string(deleteJSON))

	conn.Logout(logger)
	assert.Nil(t, err)
}

func TestMixedCaseSql(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	update := `update petone set NAME = ?NAME, Owner = ?Owner, sPecies = ?sPecies , SeX = ?SeX where sPecies = 'Manx';`
	var inputJSON = []byte(`{
		"parameters": {
			"NAME": "Satan",
			"Owner": "Mystopheles",
			"sPecies": "eurokyte",
			"SeX": "z"
		},
		"values": [{
			"NAME": "Satan",
			"Owner": "Mystopheles",
			"sPecies": "eurokyte",
			"SeX": "z"
		}]
	}`)
	var inputFields []interface{}
	inputParams := &mysql.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	assert.Nil(t, err)
	result, err := conn.PreparedUpdate(update, inputParams, inputFields, logger)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Errr %s\n", err)
	}
	assert.NotNil(t, result)
	resultJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("%s\n", string(resultJSON))

	query := `select * from petone where sPecies = 'eurokyte';`
	var queryJSON = []byte(`{
		"parameters": {
		}
	}`)

	queryParams := &mysql.Input{}
	err = json.Unmarshal(queryJSON, queryParams)
	assert.Nil(t, err)
	resultQuery, err := conn.PreparedQuery(query, queryParams, logger)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Errr %s\n", err)
	}
	assert.NotNil(t, resultQuery)
	resultJSON, err = json.Marshal(resultQuery)
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
	query := `update petone set NAME = ?NAME, Owner = ?Owner, sPecies = ?sPecies , SeX = ?SeX where sPecies = 'eurokyte';`
	connector, err := getConnector(t)
	assert.Nil(t, err)

	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	var inputFields []interface{}
	tc.SetInput(FieldsProperty, inputFields)

	var inputParams interface{}
	var inputJSON = []byte(`{
			"parameters": {
				"NAME": "George",
				"Owner": "The Mad Hatter",
				"sPecies": "Monkey",
				"SeX": "F"
			}
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams.(map[string]interface{}))

	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not execute prepared query %s for reason: %s", query, err)
		t.Fail()
	}

	outputData := tc.GetOutput(OutputProperty)
	assert.NotNil(t, outputData)
	dataBytes, err := json.Marshal(outputData)
	assert.Nil(t, err)
	assert.NotNil(t, dataBytes)
	fmt.Printf("%s\n", string(dataBytes))

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
            "Type": "IMAGE",
            "Selected": false,
            "Parameter": true,
            "isEditable": false,
            "Value": true
		}
    ]
`)

func TestUpdateBlob(t *testing.T) {
	log.SetLogLevel(logger, log.DebugLevel)

	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	var fields interface{}
	err = json.Unmarshal(blobFields, &fields)
	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}
	update := `update testblob set fname=?fname,lname=?lname,photo=?photo where id = 34;`
	var inputJSON = []byte(`{
		"parameters": {
			"fname": "cerino",
			"lname" : "debergirac"
		}
	}`)

	inputParams := &mysql.Input{}
	err = json.Unmarshal(inputJSON, inputParams)
	photoBytes, err := ioutil.ReadFile("data/billthecat.jpeg")
	photoString := base64.StdEncoding.EncodeToString(photoBytes)
	inputParams.Parameters["photo"] = photoString
	assert.Nil(t, err)
	result, err := conn.PreparedUpdate(update, inputParams, fields, logger)
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
*/
