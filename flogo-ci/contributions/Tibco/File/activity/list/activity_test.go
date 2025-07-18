package list

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

func TestListFilesOperation(t *testing.T) {
	mode := "Files and Directories"
	//mode := "Only Directories"
	//mode := "Only Files"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/abc"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, mode)
}

func TestListFilesOperation1(t *testing.T) {
	mode := "Files and Directories"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/abc/*"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, mode)
}

func TestListFilesOperation2(t *testing.T) {
	mode := "Only Directories"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/abc/*"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, mode)
}

func TestListFilesOperation3(t *testing.T) {
	mode := "Only Files"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/abc/*"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, mode)
}

func TestListFilesOperation4(t *testing.T) {
	mode := "Only Files"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/abc/*.txt"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, mode)
}

func Execute(t *testing.T, inputParams map[string]interface{}, mode string) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-list"), activityName: "list"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{Mode: mode, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform list operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of list operation: %s", err.Error())
		t.Fail()
	}
}
