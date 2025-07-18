package dfpipeline

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata
var dfConnectionJSON = `{
	"id": "f2827250-c180-11e8-acae-a108d6106722",
	"type": "flogo:connector",
	"version": "1.0.0",
	"name": "Azure",
	"inputMappings": {},
	"outputMappings": {},
	"title": "Microsoft Azure Connector",
	"description": "Establish connection to your Azure account",
	"ref": "git.tibco.com/git/product/ipaas/wi-azureconnection.git/src/app/Azure/connector/connection",
	"settings": [{
			"name": "name",
			"type": "string",
			"required": true,
			"display": {
				"name": "Connection Name",
				"visible": true
			},
			"value": "lklk"
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
			"name": "tenantId",
			"type": "password",
			"required": true,
			"display": {
				"name": "Tenant ID",
				"visible": true,
				"encryptable": true
			},
			"value": "cde6fa59-abb3-4971-be01-2443c417cbda"
		},
		{
			"name": "clientID",
			"type": "password",
			"required": true,
			"display": {
				"name": "Client ID",
				"visible": true,
				"encryptable": true
			},
			"value": "dcf74962-7433-4ac4-9e4d-9a9c05c1af72"
		},
		{
			"name": "userName",
			"type": "string",
			"required": true,
			"display": {
				"name": "User Name",
				"visible": true
			},
			"value": "ravgupta@tibco.com"
		},
		{
			"name": "password",
			"type": "password",
			"required": true,
			"display": {
				"name": "Password",
				"visible": true,
				"encryptable": true
			},
			"value": "Tibco@201*"
		},
		{
			"name": "WI_STUDIO_OAUTH_CONNECTOR_INFO",
			"type": "string",
			"required": true,
			"display": {
				"visible": false,
				"readonly": false,
				"valid": true
			},
			"value": "{\"token_type\":\"Bearer\",\"scope\":\"user_impersonation\",\"expires_in\":\"3599\",\"ext_expires_in\":\"0\",\"expires_on\":\"1538121835\",\"not_before\":\"1538117935\",\"resource\":\"https://management.azure.com\",\"access_token\":\"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Imk2bEdrM0ZaenhSY1ViMkMzbkVRN3N5SEpsWSIsImtpZCI6Imk2bEdrM0ZaenhSY1ViMkMzbkVRN3N5SEpsWSJ9.eyJhdWQiOiJodHRwczovL21hbmFnZW1lbnQuYXp1cmUuY29tIiwiaXNzIjoiaHR0cHM6Ly9zdHMud2luZG93cy5uZXQvY2RlNmZhNTktYWJiMy00OTcxLWJlMDEtMjQ0M2M0MTdjYmRhLyIsImlhdCI6MTUzODExNzkzNSwibmJmIjoxNTM4MTE3OTM1LCJleHAiOjE1MzgxMjE4MzUsImFjciI6IjEiLCJhaW8iOiI0MkJnWUtoZmRsQm0wL3VHZGJ2NVhUOGF5ZlV5L2N4NXBiZXZUbUR0b2I3QzZ4K2N0SlVBIiwiYW1yIjpbInB3ZCJdLCJhcHBpZCI6IjRmMThhOGEzLTgxZTUtNDc4OC04YzIwLWRkZWM1YmUwMTdkNSIsImFwcGlkYWNyIjoiMCIsImZhbWlseV9uYW1lIjoiUGVkZGFkYSIsImdpdmVuX25hbWUiOiJWYXJlbnlhIiwiZ3JvdXBzIjpbImQ2MzU0YzZhLTA2NjMtNGZlNi1iMDM4LTU4OGM1NzhkMzc2MyJdLCJpcGFkZHIiOiIxMjUuMjEuNTUuNiIsIm5hbWUiOiJWYXJlbnlhIFBlZGRhZGEiLCJvaWQiOiIwNmY1YTNkYy1iMzIzLTRiN2EtYTkwOS1kYjRhNDM4NjdmZTYiLCJvbnByZW1fc2lkIjoiUy0xLTUtMjEtMjQ1OTk4NjMyMS0xNDU2MDUxOTk5LTcwMzM0NzExNC00NDAwNSIsInB1aWQiOiIxMDAzN0ZGRUE1OUQxOTQwIiwic2NwIjoidXNlcl9pbXBlcnNvbmF0aW9uIiwic3ViIjoiOEV4RHd4RjJqSjVBa05MLW00end0eGhUR3VWanRkd0RjTFY3ZDRidGFvUSIsInRpZCI6ImNkZTZmYTU5LWFiYjMtNDk3MS1iZTAxLTI0NDNjNDE3Y2JkYSIsInVuaXF1ZV9uYW1lIjoidnBlZGRhZGFAdGliY28uY29tIiwidXBuIjoidnBlZGRhZGFAdGliY28uY29tIiwidXRpIjoicmRIU3VtTTRqVVN5ZFlkVVU5b2lBUSIsInZlciI6IjEuMCIsIndpZHMiOlsiY2YxYzM4ZTUtMzYyMS00MDA0LWE3Y2ItODc5NjI0ZGNlZDdjIl19.dbTcRp3UVF5QDqt2a89EygzBUmir7s0fBUOXtERveZJ0kmyGmAvtSDX6TPUbaVyOurwqcx8_9wxzT0ebLQ1sCOgTMd26l7ElscsWcY-LXc9P0S6RZUVRhFde9pbAbeC4Ma5ZOkz7GUeqU_j78tb7WoOq6MfhlcWLp9x9IKJVCtwgAaeNJgoBTRQkSP0Svt8wwRG4rVRzbnJBauVG6AAzf8z2ajScqnVG7AlYg5EIPNJqesiZjwLX30Wu9iDW9ecz07wgPLZxn-TTBR1Ziwgr2S8DbhXfL9LHiLgP3j3Rpbm70YCgKK4GgP0JrevkUHEUX1lN-CChwuubAMAQgcs7Ng\",\"refresh_token\":\"AQABAAAAAADXzZ3ifr-GRbDT45zNSEFEmCsF2SLtUDKWLOUoDRk1ViyC4UbNeQd6ogS_3vlRple1FAalzn_h5NuqZThAqgL_YFsQ8_gnWhbHJliebv9c1Ic2wfPD-CvoXu463nRyZOvbovCY6JlAxQ9pPoemjHf5N0oEdTWi7tmkvNSCpT0a1KdVkQEAoBdsrI9MStU9x8KVyKal5ghGm3BDP2bzwU3D_67ltoszBLVzHtUw6z7C3NlkJe827Cfg83gdn_j2ZgIedRrXnsvx-6zIm9Cb3Wo3vXhZlywyyuYqi4vtLsWEbl77I7MAhB-eYOnSGy93304yWhFryD0TWZwhZY8hjs-Rw9QkaoxiwpyChFl5qnkLKqOkxLgOT_drDuN2ENw1p73CGJOaoHBb3KNIa1rMVK2KwlzFUi30irNwc0eI06NcawlL0wHIkume3p7o940xDiVcPmd5eJfOMu1H-1y7pTJGSYre1UWBKxVxp1cd6-zE57DLUSHOKuIEVbpNR-yvQr72BNT8Ub9ZAGHf8GGf4f6azkvQcxX7wXQXRciKSBUycpq6itcmLARbsIod0Vc5cnZFoek4u_zyO5-fia9F6kge5acdqnl_pP4_lorET6TYSfhZXj8VnYo7-4OnVkju7L6vD5w0YlbHDjvO8Z8-usPlxl56rVRyZsmOtHEg-Ber9h4iZ8_RWpVx93np0M570mKx1JC_jYOOhEBHUFM8JFzQEH-S3NU1vwYyRLuICWd6Nnl_bQOd_QPBe_2RGMwd-NggAA\",\"id_token\":\"eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiI0ZjE4YThhMy04MWU1LTQ3ODgtOGMyMC1kZGVjNWJlMDE3ZDUiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9jZGU2ZmE1OS1hYmIzLTQ5NzEtYmUwMS0yNDQzYzQxN2NiZGEvIiwiaWF0IjoxNTM4MTE3OTM1LCJuYmYiOjE1MzgxMTc5MzUsImV4cCI6MTUzODEyMTgzNSwiYW1yIjpbInB3ZCJdLCJmYW1pbHlfbmFtZSI6IlBlZGRhZGEiLCJnaXZlbl9uYW1lIjoiVmFyZW55YSIsImlwYWRkciI6IjEyNS4yMS41NS42IiwibmFtZSI6IlZhcmVueWEgUGVkZGFkYSIsIm9pZCI6IjA2ZjVhM2RjLWIzMjMtNGI3YS1hOTA5LWRiNGE0Mzg2N2ZlNiIsIm9ucHJlbV9zaWQiOiJTLTEtNS0yMS0yNDU5OTg2MzIxLTE0NTYwNTE5OTktNzAzMzQ3MTE0LTQ0MDA1Iiwic3ViIjoibTdSeS1KU2J6Y2NtckVFOEhWYzZTNVBPMkdJdjBCVW56eGZzWDA4ZzJLZyIsInRpZCI6ImNkZTZmYTU5LWFiYjMtNDk3MS1iZTAxLTI0NDNjNDE3Y2JkYSIsInVuaXF1ZV9uYW1lIjoidnBlZGRhZGFAdGliY28uY29tIiwidXBuIjoidnBlZGRhZGFAdGliY28uY29tIiwidmVyIjoiMS4wIn0.\"}"
		},
		{
			"name": "configProperties",
			"type": "string",
			"required": true,
			"display": {
				"visible": false
			},
			"value": ""
		}
	],
	"outputs": [],
	"inputs": [],
	"handler": {
		"settings": []
	},
	"reply": [],
	"s3Prefix": "flogo",
	"key": "flogo/Azure/connector/connection/connector.json",
	"display": {
		"description": "Establish connection to your Azure account",
		"category": "Azure",
		"visible": true
	},
	"actions": [{
		"name": "Login"
	}],
	"keyfield": "name",
	"isValid": true,
	"lastUpdatedTime": 1537962020341,
	"createdTime": 1537962020341,
	"user": "flogo",
	"subscriptionId": "flogo_sbsc",
	"connectorName": " ",
	"connectorDescription": " "
}`
var inputSchemaStartTrigger = `{
	"parameters": {
		"factoryName": "tibcosecondfactory",
		"pipelineRunId": "28b0517f-2d67-4657-a455-6bdca4616ad8",
		"resourceGroupName": "johnsorg",
		"subscriptionId": "3d9f846f-9849-4329-a35e-91fa5ea16488"
	}
}`
var inputSchemaCreatePipelineRun = `{
	"parameters": {
		"factoryName": "tibcosecondfactory",
		"pipelineName": "pipeline2",
		"resourceGroupName": "johnsorg",
		"subscriptionId": "3d9f846f-9849-4329-a35e-91fa5ea16488"
	}
}`
var inputSchemaQueryPipelineRuns = `{
	"parameters": {	
		"lastUpdatedBefore":"2018-10-16T06:36:44.3345758Z",
		"lastUpdatedAfter":"2018-10-10T06:49:48.3686473Z"
	}
}`

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
	}
	return activityMetadata
}

