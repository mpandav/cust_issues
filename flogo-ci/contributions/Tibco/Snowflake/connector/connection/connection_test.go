package connection

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/project-flogo/core/support/log"
)

var connectionLog = log.ChildLogger(log.RootLogger(), "Snowflake-connection")

var snowflakeBasicAuthConnectionJSON = []byte(`{
	"name": "snowflakeBasic",
	"description": "",
	"account": "tibco_partner",
	"warehouse": "DEMO_WH",
	"database": "PLUGINDB",
	"schema": "QA",
	"authType": "Basic Authentication",
	"user": "vcastell",
	"password": "Tibco123$",
	"role": "",
	"loginTimeout": 10
}`)

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
	"provider" : "Snowflake",
	"user": "vcastell",
	"password": "L!tmein1now4",
	"clientId": "3T447dkBuaOUjDrMETzmppm8cW0=",
	"clientSecret": "uqN8N4Eh0qSYze3a8CC+rLGf7oWUoZhqOnlEZ+t4Xng=",
	"authCode": "DF368E1EB23A9BE4C7379152E21BAEE63308ADF9",
	"redirectURI" : "https://localhost.com",
	"accessToken": "SECRET:SuspFWjETvDwGNpVNJeDBlAzspDWfy36dRVwR7ddK8ELiFzY75/cw53qJrhGoR9vQCaULpv0x47do5fLZgxy26KPS6Nir41GNDslqZLRjvJbn4OWZWRcDWE6UuWEUbvbI8QuRQHsXxN6m9PwjXGt8QOPCXbkHxH5C1QfMtx7SYI3WZHNzOIeUmd13HbZARoh5F6OFL7s9beAXZ60GnA6rFrkCXKULBw1YUr0UZW+OSvs8u9RDToILeAJwb8+PWPhpr1HBnxEhlxtvrJPO5mQxxbjgJyNX9bsOZxhtWLO9Qv96NNzXDJPM2JizBhcXjj5BldqNB6lg0XnBJwtu4iWflyM7eY=",
    "accessTokenExpiry": 1660143392631,
    "refreshToken": "SECRET:zsKa5yCtKAb4kNWCp5IquxhtP2fXhiOT40Jh3Geb8TXr5R87xiCdyM4MBHDFddZzAP/yJBMugyunluugjxKFxUY5lLsjo8CuT/znIZtUc3ofAbIjme+DaJJNOCUzaz9DRLITA4LV3vep0L+6qPd0d10r1+RFW+xKClyhF4f2+s42RYRI1gGEkQCkpZyQd3ulxOMix3Ubsh+1PxD2skTwRKP7aFwq192uNa5ENphVT703vaIaa8yv66xlCVViF6VcycRw8PNcW7gvoa/YnnYfhh1VLOSGMg1OVb65wtbXXt6RXaALrgcJq4ZvqAq4awa3kRxgsPEdot+qPi2LFs07Dm8oWRuagUaWeL+6NZmlKBN87TZy3EptdpTMxKA1Fr/xwDnEJTmNZLnNcjZz0NxSAewT4m/DEbfbB4RBwpS8qPEsFJJw8blTwUPXVpuL1DxCfjV7NIrCv2io0DjyIWo9YwH8kCTZNc3/6FdhJrqmonVAq7ZsGClkp5LHul1uDsKvbhhLCpkdUsflkPMiUz5Svz0ZdpOaATn6cRbZ1ZVLeWAEZW0L7srGBR7SBD3IxDTUm8M9UDORVsAlkjCh1OP5mgHty860/1XkmxqzZ5/ckh+CIFY6CV0+lTmOL2bF4H4i",
    "refreshTokenExpiry": 1660412792631,
    "role": "",
    "loginTimeout": 10,
    "codeCheck": "2B9E1F32AA5407F3BEF4805E16222027F7FA01E4"
}`)

/*
Use below URL to get authorization code; login using (rpuranda@tibco.com/letme1n@2022):
https://dev-91275340.okta.com/oauth2/ausa3k6f8ifq4xY5x5d7/v1/authorize?client_id=0oaa3k25ryZIiddFI5d7&response_type=code&scope=session:role:sysadmin offline_access&redirect_uri=https://integration.local.cic2.pro&state=state-8600b31f-52d1-4dca-987c-386e3d8967e9&code_challenge_method=S256&code_challenge=qjrzSW9gMiUgpUvqgEPE4_-8swvyCtfOVvg55o5S_es
*/
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
	"authCode": "xO3486VQiCT3t_14AUKrs7DFQlg54o933gTXZlhNO1w",
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
// Test the Basic Auth Connection with Snowflake Database
/*func TestSnowflakeBasicAuth(t *testing.T) {
	//set logging to debug level
	log.SetLogLevel(connectionLog, log.DebugLevel)

	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(snowflakeBasicAuthConnectionJSON), &conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sdb := SnowflakeFactory{}
	_, err = sdb.NewManager(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestSnowflakeOAuth(t *testing.T) {
	//set logging to debug level
	log.SetLogLevel(connectionLog, log.DebugLevel)

	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(snowflakeOAuthConnectionJSON), &conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sdb := SnowflakeFactory{}
	_, err = sdb.NewManager(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}*/

func TestSnowflakeOktaOAuth(t *testing.T) {
	//set logging to debug level
	log.SetLogLevel(connectionLog, log.DebugLevel)

	conn := make(map[string]interface{})
	err := json.Unmarshal([]byte(snowflakeOktaOAuthConnectionJSON), &conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sdb := SnowflakeFactory{}
	_, err = sdb.NewManager(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}