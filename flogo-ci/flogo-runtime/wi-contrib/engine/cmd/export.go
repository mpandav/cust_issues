package cmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data/coerce"
	"io/ioutil"
	"os"
	"strings"
)

func init() {
	Registry("export", &exportCommand{})
	Registry("-export", &exportCommand{})
	Registry("--export", &exportCommand{})

}

type exportCommand struct {
	outOption string
}

func (b *exportCommand) Name() string {
	return "-export"
}

func (b *exportCommand) Description() string {
	return "Export app properties from app"
}

func (b *exportCommand) Run(args []string, appJson string) error {
	if len(args) <= 0 {
		return fmt.Errorf("Export command requires option props-json | props-env")
	}
	originalApp := &app.Config{}
	err := json.Unmarshal([]byte(appJson), originalApp)
	if err != nil {
		fmt.Printf("Fail to parse flogo json due to %s \n", err.Error())
		os.Exit(1)
	}
	return b.execExportCmd(args, originalApp ,appJson )
}

func (b *exportCommand) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.outOption, "o", "", "Optional, Properties output json file option")
}

func (b *exportCommand) PrintUsage() {
	execName := os.Args[0]
	usage := "Command: \n" +
		"        " + execName + " -export props-json | props-env [options] \n" +
		"Options:\n" +
		"        -o      Optional file name \n" +
		"Usage:\n" +
		"        " + execName + " -export props-json\n" +
		"        " + execName + " -export props-env\n" +
		"Example:\n" +
		"        " + execName + " -export props-json  will export app props in <appname>-props.json\n" +
		"        " + execName + " -export props-env   will export app props in <appname>-env.properties\n" +
		"        " + execName + " -export props-env -o dev-env.properties  will export app props in dev-env.properties"
	fmt.Println(usage)
}

func (b *exportCommand) IsShimCommand() bool {
	return false
}

func (b *exportCommand) Parse() bool {
	return true
}

func (b *exportCommand) execExportCmd(args []string, app *app.Config, appJson string) error {
	exportCommand := args[0]
	switch exportCommand {
	case "props-json":

		outputFile := app.Name + "-props.json"
		if b.outOption != "" {
			outputFile = b.outOption
		}

		keyVs := make(map[string]interface{}, len(app.Properties))
		for _, prop := range app.Properties {
			keyVs[prop.Name()] = prop.Value()
		}

		return exportPropsToJson(keyVs, outputFile)

	case "props-env":

		outputFile := app.Name + "-env.properties"
		if b.outOption != "" {
			outputFile = b.outOption
		}

		keyVs := make(map[string]interface{}, len(app.Properties))
		for _, prop := range app.Properties {
			keyVs[prop.Name()] = prop.Value()
		}

		return exportPropsToEnv(keyVs, outputFile)


	case "app":
		outputFile := app.Name + ".json"
		if b.outOption != "" {
			outputFile = b.outOption
		}
		return exportAppToFile( outputFile , appJson)


	default:
		fmt.Println("Invalid export usage")
		//printExportCmdUsage()
		return fmt.Errorf("Invalid export usage")
	}
}

func exportAppToFile( outputFile string , appJSON string ) error {

	pretty := jsonPrettyPrint( appJSON )

	ioutil.WriteFile(outputFile , []byte(pretty) , 0777)

	fmt.Println("App json exported successfully to file " + outputFile)
	return nil
}

func exportPropsToJson(props map[string]interface{}, outputFile string) error {
	bp, err := json.MarshalIndent(props, "", "  ")
	if err != nil {
		fmt.Println("Fail to export app properties due to: " + err.Error())
		return fmt.Errorf("Fail to export app properties due to: " + err.Error())
	}
	ioutil.WriteFile(outputFile, bp, 0777)

	fmt.Println("App properties exported successfully to json file " + outputFile)
	return nil
}

func exportPropsToEnv(props map[string]interface{}, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Fail to export app properties due to: " + err.Error())
		os.Exit(1)
	}
	defer f.Close()

	for key, value := range props {
		strV, err := coerce.ToString(value)
		if err != nil {
			fmt.Println("Fail to export app properties due to: " + err.Error())
			return fmt.Errorf("Fail to export app properties due to: " + err.Error())
		}

		key = strings.Replace(key, ".", "_", -1)
		key = strings.ToUpper(key)

		line := key + "=" + strV
		f.WriteString(line + "\n")
	}
	f.Sync()

	fmt.Println("App properties exported successfully to env file " + outputFile)
	return nil
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}