/*
 * Copyright Â© 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	//"github.com/TIBCOSoftware/flogo-lib/core/data"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		activityMetadata = activity.ToMetadata(string(jsonMetadataBytes))
	}
	return activityMetadata
}

// func TestActivityRegistration(t *testing.T) {
// 	act := NewActivity(getActivityMetadata())
// 	if act == nil {
// 		t.Error("Activity Not Registered")
// 		t.Fail()
// 		return
// 	}
// }

var officeConnectionJSON = `{
	"id": "09f7f6e0-8b20-11e8-9108-0723664c23c1",
	"type": "flogo:connector",
	"version": "1.0.0",
	"name": "Azurestorage",
	"inputMappings": {},
	"outputMappings": {},
	"title": "Microsoft Azure Storage Connector",
	"description": "Establish connection to your azurestorage account",
	"ref": "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection",
	"settings": [{
			"name": "name",
			"type": "string",
			"required": true,
			"display": {
				"name": "Connection Name",
				"visible": true
			},
			"value": "hello"
		},
		{
			"name": "description",
			"type": "string",
			"display": {
				"name": "Description",
				"visible": true
			},
			"value": ""
		},
		{
			"name": "accountName",
			"type": "string",
			"required": true,
			"display": {
				"name": "Account Name",
				"visible": true
			},
			"value": "csg3d9f846f9849x4329xa35"
		},
		{
			"name": "startDate",
			"type": "string",
			"display": {
				"name": "Start Date",
				"visible": true
			},
			"value": ""
		},
		{
			"name": "expiryDate",
			"type": "string",
			"required": true,
			"display": {
				"name": "Expiry Date",
				"visible": true
			},
			"value": "2050-12-20T21:55:07Z"
		},
		{
			"name": "accessKey",
			"type": "string",
			"required": true,
			"display": {
				"name": "Access Key",
				"visible": true
			},
			"value": "NmXONqyr8THSJUpDzPm2tkHl9miFmBgWIhLR/8O8rykO4WFeG8wdcb/vm758687r8qrJ0k+aDnahwNt7kJuMEw=="
		},
		{
			"name": "WI_STUDIO_OAUTH_CONNECTOR_INFO",
			"type": "string",
			"required": true,
			"display": {
				"visible": false
			},
			"value": "{\"account_name\":\"csg3d9f846f9849x4329xa35\",\"access_key\":\"NmXONqyr8THSJUpDzPm2tkHl9miFmBgWIhLR/8O8rykO4WFeG8wdcb/vm758687r8qrJ0k+aDnahwNt7kJuMEw==\",\"expiry_date\":\"2050-12-20T21:55:07Z\",\"start_date\":\"\",\"sas_token\":\"?sv=2017-11-09&ss=bfqt&srt=sco&sp=rwdlacup&se=2050-12-20T21:55:07Z&spr=https&sig=RD0fH8cHBV1NY6RhW19AFsZBEnWqmCUG1kvj%2FDyyjiQ%3D\"}"
		},
		{
			"name": "configProperties",
			"type": "string",
			"required": true,
			"display": {
				"visible": false
			},
			"value": ""
		},
		{
			"name": "DocsMetadata",
			"type": "string",
			"required": true,
			"display": {
				"visible": false
			},
			"value": "{\"service\":[\"File\",\"Table\"],\"paths\":{\"/File/getfile\":{\"get\":{\"summary\":\"Get File\",\"action\":\"Get File\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"\",\"required\":false,\"type\":\"string\"},\"directoryPath\":{\"description\":\"\",\"required\":false,\"type\":\"string\"},\"fileName\":{\"description\":\"\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\",\"fileName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"fileContent\":{\"description\":\"base64 string\",\"required\":false,\"type\":\"string\"}}}}}},\"/File/listdirNfiles\":{\"get\":{\"summary\":\"List Directories and Files\",\"action\":\"List Directories and Files\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"create share under specified storage account\",\"required\":false,\"type\":\"string\"},\"directoryPath\":{\"description\":\"create share under specified storage account\",\"required\":false,\"type\":\"string\"},\"prefix\":{\"description\":\"create share under specified storage account\",\"required\":false,\"type\":\"string\"},\"maxresults\":{\"description\":\"create share under specified storage account\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\",\"directoryPath\"]},\"output\":{\"$schema\":\"http://json-schema.org/draft-04/schema#\",\"type\":\"object\",\"properties\":{\"EnumerationResults\":{\"type\":\"object\",\"properties\":{\"Entries\":{\"type\":\"object\",\"properties\":{\"Directory\":{\"type\":\"array\",\"items\":[{\"type\":\"object\",\"properties\":{\"Name\":{\"type\":\"string\"},\"Properties\":{\"type\":\"string\"}}}]},\"File\":{\"type\":\"array\",\"items\":[{\"type\":\"object\",\"properties\":{\"Name\":{\"type\":\"string\"},\"Properties\":{\"type\":\"object\",\"properties\":{\"Content-Length\":{\"type\":\"string\"}}}}}]}}},\"NextMarker\":{\"type\":\"string\"}}}}}}}},\"/File/listshare\":{\"get\":{\"summary\":\"List Shares\",\"action\":\"List Shares\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{}},\"output\":{\"$schema\":\"http://json-schema.org/draft-04/schema#\",\"type\":\"object\",\"properties\":{\"EnumerationResults\":{\"type\":\"object\",\"properties\":{\"Shares\":{\"type\":\"object\",\"properties\":{\"Share\":{\"type\":\"array\",\"items\":[{\"type\":\"object\",\"properties\":{\"Name\":{\"type\":\"string\"},\"Properties\":{\"type\":\"object\",\"properties\":{\"Last-Modified\":{\"type\":\"string\"},\"Etag\":{\"type\":\"string\"},\"Quota\":{\"type\":\"string\"}}}}}]}}},\"NextMarker\":{\"type\":\"string\"}}}}}}}},\"/File/createShare\":{\"put\":{\"summary\":\"Create Share\",\"action\":\"Create Share\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"create share under specified storage account\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"message\":{\"description\":\"output message\",\"required\":false,\"type\":\"string\"}}}}}},\"/File/CreateFile\":{\"put\":{\"summary\":\"Create File\",\"action\":\"Create File\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"share name\",\"required\":false,\"type\":\"string\"},\"directoryPath\":{\"description\":\"directory Path\",\"required\":false,\"type\":\"string\"},\"fileName\":{\"description\":\"file Name\",\"required\":false,\"type\":\"string\"},\"base64String\":{\"description\":\"file contents in base64String\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\",\"fileName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"message\":{\"description\":\"output message\",\"required\":false,\"type\":\"string\"}}}}}},\"/File/CreateDirectory\":{\"put\":{\"summary\":\"Create Directory\",\"action\":\"Create Directory\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"share name\",\"required\":false,\"type\":\"string\"},\"directoryPath\":{\"description\":\"directory Path\",\"required\":false,\"type\":\"string\"},\"directoryName\":{\"description\":\"directory Name\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\",\"directoryName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"message\":{\"description\":\"output message\",\"required\":false,\"type\":\"string\"}}}}}}},\"ok\":true,\"status\":200}"
		}
	],
	"outputs": [],
	"inputs": [],
	"handler": {
		"settings": []
	},
	"reply": [],
	"s3Prefix": "flogo",
	"key": "flogo/Azurestorage/connector/connection/connector.json",
	"display": {
		"description": "Establish connection to your azurestorage account",
		"category": "Azurestorage",
		"visible": true
	},
	"actions": [{
		"name": "Login"
	}],
	"keyfield": "name",
	"isValid": true,
	"lastUpdatedTime": 1531983035726,
	"createdTime": 1531983035726,
	"user": "flogo",
	"subscriptionId": "flogo_sbsc",
	"connectorName": " ",
	"connectorDescription": " "

}`

var inputSchemaMeDrive = `{}`
var inputSchemaMeItem = `{"itemPath":"testFolder/File&c.txt"}`
var inputSchemaUsersItem = `{"idOrUserPrincipalName":"8e4cc064-2024-4874-a274-99ad20511e10","itemPath":"testFolder/File&c.txt"}`
var inputSchemaUsersPermission = `{"idOrUserPrincipalName":"8e4cc064-2024-4874-a274-99ad20511e10","itemId":"01IBVN27NF7VPZBYYFGRFI5I2QDECIOKGP","permissionId":"261f650d-88b4-42a4-b519-28aaa94f39fa"}`
var inputSchemaSiteDrive = `{"siteId":"tibcosoftware.sharepoint.com,2c216a72-f4b2-4cfb-b2e6-61d86017c3bc,f655d14d-3a0c-4783-ab73-0fa89e70eb87"}`
var inputSchemaUsersFileDownload = `{"idOrUserPrincipalName":"8e4cc064-2024-4874-a274-99ad20511e10","itemPath":"response.pdf"}`
var inputSearchSchemaMe = `{"searchText":"Attachments"}`
var inputSchemaMeFolder = `{
	"parameters": {
		"shareName": "newshare",
		"directoryPath": "",
		"fileName": "pdf-test.pdf",
		"base64String": "aGlpaWlpaWk="
	}
}`
var inputSchemaDirFail = `{
	"parameters": {
		"shareName": "myshare",
		"directoryPath": "a/b/c/d",
		"directoryName": "abc"
	}
}`
var inputSchemaListShare = `{
	"parameters": {
		"properties": {}
	}
}`
var inputSchemaListDir = `{
	"parameters": {
		"shareName": "newshare",
		"directoryPath": "",
		"prefix": "h",
		"maxresults":2.2
	}
}`

func Test_getfile(t *testing.T) {
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.RootLogger().Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
func Test_Listshares(t *testing.T) {
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("service", "File")
	tc.SetInput("operation", "Get File")
	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaMeFolder), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.RootLogger().Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
func Test_ListDirnFiles(t *testing.T) {
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("service", "File")
	tc.SetInput("operation", "List Directories and Files")
	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaListDir), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.RootLogger().Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
func Test_getFile(t *testing.T) {
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("service", "File")
	tc.SetInput("operation", "Get File")
	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaMeFolder), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.RootLogger().Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
func Test_Get_Me_Drive(t *testing.T) {

	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Me")
	tc.SetInput("resource", "drive")
	tc.SetInput("operation", "Get Metadata")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaMeDrive), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", inputIntf)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))

	assert.Nil(t, err)
}
func Test_Get_Site_Drive(t *testing.T) {

	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Sites")
	tc.SetInput("resource", "drive")
	tc.SetInput("operation", "Get Metadata")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaSiteDrive), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))

	assert.Nil(t, err)
}

func Test_Get_Me_Item(t *testing.T) {

	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Me")
	tc.SetInput("resource", "Item")
	tc.SetInput("itemIdentifier", "getByItemPath")
	//	tc.SetInput("Query", "select=name,id")
	tc.SetInput("operation", "Get Metadata")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaMeItem), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))

	assert.Nil(t, err)
}

func Test_Get_User_Item(t *testing.T) {

	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Users")
	tc.SetInput("resource", "Item")
	tc.SetInput("itemIdentifier", "getByItemPath")
	tc.SetInput("operation", "Get Metadata")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaUsersItem), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))

	assert.Nil(t, err)
}
func Test_Get_User_Permission(t *testing.T) {

	log.RootLogger().Info("Executing Test_Get_User_Permission...")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(officeConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Users")
	tc.SetInput("resource", "Permission")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaUsersPermission), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))

	assert.Nil(t, err)
}

func Test_Get_User_FileDownload(t *testing.T) {

	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Users")
	tc.SetInput("resource", "DriveItem")
	tc.SetInput("operation", "File Download")
	//	tc.SetInput("outputType", "base64 String")
	tc.SetInput("outputType", "Location Url")
	tc.SetInput("convertToPdf", true)

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaUsersFileDownload), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	//	log.Infof("testOutput is ", testOutput)
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))

	assert.Nil(t, err)
}
func Test_Get_User_Search(t *testing.T) {
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Me")
	tc.SetInput("resource", "DriveItem")
	tc.SetInput("operation", "Search Items")
	tc.SetInput("select", "name,id")
	tc.SetInput("orderby", "name")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSearchSchemaMe), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	//	log.Infof("testOutput is ", testOutput)
	// jsonOutput, _ := json.Marshal(testOutput)
	// log.Infof("jsonOutput is : %s", string(jsonOutput))
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(testOutput); err != nil {
		log.RootLogger().Info(err)
	}
	log.Infof("json codec: %v", buf.String())

	assert.Nil(t, err)
}
func Test_Get_me_recent(t *testing.T) {
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Me")
	tc.SetInput("resource", "DriveItem")
	tc.SetInput("operation", "Fetch Recent Files")
	tc.SetInput("select", "name,id")
	tc.SetInput("orderby", "name desc")
	tc.SetInput("top", "2")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaMeDrive), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	//	log.Infof("testOutput is ", testOutput)
	// jsonOutput, _ := json.Marshal(testOutput)
	// log.Infof("jsonOutput is : %s", string(jsonOutput))
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(testOutput); err != nil {
		log.RootLogger().Info(err)
	}
	log.Infof("json codec: %v", buf.String())

	assert.Nil(t, err)
}

func Test_Get_me_recent_fail(t *testing.T) {
	var inputSchemaMeItemWithOData = `{"odataparams":{"top":2.45454}}`
	log.RootLogger().Info("****TEST : Executing InsertOne start****")
	m := make(map[string]interface{})
	err1 := json.Unmarshal([]byte(settingsjson), &m)
	assert.Nil(t, err1)
	mf := mapper.NewFactory(resolve.GetBasicResolver())
	support.RegisterAlias("connection", "connection", "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection")
	fmt.Println("=======Settings========", m["settings"])
	iCtx := test.NewActivityInitContext(m["settings"], mf)
	act, err := New(iCtx)
	assert.Nil(t, err)
	tc := test.NewActivityContext(act.Metadata())
	//Setting inputs
	tc.SetInput("officeConnection", m)
	tc.SetInput("ownership", "Me")
	tc.SetInput("resource", "DriveItem")
	tc.SetInput("operation", "Fetch Recent Files")
	tc.SetInput("select", "name,id")
	tc.SetInput("orderby", "id")
	//tc.SetInput("top", "2")

	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaMeItemWithOData), &inputIntf)

	assert.Nil(t, err2)

	//complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	//	log.Infof("testOutput is ", testOutput)
	// jsonOutput, _ := json.Marshal(testOutput)
	// log.Infof("jsonOutput is : %s", string(jsonOutput))
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(testOutput); err != nil {
		log.RootLogger().Info(err)
	}
	log.Infof("json codec: %v", buf.String())
	assert.Nil(t, err)
}
