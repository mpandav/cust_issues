package main

//This is a dummy file to make sure no compile issue for shim.go
import (
	"github.com/project-flogo/core/app"
	"os"
)

const cfgJson string = ``

func exeCmd(flogoApp *app.Config) {
	os.Exit(0)
}
