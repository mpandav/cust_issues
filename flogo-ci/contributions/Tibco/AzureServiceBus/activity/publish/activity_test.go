package publish

/*import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	azservicebus "git.tibco.com/git/product/ipaas/wi-azservicebus.git/src/app/AzureServiceBus"
	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata
var serviceBusConnectionJSON = `{
	"id": "09630510-6e30-11e8-aa07-f9c6f2436efe",
	"title": "ServiceBus Connector",
	"name": "tibco-office",
	"author": "TIBCO Software Inc.",
	"type": "flogo:connector",
	"version": "1.0.0",
	"display": {
	  "description": "Operations on office",
	  "category": "office",
	  "visible": true
	},
	"ref": "git.tibco.com/git/product/ipaas/wi-office.git/office/connector/connection",
	"keyfield": "name",
	"settings": [{
		"name": "name",
		"type": "string",
		"required": true,
		"display": {
		  "name": "Connection Name",
		  "visible": true
		}
	  },
	  {
		"name": "description",
		"type": "string",
		"display": {
		  "name": "Description",
		  "visible": true
		}
	},
	{
		"name": "resourceURI",
		"type": "string",
		"required": true,
		"display": {
		  "name": "Resource URI",
		  "visible": true
		},
		"value":"tibcojetblueexp"
	  },
	  {
		"name": "authorizationRuleName",
		"type": "string",
		"required": true,
		"display": {
		  "name": "Resource URI",
		  "visible": true
		},
		"value":"golangtest"
	  },
	  {
		"name": "primarysecondaryKey",
		"type": "string",
		"required": true,
		"display": {
		  "name": "Resource URI",
		  "visible": true
		},
		"value":"vhMKaTQec0YkqLG5963xz8QEezJKpQ+unBxquWG9qMY="
	  },
	  {
		"name": "WI_STUDIO_OAUTH_CONNECTOR_INFO",
		"type": "string",
		"required": true,
		"display": {
		  "visible": false
		},
		"value":"SharedAccessSignature sr=https%3A%2F%2Fspaddindev.servicebus.windows.net%2F&sig=%2F9GWuWYk3QU0SMm5aBYWiimfRl1qa%2FLUm7T9%2FVaEODI%3D&se=1533721824&skn=PluginServiceBus"
	  }
	],
	"actions": [{
	  "name": "Login"
	}]
  }`

func getConnector(t *testing.T, jsonConnection string) (map[string]interface{}, error) {
	connector := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonConnection), &connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
		return nil, err
	}
	return connector, nil
}

func getConnection(t *testing.T, jsonConnection string) (connection *azservicebus.Connection, err error) {
	connector, err := getConnector(t, jsonConnection)
	assert.NotNil(t, connector)

	connection, err = azservicebus.GetConnection(connector)
	if err != nil {
		t.Errorf("azservicebus  get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	return
}

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}
	return activityMetadata
}

func TestPublishMessagetoQueue(t *testing.T) {

	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Publish Message to Queue  start****")
	activity := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(activity.Metadata())
	var connection interface{}
	err := json.Unmarshal([]byte(serviceBusConnectionJSON), &connection)
	if err != nil {
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}
	cmap := connection.(map[string]interface{})
	settings := cmap["settings"]
	cmap["settings"] = settings
	tc.SetInput("Connection", cmap)
	tc.SetInput("entityType", "Queue")

	//	/api/campaign/version/3/do/read/id/<id>?...
	//https://pi.servicebus.com/api/login/version/3
	// queryURL := connection.baseURL + "/" + objectName + "/version/3/do/read/{id}"

	var inputParams interface{}
	var inputJSON = []byte(`{"parameters":{
		"queueName": "queue1aaaa",
		"messageString":"<string xmlns=\"http://schemas.microsoft.com/2003/10/Serialization/\">This is a test message now.</string>",
		"brokerProperties":{"Label":"Test","SessionId":"12"}
		}
	}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	complexInput := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complexInput)

	_, err = activity.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not fetch campaign by id, %s", err.Error())
		fmt.Printf("Error: %s", err.Error())
		t.Fail()
	} else {
		complexOutput := tc.GetOutput("output")
		assert.NotNil(t, complexOutput)
		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)

		jsonString := string(dataBytes)
		t.Logf("%s", jsonString)
		fmt.Println(jsonString)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}
func TestPublishMessagetoTopic(t *testing.T) {

	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Publish Message to Topic start****")
	activity := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(activity.Metadata())
	var connection interface{}
	err := json.Unmarshal([]byte(serviceBusConnectionJSON), &connection)
	if err != nil {
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}
	cmap := connection.(map[string]interface{})
	settings := cmap["settings"]
	cmap["settings"] = settings
	tc.SetInput("Connection", cmap)
	tc.SetInput("entityType", "Topic")

	//	/api/campaign/version/3/do/read/id/<id>?...
	//https://pi.servicebus.com/api/login/version/3
	// queryURL := connection.baseURL + "/" + objectName + "/version/3/do/read/{id}"

	var inputParams interface{}
	var inputJSON = []byte(`{"parameters" :{
			"topicName": "topic1",
			"messageString":"<string xmlns=\"http://schemas.microsoft.com/2003/10/Serialization/\">This is a test message.</string>,
			"SessionId":""
			}
		}`)

	err = json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	complexInput := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complexInput)

	_, err = activity.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Could not fetch campaign by id, %s", err.Error())
		fmt.Printf("Error: %s", err.Error())
		t.Fail()
	} else {
		complexOutput := tc.GetOutput("output")
		assert.NotNil(t, complexOutput)
		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)

		jsonString := string(dataBytes)
		t.Logf("%s", jsonString)
		fmt.Println(jsonString)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
	}
}
*/
