package azdatafactory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// GetAzureClient authenticates a user with a password for given clientID parameter
func GetAzureClient(tenantID string, clientID string, userName string, password string) (*azidentity.UsernamePasswordCredential, error) {
	clientCred, err := azidentity.NewUsernamePasswordCredential(tenantID, clientID, userName, password, nil)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while authenticating username/password credential : %v", err)
	}
	return clientCred, nil
}

// GetBody returns a new bytes buffer from byte array of content
func GetBody(content interface{}) (io.Reader, error) {
	var reqBody io.Reader
	switch content.(type) {
	case string:
		reqBody = bytes.NewBuffer([]byte(content.(string)))
	default:
		b, err := json.Marshal(content)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(b)
	}
	return reqBody, nil
}
