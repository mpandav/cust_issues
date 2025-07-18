package cmd

import "sync"

var registedCommands = make(map[string]BSCommand)
var lock sync.Mutex

func Registry(name string, command BSCommand) {
	lock.Lock()
	registedCommands[name] = command
	lock.Unlock()
}

func GetAllCommands() map[string]BSCommand {
	return registedCommands
}

func GetCommand(name string) BSCommand {
	return GetAllCommands()[name]
}

func GetShimCommand(name string) BSCommand {
	return GetUnshimCommands()[name]
}

func GetUnshimCommands() map[string]BSCommand {
	c := make(map[string]BSCommand)
	for k, v := range registedCommands {
		if !v.IsShimCommand() {
			c[k] = v
		}
	}
	return c
}

func GetShimCommands() map[string]BSCommand {
	c := make(map[string]BSCommand)
	for k, v := range registedCommands {
		if v.IsShimCommand() {
			c[k] = v
		}
	}
	return c
}
