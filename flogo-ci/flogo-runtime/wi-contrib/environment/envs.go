package environment

import (
	"os"
	"strings"
)

var url, subId, userName, sfClientSecret string
var appName, appId, containerId string

func init() {
	defer func() {
		os.Unsetenv("TCI_WI_SALESFORCE_CLIENT_SECRET")
		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			if len(pair) > 0 {
				varName := pair[0]
				if strings.HasPrefix(varName, "TIBCO_INTERNAL") {
					if varName == "TIBCO_INTERNAL_INTERCOM_URL" || varName == "TIBCO_INTERNAL_TCI_SUBSCRIPTION_ID" || varName == "TIBCO_INTERNAL_SERVICE_NAME" || varName == "TIBCO_INTERNAL_TCI_SUBSCRIPTION_UNAME" || varName == "TIBCO_INTERNAL_TCI_TIBTUNNEL_ACCESS_KEY" || varName == "TIBCO_INTERNAL_CONTAINER_AGENT_HTTP_PORT" || varName == "TIBCO_INTERNAL_TCI_IS_CIC2_ENV" {
						// Not removing them for now  as Liveapps team is using them
						// TODO Should be removed once Liveapps team use APIs
						continue
					}
					// Removing all TIBCO Internal Env vars
					os.Unsetenv(varName)
				}
			}
		}
	}()

	//Salesforce client secret
	sfClientSecret = os.Getenv("TCI_WI_SALESFORCE_CLIENT_SECRET")
	url = os.Getenv("TIBCO_INTERNAL_INTERCOM_URL")
	subId = os.Getenv("TIBCO_INTERNAL_TCI_SUBSCRIPTION_ID")
	userName = os.Getenv("TIBCO_INTERNAL_TCI_SUBSCRIPTION_UNAME")
	appName = os.Getenv("TIBCO_INTERNAL_TCI_APP_NAME")
	appId = os.Getenv("TIBCO_INTERNAL_SERVICE_NAME")
	containerId = os.Getenv("HOSTNAME")
}

// GetSalesforceClientSecret to get salesforce client secret from env variable
func GetSalesforceClientSecret() string {
	return sfClientSecret
}

func GetIntercomURL() string {
	return url
}

func GetTCISubscriptionId() string {
	return subId
}

func GetTCISubscriptionUName() string {
	return userName
}

func GetTCIAppId() string {
	return appId
}

func GettCIContainerId() string {
	return containerId
}

func GetTCIAppName() string {
	return appName
}

func IsTCIEnv() bool {
	_, ok := os.LookupEnv("TIBCO_INTERNAL_TCI_SUBSCRIPTION_ID")
	return ok
}

func IsCIC2Env() bool {
	_, ok := os.LookupEnv("TIBCO_INTERNAL_TCI_IS_CIC2_ENV")
	return ok
}

func IsEnvHybridMon() bool {
	_, ok := os.LookupEnv("TCI_HYBRID_AGENT_HOST")
	return ok
}

func IsTesterEnv() bool {
	_, ok := os.LookupEnv("FLOGO_DEBUGGER_ENV")
	return ok
}
