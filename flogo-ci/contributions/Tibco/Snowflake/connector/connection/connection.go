package connection

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tibco/wi-contrib/connection/generic"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"

	sf "github.com/snowflakedb/gosnowflake"
)

var logCache = log.ChildLogger(log.RootLogger(), "Snowflake-connection")
var factory = &SnowflakeFactory{}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

// Settings corresponds to connection.json settings
type Settings struct {
	Name                  string `md:"name,required"`
	Account               string `md:"account,required"`
	Warehouse             string `md:"warehouse,required"`
	Database              string `md:"database,required"`
	Schema                string `md:"schema,required"`
	AuthType              string `md:"authType,required"`
	Provider              string `md:"provider,required"`
	User                  string `md:"user,required"`
	Password              string `md:"password,required"`
	ClientID              string `md:"clientId,required"`
	ClientSecret          string `md:"clientSecret,required"`
	OktaTokenEndpoint     string `md:"oktaTokenEndpoint,required"`
	AuthCode              string `md:"authCode,required"`
	RedirectURI           string `md:"redirectURI,required"`
	OktaCodeVerifier      string `md:"oktaCodeVerifier,required"`
	OktaCodeChallenge     string `md:"oktaCodeChallenge,required"`
	OktaAccessToken       string `md:"oktaAccessToken,required"`
	OktaAccessTokenExpiry int64  `md:"oktaAccessTokenExpiry,required"`
	Scope                 string `md:"scope,required"`
	AccessToken           string `md:"accessToken,required"`
	AccessTokenExpiry     int64  `md:"accessTokenExpiry,required"`
	RefreshToken          string `md:"refreshToken,required"`
	RefreshTokenExpiry    int64  `md:"refreshTokenExpiry,required"`
	OktaRefreshToken      string `md:"oktaRefreshToken,required"`
	Role                  string `md:"role,required"`
	LoginTimeout          int    `md:"loginTimeout,required"`
	CodeCheck             string `md:"codeCheck,required"`
}

// SnowflakeFactory structure
type SnowflakeFactory struct {
}

// Type method of connection.ManagerFactory must be implemented by SnowflakeFactory
func (*SnowflakeFactory) Type() string {
	return "Snowflake"
}

func (sn *Settings) Validate() error {
	if sn.Account == "" {
		return errors.New("Required parameter 'Account' not specified")
	}

	if sn.Warehouse == "" {
		return errors.New("Required parameter 'Warehouse' not specified")
	}

	if sn.Database == "" {
		return errors.New("Required parameter 'Database' not specified")
	}

	/*if sn.Schema == "" {
		return errors.New("Required parameter 'Schema' not specified")
	}*/

	if sn.LoginTimeout < 1 {
		return errors.New("Required parameter 'Login Timeout' not specified")
	}

	if sn.AuthType == "Basic Authentication" {
		if sn.User == "" {
			return errors.New("Required parameter 'User' not specified")
		}

		if sn.Password == "" {
			return errors.New("Required parameter 'Password' not specified")
		}
	} else {
		if sn.AuthType == "OAuth" {
			if sn.ClientID == "" {
				return errors.New("Required parameter 'ClientId' not specified")
			}
			if sn.RedirectURI == "" {
				return errors.New("Required parameter 'Redirect URI' not specified")
			}
			if sn.AuthCode == "" {
				return errors.New("Required parameter 'Authorization Code' not specified")
			}
			if sn.Provider == "Snowflake" {
				if sn.ClientSecret == "" {
					return errors.New("Required parameter 'ClientSecret' not specified")
				}
			} else {
				// okta
				if sn.OktaTokenEndpoint == "" {
					return errors.New("Required parameter 'OktaTokenEndpoint' not specified")
				}
				if sn.OktaCodeVerifier == "" {
					return errors.New("Required parameter 'OktaCodeVerifier' not specified")
				}
				if sn.OktaCodeChallenge == "" {
					return errors.New("Required parameter 'OktaCodeChallenge' not specified")
				}
				// scope necessary for okta
				if sn.Scope == "" {
					return errors.New("Required parameter 'Scope' not specified")
				}
			}

		}

	}

	return nil
}

