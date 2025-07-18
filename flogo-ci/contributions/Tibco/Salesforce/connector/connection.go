package connector

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

	"github.com/golang-jwt/jwt/v4"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
	"github.com/tibco/wi-contrib/environment"
)

const (
	DEFAULT_APIVERSION = "v52.0"
	SF_PROD            = "Production"
	SF_SANDBOX         = "Sandbox"
	TOKEN_URL_PROD     = "https://login.salesforce.com/services/oauth2/token"
	TOKEN_URL_SANDBOX  = "https://test.salesforce.com/services/oauth2/token"
)

type Settings struct {
	Name                           string `md:"name"`
	Description                    string `md:"description"`
	AuthType                       string `md:"authType"`
	Environment                    string `md:"environment"`
	CustomOAuth2Credentials        bool   `md:"customOAuth2Credentials"`
	GenerateJWT                    bool   `md:"generateJwt"`
	ClientID                       string `md:"clientId"`
	ClientSecret                   string `md:"clientSecret"`
	JWT                            string `md:"jwt"`
	Subject                        string `md:"subject"`
	JWTExpiry                      int    `md:"jwtExpiry"`
	ClientKey                      string `md:"clientKey"`
	APIVersion                     string `md:"apiVersion"`
	WI_STUDIO_OAUTH_CONNECTOR_INFO string `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
}
type SalesforceToken struct {
	ClientId     string `json:"client_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	InstanceUrl  string `json:"instance_url"`
	Scope        string `json:"scope"`
}
type SalesforceSharedConfigManager struct {
	SalesforceToken         *SalesforceToken
	ConnectionName          string
	AuthType                string
	EnvironmentType         string
	CustomOAuth2Credentials bool
	GenerateJWT             bool
	ClientId                string
	ClientSecret            string
	JWT                     string
	Subject                 string
	JWTExpiry               int
	ClientKey               string
	APIVersion              string
}

var logCache = log.ChildLogger(log.RootLogger(), "salesforce.connection")

var factory = &SalesforceManagerFactory{}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// SalesforceManagerFactory implements ManagerFactory from connection
type SalesforceManagerFactory struct {
}

func (*SalesforceManagerFactory) Type() string {
	return "Salesforce"
}

func (*SalesforceManagerFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {

	logCache.Debug("Initializing Salesforce connection")

	sscm := &SalesforceSharedConfigManager{}
	var err error
	err = sscm.getSalesforceClientConfig(settings)
	if err != nil {
		return nil, err
	}

	logCache.Infof("Salesforce connection name: %s and environment type: %s", sscm.ConnectionName, sscm.EnvironmentType)
	return sscm, nil
}

// getSalesforceClientConfig returns SalesforceToken which will be used for further API calls
func (sscm *SalesforceSharedConfigManager) getSalesforceClientConfig(settings map[string]interface{}) error {

	s := &Settings{}

	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		logCache.Errorf("Error occured during Settings MapToStruct conversion in getSalesforceClientConfig()..")
		return err
	}
	sscm.ConnectionName = s.Name

	if sscm.ConnectionName != "" {
		sscm.EnvironmentType = s.Environment
		sscm.AuthType = s.AuthType
		sscm.APIVersion = s.APIVersion
		sscm.GenerateJWT = s.GenerateJWT

		sscm.SalesforceToken = &SalesforceToken{}
		if sscm.AuthType != "OAuth 2.0 JWT Bearer Flow" {

			sscm.CustomOAuth2Credentials = s.CustomOAuth2Credentials
			tokenValue := s.WI_STUDIO_OAUTH_CONNECTOR_INFO

			if strings.HasPrefix(tokenValue, "{") {
				err = json.Unmarshal([]byte(tokenValue), sscm.SalesforceToken)
				if err != nil {
					logCache.Errorf("Error occured while unmarshalling token ", err)
					return err
				}
			} else {
				logCache.Debug("Decoding Salesforce connection credentials")
				decodedTokenByteData, err := base64.StdEncoding.DecodeString(tokenValue)
				if err != nil {
					logCache.Errorf("Error occured while decoding token ", err)
					return err
				}
				err = json.Unmarshal(decodedTokenByteData, sscm.SalesforceToken)
				if err != nil {
					logCache.Errorf("Error occured while unmarshalling token ", err)
					return err
				}
			}

			sscm.ClientId = sscm.SalesforceToken.ClientId

			isTCIEnv := environment.IsTCIEnv()
			logCache.Debug("isTCIEnv?: ", isTCIEnv)
			logCache.Debug("isCustomOAuth2Credentials?: ", sscm.CustomOAuth2Credentials)

			if isTCIEnv && !sscm.CustomOAuth2Credentials {
				sscm.ClientSecret = environment.GetSalesforceClientSecret()
			} else if !isTCIEnv && !sscm.CustomOAuth2Credentials {
				logCache.Errorf("Client Id and client secret are not set in Salesforce connection. While running application on-premise, you must configure custom OAuth2 credentials on the Salesforce connection.")
				return errors.New("Client Id and client secret are not set in Salesforce connection. While running application on-premise, you must configure custom OAuth2 credentials on the Salesforce connection.")
			} else {
				logCache.Debug("Using custom client secret")
				sscm.ClientSecret = s.ClientSecret
			}
			err = sscm.DoRefreshToken(logCache)
			if err != nil {
				logCache.Errorf("Error occured while getting token ", err)
				return err
			}
		} else if sscm.AuthType == "OAuth 2.0 JWT Bearer Flow" && !sscm.GenerateJWT {

			sscm.JWT = s.JWT

			err = sscm.DoRefreshTokenUsingJWT(logCache)
			if err != nil {
				logCache.Errorf("Error occured while getting token using jwt ", err)
				return err
			}
		} else {

			err := sscm.GenerateJSONWebToken(s.ClientID, s.Subject, s.JWTExpiry, s.ClientKey, logCache)

			if err != nil {
				return err
			}

			err = sscm.DoRefreshTokenUsingJWT(logCache)
			if err != nil {
				logCache.Errorf("Error occured while getting token using jwt ", err)
				return err
			}

		}

		//fmt.Printf("sscm : %+v\n", sscm)

		return nil
	}
	return fmt.Errorf("The connection name is empty")
}

