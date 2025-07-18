package pubsub

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"google.golang.org/api/option"
)

var logger = log.ChildLogger(log.RootLogger(), "googlecloudpubsub.connection")
var factory = &GooglePubSubConnectionFactory{}

type GooglePubSubConnectionFactory struct {
}

type GooglePubSubConnectionManager struct {
	client *pubsub.Client
}

type Settings struct {
	Name              string      `md:"name,required"`
	ServiceAccountKay interface{} `md:"serviceAccountKey,required"`
	ProjectId         string      `md:"projectId,required"`
}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

func (g GooglePubSubConnectionFactory) Type() string {
	return "GooglePubSub"
}

func (g GooglePubSubConnectionFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	connectionManager := &GooglePubSubConnectionManager{}
	s := &Settings{}
	var err error
	err = metadata.MapToStruct(settings, s, true)
	if err != nil {
		return nil, err
	}
	logger.Infof("Reading configuration for the connection [%s]", s.Name)
	serviceAcctKeyValAsObject, err := coerce.ToParams(s.ServiceAccountKay)
	var serviceAcctKeyValAsString string
	if err == nil {
		serviceAcctKeyValAsString = serviceAcctKeyValAsObject["content"]
	} else {
		serviceAcctKeyValAsString, _ = s.ServiceAccountKay.(string)
	}

	if serviceAcctKeyValAsString == "" {
		logger.Errorf("Service Account Key is not set. Check connection [%s] configuration.", s.Name)
		return nil, errors.New("service account key not set")
	}
	// Remove encoding
	index := strings.IndexAny(serviceAcctKeyValAsString, ",")
	if index > -1 {
		serviceAcctKeyValAsString = serviceAcctKeyValAsString[index+1:]
	}
	//decode
	sa, err := base64.StdEncoding.DecodeString(serviceAcctKeyValAsString)
	if err != nil {
		return nil, err
	}

	logger.Infof("Google Project set to '%s'", s.ProjectId)
	connectionManager.client, err = pubsub.NewClient(context.Background(), s.ProjectId, option.WithCredentialsJSON(sa))
	if err != nil {
		return nil, err
	}
	logger.Infof("Connection [%s] successfully created", s.Name)
	return connectionManager, nil
}

func (cm *GooglePubSubConnectionManager) Type() string {
	return "GooglePubSub"
}

func (cm *GooglePubSubConnectionManager) GetConnection() interface{} {
	return cm.client
}

func (cm *GooglePubSubConnectionManager) ReleaseConnection(connection interface{}) {

}

func (cm *GooglePubSubConnectionManager) Start() error {
	return nil
}

func (cm *GooglePubSubConnectionManager) Stop() error {
	return nil
}
