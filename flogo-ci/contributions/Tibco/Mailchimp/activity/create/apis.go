package create

import (
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/support/log"
	mailchimp "github.com/tibco/wi-mailchimp/src/app/Mailchimp/activity"
	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"
)

const (
	listPath   = "/lists"
	memberPath = "/lists/%s/members"
)

type ApiCreate struct {
	Log           log.Logger
	ActivityInput map[string]interface{}
	ApiToken      *mailchimpConn.Token
}

func (a *ApiCreate) DoCreate(url string) (interface{}, error) {

	var reqBody []byte
	var err error
	if a.ActivityInput != nil {
		reqBody, err = json.Marshal(a.ActivityInput)
		if err != nil {
			return nil, err

		}
		a.Log.Debugf("===>Activity input: %s", string(reqBody))

	} else {
		a.Log.Debug("===>No request body for the call")
	}

	apiResponse, err := mailchimp.PostCall(url, a.ApiToken, nil, reqBody, a.Log)

	if err != nil {
		return nil, fmt.Errorf("Fail to call Mailchimp API", err.Error())
	}
	a.Log.Debug("Status code: ", apiResponse.StatusCode)

	return apiResponse.Body, nil
}

func (a *ApiCreate) List() (interface{}, error) {
	url := a.ApiToken.APIEndpoint + mailchimp.API_VERSION_PATH + listPath
	return a.DoCreate(url)
}

func (a *ApiCreate) Member(listId string) (interface{}, error) {
	url := a.ApiToken.APIEndpoint + mailchimp.API_VERSION_PATH + fmt.Sprintf(memberPath, listId)
	return a.DoCreate(url)
}