// SalesforceSharedConfigManager implements Manager from connection

func (s *SalesforceSharedConfigManager) Type() string {
	return "Salesforce"
}

func (s *SalesforceSharedConfigManager) GetConnection() interface{} {
	return s
}

func (s *SalesforceSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {

	var cManager connection.Manager
	var err error
	_, ok := conn.(map[string]interface{})
	if ok {
		cManager, err = handleLegacyConnection(conn)
	} else {
		cManager, err = coerce.ToConnection(conn)
	}

	if err != nil {
		return nil, err
	}
	return cManager, nil
}

func handleLegacyConnection(conn interface{}) (connection.Manager, error) {

	connectionObject, _ := coerce.ToObject(conn)
	if connectionObject == nil {
		return nil, errors.New("Connection object is nil")
	}

	id := connectionObject["id"].(string)

	cManager := connection.GetManager(id)
	if cManager == nil {

		connObject, err := generic.NewConnection(connectionObject)
		if err != nil {
			return nil, err
		}

		cManager, err = factory.NewManager(connObject.Settings())
		if err != nil {
			return nil, err
		}

		err = connection.RegisterManager(id, cManager)
		if err != nil {
			return nil, err
		}
	}

	return cManager, nil

}

func (sscm *SalesforceSharedConfigManager) DoRefreshToken(log log.Logger) error {

	queryParam := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {sscm.ClientId},
		"client_secret": {sscm.ClientSecret},
		"refresh_token": {sscm.SalesforceToken.RefreshToken},
	}

	body := strings.NewReader(queryParam.Encode())

	tokenUrl := TOKEN_URL_PROD
	if strings.EqualFold(sscm.EnvironmentType, SF_SANDBOX) || strings.EqualFold(sscm.EnvironmentType, "sandbox") || strings.EqualFold(sscm.EnvironmentType, "SANDBOX") {
		tokenUrl = TOKEN_URL_SANDBOX
	}

	log.Debugf("Refreshing token for environment: %s against endpoint: %s", sscm.EnvironmentType, tokenUrl)
	request, err := http.NewRequest("POST", tokenUrl, body)
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

	var errs = struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}{}

	if err := json.Unmarshal(respBytes, &errs); err == nil {
		if errs.Error != "" {
			return fmt.Errorf("Refresh token error %s %s", errs.Error, errs.Description)
		}
	}

	newToken := &SalesforceToken{}
	if err := json.Unmarshal(respBytes, newToken); err != nil {
		return fmt.Errorf("Unable to unmarshal authentication response: %v", err)
	}

	if newToken.AccessToken == "" {
		return fmt.Errorf("Token refresh with invalid result: %v", newToken)
	}

	newToken.RefreshToken = sscm.SalesforceToken.RefreshToken
	sscm.SalesforceToken = newToken
	return nil
}

