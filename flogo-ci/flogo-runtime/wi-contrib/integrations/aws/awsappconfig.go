package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appconfigdata"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"
	"github.com/tibco/wi-contrib/integrations"
	"gopkg.in/yaml.v2"
)

var appConfigData *appconfigdata.AppConfigData = nil
var appconfigLogger = log.ChildLogger(log.RootLogger(), "appprops.awsappconfig.resolver")
var appConfig *awsAppConfig
var preloadedkv = make(map[string]interface{})
var accessDeniedError = fmt.Sprintf("Access denied to appconfig store. Ensure that required policy is configured for IAM role.")

// AWSAppConfigKey must be set to true to enable this feature
const (
	AWSAppConfigKey = "FLOGO_APP_PROPS_AWS_APPCONFIG" // set to true to use it

	EnvAWSAccessKeyID     = "AWS_APPCONFIG_ACCESS_KEY_ID"
	EnvAWSSecretAccessKey = "AWS_APPCONFIG_SECRET_ACCESS_KEY"
	EnvAWSSessionToken    = "AWS_APPCONFIG_SESSION_TOKEN"
	EnvAppConfigRegion    = "AWS_APPCONFIG_REGION"

	EnvAppConfigAppName     = "AWS_APPCONFIG_APP_IDENTIFIER_NAME"
	EnvAppConfigProfileName = "AWS_APPCONFIG_PROFILE_NAME"
	EnvAppConfigEnvName     = "AWS_APPCONFIG_ENV_NAME"

	EnvAWSAssumeRoleARN   = "AWS_APPCONFIG_ASSUMEDROLE_ARN"
	EnvAWSRoleSessionName = "AWS_APPCONFIG_ROLESESSION_NAME"
	EnvAWSExternalID      = "AWS_APPCONFIG_EXTERNAL_ID"
)

const (
	AppConfigResolverName = "awsappconfig"
)

// AppConfigValueResolver ...
type AppConfigValueResolver struct {
}

func useAWSAppConfiguration() bool {
	key := os.Getenv(AWSAppConfigKey)
	val, err := coerce.ToBool(key)
	if err == nil {
		return val
	}
	return false
}

type awsAppConfig struct {
	AccessKeyID     string // AWS_APPCONFIG_ACCESS_KEY_ID
	SecretAccessKey string // AWS_APPCONFIG_SECRET_ACCESS_KEY
	SessionToken    string // AWS_APPCONFIG_SESSION_TOKEN
	Region          string // AWS_APPCONFIG_REGION

	AppName        string // AWS_APPCONFIG_APP_IDENTIFIER_NAME , if not set then get from engine
	ConfigProfName string // AWS_APPCONFIG_PROFILE_NAME
	EnvName        string // AWS_APPCONFIG_ENV_NAME

	AssumeRoleARN   string // AWS_APPCONFIG_ASSUMEDROLE_ARN [aws:uid:arn::] // follow aws connection
	RoleSessionName string //AWS_APPCONFIG_ROLESESSION_NAME
	ExternalID      string //AWS_APPCONFIG_EXTERNAL_ID
}

func init() {
	if !useAWSAppConfiguration() {
		return
	}
	// populate from env var
	appConfig = &awsAppConfig{}
	appConfig.AccessKeyID = os.Getenv(EnvAWSAccessKeyID)
	appConfig.SecretAccessKey = os.Getenv(EnvAWSSecretAccessKey)
	appConfig.SessionToken = os.Getenv(EnvAWSSessionToken)
	appConfig.Region = os.Getenv(EnvAppConfigRegion)

	appConfig.ConfigProfName = os.Getenv(EnvAppConfigProfileName)
	appConfig.EnvName = os.Getenv(EnvAppConfigEnvName)

	appConfig.AssumeRoleARN = os.Getenv(EnvAWSAssumeRoleARN)
	appConfig.RoleSessionName = os.Getenv(EnvAWSRoleSessionName)
	appConfig.ExternalID = os.Getenv(EnvAWSExternalID)

	if appConfig.ConfigProfName == "" || appConfig.EnvName == "" {
		err := fmt.Errorf("Required environment variables %s and/or %s not configured", EnvAppConfigProfileName, EnvAppConfigEnvName)
		appconfigLogger.Error(err)
		panic(err)
	}

	property.RegisterPropertyResolver(&AppConfigValueResolver{})
	envProp := os.Getenv(engine.EnvAppPropertyResolvers)
	if envProp == "" {
		//Make awsappconfig resolver default since FLOGO_APPCONFIG_PROPS_AWS is set
		os.Setenv(engine.EnvAppPropertyResolvers, AppConfigResolverName)
	} else if envProp == dynamicprops.ResolverName {
		//If only dynamic property resolver is enabled append awsappconfig resolver after it
		os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", dynamicprops.ResolverName, AppConfigResolverName))
	}
	appconfigLogger.Debug("AWS appconfig resolver registered")
}

