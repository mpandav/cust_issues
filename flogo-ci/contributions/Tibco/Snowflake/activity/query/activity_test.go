package query

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
var myActivity = &MyActivity{logger: log.ChildLogger(log.RootLogger(), "Snowflake-activity-query"), activityName: "query"}

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
	"role": "ACCOUNTADMIN",
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
	"authCode": "4500AEFA7709547AB9A8EBD570680587F1EE2B97",
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
	"account": "tibco_partner",
    "warehouse": "DEMO_WH",
    "database": "PLUGINDB",
    "schema": "QA",
	"authType": "OAuth",
	"provider" : "Okta with PKCE",
	"clientId": "0oaa3k25ryZIiddFI5d7",
	"authCode": "Rg_QVIiBcZcN00-mIVLKdW8E3YxeB4tgoeTfFP2IuZM",
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
	err := json.Unmarshal([]byte(snowflakeBasicAuthConnectionJSON), &conn)
	//err := json.Unmarshal([]byte(snowflakeOAuthConnectionJSON), &conn)
	// err := json.Unmarshal([]byte(snowflakeOktaOAuthConnectionJSON), &conn)
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

	createFLOGOEMPTable := `create or replace table FLOGOEMP (ID number, NAME string);`
	insertIntoFLOGOEMP := `insert into FLOGOEMP values (1, 'tom'), (2, 'chris')`
	db.Exec(createFLOGOEMPTable)
	db.Exec(insertIntoFLOGOEMP)

	createFLOGOSTUDENTTable := `create or replace table FLOGOSTUDENT (ROLLNUM number, STDNAME string, MARKS number);`
	insertIntoFLOGOSTUDENT := `insert into FLOGOSTUDENT values (1,'Tom',23), (2,'Tom',25), (3,'John',20), (100, 'Hero', 100)`
	db.Exec(createFLOGOSTUDENTTable)
	db.Exec(insertIntoFLOGOSTUDENT)

	createFLOGODATATYPESTable := `create or replace table FLOGODATATYPES (NUM number, DEC decimal(20,2), NUMERIC numeric(30,3),INT int, INTEGER integer, D double, F float, DP double precision, R real, B boolean, V varchar, C char, S string, TXT text, BI binary, VB varbinary, T time, T_TZ timestamp_tz, T_LTZ timestamp_ltz, D1 date, VAR variant, ARR array, OBJ object);`
	insertFLOGODATATYPESTable := `INSERT INTO FLOGODATATYPES SELECT 691,654.88,98.56,238,654,867.987,654.876,87.9867,654.876,FALSE,'TEST DELETE ACTIVITY','t','test','test','b5eb2d','b5eb2d',
	'21:32:52','2001-10-26 12:32:52.123 -0700','2001-10-26 12:32:52.123 -0700','2022-06-06',PARSE_JSON('{"key1":"value1","key2":"value2"} '),ARRAY_CONSTRUCT(1,2,3),PARSE_JSON('{"outer_key1":{"inner_key1A":"1a","inner_key1B":"1b"},'||'"outer_key2":{"inner_key2":2}}');`
	db.Exec(createFLOGODATATYPESTable)
	db.Exec(insertFLOGODATATYPESTable)

	createVARIANTTable1 := `create or replace table car_sales1
    (
      src variant
    )
    as
    select parse_json(column1) as src
    from values
    ('{
        "date" : "2017-04-28",
        "dealership" : "Valley View Auto Sales",
        "salesperson" : {
          "id": "55",
          "name": "Frank Beasley"
        },
        "customer" : [
          {"name": "Joyce Ridgely", "phone": "16504378889", "address": "San Francisco, CA"}
        ],
        "vehicle" : [
          {"make": "Honda", "model": "Civic", "year": "2017", "price": "20275", "extras":["ext warranty", "paint protection"]}
        ]
    }'),
    ('{
        "date" : "2017-04-28",
        "dealership" : "Tindel Toyota",
        "salesperson" : {
          "id": "274",
          "name": "Greg Northrup"
        },
        "customer" : [
          {"name": "Bradley Greenbloom", "phone": "12127593751", "address": "New York, NY"}
        ],
        "vehicle" : [
          {"make": "Toyota", "model": "Camry", "year": "2017", "price": "23500", "extras":["ext warranty", "rust proofing", "fabric protection"]}
        ]
    }') v;`
	db.Exec(createVARIANTTable1)

	createVARIANTTable2 := `create or replace table colors (v variant);`
	insertIntoVARIANTTable2 := `insert into
    colors
    select
       parse_json(column1) as v
    from
    values
      ('[{r:255,g:12,b:0},{r:0,g:255,b:0},{r:0,g:0,b:255}]'),
      ('[{c:0,m:1,y:1,k:0},{c:1,m:0,y:1,k:0},{c:1,m:1,y:0,k:0}]')
     v;`
	db.Exec(createVARIANTTable2)
	db.Exec(insertIntoVARIANTTable2)

	createDemoTable := `create table demonstration1 (
        id integer,
        array1 array,
        object1 object
        );`
	insertIntoDemoTable := `insert into demonstration1 (id, array1, object1) 
		select 
		  1, 
		  array_construct(1, 2, 3),
		  parse_json(' { "outer_key1": { "inner_key1A": "1a", "inner_key1B": "1b" }, "outer_key2": { "inner_key2": 2 } } ');`
	db.Exec(createDemoTable)
	db.Exec(insertIntoDemoTable)

	createBinaryTable := `create or replace table demo_binary1(i int, v varchar, b binary, vb varbinary);`
	insertIntoBinaryTable1 := `insert into demo_binary1 (i, v, b, vb) select 1, 'AB', to_binary('AB'), to_binary('AB');`
	insertIntoBinaryTable2 := `insert into demo_binary1 (i, v, b, vb) select 2, 'TEST', to_binary(base64_encode('TEST'), 'BASE64'), to_binary(base64_encode('TEST'), 'BASE64');`
	insertIntoBinaryTable3 := `insert into demo_binary1 (i, v, b, vb) select 3, 'TEST', to_binary(hex_encode('TEST'), 'HEX'), to_binary(hex_encode('TEST'), 'HEX');`
	insertIntoBinaryTable4 := `insert into demo_binary1 (i, v, b, vb) select 4, 'TEST', to_binary('TEST', 'UTF-8'), to_binary('TEST', 'UTF-8');`
	insertIntoBinaryTable5 := `insert into demo_binary1 (i, v, b, vb) select 5, 'FLOGO', to_binary('FLOGO', 'UTF-8'), to_binary('FLOGO', 'UTF-8');`
	db.Exec(createBinaryTable)
	db.Exec(insertIntoBinaryTable1)
	db.Exec(insertIntoBinaryTable2)
	db.Exec(insertIntoBinaryTable3)
	db.Exec(insertIntoBinaryTable4)
	db.Exec(insertIntoBinaryTable5)
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

