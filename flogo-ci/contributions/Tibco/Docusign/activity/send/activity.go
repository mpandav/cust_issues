package send

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

var (
	emailRegx = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Envelope struct
type Envelope struct {
	Status       string     `json:"status"`
	EmailSubject string     `json:"emailSubject"`
	Documents    []Document `json:"documents"`
	Recipients   Recipients `json:"recipients"`
}

// Document struct
type Document struct {
	DocumentID string `json:"documentId"`
	Name       string `json:"name"`
	Content    string `json:"content,omitempty"`
	Extension  string `json:"fileExtension"`
	// DocumentBase64 string `json:"documentBase64"`
}

// Recipients struct
type Recipients struct {
	Signers []Signer `json:"signers"`
}

// Signer struct
type Signer struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	RecipientID  string `json:"recipientId"`
	RoutingOrder string `json:"routingOrder"`
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

type CreateEnvelopeActivity struct {
}

func init() {
	_ = activity.Register(&CreateEnvelopeActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &CreateEnvelopeActivity{}, nil
}

func (*CreateEnvelopeActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval runtime execution
func (docActivity *CreateEnvelopeActivity) Eval(context activity.Context) (done bool, err error) {

	logger := context.Logger()
	logger.Info("Executing DocuSign CreateEnvelope Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	recipients := input.Recipients
	if recipients == "" {
		return false, fmt.Errorf("Recipient list is empty")
	}

	var documents []Document
	var fileExtension string
	fileName := input.FileName
	fileContent := input.FileContent

	isMultiDoc := input.IsMultiDoc
	if isMultiDoc {
		documents = input.Documents
		if len(documents) < 1 {
			return false, fmt.Errorf("No documents added")
		}
		for index, _ := range documents {
			documents[index].DocumentID = strconv.Itoa(index + 1)
		}
	} else {
		if fileName == "" {
			return false, fmt.Errorf("File name is empty")
		}

		if fileContent == "" {
			return false, fmt.Errorf("File content is empty")
		}
		doc := Document{
			DocumentID: "1",
			Name:       fileName,
			Content:    fileContent,
		}
		documents = append(documents, doc)
	}

	// logger.Infof("documents: %#v", documents)

	signingOrder := input.SigningOrder

	dscm := input.DocusignConnection.(*docusignconnection.DocusignSharedConfigManager)
	if dscm == nil {
		return false, fmt.Errorf("error: connection manager is nil")
	}
	token := dscm.DocusignToken

	recipientArray := strings.Split(recipients, ",")
	envelope, err := createEnvelope(documents, recipientArray, signingOrder, logger)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	logger.Debug("Envelope created successfully")

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return false, fmt.Errorf("Failed to parse envelope [%s]", err.Error())
	}

	body := &bytes.Buffer{}
	errBody := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	mh := make(textproto.MIMEHeader)
	mh.Set("Content-Type", "application/json")
	mh.Set("Content-Disposition", "form-data")
	partWriter, err := writer.CreatePart(mh)

	if nil != err {
		return false, err
	}
	_, err = io.Copy(partWriter, bytes.NewBufferString(string(envelopeBytes)))
	if err != nil {
		return false, err
	}

	for _, doc := range documents {

		extArr := strings.Split(doc.Name, ".")
		if len(extArr) > 1 {
			fileExtension = extArr[len(extArr)-1]
		} else {
			return false, fmt.Errorf("File Extension not given for file: %s", doc.Name)
		}

		bytes := []byte(doc.Content)
		contentType := http.DetectContentType(bytes)
		// contentType := mime.TypeByExtension("." + fileExtension)
		// logger.Info("content type: ", contentType)
		reader := strings.NewReader(doc.Content)

		mh = make(textproto.MIMEHeader)
		mh.Set("Content-Type", contentType)
		mh.Set("Content-Disposition", "file; filename="+doc.Name+"; fileExtension="+fileExtension+"; documentid="+doc.DocumentID)

		partWriter, err = writer.CreatePart(mh)
		_, err = io.Copy(partWriter, reader)
	}

	err = writer.Close()
	if err != nil {
		return false, err
	}

	accInfo := dscm.DocusignAccount

	errBody.Write(body.Bytes())
	requestURL := accInfo.BaseURI + "/restapi/v2.1/accounts/" + accInfo.AccountId + "/envelopes"
	method := http.MethodPost

	// logger.Debugf("URL: %s", requestURL)
	req, _ := http.NewRequest(method, requestURL, body)

	authHeader := "Bearer " + token.AccessToken
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// logger.Info(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "dial tcp: lookup") ||
			strings.Contains(err.Error(), "TLS handshake timeout") || strings.Contains(err.Error(), "timeout") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to send envelope due to error %s", err.Error()), "docusign-send-4001", nil)
		}
		return false, fmt.Errorf("Failed to send envelope [%s]", err.Error())
	}
	defer res.Body.Close()

	respBody, err := docusignconnection.ReadResponseBody(res)
	if err != nil || respBody == nil {
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
			return false, activity.NewRetriableError(fmt.Sprintf("Failed to send envelope due to error %s", err.Error()), "docusign-send-4001", nil)
		}
		return false, fmt.Errorf("Failed to get response body [%s]", err.Error())
	}

	output := &Output{}
	var result interface{}
	json.Unmarshal(respBody, &result)
	resultMap := result.(map[string]interface{})
	// logger.Infof("result map %#v", resultMap)
	errMap := make(map[string]interface{})

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		output.Envelope = resultMap
		err = context.SetOutputObject(output)
		if err != nil {
			return false, err
		}
	} else if res.StatusCode == http.StatusUnauthorized {
		// refresh token
		err := dscm.DoRefreshToken()
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
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to send envelope due to error %s", err.Error()), "docusign-send-4001", nil)
			}
			return false, fmt.Errorf("Failed to send envelope [%s]", err.Error())
		}
		defer res.Body.Close()

		respBody, err = docusignconnection.ReadResponseBody(res)
		if err != nil || respBody == nil {
			if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "connection timed out") {
				return false, activity.NewRetriableError(fmt.Sprintf("Failed to send envelope due to error %s", err.Error()), "docusign-send-4001", nil)
			}
			return false, fmt.Errorf("Failed to get response body [%s]", err.Error())
		}

		json.Unmarshal(respBody, &result)
		resultMap = result.(map[string]interface{})

		if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
			output.Envelope = resultMap
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

	errBody.Reset()
	return true, nil
}

