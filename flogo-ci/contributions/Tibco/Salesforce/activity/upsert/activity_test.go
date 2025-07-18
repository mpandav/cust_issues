package upsert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

func TestUpsertAccount(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile("../../../../tests/connectionData.json")
	if err != nil {
		panic("connectionData.json file not found.")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "salesforce", "github.com/tibco/wi-salesforce/src/app/Salesforce/connector")

	connmgr, err := coerce.ToConnection(connectionObj)

	if err != nil {
		panic(err)
	}
	act := &UpsertActivity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connectionName", connmgr)
	tc.SetInput("objectName", "Account")
	tc.SetInput("externalIdFieldName", "MyExternalID__c")
	tc.SetInput("allOrNone", false)

	var inputData interface{}
	var inputJSON = []byte(`{
		"Account": {
			"records": [
				{
					"name": "SampleAccount1",
					"phone": "1111111111",
					"website": "www.salesforce1.com",
					"numberOfEmployees": 100,
					"industry": "Banking",
					"MyExternalID__c": "tyu"
				},
				{
					"name": "jj",
					"phone": "66",
					"MyExternalID__c": "ffffjkmmmcdd"
				}
			]
		}
	}`)
	err = json.Unmarshal(inputJSON, &inputData)
	assert.Nil(t, err)
	tc.SetInput("input", inputData)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not upsert account")
		t.Fail()
	} else {
		output := tc.GetOutput("output")
		assert.NotNil(t, output)
		dataBytes, err := json.Marshal(output)
		jsonString := string(dataBytes)
		t.Logf("%s", jsonString)
		fmt.Println(jsonString)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}
func TestAllOrNoneAsFalse(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile("../../../../tests/connectionData.json")
	if err != nil {
		panic("connectionData.json file not found.")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "salesforce", "github.com/tibco/wi-salesforce/src/app/Salesforce/connector")

	connmgr, err := coerce.ToConnection(connectionObj)
	if err != nil {
		panic(err)
	}
	act := &UpsertActivity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connectionName", connmgr)
	tc.SetInput("objectName", "Account")
	tc.SetInput("externalIdFieldName", "MyExternalID__c")
	tc.SetInput("allOrNone", false)

	var inputData interface{}
	var inputJSON = []byte(`{
		"Account": {
			"records": [
				{
					"name": "SampleAccount1",
					"phone": "1111111111",
					"website": "www.salesforce1.com",
					"numberOfEmployees": 100,
					"industry": "Banking",
					"MyExternalID__c": "jtyu"
				},
				{
					"name": "SampleAccount2",
					"phone": "2222222222",
					"website": "www.salesforce2.com",
					"numberOfEmployees": 250,
					"industry": "Banking",
					"MyExternalID__c": "ffffjkmmmcdd"
				},
				{
					"name": "jj",
					"phone": "66",
					"MyExternalID__c": "ffffjkmmmcdd"
				}
			]
		}
	}`)
	err = json.Unmarshal(inputJSON, &inputData)
	assert.Nil(t, err)
	tc.SetInput("input", inputData)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not upsert account")
		t.Fail()
	} else {
		output := tc.GetOutput("output")
		assert.NotNil(t, output)
		dataBytes, err := json.Marshal(output)
		jsonString := string(dataBytes)
		t.Logf("%s", jsonString)
		fmt.Println(jsonString)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}
func TestAllOrNoneAsTrue(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile("../../../../tests/connectionData.json")
	if err != nil {
		panic("connectionData.json file not found.")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "salesforce", "github.com/tibco/wi-salesforce/src/app/Salesforce/connector")

	connmgr, err := coerce.ToConnection(connectionObj)
	if err != nil {
		panic(err)
	}
	act := &UpsertActivity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connectionName", connmgr)
	tc.SetInput("objectName", "Account")
	tc.SetInput("externalIdFieldName", "MyExternalID__c")
	tc.SetInput("allOrNone", true)

	var inputData interface{}
	var inputJSON = []byte(`{
		"Account": {
			"records": [
				{
					"name": "SampleAccount1",
					"phone": "1111111111",
					"website": "www.salesforce1.com",
					"numberOfEmployees": 100,
					"industry": "Banking",
					"MyExternalID__c": "tyu"
				},
				{
					"name": "SampleAccount2",
					"phone": "2222222222",
					"website": "www.salesforce2.com",
					"numberOfEmployees": 250,
					"industry": "Banking",
					"MyExternalID__c": "ffffjkmmmcdd"
				},
				{
					"name": "jj",
					"phone": "66",
					"MyExternalID__c": "ffffjkmmmcdd"
				}
			]
		}
	}`)
	err = json.Unmarshal(inputJSON, &inputData)
	assert.Nil(t, err)
	tc.SetInput("input", inputData)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not upsert account")
		t.Fail()
	} else {
		output := tc.GetOutput("output")
		assert.NotNil(t, output)
		dataBytes, err := json.Marshal(output)
		jsonString := string(dataBytes)
		t.Logf("%s", jsonString)
		fmt.Println(jsonString)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}
func TestWithoutExternalID(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile("../../../../tests/connectionData.json")
	if err != nil {
		panic("connectionData.json file not found.")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "salesforce", "github.com/tibco/wi-salesforce/src/app/Salesforce/connector")

	connmgr, err := coerce.ToConnection(connectionObj)
	if err != nil {
		panic(err)
	}
	act := &UpsertActivity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connectionName", connmgr)
	tc.SetInput("objectName", "Account")
	tc.SetInput("externalIdFieldName", "MyExternalID__c")
	tc.SetInput("allOrNone", false)

	var inputData interface{}
	var inputJSON = []byte(`{
		"Account": {
			"records": [
				{
					"name": "SampleAccount1",
					"phone": "1111111111",
					"website": "www.salesforce1.com",
					"numberOfEmployees": 1002
				},
				{
					"name": "jj",
					"phone": "66",
					"MyExternalID__c": "ffffhhjkmmmcdd"
				}
			]
		}
	}`)
	err = json.Unmarshal(inputJSON, &inputData)
	assert.Nil(t, err)
	tc.SetInput("input", inputData)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not upsert account")
		t.Fail()
	} else {
		output := tc.GetOutput("output")
		assert.NotNil(t, output)
		dataBytes, err := json.Marshal(output)
		jsonString := string(dataBytes)
		t.Logf("%s", jsonString)
		fmt.Println(jsonString)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}
