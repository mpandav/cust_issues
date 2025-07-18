package consul

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"

	"github.com/hashicorp/consul/api"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"
	"github.com/tibco/wi-contrib/integrations"
)

var consulClient *api.Client

var configKVConsul *consulSvcConfig

var consulLoggger = log.ChildLogger(log.RootLogger(), "app-props-consul-resolver")

const (
	ConsulConfigKey = "FLOGO_APP_PROPS_CONSUL"
	ResolverName    = "consul"
)

func init() {

	consulFile := getConsulKVConfiguration()

	if consulFile != "" {
		if strings.HasSuffix(consulFile, ".json") {
			configKVConsul = consulConfigFromJSON(consulFile)
		} else if strings.HasPrefix(consulFile, "{") {
			configKVConsul = consulConfigFromInlineJSON(consulFile)
		} else {
			errMsg := fmt.Sprintf("Invalid value set for %s variable. It must be a valid JSON or key/value pair. See documentation for more details.", ConsulConfigKey)
			consulLoggger.Error(errMsg)
			panic("")
		}

		if configKVConsul != nil {
			property.RegisterPropertyResolver(&ConsulValueResolver{})
			envProp := os.Getenv(engine.EnvAppPropertyResolvers)
			if envProp == "" {
				//Make consul resolver default since FLOGO_APP_PROPS_CONSUL_KVSTORE_CONFIG is set
				os.Setenv(engine.EnvAppPropertyResolvers, ResolverName)
			} else if envProp == dynamicprops.ResolverName {
				//If only dynamic property resolver is enabled append consul after it
				os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", dynamicprops.ResolverName, ResolverName))
			}
			consulLoggger.Debug("Consul key/value store resolver registered")
		} else {
			consulLoggger.Error("Failed to read Consul Key/Value store configuration from JSON. See logs for more details.")
			panic("")
		}
	}
}

type consulSvcConfig struct {
	Address            string `json:"server_address"`
	CAFile             string `json:"ca_file"`
	CertFile           string `json:"cert_file"`
	CAPath             string `json:"ca_path"`
	KeyFile            string `json:"key_file"`
	Token              string `json:"acl_token"`
	KeyPrefix          string `json:"key_prefix"`
	InsecureConnection string `json:"insecure_connection"`
}

func getConsulKVConfiguration() string {
	key := os.Getenv(ConsulConfigKey)
	if len(key) > 0 {
		return key
	}
	return ""
}

func createClient() error {
	if configKVConsul != nil {
		consulConfig := api.DefaultConfig()
		consulConfig.Address = configKVConsul.Address
		consulConfig.TLSConfig.CAFile = configKVConsul.CAFile
		consulConfig.TLSConfig.CertFile = configKVConsul.CertFile
		consulConfig.TLSConfig.KeyFile = configKVConsul.KeyFile
		skiSsl, _ := strconv.ParseBool(configKVConsul.InsecureConnection)
		consulConfig.TLSConfig.InsecureSkipVerify = skiSsl

		consulConfig.Token = integrations.DecryptIfEncrypted(configKVConsul.Token)

		configKVConsul.KeyPrefix = integrations.SubstituteTemplate(configKVConsul.KeyPrefix)

		client, err := api.NewClient(consulConfig)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to create Consul client due to error - '%v'", err)
			return errors.New(errMsg)
		}
		consulClient = client
	}
	return nil
}

func consulConfigFromJSON(consulFile string) *consulSvcConfig {

	var configConsul consulSvcConfig
	file, err1 := ioutil.ReadFile(consulFile)
	if err1 == nil {
		err2 := json.Unmarshal(file, &configConsul)
		if err2 != nil {
			consulLoggger.Errorf("Error - '%v' occurred while parsing Consul configuration JSON file", err2)
			return nil
		}
	} else {
		consulLoggger.Errorf("Error - '%v' occurred while reading Consul configuration JSON file", err1)
		return nil
	}
	return &configConsul
}

func consulConfigFromInlineJSON(consulFile string) *consulSvcConfig {

	var configConsul consulSvcConfig
	err2 := json.Unmarshal([]byte(consulFile), &configConsul)
	if err2 != nil {
		consulLoggger.Errorf("Error - '%v' occurred while parsing Consul configuration JSON", err2)
		return nil
	}
	return &configConsul
}

func checkKnownErrors(err error) {
	// TODO Due to lack of proper error reporting by Consul, using string match
	if strings.Contains(err.Error(), "connect: connection refused") {
		consulLoggger.Error("Unable to connect to Consul server. Make sure server is running and reachable.")
		panic("")
	}

	if strings.Contains(err.Error(), "Unexpected response code: 403") {
		consulLoggger.Error("Unauthorized access to key/value store. Make sure correct token is specified in the configuration.")
		panic("")
	}
}

type ConsulValueResolver struct {
}

var prelodedConsulProps = make(map[string]interface{})

func preloadConsulProps() {
	consulLoggger.Debugf("Loading keys from path - %s", configKVConsul.KeyPrefix)
	pairs, _, err := consulClient.KV().List(configKVConsul.KeyPrefix, nil)
	if err != nil {
		checkKnownErrors(err)
		consulLoggger.Warnf("Value lookup for prefix - '%s' is not successful due to error - '%v'", configKVConsul.KeyPrefix, err)
	}
	if pairs != nil {
		for _, pair := range pairs {
			prelodedConsulProps["/"+pair.Key] = string(pair.Value)
		}
	}
}

func (resolver *ConsulValueResolver) Name() string {
	return ResolverName
}

func (resolver *ConsulValueResolver) LookupValue(toResolve string) (interface{}, bool) {

	if consulClient == nil {
		err := createClient()
		if err != nil {
			consulLoggger.Error(err.Error())
			panic("")
		}

		if len(configKVConsul.KeyPrefix) > 0 {
			// preload props from given prefix
			preloadConsulProps()
		}
	}

	if strings.Contains(toResolve, ".") {
		// replace . with /
		toResolve = strings.Replace(toResolve, ".", "/", -1)
	}

	aPath := integrations.SubstituteTemplate(toResolve)
	if len(configKVConsul.KeyPrefix) > 0 {
		aPath = configKVConsul.KeyPrefix + "/" + aPath
	}

	if !strings.HasPrefix(aPath, "/") {
		aPath = "/" + aPath
	}

	consulLoggger.Debugf("Resolving key - %s", aPath)

	if len(prelodedConsulProps) > 0 {
		// do preload lookup
		value, ok := prelodedConsulProps[aPath]
		return value, ok
	} else {
		// Call Consul service
		pair, _, err := consulClient.KV().Get(aPath, nil)
		if err != nil {
			checkKnownErrors(err)
			consulLoggger.Warnf("key - '%s' lookup is not successful due to error - '%v'", aPath, err)
			return nil, false
		}
		if pair != nil {
			consulLoggger.Debugf("Key - %s found in the param store", aPath)
			return string(pair.Value), true
		}
	}
	return nil, false
}
