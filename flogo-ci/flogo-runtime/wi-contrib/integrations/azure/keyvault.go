package azure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

var keyVaultLogger = log.ChildLogger(log.RootLogger(), "appprops.azurekeyvault.resolver")

var config *keyVaultConfig

const (
	AzureKeyVaultConfigKey = "FLOGO_APP_PROPS_AZURE_KEYVAULT" //has to be set to true to enable this feature
	KeyVaultResolverName   = "azurekeyvault"
)

func init() {
	if !useAzureKeyVaultConfiguration() {
		return
	}

	keyVaultName := os.Getenv("FLOGO_AZURE_KEYVAULT_NAME")
	if keyVaultName == "" {
		keyVaultLogger.Error("Azure Key Vault name is required.")
		panic("Azure Key Vault name is required")
	}

	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", keyVaultName)

	// Build the DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		keyVaultLogger.Error(err.Error())
		panic("Failed to obtain credential")
	}

	// Create a new secret client
	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		keyVaultLogger.Error(err.Error())
		panic("Failed to create secret client")
	}

	// Check presence of dummy secret, and ignore if error is SecretNotFound
	// This is to check if the credentails are valid, vault exists, client can access the vault
	_, err = client.GetSecret(context.Background(), "dummy", "", nil)
	if err != nil {
		if !strings.Contains(err.Error(), "SecretNotFound") {
			keyVaultLogger.Error(err.Error())
			panic(err.Error())
		}
	}

	config = &keyVaultConfig{
		client: client,
	}

	property.RegisterPropertyResolver(&KeyVaultValueResolver{})
	envProp := os.Getenv(engine.EnvAppPropertyResolvers)
	if envProp == "" {
		//Make azurekeyvault resolver default since FLOGO_APP_PROPS_AZURE_KEYVAULT is set
		os.Setenv(engine.EnvAppPropertyResolvers, KeyVaultResolverName)
	} else if envProp == dynamicprops.ResolverName {
		//If only dynamic property resolver is enabled append azurekeyvault resolver after it
		os.Setenv(engine.EnvAppPropertyResolvers, fmt.Sprintf("%s,%s", dynamicprops.ResolverName, KeyVaultResolverName))
	}
	keyVaultLogger.Debug("Azure Key Vault resolver registered")

}

func useAzureKeyVaultConfiguration() bool {
	key := os.Getenv(AzureKeyVaultConfigKey)
	val, err := coerce.ToBool(key)
	if err == nil {
		return val
	}
	return false
}

type keyVaultConfig struct {
	client *azsecrets.Client
}

// KeyVaultValueResolver
type KeyVaultValueResolver struct {
}

// Name
func (resolver *KeyVaultValueResolver) Name() string {
	return KeyVaultResolverName
}

func (resolver *KeyVaultValueResolver) LookupValue(key string) (interface{}, bool) {
	// Replace dot with dash e.g. a.b would be a-b
	if strings.Contains(key, ".") {
		keyVaultLogger.Debugf("Replacing '.' with '-' in secret: %s", key)
		key = strings.Replace(key, ".", "-", -1)
	}

	// Replace underscore with dash e.g. a_b would be a-b
	if strings.Contains(key, "_") {
		keyVaultLogger.Debugf("Replacing '_' with '-' in secret: %s", key)
		key = strings.Replace(key, "_", "-", -1)
	}

	keyVaultLogger.Debugf("Looking up secret [%s] in Azure Key Vault...", key)
	//get the secret from key vault
	resp, err := config.client.GetSecret(context.Background(), key, "", nil)
	if err != nil {
		if strings.Contains(err.Error(), "SecretNotFound") {
			keyVaultLogger.Warnf("Secret '%s' not found in the Key Vault.", key)
			return nil, false
		} else if strings.Contains(err.Error(), "SecretDisabled") {
			keyVaultLogger.Warnf("Secret '%s' is disabled in the Key Vault.", key)
			return nil, false
		} else {
			keyVaultLogger.Warnf(err.Error())
			return nil, false
		}
	}

	return *resp.Value, true
}
