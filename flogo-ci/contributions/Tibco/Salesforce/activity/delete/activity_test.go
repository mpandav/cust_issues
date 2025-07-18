package delete

import (
	"encoding/json"
	"fmt"
	"testing"

	"io/ioutil"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

func TestDeleteAccount(t *testing.T) {
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
	act := &DeleteActivity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("Connection Name", connmgr)
	tc.SetInput("Object Name", "Account")

	var inputData interface{}
	var inputJSON = []byte(`{
		"data": [
			{
				"id": "0012w00001mGIAaAAO"
			},
			{
				"id": "0012w00001mGIXKAA4"
			}
		]
	}`)
	err = json.Unmarshal(inputJSON, &inputData)
	assert.Nil(t, err)
	tc.SetInput("input", inputData)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not delete account")
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
