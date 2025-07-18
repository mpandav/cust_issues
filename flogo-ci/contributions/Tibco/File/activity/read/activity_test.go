package read

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

func TestReadTextFileOperation(t *testing.T) {
	readAs := "Text"
	compress := "None"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/a.txt"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, readAs, compress)
}

func TestReadBinaryFileOperation(t *testing.T) {
	readAs := "Binary"
	compress := "None"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/a.txt"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, readAs, compress)
}

// C:/Flogo/textcompress.txt is a compressed text file using write file test
func TestReadCompressedTextFile(t *testing.T) {
	readAs := "Text"
	compress := "GUnZip"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/textcompress.txt"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, readAs, compress)
}

// C:/Flogo/binarycompress.txt is a compressed binary file using write file test
func TestReadCompressedBinaryFile(t *testing.T) {
	readAs := "Text"
	compress := "GUnZip"

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "fileName": "C:/Flogo/binarycompress.txt"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, readAs, compress)
}

func Execute(t *testing.T, inputParams map[string]interface{}, readAs string, compress string) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "File-read"), activityName: "read"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	aInput := &Input{ReadAs: readAs, Compress: compress, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform read operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of read operation: %s", err.Error())
		t.Fail()
	}
}
