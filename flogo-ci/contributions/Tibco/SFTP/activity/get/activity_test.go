package get

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"

	sftpconnection "github.com/tibco/flogo-sftp/src/app/SFTP/connector/connection"
)

var sftpUserPasswordConnectionJSON = []byte(`{
	"name": "sftp",
	"description": "",
	"host": "localhost",
	"port": 22,
	"user": "tester",
	"password": "password",
	"publicKeyFlag": false,
	"hostKeyFlag": false
}`)

func TestRegister(t *testing.T) {
	ref := activity.GetRef(&MyActivity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestFileTransferGetOperation(t *testing.T) {
	processData := false
	overwrite := false
	binary := false // not being used for file to file transfer
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote File Name": "test11.PNG",
		 "Local File Name": "C:/Users/akabra/Downloads/SFTP/goTest.PNG"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, processData, overwrite, binary)
}

func TestProcessDataGetOperation(t *testing.T) {
	processData := true
	overwrite := true //this will not make any impact, as overwrite flag is not considered when processData is true
	binary := false
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote File Name": "test.txt"
	 }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, processData, overwrite, binary)
}

func TestProcessDataGetOperationBinary(t *testing.T) {
	processData := true
	overwrite := true //this will not make any impact, as overwrite flag is not considered when processData is true
	binary := true
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote File Name": "small.csv"
	 }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, processData, overwrite, binary)
}

func Execute(t *testing.T, inputParams map[string]interface{}, processData bool, overwrite bool, binary bool) {
	getActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "SFTP-get"), activityName: "get"}
	//set logging to debug level
	log.SetLogLevel(getActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(getActivity.Metadata())

	//connection
	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(sftpUserPasswordConnectionJSON), &conn)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	s := &sftpconnection.SftpFactory{}
	connManager, err := s.NewManager(conn)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	aInput := &Input{Connection: connManager, ProcessData: processData, Overwrite: overwrite, Binary: binary, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := getActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform get operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of get operation: %s", err.Error())
		t.Fail()
	}
}
