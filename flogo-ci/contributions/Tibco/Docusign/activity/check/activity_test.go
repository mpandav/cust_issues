package check

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

func TestCheckDocuments(t *testing.T) {
	connectionJSON := `{
		"name":"docusignTest",
		"ref":"#docusign",
		"id":"docusignTest",
		"DocusignToken": {
			"connection_id":"docusignTest",
			"client_id":"",
			"client_secret":"",
			"access_token": "",
			"expires_in":,
			"refresh_token":"",
			"token_type":"Bearer",
			"env":"account-d",
			"scope":"temp"
		},
		"DocusignAccount": {
			"account_id":"",
			"account_name":"sample",
			"is_default":true,
			"base_uri":"https://demo.docusign.net"
		}
	}`

	fmt.Println("conncetionJSOn", connectionJSON)
	connectionObject := &docusignconnection.DocusignSharedConfigManager{}
	// connectionObject := make(map[string]interface{})
	err := json.Unmarshal([]byte(connectionJSON), connectionObject)
	if err != nil {
		panic("error while unmarhsalling connection: " + err.Error())
	}

	support.RegisterAlias("connection", "docusign", "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign")
	conn, err := docusignconnection.GetSharedConfiguration(connectionObject)
	if err != nil {
		panic("\n\nerror while coercing connection" + err.Error())
	}

	checkActivity := &CheckStatusActivity{}
	activityContext := test.NewActivityContext(checkActivity.Metadata())
	activityContext.SetInput("docusignConnection", conn.(*docusignconnection.DocusignSharedConfigManager))
	activityContext.SetInput("envelopeId", "")

	_, err = checkActivity.Eval(activityContext)
	assert.Nil(t, err)
	if err != nil {
		t.Error("could not fetch status")
		t.Fail()
	} else {
		output := activityContext.GetOutput("status")
		fmt.Println("status: ", output)
		assert.NotNil(t, output)
		errOutput := activityContext.GetOutput("error")
		assert.Nil(t, errOutput)
	}
}
