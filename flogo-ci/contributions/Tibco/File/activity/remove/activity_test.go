package remove

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	ref := activity.GetRef(&MyActivity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

// Path C:/Flogo/new/new1/a.txt is expected to be present for this test case
func TestRemoveFileOperation(t *testing.T) {
	removeRecursive := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/new/new1/a.txt"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, removeRecursive)
}

func TestRemoveDirOperation(t *testing.T) {
	removeRecursive := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/new/new1"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, removeRecursive)
}

// Path C:/Flogo/new is non-empty directory
// this should fail with error as we are trying to delete non-empty directory with removeRecursive false
func TestRemoveNonEmptyDirOperation(t *testing.T) {
	removeRecursive := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/new"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, removeRecursive)
}

func TestRemoveNonEmptyDirOperation1(t *testing.T) {
	removeRecursive := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/new"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, removeRecursive)
}

func Execute(t *testing.T, inputParams map[string]interface{}, removeRecursive bool) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-remove"), activityName: "remove"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{Input: inputParams, RemoveRecursive: removeRecursive}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform remove operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of remove operation: %s", err.Error())
		t.Fail()
	}
}
