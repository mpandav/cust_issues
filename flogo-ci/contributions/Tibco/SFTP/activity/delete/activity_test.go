package delete

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

func TestDeleteOperation(t *testing.T) {
	inputParams := make(map[string]interface{})
	//multiple files
	var inputJSON = []byte(`{
		 "Remote File Name": "test*.txt"
	}`)

	//single file
	/*var inputJSON = []byte(`{
		"Remote File Name": "test.txt"
	}`)*/

	//empty directory
	/*var inputJSON = []byte(`{
		"Remote File Name": "pqr/x"
	}`)*/

	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	Execute(t, inputParams)
}

func Execute(t *testing.T, inputParams map[string]interface{}) {
	deleteActivity := &MyActivity{logger: log.ChildLogger(log.RootLogger(), "SFTP-delete"), activityName: "delete"}
	//set logging to debug level
	log.SetLogLevel(deleteActivity.logger, log.DebugLevel)

	tc := test.NewActivityContext(deleteActivity.Metadata())

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
	ok, err := deleteActivity.Eval(tc)
	assert.True(t, ok)
	if err != nil {
		t.Errorf("Failed to perform delete operation: %s", err.Error())
		t.Fail()
	}
}
