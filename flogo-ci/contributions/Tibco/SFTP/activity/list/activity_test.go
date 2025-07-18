package list

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

func TestListOperation(t *testing.T) {
	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "Remote Directory": "folder"
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams)
}

func Execute(t *testing.T, inputParams map[string]interface{}) {
	listActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "SFTP-list"), activityName: "list"}
	//set logging to debug level
	log.SetLogLevel(listActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(listActivity.Metadata())

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

	aInput := &Input{Connection: connManager, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, err := listActivity.Eval(tc)
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
