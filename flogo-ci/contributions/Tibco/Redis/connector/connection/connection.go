package redis

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	redis "github.com/redis/go-redis/v9"
	"github.com/tibco/wi-contrib/connection/generic"
)

var logCache = log.ChildLogger(log.RootLogger(), "connection.redis")

type Settings struct {
	Host                 string `md:"Host"`
	Port                 int    `md:"Port"`
	Description          string `md:"Description"`
	AuthMode             bool   `md:"AuthMode"`
	Name                 string `md:"Name"`
	DefaultDatabaseIndex int    `md:"DefaultDatabaseIndex"`
	ClientCert           string `md:"ClientCert"`
	ClientKey            string `md:"ClientKey"`
	CaCert               string `md:"CaCert"`
	Password             string `md:"Password"`
	DocsMetadata         string `md:"DocsMetadata"`
}

// RedisToken  Needed for holding auth creds
type RedisToken struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	Password string `json:"Password"`
}

type RedisSharedConfigManager struct {
	RedisToken           *RedisToken
	ConnectionName       string
	DocsMetadata         string
	DocsObject           map[string]interface{}
	ConfigProperties     string
	ClientCert           string
	ClientKey            string
	CaCert               string
	RedisClient          *redis.Client
	DefaultDatabaseIndex float64
	AuthMode             bool
}

var factory = &RedisManagerFactory{}

func init() {
	logCache.Debug("Calling init()")
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type RedisManagerFactory struct {
}

func (*RedisManagerFactory) Type() string {
	return "Redis"
}

func (s *RedisSharedConfigManager) Type() string {
	return "Redis"
}

func (s *RedisSharedConfigManager) GetConnection() interface{} {
	return s
}

func (s *RedisSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func (*RedisManagerFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &RedisSharedConfigManager{}
	var err error
	err = sharedConn.getRedisClientConfig(settings)
	if err != nil {
		return nil, err
	}
	return sharedConn, nil
}

func (rcm *RedisSharedConfigManager) getRedisClientConfig(settings map[string]interface{}) error {
	s := &Settings{}
	var token RedisToken

	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		logCache.Errorf("Error occured during Settings MapToStruct conversion in getRedisClientConfig()..")
		return err
	}

	rcm.ConnectionName = s.Name
	if rcm.ConnectionName != "" {
		if rcm.DocsObject == nil {
			var docsMetadata interface{}
			err = json.Unmarshal([]byte(s.DocsMetadata), &docsMetadata)

			if err != nil {
				return fmt.Errorf("Cannot deserialize schema document from connection %s", err.Error())

			}
			rcm.DocsObject = docsMetadata.(map[string]interface{})
		}

		rcm.DocsMetadata = s.DocsMetadata
		token.Host = s.Host
		token.Password = s.Password
		token.Port = s.Port
		rcm.DefaultDatabaseIndex = float64(s.DefaultDatabaseIndex)
		rcm.AuthMode = s.AuthMode

		if s.AuthMode == true {

			rcm.ClientCert = s.ClientCert
			rcm.ClientKey = s.ClientKey
			rcm.CaCert = s.CaCert

		}

		//token.PrimarysecondaryKey = s.PrimarysecondaryKey
		//token.ResourceURI = s.ResourceURI ///this need to be set correcctly

		logCache.Debugf("value set to token struct")

		rcm.ConnectionName = s.Name
		if rcm.ConnectionName != "" {
			rcm.RedisToken = &token

			return nil
		}
	}
	return fmt.Errorf("The connection name is empty")
}

// Start ...
func (s *RedisSharedConfigManager) Start() error {
	return nil
}

// Stop ...
func (s *RedisSharedConfigManager) Stop() error {
	return nil
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
