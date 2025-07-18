package delete

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"

	snowflakedb "github.com/tibco/wi-snowflake/src/app/Snowflake/connector/connection"
)

var db *sql.DB
var connManager connection.Manager
var myActivity = &MyActivity{logger: log.ChildLogger(log.RootLogger(), "Snowflake-activity-delete"), activityName: "delete"}

var snowflakeBasicAuthConnectionJSON = []byte(`{
	"name": "snowflakeBasic",
	"description": "",
	"account": "fqb58399.us-east-1",
	"warehouse": "COMPUTE_WH",
	"database": "ESECONNECTOR",
	"schema": "ESESCHEMA",
	"authType": "Basic Authentication",
	"user": "vcastell",
	"password": "L!tmein1now4",
	"role": "SYSADMIN",
	"loginTimeout": 10
}`)

/*
   We are using a trial account for testing purpose.
   When the trial account expires, we would need to create a new account and update below two connection json.
*/

/*
   The Refresh Token validity is 90 days.
   After it is expired, we need to update below json with new refresh token and access token.
*/

/*
   ==> Execute below sql with ACCOUNTADMIN role to create an integration object.

   create security integration test_integration_object
   type = oauth
   enabled = true
   oauth_client = CUSTOM
   oauth_client_type = 'CONFIDENTIAL'
   oauth_redirect_uri = 'https://oauthdebugger.com/debug'
   oauth_issue_refresh_tokens = true
   oauth_refresh_token_validity = 86400
   OAUTH_ALLOW_NON_TLS_REDIRECT_URI = false;

   ==> Execute below sql with ACCOUNTADMIN role to get the cliend id and client secret.

   select SYSTEM$SHOW_OAUTH_CLIENT_SECRETS('TEST_INTEGRATION_OBJECT')
*/

/*
Worksheet - https://app.snowflake.com/us-east-1/fqb58399/w5AGTRnNIsTA#query
Authorization Code: Kindly copy paste below URL in the browser to get authorization code and login credentials(vcastell/L!tmein1now4):
https://fqb58399.us-east-1.snowflakecomputing.com/oauth/authorize?response_type=code&client_id=3T447dkBuaOUjDrMETzmppm8cW0%3D&scope=refresh_token&redirect_uri=https://localhost.com
*/
var snowflakeOAuthConnectionJSON = []byte(`{
	"name": "snowflakeOAuth",
	"description": "",
	"account": "fqb58399.us-east-1",
	"warehouse": "COMPUTE_WH",
	"database": "ESECONNECTOR",
	"schema": "ESESCHEMA",
	"authType": "OAuth",
	"user": "vcastell",
	"password": "L!tmein1now4",
	"clientId": "3T447dkBuaOUjDrMETzmppm8cW0=",
	"clientSecret": "uqN8N4Eh0qSYze3a8CC+rLGf7oWUoZhqOnlEZ+t4Xng=",
	"authCode": "8C644D1FDEC85C488785D9C76221DD9924A56579",
	"redirectURI" : "https://localhost.com",
	"accessToken": "SECRET:SuspFWjETvDwGNpVNJeDBlAzspDWfy36dRVwR7ddK8ELiFzY75/cw53qJrhGoR9vQCaULpv0x47do5fLZgxy26KPS6Nir41GNDslqZLRjvJbn4OWZWRcDWE6UuWEUbvbI8QuRQHsXxN6m9PwjXGt8QOPCXbkHxH5C1QfMtx7SYI3WZHNzOIeUmd13HbZARoh5F6OFL7s9beAXZ60GnA6rFrkCXKULBw1YUr0UZW+OSvs8u9RDToILeAJwb8+PWPhpr1HBnxEhlxtvrJPO5mQxxbjgJyNX9bsOZxhtWLO9Qv96NNzXDJPM2JizBhcXjj5BldqNB6lg0XnBJwtu4iWflyM7eY=",
    "accessTokenExpiry": 1660143392631,
    "refreshToken": "SECRET:zsKa5yCtKAb4kNWCp5IquxhtP2fXhiOT40Jh3Geb8TXr5R87xiCdyM4MBHDFddZzAP/yJBMugyunluugjxKFxUY5lLsjo8CuT/znIZtUc3ofAbIjme+DaJJNOCUzaz9DRLITA4LV3vep0L+6qPd0d10r1+RFW+xKClyhF4f2+s42RYRI1gGEkQCkpZyQd3ulxOMix3Ubsh+1PxD2skTwRKP7aFwq192uNa5ENphVT703vaIaa8yv66xlCVViF6VcycRw8PNcW7gvoa/YnnYfhh1VLOSGMg1OVb65wtbXXt6RXaALrgcJq4ZvqAq4awa3kRxgsPEdot+qPi2LFs07Dm8oWRuagUaWeL+6NZmlKBN87TZy3EptdpTMxKA1Fr/xwDnEJTmNZLnNcjZz0NxSAewT4m/DEbfbB4RBwpS8qPEsFJJw8blTwUPXVpuL1DxCfjV7NIrCv2io0DjyIWo9YwH8kCTZNc3/6FdhJrqmonVAq7ZsGClkp5LHul1uDsKvbhhLCpkdUsflkPMiUz5Svz0ZdpOaATn6cRbZ1ZVLeWAEZW0L7srGBR7SBD3IxDTUm8M9UDORVsAlkjCh1OP5mgHty860/1XkmxqzZ5/ckh+CIFY6CV0+lTmOL2bF4H4i",
    "refreshTokenExpiry": 1660412792631,
    "role": "",
    "loginTimeout": 10,
    "codeCheck": "2B9E1F32AA5407F3BEF4805E16222027F7FA01E4"
}`)

