package salesforce

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"
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
	//log.Infof("Response: %s", string(body))
	apiErrors := &ApiErrors{}
	if err := json.Unmarshal(body, apiErrors); err == nil {
		// Check if api error is valid
		if apiErrors.Validate() && res.StatusCode != 200 {
			if apiErrors.InvalidTokenErorr() {
				log.Debug("Access token expired, do token refresh...")
				var err error
				if sscm.AuthType == "OAuth 2.0 JWT Bearer Flow" {
					if sscm.GenerateJWT {
						err = sscm.GenerateJSONWebToken(sscm.ClientId, sscm.Subject, sscm.JWTExpiry, sscm.ClientKey, log)
					}
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
					return nil, activity.NewError("API call failed: "+err.Error(), strconv.Itoa(res.StatusCode), err)
				}
			} else {
				log.Errorf("API call error: %s", apiErrors.Error())
				return nil, activity.NewError("API call failed: "+apiErrors.Error(), strconv.Itoa(res.StatusCode), apiErrors)
			}
		}
	}

	return body, nil
}

func RestCallForCSVResponse(sscm *sfconnection.SalesforceSharedConfigManager, method string, url string, reqBytes []byte, log log.Logger) (map[string]interface{}, error) {

	var reqBody io.Reader
	respMap := make(map[string]interface{})
	if reqBytes != nil {
		reqBody = bytes.NewBuffer(reqBytes)
	} else {
		reqBody = nil
	}
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Add("Authorization", "Bearer "+sscm.SalesforceToken.AccessToken)
	req.Header.Add("Accept", "text/csv")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	code := res.StatusCode
	if code != 200 && code != 204 {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
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
					respMap, err = RestCallForCSVResponse(sscm, method, url, reqBytes, log)
					if err != nil {
						log.Errorf("API call error: %s", err.Error())
						return nil, activity.NewError("API call failed: "+err.Error(), strconv.Itoa(res.StatusCode), err)
					}
				} else {
					log.Errorf("API call error: %s", apiErrors.Error())
					return nil, activity.NewError("API call failed: "+apiErrors.Error(), strconv.Itoa(res.StatusCode), apiErrors)
				}
			}
		}
	} else {

		defer res.Body.Close()
		reader := csv.NewReader(res.Body)
		content, _ := reader.ReadAll()
		if content == nil {
			log.Info("No record fetched")
			return nil, nil
		}
		headersArr := make([]string, 0)
		for _, headE := range content[0] {
			headersArr = append(headersArr, headE)
		}

		//Remove the header row
		content = content[1:]

		recordsMap := make([]map[string]interface{}, 0)
		for _, d := range content {
			recordMap := make(map[string]interface{})
			for j, y := range d {
				recordMap[headersArr[j]] = y
			}
			recordsMap = append(recordsMap, recordMap)
		}

		maxRecords, _ := strconv.Atoi(strings.Join(res.Header["Sforce-Numberofrecords"], ""))
		respMap["records"] = recordsMap
		respMap["maxRecords"] = maxRecords
		respMap["locator"] = strings.Join(res.Header["Sforce-Locator"], "")

	}

	return respMap, nil
}
