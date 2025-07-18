package listactivity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/project-flogo/core/activity"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

type EnvelopeDocuments struct {
	DocumentID string `json:"documentId"`
	Name       string `json:"name"`
	URI        string `json:"uri"`
}

type EnvelopeDetails struct {
	EnvelopeId   string              `json:"envelopeId"`
	EnvelopeDocs []EnvelopeDocuments `json:"envelopeDocuments"`
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

type ListDocumentsActivity struct {
}

func init() {
	_ = activity.Register(&ListDocumentsActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &ListDocumentsActivity{}, nil
}

func (*ListDocumentsActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (docActivity *ListDocumentsActivity) Eval(context activity.Context) (done bool, err error) {

	logger := context.Logger()
	logger.Info("Executing DocuSign ListDocuments Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	envelopeId := input.EnvelopeID
	if envelopeId == "" {
		return false, fmt.Errorf("EnvelopeID is empty")
	}

	dscm := input.DocusignConnection.(*docusignconnection.DocusignSharedConfigManager)
	if dscm == nil {
		return false, fmt.Errorf("error: connection manager is nil")
	}
	token := dscm.DocusignToken
	accInfo := dscm.DocusignAccount

	requestURL := accInfo.BaseURI + "/restapi/v2.1/accounts/" + accInfo.AccountId + "/envelopes/" + envelopeId + "/documents"

	method := http.MethodGet
	req, _ := http.NewRequest(method, requestURL, nil)
	authHeader := "Bearer " + token.AccessToken
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "dial tcp: lookup") ||
			strings.Contains(err.Error(), "TLS handshake timeout") || strings.Contains(err.Error(), "timeout") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
		}
		return false, fmt.Errorf("failed to get document [%s]", err.Error())
	}
	defer res.Body.Close()

	respBody, err := docusignconnection.ReadResponseBody(res)
	if err != nil || respBody == nil {
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
		}
		return false, fmt.Errorf("failed to get response body [%s]", err.Error())
	}

	output := &Output{}
	var result interface{}
	json.Unmarshal(respBody, &result)
	resultMap := result.(map[string]interface{})
	errMap := make(map[string]interface{})

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		jsonstring, err := json.Marshal(resultMap)
		if err != nil {
			return false, fmt.Errorf("error while marshalling repsonce [%s]", err.Error())
		}
		envelopeDetails := EnvelopeDetails{}
		err = json.Unmarshal(jsonstring, &envelopeDetails)
		if err != nil {
			return false, fmt.Errorf("error while unmarshalling reponse [%s]", err.Error())
		}
		if envelopeId != envelopeDetails.EnvelopeId {
			return false, fmt.Errorf("received envelopeId defers from provided envelopeId")
		}

		resMap := make(map[string]interface{})
		resultstring, err := json.Marshal(envelopeDetails)
		if err != nil {
			return false, fmt.Errorf("error while marshalling repsonce [%s]", err.Error())
		}
		err = json.Unmarshal(resultstring, &resMap)
		if err != nil {
			return false, fmt.Errorf("error while unmarshalling reponse [%s]", err.Error())
		}

		output.Output = resMap
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}
	} else if res.StatusCode == http.StatusUnauthorized {
		err = dscm.DoRefreshToken()
		if err != nil {
			return false, fmt.Errorf("failed to refresh token: %s", err.Error())
		}
		token = dscm.DocusignToken
		authHeader := "Bearer " + token.AccessToken
		req.Header.Set("Authorization", authHeader)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "dial tcp: lookup") ||
				strings.Contains(err.Error(), "TLS handshake timeout") || strings.Contains(err.Error(), "timeout") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
			}
			return false, fmt.Errorf("failed to get document [%s]", err.Error())
		}
		defer res.Body.Close()

		respBody, err := docusignconnection.ReadResponseBody(res)
		if err != nil || respBody == nil {
			if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
			}
			return false, fmt.Errorf("failed to get response body [%s]", err.Error())
		}

		json.Unmarshal(respBody, &result)
		resultMap = result.(map[string]interface{})

		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
			jsonstring, err := json.Marshal(resultMap)
			if err != nil {
				return false, fmt.Errorf("error while marshalling repsonce [%s]", err.Error())
			}
			envelopeDetails := EnvelopeDetails{}
			err = json.Unmarshal([]byte(jsonstring), &envelopeDetails)
			if err != nil {
				return false, fmt.Errorf("error while unmarshalling reponse [%s]", err.Error())
			}
			if envelopeId != envelopeDetails.EnvelopeId {
				return false, fmt.Errorf("received envelopeId defers from provided envelopeId")
			}
			resMap := make(map[string]interface{})
			resultstring, err := json.Marshal(envelopeDetails)
			if err != nil {
				return false, fmt.Errorf("error while marshalling repsonce [%s]", err.Error())
			}
			err = json.Unmarshal(resultstring, &resMap)
			if err != nil {
				return false, fmt.Errorf("error while unmarshalling reponse [%s]", err.Error())
			}
			output.Output = resMap
			err = context.SetOutputObject(output)
			if err != nil {
				return false, err
			}
		} else {
			errMap["errorCode"] = res.StatusCode
			errMap["message"] = resultMap["message"].(string)
			output.Error = errMap
			err = context.SetOutputObject(output)
			if err != nil {
				return false, err
			}
			return false, fmt.Errorf("%s", errMap["message"].(string))
		}
	} else {
		errMap["errorCode"] = res.StatusCode
		errMap["message"] = resultMap["message"].(string)
		output.Error = errMap
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}

		return false, fmt.Errorf("%s", errMap["message"].(string))
	}

	return true, nil
}