var snowflakeOktaOAuthConnectionJSON = []byte(`{
	"name": "snowflakeOktaOAuth",
	"description": "",
	"account": "fqb58399.us-east-1",
	"warehouse": "COMPUTE_WH",
	"database": "ESECONNECTOR",
	"schema": "ESESCHEMA",
	"authType": "OAuth",
	"provider" : "Okta with PKCE",
	"clientId": "0oaa3k25ryZIiddFI5d7",
	"authCode": "mvny4wzfOu_UTF3deuEthmGJ1fP99_Vc3y0j36US4c8",
	"scope": "session:role:sysadmin offline_access",
    "oktaCodeVerifier": "M25iVXpKU3puUjFaYWg3T1NDTDQtcW1ROUY5YXlwalNoc0hhakxifmZHag",
    "oktaCodeChallenge": "qjrzSW9gMiUgpUvqgEPE4_-8swvyCtfOVvg55o5S_es",
	"redirectURI" : "https://integration.local.cic2.pro",
	"oktaAccessToken": "SECRET:+r81KL4SCHfI0Abfl4T5swcdwWsv1EUGX/WFINgxOktvCTn652I10X24oOeZWlRLwID5iCr/LGsXTD6VhOYtM8TpsfkFazbRI2MWVL5I3Mxg0wvQc6w8kAntCOPSyJLPnwqwT9zmgbN+r8r4omlxh7R0yOjGkQZdIp314MwHI5+UZvnA866IqhYVAnWInONr7HHlPYiRWLZURGdDVvMV2N/0pEGGrdTh04TMHL2KZLxNiMZuUC8hW+JF3Lp/6dxLD7xsLSI2YMVGXvH1jkRk04DHAEveTVXPUbwzLLRMcjEzj+JFDnT0xicx58dEMN7VoJLpz3YfnpH1jBugrEwcogwpRD9aY9mB/lfdQjs5oiIJ1861vlGgJkZr1+ZALk0xUt73M7v1rcUv37wMbNRynbbhcq/oB7LCVeFGrPfyxOJALvMVrWmh1uW2nL91I2n4iKD4ssBauGNMf9x+34SJTjbtUhCYSSXKz/UzA0/20h4v6tn/uBnItp++46gQa8FdOSdbEiyI0d7l53Bg/oQscggXxGAoSfslpJYfS88hR0YbEexr6c3lC0G9XrLhW64PmGGrIGMKg/3KBiC8OVbf+QAC/MB30nh2UR5g5T+gxWvIacT+/wzgVvHIo4eYC9eolT3UpRiyAVbVG6ot3qe3Tj6Capkt1nISgBTT1pFsB+x32U2cTFK9zefRy/E5ts0Wkq4ShrKu9PrLFiXg16ex7jokYG1GIQLQxxqnMg9Vl9Aa1zrj4WONZkzm3chceQ7aEVdRYd7rE6u6vjwBDofQ0V8bBIAf6W2x+8ZU726p/CdGWVEKSUV8e/o6sotmeJ0Od0mYyu13jZBgECg55//6un44i9bhbkqdiABShDPshAXeLqqi0bxUDKztkGVj+GYQiFWQ0oAicv+UDwtWd3BQkk1qaXGHk7NLHDMoC6/WxWNX0f4JuFHlu0Bn54Sut/rApYVNMqCOypV5rvB5NSwxaybHqR9bxKCmX9TAtwOx14KFYJuRQeeE8wUibEq4fJW78Epj5CkvHOumGrNL1QbsTK25w4NE4nAdSTcJ8VzU44SvZxUuGYufz8d2ohs1m4SyMIMmvjtYdXjsxe9EKg5b2AnhbFgTepWqRThYKeXL6jc1JL2xuyS5ngTitGRGU2tgHz5geQTcZQ1MOf4+GHMKdwdLHGQ5Kzdy8aEEDwTG8fgHh7tF97+qfGqkvlWEmo1HTQvfBEiffHFQvF/Wvf1l1uScp22YWcnHWqu1rdvxYe36qNah+qMdHTywJtBhAnoKup9Ke2NAqbi0yfxhbaYWEi5vpzKQsH8t",
    "oktaAccessTokenExpiry": 1687944041000,
	"oktaRefreshToken": "4Mwvacpw7ZQar37nCxJktb9Jlxel1cZUXGt5IVzCB78",
	"loginTimeout": 10,
	"oktaTokenEndpoint": "https://dev-91275340.okta.com/oauth2/ausa3k6f8ifq4xY5x5d7/v1/token",
    "codeCheck": "xO3486VQiCT3t_14AUKrs7DFQlg54o933gTXZlhNO1w"
}`)

