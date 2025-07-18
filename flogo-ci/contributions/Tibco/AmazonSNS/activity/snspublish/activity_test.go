package snspublish

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	_ "github.com/tibco/flogo-aws/src/app/AWS/connector"
)

const (
	TopicArn = ""
)

func getConnectionManager() interface{} {
	connectionBytes, err := ioutil.ReadFile("../connectionData.json")
	if err != nil {
		panic("connectionData.json file found")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "connector", "git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/AWS/connector")
	connmgr, _ := coerce.ToConnection(connectionObj)
	return connmgr
}

func setupActivity(t *testing.T) (*Activity, *test.TestActivityContext) {
	act := &Activity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connection", getConnectionManager())
	return act, tc
}

func logOutput(t *testing.T, tc *test.TestActivityContext) {
	output := tc.GetOutput("output")
	assert.NotNil(t, output)
	outputBytes, err := json.Marshal(output)
	assert.Nil(t, err)
	tc.Logger().Info("output:", string(outputBytes))
}

func TestPublishPlainText(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("messageType", "plainText")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Message": "Hello from Go Test",
		"TopicArn": "` + TopicArn + `"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Publish Message due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestPublishPlainTextWithAttributes(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("messageType", "plainText")
	var attribParams []map[string]interface{}
	var attribJSON = []byte(`[
		{
				"AttributeName": "sendSMS",
				"AttributeType": "String"
		},
		{
				"AttributeName": "NumericAttr",
				"AttributeType": "Number"
		},
		{
				"AttributeName": "ArrayAttr",
				"AttributeType": "String.Array"
		}
	]`)
	err := json.Unmarshal(attribJSON, &attribParams)
	tc.SetInput("messageAttributeNames", attribParams)

	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Message": "Hello from Go Test with Attributes",
		"MessageAttributes": {
			"sendSMS": "true",
			"NumericAttr": "123",
			"ArrayAttr": "[\"string\", 123, true, null]"
		},
		"TopicArn": "` + TopicArn + `"
	}`)
	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Publish Message due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
func TestPublishCustomWithJSON(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("messageType", "custom")
	var inputParams map[string]interface{}

	var inputJSON = `{
		"Message": {
			"default": "Sample fallback message with custom JSON",
			"email": "Sample message for email endpoints with custom JSON",
			"GCM": { 
				"notification": { 
					"title": "Sample Notification with custom JSON", 
					"body": "Sample message for Android endpoints with custom JSON", 
					"color": "#99ccff", 
					"tag":"mynotiftag" 
				}
			},
			"APNS": {
				"aps": {
					"alert": "Sample message for iOS endpoints"
				}
			},
			"APNS_SANDBOX": {
				"aps": {
					"alert": "Sample message for iOS development endpoints"
				}
			}
		},
		"TopicArn": "` + TopicArn + `"
	}`

	err := json.Unmarshal([]byte(inputJSON), &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Publish Message due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
