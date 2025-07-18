package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"
	"github.com/tibco/wi-contrib/integrations"

	"github.com/aws/aws-sdk-go/service/ssm"
)

var ssmc *ssm.SSM = nil

var paramstoreLogger = log.ChildLogger(log.RootLogger(), "appprops.awsparamstore.resolver")

var paramPrefix string

var configAWSParam *awsConfig
var prelodedkv = make(map[string]interface{})
var ErrAccessDenied = fmt.Sprintf("Access denied to param store. Ensure that below policy is configured for IAM role. \n %s", "{ \r\n   \"Version\":\"2012-10-17\",\r\n   \"Statement\":[ \r\n      { \r\n         \"Action\":[ \r\n            \"ssm:GetParameter\",\r\n            \"ssm:GetParametersByPath\",\r\n         ],\r\n         \"Effect\":\"Allow\",\r\n         \"Resource\":\"*\"\r\n      }\r\n   ]\r\n}")

const (
	AWSParamStoreConfigKey = "FLOGO_APP_PROPS_AWS"
	ParamStoreResolverName = "awsparamstore"
)

func init() {

	awsParamFile := getAWSParamConfigurationKey()
	if awsParamFile != "" {

		if strings.HasSuffix(awsParamFile, ".json") {
			configAWSParam = configFromJSON(awsParamFile)
		} else if strings.HasPrefix(awsParamFile, "{") {
			configAWSParam = configFromInlineJSON(awsParamFile)
		} else {
			errMsg := fmt.Sprintf("Invalid value set for %s variable. It must be a valid JSON or key/value pair. See documentation for more details.", AWSParamStoreConfigKey)
			paramstoreLogger.Error(errMsg)
			panic("")
		}

		if configAWSParam != nil {
			// Set resolver to param store

			property.RegisterPropertyResolver(&ParamStoreValueResolver{})
			envProp := os.Getenv(engine.EnvAppPropertyResolvers)
			if envProp == "" {
				//Make awsparamstore resolver default since FLOGO_APP_PROPS_AWS_PARAMSTORE_CONFIG is set
				os.Setenv(engine.EnvAppPropertyResolvers, ParamStoreResolverName)
			} else if envProp == dynamicprops.ResolverName {
				//If only dynamic property resolver is enabled append awsparamstore after it
				os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", dynamicprops.ResolverName, ParamStoreResolverName))
			}

			paramstoreLogger.Debug("AWS parameter store resolver registered")
		} else {
			paramstoreLogger.Error("Failed to read AWS param store configuration from JSON. See logs for more details.")
			panic("")
		}

	}
}

func createSSMClient() error {
	if configAWSParam != nil {
		sess := session.Must(session.NewSession())
		if sess == nil {
			return errors.New("Failed to create AWS session. Can not fetch paramter.")
		}

		awsConfig := &aws.Config{}

		if configAWSParam.Region == "" {
			// Lookup region from the environment
			configAWSParam.Region = os.Getenv("AWS_REGION")
			if configAWSParam.Region == "" {
				// Use metadata API
				region, err := getRegionFromEnv(sess)
				if err != nil {
					return errors.New("Failed to find region using metadata API. 'Region' must be set in the AWS configuration JSON.")
				}
				configAWSParam.Region = region
			}
			paramstoreLogger.Infof("Region set to - '%s'", configAWSParam.Region)
		}

		awsConfig.Region = &configAWSParam.Region
		if configAWSParam.UseExecutionEnv == false {
			if configAWSParam.AccessKeyID != "" && configAWSParam.SecretAccessKey != "" {
				var v credentials.Value
				v.AccessKeyID = integrations.DecryptIfEncrypted(configAWSParam.AccessKeyID)
				v.SecretAccessKey = integrations.DecryptIfEncrypted(configAWSParam.SecretAccessKey)
				v.SessionToken = integrations.DecryptIfEncrypted(configAWSParam.SessionToken)
				awsConfig.Credentials = credentials.NewStaticCredentialsFromCreds(v)
			} else {
				paramstoreLogger.Error("AccessKeyID and SecretAccessKey must be set in the configuration")
				panic("")
			}
		}

		ssmc = ssm.New(sess, awsConfig)
		if ssmc != nil {
			paramstoreLogger.Debug("SSM client successfully created ")
		}
		paramPrefix = integrations.SubstituteTemplate(configAWSParam.ParamPrefix)
	}
	return nil
}

