package write

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

func TestWriteTextFileOperation(t *testing.T) {
	writeAs := "Text"
	compress := "None"
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/a.txt",
		 "overwrite": false,
		 "textContent": "Hello World"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

// Append the content
func TestWriteTextFileOperation1(t *testing.T) {
	writeAs := "Text"
	compress := "None"
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/a.txt",
		 "overwrite": false,
		 "textContent": "Hello World"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

// Overwrite the existing file
func TestWriteTextFileOperation2(t *testing.T) {
	writeAs := "Text"
	compress := "None"
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/a.txt",
		 "overwrite": true,
		 "textContent": "Hello World"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

// Demo folder does not exist under C:/Flogo, and createNonExistingDir is false so this test case will fail
func TestWriteTextFileOperation3(t *testing.T) {
	writeAs := "Text"
	compress := "None"
	createNonExistingDir := false

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Demo/a.txt",
		 "overwrite": true,
		 "textContent": "Hello World"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

func TestWriteTextFileOperation4(t *testing.T) {
	writeAs := "Text"
	compress := "None"
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Demo/a.txt",
		 "overwrite": true,
		 "textContent": "Hello World"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

func TestWriteBinaryFileOperation1(t *testing.T) {
	writeAs := "Binary"
	compress := "None"
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/Demo/binary.txt",
		 "overwrite": true,
		 "binaryContent": "SGVsbG8gV29ybGQ="
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

func TestWriteTextCompress(t *testing.T) {
	writeAs := "Text"
	compress := "GZip"
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/textcompress.txt",
		 "overwrite": true,
		 "textContent": "Hello World"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

func TestWriteBinaryCompress(t *testing.T) {
	writeAs := "Binary"
	compress := "GZip"
	createNonExistingDir := true

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/binarycompress.txt",
		 "overwrite": true,
		 "binaryContent": "SGVsbG8gV29ybGQ="
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, writeAs, compress, createNonExistingDir)
}

func Execute(t *testing.T, inputParams map[string]interface{}, writeAs string, compress string, createNonExistingDir bool) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-write"), activityName: "write"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{WriteAs: writeAs, CreateNonExistingDir: createNonExistingDir, Compress: compress, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform write operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of write operation: %s", err.Error())
		t.Fail()
	}
}
