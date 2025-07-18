/*
 * Copyright Â© 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

package update
/***
import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"testing"

	"git.tibco.com/git/product/ipaas/wi-postgres.git/src/app/PostgreSQL/connector/connection/connection"
	"git.tibco.com/git/product/ipaas/wi-postgres.git/src/app/github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
	"github.com/project-flogo/core/activity"
)

var activityMetadata *activity

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

func TestUpdateStatements(t *testing.T) {
	// log.SetLogLevel(logger.DebugLevel)

	var tests = []struct {
		id       string
		query    string
		input    string
		expected string
		errorMsg string
	}{
		{`UpdateSelectTest`,
			`UPDATE accounts SET (contact_first_name, contact_last_name) = ((SELECT first_name, last_name FROM salesmen WHERE (salesmen.id = accounts.sales_id)) );`,
			`{"parameters":{}}`,
			`{"records":null}`,
			`no error`,
		},
		{`UpdateAccounts`,
			`UPDATE employeesa SET sales_count = sales_count + 1 WHERE employeesa.employee_id = (SELECT sales_person FROM accounts WHERE name = 'Robot Maker Corporation');`,
			`{"parameters":{}}`,
			`{"records":null}`,
			`no error`,
		},
		{`UpdateAccountsParameters`,
			`UPDATE employeesa SET sales_count = sales_count + ?mycount WHERE employeesa.employee_id = (SELECT sales_person FROM accounts WHERE name = ?corpname);`,
			`{"parameters":{ "corpname": "Acme Corporation", "mycount" : 2}}`,
			`{"records":null}`,
			`no error`,
		},
		{`UpdateAccountsParametersReturning`,
			`UPDATE employeesa SET sales_count = sales_count + ?mycount WHERE employeesa.employee_id = (SELECT sales_person FROM accounts WHERE name = ?corpname) RETURNING last_name;`,
			`{"parameters":{ "corpname": "Acme Corporation", "mycount" : 2}}`,
			`{"records":[{"last_name":"Greenberg"}]}`,
			`no error`,
		},
	}

	conn, err := GetPostgresObject()
	if err != nil {
		t.Errorf("connection failed: %s", err.Error())
		t.FailNow()
		return
	}
	for _, test := range tests {
		t.Logf("Running test %s\n", test.id)
		testQuery(t, test.id, test.query, test.input, test.expected, test.errorMsg, conn)
	}
}

func TestBlobs(t *testing.T) {
	//log.SetLogLevel(logger.DebugLevel)

	var tests = []struct {
		id        string
		query     string
		input     string
		fields    string
		expected  string
		errorMsg  string
		fieldName string
		fileName  string
	}{
		{id: `UpdatePngBlob`,
			query:     `UPDATE flogo.connector SET logo = ?logo;`,
			input:     `{"parameters":{}}`,
			fields:    `[{"FieldName":"logo","Type":"BYTEA","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"product_id","Type":"INT4","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"version","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"jiraid","Type":"TEXT","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"description","Type":"VARCHAR","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"NUMERIC","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
			expected:  `{"records":null}`,
			errorMsg:  `no error`,
			fieldName: "logo",
			fileName:  "../insert/icons/ic-postgres-insert@2x.png", // update to the middle size
		},
	}
	conn, err := GetPostgresObject()
	if err != nil {
		t.Errorf("connection failed: %s", err.Error())
		t.FailNow()
		return
	}
	for _, test := range tests {
		t.Logf("Running test %s\n", test.id)
		testBlob(t, test.id, test.query, test.input, test.fields, test.fieldName, test.fileName, test.expected, test.errorMsg, conn)
	}

}

// TODO merge all these tests later into one file
func testBlob(t *testing.T, id string, query string, input string, fields string,
	fieldName string, fileName string, expected string, errorMsg string, conn interface{}) {

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	connector := conn.(map[string]interface{})["connector"]
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, id)
	tc.SetInput(FieldsProperty, fields)

	inputParams := postgres.Input{}
	err := json.Unmarshal([]byte(input), &inputParams)

	photoBytes, err := ioutil.ReadFile(fileName)
	photoString := base64.StdEncoding.EncodeToString(photoBytes)
	inputParams.Parameters[fieldName] = photoString

	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complex)

	_, err = act.Eval(tc)
	if err != nil {
		if err.Error() == errorMsg {
			return
		}
		t.Errorf("%s", err.Error())
		return
	}
	complexOutput := tc.GetOutput(OutputProperty)
	outputData := complexOutput.(*data.ComplexObject).Value
	dataBytes, err := json.Marshal(outputData)
	if err != nil {
		t.Errorf("invalid response format")
		return
	}
	value := string(dataBytes)
	if expected != value {
		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
		return
	}
}

// GetPostgresObject for testing
func GetPostgresObject() (connector interface{}, err error) {
	cb, err := ioutil.ReadFile("data/connectionData.json")
	if err != nil {
		return connector, err
	}
	err = json.Unmarshal(cb, &connector)
	if err != nil {
		return connector, err
	}
	return
}

func testQuery(t *testing.T, id string, query string, input string, expected string, errorMsg string, conn interface{}) {

	act := NewActivity(getActivityMetadata())
	tc := test.NewTestActivityContext(act.Metadata())

	connector := conn.(map[string]interface{})["connector"]
	tc.SetInput(ConnectionProp, connector)
	tc.SetInput(QueryProperty, query)
	tc.SetInput(QueryNameProperty, id)

	var inputParams interface{}
	err := json.Unmarshal([]byte(input), &inputParams)
	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
	tc.SetInput("input", complex)

	_, err = act.Eval(tc)
	if err != nil {
		if err.Error() == errorMsg {
			return
		}
		t.Errorf("%s", err.Error())
		return
	}
	complexOutput := tc.GetOutput(OutputProperty)
	outputData := complexOutput.(*data.ComplexObject).Value
	dataBytes, err := json.Marshal(outputData)
	if err != nil {
		t.Errorf("invalid response format")
		return
	}
	value := string(dataBytes)
	if expected != value {
		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
		return
	}
}
***/