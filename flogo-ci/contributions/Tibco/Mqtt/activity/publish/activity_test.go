package publish

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	mqttconnection "github.com/tibco/wi-mqtt/src/app/Mqtt/connector/connection"
)

var activityMetadata *activity.Metadata

func getConnectionManager() interface{} {
	connectionBytes, err := ioutil.ReadFile("../connectionData.json")
	if err != nil {
		panic("connectionData.json file found")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	connmgr, _ := mqttconnection.GetSharedConfiguration(connectionObj)
	return connmgr
}

func setupActivity(t *testing.T) (*MqttActivity, *test.TestActivityContext) {
	act := &MqttActivity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("Connection", getConnectionManager())
	return act, tc
}

func TestPublishString(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("topic", "my-topic")
	tc.SetInput("retain", true)
	tc.SetInput("qos", 2)
	tc.SetInput("valueType", "String")
	tc.SetInput("stringValue", "Hello World")
	_, err := act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Get Buckets due to %s", err.Error())
		t.Fail()
	}
}
