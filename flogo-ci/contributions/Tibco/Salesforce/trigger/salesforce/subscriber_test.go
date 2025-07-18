package salesforce

import (
	"encoding/json"
	"testing"

	"github.com/project-flogo/core/trigger"
	"github.com/stretchr/testify/assert"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

var tokenStr = `{
  "access_token": "00D2w000003ygtI!ARcAQE5WfbBr1sjMIWr37cEaNhV0_GWZ_81x14pTS5DUz3SbdQyCa5K79iVBtk7jIm5aNTOI4P8UynlXx2Vi2O1JuLUxswz2",
  "refresh_token": "5Aep861ZBQbtA4s3JXHnUuWBZzblq_FoeG2MTZlkyeNaT3MDtqoUqoO5DGdkHH3GLkxu1Sa2kiIRIdFP1X_BsX3",
  "instance_url": "https://ap16.salesforce.com",
  "scope": "refresh_token visualforce wave_api custom_permissions web openid chatter_api api id eclair_api full"
  }`

func TestTrigger(t *testing.T) {
	token := &sfconnection.SalesforceToken{}
	err := json.Unmarshal([]byte(tokenStr), token)
	assert.Nil(t, err)
	sscm := &sfconnection.SalesforceSharedConfigManager{SalesforceToken: token, APIVersion: "v52.0"}
	sscm.ClientId = "3MVG9n_HvETGhr3B3UbvR1yl3tO4Sf7lOh9pivoacwEE8F1AukeZzeRUEiYPEQH53V9ew0kba7ZBbYtA9hghj"
	sscm.ClientSecret = "A3D8515EF31D4683F9D56AF9B583EBE4E13959808B94078B1F89251A7922C3E7"
	pushTopic := PushTopic{}
	pushTopic.Name = "AccountUpdates"
	sub := Subscriber{}
	sub.salesforceSharedConfigManager = sscm
	sub.topic = pushTopic
	fn := func(handler trigger.Handler, eventData interface{}) {
		logCache.Debug("-- Executing action --")
	}
	booleanVal := sub.ListenToPushTopic(fn)
	assert.True(t, booleanVal)
}
func TestChangeDataCapture(t *testing.T) {
	token := &sfconnection.SalesforceToken{}
	err := json.Unmarshal([]byte(tokenStr), token)
	assert.Nil(t, err)
	sscm := &sfconnection.SalesforceSharedConfigManager{SalesforceToken: token, APIVersion: "v52.0"}
	sscm.ClientId = "3MVG9n_HvETGhr3B3UbvR1yl3tEzI9xxdLMdnbHWM.BOsHdTwrZJDtHumqPDxEHMjz6ZdkBThgd6p7.rPWemS"
	sscm.ClientSecret = "D20139992D7886842BFD35BD055CCA3F0775AC2D085B5807F0ACF9FA1ECC512B"
	changeDataCapture := ChangeDataCapture{}
	changeDataCapture.Name = "ContactChangeEvent"
	sub := Subscriber{subscriberType: "Change Data Capture"}
	sub.salesforceSharedConfigManager = sscm
	sub.changeDataCapture = changeDataCapture
	fn := func(handler trigger.Handler, eventData interface{}) {
		logCache.Debug("-- Executing action --")
	}
	booleanVal := sub.ListenToChangeDataCapture(fn)
	assert.True(t, booleanVal)
}
func TestPlatformEvent(t *testing.T) {
	token := &sfconnection.SalesforceToken{}
	err := json.Unmarshal([]byte(tokenStr), token)
	assert.Nil(t, err)
	sscm := &sfconnection.SalesforceSharedConfigManager{SalesforceToken: token, APIVersion: "v52.0"}
	sscm.ClientId = "3MVG9n_HvETGhr3B3UbvR1yl3tEzI9xxdLMdnbHWM.BOsHdTwrZJDtHumqPDxEHMjz6ZdkBThgd6p7.rPWemS"
	sscm.ClientSecret = "D20139992D7886842BFD35BD055CCA3F0775AC2D085B5807F0ACF9FA1ECC512B"
	platformEvent := PlatformEvent{}
	platformEvent.Name = "PlatFormEventT1__e"
	sub := Subscriber{subscriberType: "Platform Event"}
	sub.salesforceSharedConfigManager = sscm
	sub.platformEvent = platformEvent
	fn := func(handler trigger.Handler, eventData interface{}) {
		logCache.Debug("-- Executing action --")
	}
	booleanVal := sub.ListenToPlatformEvent(fn)
	assert.True(t, booleanVal)
}

// func TestStop(t *testing.T) {
// 	token := &sfconnection.SalesforceToken{}
// 	err := json.Unmarshal([]byte(tokenStr), token)
// 	assert.Nil(t, err)
// 	pushTopic := PushTopic{}
// 	pushTopic.Name = "AccountUpdates"
// 	sub := Subscriber{}
// 	sub.token = token
// 	sub.topic = pushTopic
// 	fn := func(handler trigger.Handler, eventData EventData) {
// 		logCache.Debug("-- Executing action --")
// 	}
// 	booleanVal := sub.Listen(fn)
// 	assert.True(t, booleanVal)
// }
