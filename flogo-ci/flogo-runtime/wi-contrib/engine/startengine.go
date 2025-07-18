package engine

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/starter"
)

var cpuprofile = flag.String("cpuprofile", "", "Writes CPU profiling for the current process to the specified file")
var memprofile = flag.String("memprofile", "", "Writes memory profiling for the current process to the specified file")
var engineConfig = `{
  "type": "flogo:engine",
  "actionSettings": {
    "github.com/project-flogo/flow": {
      "stateRecordingMode": "debugger"
    }
  },
  "services": [
    {
      "name": "stateful",
      "ref": "github.com/tibco/wi-contrib/engine/stateful",
      "enabled": true,
      "settings": {
      }
    }
  ]
}`

var properties []*data.Attribute

func StartProfile() error {
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "cpuprofile") || strings.Contains(os.Args[1], "memprofile") {
			//Only for profile case, other case handled by commander registry
			flag.Parse()
		}

		if *cpuprofile != "" {
			f, err := os.Create(*cpuprofile)
			if err != nil {
				return fmt.Errorf("Failed to create CPU profiling file due to error - %s", err.Error())
			}
			pprof.StartCPUProfile(f)
		}
	}
	return nil
}

func HasProfile() bool {
	return *memprofile != "" || *cpuprofile != ""
}

func stopProfile() error {
	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			return fmt.Errorf("Failed to create memory profiling file due to error - %s", err.Error())
		}

		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			return fmt.Errorf("Failed to write memory profiling data to file due to error - %s", err.Error())
		}
		f.Close()
	}
	return nil
}

func CreateEngine(appJson string) (engine.Engine, error) {
	flogoApp, err := CreateFlogoApp(appJson)
	if err != nil {
		return nil, fmt.Errorf("create flogo app from flogo.json error: %s", err.Error())
	}

	e, err := engine.New(flogoApp, engine.ConfigOption(getEngineConfig(), false))
	if err != nil {
		return nil, fmt.Errorf("Failed to create engine instance due to error: %s", err.Error())
	}
	return e, nil
}

func getEngineConfig() string {

	configPath := engine.GetFlogoEngineConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		flogo, err := os.Open(configPath)
		if err != nil {
			return engineConfig
		}

		jsonBytes, err := ioutil.ReadAll(flogo)
		if err != nil {
			return engineConfig
		}
		return string(jsonBytes)
	}

	return engineConfig
}

func CreateFlogoApp(appJson string) (*app.Config, error) {

	if appJson == "" {
		var err error
		appJson, err = GetAppJsonFromCurrentDir()
		if err != nil {
			return nil, err
		}
	}

	cfgJson, err := UpgradeAppJson(appJson)
	if err != nil {
		return nil, fmt.Errorf("Failed to upgrade legacy app json: %s", err.Error())
	}

	setAppVariables(appJson)

	flogoApp, err := engine.LoadAppConfig(cfgJson, false)

	if err != nil {
		return nil, fmt.Errorf("Failed to create engine: %s", err.Error())
	}
	updateImports(flogoApp)
	return flogoApp, err
}

func GetAppJsonFromCurrentDir() (string, error) {
	// a json string wasn't provided, so lets lookup the file in path
	configPath := engine.GetFlogoAppConfigPath()

	flogo, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("Failed to load app: %s", err.Error())
	}
	jsonBytes, err := ioutil.ReadAll(flogo)
	if err != nil {
		return "", fmt.Errorf("Failed to read app file: %s", err.Error())
	}
	return string(jsonBytes), nil
}

func StartEngine(e engine.Engine) error {

	err := e.Start()
	if err != nil {
		return fmt.Errorf("Failed to start engine due to error: %s", err.Error())
	}
	return nil
}

func StartOthers(e engine.Engine, startTime time.Time) error {
	mangement := NewManagement(e)
	go mangement.startServer()

	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
	println("     Runtime started in " + time.Since(startTime).String())
	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

	if starter.HasShimRunner() {
		for name, runner := range starter.AllRunner() {
			runner.Init(os.Args)
			nextStep, err := runner.Run(os.Args)
			if err != nil {
				log.RootLogger().Errorf("Failed to run runner [%s] due to error: %s", name, err.Error())
			}
			if nextStep == starter.TERMINATE {
				os.Exit(0)
			}
		}
	}

	exitChan := setupSignalHandling()

	code := <-exitChan

	e.Stop()

	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
	println("     Runtime was up since " + time.Since(startTime).String())
	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

	if err := stopProfile(); err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("exit %d", code)
	}
	return nil
}

func setupSignalHandling() chan int {

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int, 1)
	select {
	case s := <-signalChan:
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitChan <- 0
		default:
			log.RootLogger().Debug("Unknown signal.")
			exitChan <- 1
		}
	}
	return exitChan
}

func updateImports(config *app.Config) {
	var completeMap struct {
		AllCons []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Tag     string `json:"tag"`
			Ref     string `json:"ref"`
		} `json:"conDescs"`
		Old2NewMap map[string]string `json:"old2newRef"`
	}

	js := GetSharedData("connectionVersionJson")

	if js == nil {

		return
	}

	err := json.Unmarshal([]byte(js.(string)), &completeMap)
	if err != nil {
		return
	}

	imports := config.Imports

	var updatedImports []string

	if completeMap.Old2NewMap == nil || len(completeMap.Old2NewMap) == 0 {
		return
	}

	for _, value := range imports {

		if newRef, ok := completeMap.Old2NewMap[value]; ok {
			updatedImports = append(updatedImports, newRef)
		} else {
			updatedImports = append(updatedImports, value)
		}
	}
	config.Imports = updatedImports
}
