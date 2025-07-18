package cmd

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	feengine "github.com/tibco/wi-contrib/engine"
)

var unsupportedRef = []string{"github.com/TIBCOSoftware/flogo-contrib/trigger/lambda", "git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/Lambda/trigger/lambda", "github.com/project-flogo/grpc/trigger/grpc"}

var unsupportedAct = []string{"github.com/project-flogo/grpc/activity/grpc", "git.tibco.com/git/product/ipaas/wi-contrib.git/contributions/General/activity/protobuf2json"}

var unsupportedRefVals = make(map[string]string)

func init() {
	Registry("app", &appCommand{})
	Registry("-app", &appCommand{})
	Registry("--app", &appCommand{})

	unsupportedRefVals["git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/Lambda/trigger/lambda"] = "Lambda"

	unsupportedRefVals["github.com/project-flogo/grpc/trigger/grpc"] = "gRPC"

	unsupportedRefVals["github.com/TIBCOSoftware/flogo-contrib/trigger/lambda"] = "Lambda"
	unsupportedRefVals["github.com/project-flogo/grpc/activity/grpc"] = "gRPC Invoke"
	unsupportedRefVals["git.tibco.com/git/product/ipaas/wi-contrib.git/contributions/General/activity/protobuf2json"] = "Protobuf To JSON"
}

type appCommand struct {
	debug bool
}

func (b *appCommand) Name() string {
	return "-app"
}

func (b *appCommand) Description() string {
	return "Provide the Application JSON File"
}

func (b *appCommand) Run(args []string, appJSON string) error {
	baseEngine, _ := coerce.ToBool(feengine.GetSharedData("isBaseEngine"))
	if len(args) > 0 {
		//Only provide app json to run
		content, err := ioutil.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("Read application json file [%s] error: %s", args[0], err.Error())
		}
		if baseEngine {
			validate, err := validateApp(string(content))
			if !validate {
				return err
			}

			if len(args) == 1 {
				startTime := time.Now()
				println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
				println("     Starting TIBCO Flogo® Runtime")
				println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
				e, err := feengine.CreateEngine(string(content))
				if err != nil {
					return err
				}
				addAppJSONToMap(string(content))
				PrintContributions(string(content))
				//Add validate here to make sure no shim trigger.
				err = feengine.StartEngine(e)
				if err != nil {
					return err
				}
				err = feengine.StartOthers(e, startTime)
				if err != nil {
					return err
				}
			} else {
				//additional commands
				err := HandleCommandline(args[1:], string(content))
				if err != nil {
					return err
				}
			}
		} else {
			//This is for general binary
			startTime := time.Now()
			println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
			println("     Starting TIBCO Flogo® Runtime")
			println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

			e, err := feengine.CreateEngine(string(content))
			if err != nil {
				log.RootLogger().Error(err)
				os.Exit(1)
			}
			addAppJSONToMap(string(content))
			err = feengine.StartEngine(e)
			if err != nil {
				log.RootLogger().Error(err)
				os.Exit(1)
			}

			err = feengine.StartOthers(e, startTime)
			if err != nil {
				log.RootLogger().Error(err)
				os.Exit(1)
			}
		}
	}

	return nil
}

func addAppJSONToMap(appJSON string) {
	if appJSON == "" {
		return
	}
	log.RootLogger().Debug("Adding app json to shared data")
	feengine.AddSharedData("flogoJSON", appJSON)
}

func (b *appCommand) AddFlags(fs *flag.FlagSet) {
}

func (b *appCommand) PrintUsage() {
	execName := os.Args[0]

	usage := "Command: \n" +
		"        " + execName + " --app <application path> \n" +
		"Usage:\n" +
		"        " + execName + " --app /Users/app/app.json \n"
	fmt.Println(usage)
}

func (b *appCommand) IsShimCommand() bool {
	return false
}

func (b *appCommand) Parse() bool {
	return false
}

// PrintContributions ...
func PrintContributions(appJSON string) {
	appMap := make(map[string]interface{})

	var ContribConf []struct {
		Ref        string `json:"ref"`
		S3Location string `json:"s3location"`
		Type       string `json:"type"`
	}

	err := json.Unmarshal([]byte(appJSON), &appMap)
	if err != nil {
		return
	}

	contribIn := appMap["contrib"]
	if contribIn != nil {
		contribByte, err := base64.StdEncoding.DecodeString(contribIn.(string))
		if err != nil {
			return
		}
		err = json.Unmarshal(contribByte, &ContribConf)
		if err != nil {
			return
		}
	}

	var conDescs []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Tag     string `json:"tag"`
		Ref     string `json:"ref"`
	}

	js := feengine.GetSharedData("connectionVersionJson")

	err = json.Unmarshal([]byte(js.(string)), &conDescs)
	if err != nil {
		return
	}

	for _, contrib := range ContribConf {

		for _, entry := range conDescs {

			if contrib.Ref == entry.Ref {

				if entry.Tag != "" {
					println(entry.Name + " -  " + entry.Version + "." + entry.Tag)
				} else {
					println(entry.Name + " -  " + entry.Version)
				}

			}
		}

	}
}

func validateApp(appJSON string) (bool, error) {
	config, _ := feengine.CreateFlogoApp(appJSON)
	imports := config.Imports

	for i := 0; i < len(imports); i++ {
		if containsRef(unsupportedRef, imports[i]) {
			return false, fmt.Errorf("The App JSON file contains the trigger [" + unsupportedRefVals[imports[i]] + "] which is not supported in Engine Binary mode")
		}

		if containsRef(unsupportedAct, imports[i]) {
			return false, fmt.Errorf("The App JSON file contains the activity [" + unsupportedRefVals[imports[i]] + "] which is not supported in Engine Binary mode")
		}

	}

	return true, nil

}

func containsRef(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
