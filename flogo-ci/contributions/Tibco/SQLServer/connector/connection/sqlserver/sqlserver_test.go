/*
 * Copyright Â© 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package sqlserver

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"testing"

// 	"git.tibco.com/git/product/ipaas/wi-contrib.git/connection/generic"
// 	"git.tibco.com/git/product/ipaas/wi-mssql.git/src/app/SQLServer/connector/connection/sqlserver"

// 	"git.tibco.com/git/product/ipaas/wi-mssql.git/src/app/SQLServer/activity/insert"
// 	"git.tibco.com/git/product/ipaas/wi-mssql.git/src/app/SQLServer/activity/query"
// 	"git.tibco.com/git/product/ipaas/wi-mssql.git/src/app/SQLServer/activity/update"
// 	"github.com/TIBCOSoftware/flogo-contrib/action/flow/test"
// 	"github.com/TIBCOSoftware/flogo-lib/core/activity"
// 	"github.com/TIBCOSoftware/flogo-lib/core/data"
// 	"github.com/TIBCOSoftware/flogo-lib/logger"
// 	"github.com/stretchr/testify/assert"
// )

// var log = logger.GetLogger("flogo.sqlserver_test")
// var insertlogger = logger.GetLogger("sqlserver-insert")
// var deletelogger = logger.GetLogger("sqlserver-delete")
// var updatelogger = logger.GetLogger("sqlserver-update")
// var querylogger = logger.GetLogger("sqlserver-query")

// var connector *sqlserver.Connection

// var connectionJSON = `{
// 	"id" : "SqlServerTestConnection",
// 	"name": "tibco-sqlserver",
// 	"description" : "SqlServer Test Connection",
// 	"title": "SqlServer Connector",
// 	"type": "flogo:connector",
// 	"version": "1.0.0",
// 	"ref": "https://git.tibco.com/git/product/ipaas/wi-mssql.git/activity/query",
// 	"keyfield": "name",
// 	"settings": [
// 		{
// 		  "name": "name",
// 		  "value": "MyConnection",
// 		  "type": "string"
// 		},
// 		{
// 		  "name": "description",
// 		  "value": "SqlServer Connection",
// 		  "type": "string"
// 		},
// 		{
// 		  "name": "host",
// 		  "value": "flex-linux-gazelle",
// 		  "type": "string"
// 		},
// 		{
// 		  "name": "port",
// 		  "value": 1433,
// 		  "type": "int"
// 		},
// 		{
// 		  "name": "databaseName",
// 		  "value": "NORTHWND",
// 		  "type": "string"

// 		},
// 		{
// 		  "name": "user",
// 		  "value": "widev",
// 		  "type": "string"

// 		},
// 		{
// 		  "name": "password",
// 		  "value": "widev",
// 		  "type": "string"

// 		}
// 	  ]
// }`

// var invalidConnectionJSON = []byte(`{
//    "id" : "MySQLTestConnection",
//    "name": "tibco-sqlserver",
//    "description" : "SqlServer Test Connection",
//    "title": "SqlServer Connector",
//    "type": "flogo:connector",
//    "version": "1.0.0",
//    "ref": "https://git.tibco.com/git/product/ipaas/wi-mssql.git/activity/query",
//    "keyfield": "name",
//    "settings": [
// 	   {
// 		   "name": "name",
// 		   "value": "MyConnection",
// 		   "type": "string"
// 	   },
// 	   {
// 		   "name": "description",
// 		   "value": "SqlServer Connection",
// 		   "type": "string"
// 	   },
// 	   {
// 		   "name": "host",
// 		   "value": "flex-linux-gazelle",
// 		   "type": "string"
// 	   },
// 	   {
// 		   "name": "port",
// 		   "value": 1433,
// 		   "type": "int"
// 	   },
// 	   {
// 		   "name": "databaseName",
// 		   "value": "NORTHWND",
// 		   "type": "string"

// 	   },
// 	   {
// 		   "name": "user",
// 		   "value": "dead",
// 		   "type": "string"

// 	   },
// 	   {
// 		   "name": "password",
// 		   "value": "dead",
// 		   "type": "string"

// 	   }
// 	   ]
// }`)

// func getConnector(t *testing.T) (connector map[string]interface{}, err error) {

// 	connector = make(map[string]interface{})
// 	err = json.Unmarshal([]byte(connectionJSON), &connector)
// 	if err != nil {
// 		t.Errorf("Error: %s", err.Error())
// 	}
// 	return
// }

// func setSchema(conn map[string]interface{}, schema string) {
// 	settings := conn["settings"].([]interface{})
// 	for _, setting := range settings {
// 		settingName := setting.(map[string]interface{})["name"]
// 		if settingName == "databaseName" {
// 			setting.(map[string]interface{})["value"] = schema
// 			return
// 		}
// 	}

// }

// func getConnection(t *testing.T) (connection *sqlserver.Connection, err error) {
// 	log.SetLogLevel(logger.DebugLevel)
// 	connector, err := getConnector(t)
// 	assert.NotNil(t, connector)
// 	connectionObj, err := generic.NewConnection(connector)
// 	if err != nil {
// 		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
// 		t.Fail()
// 	}

// 	connection, err = sqlserver.GetConnection(connectionObj, log)
// 	if err != nil {
// 		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
// 		t.Fail()
// 	}
// 	assert.NotNil(t, connection)
// 	return
// }

// var mdatas map[string]*activity.Metadata

// func init() {
// 	mdatas = make(map[string]*activity.Metadata)
// }

// // GetActivityMetadata is used only with tests
// func GetActivityMetadata(activityType string) *activity.Metadata {
// 	activityMetadata, ok := mdatas[activityType]
// 	if !ok {
// 		file := "../../../activity/" + activityType + "/activity.json"
// 		jsonMetadataBytes, err := ioutil.ReadFile(file)
// 		if err != nil {
// 			fmt.Printf("failed to load json metadata file: %s", file)
// 			panic("failed to load activity metadata")
// 		}
// 		activityMetadata = activity.NewMetadata(string(jsonMetadataBytes))
// 		mdatas[activityType] = activityMetadata
// 	}
// 	return activityMetadata
// }

// func TestQueries(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)

// 	var tests = []struct {
// 		id       string
// 		query    string
// 		input    string
// 		expected string
// 		errorMsg string
// 	}{
// 		// {`MultipleSameParams`,
// 		// 	`select count(*) from Employees WHERE FirstName=?abc and LastName=?abcxyx and TitleOfCourtesy=?abc;`,
// 		// 	`{"parameters":{"abcxyx": "jobs", "abc": "steve"}}`,
// 		// 	`{"records":[{"":0}]}`,
// 		// 	``,
// 		// },
// 		// {`SubstringParams`,
// 		// 	`select count(*) from Employees WHERE FirstName=?abc and LastName=?abcxyx and TitleOfCourtesy=?abc;`,
// 		// 	`{"parameters":{"abcxyx": "jobs", "abc": "steve"}}`,
// 		// 	`{"records":[{"":0}]}`,
// 		// 	``,
// 		// },
// 		// {`WIMYSQ-144`,
// 		// 	`select count(*) from tempdb.dbo.practice_wcn WHERE practice_name=?z_emr_practice_name and emr_practice_name=?emr_practice_name;`,
// 		// 	`{"parameters":{"emr_practice_name": "1959031", "z_emr_practice_name": 105701}}`,
// 		// 	`{"records":[{"":0}]}`,
// 		// 	``,
// 		// },
// 		// {`WIMYSQ-144`,
// 		// 	`select count(*) from tempdb.dbo.practice_wcn WHERE practice_name='105701' and emr_practice_name='1959031';`,
// 		// 	`{"parameters":{}}`,
// 		// 	`{"records":[{"":0}]}`,
// 		// 	``,
// 		// },
// 	}

// 	conn, err := getConnector(t)
// 	if err != nil {
// 		t.Errorf("connection failed: %s", err.Error())
// 		t.FailNow()
// 		return
// 	}
// 	for _, test := range tests {
// 		t.Logf("Running test %s\n", test.id)
// 		testQuery(t, test.id, test.query, test.input, test.expected, test.errorMsg, conn)
// 	}
// }

// func testQuery(t *testing.T, id string, queryString string, input string, expected string, errorMsg string, conn interface{}) {
// 	querylogger.SetLogLevel(logger.DebugLevel)

// 	act := query.NewActivity(GetActivityMetadata("query"))
// 	tc := test.NewTestActivityContext(act.Metadata())

// 	// connector := conn.(map[string]interface{})["connector"]
// 	tc.SetInput("Connection", conn)
// 	tc.SetInput("Query", queryString)
// 	tc.SetInput("QueryName", id)

// 	var inputParams interface{}
// 	err := json.Unmarshal([]byte(input), &inputParams)
// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	if err != nil {
// 		if err.Error() == errorMsg {
// 			return
// 		}
// 		t.Errorf("%s", err.Error())
// 		return
// 	}
// 	complexOutput := tc.GetOutput("Output")
// 	outputData := complexOutput.(*data.ComplexObject).Value
// 	dataBytes, err := json.Marshal(outputData)
// 	if err != nil {
// 		t.Errorf("invalid response format")
// 		return
// 	}
// 	value := string(dataBytes)
// 	if expected != value {
// 		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
// 		return
// 	}
// }

// func TestInserts(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)
// 	insertlogger.SetLogLevel(logger.DebugLevel)
// 	updatelogger.SetLogLevel(logger.DebugLevel)
// 	querylogger.SetLogLevel(logger.DebugLevel)
// 	deletelogger.SetLogLevel(logger.DebugLevel)

// 	var tests = []struct {
// 		id           string
// 		testType     string
// 		schema       string
// 		query        string
// 		input        string
// 		fields       string
// 		expected     string
// 		errorMsg     string
// 		outputSchema string
// 		inputSchema  string
// 	}{
// 		// {`InsertFromSelectNoParms`, `insert`, `NORTHWND`,
// 		// 	`INSERT INTO Customers_wcn (CustomerID , CompanyName, ContactName,ContactTitle,PostalCode) SELECT CustomerID,CompanyName, ContactName,ContactTitle,PostalCode FROM Customers where City=?city ORDER BY CustomerID;`,
// 		// 	`{"parameters":{"city":"Paris"},"values":[]}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":0}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`WIMYSQ-161-Update`, `update`, `university`,
// 		// 	`UPDATE persons SET FirstName = ?name, LastName= ?na, City= ?na3 where ID= 1;`,
// 		// 	`{"parameters":{"name":"Elon The Great", "na":"Very Musk","na3":"New Palo Alto"}}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"rowsAffected":1}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`WIMYSQ-167`, `insert`, `university`,
// 		// 	`INSERT INTO shirts (name, size) VALUES ('dress shirt','large'), ('t-shirt','medium');`,
// 		// 	`{"parameters":{},
// 		//       "values":[]}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":2}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`WIDBSV-161`, `insert`, `university`,
// 		// 	`insert into Persons values (?ID, ?name, ?na, ?age, ?na3);;`,
// 		// 	`{"parameters":{"name":"Elon", "na":"Musk", "age":48, "na3":"Palo Alto"},
// 		//       "values":[{"ID":1}]}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":1}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`InsertRecords`, `insert`, `university`,
// 		// 	`INSERT INTO products (product_no, name, price) VALUES (1, 'Cheese', ?Price),(2, 'Bread', ?Price), (3, 'Milk', ?Price);`,
// 		// 	`{"values":[{"Price":2.99}, {"Price":3.99}, {"Price":4.99}]}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":3}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`InsertRecords`, `insert`, `university`,
// 		// 	`INSERT INTO employee_wcn (emp_id, emp_name, performance, salary) VALUES (?emp_id, ?emp_name, ?performance, ?salary);`,
// 		// 	`{"values":[{"emp_id":1, "emp_name": "walter", "performance": 22, "salary":50000},
// 		// 		{"emp_id":2, "emp_name": "jane of the seven seas crossed tith hane of ark", "performance": 23, "salary":60000},
// 		// 		{"emp_id":78, "emp_name": "jane of the seven seas crossed tith hane of ark", "performance": 23, "salary":60000},
// 		// 		{"emp_id":79, "emp_name": "jane of the seven seas crossed tith hane of ark", "performance": 23, "salary":60000},
// 		// 		{"emp_id":80, "emp_name": "mary", "performance": 24, "salary":70000}]}`,
// 		// 	`[{"FieldName":"emp_id","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},
// 		// 		{"FieldName":"emp_name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},
// 		// 		{"FieldName":"performance","Type":"NUMERIC","Selected":false,"Parameter":false,"isEditable":false,"Value":true},
// 		// 		{"FieldName":"salary","Type":"NUMERIC","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":3}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`InsertOneValueMultipleRecords`, `insert`, `university`,
// 		// 	`INSERT INTO products (product_no, name, price) VALUES (?product_no, ?name, ?price);`,
// 		// 	`{"values":[{"price":2.55, "product_no":6, "name":"Corn"}, {"price":3.55, "product_no":7, "name":"Turmeric"}, {"price":6.55, "product_no":8, "name":"Basil"}]}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":3}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`InsertWithParameters`, `insert`, `university`,
// 		// 	`INSERT INTO products (product_no, name, price) VALUES (?myproduct, ?myname, ?myprice);`,
// 		// 	`{"parameters":{"myproduct":9, "myname":"Salt", "myprice":0.99}}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":1}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		// {`InsertWithParametersAndValues`, `insert`, `university`, // This will not process the parameters, just the values cause we're not supporting mixing them.  Till someone changes my mind.  wcn
// 		// 	`INSERT INTO products (product_no, name, price) VALUES (?product_no, ?name, ?price), (?myproduct, ?myname, ?myprice);`,
// 		// 	`{"parameters":{"myproduct":14, "myname":"Pepper", "myprice":1.99},
// 		//       "values":[{"price":8.51, "product_no":11, "name":"Flour"}, {"price":9.99, "product_no":12, "name":"Maple Syrup"}]}`,
// 		// 	`[{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false},{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true}]`,
// 		// 	`{"lastInsertId":0,"rowsAffected":2}`,
// 		// 	`no error`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"records":{"type":"array","items":{"type":"object","properties":{}}}}}`,
// 		// 	`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		// },
// 		{`WIMYSQ-157`, `insert`, `university`,
// 			`INSERT INTO testgroupconcat (empid, fname, lname, deptid, strength) VALUES ( ?a, ?b, CONCAT(?b,' dup'), 7, 'anything');`,
// 			`{"parameters":{"a" : 1, "b": "John"}}`,
// 			`[{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"Image","Type":"BLOB","Selected":false,"Parameter":true,"isEditable":false,"Value":true}]`,
// 			`{"lastInsertId":0,"rowsAffected":1}`,
// 			`no error`,
// 			`"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"ProductNo":{"type":"integer"},"Name":{"type":"string"},"Price":{"type":"number"},"Image":{"type":"string"}}}},"parameters":{"type":"object","properties":{}}}}`,
// 			`{"$schema":"http://json-schema.org/draft-04/schema#","type":"object","definitions":{},"properties":{"values":{"type":"array","items":{"type":"object","properties":{"Name":{"type":"string"},"Price":{"type":"number"},"ProductNo":{"type":"integer"}}}},"parameters":{"type":"object","properties":{"price":{"type":"number"}}}}}`,
// 		},
// 	}

// 	conn, err := getConnector(t)
// 	if err != nil {
// 		t.Errorf("connection failed: %s", err.Error())
// 		t.FailNow()
// 		return
// 	}

// 	for _, test := range tests {
// 		setSchema(conn, test.schema)
// 		t.Logf("Running test %s\n", test.id)
// 		switch test.testType {
// 		case "insert":
// 			testInsert(t, test.id, test.query, test.input, test.fields, test.expected, test.errorMsg, conn)
// 		case "update":
// 			testUpdate(t, test.id, test.query, test.input, test.fields, test.expected, test.errorMsg, conn)
// 		case "query":
// 			testQuery(t, test.id, test.query, test.input, test.expected, test.errorMsg, conn)
// 		default:
// 			t.Errorf("unrecognized test type: %s", test.testType)

// 		}
// 	}
// }

// func testInsert(t *testing.T, id string, queryString string, input string, fields string, expected string, errorMsg string, conn interface{}) {

// 	act := insert.NewActivity(GetActivityMetadata("insert"))
// 	tc := test.NewTestActivityContext(act.Metadata())

// 	tc.SetInput("Connection", conn)
// 	tc.SetInput("InsertStatement", queryString)
// 	tc.SetInput("QueryName", id)
// 	tc.SetInput("Fields", fields)

// 	var inputParams interface{}
// 	err := json.Unmarshal([]byte(input), &inputParams)
// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	if err != nil {
// 		if err.Error() == errorMsg {
// 			return
// 		}
// 		t.Errorf("%s", err.Error())
// 		return
// 	}
// 	complexOutput := tc.GetOutput("Output")
// 	outputData := complexOutput.(*data.ComplexObject).Value
// 	dataBytes, err := json.Marshal(outputData)
// 	if err != nil {
// 		t.Errorf("invalid response format")
// 		return
// 	}
// 	value := string(dataBytes)
// 	if expected != value {
// 		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
// 		return
// 	}
// }

// func testUpdate(t *testing.T, id string, queryString string, input string, fields string, expected string, errorMsg string, conn interface{}) {

// 	act := update.NewActivity(GetActivityMetadata("update"))
// 	tc := test.NewTestActivityContext(act.Metadata())

// 	tc.SetInput("Connection", conn)
// 	tc.SetInput("UpdateStatement", queryString)
// 	tc.SetInput("QueryName", id)
// 	tc.SetInput("Fields", fields)

// 	var inputParams interface{}
// 	err := json.Unmarshal([]byte(input), &inputParams)
// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	if err != nil {
// 		if err.Error() == errorMsg {
// 			return
// 		}
// 		t.Errorf("%s", err.Error())
// 		return
// 	}
// 	complexOutput := tc.GetOutput("Output")
// 	outputData := complexOutput.(*data.ComplexObject).Value
// 	dataBytes, err := json.Marshal(outputData)
// 	if err != nil {
// 		t.Errorf("invalid response format")
// 		return
// 	}
// 	value := string(dataBytes)
// 	if expected != value {
// 		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
// 		return
// 	}
// }

// func TestBlobs(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)

// 	var tests = []struct {
// 		id        string
// 		testType  string
// 		query     string
// 		input     string
// 		fields    string
// 		expected  string
// 		errorMsg  string
// 		imageFile string
// 	}{

// 		{`InsertBlob`, `insert`,
// 			`INSERT INTO university.products VALUES (4, 'Jam', ?Price, ?Image), (5, 'Raisins', ?Price, ?Image);`,
// 			`{
// 		        "values": [{
// 		          "Price": 8.99,
// 		          "Image": "iVBORw0KGgoAAAANSUhEUgAAABEAAAAUCAYAAABroNZJAAAABHNCSVQICAgIfAhkiAAAA31JREFU OI11lN1PW2Ucxz/P6enLgXaFQmkpo7x0c4AOnBvzJUucElEkcYlmISYmJl544aV/gomXXuzKOOMS NUuGi25xmsjMXDZ1jpCJBNhb1rp1oy8WRlt6OD3tefGiwGDK9+6XPN/v8/19n9/zE+Vy2WYTTNPk xtwMyXicaCyG4lH47NgnBMMtDL9+hIEDBwGxmYK8ubhy5Vduz0wzPfUn0a4YdxNxhkZGkUWWcMDk 8sWLFFWVFw69iMPh2OBJYAMWYPPNiRPgduL1FtjXN4FamOfq1CSyVMY2U6TuP+CjD95hpZD/fyeq WsIwBK89N4W5VKRSWWVpaZHVyl/ISgCXa4XRUR/ZTISHD5fx+Rs23EggSC8scObUSR7c/5uV/G2q hgZAV5fG20edGzcK9RJ2JcfZYx9z6ecfMUyjJmIaBlcvX0BXC/QEg1tsGpU82vLkRm07JGQlQN/N aSbHvyJ5L1kTqRpVkGQGmpppmE9tEbHtLbmDqL1KZ52XXaZFJpOpZSIQpNNJzk1MYIUMPv3Swl9f pbNT4JCKQBMADqfN6e801IzBcQoULXjfNmsistNJsLGZZ19+hSP9T3Nq/CRTM7P07KnHtBRAUDVM /slp5G4YfPjueyh+P+fm5gmGIzURSZJoibSjrmro588ypBfAcvPHTwYFZ4XFRQeJWxk8CY0xvwf/ b+fJV6qoje0EW4Jr7QhBoLmFeCKO7lJwl4oM+xTUqk2p6sK6Ixhx2LhND3WKgNUSJcOmaddu6r2+ R3MSbmtDNyzkvr2Ys5Os4uCaJvGLruOSdTy2xLCrjqfqnDiBmUKe6O4exFrQMoDP52clk0SN7MdZ SPKtCBJ78y2OB+9RnzpNIu9gfK4VdyVAr0siR5XDXR2bxx48ikJzKExKdqM3dFOKhBk7EMGXnUBY Bt07dA7uXKYiGdhACplo9DERgI7Yk6Rzy3hbO2hOL/LD1RT6jv0gJApVmdmMF7cl0GwQoW4aGgP/ Fel6Ikb8zk1Ch19iRHHy+9dfMPR5gqHv2zl6Jkib5qfX7WI6m2XP4CEkaYO69gFtiER2snDrOtmB PoziXcZao4ytn/ICFmTKGtdLJV7d118j1cYYeb3weOp4ZnCQC+kc9vNvsB2crb307x1g82IS5bJm r7vJ5/PoenlbgXW0hEKI9SQEiNp6tNcMicc33/Z41A3/AmQaYxgRTeWaAAAAAElFTkSuQmCC"
// 		        }, {
// 		          "Price": 3.99,
// 		          "Image": "VBORw0KGgoAAAANSUhEUgAAABQAAAAVCAYAAABG1c6oAAAABHNCSVQICAgIfAhkiAAAAwJJREFU OI2VlU1MFFccwH9vdmdnl53ZZTXlQzcUrK5YMQ0EYmtbUgGpTbSIGg4aU9NED+2hh7YHLx5MjDEm TXowtgdPbfoRY0pTPDRpS5qepJpg+TQrqCsE1pUNRZldZuGNhwFlnQXiP5lkZt683/u9//u/NyKb zdrkhY1tgxAAghfDtp3PhXC3AXgLvbRtiWVZxIeGGL09gGVlCL0SZfuONygtK8WjKAUHAxD5hs7t VCrFNxe/Rg8WURGNYs78j2d2luGxQXbtaWd3y35U1QMoaxuapslnHx2m7ejHtNaUot08i3zygOsT YYqzD+k8/wdFviDvNLeuZmgDNlLCV2dOU1JeyrGWGmTncWQuCQrMWB7mpeDhY8n561EuXbuBYYRc wGXOguTkBHf6B+joaGPh91PYuSRCcbL11/0wqVkVw++lqTzBlR+/L2i4DGhzbyROQ3MTvkQPpHvz Wlsrp9kaybBRt2jcrPBP958FgXk5nMtkKI6sQ4z3YoslbyeCqmRp9TboktRUei1D0A2DR8k0MliG wF0YS8+TppeqqkqWcr8CUBB7vYa+3n9Jq+vBW7SYiPwuuQXBL33zHGhrX8vQJhQuZvfeDzj3w6+Y G5pAujv8nRAkSlp4+72WF+XchmCzr62dV6u2891NQI8h5POpxtMBehYa+fTzC/gDftd0lwHFs0vT /Jz85AuSeowbVi1CdWotk4OuRIB3j3zJllg1Tj25t6B77wiJ3+/nYPshusd93M+WkDJVbk36mCv7 kIa6eqQskIsVgSgIIXgttoVIRTWXhzSuPS7n57EG9hw6ghEK41E8LwN0QtcN6na+Rfk2g9r3w0hj E7UNbz7P0MsChRCoWgBDF4RDPoqCGj6fb2XSykBn5bLZDAO9PeTmvahehZH4MMN9/y0esAXqZS3g 3ZFREvGf2N9qEi0TnDg+zdXL32JZc6sCC5zYzhiBQACZ0en6rZ+sqYG2QEnFRhTFw2pJFO5/imOZ y+UYvT3M2L1BVC0EzLGjvpFIZP2iYWHoU47sBvOnzdmGAAAAAElFTkSuQmCC"
// 		        }]
// 		      }`,
// 			`[{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"Image","Type":"BLOB","Selected":false,"Parameter":true,"isEditable":false,"Value":true}]`,
// 			`{"lastInsertId":0,"rowsAffected":2}`,
// 			`no error`,
// 			``,
// 		},
// 		{`QueryBlob`, `query`,
// 			`select Name, Image from products where ProductNo=?ProductNo;`,
// 			`{"parameters":{"ProductNo" : 4}}`,
// 			`[{"FieldName":"ProductNo","Type":"INT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Name","Type":"TEXT","Selected":false,"Parameter":false,"isEditable":false,"Value":true},{"FieldName":"Price","Type":"DECIMAL","Selected":false,"Parameter":true,"isEditable":false,"Value":true},{"FieldName":"Image","Type":"BLOB","Selected":false,"Parameter":true,"isEditable":false,"Value":true}]`,
// 			`{"records":[{"Image":"aVZCT1J3MEtHZ29BQUFBTlNVaEVVZ0FBQUJFQUFBQVVDQVlBQUFCcm9OWkpBQUFBQkhOQ1NWUUlDQWdJZkFoa2lBQUFBMzFKUkVGVSBPSTExbE4xUFcyVWN4ei9QNmVuTGdYYUZRbWtwbzd4MGM0QU9uQnZ6SlV1Y0VsRWtjWWxtSVNZbUpsNTQ0YVYvZ29tWFh1ektPT01TIE5VdUdpMjV4bXNqTVhEWjFqcENKQk5oYjFycDFveThXUmx0Nk9EM3RlZkdpd0dESzkrNlhQTi92OC8xOW45L3pFK1Z5MldZVFROUGsgeHR3TXlYaWNhQ3lHNGxINDdOZ25CTU10REw5K2hJRURCd0d4bVlLOHViaHk1VmR1ejB3elBmVW4wYTRZZHhOeGhrWkdrVVdXY01EayA4c1dMRkZXVkZ3NjlpTVBoMk9CSllBTVdZUFBOaVJQZ2R1TDFGdGpYTjRGYW1PZnExQ1N5Vk1ZMlU2VHVQK0NqRDk1aHBaRC9meWVxIFdzSXdCSzg5TjRXNVZLUlNXV1ZwYVpIVnlsL0lTZ0NYYTRYUlVSL1pUSVNIRDVmeCtSczIzRWdnU0M4c2NPYlVTUjdjLzV1Vi9HMnEgaGdaQVY1ZkcyMGVkR3pjSzlSSjJKY2ZaWXg5ejZlY2ZNVXlqSm1JYUJsY3ZYMEJYQy9RRWcxdHNHcFU4MnZMa1JtMDdKR1FsUU4vTiBhU2JIdnlKNUwxa1RxUnBWa0dRR21wcHBtRTl0RWJIdExibURxTDFLWjUyWFhhWkZKcE9wWlNJUXBOTkp6azFNWUlVTVB2M1N3bDlmIHBiTlQ0SkNLUUJNQURxZk42ZTgwMUl6QmNRb1VMWGpmTm1zaXN0TkpzTEdaWjE5K2hTUDlUM05xL0NSVE03UDA3S25IdEJSQVVEVk0gL3NscDVHNFlmUGp1ZXloK1ArZm01Z21HSXpVUlNaSm9pYlNqcm1ybzU4OHlwQmZBY3ZQSFR3WUZaNFhGUlFlSld4azhDWTB4dndmLyBiK2ZKVjZxb2plMEVXNEpyN1FoQm9MbUZlQ0tPN2xKd2w0b00reFRVcWsycDZzSzZJeGh4MkxoTkQzV0tnTlVTSmNPbWFkZHU2cjIrIFIzTVNibXRETnl6a3ZyMllzNU9zNHVDYUp2R0xydU9TZFR5MnhMQ3JqcWZxbkRpQm1VS2U2TzRleEZyUU1vRFA1MmNsazBTTjdNZFogU1BLdENCSjc4eTJPQis5Um56cE5JdTlnZks0VmR5VkFyMHNpUjVYRFhSMmJ4eDQ4aWtKektFeEtkcU0zZEZPS2hCazdFTUdYblVCWSBCdDA3ZEE3dVhLWWlHZGhBQ3BsbzlERVJnSTdZazZSenkzaGJPMmhPTC9MRDFSVDZqdjBnSkFwVm1kbU1GN2NsMEd3UW9XNGFHZ1AvIEZlbDZJa2I4emsxQ2gxOWlSSEh5KzlkZk1QUjVncUh2MnpsNkpraWI1cWZYN1dJNm0yWFA0Q0VrYVlPNjlnRnRpRVIyc25Eck90bUIgUG96aVhjWmFvNHl0bi9JQ0ZtVEtHdGRMSlY3ZDExOGoxY1lZZWIzd2VPcDRabkNRQytrYzl2TnZzQjJjcmIzMDd4MWc4MklTNWJKbSByN3ZKNS9Qb2VubGJnWFcwaEVLSTlTUUVpTnA2dE5jTWljYzMzL1o0MUEzL0FtUWFZeGdSVGVXYUFBQUFBRWxGVGtTdVFtQ0M=","Name":"Jam"}]}`,
// 			``,
// 			``,
// 		},
// 	}

// 	conn, err := getConnector(t)
// 	if err != nil {
// 		t.Errorf("connection failed: %s", err.Error())
// 		t.FailNow()
// 		return
// 	}

// 	for _, test := range tests {
// 		t.Logf("Running test %s\n", test.id)
// 		switch test.testType {
// 		case "insert":
// 			testBlobInsert(t, test.id, test.query, test.input, test.fields, test.expected, test.errorMsg, test.imageFile, conn)
// 		case "query":
// 			testBlobQuery(t, test.id, test.query, test.input, test.fields, test.expected, test.errorMsg, conn)
// 		default:
// 			t.Errorf("unrecognized test type: %s", test.testType)

// 		}
// 	}
// }

// func testBlobInsert(t *testing.T, id string, queryString string, input string, fields string,
// 	expected string, errorMsg string, imageFile string, conn interface{}) {

// 	act := insert.NewActivity(GetActivityMetadata("insert"))
// 	tc := test.NewTestActivityContext(act.Metadata())

// 	tc.SetInput("Connection", conn)
// 	tc.SetInput("Query", queryString)
// 	tc.SetInput("QueryName", id)
// 	tc.SetInput("Fields", fields)

// 	var inputParams interface{}
// 	err := json.Unmarshal([]byte(input), &inputParams)
// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	if err != nil {
// 		if err.Error() == errorMsg {
// 			return
// 		}
// 		t.Errorf("%s", err.Error())
// 		return
// 	}
// 	complexOutput := tc.GetOutput("Output")
// 	outputData := complexOutput.(*data.ComplexObject).Value
// 	dataBytes, err := json.Marshal(outputData)
// 	if err != nil {
// 		t.Errorf("invalid response format")
// 		return
// 	}
// 	value := string(dataBytes)
// 	if expected != value {
// 		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
// 		return
// 	}
// }

// func testBlobQuery(t *testing.T, id string, queryString string, input string, fields string, expected string, errorMsg string, conn interface{}) {

// 	act := query.NewActivity(GetActivityMetadata("query"))
// 	tc := test.NewTestActivityContext(act.Metadata())

// 	tc.SetInput("Connection", conn)
// 	tc.SetInput("Query", queryString)
// 	tc.SetInput("QueryName", id)
// 	tc.SetInput("Fields", fields)

// 	var inputParams interface{}
// 	err := json.Unmarshal([]byte(input), &inputParams)
// 	complex := &data.ComplexObject{Metadata: "", Value: inputParams}
// 	tc.SetInput("input", complex)

// 	_, err = act.Eval(tc)
// 	if err != nil {
// 		if err.Error() == errorMsg {
// 			return
// 		}
// 		t.Errorf("%s", err.Error())
// 		return
// 	}
// 	complexOutput := tc.GetOutput("Output")
// 	outputData := complexOutput.(*data.ComplexObject).Value
// 	dataBytes, err := json.Marshal(outputData)
// 	if err != nil {
// 		t.Errorf("invalid response format")
// 		return
// 	}
// 	value := string(dataBytes)
// 	if expected != value {
// 		t.Errorf("query response has wrong value, got:  %s -- expected: %s", value, expected)
// 		return
// 	}
// }

// func TestGetConnection(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)
// 	connector := &sqlserver.Connector{}
// 	err := json.Unmarshal([]byte(connectionJSON), connector)
// 	if err != nil {
// 		t.Errorf("Error: %s", err.Error())
// 	}
// 	connectionObj, err := generic.NewConnection(connector)
// 	if err != nil {
// 		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
// 		t.Fail()
// 	}
// 	connection, err := sqlserver.GetConnection(connectionObj, log)
// 	if err != nil {
// 		t.Errorf("SqlServer get connection failed %s", err.Error())
// 		t.Fail()
// 	}
// 	assert.NotNil(t, connection)
// 	_, err = connection.Login(log)
// 	if err != nil {
// 		t.Errorf("SqlServer Login failed %s", err.Error())
// 	}
// 	connection.Logout(log)
// }

// func TestInvalidGetConnection(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)

// 	connector := &sqlserver.Connector{}
// 	err := json.Unmarshal([]byte(invalidConnectionJSON), connector)
// 	if err != nil {
// 		t.Errorf("Error: %s", err.Error())
// 	}
// 	connectionObj, err := generic.NewConnection(connector)
// 	if err != nil {
// 		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
// 		t.Fail()
// 	}
// 	connection, err := sqlserver.GetConnection(connectionObj, log)
// 	if err != nil {
// 		t.Errorf("SqlServer get connection failed %s", err.Error())
// 		t.Fail()
// 	}

// 	assert.NotNil(t, connection)
// 	_, err = connection.Login(log)

// 	if err != nil {
// 		fmt.Printf("SqlServer Login failed %s as expected \n", err.Error())
// 	}
// 	assert.Error(t, err)
// 	connection.Logout(log)
// }
