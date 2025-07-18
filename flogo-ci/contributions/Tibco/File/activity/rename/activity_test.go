package rename

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

func TestRenameFileOperation(t *testing.T) {
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fromFileName": "C:/Flogo/a.txt",
		 "toFileName": "C:/Flogo/b.txt",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir)
}

// this test will fail as b.txt exists from above test case and overwrite is set to false
func TestRenameFileOperation1(t *testing.T) {
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fromFileName": "C:/Flogo/a1.txt",
		 "toFileName": "C:/Flogo/b.txt",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir)
}

func TestRenameFileOperation2(t *testing.T) {
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fromFileName": "C:/Flogo/a1.txt",
		 "toFileName": "C:/Flogo/b.txt",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir)
}

// folder new does not exist
func TestRenameFileOperation3(t *testing.T) {
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fromFileName": "C:/Flogo/b.txt",
		 "toFileName": "C:/Flogo/new/c.txt",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir)
}

func TestRenameFolderOperation(t *testing.T) {
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fromFileName": "C:/Flogo/new1",
		 "toFileName": "C:/Flogo/new2",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir)
}

func Execute(t *testing.T, inputParams map[string]interface{}, createNonExistingDir bool) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-rename"), activityName: "rename"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{CreateNonExistingDir: createNonExistingDir, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform rename operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of rename operation: %s", err.Error())
		t.Fail()
	}
}
