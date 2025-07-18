package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func init() {
	Registry("help", &helpCommand{})
	Registry("-h", &helpCommand{})
	Registry("--h", &helpCommand{})
}

type helpCommand struct {
}

func (b *helpCommand) Name() string {
	return "help"
}

func (b *helpCommand) Description() string {
	return "Help command"
}

func (b *helpCommand) Run(args []string, appJson string) error {
	exeName := os.Args[0]
	if len(args) <= 0 {
		fmt.Println("Usage:")
		fmt.Println("\t " + exeName + " help COMMAND")
		fmt.Println("Avalibale Commands:")
		for k, c := range GetAllCommands() {
			if "help" != c.Name() {
				if !strings.HasPrefix(k, "-") {
					fmt.Println(fmt.Sprintf("\t %s - %s", c.Name(), c.Description()))
				}
			}
		}
		return nil
	}

	command := GetAllCommands()[args[0]]
	if command != nil {
		command.PrintUsage()
		return nil
	} else {
		return fmt.Errorf("Unknow command %s", args[0])
	}

	return nil
}

func (b *helpCommand) AddFlags(fs *flag.FlagSet) {

}

func (b *helpCommand) PrintUsage() {

}

func (b *helpCommand) IsShimCommand() bool {
	return false
}

func (b *helpCommand) Parse() bool {
	return true
}
