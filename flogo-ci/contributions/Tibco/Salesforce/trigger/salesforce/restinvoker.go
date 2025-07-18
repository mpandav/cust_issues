package salesforce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/project-flogo/core/support/log"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

func RestCall(sscm *sfconnection.SalesforceSharedConfigManager, method string, url string, reqBytes []byte, log log.Logger) ([]byte, error) {
	var reqBody io.Reader

	if reqBytes != nil {
		reqBody = bytes.NewBuffer(reqBytes)

	} else {
		reqBody = nil
	}

	log.Debugf("Request Url [%s] with method [%s]", url, method)

	req, _ := http.NewRequest(method, url, reqBody)

	req.Header.Add("Authorization", "Bearer "+sscm.SalesforceToken.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debugf("Response status code %s", strconv.Itoa(res.StatusCode))
	apiErrors := &ApiErrors{}
	if err := json.Unmarshal(body, apiErrors); err == nil {
		// Check if api error is valid
		if apiErrors.Validate() {
			if apiErrors.InvalidTokenErorr() {
				log.Debug("Access token expired, do token refresh...")
				var err error
				if sscm.AuthType == "OAuth 2.0 JWT Bearer Flow" {
					err = sscm.DoRefreshTokenUsingJWT(log)
				} else {
					err = sscm.DoRefreshToken(log)
				}
				if err != nil {
					log.Errorf("Token refresh error [%s]", err.Error())
					return nil, fmt.Errorf("Token refresh error [%s]", err.Error())
				}
				//Send Request again.
				body, err = RestCall(sscm, method, url, reqBytes, log)
				if err != nil {
					log.Errorf("API call error: %s", err.Error())
					return nil, fmt.Errorf("API call failure %s", err.Error())
				}
			} else {
				log.Errorf("API call error: %s", apiErrors.Error())
				return nil, fmt.Errorf("API call failure %s", err.Error())
			}
		}
	}

	return body, nil
}
