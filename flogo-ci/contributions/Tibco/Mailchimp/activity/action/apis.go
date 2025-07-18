package action

import (
	"fmt"

	"github.com/project-flogo/core/support/log"
	mailchimp "github.com/tibco/wi-mailchimp/src/app/Mailchimp/activity"
	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"
)

const (
	sendPath     = "/campaigns/%s/actions/send"
	schedulePath = "/campaigns/%s/actions/schedule"
	testPath     = "/campaigns/%s/actions/test"
)

type ApiAction struct {
	Log        log.Logger
	CampaignId string
	Data       interface{}
	ApiToken   *mailchimpConn.Token
}

func (a *ApiAction) DoCreate(url string) (interface{}, error) {

	var reqBody []byte
	if a.Data != nil {
		str, _ := mailchimp.InputObjectToStr(a.Data)

		a.Log.Debugf("===>Activity input body: %s", str)
		reqBody = []byte(str)
	} else {
		a.Log.Debug("===>No request body for the call")
	}

	apiResponse, err := mailchimp.PostCall(url, a.ApiToken, nil, reqBody, a.Log)
	if err != nil {
		return nil, fmt.Errorf("Fail to call Mailchimp API", err.Error())
	}
	if apiResponse.StatusCode == 204 || string(apiResponse.Body) == "" {
		// This is Succesful response with no body for all action apis on mailchimp
		return true, nil
	}

	return false, nil
}

func (a *ApiAction) CampaignSend() (interface{}, error) {
	url := a.ApiToken.APIEndpoint + mailchimp.API_VERSION_PATH + fmt.Sprintf(sendPath, a.CampaignId)
	return a.DoCreate(url)
}

func (a *ApiAction) CampaignSchedule() (interface{}, error) {
	url := a.ApiToken.APIEndpoint + mailchimp.API_VERSION_PATH + fmt.Sprintf(schedulePath, a.CampaignId)
	return a.DoCreate(url)
}

func (a *ApiAction) CampaginTest() (interface{}, error) {
	url := a.ApiToken.APIEndpoint + mailchimp.API_VERSION_PATH + fmt.Sprintf(testPath, a.CampaignId)
	return a.DoCreate(url)
}