func TestNoInputParamSQL1(t *testing.T) {
	query := `select * from FLOGOEMP;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

/*func TestAllDatatypeSQL(t *testing.T) {
	query := `select * from FLOGODATATYPES;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestWhereClauseSQL(t *testing.T) {
	query := `select ID, NAME from FLOGOEMP where ID = ?ID;`

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

func TestWhereClauseSQLRepeatedParameters(t *testing.T) {
	query := `select * from FLOGOSTUDENT where ROLLNUM = ?a AND MARKS = ?a;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "a": 100
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestMultipleConditionalClauseSQL(t *testing.T) {
	query := `select * from FLOGOSTUDENT where ROLLNUM = ?ROLLNUM AND STDNAME = ?STDNAME;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
        "parameters": {
            "ROLLNUM": 1,
            "STDNAME": "Tom"
        }
    }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestMaxFunc(t *testing.T) {
	query := `select max(ROLLNUM) as STUDENT_ID from FLOGOSTUDENT;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestMultiSelect(t *testing.T) {
	query := `select * from FLOGOEMP where ID = (select min(ROLLNUM) from FLOGOSTUDENT);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
         }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectDistinct(t *testing.T) {
	query := `select distinct(ROLLNUM) from FLOGOSTUDENT;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectOrderBy(t *testing.T) {
	query := `select * from FLOGOSTUDENT order by ROLLNUM desc;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectSumFuncAndGroupBy(t *testing.T) {
	query := `SELECT STDNAME, SUM(MARKS) AS Total_Marks FROM FLOGOSTUDENT GROUP BY STDNAME;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype1(t *testing.T) {
	query := `select * from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype2(t *testing.T) {
	query := `select src:dealership from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype3(t *testing.T) {
	query := `select src:salesperson.name from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype4(t *testing.T) {
	query := `select src['salesperson']['name'] from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype5(t *testing.T) {
	query := `select src:vehicle[0] from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype6(t *testing.T) {
	query := `select src:vehicle[0].price from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype7(t *testing.T) {
	query := `select src:salesperson.id::string from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype8(t *testing.T) {
	query := `select
    value:name::string as "Customer Name",
    value:address::string as "Address"
    from
      car_sales1
    , lateral flatten(input => src:customer);`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype9(t *testing.T) {
	query := `select
    vm.value:make::string as make,
    vm.value:model::string as model,
    ve.value::string as "Extras Purchased"
    from
      car_sales1
    , lateral flatten(input => src:vehicle) vm
    , lateral flatten(input => vm.value:extras) ve;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype10(t *testing.T) {
	//query := `select get_path(src, 'vehicle[0]:make') from car_sales1;`
	//query := `select get_path(src, 'vehicle[0].make') from car_sales1;`
	query := `select src:vehicle[0].make from car_sales1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectVariantDatatype11(t *testing.T) {
	query := `select *, get(v, array_size(v)-1) from colors;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectDemoTable(t *testing.T) {
	query := `select * from demonstration1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectBinaryTable1(t *testing.T) {
	query := `select * from demo_binary1;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}

func TestSelectBinaryTable2(t *testing.T) {
	query := `select B, hex_decode_string(to_varchar(B)) from demo_binary1 where I = 3;`

	inputParams := make(map[string]interface{})
	var inputJSON = []byte(`{
		 "parameters": {
		 }
	 }`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)

	PreparedAndExecuteSQL(t, query, inputParams)
}*/

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
