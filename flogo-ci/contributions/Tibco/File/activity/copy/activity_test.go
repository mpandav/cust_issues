package copy

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

// Prerequisite :-
// Folder C:\Flogo has a.txt, pqr.txt, test.PNG
// Folder C:\Flogo\abc used in the tests has
// a.txt, a1.txt, b.txt, newfile, folder -> a.txt, folder -> folder1 -> a.txt

// copy file

// copy png file
func TestCopyFileOperation0(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/test.PNG",
		"toFileName": "C:/Flogo/Flogo1/test1.PNG",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// toFileName only has a location and filename
func TestCopyFileOperation(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1/b.txt",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// toFileName only has a location and not filename
func TestCopyFileOperation1(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// target file should be overwritten
func TestCopyFileOperationOverwrite(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// this will create a file of name Flogo2 and copy the content of a.txt to it
func TestCopyFileOperation2(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1/Flogo2",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// File create non existing directory tests

// this will create non existing folder Flogo3 first and then
// a file of name Flogo2 and copy the content of a.txt to it
func TestCopyFileOperationNonExisting(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1/Flogo3/Flogo2",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

func TestCopyFileOperationNonExisting1(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1/Flogo3/Flogo4/b.txt",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// here Flogo4 directory will be available from above test so the toFileName is an available folder structure
// create a.txt under Flogo4 and copy the content
func TestCopyFileOperationNonExisting2(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1/Flogo3/Flogo4",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// Test case will show a warn message as source file zz.txt does not exists
func TestCopyFileOperationSourceFileNotPresent(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/zz.txt",
		"toFileName": "C:/Flogo/zz1.txt",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// this test case should fail as we are trying to copy a folder to a file
// here dest file should exist
func TestCopyFileOperationDirectoryToFile(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/pqr.txt",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// Flogo30 does not exist and create non existing is set to false, so this test case will fail
func TestCopyFileOperationNonExistingFalse(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/a.txt",
		"toFileName": "C:/Flogo/Flogo1/Flogo30/b.txt",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// copy file ends

// copy folder

// this test will create a folder abc at path C:/Flogo/Flogo1
// and copy all files from source abc folder
func TestCopyFolderOperation(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo1",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// Folder include subdirectories tests

// this test will create a folder abc at path C:/Flogo/Flogo11
// and copy all files and folders from source abc folder including subdirectories
func TestCopyFolderOperation1(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := true // this is changed

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo11",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// this test will create a folder abc at path C:/Flogo/Flogo1/abc (existing)
// and copy all files and folders from source abc folder
func TestCopyFolderOperation2(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo1/abc",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// Folder overwrite tests

// folder abc already exists under destination folder C:/Flogo/Flogo1
// as overwrite is set to false this will fail
func TestCopyFolderOperationOverwrite(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo1",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// folder abc already exists under destination folder C:/Flogo/Flogo1
// as overwrite is set to true but folder deletion will fail as it is not empty
func TestCopyFolderOperationOverwrite1(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo1",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// folder abc already exists (make sure it is empty) under destination folder C:/Flogo/Flogo1
// since overwrite is true, abc will be deleted and source abc will be copied
func TestCopyFolderOperationOverwrite2(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo1",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// Folder create non existing directory tests
func TestCopyFolderOperationNonExisting(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc",
		"toFileName": "C:/Flogo/Flogo1/Flogo4/Flogo3",
		"overwrite": true
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

//copy folder ends

// wildcard

// only *.txt files will be copied to Flogo4 folder
func TestCopyWildcardOperation(t *testing.T) {
	createNonExistingDir := false
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc/*.txt",
		"toFileName": "C:/Flogo/Flogo1/Flogo4",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// all files and directories are copied to FLogo3 folder
// but directories under directory will not be copied as includeSubDirectories is set false
func TestCopyWildcardOperation1(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc/*",
		"toFileName": "C:/Flogo/Flogo3",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// all files and sub directories will be copied to Flogo4
func TestCopyWildcardOperation2(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := true //changed this

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc/*",
		"toFileName": "C:/Flogo/Flogo4",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

func TestCopyWildcardOperationNonExistingDirectory(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc/*",
		"toFileName": "C:/Flogo/Flogo4/Flogo5",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// this test case should fail as we are trying to copy multiple source to single destination file
// source and destination should exist
func TestCopyWildcardOperationMultipleSourceToSingleDestFile(t *testing.T) {
	createNonExistingDir := true
	includeSubDirectories := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		"fromFileName": "C:/Flogo/abc/*",
		"toFileName": "C:/Flogo/a.txt",
		"overwrite": false
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, createNonExistingDir, includeSubDirectories)
}

// wildcard ends

func Execute(t *testing.T, inputParams map[string]interface{}, createNonExistingDir bool, includeSubDirectories bool) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-copy"), activityName: "copy"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{CreateNonExistingDir: createNonExistingDir, IncludeSubDirectories: includeSubDirectories, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform copy operation: %s", err.Error())
		t.Fail()
	}
}
