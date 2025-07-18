package azureservicebusconnection

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
)

type Settings struct {
	ResourceURI           string `md:"resourceURI"`
	AuthorizationRuleName string `md:"authorizationRuleName"`
	Description           string `md:"description"`
	DocsMetadata          string `md:"DocsMetadata"`
	Name                  string `md:"name"`
	PrimarysecondaryKey   string `md:"primarysecondaryKey"`
	Count                 int    `md:"count"`
	Interval              int    `md:"interval"`
	AuthMode              string `md:"authMode"`
	TenantID              string `md:"tenantID"`
	ClientID              string `md:"clientID"`
	ClientSecret          string `md:"clientSecret"`
}

type AzureToken struct {
	AuthorizationRuleName string `json:"authorizationRuleName"`
	PrimarysecondaryKey   string `json:"primarysecondaryKey"`
	ResourceURI           string `json:"resource_url"`
}

type AzureServiceBusSharedConfigManager struct {
	AzureToken          *AzureToken
	ConnectionName      string
	AccountSID          string
	PrimarysecondaryKey string
	DocsMetadata        string
	DocsObject          map[string]interface{}
	senderCache         map[string]*azservicebus.Sender
	ServiceBusClient    *azservicebus.Client
	senderCacheMutex    *sync.RWMutex
	retrycount          int
	retryinterval       int
	AuthMode            string
	TenantID            string
	ClientID            string
	ClientSecret        string
}

var logCache = log.ChildLogger(log.RootLogger(), "azureServiceBus-connection")
var factory = &AzureServiceBusManagerFactory{}

func init() {
	logCache.Debug("Calling init()")
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type AzureServiceBusManagerFactory struct {
}

func (*AzureServiceBusManagerFactory) Type() string {
	return "AzureServiceBus"
}

func (s *AzureServiceBusSharedConfigManager) Type() string {
	return "AzureServiceBus"
}

func (s *AzureServiceBusSharedConfigManager) GetConnection() interface{} {
	return s
}

func (s *AzureServiceBusSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func (ascm *AzureServiceBusSharedConfigManager) Start() error {
	var err error
	ascm.senderCache = make(map[string]*azservicebus.Sender)
	retryOptions := azservicebus.RetryOptions{
		MaxRetries:    int32(ascm.retrycount),
		RetryDelay:    time.Duration(ascm.retryinterval) * time.Millisecond,
		MaxRetryDelay: time.Duration(ascm.retryinterval) * time.Millisecond,
	}
	if ascm.AuthMode == "SAS Token" {
		connStr := ""
		if strings.HasPrefix(ascm.AzureToken.ResourceURI, "https") {
			u, err := url.Parse(ascm.AzureToken.ResourceURI)
			if err != nil {
				return fmt.Errorf("Unable to parse namespace url %s", err.Error())
			}
			namespace := u.Host
			if u.Path != "" {
				connStr = "Endpoint=sb://" + namespace + "/;SharedAccessKeyName=" + ascm.AzureToken.AuthorizationRuleName + ";SharedAccessKey=" + ascm.AzureToken.PrimarysecondaryKey + ";EntityPath=" + u.Path
			} else {
				connStr = "Endpoint=sb://" + namespace + "/;SharedAccessKeyName=" + ascm.AzureToken.AuthorizationRuleName + ";SharedAccessKey=" + ascm.AzureToken.PrimarysecondaryKey
			}
		} else {
			connStr = "Endpoint=sb://" + ascm.AzureToken.ResourceURI + ".servicebus.windows.net/;SharedAccessKeyName=" + ascm.AzureToken.AuthorizationRuleName + ";SharedAccessKey=" + ascm.AzureToken.PrimarysecondaryKey
		}

		ascm.ServiceBusClient, err = azservicebus.NewClientFromConnectionString(connStr,
			&azservicebus.ClientOptions{
				RetryOptions: retryOptions,
			})
		if err != nil {
			return err
		}
	} else {
		cred, err := azidentity.NewClientSecretCredential(ascm.TenantID, ascm.ClientID, ascm.ClientSecret, nil)
		if err != nil {
			return err
		}
		var ResourceURI string
		if strings.HasPrefix(ascm.AzureToken.ResourceURI, "https") {
			u, err := url.Parse(ascm.AzureToken.ResourceURI)
			if err != nil {
				return fmt.Errorf("Unable to parse namespace url %s", err.Error())
			}
			ResourceURI = u.Host
		} else {
			ResourceURI = ascm.AzureToken.ResourceURI + ".servicebus.windows.net"
		}

		ascm.ServiceBusClient, err = azservicebus.NewClient(ResourceURI, cred, &azservicebus.ClientOptions{
			RetryOptions: retryOptions,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AzureServiceBusSharedConfigManager) Stop() error {
	for _, sender := range s.senderCache {
		sender.Close(context.Background())
	}
	s.senderCache = nil
	return nil
}

func (*AzureServiceBusManagerFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &AzureServiceBusSharedConfigManager{}
	err := sharedConn.getAzureServiceBusClientConfig(settings)
	if err != nil {
		return nil, err
	}
	sharedConn.senderCacheMutex = &sync.RWMutex{}
	//init service bus logger to log connection attempt
	engineLogLevel := os.Getenv(log.EnvKeyLogLevel)
	if log.ToLogLevel(engineLogLevel) <= log.DebugLevel {
		InitLogger()
	}
	return sharedConn, nil
}

func (ascm *AzureServiceBusSharedConfigManager) getAzureServiceBusClientConfig(settings map[string]interface{}) error {
	s := &Settings{}
	var token AzureToken

	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		logCache.Errorf("Error occured during Settings MapToStruct conversion in getAzureServiceBusClientConfig()..")
		return err
	}

	ascm.ConnectionName = s.Name

	if ascm.ConnectionName != "" {

		if ascm.DocsObject == nil {
			var docsMetadata interface{}
			err = json.Unmarshal([]byte(s.DocsMetadata), &docsMetadata)

			if err != nil {
				return fmt.Errorf("Cannot deserialize schema document from connection %s", err.Error())

			}
			ascm.DocsObject = docsMetadata.(map[string]interface{})
		}

		ascm.DocsMetadata = s.DocsMetadata
		token.AuthorizationRuleName = s.AuthorizationRuleName
		token.PrimarysecondaryKey = s.PrimarysecondaryKey
		token.ResourceURI = s.ResourceURI

		ascm.AzureToken = &token

		ascm.retrycount = s.Count
		ascm.retryinterval = s.Interval
		ascm.AuthMode = s.AuthMode
		ascm.TenantID = s.TenantID
		ascm.ClientID = s.ClientID
		ascm.ClientSecret = s.ClientSecret
		return nil
	}
	return fmt.Errorf("The connection name is empty")
}

func (connection *AzureServiceBusSharedConfigManager) GetSenderConnection(senderName string) (*azservicebus.Sender, error) {

	//read existing connection
	connection.senderCacheMutex.RLock()
	sender, ok := connection.senderCache[senderName]
	connection.senderCacheMutex.RUnlock()
	if ok {
		return sender, nil
	}

	//cerate new connection
	connection.senderCacheMutex.Lock()
	defer connection.senderCacheMutex.Unlock()
	sender, err := connection.ServiceBusClient.NewSender(senderName, nil)
	if err != nil {
		return nil, err
	}
	connection.senderCache[senderName] = sender
	return sender, nil
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
