package odata

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

var logger = log.ChildLogger(log.RootLogger(), "odata.connector.connection")

var factory = &connectionFactory{}

const (
	OAuth2 = "OAuth2"
	Basic  = "Basic"
)

type AuthorizationConnection struct {
	Name                 string `md:"name"`
	Type                 string `md:"type"`
	RootURL              string `md:"rootURL"`
	Username             string `md:"userName"`
	Password             string `md:"password"`
	GrantType            string `md:"grantType"`
	AccessTokenURL       string `md:"accessTokenURL"`
	ClientID             string `md:"clientId"`
	ClientSecret         string `md:"clientSecret"`
	Scope                string `md:"scope"`
	ClientAuthentication string `md:"clientAuthentication"`
	TokenInfo            string `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
	AuthToken            *authorizationToken
	lock                 sync.Mutex
}

type authorizationToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	IssuedAt     string `json:"issued_at"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (t *authorizationToken) UnmarshalJSON(data []byte) error {
	ser := &struct {
		AccessToken  string      `json:"access_token"`
		TokenType    string      `json:"token_type"`
		ExpiresIn    interface{} `json:"expires_in"`
		IssuedAt     interface{} `json:"issued_at"`
		RefreshToken string      `json:"refresh_token"`
		Scope        string      `json:"scope"`
	}{}

	err := json.Unmarshal(data, &ser)
	if err != nil {
		return err
	}
	t.AccessToken = ser.AccessToken
	t.TokenType = ser.TokenType
	t.ExpiresIn, _ = coerce.ToString(ser.ExpiresIn)
	t.IssuedAt, _ = coerce.ToString(ser.IssuedAt)
	t.RefreshToken = ser.RefreshToken
	t.Scope = ser.Scope
	return nil
}
func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type connectionFactory struct {
}

func (connectionFactory) Type() string {
	return "OData"
}

func (connectionFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	authManager := &AuthorizationManager{}
	config, err := getConnectionConfig(settings)
	if err != nil {
		return nil, err
	}
	authManager.conn = config
	return authManager, nil

}

type AuthorizationManager struct {
	conn *AuthorizationConnection
}

func (a *AuthorizationManager) Type() string {
	return "OData"
}

func (a *AuthorizationManager) GetConnection() interface{} {
	return a.conn
}

func (a *AuthorizationManager) ReleaseConnection(connection interface{}) {
}

func (conn *AuthorizationConnection) SendRequest(client *http.Client, method, uri string, header http.Header, body interface{}, requestType string, log log.Logger) (*http.Response, error) {
	var reqBody io.Reader

	var err error
	reqBody, err = GetRequestBody(method, body, requestType, log)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest(method, uri, reqBody)

	req.Header = header

	log.Debugf("Request Headers: %+v", req.Header)
	if conn.Type == Basic {
		req.Header.Set("Authorization", "Basic "+basicAuth(conn.Username, conn.Password))
	} else if conn.Type == OAuth2 {
		req.Header.Set("Authorization", "Bearer "+conn.AuthToken.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		if refreshTokenRequired(conn.Type, 0, err.Error()) {
			logger.Infof("session timeout or access token is invalid, trying to refresh token for connection [%s]", conn.Name)
			err = conn.RefreshToken()
			if err != nil {
				return nil, err
			}
			logger.Infof("token refreshed for connection [%s]", conn.Name)
			return conn.SendRequest(client, method, uri, header, body, requestType, log)
		}
		return resp, err
	}

	if refreshTokenRequired(conn.Type, resp.StatusCode, "") {
		logger.Infof("session timeout or access token is invalid, trying to refresh token for connection [%s]", conn.Name)
		err = conn.RefreshToken()
		if err != nil {
			return nil, err
		}
		logger.Infof("token refreshed for connection [%s]", conn.Name)
		return conn.SendRequest(client, method, uri, header, body, requestType, log)
	}
	return resp, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func refreshTokenRequired(connType string, statusCode int, errMsg string) bool {
	// Invalid token
	// code with 401
	// Session expired
	// INVALID_SESSION_ID
	// No refresh token required for basic
	if connType == OAuth2 {
		errMsg = strings.ToLower(errMsg)
		if strings.Contains(errMsg, "invalid") && strings.Contains(errMsg, "token") {
			return true
		}
		if strings.Contains(errMsg, "session") && strings.Contains(errMsg, "expired") {
			return true
		}
		if strings.Contains(errMsg, "invalid_session_Id") {
			return true
		}

		if statusCode == 401 {
			return true
		}
	}
	return false
}

func (conn *AuthorizationConnection) RefreshToken() error {
	conn.lock.Lock()
	defer conn.lock.Unlock()
	req, err := conn.createRefreshTokenRequest()
	if err != nil {
		return fmt.Errorf("error refreshing token for connection [%s]: %s", conn.Name, err.Error())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending refresh token request: %v", err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading refresh token response bytes: %v", err)
	}

	if resp.StatusCode == 200 {
		newToken := &authorizationToken{}
		if err := json.Unmarshal(respBytes, newToken); err != nil {
			return fmt.Errorf("unable to unmarshal authrizationToken response: %v", err)
		}

		if len(newToken.AccessToken) > 0 {
			conn.AuthToken.AccessToken = newToken.AccessToken
		}
		if len(newToken.ExpiresIn) > 0 {
			conn.AuthToken.ExpiresIn = newToken.ExpiresIn
		}

		if len(newToken.RefreshToken) > 0 {
			conn.AuthToken.RefreshToken = newToken.RefreshToken
		}

		if len(newToken.TokenType) > 0 {
			conn.AuthToken.TokenType = newToken.TokenType
		}

	} else {
		return fmt.Errorf("refresh token status code not in 200 [%s]", string(respBytes))
	}
	return nil
}

func (conn *AuthorizationConnection) createRefreshTokenRequest() (*http.Request, error) {

	var request *http.Request
	var err error
	if conn.GrantType == "Client Credentials" {
		if conn.AuthToken != nil {
			request, err = conn.createRequest("client_credentials", conn.AuthToken.RefreshToken)
		} else {
			request, err = conn.createRequest("client_credentials", "")
		}
	} else {
		if conn.AuthToken != nil && conn.AuthToken.RefreshToken != "" {
			request, err = conn.createRequest("refresh_token", conn.AuthToken.RefreshToken)
		} else {
			request, err = conn.createRequest("refresh_token", "")
		}
	}

	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "Flogo authorization connection")
	return request, nil
}

func (conn *AuthorizationConnection) createRequest(grantType, refreshToken string) (*http.Request, error) {
	if conn.ClientAuthentication == "Query" {
		queryParam := url.Values{
			"grant_type":    {grantType},
			"client_id":     {conn.ClientID},
			"client_secret": {conn.ClientSecret},
		}
		if len(refreshToken) > 0 {
			queryParam["refresh_token"] = []string{refreshToken}
		} else {
			if len(conn.Scope) > 0 {
				// Set initial scope
				queryParam.Set("scope", conn.Scope)
			}
		}

		return http.NewRequest("GET", conn.AccessTokenURL+"?"+queryParam.Encode(), nil)
	} else if conn.ClientAuthentication == "Header" {
		queryParam := url.Values{
			"grant_type": {grantType},
		}
		if len(refreshToken) > 0 {
			queryParam["refresh_token"] = []string{refreshToken}
		} else {
			if len(conn.Scope) > 0 {
				// Set initial scope
				queryParam.Set("scope", conn.Scope)
			}
		}
		body := strings.NewReader(queryParam.Encode())
		req, err := http.NewRequest("POST", conn.AccessTokenURL, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(conn.ClientID, conn.ClientSecret)
		return req, nil
	} else if conn.ClientAuthentication == "Body" {
		queryParam := url.Values{
			"grant_type":    {grantType},
			"client_id":     {conn.ClientID},
			"client_secret": {conn.ClientSecret},
		}
		if len(refreshToken) > 0 {
			queryParam["refresh_token"] = []string{refreshToken}
		} else {
			if len(conn.Scope) > 0 {
				// Set initial scope
				queryParam.Set("scope", conn.Scope)
			}
		}
		body := strings.NewReader(queryParam.Encode())
		req, err := http.NewRequest("POST", conn.AccessTokenURL, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		return req, nil
	}
	return nil, nil
}

func getConnectionConfig(settings map[string]interface{}) (*AuthorizationConnection, error) {
	s := &AuthorizationConnection{lock: sync.Mutex{}}
	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}

	tokenValue := s.TokenInfo

	if s.Type == "OAuth2" {

		authToken := &authorizationToken{}
		if strings.HasPrefix(tokenValue, "{") {
			err = json.Unmarshal([]byte(tokenValue), authToken)
			if err != nil {
				logger.Errorf("Error occured while unmarshalling token ", err)
				return nil, err
			}
		} else {
			logger.Debug("Decoding OData connection credentials")
			decodedTokenByteData, err := base64.StdEncoding.DecodeString(tokenValue)
			if err != nil {
				logger.Errorf("Error occured while decoding token ", err)
				return nil, err
			}
			err = json.Unmarshal(decodedTokenByteData, authToken)
			if err != nil {
				logger.Errorf("Error occured while unmarshalling token ", err)
				return nil, err
			}
		}
		s.AuthToken = authToken
	}

	return s, nil
}
func GetRequestBody(method string, body interface{}, requestType string, log log.Logger) (reqBody io.Reader, err error) {
	if method == http.MethodPost || method == http.MethodPatch || method == http.MethodDelete {
		value := body
		if value != nil {
			if requestType == "application/json" {
				reqBody, err = getBody(value, log)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("unsupport request type %s", requestType)
			}
		}
	}
	return reqBody, nil
}

func getBody(content interface{}, log log.Logger) (io.Reader, error) {
	var reqBody io.Reader
	switch content.(type) {
	case string:
		log.Debugf("Request Body [%s]", content.(string))
		reqBody = bytes.NewBuffer([]byte(content.(string)))
	default:
		b, err := json.Marshal(content) //todo handle error
		if err != nil {
			return nil, err
		}
		log.Debugf("Request Body [%s]", string(b))
		reqBody = bytes.NewBuffer(b)
	}
	return reqBody, nil
}
