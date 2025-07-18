package connection

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

const API_VERSION_PATH = "/3.0"
const TOKEN_URL = ""

var logCache = log.ChildLogger(log.RootLogger(), "Mailchimp.connection")

var factory = &MailChimpFactory{}

func init() {
	if os.Getenv(log.EnvKeyLogLevel) == "DEBUG" {
		logCache.DebugEnabled()
	}

	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// Settings for Mailchimp Connection
type Settings struct {
	Name                           string `md:"name,required"`
	Description                    string `md:"description"`
	ClientId                       string `md:"clientId,required"`
	ClientSecret                   string `md:"clientSecret,required"`
	WI_STUDIO_OAUTH_CONNECTOR_INFO string `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
}

//	type Token struct {
type Token struct {
	ClientId     string  `json:"client_id"`
	ClientSecret string  `json:"client_secret"`
	AccessToken  string  `json:"access_token"`
	ExpiresIn    int     `json:"expires_in"`
	Scope        *string `json:"scope"` // Handle null using *string
	DC           string  `json:"dc"`
	Role         string  `json:"role"`
	AccountName  string  `json:"accountname"`
	UserID       int     `json:"user_id"`
	Login        struct {
		Email      string  `json:"email"`
		Avatar     *string `json:"avatar"` // Handle null using *string
		LoginID    int     `json:"login_id"`
		LoginName  string  `json:"login_name"`
		LoginEmail string  `json:"login_email"`
	} `json:"login"`
	LoginURL    string `json:"login_url"`
	APIEndpoint string `json:"api_endpoint"`
}

// MailchimpConnectionManager details
type MailchimpConnectionManager struct {
	Token          *Token
	ConnectionName string
	ClientId       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
}

type MailChimpFactory struct {
}

// Type MailChimpFactory
func (*MailChimpFactory) Type() string {
	return "Mailchimp"
}

// NewManager MailChimpFactory
func (*MailChimpFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &MailchimpConnectionManager{
		Token: &Token{},
	}
	var err error

	s := &Settings{}
	err = metadata.MapToStruct(settings, s, false)

	if err != nil {
		return nil, err
	}
	sharedConn.ConnectionName = s.Name
	clientId := s.ClientId
	if clientId == "" {
		return nil, errors.New("Required Parameter client_id Name is empty")
	}
	sharedConn.ClientId = clientId
	clientSecret := s.ClientSecret
	if clientSecret == "" {
		return nil, errors.New("Required Parameter clientSecret is empty")
	}
	logCache.Debug("Using custom client secret")
	sharedConn.ClientSecret = clientSecret

	tokenValue := s.WI_STUDIO_OAUTH_CONNECTOR_INFO

	if strings.HasPrefix(tokenValue, "{") {
		err = json.Unmarshal([]byte(tokenValue), sharedConn.Token)
		if err != nil {
			logCache.Errorf("Error occured while unmarshalling token ", err)
			return nil, err
		}
	}

	return sharedConn, nil
}

// Type MailchimpSharedConfigManager details
func (p *MailchimpConnectionManager) Type() string {

	return "Mailchimp"
}

func (cm *MailchimpConnectionManager) GetConnection() interface{} {
	return cm
}
func (cm *MailchimpConnectionManager) GetOauthToken() interface{} {
	return cm.Token
}
func (cm *MailchimpConnectionManager) ReleaseConnection(connection interface{}) {

}

func (cm *MailchimpConnectionManager) Start() error {
	return nil
}

func (cm *MailchimpConnectionManager) Stop() error {
	return nil
}

