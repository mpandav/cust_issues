package integrations

import (
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/engine/secret"
	"strings"
)

func DecryptIfEncrypted(val string) string {

	if len(val) == 0 {
		return val
	}

	if strings.HasPrefix(val, "SECRET:") {
		encodedValue := string(val[7:])
		decodedValue, _ := secret.GetSecretValueHandler().DecodeValue(encodedValue)
		return decodedValue
	}
	return val
}

func SubstituteTemplate(key string) string {

	if len(key) == 0 {
		return key
	}

	if strings.Contains(key, "<APPNAME>") {
		key = strings.Replace(key, "<APPNAME>", engine.GetAppName(), -1)
	}

	if strings.Contains(key, "<APPVERSION>") {
		key = strings.Replace(key, "<APPVERSION>", engine.GetAppVersion(), -1)
	}

	if strings.HasSuffix(key, "/") {
		key = key[:len(key)-1]
	}

	return key
}
