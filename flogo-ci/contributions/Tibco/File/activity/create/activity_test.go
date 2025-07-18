package create

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

func TestCreateFileOperationRelativePath(t *testing.T) {
	isDir := false
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "./abc/pqr/newFile.txt",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// Path C:/Flogo/Flogo1 is expected to be present
func TestCreateFileOperation(t *testing.T) {
	isDir := false
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo1/newFile.txt",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// this will pass since overwrite is true
func TestCreateFileOperation1(t *testing.T) {
	isDir := false
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo1/newFile.txt",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// expected to fail as above testcase will already create same file and in this
// testcase overwrite is set to false
func TestCreateFileOperation2(t *testing.T) {
	isDir := false
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo1/newFile.txt",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// expected to fail as create non existing is set to false and Flogo2 does not exist
func TestCreateFileOperation3(t *testing.T) {
	isDir := false
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo1/Flogo2/newFile.txt",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

func TestCreateFileOperation4(t *testing.T) {
	isDir := false
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo1/Flogo2/newFile.txt",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

func TestCreateDirOperationRelativePath(t *testing.T) {
	isDir := true
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "./abc/pqr2/newFile1",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

func TestCreateDirOperation(t *testing.T) {
	isDir := true
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo3",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// this will pass as overwrite is set to true
func TestCreateDirOperation1(t *testing.T) {
	isDir := true
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo3",
		 "overwrite": true
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// expected to fail as above testcase will already create same folder and in this
// testcase overwrite is set to false
func TestCreateDirOperation2(t *testing.T) {
	isDir := true
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo3",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

// expected to fail as create non existing is set to false and Flogo4 does not exist
func TestCreateDirOperation3(t *testing.T) {
	isDir := true
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo3/Flogo4/Flogo5",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

func TestCreateDirOperation4(t *testing.T) {
	isDir := true
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Flogo3/Flogo4/Flogo5",
		 "overwrite": false
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, isDir)
}

func Execute(t *testing.T, inputParams map[string]interface{}, createNonExistingDir bool, isDir bool) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-create"), activityName: "create"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{IsDir: isDir, CreateNonExistingDir: createNonExistingDir, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform create operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of create operation: %s", err.Error())
		t.Fail()
	}
}
