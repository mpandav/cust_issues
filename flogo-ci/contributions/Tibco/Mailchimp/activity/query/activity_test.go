package query

/*
import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata

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

func TestRegistered(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	if act == nil {
		t.Error("Activity Not Registered")
		t.Fail()
		return
	}
}

var connectionJson = `{
    "id": "dddddddd",
    "settings": [
        {
            "display": {
                "name": "Name",
                "placeholder": "Connection name",
                "visible": true
            },
            "inputType": "text",
            "name": "name",
            "required": true,
            "type": "string",
            "value": "dfdsfds"
        },
        {
          "name": "WI_STUDIO_OAUTH_CONNECTOR_INFO",
          "type": "string",
          "required": true,
          "display": {
           "visible": false,
           "readonly": false,
           "valid": false
          },
          "value": "{\"client_id\":\"440725332451\",\"client_secret\":\"f92a3536c3dba5b55790bc87ba26e4dd2ede84ca2b0e0aec12\",\"access_token\":\"2fb62c2692ea7825dda2da8790d9a325\",\"expires_in\":0,\"scope\":null,\"dc\":\"us17\",\"role\":\"owner\",\"accountname\":\"Ecommerce\",\"login_url\":\"https://login.mailchimp.com\",\"api_endpoint\":\"https://us17.api.mailchimp.com\"}"
         }
    ]
}`

const input = `{
	"status": "sent"
}`

func TestGetCampaigns(t *testing.T) {
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(connectionJson), &m)
	assert.Nil(t, err)
	//setup attrs
	tc.SetInput(RESOURCE, CAMPAIGNS)
	tc.SetInput(CONNECTION, m)
	var body interface{}
	err2 := json.Unmarshal([]byte(input), &body)
	assert.Nil(t, err2)

	complex := &data.ComplexObject{Metadata: "", Value: body}
	tc.SetInput("input", complex)

	//eval
	_, err = act.Eval(tc)
	assert.Nil(t, err)
}
*/
