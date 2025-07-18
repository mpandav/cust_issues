package get

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
	ExistingBucket      = "awakchau-s3-gotest"
	ExistingObject      = "hello"
	DestinationFilePath = "/home/abhijit"
	ListBucketPrefix    = "awakchau"
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

func TestGetBuckets(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "listBuckets")
	_, err := act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Get Buckets due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestGetBucketsWithPrefix(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "listBuckets")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Prefix": "` + ListBucketPrefix + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestGetObjects(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "listObjects")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + ExistingBucket + `"
	}`)
	json.Unmarshal(inputJSON, &inputParams)
	tc.SetInput("input", inputParams)
	_, err := act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf(err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestInvalidGetObjects(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "listObjects")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "xxx` + ExistingBucket + `xxx"
	}`)
	json.Unmarshal(inputJSON, &inputParams)
	tc.SetInput("input", inputParams)
	_, err := act.Eval(tc)
	if assert.NotNil(t, err) {
		tc.Logger().Info("Failed due to invalid bucket name: xxx" + ExistingBucket + "xxx")
	}
}

func TestGetSingleObjectAsText(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "single")
	tc.SetInput("outputType", "text")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + ExistingBucket + `",
		"Key": "` + ExistingObject + `"
	}`)
	json.Unmarshal(inputJSON, &inputParams)
	tc.SetInput("input", inputParams)
	_, err := act.Eval(tc)
	if !assert.Nil(t, err) {
		t.Errorf(err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestGetSingleObjectAsFile(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("operationType", "single")
	tc.SetInput("outputType", "file")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + ExistingBucket + `",
		"Key": "` + ExistingObject + `",
		"DestinationFilePath": "` + DestinationFilePath + `"
	}`)
	json.Unmarshal(inputJSON, &inputParams)
	tc.SetInput("input", inputParams)
	_, err := act.Eval(tc)
	if !assert.Nil(t, err) {
		t.Errorf(err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