func getDSN(snowflakeConn *Settings) (string, error) {
	cfg := &sf.Config{
		Application:  "TIBCO_FLOGO",
		Account:      snowflakeConn.Account,
		Warehouse:    snowflakeConn.Warehouse,
		Database:     snowflakeConn.Database,
		Schema:       snowflakeConn.Schema,
		LoginTimeout: time.Duration(snowflakeConn.LoginTimeout) * time.Second,
	}

	if strings.TrimSpace(snowflakeConn.Role) != "" {
		cfg.Role = snowflakeConn.Role
	}

	if snowflakeConn.AuthType == "Basic Authentication" {
		cfg.User = snowflakeConn.User
		cfg.Password = snowflakeConn.Password
		cfg.Authenticator = sf.AuthTypeSnowflake
	} else {
		cfg.Authenticator = sf.AuthTypeOAuth
		if snowflakeConn.Provider == "Snowflake" {
			if snowflakeConn.AuthCode != snowflakeConn.CodeCheck {
				err := getAccessTokenFromAuthCode(snowflakeConn)
				if err != nil {
					return "", err
				}
				cfg.Token = snowflakeConn.AccessToken
				snowflakeConn.CodeCheck = snowflakeConn.AuthCode
			} else if snowflakeConn.AccessTokenExpiry > time.Now().Unix() {
				logCache.Info("Access Token expired. Generating new access token.")

				if snowflakeConn.RefreshTokenExpiry < time.Now().Unix() {
					return "", errors.New("Refresh Token has expired. Please provide new Authorization Code to generate new Access Token and Refresh Token")
				}

				err := getAccessTokenFromRefreshToken(snowflakeConn)
				if err != nil {
					return "", err
				}
				cfg.Token = snowflakeConn.AccessToken
			} else {
				cfg.Token = snowflakeConn.AccessToken
			}
		} else {
			// okta oauth
			timeNow := time.Now().Unix() * 1000
			if snowflakeConn.AuthCode != snowflakeConn.CodeCheck {
				err := getOktaAccessTokenFromAuthCode(snowflakeConn)
				if err != nil {
					return "", err
				}
				cfg.Token = snowflakeConn.OktaAccessToken
				snowflakeConn.CodeCheck = snowflakeConn.AuthCode
			} else if snowflakeConn.OktaAccessTokenExpiry < timeNow {
				logCache.Info("Okta Access Token expired. Generating new access token.")
				err := getOktaAccessTokenFromRefreshToken(snowflakeConn)
				if err != nil {
					return "", err
				}
				cfg.Token = snowflakeConn.OktaAccessToken
			} else {
				cfg.Token = snowflakeConn.OktaAccessToken
			}
		}
	}
	dsn, err := sf.DSN(cfg)
	return dsn, err
}

// NewManager method of connection.ManagerFactory must be implemented by SnowflakeFactory
func (*SnowflakeFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &SnowflakeSharedConfigManager{}
	var err error
	s := &Settings{}

	err = metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}
	//1. Validate connection
	err = s.Validate()
	if err != nil {
		return nil, fmt.Errorf("Snowflake connection validation error: %s", err.Error())
	}

	//2. Get the DSN
	dsn, err := getDSN(s)
	if err != nil {
		return nil, fmt.Errorf("Failed to create DSN from Connection parameters, error: %v", err.Error())
	}

	//3. Login and Ping DB
	logCache.Infof("Logging into Snowflake DB. Connection name : %s", s.Name)
	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, fmt.Errorf("Cannot login Database. Error: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Cannot login Database. Error: %s", err.Error())
	}

	logCache.Infof("Login successful. Connection name : %s", s.Name)

	sharedConn.connName = s.Name
	sharedConn.db = db
	return sharedConn, nil
}

// SnowflakeSharedConfigManager structure
type SnowflakeSharedConfigManager struct {
	connName string
	db       *sql.DB
}

// Type method of connection.Manager must be implemented by SnowflakeSharedConfigManager
func (o *SnowflakeSharedConfigManager) Type() string {
	return "Snowflake"
}

// GetConnection method of connection.Manager must be implemented by SnowflakeSharedConfigManager
func (s *SnowflakeSharedConfigManager) GetConnection() interface{} {
	return s.db
}

// ReleaseConnection method of connection.Manager must be implemented by SnowflakeSharedConfigManager
func (o *SnowflakeSharedConfigManager) ReleaseConnection(connection interface{}) {
}

// GetSharedConfiguration returns connection.Manager based on connection selected
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

// Start method would have business logic to start the shared resource. Since db is already initialized returning nil.
func (o *SnowflakeSharedConfigManager) Start() error {
	return nil
}

// Stop method would do business logic to stop the the shared resource. Closing db connection in this method.
func (s *SnowflakeSharedConfigManager) Stop() error {
	if s.db == nil {
		return nil
	}

	logCache.Infof("Logging out of Snowflake DB. Connection name : %s", s.connName)
	err := s.db.Close()
	if err != nil {
		return err
	}
	logCache.Infof("Logged out of Snowflake DB successfully. Connection name : %s", s.connName)

	return nil
}
