package connector

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
)

var logger = log.ChildLogger(log.RootLogger(), "aws-connection")
var factory = &awsFactory{}

const (
	DEFAULT_CREDENTIALS = "Default Credentials"
	AWS_CREDENTIALS     = "AWS Credentials"
)

type awsConnection struct {
	Name               string `md:"name"`
	Region             string `md:"region"`
	EnableEndpoint     bool   `md:"customEndpoint"`
	CustEndpoint       string `md:"endpoint"`
	AuthenticationType string `md:"authenticationType"`
	AccessKey          string `md:"accessKey"`
	SecretKey          string `md:"secretKey"`
	AssumeRole         bool   `md:"assumeRole"`
	RoleArn            string `md:"roleArn"`
	RoleSessionName    string `md:"roleSessionName"`
	ExternalID         string `md:"externalId"`
	ExpirationDuration int    `md:"expirationDuration"`
	SessionToken       string `md:"sessionToken"`
}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

type awsFactory struct {
}

func (*awsFactory) Type() string {
	return "aws"
}

func (*awsFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	awsManger := &awsManager{}
	var err error
	awsManger.config, err = getConnectionConfig(settings)
	if err != nil {
		return nil, err
	}
	awsManger.session = awsManger.NewSession()
	return awsManger, nil
}

type awsManager struct {
	config  *awsConnection
	session *session.Session
}

func (a *awsManager) Type() string {
	return "aws"
}

func (a *awsManager) GetConnection() interface{} {
	if a.session.Config.Endpoint != nil {
		//in case of custom endpoint create new session, as it will be modified at connector level
		return a.NewSession()
	}
	return a.session
}

func (a *awsManager) NewSession() *session.Session {
	sess := session.Must(session.NewSession(a.GetConfig()))
	if a.config.AssumeRole {
		logger.Infof("Enabled Assume Role for connection [%s]", a.config.Name)
		sess.Config.Credentials = stscreds.NewCredentials(sess, a.config.RoleArn, func(p *stscreds.AssumeRoleProvider) {
			if len(a.config.ExternalID) > 0 {
				p.ExternalID = aws.String(a.config.ExternalID)
			}
			p.RoleSessionName = a.config.RoleSessionName
			p.Duration = time.Duration(a.config.ExpirationDuration) * time.Second
		})
	}
	return sess
}

func (a *awsManager) GetConfig() *aws.Config {
	conf := &aws.Config{Region: aws.String(a.config.Region)}
	if a.config.EnableEndpoint && len(a.config.CustEndpoint) > 0 {
		conf.Endpoint = aws.String(a.config.CustEndpoint)
	}
	if a.config.AuthenticationType == DEFAULT_CREDENTIALS {
		logger.Infof("Using default AWS credential provider chain")
	} else {
		if a.config.AuthenticationType != AWS_CREDENTIALS {
			logger.Infof("Defaulting authentication type to AWS Credentials")
		}
		conf.Credentials = credentials.NewStaticCredentials(a.config.AccessKey, a.config.SecretKey, a.config.SessionToken)
	}
	return conf
}

func (a *awsManager) ReleaseConnection(connection interface{}) {
	//No nothing for aws connection
}

func (a *awsManager) Start() error {
	return nil
}

func (a *awsManager) Stop() error {
	return nil
}

//
//func NewConnectionManager(conn interface{}) (connection.Manager, error) {
//
//	var cManager connection.Manager
//	var err error
//	_, ok := conn.(map[string]interface{})
//	if ok {
//		cManager, err = handleLegacyConnection(conn)
//	} else {
//		cManager, err = coerce.ToConnection(conn)
//	}
//
//	if err != nil {
//		return nil, err
//	}
//
//	return cManager, nil
//}
//
//func handleLegacyConnection(conn interface{}) (connection.Manager, error) {
//
//	connectionObject, _ := coerce.ToObject(conn)
//	if connectionObject == nil {
//		return nil, errors.New("Connection object is nil")
//	}
//
//	id := connectionObject["id"].(string)
//
//	cManager := connection.GetManager(id)
//	if cManager == nil {
//
//		connObject, err := generic.NewConnection(connectionObject)
//		if err != nil {
//			return nil, err
//		}
//
//		cManager, err = factory.NewManager(connObject.Settings())
//		if err != nil {
//			return nil, err
//		}
//
//		err = connection.RegisterManager(id, cManager)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	return cManager, nil
//}

func getConnectionConfig(settings map[string]interface{}) (*awsConnection, error) {
	s := &awsConnection{}
	err := metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}
	return s, nil
}
