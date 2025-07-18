package connection

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//Response represents the response received from token endpoint while generating new access token
type Response struct {
	ExpiresIn             int    `json:"expires_in"`
	AccessToken           string `json:"access_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	RefreshToken          string `json:"refresh_token"`
	Data                  string `json:"data"`
	Message               string `json:"message"`
	Code                  string `json:"code"`
	Success               bool   `json:"success"`
	Error                 string `json:"error"`
}

func getAccessTokenFromRefreshToken(snowFlakeConn *Settings) error {

	tokenEndpoint := "https://" + snowFlakeConn.Account + ".snowflakecomputing.com/oauth/token-request"
	authorizationHeaderValue := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", snowFlakeConn.ClientID, snowFlakeConn.ClientSecret))))
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", snowFlakeConn.RefreshToken)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authorizationHeaderValue)
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(req)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result := Response{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		if !result.Success {
			if result.Error == "invalid_grant" {
				errString := "Error while generating access token. The refresh token is not valid. Please provide new 'Authorization Code'."
				return errors.New(errString)
			} else if result.Error == "invalid_client" {
				errString := "Error while generating access token. The 'Client ID' or 'Client Secret' is incorrect."
				return errors.New(errString)
			} else {
				return errors.New(result.Message)
			}
		}
	}

	snowFlakeConn.AccessToken = result.AccessToken
	snowFlakeConn.AccessTokenExpiry = (time.Now().Unix() + int64(result.ExpiresIn)) * 1000
	return nil
}

func getAccessTokenFromAuthCode(snowFlakeConn *Settings) error {

	tokenEndpoint := "https://" + snowFlakeConn.Account + ".snowflakecomputing.com/oauth/token-request"
	authorizationHeaderValue := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", snowFlakeConn.ClientID, snowFlakeConn.ClientSecret))))
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", snowFlakeConn.AuthCode)
	data.Set("redirect_uri", snowFlakeConn.RedirectURI)

	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", authorizationHeaderValue)
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(req)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result := Response{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		if !result.Success {
			if result.Error == "invalid_grant" {
				errString := "Error while generating access token. The given 'Authorization Code' is not valid."
				return errors.New(errString)
			} else if result.Error == "invalid_client" {
				errString := "Error while generating access token. The 'Client ID' or 'Client Secret' is incorrect."
				return errors.New(errString)
			} else if result.Error == "invalid_request" {
				errString := "Error while generating access token. Please check if 'Redirect URI' is correct."
				return errors.New(errString)
			} else {
				return errors.New(result.Message)
			}
		}
	}

	snowFlakeConn.AccessToken = result.AccessToken
	snowFlakeConn.AccessTokenExpiry = (time.Now().Unix() + int64(result.ExpiresIn)) * 1000
	snowFlakeConn.RefreshToken = result.RefreshToken
	snowFlakeConn.RefreshTokenExpiry = (time.Now().Unix() + int64(result.RefreshTokenExpiresIn)*1000)
	return nil
}

func getOktaAccessTokenFromAuthCode(snowFlakeConn *Settings) error {

	//tokenEndpoint := "https://" + snowFlakeConn.Account + ".snowflakecomputing.com/oauth/token-request"
	// authorizationHeaderValue := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", snowFlakeConn.ClientID, snowFlakeConn.ClientSecret))))
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", snowFlakeConn.AuthCode)
	data.Set("redirect_uri", snowFlakeConn.RedirectURI)
	data.Set("client_id", snowFlakeConn.ClientID)
	data.Set("scope", snowFlakeConn.Scope)
	data.Set("code_verifier", snowFlakeConn.OktaCodeVerifier)
	req, err := http.NewRequest("POST", snowFlakeConn.OktaTokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Add("Authorization", authorizationHeaderValue)
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(req)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result := Response{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		if !result.Success {
			if result.Error == "invalid_grant" {
				errString := "Error while generating access token. The given 'Authorization Code' is not valid."
				return errors.New(errString)
			} else if result.Error == "invalid_client" {
				errString := "Error while generating access token. The 'Client ID' or 'Client Secret' is incorrect."
				return errors.New(errString)
			} else if result.Error == "invalid_request" {
				errString := "Error while generating access token. Please check if 'Redirect URI' is correct."
				return errors.New(errString)
			} else {
				return errors.New(result.Message)
			}
		}
	}

	snowFlakeConn.OktaAccessToken = result.AccessToken
	snowFlakeConn.OktaAccessTokenExpiry = (time.Now().Unix() + int64(result.ExpiresIn)) * 1000
	snowFlakeConn.OktaRefreshToken = result.RefreshToken
	//snowFlakeConn.RefreshTokenExpiry = (time.Now().Unix() + int64(result.RefreshTokenExpiresIn)*1000)
	return nil
}

func getOktaAccessTokenFromRefreshToken(snowFlakeConn *Settings) error {

	//tokenEndpoint := "https://" + snowFlakeConn.Account + ".snowflakecomputing.com/oauth/token-request"
	//authorizationHeaderValue := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", snowFlakeConn.ClientID, snowFlakeConn.ClientSecret))))
	client := &http.Client{}
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", snowFlakeConn.OktaRefreshToken)
	data.Set("client_id", snowFlakeConn.ClientID)
	data.Set("redirect_uri", snowFlakeConn.RedirectURI)
	data.Set("code_verifier", snowFlakeConn.OktaCodeVerifier)
	data.Set("scope", snowFlakeConn.Scope)
	req, err := http.NewRequest("POST", snowFlakeConn.OktaTokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	//req.Header.Add("Authorization", authorizationHeaderValue)
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, _ := client.Do(req)
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result := Response{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		if !result.Success {
			if result.Error == "invalid_grant" {
				errString := "Error while generating access token. The refresh token is not valid. Please provide new 'Authorization Code'."
				return errors.New(errString)
			} else if result.Error == "invalid_client" {
				errString := "Error while generating access token. The 'Client ID' or 'Client Secret' is incorrect."
				return errors.New(errString)
			} else {
				return errors.New(result.Message)
			}
		}
	}

	snowFlakeConn.OktaAccessToken = result.AccessToken
	snowFlakeConn.OktaAccessTokenExpiry = (time.Now().Unix() + int64(result.ExpiresIn)) * 1000
	snowFlakeConn.OktaRefreshToken = result.RefreshToken
	return nil
}