package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/project-flogo/core/support/log"

	"github.com/project-flogo/core/activity"
)

const (
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"
)

var UCSREQUEST_ID = "ucsRequestId"

func GetHttpClient(logCache log.Logger, timeout int) (http.Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		logCache.Errorf("Got error while creating cookie jar %s", err.Error())
	}
	client := &http.Client{Jar: jar}

	httpTransportSettings := &http.Transport{}

	if timeout > 0 {
		httpTransportSettings.ResponseHeaderTimeout = time.Second * time.Duration(timeout)
	}

	client.Transport = httpTransportSettings

	return *client, nil
}

func GetRestResponse(activityLogger log.Logger, client http.Client, method string, uri string, headers map[string]string, reqBody io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, uri, reqBody)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	//UCS-359
	requestId := req.Header.Get(UCSREQUEST_ID)
	if requestId == "" {
		requestId, _ = GetUniqueId()
		requestId = "flogo-RT-" + requestId
		//insert this header to identify this request uniquely in the logs for correlation etc.
		req.Header.Add(UCSREQUEST_ID, requestId)
	}

	resp, err := client.Do(req)
	if err != nil {
		activityLogger.Errorf("Error occurred while executing operation: %s", err)
		return nil, err
	}

	// check for 503, 504 and return retriable error
	if resp.StatusCode == 503 || resp.StatusCode == 504 {
		return nil, activity.NewRetriableError("Failed to execute operation", "ucs-activity-4001", nil)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		if resp.Body != nil {
			badresponsebody := GetBodyAsText(resp.Body)
			if badresponsebody != "" {
				return nil, errors.New("ResponseCode: " + resp.Status + "Bad Request: " + badresponsebody)
			} else {
				return nil, errors.New("ResponseCode: " + resp.Status)
			}
		}
		return nil, errors.New("ResponseCode: " + resp.Status)
	}

	if resp.StatusCode == 204 {
		return nil, activity.NewRetriableError("Failed to execute operation", "ucs-activity-4001", nil)
	}

	if resp == nil {
		return resp, errors.New("Empty Response")
	}

	return resp, nil
}

func GetBodyAsText(respBody io.ReadCloser) string {

	defer func() {
		if respBody != nil {
			_ = respBody.Close()
		}
	}()

	var response = ""

	if respBody != nil {
		b := new(bytes.Buffer)
		b.ReadFrom(respBody)
		response = b.String()
	}

	return response
}

func GetBodyAsJSON(respBody io.ReadCloser) (interface{}, error) {

	defer func() {
		if respBody != nil {
			_ = respBody.Close()
		}
	}()

	d := json.NewDecoder(respBody)
	d.UseNumber()
	var response interface{}
	err := d.Decode(&response)
	if err != nil {
		switch {
		case err == io.EOF:
			return nil, nil
		default:
			return nil, err
		}
	}

	return response, nil
}