func createAppConfigClient() error {
	appConfig.AppName = os.Getenv(EnvAppConfigAppName)
	if appConfig.AppName == "" {
		appConfig.AppName = engine.GetAppName()
	}

	if appConfig.Region == "" {
		// Lookup region from the environment
		appConfig.Region = os.Getenv("AWS_REGION")
		if appConfig.Region == "" {
			// Use metadata API
			appconfigLogger.Infof("Region is not set, trying to get aws region from EC2 metadata API")
			region, err := getRegionFromEnv(session.Must(session.NewSession()))
			if err != nil {
				return errors.New("failed to find region using metadata API, 'Region' must be configured as environment variable")
			}
			appConfig.Region = region
		}
		appconfigLogger.Infof("Region set to - '%s'", appConfig.Region)
	}

	appConfig.AccessKeyID = integrations.DecryptIfEncrypted(appConfig.AccessKeyID)
	appConfig.SecretAccessKey = integrations.DecryptIfEncrypted(appConfig.SecretAccessKey)
	appConfig.SessionToken = integrations.DecryptIfEncrypted(appConfig.SessionToken)

	conf := &aws.Config{Region: aws.String(appConfig.Region)}

	if appConfig.AccessKeyID != "" && appConfig.SecretAccessKey != "" {
		conf.Credentials = credentials.NewStaticCredentials(appConfig.AccessKeyID, appConfig.SecretAccessKey, appConfig.SessionToken)
	} else {
		appconfigLogger.Info("AccessKeyID or SecretAccessKey not provided, using the default credential provider chain to get the AWS credentials")
	}

	sess := session.Must(session.NewSession(conf))
	if strings.TrimSpace(appConfig.AssumeRoleARN) != "" {
		appconfigLogger.Debug("Assume role enabled for AWS app config resolver")
		sess.Config.Credentials = stscreds.NewCredentials(sess, appConfig.AssumeRoleARN, func(p *stscreds.AssumeRoleProvider) {
			if len(appConfig.ExternalID) > 0 {
				p.ExternalID = aws.String(appConfig.ExternalID)
			}
			if len(appConfig.RoleSessionName) > 0 {
				p.RoleSessionName = appConfig.RoleSessionName
			}
		})
	}
	appConfigData = appconfigdata.New(sess)
	if appConfigData != nil {
		appconfigLogger.Debug("AppConfig client successfully created ")
	}
	return nil
}

func preloadValues() {
	op, err := appConfigData.StartConfigurationSession(&appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:          aws.String(appConfig.AppName),
		ConfigurationProfileIdentifier: aws.String(appConfig.ConfigProfName),
		EnvironmentIdentifier:          aws.String(appConfig.EnvName),
	})
	if err != nil {
		appconfigLogger.Error(err.Error())
		panic(err)
	}

	outPut, err := appConfigData.GetLatestConfiguration(&appconfigdata.GetLatestConfigurationInput{ConfigurationToken: op.InitialConfigurationToken})
	if err != nil {
		appconfigLogger.Error(err.Error())
		panic(err)
	}
	var appcfgvalMap map[string]interface{}
	if outPut.Configuration != nil {
		if strings.EqualFold(*outPut.ContentType, "application/json") {
			err := json.Unmarshal(outPut.Configuration, &appcfgvalMap)
			if err != nil {
				appconfigLogger.Error(err.Error())
				panic("Error while deserializing appconfig data")
			}
		} else if strings.EqualFold(*outPut.ContentType, "application/x-yaml") {
			var dataMap map[interface{}]interface{}
			err := yaml.Unmarshal(outPut.Configuration, &dataMap)
			if err != nil {
				appconfigLogger.Error(err.Error())
				panic("Error while deserializing appconfig data")
			} else {
				appcfgvalMap = make(map[string]interface{})
				flatYamlMap(dataMap, "", &appcfgvalMap)
			}
		} else if strings.EqualFold(*outPut.ContentType, "text/plain") {
			appcfgvalMap = make(map[string]interface{})
			lines := strings.Split(string(outPut.Configuration), "\n")
			for _, row := range lines {
				if strings.Contains(row, "=") {
					kv := strings.Split(row, "=")
					if len(kv) == 2 {
						appcfgvalMap[kv[0]] = kv[1]
					}
				}
			}
		}
	}
	if len(appcfgvalMap) > 0 {
		for k, v := range appcfgvalMap {
			preloadedkv[k] = v
		}
	}
}

// Name ...
func (resolver *AppConfigValueResolver) Name() string {
	return AppConfigResolverName
}

// LookupValue ...
func (resolver *AppConfigValueResolver) LookupValue(toResolve string) (interface{}, bool) {
	if appConfigData == nil {
		err := createAppConfigClient()
		if err != nil {
			appconfigLogger.Error(err.Error())
			panic("")
		}
	}
	if len(preloadedkv) == 0 {
		preloadValues()
	}
	value, ok := preloadedkv[toResolve]
	if ok {
		appconfigLogger.Debugf("Parameter - %s found in AWS appConfig", toResolve)
		return value, true
	}
	return nil, false
}

func flatYamlMap(in map[interface{}]interface{}, prefix string, res *map[string]interface{}) {
	for k, v := range in {
		switch v.(type) {
		case map[interface{}]interface{}:
			if prefix != "" {
				prefix = prefix + "." + k.(string)
			} else {
				prefix = k.(string)
			}
			flatYamlMap(v.(map[interface{}]interface{}), prefix, res)
			prefix = ""
		case []interface{}:
			if prefix != "" {
				prefix = prefix + "." + k.(string)
			} else {
				prefix = k.(string)
			}
			for _, v1 := range v.([]interface{}) {
				flatYamlMap(v1.(map[interface{}]interface{}), prefix, res)
			}
			prefix = ""
		default:
			var keytoenter string
			if prefix != "" {
				keytoenter = prefix + "." + k.(string)
			} else {
				keytoenter = k.(string)
			}
			(*res)[keytoenter] = v
		}
	}
	prefix = ""
}
