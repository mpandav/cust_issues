package springcloud

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"
)

var springCloudLogger = log.ChildLogger(log.RootLogger(), "app-props-spring-resolver")
var configSpringCloud *springCloudConfig
var prelodedKV = make(map[string]interface{})
var insidePCF = ""

const (
	SpringCloudConfigKey = "FLOGO_APP_PROPS_SPRING_CLOUD"
	ResolverName         = "springCloud"
)

func init() {
	insidePCF = os.Getenv("VCAP_SERVICES")
	springCloudFile := getSpringCloudConfigurationKey()
	if springCloudFile != "" {

		if strings.HasSuffix(springCloudFile, ".json") {
			configSpringCloud = configFromJSON(springCloudFile)
		} else if strings.HasPrefix(springCloudFile, "{") {
			configSpringCloud = configFromInlineJSON(springCloudFile)
		} else {
			errMsg := fmt.Sprintf("Invalid value set for %s variable. It must be a valid JSON. See documentation for more details.", SpringCloudConfigKey)
			springCloudLogger.Error(errMsg)
			panic("")
		}

		if configSpringCloud != nil {
			property.RegisterPropertyResolver(&SpringCloudValueResolver{})

			envProp := os.Getenv(engine.EnvAppPropertyResolvers)
			if envProp == "" {
				//Make consul resolver default since FLOGO_APP_PROPS_CONSUL_KVSTORE_CONFIG is set
				os.Setenv(engine.EnvAppPropertyResolvers, ResolverName)
			} else if envProp == dynamicprops.ResolverName {
				//If only dynamic property resolver is enabled append consul after it
				os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", dynamicprops.ResolverName, ResolverName))
			}

			springCloudLogger.Debug("Spring Cloud Config resolver registered")

		} else {
			springCloudLogger.Error("Failed to read Spring Cloud configuration from JSON. See logs for more details.")
			panic("")
		}
	}
}

type SpringCloudValueResolver struct {
}

type springCloudConfig struct {
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
	Profile  string `json:"profile"`
	AppID    string `json:"app_id"`
}

func getSpringCloudConfigurationKey() string {
	key := os.Getenv(SpringCloudConfigKey)
	if len(key) > 0 {
		return key
	}
	return ""
}

func configFromJSON(configFile string) *springCloudConfig {

	var configSpring springCloudConfig
	file, err1 := ioutil.ReadFile(configFile)
	if err1 == nil {

		err2 := json.Unmarshal(file, &configSpring)
		if err2 != nil {
			springCloudLogger.Errorf("Error - '%v' occurred while parsing Spring Cloud configuration JSON", err2)
			return nil
		}

		if configSpring.Profile == "" && insidePCF != "" {
			springCloudLogger.Errorf("Error - '%v' no profile found while parsing Spring Cloud configuration JSON", err2)
			return nil
		}

	} else {
		springCloudLogger.Errorf("Error - '%v' occurred while reading Spring Cloud configuration JSON file", err1)
		return nil
	}
	return &configSpring
}

func configFromInlineJSON(configFile string) *springCloudConfig {
	var configSpring springCloudConfig

	err2 := json.Unmarshal([]byte(configFile), &configSpring)
	if err2 != nil {
		springCloudLogger.Errorf("Error - '%v' occurred while parsing Spring Cloud configuration JSON", err2)
		return nil
	}

	if configSpring.Profile == "" && insidePCF != "" {
		springCloudLogger.Errorf("Error - '%v' no profile found while parsing Spring Cloud configuration JSON", err2)
		return nil
	}

	return &configSpring
}

type KV map[string]string

type propSource struct {
	Name   string
	Source KV
}

type spring struct {
	Name            string
	Profiles        []string
	Label           string
	Version         string
	State           string
	PropertySources []propSource
}

type tkn struct {
	Access_token string
	Token_type   string
	Expiry       string
	Scope        string
	Jti          string
}

type config_server struct {
	P_Config_Server []interface{} `json:"p-config-server"`
}

func preloadKV(cre map[string]interface{}) error {

	if cre != nil {

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

		access_token_uri := cre["access_token_uri"].(string)
		client_id := cre["client_id"].(string)
		client_secret := cre["client_secret"].(string)
		auth_uri := cre["uri"].(string)

		OAuthToken := access_token_uri + "?client_id=" + client_id + "&grant_type=client_credentials&client_secret=" + client_secret + "&token_format=jwt"

		accessToken, err := client.Post(OAuthToken, "application/x-www-form-urlencoded", nil)

		if err != nil {
			return err
		}

		data, _ := ioutil.ReadAll(accessToken.Body)

		var authResult tkn
		err = json.Unmarshal(data, &authResult)

		if err != nil {
			return err
		}

		OAuthKV := auth_uri + "/" + engine.GetAppName() + "/" + configSpringCloud.Profile + "?access_token=" + authResult.Access_token

		resp, err := client.Get(OAuthKV)

		if err == nil {
			data, _ := ioutil.ReadAll(resp.Body)

			var s spring
			err := json.Unmarshal(data, &s)
			if err == nil {
				for i := 0; i <= len(s.PropertySources)-1; i++ {
					for k, v := range s.PropertySources[i].Source {

						_, con := prelodedKV[k]
						if !con {
							prelodedKV[k] = v
						}
					}
				}
			} else {
				return err
			}
		} else {
			return err
		}

		springCloudLogger.Debug("Configuration successfully loaded from Spring Cloud config service ")

		return nil
	}
	return errors.New("Required credentials are not set.")
}

/*func createSpringCloudClient() error {
	if configSpringCloud != nil {
		cfClient := &cfclient.Config{
			ApiAddress:        configSpringCloud.Address,
			Username:          configSpringCloud.Username,
			Password:          configSpringCloud.Password,
			SkipSslValidation: true,
		}
		client, err := cfclient.NewClient(cfClient)
		if err != nil {
			fmt.Println("Error 1: ", err)
		}
		appl, err := client.GetAppEnv(configSpringCloud.AppID)
		if err != nil {
			fmt.Println(err)
		}

		m := appl.SystemEnv

		cred := m["VCAP_SERVICES"].(map[string]interface{})["p-config-server"].([]interface{})[0].(map[string]interface{})["credentials"].(map[string]interface{})

		preloadKV(cred)
	}
	return nil
}*/

func createSpringCloudClientInsidePCF() error {

	var obj config_server
	var cred map[string]interface{} = nil

	err := json.Unmarshal([]byte(insidePCF), &obj)

	if err != nil {
		return err
	}

	for i := 0; i < len(obj.P_Config_Server); i++ {

		m := obj.P_Config_Server[i].(map[string]interface{})
		tags := m["tags"].([]interface{})

		var tagString string
		if len(tags) == 2 {
			tagString += tags[0].(string)
			tagString += " "
			tagString += tags[1].(string)
		}

		if strings.Contains(tagString, "configuration") && strings.Contains(tagString, "spring-cloud") {
			cred = m["credentials"].(map[string]interface{})
			break
		}
	}

	err = preloadKV(cred)

	if err != nil {
		return err
	}

	return nil
}

func (resolver *SpringCloudValueResolver) Name() string {
	return ResolverName
}

func (resolver *SpringCloudValueResolver) LookupValue(toResolve string) (interface{}, bool) {

	if len(prelodedKV) == 0 {
		if insidePCF != "" {
			err := createSpringCloudClientInsidePCF()
			if err != nil {
				springCloudLogger.Error(err.Error())
				panic("")
			}
		}
	}

	value, ok := prelodedKV[toResolve]
	if ok {
		return value, true
	} else {
		return value, false
	}

}
