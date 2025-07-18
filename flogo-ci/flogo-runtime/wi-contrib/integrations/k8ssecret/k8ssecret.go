package k8ssecret

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"

	"errors"
)

var k8sSecretLoggger = log.ChildLogger(log.RootLogger(), "app-props-k8s-volume-resolver")

const (
	ConfigKey    = "FLOGO_APP_PROPS_K8S_VOLUME"
	ResolverName = "k8s-volume"
)

type k8sSecretResolverConfig struct {
	Volume string `json:"volume_path"`
}

var configK8sSecrets *k8sSecretResolverConfig

func init() {

	resolverConfig := getK8sSecretsResolverConfiguration()

	if resolverConfig != "" {
		if strings.HasSuffix(resolverConfig, ".json") {
			configK8sSecrets = k8sSecretsConfigFromJSON(resolverConfig)
		} else if strings.HasPrefix(resolverConfig, "{") {
			configK8sSecrets = k8sSecretsConfigFromInlineJSON(resolverConfig)
		} else {
			errMsg := fmt.Sprintf("Invalid value set for %s variable. It must be a valid JSON or key/value pair. See documentation for more details.", ConfigKey)
			k8sSecretLoggger.Error(errMsg)
			panic("")
		}

		if configK8sSecrets != nil {
			property.RegisterPropertyResolver(&K8sSecretResolver{})
			envProp := os.Getenv(engine.EnvAppPropertyResolvers)
			if envProp == "" {
				//Make K8s Secrets resolver default since FLOGO_APP_PROPS_K8S_SECRETS is set
				os.Setenv(engine.EnvAppPropertyResolvers, ResolverName)
			} else if envProp == dynamicprops.ResolverName {
				//If only dynamic property resolver is enabled append k8s-volume after it
				os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", dynamicprops.ResolverName, ResolverName))
			}
			k8sSecretLoggger.Debug("Resolver is registered")

			err := preloadPropsFromFile()
			if err != nil {
				errMsg := fmt.Sprintf("Failed to read properties from mounted volume due to error - %v", err)
				k8sSecretLoggger.Error(errMsg)
				panic("")
			}
		} else {
			k8sSecretLoggger.Error("Failed to read configuration from JSON. See logs for more details.")
			panic("")
		}
	}
}

func getK8sSecretsResolverConfiguration() string {
	key := os.Getenv(ConfigKey)
	if len(key) > 0 {
		return key
	}
	return ""
}

func k8sSecretsConfigFromJSON(k8sSecretsFile string) *k8sSecretResolverConfig {

	var configk8sSecretsFromFile k8sSecretResolverConfig
	file, err1 := ioutil.ReadFile(k8sSecretsFile)
	if err1 == nil {
		err2 := json.Unmarshal(file, &configk8sSecretsFromFile)
		if err2 != nil {
			k8sSecretLoggger.Errorf("Error - '%v' occurred while parsing configuration JSON file", err2)
			return nil
		}
	} else {
		k8sSecretLoggger.Errorf("Error - '%v' occurred while reading configuration JSON file", err1)
		return nil
	}
	return &configk8sSecretsFromFile
}

func k8sSecretsConfigFromInlineJSON(k8sSecretsFile string) *k8sSecretResolverConfig {

	var configK8sSecrets k8sSecretResolverConfig
	err2 := json.Unmarshal([]byte(k8sSecretsFile), &configK8sSecrets)
	if err2 != nil {
		k8sSecretLoggger.Errorf("Error - '%v' occurred while parsing configuration JSON", err2)
		return nil
	}
	return &configK8sSecrets
}

var prelodedProps = make(map[string]interface{})

func preloadPropsFromFile() error {
	if configK8sSecrets.Volume == "" {
		return errors.New("Volume path must not be empty.")
	}

	err := filepath.Walk(configK8sSecrets.Volume, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(info.Name(), "..") || info.Mode().IsRegular() {
			// Dont process if its directory or regular file.
			// K8s creates links for appropriate files in mounted directory
			return nil
		}

		name := info.Name()
		if name[0] == '.' {
			// Hidden file
			name = name[1:]
		}

		extn := filepath.Ext(name)
		if extn != "" {
			// Remove extension
			name = strings.Replace(name, extn, "", -1)
		}

		name = strings.ToLower(name)
		_, ok := prelodedProps[name]
		if !ok {
			k8sSecretLoggger.Debugf("Reading property file from path: %s", path)
			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			if len(contents) == 0 {
				k8sSecretLoggger.Warnf("Skipping empty file - %s", path)
				return nil
			}
			k8sSecretLoggger.Debugf("Adding property with key - %s", name)
			prelodedProps[name] = string(contents)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type K8sSecretResolver struct {
}

func (resolver *K8sSecretResolver) Name() string {
	return ResolverName
}

func (resolver *K8sSecretResolver) LookupValue(toResolve string) (interface{}, bool) {
	toResolve = strings.Replace(toResolve, ".", "_", -1)
	// Standardise on lower case for easy lookup
	toResolve = strings.ToLower(toResolve)
	k8sSecretLoggger.Debugf("Resolving property with key - %s", toResolve)
	val, ok := prelodedProps[toResolve]
	return val, ok
}
