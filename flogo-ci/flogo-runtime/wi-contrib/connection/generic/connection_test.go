package generic

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var connectionObject = `
{
                  "id": "12c237d0-2fbe-11e8-aab3-2714b6f3173f",
                  "type": "flogo:connector",
                  "version": "1.0.0",
                  "name": "tibco-chargify",
                  "inputMappings": {},
                  "outputMappings": {},
                  "title": "Chargify Connector",
                  "description": "This is Chargify connector",
                  "ref": "git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/chargify/connector/chargify",
                  "settings": [
                    {
                      "name": "name",
                      "type": "string",
                      "required": true,
                      "display": {
                        "name": "Connection Name",
                        "description": "Name of the connection"
                      },
                      "value": "mychargify"
                    },
                    {
                      "name": "description",
                      "type": "string",
                      "display": {
                        "name": "Description",
                        "description": "Connection description"
                      },
                      "value": ""
                    },
                    {
                      "name": "subDomain",
                      "type": "string",
                      "required": true,
                      "display": {
                        "name": "Site's Subdomain",
                        "description": "Enter your site's subdomain (<your-subdomain>.chargify.com) here "
                      },
                      "value": "tibco-software-inc"
                    },
                    {
                      "name": "apiKey",
                      "type": "string",
                      "required": true,
                      "display": {
                        "name": "API Key",
                        "type": "password",
                        "description": "Enter your API V1 key here. Refer https://help.chargify.com/integrations/api-keys-chargify-direct.html"
                      },
                      "value": "axaxaxaxaxaxaxaxaaxaaxxaaxaxax"
                    }
                  ],
                  "outputs": [],
                  "inputs": [],
                  "handler": {
                    "settings": []
                  },
                  "reply": [],
                  "s3Prefix": "flogo",
                  "key": "flogo/Chargify/connector/chargify/connector.json",
                  "display": {
                    "description": "This is Chargify connector",
                    "category": "Chargify",
                    "visible": true,
                    "smallIcon": "chargify.png"
                  },
                  "actions": [
                    {
                      "name": "Connect",
                      "display": {
                        "readonly": false,
                        "valid": true,
                        "visible": true
                      }
                    }
                  ],
                  "keyfield": "name",
                  "isValid": true,
                  "lastUpdatedTime": 1521935403981,
                  "createdTime": 1521935403981,
                  "user": "flogo",
                  "subscriptionId": "flogo_sbsc",
                  "connectorName": " ",
                  "connectorDescription": " "
                }
`

func TestConnection(t *testing.T) {
	conn, err := NewConnection(connectionObject)
	if conn == nil {
		t.Error(err)
		t.Fail()
		return
	}

	assert.Equal(t, conn.GetName(), "mychargify")
	assert.Equal(t, conn.GetId(), "12c237d0-2fbe-11e8-aab3-2714b6f3173f")
	assert.Equal(t, conn.GetSetting("subDomain"), "tibco-software-inc")

}
