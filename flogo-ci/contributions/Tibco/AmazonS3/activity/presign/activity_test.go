package presign

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	_ "github.com/tibco/flogo-aws/src/app/AWS/connector"
)

var activityMetadata *activity.Metadata

const (
	ExistingBucket = "awakchau-s3-test-presign-url"
	ExistingObject = "hello"
)

func getConnectionManager() interface{} {
	connectionBytes, err := ioutil.ReadFile("../connectionData.json")
	if err != nil {
		panic("connectionData.json file found")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "connector", "git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/AWS/connector")
	connmgr, _ := coerce.ToConnection(connectionObj)
	return connmgr
}

func setupActivity(t *testing.T) (*Activity, *test.TestActivityContext) {
	act := &Activity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connection", getConnectionManager())
	return act, tc
}

func logOutput(t *testing.T, tc *test.TestActivityContext) {
	output := tc.GetOutput("output")
	assert.NotNil(t, output)
	outputBytes, err := json.Marshal(output)
	assert.Nil(t, err)
	tc.Logger().Info("output:", string(outputBytes))
}

func TestPresignGET(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "get")
	tc.SetInput("expirationTimeSec", 300)
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + ExistingBucket + `",
		"Key":"` + ExistingObject + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Get Buckets due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestPresignPUT(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "put")
	tc.SetInput("expirationTimeSec", 300)
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + ExistingBucket + `",
		"Key":"` + ExistingObject + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Get Buckets due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestPresignDELETE(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "delete")
	tc.SetInput("expirationTimeSec", 300)
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + ExistingBucket + `",
		"Key":"` + ExistingObject + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Get Buckets due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