func (sscm *SalesforceSharedConfigManager) DoRefreshTokenUsingJWT(log log.Logger) error {

	queryParam := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {sscm.JWT},
	}

	body := strings.NewReader(queryParam.Encode())

	tokenUrl := TOKEN_URL_PROD
	if strings.EqualFold(sscm.EnvironmentType, SF_SANDBOX) || strings.EqualFold(sscm.EnvironmentType, "sandbox") || strings.EqualFold(sscm.EnvironmentType, "SANDBOX") {
		tokenUrl = TOKEN_URL_SANDBOX
	}

	log.Debugf("Getting token for environment: %s against endpoint: %s", sscm.EnvironmentType, tokenUrl)
	request, err := http.NewRequest("POST", tokenUrl, body)
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

	var errs = struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}{}

	if err := json.Unmarshal(respBytes, &errs); err == nil {
		if errs.Error != "" {
			return fmt.Errorf("Refresh token error %s %s", errs.Error, errs.Description)
		}
	}

	newToken := &SalesforceToken{}
	if err := json.Unmarshal(respBytes, newToken); err != nil {
		return fmt.Errorf("Unable to unmarshal authentication response: %v", err)
	}

	if newToken.AccessToken == "" {
		return fmt.Errorf("Token refresh with invalid result: %v", newToken)
	}
	sscm.SalesforceToken = newToken

	return nil
}

func (sscm *SalesforceSharedConfigManager) GenerateJSONWebToken(clientID string, subject string, jwtExp int, clientPrivateKey string, log log.Logger) error {

	url := "https://login.salesforce.com"
	if strings.EqualFold(sscm.EnvironmentType, SF_SANDBOX) || strings.EqualFold(sscm.EnvironmentType, "sandbox") || strings.EqualFold(sscm.EnvironmentType, "SANDBOX") {
		url = "https://test.salesforce.com"
	}

	log.Debugf("Generating json web token for environment: %s against endpoint: %s", sscm.EnvironmentType, url)

	sscm.ClientId = clientID
	sscm.Subject = subject
	sscm.JWTExpiry = jwtExp
	sscm.ClientKey = clientPrivateKey

	atClaims := jwt.MapClaims{}
	atClaims["iss"] = sscm.ClientId
	atClaims["aud"] = url
	atClaims["exp"] = time.Now().Add(time.Minute * time.Duration(sscm.JWTExpiry)).Unix()
	atClaims["sub"] = sscm.Subject
	clientKey, err := decodeCerts(sscm.ClientKey)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, atClaims)
	pkey, err := jwt.ParseRSAPrivateKeyFromPEM(clientKey)

	if err != nil {
		return err
	}
	jwtToken, err := token.SignedString(pkey)
	if err != nil {
		return err
	}
	sscm.JWT = jwtToken
	return nil
}

func decodeCerts(certVal string) ([]byte, error) {
	if certVal == "" {
		return nil, fmt.Errorf("Certificate is not configured")
	}

	//if certificate comes from fileselctor it will be base64 encoded
	if strings.HasPrefix(certVal, "{") {
		certObj, err := coerce.ToObject(certVal)
		if err == nil {
			certRealValue, ok := certObj["content"].(string)
			if !ok || certRealValue == "" {
				return nil, fmt.Errorf("Invalid certificate value")
			}

			index := strings.IndexAny(certRealValue, ",")
			if index > -1 {
				certRealValue = certRealValue[index+1:]
			}

			encodedDataOfCert, err := base64.StdEncoding.DecodeString(certRealValue)
			if err != nil {
				return nil, fmt.Errorf("Invalid base64 encoded certificate value")
			}
			return []byte(encodedDataOfCert), nil
		}
		return nil, err
	}

	encodedDataOfCert, err := base64.StdEncoding.DecodeString(certVal)
	if err != nil {
		return nil, fmt.Errorf("Invalid base64 encoded certificate. Check override value configured to the application property.")
	}
	return []byte(encodedDataOfCert), nil
}
