package put

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
	BucketName                 = "awakchau-s3-gotest"
	PutObjectAsTextKey         = "PutObject-withUpload-asText"
	PutObjectAsFileKey         = "PutObject-withUpload-asFile"
	PutObjectAsFileSourcePath  = "/home/abhijit/hello"
	PutObjectWithCopyObjectKey = "CopyOf-PutObject-withUpload-asFile"
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

func TestPutBucket(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Put Bucket due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestInvalidPutBucket(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.NotNil(t, err)
	if err != nil {
		tc.Logger().Info(err.Error())
	}
}

func TestPutObjectWithUploadText(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "object")
	tc.SetInput("putType", "upload")
	tc.SetInput("inputType", "text")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
		"Key": "` + PutObjectAsTextKey + `",
		"TextContent": "Hello with upload as text"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Put Object due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestPutObjectWithUploadFile(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "object")
	tc.SetInput("putType", "upload")
	tc.SetInput("inputType", "file")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
    "Key": "` + PutObjectAsFileKey + `",
    "SourceFilePath": "` + PutObjectAsFileSourcePath + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Put Object due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestPutObjectWithCopyObject(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "object")
	tc.SetInput("putType", "copy")
	tc.SetInput("preserveACL", true)

	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
    "CopySource": "` + BucketName + `/` + PutObjectAsFileKey + `",
		"Key": "` + PutObjectWithCopyObjectKey + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Put Object due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
