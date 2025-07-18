package docusign

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"

	jwt "github.com/golang-jwt/jwt/v4"
)

var logger = log.ChildLogger(log.RootLogger(), "connection.docusign")

type Settings struct {
	Name                           string `md:"name"`
	Description                    string `md:"description"`
	Environment                    bool   `md:"environment"`
	IntegratorKey                  string `md:"integratorKey"`
	SecretKey                      string `md:"secretKey"`
	AuthenticationType             string `md:"authenticationType"`
	UserID                         string `md:"userId"`
	RSAPrivateKey                  string `md:"privateKey"`
	WI_STUDIO_OAUTH_CONNECTOR_INFO string `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
}

type Token struct {
	ConnectionID string `json:"connection_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Env          string `json:"env"`
	Scope        string `json:"scope"`
	PrivateKey   string
	UserId       string
}

type RespError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type UserInfo struct {
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	AccountId   string `json:"account_id"`
	AccountName string `json:"account_name"`
	IsDefault   bool   `json:"is_default"`
	BaseURI     string `json:"base_uri"`
}

type DocusignSharedConfigManager struct {
	DocusignToken   *Token
	DocusignAccount *Account
	// DocusignHttpClient *http.Client
}

type JWTClaim struct {
	jwt.StandardClaims
	Scope string `json:"scope"`
}

var factory = &DocusignFactory{}

func init() {
	logger.Debug("Calling init()")
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type DocusignFactory struct {
}

func (*DocusignFactory) Type() string {
	return "Docusign"
}

func (*DocusignFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	s := &Settings{}
	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}

	// logger.Infof("Connection settings: %#v", s)

	acc_token := &Token{}
	dscm := &DocusignSharedConfigManager{}

	connName := s.Name
	tokenInfo := s.WI_STUDIO_OAUTH_CONNECTOR_INFO

	if tokenInfo == "" {
		return nil, fmt.Errorf("Token Information not found")
	}

	if connName != "" {
		if s.AuthenticationType == "Authorization Code Grant" || s.AuthenticationType == "" {
			if strings.HasPrefix(tokenInfo, "{") {
				err := json.Unmarshal([]byte(tokenInfo), acc_token)
				if err != nil {
					return nil, err
				}
			} else {
				decodedToken, err := base64.StdEncoding.DecodeString(tokenInfo)
				if err != nil {
					return nil, fmt.Errorf("unable to decode token error: %s", err.Error())
				}
				err = json.Unmarshal(decodedToken, acc_token)
				if err != nil {
					return nil, fmt.Errorf("error while unmarshalling token error: %s", err.Error())
				}
			}
			acc_token.ConnectionID = connName
			dscm.DocusignToken = acc_token
		} else if s.AuthenticationType == "JWT Grant" {
			err = dscm.getAccessTokenWithJWT(s)
			if err != nil {
				return nil, err
			}
		}

		// dscm.DocusignHttpClient = &http.Client{}

		// logger.Infof("Token: %#v", dscm.DocusignToken)

		err = dscm.GetAccountInfo()
		if err != nil {
			return nil, err
		}

	} else {
		return nil, fmt.Errorf("Connection Name cannot be empty")
	}

	return dscm, nil
}

func (dscm *DocusignSharedConfigManager) DoRefreshToken() error {

	logger.Info("Refreshing token")

	expiredToken := dscm.DocusignToken

	if expiredToken.RefreshToken == "" {

		// logger.Infof("In JWT refresh")

		settings := &Settings{}
		settings.Environment = getEnv(expiredToken.Env)
		settings.IntegratorKey = expiredToken.ClientID
		settings.Name = expiredToken.ConnectionID
		settings.UserID = expiredToken.UserId
		settings.RSAPrivateKey = expiredToken.PrivateKey

		// logger.Infof("Created settings: %#v", settings)

		err := dscm.getAccessTokenWithJWT(settings)
		if err != nil {
			return fmt.Errorf("error while refreshing JWT: %s", err.Error())
		}

		// logger.Infof("Refreshed token: %#v", dscm.DocusignToken)

	} else {

		// logger.Info("In authenticate code refresh")

		queryParam := url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {expiredToken.RefreshToken},
		}

		body := strings.NewReader(queryParam.Encode())

		request, err := http.NewRequest("POST", getAuthTokenUrl(expiredToken.Env), body)
		if err != nil {
			return fmt.Errorf("Error creating authenitcation request: %v", err)
		}
		request.Header.Set("User-Agent", "Web Integrator")
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(expiredToken.ClientID+":"+expiredToken.ClientSecret))
		request.Header.Set("Authorization", authHeader)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return fmt.Errorf("Error sending authentication request: %v", err)
		}
		defer resp.Body.Close()

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Error reading authentication response bytes: %v", err)
		}

		respErr := &RespError{}

		if err := json.Unmarshal(respBytes, respErr); err == nil {
			if respErr.Error != "" {
				return fmt.Errorf("Refresh token error %s %s", respErr.Error, respErr.Description)
			}
		}

		newToken := &Token{}
		if err := json.Unmarshal(respBytes, newToken); err != nil {
			return fmt.Errorf("Unable to unmarshal authentication response: %v", err)
		}

		if newToken.AccessToken == "" {
			return fmt.Errorf("Token refresh with invalid result: %v", newToken)
		}

		newToken.ConnectionID = expiredToken.ConnectionID
		newToken.ClientID = expiredToken.ClientID
		newToken.ClientSecret = expiredToken.ClientSecret
		newToken.Env = expiredToken.Env

		dscm.DocusignToken = newToken
	}
	return nil
}

