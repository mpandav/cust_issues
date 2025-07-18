package rediscommand

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata
var connector *Connection

var testconnection = `{
	"id" : "RedisTestConnection",
	"name": "tibco-redis",
	"description" : "Redis Test Connection",
	"title": "Redis Connector",
	"type": "flogo:connector",
	"version": "1.0.0",
	"ref": "git.tibco.com/git/product/ipaas/wi-redis.git/Redis/connector/connection",
	"keyfield": "name",
	"settings": [
		{
		  "name": "name",
		  "value": "MyConnection",
		  "type": "string"
		},
		{
		  "name": "description",
		  "value": "My Redis Connection",
		  "type": "string"
		},
		{
		  "name": "host",
		  "value": "localhost",
		  "type": "string"
		},
		{
		  "name": "port",
		  "value": 6379,
		  "type": "int"
		},
		{
		  "name": "databaseIndex",
		  "value": 0,
		  "type": "int"
		  
		},  
		{
		  "name": "password",
		  "value": "password",
		  "type": "string"
		  
		}
	  ]
}`

func getConnector(t *testing.T, jsonConnection string) (map[string]interface{}, error) {

	connector := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonConnection), &connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return nil, err
	}
	return connector, nil
}

func getConnection(t *testing.T, jsonConnection string) (connection *Connection, err error) {
	connector, err := getConnector(t, jsonConnection)
	assert.NotNil(t, connector)

	connection, err = GetConnection(connector)
	if err != nil {
		t.Errorf("redis get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	return
}

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

func TestCreate(t *testing.T) {

	act := NewActivity(getActivityMetadata())

	if act == nil {
		t.Error("Activity Not Created")
		t.Fail()
		return
	}
}
func TestGetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	connectionBytes, err := ioutil.ReadFile("connectionFull.json")
	if err != nil {
		panic("connectionFull.json file found")
	}
	var connection interface{}
	err = json.Unmarshal(connectionBytes, &connection)
	if err != nil {
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}
	cmap := connection.(map[string]interface{})
	cname := cmap["connectorName"]

	log.Debug("connection name is %s", cname)
	connector := cmap["connector"]
	settings := connector.(map[string]interface{})["settings"]
	cmap["settings"] = settings
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())

	//setup attrs

	tc.SetInput(connectionProp, connection)
	tc.SetInput(ivCommand, "GET")

	var inputParams interface{}
	var inputJSON = []byte(`{
				"key":"testkey1",
				"DatabaseIndex": 2
				
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		assert.NotNil(t, complexOutput)
		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)

		jsonString := string(dataBytes)
		fmt.Println(jsonString)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SET")

	var inputParams interface{}
	var inputJSON = []byte(`{
				"key": "name",
				"value": "vinayak",
				"PX": 100,
				"NX|XX": "XX"
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		assert.NotNil(t, complexOutput)
		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)

		jsonString := string(dataBytes)
		fmt.Println(jsonString)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestMGetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "MGET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
		"keys":[{
			"key": "testKey1"
		}]			
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		assert.NotNil(t, complexOutput)
		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)

		jsonString := string(dataBytes)
		fmt.Println(jsonString)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestMSetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "MSET")

	var inputParams interface{}
	var inputJSON = []byte(`{
				"keyvalues": [{
					"key": "testKey",
				 	"value": "testVal"
				},{
					"key": "testKey1",
				 	"value": "testVal1"
				}]
			}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestLPopCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LPOP")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"names"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestRPopCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "RPOP")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"names"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestLPushCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LPUSH")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"names",
		"values" :[{"value": "akshay1"}, {"value": "vinayak1"}]
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestRPushCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "RPUSH")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"names",
		"values" :[{"value": "akshay1"}, {"value": "vinayak1"}]
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestLIndexCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LINDEX")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"names",
		"index" : 1
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestLSetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LSET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 		 
		"key":"names",
		"index" : 1,
		"value":"vinyak bagal"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestLRemCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LREM")

	var inputParams interface{}
	var inputJSON = []byte(`{ 		 
		"key":"names",
		"count" : 1,
		"value":"vinyak"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestLInsertCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LINSERT")

	var inputParams interface{}
	var inputJSON = []byte(`{ 		 
		"key":"names",
		"BEFORE|AFTER" :"BEFORE",
		"pivot":"akshay2",
		"value":"akshay1"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSAddCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SADD")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
		"key":"members",
		"members" :[{"member": "akshay1"}, {"member": "vinayak1"}]
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSRemCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SREM")

	var inputParams interface{}
	var inputJSON = []byte(`{ 		 
		"key":"members",
		"members" :[{"member": "akshay1"}, {"member": "vinayak1"}]
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSPopCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SPOP")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"members",
		"count": 2
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestScardCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SCARD")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"members"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSMembersCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SMEMBERS")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key":"members"
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestLRangeCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "LRANGE")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
			 
		"key": "TestBlog",
		"start": 0,
		"stop": 2
	
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}

}

func TestGetSetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "GETSET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "name",
			"value": "Tom"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestStrlenCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "STRLEN")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "name"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestGetRangeCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "GETRANGE")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "name",
			"start": 0,
			"end": 1
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestSetRangeCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "SETRANGE")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "name",
			"offset": 2,
			"value": "mJerry"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHSetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HSET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust",
			"field": "Name",
			"value": "Akshay"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHGetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HGET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust",
			"field": "Name"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHDelCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HDEL")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust",
			"fields": [{"field": "Name"}]
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHExistsCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HEXISTS")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust",
			"field": "Name"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHGetAllCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HGETALL")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHKeysCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HKEYS")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHLenCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HLEN")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHMGetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HMGET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust",
			"fields": [{"field": "Id"}, {"field": "Name"}, {"field": "Address"}, {"field": "City"}]
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHMSetCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HMSET")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust",
			"fieldvalues":[
				{
					"field": "Name",
					"value": "Vinayak"
				},
				{
					"field": "Id",
					"value": "11"
				},
				{
					"field": "Address",
					"value": "Yerwada"
				},
				{
					"field": "City",
					"value": "Pune"
				}
			]		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestHValsCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "HVALS")

	var inputParams interface{}
	var inputJSON = []byte(`{ 
				 
			"key": "Cust"
		
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestZcardCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZCARD")

	var inputParams interface{}
	var inputJSON = []byte(`{

		"key":"myset"

}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestZRemCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZREM")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset",
		"members":[{"member": "ten"}, {"member": "twenty"}]

}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestZrangeCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZRANGE")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset",
		"start" : 0,
		"stop":1,
		"WITHSCORES": true

}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestZcountCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZCOUNT")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset",
		"min" : "(0",
		"max":"1"

}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestZRankCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZRANK")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset",
		"member" : "two"
	

}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}

func TestZREMRANGEBYRANKCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZREMRANGEBYRANK")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset",
		"start" : 0,
		"stop":1


}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestZREMRANGEBYSCORECommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZREMRANGEBYSCORE")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset2",
		"min" : "20",
		"max":"100"


}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
func TestZAddCommand(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(getActivityMetadata())
	connector, err := getConnector(t, testconnection)

	assert.Nil(t, err)
	//setup attrs

	tc.SetInput(connectionProp, connector)
	tc.SetInput(ivCommand, "ZADD")

	var inputParams interface{}
	var inputJSON = []byte(`{
		"key":"myset2",
		"memberswithscores": [
			{
				"member": "one",
				"score" : 10
			}
		],		
		"NX|XX": "NX",
		"CH": true
}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput(inputProp, complex)

	ok, err := act.Eval(tc)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute command")
		t.Fail()
	} else {
		complexOutput := tc.GetOutput(outputProperty)
		log.Infof("OutputComplex %s", complexOutput)
		assert.True(t, ok)
		assert.Nil(t, err)
	}
}
