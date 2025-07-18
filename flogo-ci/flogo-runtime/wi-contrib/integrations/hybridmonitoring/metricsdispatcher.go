package hybridmonitoring

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tibco/wi-contrib/integrations/hybridmonitoring/types"
)

var client = &http.Client{Timeout: 1 * time.Minute}

func registerApp(regis *types.RegistratrationRequest) (string, error) {

	bodyBytes, _ := json.Marshal(regis)

	agentUrl := "http://" + agentConfig.HybridAgentHost + ":" + agentConfig.HybridAgentPort + types.AGENT_REGISTRATION
	request, err := http.NewRequest(http.MethodPost, agentUrl, bytes.NewBufferString(string(bodyBytes)))
	if err != nil {
		errMsg := fmt.Sprintf("Create new http POST request failed with error: %s", err.Error())
		return "", errors.New(errMsg)
	}

	resp, err := client.Do(request)
	if err != nil {
		errMsg := fmt.Sprintf("Send '%+v' request to url: '%s' failed with error: %s", "Registration", agentUrl, err.Error())
		return "", errors.New(errMsg)
	}

	if resp == nil {
		errMsg := fmt.Sprintf("Failed to send '%+v' request to url: '%s'", "Registration", agentUrl)
		return "", errors.New(errMsg)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to get response body: %s", err.Error())
			return "", errors.New(errMsg)
		}
		engResp := &types.RegistrationResponse{}
		err = json.Unmarshal(bytes, engResp)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to unmarshal response: %s", err.Error())
			return "", errors.New(errMsg)
		}

		statsLogger.Debugf("Pacemaker push app metrics to metrics server successfully. Event: %+v", "Registration")
		return engResp.Msg, nil
	}

	if resp.Header.Get("Content-Type") == "application/json" {
		errorResponse, err := GetErrorResponse(resp.Body)
		if err != nil {
			msg := fmt.Sprintf("Failed to push app metrics to metrics server. Event: %+v. Code: %d. Detail: %s", "Registration", resp.StatusCode, err.Error())
			return "", errors.New(msg)
		}
		return "", errors.New(errorResponse.ErrorMsg)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse response from metrics server. Event: %+v. Code: %d. Detail: %s", "Registration", resp.StatusCode, err.Error())
		return "", errors.New(errMsg)
	}
	errMsg := fmt.Sprintf("Failed to push app metrics to metrics server. Event: %+v. Code: %d. Detail: %s", "Registration", resp.StatusCode, string(respBytes))
	return "", errors.New(errMsg)

}

func dispatchAppEngineMetrics(appMetrics *types.AppEngineMetrics) (int, error) {

	bodyBytes, _ := json.Marshal(appMetrics)

	agentUrl := "http://" + agentConfig.HybridAgentHost + ":" + agentConfig.HybridAgentPort + types.AGENTPATH
	request, err := http.NewRequest(http.MethodPost, agentUrl, bytes.NewBufferString(string(bodyBytes)))
	if err != nil {
		errMsg := fmt.Sprintf("Create new http POST request failed with error: %s", err.Error())
		return 0, errors.New(errMsg)
	}

	resp, err := client.Do(request)
	if err != nil {
		errMsg := fmt.Sprintf("Send '%+v' request to url: '%s' failed with error: %s", appMetrics.EventType, agentUrl, err.Error())
		return 0, errors.New(errMsg)
	}

	if resp == nil {
		errMsg := fmt.Sprintf("Failed to send '%+v' request to url: '%s'", appMetrics.EventType, agentUrl)
		return 0, errors.New(errMsg)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to get response body: %s", err.Error())
			return 0, errors.New(errMsg)
		}

		engResp := &types.EngineUpdateResponse{}
		err = json.Unmarshal(bytes, engResp)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to unmarshal response: %s", err.Error())
			return 0, errors.New(errMsg)
		}

		statsLogger.Debugf("Pacemaker push app metrics to metrics server successfully. Event: %+v", appMetrics.EventType)
		statsLogger.Debugf("Getting next push interval: %+v", engResp.NextPushInterval)
		return engResp.NextPushInterval, nil
	}

	if resp.Header.Get("Content-Type") == "application/json" {
		errorResponse, err := GetErrorResponse(resp.Body)
		if err != nil {
			msg := fmt.Sprintf("Failed to push app metrics to metrics server. Event: %+v. Code: %d. Detail: %s", appMetrics.EventType, resp.StatusCode, err.Error())
			return 0, errors.New(msg)
		}
		return 0, errors.New(errorResponse.ErrorMsg)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)

	var response types.EngineUpdateResponse
	json.Unmarshal(respBytes, &response)
	fmt.Println(string(respBytes))
	fmt.Println(response.NextPushInterval)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse response from metrics server. Event: %+v. Code: %d. Detail: %s", appMetrics.EventType, resp.StatusCode, err.Error())
		return 0, errors.New(errMsg)
	}
	errMsg := fmt.Sprintf("Failed to push app metrics to metrics server. Event: %+v. Code: %d. Detail: %s", appMetrics.EventType, resp.StatusCode, string(respBytes))
	return response.NextPushInterval, errors.New(errMsg)

}

func GetErrorResponse(reader io.Reader) (*types.ErrorResponse, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	errorResponse := &types.ErrorResponse{}
	err = json.Unmarshal(bytes, errorResponse)
	if err != nil {
		return nil, err
	}
	return errorResponse, nil
}
