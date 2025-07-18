package mailchimp

import (
	"bytes"

	"fmt"
	"net/http"
	"net/url"

	"io"

	"github.com/project-flogo/core/support/log"
	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"
)

const API_VERSION_PATH = "/3.0"
const TOKEN_URL = ""

// Response struct for wrapping the response
type APIResponse struct {
	StatusCode int
	Body       []byte
}

func (res *APIResponse) ToString() string {
	if res != nil {
		if res.Body != nil {
			return string(res.Body)
		}
	}
	return ""
}

// GETCall  REST get request
func GetCall(geturl string, token *mailchimpConn.Token, queryParams map[string][]string, headers map[string]string, log log.Logger) (*APIResponse, error) {
	var requestUrl = geturl
	if queryParams != nil && len(queryParams) > 0 {
		qp := url.Values{}
		var hasQuery bool
		for k, value := range queryParams {
			qp.Add(k, ArrayToQueryParameters(value))
			hasQuery = true
		}
		if hasQuery {
			requestUrl = requestUrl + "?" + qp.Encode()
		}
	}

	log.Debugf("request URL %s", requestUrl)
	request, _ := http.NewRequest("GET", requestUrl, nil)
	request.Header.Add("Authorization", "Bearer "+token.AccessToken)
	if headers != nil {
		for k, v := range headers {
			request.Header.Add(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	apiResponse := &APIResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}
	log.Debugf("Response code [%d] and status [%s]", resp.StatusCode, resp.Status)

	if apiResponse.StatusCode >= 400 {
		// error
		return nil, fmt.Errorf("Error response with code %d, msg: %s", resp.StatusCode, string(respBody))
	}

	if apiResponse.InvalidTokenError() {
		log.Info("We are encountering an invalid token error")
		//----
		// There is no Refresh token in Mailchimp API, so this part is not needed
		//----
		// return GetCall(geturl, token, queryParams, headers, log)

	} else if apiResponse.ServerUnavailable() {
		log.Info("API server temporarily unavailable, retry")
		return GetCall(geturl, token, queryParams, headers, log)
	}

	log.Debugf("Response body %s", string(respBody))
	return apiResponse, err
}

// PostCall REST POST request
func PostCall(postURL string, mt *mailchimpConn.Token, queryParams map[string][]string, reqBodyBytes []byte, log log.Logger) (*APIResponse, error) {

	var requestUrl = postURL
	if queryParams != nil && len(queryParams) > 0 {
		qp := url.Values{}
		for k, value := range queryParams {
			for _, v := range value {
				if v != "" {
					qp.Add(k, v)
				}
			}
		}
		requestUrl = requestUrl + "?" + qp.Encode()
	}
	log.Debugf("Post call, request URL: %s", requestUrl)

	var reqBody io.Reader

	if reqBodyBytes != nil {
		reqBody = bytes.NewBuffer(reqBodyBytes)

	} else {
		reqBody = nil
	}
	req, err := http.NewRequest("POST", requestUrl, reqBody)
	if err != nil {
		log.Errorf("New http request faliure %s", err.Error())
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+mt.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Response body %s", string(respBody))

	if res.StatusCode >= 400 {
		// error
		return nil, fmt.Errorf("Error response with code %d, msg: %s", res.StatusCode, string(respBody))
	}

	apiResponse := &APIResponse{
		StatusCode: res.StatusCode,
		Body:       respBody,
	}
	// if res.StatusCode == 204 || string(body) == "" {
	// Handle 204 No Content response in Action APIs
	// }

	if apiResponse.InvalidTokenError() {
		log.Info("We are encountering an invalid token error")
		//----
		// There is no Refresh token in Mailchimp API, so this part is not needed
		//----
		// apiRsp, err = PostCall(postURL, mt, queryParams, reqBodyBytes, log)

	} else if apiResponse.ServerUnavailable() {
		log.Info("API server temporarily unavailable, retry")
		apiResponse, err = PostCall(postURL, mt, queryParams, reqBodyBytes, log)
	}

	return apiResponse, nil
}

// DeleteCall REST DELETE request
func DeleteCall(url string, mt *mailchimpConn.Token, log log.Logger) (*APIResponse, error) {

	log.Debugf("Delete call, request URL: %s", url)

	req, _ := http.NewRequest("DELETE", url, nil)

	req.Header.Add("Authorization", "Bearer "+mt.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	apiRsp := &APIResponse{
		StatusCode: res.StatusCode,
		Body:       body,
	}

	if apiRsp.InvalidTokenError() {
		log.Info("We are encountering an invalid token error", apiRsp.ToString())
		//----
		// There is no Refresh token in Mailchimp API, so this part is not needed
		//----
		// apiRsp, err = DeleteCall(url, mt, log)

	} else if apiRsp.ServerUnavailable() {
		log.Info("API server temporarily unavailable, retry")
		apiRsp, err = DeleteCall(url, mt, log)
	}

	return apiRsp, nil
}

// Token error
func (r *APIResponse) InvalidTokenError() bool {

	return false
}

// Server error
func (r *APIResponse) ServerUnavailable() bool {
	return false
}
