package put

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

func TestFileTransferPutOperation(t *testing.T) {
	processData := false
	overwrite := true
	binary := false // not being used for file to file transfer
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote File Name": "putFileTransfer.PNG",
		 "Local File Name": "C:/Users/akabra/Downloads/SFTP/goTest.PNG"
	 }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, processData, overwrite, binary)
}

func TestProcessDataPutOperation(t *testing.T) {
	processData := true
	overwrite := true
	binary := false
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote File Name": "putDataTransfer.txt",
		 "ASCII Data": "This is a put activity test"
	 }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, processData, overwrite, binary)
}

func TestProcessDataPutOperationBinary(t *testing.T) {
	processData := true
	overwrite := true
	binary := true
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote File Name": "putDataTransferBinary.txt",
		 "Binary Data": "VGhpcyBpcyBhIHB1dCBhY3Rpdml0eSB0ZXN0IGZvciBCaW5hcnkgZmxhZw=="
	 }`)

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams, processData, overwrite, binary)
}

func Execute(t *testing.T, inputParams map[string]interface{}, processData bool, overwrite bool, binary bool) {
	putActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "SFTP-put"), activityName: "put"}
	//set logging to debug level
	log.SetLogLevel(putActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(putActivity.Metadata())

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
	ok, err := putActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform put operation: %s", err.Error())
		t.Fail()
	}

	aOutput := &Output{}
	err = tc.GetOutputObject(aOutput)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Failed to get output of put operation: %s", err.Error())
		t.Fail()
	}
}
