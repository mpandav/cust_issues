package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tibco/wi-contrib/engine"
)

func init() {
	Registry("-listInstalledContribs", &listContribCommand{})
	Registry("--listInstalledContribs", &listContribCommand{})
}

type listContribCommand struct {
	debug bool
}

func (b *listContribCommand) Name() string {
	return "-listInstalledContribs"
}

func (b *listContribCommand) Description() string {
	return "List the Contribs"
}

func (b *listContribCommand) Run(args []string, appJson string) error {
	printAllContributions()
	return nil
}

func (b *listContribCommand) AddFlags(fs *flag.FlagSet) {
}

func (b *listContribCommand) PrintUsage() {
	execName := os.Args[0]

	usage := "Command: \n" +
		"        " + execName + " -listInstalledContribs\n" +
		"Usage:\n" +
		"        " + execName + " -listInstalledContribs\n"
	fmt.Println(usage)
}

func (b *listContribCommand) IsShimCommand() bool {
	return false
}

func (b *listContribCommand) Parse() bool {
	return true
}

func printAllContributions() {

	var conDescs []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Tag     string `json:"tag"`
		Ref     string `json:"ref"`
	}
	js := engine.GetSharedData("connectionVersionJson")
	err := json.Unmarshal([]byte(js.(string)), &conDescs)
	if err != nil {
		return
	}

	for _, entry := range conDescs {

		if entry.Tag != "" {
			println(entry.Name + " -  " + entry.Version + "." + entry.Tag)
		} else {
			println(entry.Name + " -  " + entry.Version)
		}
	}

}
