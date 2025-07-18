package cmd

import (
	"flag"
	"fmt"
	"os"
)

type BSCommand interface {
	Name() string
	Description() string
	Run(args []string, appJson string) error
	AddFlags(fs *flag.FlagSet)
	PrintUsage()
	IsShimCommand() bool
	Parse() bool
}

func ExecCommand(fs *flag.FlagSet, cmd BSCommand, args []string, appJson string) error {
	cmd.AddFlags(fs)
	fs.Usage = cmd.PrintUsage

	if cmd.Parse() {
		if err := fs.Parse(args); err != nil {
			os.Exit(1)
		}
		args = fs.Args()
	}
	return cmd.Run(args, appJson)

}

func HandleCommandline(args []string, appJson string) error {
	secondArg := args[0]
	commandSet := flag.NewFlagSet(secondArg, flag.ContinueOnError)
	command := GetCommand(secondArg)
	if command != nil {
		err := ExecCommand(commandSet, command, args[1:], appJson)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Undefined command [%s]", secondArg)
	}

	return nil
}
