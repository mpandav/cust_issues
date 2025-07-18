package azstorage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azfile/share"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azfile/service"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

const (
	AuthModeSAS = "SAS Token"
	AuthModeCS  = "OAuth2"
)

var factory = &azStorageFactory{}

type Settings struct {
	Name           string        `md:"name,required"`
	Description    string        `md:"description"`
	ConnectionType string        `md:"connectionType"`
	Accountname    string        `md:"accountName"`
	SAS            string        `md:"sas"`
	Accesskey      string        `md:"accessKey"`
	ExpiryDate     string        `md:"expiryDate"`
	CONNECTORINFO  string        `md:"WI_STUDIO_OAUTH_CONNECTOR_INFO"`
	DocsMetadata   string        `md:"DocsMetadata"`
	RegenerateFlag bool          `md:"regenerateFlag"`
	RegenerateTime string        `md:"regenerateTime"`
	SasParameters  SasParameters `md:"SasParameters"`
	ContainerName  string        `md:"containerName"`
	AuthMode       string        `md:"authMode"`
	TenantID       string        `md:"tenantID"`
	ClientID       string        `md:"clientID"`
	ClientSecret   string        `md:"clientSecret"`
}

type SasParameters struct {
	SP  string `md:"sp"`
	SS  string `md:"ss"`
	SRT string `md:"srt"`
	SR  string `md:"sr"`
}

var connectorLog = log.ChildLogger(log.RootLogger(), "azure-storage-connection")

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type azStorageFactory struct {
}

func (*azStorageFactory) Type() string {
	return "azstorage"
}

func (*azStorageFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	sharedConn := &AzStorageSharedConfigManager{}
	var err error
	sharedConn.Config, err = getazstorageConfig(settings)
	if err != nil {
		return nil, err
	}
	return sharedConn, nil
}

type AzStorageSharedConfigManager struct {
	Config *Settings
	Name   string
	Cred   *azidentity.ClientSecretCredential
}

func (k *AzStorageSharedConfigManager) Type() string {
	return "azstorage"
}

func (k *AzStorageSharedConfigManager) GetConnection() interface{} {
	return k.Config
}
func (k *AzStorageSharedConfigManager) GetClient() {
	return
}

func (k *AzStorageSharedConfigManager) GetClientConfiguration() *Settings {
	return k.Config
}

