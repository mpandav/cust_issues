package main

import (
	"github.com/project-flogo/core/support/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"
)

func init() {
	startTime := time.Now()
	pluginFolder := getPluginDir()

	if err := loadPlugin(pluginFolder); err != nil {
		log.RootLogger().Error("Failed to start engine due to Connector plug-in error. Contact TIBCO Support.")
		os.Exit(1)
	}

	log.RootLogger().Debugf("Loading all plugins taken [%s]", time.Since(startTime))
}

func loadPlugin(pluginDir string) error {
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil
	}

	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	if files != nil && len(files) > 0 {
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".so") {
				log.RootLogger().Debugf("Loading plugin from %s name %s", filepath.Join(pluginDir, f.Name()), f.Name())
				_, err = plugin.Open(filepath.Join(pluginDir, f.Name()))
				if err != nil {
					log.RootLogger().Errorf("Failed to load Connector plug-in [%s] , due to %s", f.Name(), err.Error())
					//return err
				}
				log.RootLogger().Debugf("Loaded plugins [%s]", f.Name())
			}

		}
	}

	return nil
}

func getPluginDir() string {
	//For lambda case
	p := os.Getenv("LAMBDA_TASK_ROOT")
	if p != "" {
		return filepath.Join(p, "plugins")
	}

	return "plugins"
}