func Test_QueryPipeline(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace start****")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(dfConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("operation", "Query Runs")
	tc.SetInput("subscriptionId", "3d9f846f-9849-4329-a35e-91fa5ea16488")
	tc.SetInput("resourceGroup", "johnsorg")
	tc.SetInput("dataFactories", "tibcothirdfactory")
	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaQueryPipelineRuns), &inputIntf)

	assert.Nil(t, err2)

	complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
func Test_CancelPipeline(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace start****")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(dfConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("operation", "Cancel pipeline Run")
	tc.SetInput("subscriptionId", "3d9f846f-9849-4329-a35e-91fa5ea16488")
	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaStartTrigger), &inputIntf)

	assert.Nil(t, err2)

	complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}

func Test_createPipelineRun(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace start****")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(dfConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("operation", "Create Pipeline Run")
	tc.SetInput("subscriptionId", "3d9f846f-9849-4329-a35e-91fa5ea16488")
	var inputIntf interface{}
	err2 := json.Unmarshal([]byte(inputSchemaCreatePipelineRun), &inputIntf)

	assert.Nil(t, err2)

	complex := &data.ComplexObject{Metadata: "", Value: inputIntf}
	tc.SetInput("input", complex)
	//Executing activity
	_, err = act.Eval(tc)
	//Getting outputs
	testOutput := tc.GetOutput("output")
	jsonOutput, _ := json.Marshal(testOutput)
	log.Infof("jsonOutput is : %s", string(jsonOutput))
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace ends****")
	assert.Nil(t, err)
}