// func sendEnvelopeReq(body *bytes.Buffer, writer *multipart.Writer, accInfo *docusignconnection.Account, token *docusignconnection.Token, logger log.Logger) (*http.Response, error) {
// 	requestURL := accInfo.BaseURI + "/restapi/v2/accounts/" + accInfo.AccountId + "/envelopes"
// 	method := http.MethodPost

// 	logger.Debugf("URL: %s", requestURL)
// 	req, _ := http.NewRequest(method, requestURL, body)

// 	authHeader := "Bearer " + token.AccessToken
// 	req.Header.Set("Authorization", authHeader)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	res, err := http.DefaultClient.Do(req)
// 	return res, err
// }

func createEnvelope(inputDocs []Document, recipientArray []string, inOrder bool, logger log.Logger) (*Envelope, error) {
	logger.Info("Creating envelope in activity")
	var documents []Document
	var signers []Signer

	for _, document := range inputDocs {
		doc := Document{
			DocumentID: document.DocumentID,
			Name:       document.Name,
		}
		documents = append(documents, doc)
	}

	for i, v := range recipientArray {
		if validateEmail(strings.TrimSpace(v)) {
			order := ""
			if inOrder {
				order = strconv.Itoa(i + 1)
			}
			signer := Signer{
				Name:         v,
				Email:        v,
				RecipientID:  strconv.Itoa(i + 1),
				RoutingOrder: order,
			}
			signers = append(signers, signer)
		} else {
			return nil, errors.New("Email is invalid")
		}
	}

	recipients := Recipients{
		Signers: signers,
	}

	// emailSubjectt := fmt.Sprintf("Please sign %s", fileName)
	emailSubject := fmt.Sprintf("Please sign the documents")
	envelope := Envelope{
		Status:       "sent",
		EmailSubject: emailSubject,
		Documents:    documents,
		Recipients:   recipients,
	}

	return &envelope, nil
}

func validateEmail(email string) bool {
	return emailRegx.MatchString(email)
}
