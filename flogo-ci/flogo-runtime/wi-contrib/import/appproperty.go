package main

import (
	"encoding/json"
	"github.com/TIBCOSoftware/flogo-lib/config"
	"github.com/TIBCOSoftware/flogo-lib/util"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/support/log"
	"io/ioutil"
	"strings"
)

var preload = make(map[string]interface{})

func init() {
	propFile := config.GetAppPropertiesOverride()
	if propFile != "" {
		_ = property.RegisterPropertyResolver(&CommandLineOverriteResolver{})
		log.RootLogger().Infof("'%s' is set. Loading overridden properties", config.ENV_APP_PROPERTY_OVERRIDE_KEY)
		if strings.HasSuffix(propFile, ".json") {
			// Override through file
			file, err := ioutil.ReadFile(propFile)
			if err != nil {
				log.RootLogger().Errorf("Can not read - %s due to error - %v", propFile, err)
				panic("")
			}
			err = json.Unmarshal(file, &preload)
			if err != nil {
				log.RootLogger().Errorf("Can not convert property - %s due to error - %v", propFile, err)
				panic("")
			}
		} else if strings.ContainsRune(propFile, '=') {
			// Override through P1=V1,P2=V2
			overrideProps := util.ParseKeyValuePairs(propFile)
			for k, v := range overrideProps {
				preload[k] = v
			}
		}
	}
}

// Resolve property value from external files
type CommandLineOverriteResolver struct {
}

func (resolver *CommandLineOverriteResolver) Name() string {
	return "CommandLine"
}

func (resolver *CommandLineOverriteResolver) LookupValue(toResolve string) (interface{}, bool) {
	val, found := preload[toResolve]
	return val, found
}
