package connection

import (
	"os"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

var logCache = log.ChildLogger(log.RootLogger(), "azuread.connection")

var factory = &AzureADFactory{}

// Settings for AzureAD
type Settings struct {
	Name                           string `md:"name,required"`
	Description                    string `md:"description"`
	TenantId                       string `md:"tenantId,required"`
	ClientID                       string `md:"clientID,required"`
	UserName                       string `md:"userName,required"`
	Password                       string `md:"password,required"`
	ResourceURL                    string `md:"resourceURL,required"`
	GrantType                      string `md:"grantType"`
	WI_STUDIO_OAUTH_CONNECTOR_INFO string `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
	ConfigProperties               string `md:"configProperties"`
}

func init() {
	if os.Getenv(log.EnvKeyLogLevel) == "DEBUG" {
		logCache.DebugEnabled()
	}

	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type AzureADFactory struct {
}

// Type AzureADFactory
func (*AzureADFactory) Type() string {
	return "AzureADFactory"
}

// AzureADConfigManager details
type AzureADConfigManager struct {
	connSettings map[string]interface{}
}

// NewManager AzureADFactory
func (*AzureADFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &AzureADConfigManager{}
	var err error

	s := &Settings{}
	err = metadata.MapToStruct(settings, s, false)

	if err != nil {
		return nil, err
	}
	sharedConn.connSettings = settings

	return sharedConn, nil
}

// Type AzureADConfigManager
func (p *AzureADConfigManager) Type() string {

	return "AzureAD"
}

// GetConnection AzureADConfigManager details
func (p *AzureADConfigManager) GetConnection() interface{} {
	return p.connSettings
}

// Start AzureADConfigManager
func (p *AzureADConfigManager) Start() error {
	return nil
}

// Stop AzureADConfigManager
func (p *AzureADConfigManager) Stop() error {
	return nil
}

// ReleaseConnection AzureADConfigManager
func (p *AzureADConfigManager) ReleaseConnection(connection interface{}) {

}
