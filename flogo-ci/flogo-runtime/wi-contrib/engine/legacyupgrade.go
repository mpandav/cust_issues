package engine

import (
	"encoding/json"
	"fmt"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/legacybridge/config"
	_ "github.com/project-flogo/legacybridge/config/flow"
)

func UpgradeAppJson(appContent string) (string, error) {
	ok, err := IsLegacyFlogoJson(appContent)
	if ok && err == nil {
		log.RootLogger().Debugf("Find the app json is an outdate, updated to new format")
		newContent, err := config.ConvertLegacyJson(appContent)
		if err != nil {
			return "", fmt.Errorf("Convert legacy json failed, due to %s", err.Error())
		}
		log.RootLogger().Debugf("Updated app: %s", newContent)
		return newContent, nil
	}

	return appContent, nil
}

func IsLegacyFlogoJson(appJson string) (bool, error) {
	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(appJson), &jsonMap)
	if err != nil {
		return false, err
	}
	appModel, ok := jsonMap["appModel"]
	if ok && appModel == "1.0.0" {
		return true, nil
	}
	return false, nil
}
