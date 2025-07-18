/*
 * Copyright Â© 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package create

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

var settingsjson = `{
	"settings": {
	  "Connection": {
		"id": "e1e890d0-de91-11e9-aef0-13201957902e",
		"name": "azurecon",
		"ref": "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection",
		"settings": {
			  "name": "mongocon", 
			  "accountName":"csg3d9f846f9849x4329xa35",
			  "connectionType":"Enter SAS",				 
			  "sas":"",
			  "expiryDate":"",
			  "accessKey":"",
			  "WI_STUDIO_OAUTH_CONNECTOR_INFO":"{\"account_name\":\"csg3d9f846f9849x4329xa35\",\"access_key\":\"NmXONqyr8THSJUpDzPm2tkHl9miFmBgWIhLR/8O8rykO4WFeG8wdcb/vm758687r8qrJ0k+aDnahwNt7kJuMEw==\",\"expiry_date\":\"2050-12-20T21:55:07Z\",\"start_date\":\"\",\"sas_token\":\"sv=2017-11-09&ss=bfqt&srt=sco&sp=rwdlacup&se=2050-12-20T21:55:07Z&spr=https&sig=RD0fH8cHBV1NY6RhW19AFsZBEnWqmCUG1kvj%2FDyyjiQ%3D\"}"				
			  }		
		},
	"operation": "Create File",
	"service": "File",
	"path":""
}
}`

var activityMetadata *activity.Metadata

// func getActivityMetadata() *activity.Metadata {
// 	if activityMetadata == nil {
// 		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
// 		if err != nil {
// 			panic("No Json Metadata found for activity.json path")
// 		}
// 		activityMetadata = activity.ToMetadata(string(jsonMetadataBytes))
// 	}
// 	return activityMetadata
// }

// func TestActivityRegistration(t *testing.T) {
// 	act := NewActivity(getActivityMetadata())
// 	if act == nil {
// 		t.Error("Activity Not Registered")
// 		t.Fail()
// 		return
// 	}
// }

var officeConnectionJSON = `{
	 "id": "66f93b00-8987-11e8-a466-0d29bc86ee25",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "name": "Azurestorage",
	 "inputMappings": {},
	 "outputMappings": {},
	 "title": "Microsoft Azure Storage Connector",
	 "description": "Establish connection to your azurestorage account",
	 "ref": "git.tibco.com/git/product/ipaas/wi-azstorage.git/src/app/Azurestorage/connector/connection",
	 "settings": [
	  {
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
	   "value": "{\"account_name\":\"csg3d9f846f9849x4329xa35\",\"access_key\":\"NmXONqyr8THSJUpDzPm2tkHl9miFmBgWIhLR/8O8rykO4WFeG8wdcb/vm758687r8qrJ0k+aDnahwNt7kJuMEw==\",\"expiry_date\":\"2050-12-20T21:55:07Z\",\"start_date\":\"\",\"sas_token\":\"sv=2017-11-09&ss=bfqt&srt=sco&sp=rwdlacup&se=2050-12-20T21:55:07Z&spr=https&sig=RD0fH8cHBV1NY6RhW19AFsZBEnWqmCUG1kvj%2FDyyjiQ%3D\"}"
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
	   "value": "{\"service\":[\"File\",\"Table\"],\"paths\":{\"/File/createShare\":{\"put\":{\"summary\":\"Create Share\",\"action\":\"Create Share\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"create share under specified storage account\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"message\":{\"description\":\"output message\",\"required\":false,\"type\":\"string\"}}}}}},\"/File/CreateFile\":{\"put\":{\"summary\":\"Create File\",\"action\":\"Create File\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"share name\",\"required\":false,\"type\":\"string\"},\"directoryPath\":{\"description\":\"directory Path\",\"required\":false,\"type\":\"string\"},\"fileName\":{\"description\":\"file Name\",\"required\":false,\"type\":\"string\"},\"base64String\":{\"description\":\"file contents in base64String\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\",\"fileName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"message\":{\"description\":\"output message\",\"required\":false,\"type\":\"string\"}}}}}},\"/File/CreateDirectory\":{\"put\":{\"summary\":\"Create Directory\",\"action\":\"Create Directory\",\"schema\":{\"parameters\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"shareName\":{\"description\":\"share name\",\"required\":false,\"type\":\"string\"},\"directoryPath\":{\"description\":\"directory Path\",\"required\":false,\"type\":\"string\"},\"directoryName\":{\"description\":\"directory Name\",\"required\":false,\"type\":\"string\"}},\"required\":[\"shareName\",\"directoryName\"]},\"output\":{\"$id\":\"http://example.com/example.json\",\"type\":\"object\",\"definitions\":{},\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"additionalProperties\":false,\"properties\":{\"message\":{\"description\":\"output message\",\"required\":false,\"type\":\"string\"}}}}}}},\"ok\":true,\"status\":200}"
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
	 "actions": [
	  {
	   "name": "Login"
	  }
	 ],
	 "keyfield": "name",
	 "isValid": true,
	 "lastUpdatedTime": 1531807527600,
	 "createdTime": 1531807527600,
	 "user": "flogo",
	 "subscriptionId": "flogo_sbsc",
	 "connectorName": " ",
	 "connectorDescription": " "
   }`

var inputSchemaMeFolder = `{
	"parameters": {
		"shareName": "newshare1",
		"directoryPath": "",
		"fileName": "fileNew1.txt",
		"fileContent": "aGlpaWlpaWk="
	}
}`
var inputSchemaDirFail = `{
	"parameters": {
		"shareName": "myshare",
		"directoryPath": "a/b/c/d",
		"directoryName": "abc"
	}
}`

func Test_Create_file_ConflictReplace(t *testing.T) {
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
	tc.SetInput("input", `{
		"shareName": "myshare",
		"directoryPath": "" ,
		"fileName": "test-File.txt",
		"fileContent": "Test"
		  }`)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.RootLogger().Infof("jsonOutput is : %s", string(jsonOutput))
	log.RootLogger().Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}

func Test_Create_dir_fail(t *testing.T) {
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
	tc.SetInput("input", `{
		"shareName": "myshare-1",
		"directoryPath": "" ,
		"fileName": "test-File.txt",
		"fileContent": "Test"
		  }`)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.RootLogger().Infof("jsonOutput is : %s", string(jsonOutput))
	log.RootLogger().Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
