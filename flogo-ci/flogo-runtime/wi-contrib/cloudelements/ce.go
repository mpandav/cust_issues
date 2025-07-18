package cloudelements

import "os"

var apiDomain, consoleDomain, orgSecret, userSecret string

func init() {

	defer func() {
		// Cleanup to avoid exposure
		os.Unsetenv("TCI_WI_CLOUD_ELEMENTS_API_URL")
		os.Unsetenv("TCI_WI_CLOUD_ELEMENTS_CONSOLE_URL")
		os.Unsetenv("TCI_WI_CLOUD_ELEMENTS_ORG_TOKEN")
		os.Unsetenv("TCI_WI_CLOUD_ELEMENTS_USER_SECRET")
	}()

	//Read values from environment variable
	apiDomain = os.Getenv("TCI_WI_CLOUD_ELEMENTS_API_URL")
	consoleDomain = os.Getenv("TCI_WI_CLOUD_ELEMENTS_CONSOLE_URL")
	orgSecret = os.Getenv("TCI_WI_CLOUD_ELEMENTS_ORG_TOKEN")
	userSecret = os.Getenv("TCI_WI_CLOUD_ELEMENTS_USER_SECRET")
}

func GetAPIDomain() string {
	return apiDomain
}

func GetConsoleDomain() string {
	return consoleDomain
}

func GetOrgSecret() string {
	return orgSecret
}
func GetUserSecret() string {
	return userSecret
}