func (dscm *DocusignSharedConfigManager) GetAccountInfo() error {
	logger.Info("Getting Account Information")

	requestUrl := "https://" + dscm.DocusignToken.Env + ".docusign.com/oauth/userinfo"
	method := http.MethodGet
	req, _ := http.NewRequest(method, requestUrl, nil)
	authHeader := "Bearer " + dscm.DocusignToken.AccessToken

	req.Header.Set("Authorization", authHeader)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to get account info: %s", err.Error())
	}

	defer res.Body.Close()

	respBody, err := ReadResponseBody(res)
	if err != nil || respBody == nil {
		return fmt.Errorf("Failed to get response body: %s", err.Error())
	}

	var userInfo = &UserInfo{}
	if res.StatusCode == http.StatusOK {
		json.Unmarshal(respBody, userInfo)
		// logger.Infof("uinfo: %#v", userInfo)
		// 		if len(userInfo.Accounts) > 0 {
		// 			return &userInfo.Accounts[0], nil
		// 		}
		ulength := len(userInfo.Accounts)
		if ulength == 1 {
			// logger.Info("In single account")
			dscm.DocusignAccount = &userInfo.Accounts[0]
			return nil
		} else if ulength > 1 {
			for _, acc := range userInfo.Accounts {
				// index ignored in for statement
				// logger.Infof("account at index %d : %#v", index, acc)
				if acc.IsDefault {
					dscm.DocusignAccount = &acc
					return nil
				}
			}
			return fmt.Errorf("No default account set")
		} else {
			return fmt.Errorf("No accounts fetched")
		}
	} else if res.StatusCode == http.StatusUnauthorized {
		// refresh token
		err = dscm.DoRefreshToken()
		if err != nil {
			return fmt.Errorf("Failed to refresh token: %s", err.Error())
		}

		authHeader := "Bearer " + dscm.DocusignToken.AccessToken
		req.Header.Set("Authorization", authHeader)
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("Failed to get account info: %s", err.Error())
		}
		defer res.Body.Close()

		respBody, err = ReadResponseBody(res)
		if err != nil || respBody == nil {
			return fmt.Errorf("Failed to get response body: %s", err.Error())
		}

		if res.StatusCode == http.StatusOK {
			json.Unmarshal(respBody, userInfo)
			// if len(userInfo.Accounts) > 0 {
			// 	dscm.DocusignAccount = &userInfo.Accounts[0]
			// 	return nil
			// }
			ulength := len(userInfo.Accounts)
			if ulength == 1 {
				dscm.DocusignAccount = &userInfo.Accounts[0]
				return nil
			} else if ulength > 1 {
				for _, acc := range userInfo.Accounts {
					// logger.Infof("account at index %d : %#v", index, acc)
					if acc.IsDefault {
						dscm.DocusignAccount = &acc
						return nil
					}
				}
				return fmt.Errorf("No default account set")
			} else {
				return fmt.Errorf("No accounts fetched")
			}
		}

	}

	return err
}

// func sendGetUserInfoRequest(token *Token) (*http.Response, error) {
// 	requestUrl := "https://" + token.Env + ".docusign.com/oauth/userinfo"
// 	method := http.MethodGet
// 	req, _ := http.NewRequest(method, requestUrl, nil)
// 	authHeader := "Bearer " + token.AccessToken

// 	req.Header.Set("Authorization", authHeader)
// 	res, err := http.DefaultClient.Do(req)
// 	return res, err
// }

func (dscm *DocusignSharedConfigManager) generateJWT(s *Settings) (string, error) {

	logger.Debug("Generating JWT")
	issueTime := time.Now().Unix()
	integratorKey := s.IntegratorKey
	userId := s.UserID
	if userId == "" {
		return "", fmt.Errorf("User ID cannot be empty")
	}

	privateKeyData := s.RSAPrivateKey
	if strings.HasPrefix(privateKeyData, "{") {
		type privateKeyFile struct {
			Content  string
			Filename string
		}
		privateKey := &privateKeyFile{}

		err := json.Unmarshal([]byte(privateKeyData), privateKey)
		if err != nil {
			return "", err
		}

		if privateKey.Filename == "" {
			return "", fmt.Errorf("Private Key file name not found")
		}
		privateKeyData = privateKey.Content
		if privateKeyData == "" {
			return "", fmt.Errorf("Private Key content not found")
		}

		index := strings.IndexAny(privateKeyData, ",")
		if index != -1 {
			privateKeyData = privateKeyData[index+1:]
		}
	}

	privateKeyValue, err := base64.StdEncoding.DecodeString(privateKeyData)
	if err != nil {
		return "", fmt.Errorf("Unable to decode private key %s", err.Error())
	}

	authURI := getDomain(s.Environment) + ".docusign.com"

	claims := &JWTClaim{
		jwt.StandardClaims{
			Audience:  authURI,
			ExpiresAt: issueTime + 3600,
			IssuedAt:  issueTime,
			Issuer:    integratorKey,
			Subject:   userId,
		},
		"signature impersonation",
	}

	// logger.Info("claims: \n", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signingKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyValue)
	if err != nil {
		return "", fmt.Errorf("Error occured while parsing key")
	}

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("Error occured while signing JWT")
	}

	return signedToken, nil

}

