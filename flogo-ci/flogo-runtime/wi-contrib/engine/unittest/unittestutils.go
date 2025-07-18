package unittest

import (
	"os"
)

func isTCIEnv() bool {
	return os.Getenv("TIBCO_INTERNAL_TCI_SUBSCRIPTION_ID") != ""
}

func getUnitTestModel(version interface{}) int {
	if version == nil {
		return V2Internal
	}
	switch version.(string) {
	case "1.0.0":
		return V1
	case "1.1.0":
		return V2
	case "1.1.1":
		return V3
	}
	return V1
}

func showInfo(message string) string {
	color := "\033[1;34m" + message + "\033[0m"
	return color
}

func showError(message string) string {
	errorString := "\033[1;31m" + message + "\033[0m"
	return errorString
}

func showSuccess(message string) string {
	successString := "\033[1;32m" + message + "\033[0m"
	return successString
}
