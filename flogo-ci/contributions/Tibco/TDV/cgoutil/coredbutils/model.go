package coredbutils

import "github.com/project-flogo/core/support/log"

var logCache = log.ChildLogger(log.RootLogger(), "tdv.connection")

type DBDiagnostic struct {
	State       string `json:"State"`
	NativeError string `json:"NativeError"`
	Messge      string `json:"Message"`
}
