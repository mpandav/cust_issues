package main

import (
	"github.com/project-flogo/core/engine/secret"
	"os"
)

var DATA_SECRET_KEY_DEFAULT string

func init() {
	secret.SetSecretValueHandler(&secret.KeyBasedSecretValueHandler{Key: GetDataSecretKey()})
}

func GetDataSecretKey() string {
	key := os.Getenv(secret.EnvKeyDataSecretKey)
	if len(key) > 0 {
		return key
	}
	return DATA_SECRET_KEY_DEFAULT
}
