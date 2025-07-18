package delete

/*
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"git.tibco.com/git/product/ipaas/wi-contrib.git/connection/generic"
	"git.tibco.com/git/product/ipaas/wi-mysql.git/src/app/MySQL/connector/connection/mysql"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/stretchr/testify/assert"
)

var activityMetadata *activity.Metadata
var connector *mysql.Connection

var connectionJSON = []byte(`{
	 "id" : "MySQLTestConnection",
	 "name": "tibco-mysql",
	 "description" : "MySQL Test Connection",
	 "title": "MySQL Connector",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "ref": "https://git.tibco.com/git/product/ipaas/wi-mysql.git/activity/query",
	 "keyfield": "name",
	 "settings": [
		 {
		   "name": "name",
		   "value": "MyConnection",
		   "type": "string"
		 },
		 {
		   "name": "description",
		   "value": "MySQL Connection",
		   "type": "string"
		 },
		 {
		   "name": "host",
		   "value": "flex-linux-gazelle.na.tibco.com",
		   "type": "string"
		 },
		 {
		   "name": "port",
		   "value": 3306,
		   "type": "int"
		 },
		 {
		   "name": "databaseName",
		   "value": "university",
		   "type": "string"

		 },
		 {
		   "name": "user",
		   "value": "widev",
		   "type": "string"

		 },
		 {
		   "name": "password",
		   "value": "widev",
		   "type": "string"

		 }
	   ]
 }`)

var invalidConnectionJSON = []byte(`{
	 "id" : "MySQLTestConnection",
	 "name": "tibco-MySQL",
	 "description" : "MySQL Test Connection",
	 "title": "AWS MySQL Connector",
	 "type": "flogo:connector",
	 "version": "1.0.0",
	 "ref": "https://git.tibco.com/git/product/ipaas/wi-MySQL.git/activity/query",
	 "keyfield": "name",
	 "settings": [
		 {
		   "name": "name",
		   "value": "MyConnection",
		   "type": "string"
		 },
		 {
		   "name": "description",
		   "value": "My MySQL Connection",
		   "type": "string"
		 },
		 {
		   "name": "host",
		   "value": "flex-linux-gazelle.na.tibco.com",
		   "type": "string"
		 },
		 {
		   "name": "port",
		   "value": 3306,
		   "type": "int"
		 },
		 {
		   "name": "databaseName",
		   "value": "university",
		   "type": "string"

		 },
		 {
		   "name": "user",
		   "value": "widev",
		   "type": "string"

		 },
		 {
		   "name": "password",
		   "value": "wrongpassword",
		   "type": "string"

		 }
	   ]
 }`)

var flexConnectionJSON = []byte(`{
  "id" : "MySQLTestConnection",
  "name": "tibco-mysql",
  "description" : "MySQL Test Connection",
  "title": "MySQL Connector",
  "type": "flogo:connector",
  "version": "1.0.0",
  "ref": "https://git.tibco.com/git/product/ipaas/wi-mysql.git/activity/query",
  "keyfield": "name",
  "settings": [
    {
      "name": "name",
      "value": "MyConnection",
      "type": "string"
    },
    {
      "name": "description",
      "value": "MySQL Connection",
      "type": "string"
    },
    {
      "name": "host",
      "value": "flex-linux-gazelle.na.tibco.com",
      "type": "string"
    },
    {
      "name": "port",
      "value": 3306,
      "type": "int"
    },
    {
      "name": "databaseName",
      "value": "northwind",
      "type": "string"

    },
    {
      "name": "user",
      "value": "widev",
      "type": "string"

    },
    {
      "name": "password",
      "value": "widev",
      "type": "string"

    }
    ]
}`)

var loggger = log.ChildLogger(log.RootLogger(), "flogo-mysql-delete-test")

func getActivityMetadata() *activity.Metadata {
	if activityMetadata == nil {
		jsonMetadataBytes, err := ioutil.ReadFile("activity.json")
		if err != nil {
			panic("No Json Metadata found for activity.json path")
		}
		var activityMap map[string]interface{}
		var activityMetadata *Input
		err = json.Unmarshal(jsonMetadataBytes, activityMap)
		activityMetadata.FromMap(activityMap)
	}
	return activityMetadata
}

func getConnector(t *testing.T) (connector map[string]interface{}, err error) {

	connector = make(map[string]interface{})
	err = json.Unmarshal([]byte(flexConnectionJSON), &connector)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	return
}

func getConnection(t *testing.T) (connection *mysql.Connection, err error) {
	connector, err := getConnector(t)
	assert.NotNil(t, connector)
	connectionObj, err := generic.NewConnection(connector)
	if err != nil {
		t.Errorf("Mygo debutgo debutSQL get connection failed %s", err.Error())
		t.Fail()
	}

	connection, err = mysql.GetConnection(connectionObj)
	if err != nil {
		t.Errorf("MySQL get connection failed %s", err.Error())
		t.Fail()
	}
	assert.NotNil(t, connection)
	return
}

func TestGetConnection(t *testing.T) {
	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	if err != nil {
		t.Errorf("MySQL Login failed %s", err.Error())
	}
	conn.Logout(logger)
}

// func TestActivityRegistration(t *testing.T) {
// 	act := NewActivity(getActivityMetadata())
// 	if act == nil {
// 		t.Error("Activity Not Registered")
// 		t.Fail()
// 		return
// 	}
// }

func TestLogin(t *testing.T) {

	// conn := mysql.Connection{
	// 	DatabaseURL: "",
	// 	Host:        "wasp-deva.na.tibco.com",
	// 	Port:        3306,
	// 	User:        "root",
	// 	Password:    "admin",
	// 	DbName:      "university",
	// }
	conn := mysql.Connection{
		DatabaseURL: "",
		Host:        "flex-linux-gazelle.na.tibco.com",
		Port:        "3306",
		User:        "widev",
		Password:    "widev",
		DbName:      "northwind",
	}
	status, err := conn.Login(logger)
	if status == false {
		fmt.Println(err)
		t.Fail()
	}
	conn.Logout(logger)
}

func TestDeleteString(t *testing.T) {
	var prepareInsert = `INSERT INTO wcntest VALUES ('1','mary','popins','109');`
	var deleteStatement = `DELETE FROM wcntest where id = '1';`
	var fields []interface{}
	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)

	inputParams := &mysql.Input{}

	result, err := conn.PreparedInsert(prepareInsert, inputParams, fields, logger)
	assert.Nil(t, err)
	insertJSON, err := json.Marshal(result)

	result, err = conn.PreparedDelete(deleteStatement, inputParams, logger)
	assert.Nil(t, err)
	deleteJSON, err := json.Marshal(result)
	assert.Nil(t, err)
	fmt.Printf("\nResult  for insert:  %s \nResult for delete: %s", string(insertJSON), string(deleteJSON))

	conn.Logout(logger)
	assert.Nil(t, err)
}

func TestDeletePerf(t *testing.T) {
	var prepareInsert = `INSERT INTO wcntest VALUES ('1','mary','popins','109');`
	var deleteStatement = `DELETE FROM wcntest where id = '1';`
	var fields []interface{}
	conn, err := getConnection(t)
	assert.Nil(t, err)
	_, err = conn.Login(logger)
	assert.Nil(t, err)

	inputParams := &mysql.Input{}
	start := time.Now()

	for i := 0; i < 1000; i++ {
		_, err := conn.PreparedInsert(prepareInsert, inputParams, fields, logger)
		assert.Nil(t, err)
		// insertJSON, err := json.Marshal(result)

		_, err = conn.PreparedDelete(deleteStatement, inputParams, logger)
		assert.Nil(t, err)
		// deleteJSON, err := json.Marshal(result)
		assert.Nil(t, err)
		//		fmt.Printf("\nResult  for insert:  %s \nResult for delete: %s", string(insertJSON), string(deleteJSON))

	}
	end := time.Now()
	logger.Infof("Test performance ran 1000 queries in: %s \n", end.Sub(start).String())

	conn.Logout(logger)
	assert.Nil(t, err)
}
*/
