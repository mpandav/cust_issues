package check

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// CheckStatusActivity struct
type CheckStatusActivity struct {
}

func init() {
	_ = activity.Register(&CheckStatusActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &CheckStatusActivity{}, nil
}

func (*CheckStatusActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval runtime execution
func (docActivity *CheckStatusActivity) Eval(context activity.Context) (done bool, err error) {

	logger := context.Logger()
	logger.Info("Executing DocuSign CheckStatus Activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	envelopeId := input.EnvelopeID
	if envelopeId == "" {
		return false, fmt.Errorf("Envelope Id is required")
	}

	dscm := input.DocusignConnection.(*docusignconnection.DocusignSharedConfigManager)
	if dscm == nil {
		return false, fmt.Errorf("error: connection manager is nil")
	}
	token := dscm.DocusignToken
	accInfo := dscm.DocusignAccount

	//form a request
	requestURL := accInfo.BaseURI + "/restapi/v2.1/accounts/" + accInfo.AccountId + "/envelopes/" + envelopeId
	method := http.MethodGet
	req, _ := http.NewRequest(method, requestURL, nil)
	authHeader := "Bearer " + token.AccessToken
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "dial tcp: lookup") ||
			strings.Contains(err.Error(), "TLS handshake timeout") || strings.Contains(err.Error(), "timeout") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to check envelope status due to error %s", err.Error()), "docusign-check-4003", nil)
		}
		return false, fmt.Errorf("Failed to check envelope status [%s]", err.Error())
	}
	defer res.Body.Close()

	respBody, err := docusignconnection.ReadResponseBody(res)
	if err != nil || respBody == nil {
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to check envelope status due to error %s", err.Error()), "docusign-check-4003", nil)
		}
		return false, fmt.Errorf("Failed to get response body [%s]", err.Error())
	}

	output := &Output{}
	var result interface{}
	json.Unmarshal(respBody, &result)
	envelopeMap := result.(map[string]interface{})
	errMap := make(map[string]interface{})

	if res.StatusCode == http.StatusOK {
		status := envelopeMap["status"]
		output.Status = status.(string)
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}
	} else if res.StatusCode == http.StatusUnauthorized {
		// refresh token
		err = dscm.DoRefreshToken()
		if err != nil {
			return false, fmt.Errorf("Failed to refresh token: %s", err.Error())
		}

		token = dscm.DocusignToken

		authHeader = "Bearer " + token.AccessToken
		req.Header.Set("Authorization", authHeader)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "dial tcp: lookup") ||
				strings.Contains(err.Error(), "TLS handshake timeout") || strings.Contains(err.Error(), "timeout") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to check envelope status due to error %s", err.Error()), "docusign-check-4003", nil)
			}
			return false, fmt.Errorf("Failed to check envelope status [%s]", err.Error())
		}
		defer res.Body.Close()

		respBody, err := docusignconnection.ReadResponseBody(res)
		if err != nil || respBody == nil {
			if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to check envelope status due to error %s", err.Error()), "docusign-check-4003", nil)
			}
			return false, fmt.Errorf("Failed to get response body [%s]", err.Error())
		}

		json.Unmarshal(respBody, &result)
		envelopeMap = result.(map[string]interface{})
		if res.StatusCode == http.StatusOK {
			status := envelopeMap["status"]
			output.Status = status.(string)
			err = context.SetOutputObject(output)
			if err != nil {
				return false, err
			}
		} else {
			errMap["errorCode"] = res.StatusCode
			errMap["message"] = envelopeMap["message"].(string)
			output.Error = errMap
			err = context.SetOutputObject(output)
			if err != nil {
				return false, err
			}
			return false, fmt.Errorf("%s", errMap["message"].(string))
		}

	} else {
		errMap["errorCode"] = res.StatusCode
		errMap["message"] = envelopeMap["message"].(string)
		output.Error = errMap
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}

		return false, fmt.Errorf("%s", errMap["message"].(string))
	}

	return true, nil
}

// func sendGetStatusReq(envelopeId string, accInfo *docusignconnection.Account, token *docusignconnection.Token) (*http.Response, error) {
// 	requestURL := accInfo.BaseURI + "/restapi/v2/accounts/" + accInfo.AccountId + "/envelopes/" + envelopeId
// 	method := http.MethodGet
// 	req, _ := http.NewRequest(method, requestURL, nil)
// 	authHeader := "Bearer " + token.AccessToken
// 	req.Header.Set("Authorization", authHeader)
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := http.DefaultClient.Do(req)
// 	return res, err
// }
