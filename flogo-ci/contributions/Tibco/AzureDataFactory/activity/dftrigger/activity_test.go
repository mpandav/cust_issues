package dftrigger

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

	"title": "Microsoft Azure Active Directory Connector",
	"name": "AzureAD",
	"author": "TIBCO Software Inc.",
	"type": "flogo:connector",
	"version": "1.0.0",
	"display": {
		"description": "Establish connection to your Azure account",
		"category": "AzureAD",
		"visible": true
	},
	"ref": "git.tibco.com/git/product/ipaas/wi-azadconnection.git/src/app/AzureAD/connector/connection",
	"keyfield": "name",
	"settings": [{
			"name": "name",
			"type": "string",
			"required": true,
			"display": {
				"name": "Connection Name",
				"visible": true
			},
			"value": "DFtest"
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
			"name": "resourceURL",
			"type": "string",
			"required": true,
			"display": {
				"name": "Resource URL",
				"visible": true
			},
			"value": "https://management.azure.com"
		},
		{
			"name": "grantType",
			"type": "string",
			"required": false,
			"display": {
				"visible": false
			},
			"value": "password"
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
			"value": "{\"token_type\":\"Bearer\",\"scope\":\"user_impersonation\",\"expires_in\":\"3600\",\"ext_expires_in\":\"0\",\"expires_on\":\"1539683761\",\"not_before\":\"1539679861\",\"resource\":\"https://management.azure.com\",\"access_token\":\"eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6Imk2bEdrM0ZaenhSY1ViMkMzbkVRN3N5SEpsWSIsImtpZCI6Imk2bEdrM0ZaenhSY1ViMkMzbkVRN3N5SEpsWSJ9.eyJhdWQiOiJodHRwczovL21hbmFnZW1lbnQuYXp1cmUuY29tIiwiaXNzIjoiaHR0cHM6Ly9zdHMud2luZG93cy5uZXQvY2RlNmZhNTktYWJiMy00OTcxLWJlMDEtMjQ0M2M0MTdjYmRhLyIsImlhdCI6MTUzOTY3OTg2MSwibmJmIjoxNTM5Njc5ODYxLCJleHAiOjE1Mzk2ODM3NjEsImFjciI6IjEiLCJhaW8iOiJBU1FBMi84SkFBQUFGZmNKWmFVb1R5SFNRSWJwLzgyak5PRGpNZm5hU3ZwU3ZNcVZOTndvVmVBPSIsImFtciI6WyJwd2QiXSwiYXBwaWQiOiJkY2Y3NDk2Mi03NDMzLTRhYzQtOWU0ZC05YTljMDVjMWFmNzIiLCJhcHBpZGFjciI6IjAiLCJmYW1pbHlfbmFtZSI6Ikd1cHRhIiwiZ2l2ZW5fbmFtZSI6IlJhdmkiLCJncm91cHMiOlsiZDYzNTRjNmEtMDY2My00ZmU2LWIwMzgtNTg4YzU3OGQzNzYzIl0sImlwYWRkciI6IjEyMi4xNS4yMDYuMTU3IiwibmFtZSI6IlJhdmkgR3VwdGEiLCJvaWQiOiIwMWE1ODI1OS1hYmY1LTQxMGMtOTQwMy03YTg4ZjMwZTI3ZDgiLCJvbnByZW1fc2lkIjoiUy0xLTUtMjEtMjQ1OTk4NjMyMS0xNDU2MDUxOTk5LTcwMzM0NzExNC00NjgzMiIsInB1aWQiOiIxMDAzMDAwMEE1RDNCMkE1Iiwic2NwIjoidXNlcl9pbXBlcnNvbmF0aW9uIiwic3ViIjoiWC01bnBoTklib1VzTVRPYWV2WFd0d256TXg0QTV6dG4zM2FaLXhTdWQ0SSIsInRpZCI6ImNkZTZmYTU5LWFiYjMtNDk3MS1iZTAxLTI0NDNjNDE3Y2JkYSIsInVuaXF1ZV9uYW1lIjoicmF2Z3VwdGFAdGliY28uY29tIiwidXBuIjoicmF2Z3VwdGFAdGliY28uY29tIiwidXRpIjoidXhrWDloTTZkRW1zWVNYRmRieWZBQSIsInZlciI6IjEuMCIsIndpZHMiOlsiY2YxYzM4ZTUtMzYyMS00MDA0LWE3Y2ItODc5NjI0ZGNlZDdjIl19.aXq1__7YTq2VycRbdkiFGPsM0EnOAxG1w9w--1RaX1VCpzcxNemqvbaHbm6HYbLm5xAK28NWlaGynNNGQkTH8aZVOBkXSJeSCzD72xivBz9Jpnm33SfS3SyakANxdnkADhx8puux94m1zrxEPB9-MoNgduPz8PZ8M28R4vEk-jUMAPSK3Er6NHX0Zfc1SA-QwqVD9o404uNo1MIObUme1aUlv3rVh3Ii9TZmG3UXrT2CC7rLDcxMDNfxGv-CCwKMryOwD3ePdr_0lBe5AMC7GiqOBDWqGxpq1ANdBEbiTqaSFJYiT7P2R-F968Fh7oNQqfH066u0kFWPn7S9ZswQzw\",\"refresh_token\":\"AQABAAAAAAC5una0EUFgTIF8ElaxtWjTMcKT_T6FfjaTP2aHjJ6Ht-QFC92F5KfQnma20FIz02ezyRcYolF5vm31vIb4f_SuEEUtDIfVNTk1fBCoKNLrUJf7p3gLTY043-fWSeT_hh0Ng-rvoQvc6_ZCa6uDUjTOUmjeGKzgpqAICuYeYNhwt4-jkLoWPdCc6o7AfodG0fkqziyVomClhqQxv2F-nnkzpA3FYx-8yRkonsFfT42UUPo5Z-7V_75MCIUo9oOFlMnA_1wY-4NoqCCN2NVJTft3ZsAP11WKWXT94RuHEuRKkgj_oooJi5a4yr-QR-kdYzNzEDcFisP6BM7fy-xDJUgJZjC1_Xmbl-hVnoZYgUoFagK4YraF9XD-9oY7zaMdFZcaxQAdHPsvWK3CdQgGvWljmWBvWwrnNxGpuVlDenD_tHeRrstO_FkxW0rAZl8dXSpeBEeLVM9r2AukGSpWwq72nlgIPsRF83PcAgTWImmhnv7ucn1ofderN5NdxrY9iRTXRMo9peGhwF9vpXcdGvv9Xefo0Tqpoy-2kk0J8d54alyD-TKcKwJMQo_jJUTzLGvCpiKU9w4-rEWS-bugaDhIahv_mEdT_c8rxfrAWQgB5d7ve1zx64EPZ1Py7b6nYq0F45XKnxodO_eCnuMpiGzWpEOI5yulixPdOPBUnAwgKyO5BkJO0ujgDGKc4xAsJLO-TRx_NSI8azcpq4BFCh3W7xZ72z-6Ej4qOF5hhoKFHbO0ZId4UWIAXXdCr-oCoD8gAA\",\"id_token\":\"eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiJkY2Y3NDk2Mi03NDMzLTRhYzQtOWU0ZC05YTljMDVjMWFmNzIiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9jZGU2ZmE1OS1hYmIzLTQ5NzEtYmUwMS0yNDQzYzQxN2NiZGEvIiwiaWF0IjoxNTM5Njc5ODYxLCJuYmYiOjE1Mzk2Nzk4NjEsImV4cCI6MTUzOTY4Mzc2MSwiYW1yIjpbInB3ZCJdLCJmYW1pbHlfbmFtZSI6Ikd1cHRhIiwiZ2l2ZW5fbmFtZSI6IlJhdmkiLCJpcGFkZHIiOiIxMjIuMTUuMjA2LjE1NyIsIm5hbWUiOiJSYXZpIEd1cHRhIiwib2lkIjoiMDFhNTgyNTktYWJmNS00MTBjLTk0MDMtN2E4OGYzMGUyN2Q4Iiwib25wcmVtX3NpZCI6IlMtMS01LTIxLTI0NTk5ODYzMjEtMTQ1NjA1MTk5OS03MDMzNDcxMTQtNDY4MzIiLCJzdWIiOiJFZzRwcG96UU5adHNkZlA5RnRYLXpOUUdxOHVwSk5McWJFQlA4Mk5iRkNvIiwidGlkIjoiY2RlNmZhNTktYWJiMy00OTcxLWJlMDEtMjQ0M2M0MTdjYmRhIiwidW5pcXVlX25hbWUiOiJyYXZndXB0YUB0aWJjby5jb20iLCJ1cG4iOiJyYXZndXB0YUB0aWJjby5jb20iLCJ2ZXIiOiIxLjAifQ.\"}"
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
	"actions": [{
		"name": "Login"
	}],
	"s3Prefix": "flogo",
	"lastModifiedDate": "2018-09-28T09:24:58.854Z",
	"key": "flogo/AzureAD/connector/connection/connector.json",
	"isValid": true,
	"lastUpdatedTime": 1539680161564,
	"createdTime": 1538562570208,
	"user": "flogo",
	"subscriptionId": "flogo_sbsc",
	"id": "361fde00-c6f7-11e8-acae-a108d6106722",
	"connectorName": "DFtest",
	"connectorDescription": " "

}`
var inputSchemaStartTrigger = `{
	"parameters": {
		"factoryName": "tibcosecondfactory",
		"triggerName": "trigger2",
		"resourceGroupName": "johnsorg"
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
func Test_QyeryTriggerRuns(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace start****")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(dfConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("operation", "Query Trigger Runs")
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
func Test_startTrigger(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace start****")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(dfConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("operation", "Activate/Start Trigger")
	tc.SetInput("resourceGroup", "johnsorg")
	tc.SetInput("dataFactories", "tibcosecondfactory")
	tc.SetInput("dfTrigger", "trigger2")
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

func Test_stopTrigger(t *testing.T) {
	log.SetLogLevel(logger.DebugLevel)
	log.Info("****TEST : Executing Create folder test for testing conflict behavior replace start****")
	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(dfConnectionJSON), &m)
	assert.Nil(t, err)
	//Setting inputs
	tc.SetInput("Connection", m)
	tc.SetInput("operation", "De-Activate/Stop Trigger")
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