func getRegionFromEnv(sess *session.Session) (string, error) {
	svc := ec2metadata.New(sess)
	return svc.Region()
}

type ParamStoreValueResolver struct {
}

type awsConfig struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Region          string `json:"region"`
	ParamPrefix     string `json:"param_prefix"`
	UseExecutionEnv bool   `json:"use_iam_role"`
	SessionToken    string `json:"session_token"`
}

func getAWSParamConfigurationKey() string {
	key := os.Getenv(AWSParamStoreConfigKey)
	if len(key) > 0 {
		return key
	}
	return ""
}

func configFromJSON(configFile string) *awsConfig {

	var configAws awsConfig
	file, err1 := ioutil.ReadFile(configFile)
	if err1 == nil {
		err2 := json.Unmarshal(file, &configAws)
		if err2 != nil {
			paramstoreLogger.Errorf("Error - '%v' occurred while parsing AWS configuration JSON file", err2)
			return nil
		}
	} else {
		paramstoreLogger.Errorf("Error - '%v' occurred while reading AWS configuration JSON file", err1)
		return nil
	}
	return &configAws
}

func configFromInlineJSON(consulFile string) *awsConfig {
	var configAws awsConfig

	err2 := json.Unmarshal([]byte(consulFile), &configAws)
	if err2 != nil {
		paramstoreLogger.Errorf("Error - '%v' occurred while parsing AWS configuration JSON", err2)
		return nil
	}
	return &configAws
}

func preload(token *string) {

	decryption := true
	//TODO recursive iteration could be costly in case of deep hierarchy
	recursive := true
	outPut, err := ssmc.GetParametersByPath(&ssm.GetParametersByPathInput{
		Path:           &paramPrefix,
		WithDecryption: &decryption,
		NextToken:      token,
		Recursive:      &recursive,
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == organizations.ErrCodeAccessDeniedException {
				paramstoreLogger.Error(ErrAccessDenied)
			} else {
				paramstoreLogger.Error(awsErr.Message())
			}
			panic("")
		} else {
			paramstoreLogger.Error(err.Error())
			panic("")
		}
	}

	for _, param := range outPut.Parameters {
		prelodedkv[*param.Name] = *param.Value
	}

	if outPut.NextToken != nil {
		// Load next set of token
		preload(outPut.NextToken)
	}
}

func (resolver *ParamStoreValueResolver) Name() string {
	return ParamStoreResolverName
}

func (resolver *ParamStoreValueResolver) LookupValue(toResolve string) (interface{}, bool) {

	if ssmc == nil {
		err := createSSMClient()
		if err != nil {
			paramstoreLogger.Error(err.Error())
			panic("")
		}

		if len(paramPrefix) > 0 {
			// preload props from given prefix
			paramstoreLogger.Debugf("Loading params from path - %s", paramPrefix)
			preload(nil)
		}
	}

	if strings.Contains(toResolve, ".") {
		// replace . with /
		toResolve = strings.Replace(toResolve, ".", "/", -1)
	}

	aPath := integrations.SubstituteTemplate(toResolve)
	if len(paramPrefix) > 0 {
		aPath = paramPrefix + "/" + aPath
	}

	if !strings.HasPrefix(aPath, "/") {
		aPath = "/" + aPath
	}

	paramstoreLogger.Debugf("Resolving Param: %s", aPath)

	if len(prelodedkv) > 0 {
		// do preload lookup
		value, ok := prelodedkv[aPath]
		return value, ok
	} else {
		decryption := true
		param, err := ssmc.GetParameter(&ssm.GetParameterInput{
			Name:           &aPath,
			WithDecryption: &decryption})

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				// panic if error other than param not found is returned
				if awsErr.Code() != ssm.ErrCodeParameterNotFound {
					if awsErr.Code() == organizations.ErrCodeAccessDeniedException {
						paramstoreLogger.Error(ErrAccessDenied)
					} else {
						paramstoreLogger.Error(awsErr.Message())
					}
					panic("")
				}
			}
			paramstoreLogger.Debugf("Param - '%s' lookup is not successful due to error - '%v'", aPath, err)
			return nil, false
		}

		if param != nil {
			paramstoreLogger.Debugf("Parameter - %s found in AWS param store", aPath)
			return *param.Parameter.Value, true
		}
	}
	return nil, false
}