func TestRegister(t *testing.T) {
	ref := activity.GetRef(&MyActivity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func setup() {
	conn := make(map[string]interface{})

	//comment and uncomment based on type of connection JSON you want to use - Basic or OAuth
	//err := json.Unmarshal([]byte(snowflakeBasicAuthConnectionJSON), &conn)
	err := json.Unmarshal([]byte(snowflakeOAuthConnectionJSON), &conn)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sdb := &snowflakedb.SnowflakeFactory{}
	connManager, err = sdb.NewManager(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db = connManager.GetConnection().(*sql.DB)

	createFLOGOEMPDELETETable := `create or replace table FLOGOEMP_DELETE (ID number, NAME string);`
	insertIntoFLOGOEMPDELETE := `insert into FLOGOEMP_DELETE values (1, 'tom'), (2, 'chris')`
	db.Exec(createFLOGOEMPDELETETable)
	db.Exec(insertIntoFLOGOEMPDELETE)

	createFLOGOSTUDENTDELETETable := `create or replace table FLOGOSTUDENT_DELETE (ROLLNUM number, STDNAME string, MARKS number);`
	insertIntoFLOGOSTUDENTDELETE := `insert into FLOGOSTUDENT_DELETE values (1,'Tom',23), (2,'Tom',25), (3,'John',20), (100, 'Hero', 100)`
	db.Exec(createFLOGOSTUDENTDELETETable)
	db.Exec(insertIntoFLOGOSTUDENTDELETE)
}

func shutdown() {
	if db != nil {
		db.Close()
	}
}

func TestMain(m *testing.M) {
	//set logging to debug level
	log.SetLogLevel(myActivity.logger, log.DebugLevel)

	setup()
	code := m.Run()
	shutdown()
	os.Exit(code)
}

func TestDeleteSQLNoParams(t *testing.T) {
	query := `DELETE FROM FLOGOEMP_DELETE;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestDeleteSQL(t *testing.T) {
	query := `DELETE FROM FLOGOEMP_DELETE WHERE ID = ?ID;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "ID": 1
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestDeleteSQLRepeatedParameters(t *testing.T) {
	query := `DELETE FROM FLOGOSTUDENT_DELETE WHERE ROLLNUM = ?input AND MARKS = ?input;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "input": 100
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func PreparedAndExecuteSQL(t *testing.T, query string, inputParams map[string]interface{}) {
	tc := test.NewActivityContext(myActivity.Metadata())

	aInput := &Input{Connection: connManager, Query: query, Input: inputParams}
	tc.SetInputObject(aInput)
	ok, _ := myActivity.Eval(tc)
	assert.True(t, ok)
	aOutput := &Output{}
	err := tc.GetOutputObject(aOutput)
	assert.Nil(t, err)

	if err != nil {
		t.Errorf("Could not execute prepared query %s", query)
		t.Fail()
	}
}