func (k *AzStorageSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func (k *AzStorageSharedConfigManager) Start() error {
	if k.Config.AuthMode == AuthModeCS {
		cred, err := azidentity.NewClientSecretCredential(k.Config.TenantID, k.Config.ClientID, k.Config.ClientSecret, nil)
		if err != nil {
			return err
		}
		k.Cred = cred
	}
	return nil
}

func (k *AzStorageSharedConfigManager) Stop() error {

	return nil
}

func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {
	var cManager connection.Manager
	var err error
	cManager, err = coerce.ToConnection(conn)

	if err != nil {
		return nil, err
	}
	return cManager, nil
}

func getazstorageConfig(settings map[string]interface{}) (*Settings, error) {
	connectionConfig := &Settings{}
	err := metadata.MapToStruct(settings, connectionConfig, false)
	if err != nil {
		return nil, err
	}
	if connectionConfig.ConnectionType == "Generate SAS" {
		stringToSign := connectionConfig.Accountname + "\n" + "rwdlacup" + "\n" + "bfqt" + "\n" + "sco" + "\n" + ""
		stringToSign = stringToSign + "\n" + connectionConfig.ExpiryDate
		stringToSign = stringToSign + "\n" + ""
		stringToSign = stringToSign + "" + "\n" + "https" + "\n" + "2017-11-09" + "\n"
		signingKey := connectionConfig.Accesskey
		decodekey, _ := base64.StdEncoding.DecodeString(signingKey)
		h := hmac.New(sha256.New, decodekey)
		h.Write([]byte(stringToSign))
		hashInBase64 := base64.StdEncoding.EncodeToString(h.Sum(nil))
		connectionConfig.SAS = "sv=2017-11-09&ss=bfqt&srt=sco&sp=rwdlacup&se=" + connectionConfig.ExpiryDate + "&spr=https&sig=" + url.QueryEscape(hashInBase64)
	} else if connectionConfig.ConnectionType == "Enter SAS" {
		// extracting account, SAS token, permissons from URL or SAS
		var sasParameters SasParameters
		var queryMap url.Values
		if strings.HasPrefix(connectionConfig.SAS, "http") {
			u, err := url.Parse(connectionConfig.SAS)
			if err != nil {
				return nil, err
			}
			queryMap, _ = url.ParseQuery(u.RawQuery)
			connectionConfig.SAS = u.RawQuery
			connectionConfig.ContainerName = u.EscapedPath()[1:]
		} else {
			connectionConfig.SAS = strings.TrimPrefix(connectionConfig.SAS, "?")
			queryMap, _ = url.ParseQuery(connectionConfig.SAS)
		}

		if queryMap != nil {
			if queryMap["se"] != nil {
				connectionConfig.ExpiryDate = queryMap["se"][0]
			}
			if queryMap["ss"] != nil {
				sasParameters.SS = queryMap["ss"][0]
			}
			if queryMap["srt"] != nil {
				sasParameters.SRT = queryMap["srt"][0]
			}
			if queryMap["sp"] != nil {
				sasParameters.SP = queryMap["sp"][0]
			}
			if queryMap["sr"] != nil {
				sasParameters.SR = queryMap["sr"][0]
			}
		}
		connectionConfig.SasParameters = sasParameters

	}
	return connectionConfig, nil
}

func (k *AzStorageSharedConfigManager) RenewToken() error {
	currentDate, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	duration, err := time.ParseDuration(k.Config.RegenerateTime)
	if err != nil {
		return fmt.Errorf("Error parsing duration %s", k.Config.RegenerateTime)
	}
	expiryDate := currentDate.Add(duration)
	if err != nil {
		return err
	}
	k.Config.ExpiryDate = expiryDate.Format(time.RFC3339)

	stringToSign := k.Config.Accountname + "\n" + "rwdlacup" + "\n" + "bfqt" + "\n" + "sco" + "\n" + ""
	stringToSign = stringToSign + "\n" + k.Config.ExpiryDate
	stringToSign = stringToSign + "\n" + ""
	stringToSign = stringToSign + "" + "\n" + "https" + "\n" + "2017-11-09" + "\n"
	signingKey := k.Config.Accesskey
	decodekey, _ := base64.StdEncoding.DecodeString(signingKey)
	h := hmac.New(sha256.New, decodekey)
	h.Write([]byte(stringToSign))
	hashInBase64 := base64.StdEncoding.EncodeToString(h.Sum(nil))
	k.Config.SAS = "sv=2017-11-09&ss=bfqt&srt=sco&sp=rwdlacup&se=" + k.Config.ExpiryDate + "&spr=https&sig=" + url.QueryEscape(hashInBase64)

	connectorLog.Infof("SAS Token Renewed Successfully %s", k.Config.SAS)
	return nil
}

func (k *AzStorageSharedConfigManager) GetShareClient(params map[string]string) (*share.Client, error) {
	fileUrlSAS := fmt.Sprintf("https://%s.file.core.windows.net/%s?%s", k.Config.Accountname, params["shareName"], k.Config.SAS)
	fileUrl := fmt.Sprintf("https://%s.file.core.windows.net/%s", k.Config.Accountname, params["shareName"])

	switch k.Config.AuthMode {
	case AuthModeSAS:
		return share.NewClientWithNoCredential(fileUrlSAS, nil)
	case AuthModeCS:
		return share.NewClient(fileUrl, k.Cred, &share.ClientOptions{FileRequestIntent: to.Ptr(share.TokenIntentBackup)})
	default:
		return nil, fmt.Errorf("Invalid AuthMode: %s", k.Config.AuthMode)
	}
}

func (k *AzStorageSharedConfigManager) GetBlobClient() (*azblob.Client, error) {
	blobUrl := fmt.Sprintf("https://%s.blob.core.windows.net", k.Config.Accountname)
	blobUrlSAS := fmt.Sprintf("https://%s.blob.core.windows.net/?%s", k.Config.Accountname, k.Config.SAS)

	switch k.Config.AuthMode {
	case AuthModeSAS:
		return azblob.NewClientWithNoCredential(blobUrlSAS, nil)
	case AuthModeCS:
		return azblob.NewClient(blobUrl, k.Cred, nil)
	default:
		return nil, fmt.Errorf("Invalid AuthMode: %s", k.Config.AuthMode)
	}
}

func (k *AzStorageSharedConfigManager) GetShareServiceClient() (*service.Client, error) {
	fileUrl := fmt.Sprintf("https://%s.file.core.windows.net", k.Config.Accountname)
	fileUrlSAS := fmt.Sprintf("https://%s.file.core.windows.net/?%s", k.Config.Accountname, k.Config.SAS)

	switch k.Config.AuthMode {
	case AuthModeSAS:
		return service.NewClientWithNoCredential(fileUrlSAS, nil)
	case AuthModeCS:
		return service.NewClient(fileUrl, k.Cred, &service.ClientOptions{FileRequestIntent: to.Ptr(share.TokenIntentBackup)})
	default:
		return nil, fmt.Errorf("Invalid AuthMode: %s", k.Config.AuthMode)
	}
}
