package send

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

func TestCreateEnvelope(t *testing.T) {
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

	createActivity := &CreateEnvelopeActivity{}
	activityContext := test.NewActivityContext(createActivity.Metadata())
	activityContext.SetInput("docusignConnection", conn.(*docusignconnection.DocusignSharedConfigManager))
	// activityContext.SetInput("isMultiDoc", false)
	activityContext.SetInput("isMultiDoc", true)

	activityContext.SetInput("recipients", "")
	activityContext.SetInput("signingInOrder", false)

	// activityContext.SetInput("fileName", "test.txt")
	// activityContext.SetInput("fileContent", "This test file")

	fmt.Println("IM here")

	var docs []map[string]interface{}
	for itr := 1; itr <= 2; itr++ {
		doc := make(map[string]interface{})
		doc["name"] = fmt.Sprint("file", itr, ".txt")
		doc["content"] = fmt.Sprint("this is test content file ", itr)
		docs = append(docs, doc)
	}

	activityContext.SetInput("documents", docs)

	_, err = createActivity.Eval(activityContext)
	assert.Nil(t, err)
	if err != nil {
		t.Error("could not send envelope")
		t.Fail()
	} else {
		output := activityContext.GetOutput("envelope")
		// fmt.Printf("output envelope: %#v", output)
		assert.NotNil(t, output)
		errOutput := activityContext.GetOutput("error")
		assert.Nil(t, errOutput)
	}
}
