package aws

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
)

var awsLog = log.ChildLogger(log.RootLogger(), "aws-connection-log")

const (
	DEFAULT_CREDENTIALS = "Default Credentials"
	AWS_CREDENTIALS     = "AWS Credentials"
)

type Connection struct {
	*generic.Connection
	*awsConnection
}

type awsConnection struct {
	region, accessKey, secretKey         string
	assumeRole                           bool
	roleArn, roleSessionName, externalID string
	sessionToken                         string
	authenticationType                   string
	expirationDuration                   time.Duration
}

func (conn *Connection) GetRegion() string {
	return conn.awsConnection.region
}
func (conn *Connection) GetAccessKey() string {
	return conn.awsConnection.accessKey
}
func (conn *Connection) GetSecretKey() string {
	return conn.awsConnection.secretKey
}
func (conn *Connection) GetSessionToken() string {
	return conn.awsConnection.sessionToken
}
func (conn *Connection) GetAuthenticationType() string {
	return conn.awsConnection.authenticationType
}

func NewConnection(connectionObject interface{}) (*Connection, error) {

	genericConn, err := generic.NewConnection(connectionObject)
	if err != nil {
		return nil, err
	}

	awsCong := new(awsConnection)
	accessKey := genericConn.GetSetting("accessKey")
	secretKey := genericConn.GetSetting("secretKey")
	sessionToken := genericConn.GetSetting("sessionToken")
	authenticationType, _ := genericConn.GetSetting("authenticationType").(string)

	region := genericConn.GetSetting("region")
	if region == nil {
		return nil, errors.New("Invalid AWS connection. Missing region configuration.")
	}
	awsCong.region = region.(string)

	if authenticationType == DEFAULT_CREDENTIALS {
		awsLog.Infof("Using default AWS credential provider chain")
		awsCong.authenticationType = DEFAULT_CREDENTIALS
		return &Connection{genericConn, awsCong}, nil
	} else {
		awsLog.Infof("Using authentication type as AWS Credentials")
		awsCong.authenticationType = AWS_CREDENTIALS
	}

	if accessKey == nil || secretKey == nil {
		return nil, errors.New("Invalid AWS connection. Missing accessKey or secretKey configuration.")
	}
	awsCong.accessKey = accessKey.(string)
	awsCong.secretKey = secretKey.(string)

	if sessionToken != nil {
		awsCong.sessionToken = sessionToken.(string)
	}

	var assumeRole bool
	assumeRoleValue := genericConn.GetSetting("assumeRole")
	if assumeRoleValue != nil {
		assumeRole, err = coerce.ToBool(assumeRoleValue)
		if err != nil {
			assumeRole = false
		}
	}
	if assumeRole {
		awsCong.assumeRole = true
		roleArn := genericConn.GetSetting("roleArn")
		roleSessionName := genericConn.GetSetting("roleSessionName")
		externalId := genericConn.GetSetting("externalId")
		expirationDuration := genericConn.GetSetting("expirationDuration")

		if roleArn != nil {
			s, _ := coerce.ToString(roleArn)
			awsCong.roleArn = s
		}

		if roleSessionName != nil {
			s, _ := coerce.ToString(roleSessionName)
			awsCong.roleSessionName = s
		}

		if externalId != nil {
			s, _ := coerce.ToString(externalId)
			awsCong.externalID = s
		}

		if expirationDuration != nil {
			expire, err := coerce.ToInt(expirationDuration)
			if err != nil {
				return nil, err
			}
			awsCong.expirationDuration = time.Duration(expire) * time.Second

		}
	}
	return &Connection{genericConn, awsCong}, nil
}

func (c *Connection) GetAssumeConfig(session *session.Session) *aws.Config {
	conf := &aws.Config{Region: aws.String(c.region)}
	if c.authenticationType == DEFAULT_CREDENTIALS {
		return conf
	}
	conf.Credentials = credentials.NewStaticCredentials(c.GetAccessKey(), c.GetSecretKey(), c.GetSessionToken())

	if c.assumeRole {
		creds := stscreds.NewCredentials(session, c.roleArn, func(p *stscreds.AssumeRoleProvider) {
			p.ExternalID = aws.String(c.externalID)
			p.RoleSessionName = c.roleSessionName
			p.Duration = c.expirationDuration
		})
		conf.Credentials = creds
	}
	return conf
}

func (c *Connection) GetConfig() *aws.Config {
	conf := &aws.Config{Region: aws.String(c.region)}
	if c.authenticationType == DEFAULT_CREDENTIALS {
		return conf
	}
	conf.Credentials = credentials.NewStaticCredentials(c.GetAccessKey(), c.GetSecretKey(), c.GetSessionToken())
	return conf
}

func (c *Connection) NewSession() *session.Session {
	sess := session.Must(session.NewSession(c.GetConfig()))
	if c.assumeRole {
		awsLog.Infof("Enabled Assume Role for connection [%s]", c.GetName())
		sess.Config.Credentials = stscreds.NewCredentials(sess, c.roleArn, func(p *stscreds.AssumeRoleProvider) {
			if len(c.externalID) > 0 {
				p.ExternalID = aws.String(c.externalID)
			}

			if len(c.roleSessionName) > 0 {
				p.RoleSessionName = c.roleSessionName
			}
			p.Duration = c.expirationDuration
		})
	}
	return sess
}
