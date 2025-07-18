package authorization

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/project-flogo/core/support/trace"
	"github.com/tibco/wi-contrib/environment"

	"github.com/project-flogo/core/data/coerce"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

var logger = log.ChildLogger(log.RootLogger(), "general.connection.authorization")
var factory = &authorizatrionFactory{}

// TODO  this should move to shared code of REST, same as get Body etc
const (
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodPATCH  = "PATCH"
	MethodDELETE = "DELETE"

	OAuth2      = "OAuth2"
	Basic       = "Basic"
	BearerToken = "Bearer Token"
)

type AuthorizationConnection struct {
	Name        string `md:"name"`
	Type        string `md:"type"`
	UserName    string `md:"userName"`
	Password    string `md:"password"`
	BearerToken string `md:"token"`

	GrandType            string `md:"grantType"`
	CallbackUrl          string `md:"callbackURL"`
	AuthUrl              string `md:"authURL"`
	AccessTokenUrl       string `md:"accessTokenURL"`
	ClientId             string `md:"clientId"`
	ClientSecret         string `md:"clientSecret"`
	Scope                string `md:"scope"`
	AuthQueryParameters  string `md:"authQueryParameters"`
	ClientAuthentication string `md:"clientAuthentication"`
	Audience             string `md:"audience"`
	//Method               string `md:"method"`
	TokenInfo string `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
	Token     *authrizationToken
	lock      sync.Mutex
}

type authrizationToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	IssuedAt     string `json:"issued_at"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (a *authrizationToken) UnmarshalJSON(data []byte) error {
	ser := &struct {
		AccessToken  string      `json:"access_token"`
		TokenType    string      `json:"token_type"`
		ExpiresIn    interface{} `json:"expires_in"`
		IssuedAt     interface{} `json:"issued_at"`
		RefreshToken string      `json:"refresh_token"`
		Scope        string      `json:"scope"`
	}{}

	if err := json.Unmarshal(data, ser); err != nil {
		return err
	}
	a.AccessToken = ser.AccessToken
	a.TokenType = ser.TokenType
	a.RefreshToken = ser.RefreshToken
	a.Scope = ser.Scope
	a.ExpiresIn, _ = coerce.ToString(ser.ExpiresIn)
	a.IssuedAt, _ = coerce.ToString(ser.IssuedAt)
	return nil
}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type authorizatrionFactory struct {
}

func (*authorizatrionFactory) Type() string {
	return "aws"
}

func (*authorizatrionFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	awsManger := &AuthorizationManager{}
	config, err := getConnectionConfig(settings)
	if err != nil {
		return nil, err
	}
	awsManger.conn = config
	if config.Token != nil {
		if len(config.Token.ExpiresIn) > 0 {
			logger.Infof("access token expire in [%s] for connection [%s]", config.Token.ExpiresIn, config.Name)
		} else if len(config.Token.IssuedAt) > 0 {
			logger.Infof("access token issued at [%s] for connection [%s]", config.Token.IssuedAt, config.Name)
		}
	}
	return awsManger, nil
}

type AuthorizationManager struct {
	conn *AuthorizationConnection
}

func (a *AuthorizationManager) Type() string {
	return "aws"
}

func (a *AuthorizationManager) GetConnection() interface{} {
	return a.conn
}

func (a *AuthorizationManager) ReleaseConnection(connection interface{}) {
	//No nothing for aws connection
}

func (conn *AuthorizationConnection) SendRequest(client *http.Client, method, uri string, header http.Header, body interface{}, requestType string, host string, asrEnable bool, traceCtx trace.TracingContext, log log.Logger) (*http.Response, error) {
	var reqBody io.Reader
	var multipartHeader string
	if requestType == "multipart/form-data" {
		bodyData := &bytes.Buffer{}
		writer := multipart.NewWriter(bodyData)

		for k, v := range body.(map[string]interface{}) {

			switch v.(type) {
			case string:
				_ = writer.WriteField(k, v.(string))
				break
			case map[string]interface{}:
				b := new(bytes.Buffer)
				for key, value := range v.(map[string]interface{}) {
					fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
				}
				_ = writer.WriteField(k, b.String())
				break
			case [][]uint8:
				for _, fileData := range v.([][]uint8) {
					part, err := writer.CreateFormFile(k, "tempFile")
					reader := bytes.NewReader([]byte(string(fileData)))
					if err != nil {
						return nil, err
					}
					_, err = io.Copy(part, reader)
				}
				break
			default:
				for _, fileData := range v.([]interface{}) {
					part, err := writer.CreateFormFile(k, "tempFile")
					reader := bytes.NewReader([]byte(fileData.(string)))
					if err != nil {
						return nil, err
					}
					_, err = io.Copy(part, reader)
				}

			}
		}

		err := writer.Close()
		if err != nil {
			return nil, err
		}

		reqBody = bodyData
		multipartHeader = writer.FormDataContentType()
	} else {
		var err error
		reqBody, err = GetRequestBody(method, body, requestType, asrEnable, log)
		if err != nil {
			return nil, err
		}
	}

	req, _ := http.NewRequest(method, uri, reqBody)
	if requestType == "multipart/form-data" {
		req.Header.Set("Content-Type", multipartHeader)
	}
	if host != "" {
		log.Infof("Overriding Host With: %s", host)
		req.URL.Host = host
		req.Host = host
	}
	req.Header = header
	if asrEnable {
		req.Header.Set("X-ATMOSPHERE-for-USER", environment.GetTCISubscriptionUName())
	}

	if traceCtx != nil {
		_ = trace.GetTracer().Inject(traceCtx, trace.HTTPHeaders, req)
	}

	log.Debugf("Request Headers: %+v", req.Header)
	if conn.Type == Basic {
		req.Header.Set("Authorization", "Basic "+basicAuth(conn.UserName, conn.Password))
	} else if conn.Type == OAuth2 {
		req.Header.Set("Authorization", "Bearer "+conn.Token.AccessToken)
	} else if conn.Type == BearerToken {
		req.Header.Set("Authorization", "Bearer "+conn.BearerToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		if refreshTokenRequired(conn.Type, 0, err.Error()) {
			logger.Infof("session timeout or access token is invalid, trying to refresh token [%s]", conn.Name)
			err = conn.RefreshToken()
			if err != nil {
				return nil, err
			}
			logger.Infof("token refreshed for connection [%s]", conn.Name)
			return conn.SendRequest(client, method, uri, header, body, requestType, host, asrEnable, traceCtx, log)
		}
		return resp, err
	}

	if refreshTokenRequired(conn.Type, resp.StatusCode, "") {
		logger.Infof("session timeout or access token is invalid, trying to refresh token [%s]", conn.Name)
		err = conn.RefreshToken()
		if err != nil {
			return nil, err
		}
		logger.Infof("token refreshed for connection [%s]", conn.Name)
		return conn.SendRequest(client, method, uri, header, body, requestType, host, asrEnable, traceCtx, log)
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
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading refresh token response bytes: %v", err)
	}

	if resp.StatusCode == 200 {
		newToken := &authrizationToken{}
		if err := json.Unmarshal(respBytes, newToken); err != nil {
			return fmt.Errorf("unable to unmarshal authrizationToken response: %v", err)
		}

		if len(newToken.AccessToken) > 0 {
			conn.Token.AccessToken = newToken.AccessToken
		}
		if len(newToken.ExpiresIn) > 0 {
			conn.Token.ExpiresIn = newToken.ExpiresIn
		}

		if len(newToken.RefreshToken) > 0 {
			conn.Token.RefreshToken = newToken.RefreshToken
		}

		if len(newToken.TokenType) > 0 {
			conn.Token.TokenType = newToken.TokenType
		}
		if conn.Token != nil && len(conn.Token.ExpiresIn) > 0 {
			logger.Infof("access token expire in [%s] for connection [%s]", conn.Token.ExpiresIn, conn.Name)
		} else if len(conn.Token.IssuedAt) > 0 {
			logger.Infof("access token issued at [%s] for connection [%s]", conn.Token.IssuedAt, conn.Name)
		}

	} else {
		return fmt.Errorf("refresh token status code not in 200 [%s]", string(respBytes))
	}
	return nil
}

func getConnectionConfig(settings map[string]interface{}) (*AuthorizationConnection, error) {
	s := &AuthorizationConnection{lock: sync.Mutex{}}
	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}

	tokenValue := s.TokenInfo

	if s.Type == OAuth2 && len(tokenValue) > 0 {

		authToken := &authrizationToken{}
		if strings.HasPrefix(tokenValue, "{") {
			err = json.Unmarshal([]byte(tokenValue), authToken)
			if err != nil {
				logger.Errorf("Error occured while unmarshalling token ", err)
				return nil, err
			}
		} else {
			logger.Debug("Decoding HTTP Client Authorization connection credentials")
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
		s.Token = authToken
		if s.Token.Scope != "" && s.GrandType == "Client Credentials" && !validScope(s.Token.Scope, s.Scope) {
			logger.Errorf("Configured scope [%s] does not match with fetched token scope [%s] for connection [%s]. New token requested.", s.Scope, s.Token.Scope, s.Name)
			// Requesting new token with new scope
			err := s.RefreshToken()
			if err != nil {
				return nil, err
			}
			logger.Infof("Token successfully refreshed with new scope for connection [%s]", s.Name)
		}
	}

	/*if s.Method == "" {
		s.Method = "GET"
	}*/
	return s, nil
}

func validScope(tokenScope string, configuredScope string) bool {
	if len(tokenScope) == 0 {
		return false
	}

	configuredScopes := strings.Split(configuredScope, " ")
	for _, scope := range configuredScopes {
		if !strings.Contains(tokenScope, scope) {
			return false
		}
	}
	return true
}

func (conn *AuthorizationConnection) createRefreshTokenRequest() (*http.Request, error) {

	var request *http.Request
	var err error
	//No nothing for aws connection
	if conn.GrandType == "Client Credentials" {
		if conn.Token != nil {
			request, err = conn.createRequest("client_credentials", conn.Token.RefreshToken)
		} else {
			request, err = conn.createRequest("client_credentials", "")
		}
	} else {
		if conn.Token != nil && conn.Token.RefreshToken != "" {
			request, err = conn.createRequest("refresh_token", conn.Token.RefreshToken)
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
			"client_id":     {conn.ClientId},
			"client_secret": {conn.ClientSecret},
		}
		if len(refreshToken) > 0 {
			queryParam["refresh_token"] = []string{refreshToken}
		} else {
			if len(conn.Scope) > 0 {
				// Set initial scope
				queryParam.Set("scope", conn.Scope)
			}
			if len(conn.Audience) > 0 {
				queryParam.Set("audience", conn.Audience)
			}
		}

		return http.NewRequest("GET", conn.AccessTokenUrl+"?"+queryParam.Encode(), nil)
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
			if len(conn.Audience) > 0 {
				queryParam.Set("audience", conn.Audience)
			}
		}
		body := strings.NewReader(queryParam.Encode())
		req, err := http.NewRequest("POST", conn.AccessTokenUrl, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(conn.ClientId, conn.ClientSecret)
		return req, nil
	} else if conn.ClientAuthentication == "Body" {
		queryParam := url.Values{
			"grant_type":    {grantType},
			"client_id":     {conn.ClientId},
			"client_secret": {conn.ClientSecret},
		}
		if len(refreshToken) > 0 {
			queryParam["refresh_token"] = []string{refreshToken}
		} else {
			if len(conn.Scope) > 0 {
				// Set initial scope
				queryParam.Set("scope", conn.Scope)
			}
			if len(conn.Audience) > 0 {
				queryParam.Set("audience", conn.Audience)
			}
		}
		body := strings.NewReader(queryParam.Encode())
		req, err := http.NewRequest("POST", conn.AccessTokenUrl, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
		return req, nil
	}
	return nil, nil
}

func GetRequestBody(method string, body interface{}, requestType string, asrEnabled bool, log log.Logger) (reqBody io.Reader, err error) {
	if method == MethodPOST || method == MethodPUT || method == MethodPATCH || method == MethodDELETE {
		value := body
		if value != nil {
			if requestType == "application/json" {
				reqBody, err = getBody(value, log)
			} else if requestType == "application/x-www-form-urlencoded" {
				var encodeBody = ""
				switch ty := value.(type) {
				case []interface{}:
					encodeBody, err = encodedUrlBodyFromArray(ty)
					if err != nil {
						return nil, err
					}
				case map[string]interface{}:
					encodeBody = encodedUrlBodyFromObject(ty)
				case string:
					//To array,
					a, err := coerce.ToArray(value)
					if err != nil {
						// Try to convert to map
						obj, err := coerce.ToObject(value)
						if err != nil {
							return nil, fmt.Errorf("unexpected application/x-www-form-urlencoded body [%+v], please refer to documentation", value)
						} else {
							encodeBody = encodedUrlBodyFromObject(obj)
						}
					} else {
						encodeBody, err = encodedUrlBodyFromArray(a)
						if err != nil {
							return nil, err
						}
					}
				}

				reqBody, err = getBody(encodeBody, log)
			} else if requestType == "text/plain" {
				bodyMap, err := coerce.ToObject(value)
				if err != nil {
					return nil, fmt.Errorf("unknow body strucure, %s", err.Error())
				}
				reqBody, err = getBody(bodyMap["data"], log)
			} else if requestType == "multipart/form-data" {
				if !asrEnabled {
					return nil, fmt.Errorf("unsupport request type %s", requestType)
				}

				files, err := coerce.ToParams(value)
				if err != nil {
					return nil, fmt.Errorf("error to convert file parts to base64 string: %s", err.Error())
				}
				payload := &bytes.Buffer{}
				writer := multipart.NewWriter(payload)
				for k, v := range files {
					p, err := writer.CreateFormField(k)
					if err != nil {
						return nil, fmt.Errorf("error creating multipart [%s]: %s", k, err.Error())
					}
					bts, err := coerce.ToBytes(v)
					if err != nil {
						return nil, fmt.Errorf("decodeBase64 file content to bytes error: %s", err.Error())
					}
					dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(bts)))
					n, err := base64.StdEncoding.Decode(dbuf, bts)
					if _, err := p.Write(dbuf[:n]); err != nil {
						return nil, fmt.Errorf("error writing [%s] file content to part: %s", k, err.Error())
					}
				}
				err = writer.Close()
				if err != nil {
					log.Warnf("Closing multi-part writer error: %s", err.Error())
				}
				reqBody = payload

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

func encodedUrlBodyFromArray(bodyArray []interface{}) (string, error) {
	values := url.Values{}
	for _, v := range bodyArray {
		nvMap, err := coerce.ToParams(v)
		if err != nil {
			return "", fmt.Errorf("form-urlencoded body must be name-value string paire")
		}

		name, err := url.QueryUnescape(nvMap["name"])
		if err != nil {
			name = nvMap["name"]
		}

		value, err := url.QueryUnescape(nvMap["value"])
		if err != nil {
			value = nvMap["value"]
		}
		values.Add(name, value)
	}
	return values.Encode(), nil
}

func encodedUrlBodyFromObject(body map[string]interface{}) string {
	values := url.Values{}
	for k, v := range body {
		name, err := url.QueryUnescape(k)
		if err != nil {
			name = k
		}

		val, _ := coerce.ToString(v)
		value, err := url.QueryUnescape(val)
		if err != nil {
			value = val
		}
		values.Add(name, value)
	}
	return values.Encode()
}