func (dscm *DocusignSharedConfigManager) getAccessTokenWithJWT(s *Settings) error {
	logger.Debug("Getting access token with jwt")
	jWToken, err := dscm.generateJWT(s)
	if err != nil {
		return fmt.Errorf("Error generating token %s", err.Error())
	}

	queryParam := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jWToken},
	}

	requestBody := strings.NewReader(queryParam.Encode())

	request, err := http.NewRequest("POST", getAuthTokenUrl(getDomain(s.Environment)), requestBody)
	if err != nil {
		return fmt.Errorf("Error creating authenitcation request: %v", err)
	}

	request.Header.Set("User-Agent", "Web Integrator")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error sending authentication request: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading authentication response bytes: %v", err)
	}

	respErr := &RespError{}

	if err := json.Unmarshal(respBytes, respErr); err == nil {
		if respErr.Error != "" {
			return fmt.Errorf("JWT token error %s %s", respErr.Error, respErr.Description)
		}
	}

	docuSignToken := &Token{}
	if err := json.Unmarshal(respBytes, docuSignToken); err != nil {
		return fmt.Errorf("Unable to unmarshal authentication response: %v", err)
	}

	if docuSignToken.AccessToken == "" {
		return fmt.Errorf("Access Token with invalid result: %v", docuSignToken)
	}

	// logger.Info("Contents of dosusign token: ", docuSignToken)
	// logger.Info("Client ID: ", docuSignToken.ClientID)
	// logger.Info("Client Secret: ", docuSignToken.ClientSecret)
	// logger.Info("Connection ID: ", docuSignToken.ConnectionID)
	// logger.Info("Access Token: ", docuSignToken.AccessToken)
	// logger.Info("Token type: ", docuSignToken.TokenType)
	// logger.Info("Refresh Token: ", docuSignToken.RefreshToken)
	// logger.Info("Expire IN: ", docuSignToken.ExpiresIn)
	// logger.Info("Env: ", docuSignToken.Env)

	docuSignToken.ClientID = s.IntegratorKey
	docuSignToken.ClientSecret = s.SecretKey
	docuSignToken.Env = getDomain(s.Environment)
	docuSignToken.ConnectionID = s.Name
	docuSignToken.PrivateKey = s.RSAPrivateKey
	docuSignToken.UserId = s.UserID

	// logger.Infof("JWT token: %#v", docuSignToken)

	dscm.DocusignToken = docuSignToken

	return nil
}

// func getAuthURI(domain string) string {
// 	return domain + ".docusign.com"
// }

func getDomain(prod bool) string {
	if prod {
		return "account"
	}

	return "account-d"
}

func getEnv(domain string) bool {
	if domain == "account-d" {
		return false
	} else if domain == "account" {
		return true
	}

	return true
}

func getAuthTokenUrl(domain string) string {
	return "https://" + domain + ".docusign.com/oauth/token"
}

func (dscm *DocusignSharedConfigManager) Type() string {
	return "Docusign"
}

func (dscm *DocusignSharedConfigManager) GetConnection() interface{} {
	return dscm
}

func (dscm *DocusignSharedConfigManager) ReleaseConnection(connection interface{}) {
}

func (dscm *DocusignSharedConfigManager) Start() error {
	return nil
}

func (dscm *DocusignSharedConfigManager) Stop() error {
	return nil
}

func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {

	var manager connection.Manager
	var err error
	_, ok := conn.(map[string]interface{})
	if ok {
		manager, err = handleLegacyConnection(conn)
	} else {
		manager, err = coerce.ToConnection(conn)
	}
	if err != nil {
		return nil, err
	}

	return manager, nil
}

func handleLegacyConnection(conn interface{}) (connection.Manager, error) {

	connectionObject, _ := coerce.ToObject(conn)
	if connectionObject == nil {
		return nil, fmt.Errorf("Connection object is nil")
	}

	id := connectionObject["id"].(string)

	manager := connection.GetManager(id)
	if manager == nil {

		connObject, err := generic.NewConnection(connectionObject)
		if err != nil {
			return nil, err
		}

		manager, err = factory.NewManager(connObject.Settings())
		if err != nil {
			return nil, err
		}

		connection.RegisterManager(id, manager)
	}
	return manager, nil
}

// ReadResponseBody gets the body of http reponse
func ReadResponseBody(res *http.Response) ([]byte, error) {
	if res.Body != nil {
		respBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return respBody, nil
	}
	return nil, errors.New("Response doesn't have a body")
}
