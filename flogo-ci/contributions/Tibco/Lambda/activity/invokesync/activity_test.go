package invokesync

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	_ "github.com/tibco/flogo-aws/src/app/AWS/connector"
)

func getConnectionManager() interface{} {
	connectionBytes, err := ioutil.ReadFile("connectionData.json")
	if err != nil {
		panic("connectionData.json file found")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "connector", "github.com/tibco/flogo-aws/src/app/AWS/connector")
	connmgr, _ := coerce.ToConnection(connectionObj)
	return connmgr
}

func setupActivity(t *testing.T) (*Activity, *test.TestActivityContext) {
	act := &Activity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("ConnectionName", getConnectionManager())
	return act, tc
}

func logOutput(t *testing.T, tc *test.TestActivityContext) {
	output := &Output{}
	tc.GetOutputObject(output)
	assert.NotNil(t, output)
	outputBytes, err := json.Marshal(output)
	assert.Nil(t, err)
	tc.Logger().Info(string(outputBytes))
}

func TestLambdaInvoke(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("arn", "arn:aws:lambda:us-west-2:159020444217:function:Lambda2")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"msg": "Hello from Go Test"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("payload", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to invoke lambda function due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
