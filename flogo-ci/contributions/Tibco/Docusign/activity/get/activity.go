package get

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/flow/instance"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

type RetrieveDocumentActivity struct {
}

func init() {
	_ = activity.Register(&RetrieveDocumentActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &RetrieveDocumentActivity{}, nil
}

func (*RetrieveDocumentActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval runtime execution
func (docActivity *RetrieveDocumentActivity) Eval(context activity.Context) (done bool, err error) {

	logger := context.Logger()
	logger.Info("Executing DocuSign RetrieveDocument Activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	envelopeId := input.EnvelopeID
	if envelopeId == "" {
		return false, fmt.Errorf("envelopeId is empty")
	}

	output := &Output{}
	var docURI string
	var documentId string

	getAllDocuments := input.GetAllDocs
	if getAllDocuments {
		outputType := context.(*instance.TaskInst).Task().ActivityConfig().GetOutput("outputType")
		// outputType := context.(*test.TestActivityContext).GetOutput("outputType")
		if outputType == "PDF" {
			docURI = "combined"
		} else if outputType == "ZIP" {
			docURI = "archive"
		}
	} else {
		documentId = input.DocumentID
		docId, err := strconv.Atoi(documentId)
		if err != nil {
			if documentId == "certificate" {
				docURI = documentId
			} else if documentId == "" {
				docURI = "1"
			} else {
				return false, fmt.Errorf("Enter appropriate value for documentId")
			}
		} else {
			if docId > 0 {
				docURI = documentId
			} else if docId <= 0 {
				return false, fmt.Errorf("documentId cannot be less than 1")
			}
		}
	}
	dscm := input.DocusignConnection.(*docusignconnection.DocusignSharedConfigManager)
	if dscm == nil {
		return false, fmt.Errorf("error: connection manager is nil")
	}

	token := dscm.DocusignToken
	accInfo := dscm.DocusignAccount

	requestURL := accInfo.BaseURI + "/restapi/v2.1/accounts/" + accInfo.AccountId + "/envelopes/" + envelopeId + "/documents/" + docURI

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
		return false, fmt.Errorf("Failed to get document [%s]", err.Error())
	}
	defer res.Body.Close()

	respBody, err := docusignconnection.ReadResponseBody(res)
	if err != nil || respBody == nil {
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
		}
		return false, fmt.Errorf("Failed to get response body [%s]", err.Error())
	}

	var result interface{}
	errMap := make(map[string]interface{})

	if res.StatusCode == http.StatusOK {
		base64Str := base64.StdEncoding.EncodeToString(respBody)
		output.FileContent = base64Str
		output.FileType = res.Header.Get("Content-Type")
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
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "dial tcp: lookup") ||
				strings.Contains(err.Error(), "TLS handshake timeout") || strings.Contains(err.Error(), "timeout") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
			}
			return false, fmt.Errorf("Failed to get document [%s]", err.Error())
		}
		defer res.Body.Close()

		respBody, err = docusignconnection.ReadResponseBody(res)
		if err != nil || respBody == nil {
			if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to get document due to error %s", err.Error()), "docusign-get-4002", nil)
			}
			return false, fmt.Errorf("Failed to get response body [%s]", err.Error())
		}

		if res.StatusCode == http.StatusOK {
			base64Str := base64.StdEncoding.EncodeToString(respBody)
			output.FileContent = base64Str
			output.FileType = res.Header.Get("Content-Type")
			err = context.SetOutputObject(output)
			if err != nil {
				return false, err
			}
		} else {
			json.Unmarshal(respBody, &result)
			resultMap := result.(map[string]interface{})
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
		json.Unmarshal(respBody, &result)
		resultMap := result.(map[string]interface{})
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
