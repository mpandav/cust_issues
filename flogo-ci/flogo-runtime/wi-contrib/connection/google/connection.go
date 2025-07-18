package google

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/project-flogo/core/data/coerce"
	"github.com/tibco/wi-contrib/connection/generic"
)

type Connection struct {
	*generic.Connection
	*serviceAccountInfo
}

type ServiceAccountKey struct {
	serviceAccountKey *serviceAccountInfo
}

type serviceAccountInfo struct {
	accountType, projectId, privateKeyId, privateKey, clientEmail, clientId, authUri, tokenUri, authProviderX509CertUrl, clientX509CertUrl string
}

func (conn *Connection) GetServiceAccountKey() *ServiceAccountKey {
	return &ServiceAccountKey{conn.serviceAccountInfo}
}

func (sak *ServiceAccountKey) GetAccountType() string {
	return sak.serviceAccountKey.accountType
}

func (sak *ServiceAccountKey) GetProjectId() string {
	return sak.serviceAccountKey.projectId
}

func (sak *ServiceAccountKey) GetPrivateKeyId() string {
	return sak.serviceAccountKey.privateKeyId
}

func (sak *ServiceAccountKey) GetPrivateKey() string {
	return sak.serviceAccountKey.privateKey
}

func (sak *ServiceAccountKey) GetClientEmailId() string {
	return sak.serviceAccountKey.clientEmail
}

func (sak *ServiceAccountKey) GetClientId() string {
	return sak.serviceAccountKey.clientId
}

func (sak *ServiceAccountKey) GetAuthUri() string {
	return sak.serviceAccountKey.authUri
}

func (sak *ServiceAccountKey) GetTokenUri() string {
	return sak.serviceAccountKey.tokenUri
}

func (sak *ServiceAccountKey) GetAuthProviderX509CertUrl() string {
	return sak.serviceAccountKey.authProviderX509CertUrl
}

func (sak *ServiceAccountKey) GetClientX509CertUrl() string {
	return sak.serviceAccountKey.clientX509CertUrl
}

func NewConnection(connectionObject interface{}) (*Connection, error) {

	genericConn, err := generic.NewConnection(connectionObject)
	if err != nil {
		return nil, err
	}

	if genericConn.GetSetting("seviceaccountkey") == nil {
		return nil, errors.New("Invalid Google connection. Missing Service Account Key configuration.")
	}

	serviceAcctKeyStringVal, _ := genericConn.GetSetting("seviceaccountkey").(string)
	if serviceAcctKeyStringVal == "" {
		return nil, errors.New("Invalid Google connection. Missing Service Account Key configuration.")
	}

	// Remove encoding
	index := strings.IndexAny(serviceAcctKeyStringVal, ",")
	if index > -1 {
		serviceAcctKeyStringVal = serviceAcctKeyStringVal[index+1:]
	}

	//decode
	decodedServiceAcctKey, err := base64.StdEncoding.DecodeString(serviceAcctKeyStringVal)

	if err != nil {
		return nil, err
	}
	svcAccountKey := new(serviceAccountInfo)
	serviceAcctKeyValue, err := coerce.ToObject(string(decodedServiceAcctKey))
	if err != nil {
		return nil, err
	}
	svcAccountKey.accountType = serviceAcctKeyValue["type"].(string)
	svcAccountKey.projectId = serviceAcctKeyValue["project_id"].(string)
	svcAccountKey.privateKeyId = serviceAcctKeyValue["private_key_id"].(string)
	svcAccountKey.privateKey = serviceAcctKeyValue["private_key"].(string)
	svcAccountKey.clientEmail = serviceAcctKeyValue["client_email"].(string)
	svcAccountKey.clientId = serviceAcctKeyValue["client_id"].(string)
	svcAccountKey.authUri = serviceAcctKeyValue["auth_uri"].(string)
	svcAccountKey.tokenUri = serviceAcctKeyValue["token_uri"].(string)
	svcAccountKey.authProviderX509CertUrl = serviceAcctKeyValue["auth_provider_x509_cert_url"].(string)
	svcAccountKey.clientX509CertUrl = serviceAcctKeyValue["client_x509_cert_url"].(string)
	return &Connection{genericConn, svcAccountKey}, nil
}
