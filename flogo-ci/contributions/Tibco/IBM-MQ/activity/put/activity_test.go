package put

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"

	_ "net/http/pprof"
)

var activityMetadata *activity.Metadata

const connFile = "test_connection.json"

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
func setupActivity(t *testing.T) (*PutActivity, *test.TestActivityContext) {
	act := &PutActivity{}
	tc := test.NewActivityContext(act.Metadata())
	return act, tc
}
func TestCreate(t *testing.T) {

	act, tc := setupActivity(t)

	if act == nil {
		t.Error("Activity Not Created")
		t.Fail()
		return
	}
	if tc == nil {
		t.Error("Context Not Created")
		t.Fail()
		return
	}
}

var MQMDData = `{
	"CorrelId":"aXdpc2hpd2VyZWFub3NjYXJteWVyc3dlaW5lcgo=",
	"CodedCharSetId":1208,
	"Format":"MQSTR ",
	"Priority": 5,
	"Expiry": 50,
	"ReplyToQ":"flogo.reply",
	"ReplyToQmgr":"centosmq9",
	"PutApplName": "wcnappl",
	"PutApplType": 6,
	"UserIdentifier":"wcnuser",
	"AccountingToken": "wcnaccoungtoke",
	"ApplIdentityData": "wcnapplIDData",
	"ApplOriginData": "OGDTA"
	}`

var PropData = `{
			"propertyone":"property one in the message",
			"propertytwi":"property two in the message",
			"propertythree":"3334"
			}`

func TestEval(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile(connFile)
	if err != nil {
		panic(connFile + " not found")
	}
	var connection interface{}
	err = json.Unmarshal(connectionBytes, &connection)
	if err != nil {
		fmt.Printf("%s", err)
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}
	cmap := connection.(map[string]interface{})
	cname := cmap["connectorName"]
	log.Debug("connection name is %s", cname)
	connector := cmap["Connection"]
	settings := connector.(map[string]interface{})["settings"]
	cmap["settings"] = settings

	act, tc := setupActivity(t)
	//setup attrs
	tc.SetInput("Connection", connection.(map[string]interface{})["Connection"])
	tc.SetInput("valueType", "String")
	tc.SetInput("MessageString", "hello world")
	tc.SetInput("queue", "flogo.request")
	tc.SetInput("GenCorrelationID", false)
	tc.SetInput("messageType", "Datagram")
	tc.SetInput("contextSupport", "Identity")

	mqmdData := make(map[string]interface{})
	err = json.Unmarshal([]byte(MQMDData), &mqmdData)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	complex := &data.ComplexObject{Metadata: "", Value: mqmdData}
	tc.SetInput("MQMD", complex)

	//setup props
	propData := make(map[string]interface{})
	err = json.Unmarshal([]byte(PropData), &propData)
	if err != nil {
		fmt.Printf("Error making props: %s", err)
		t.Errorf("Error: %s", err.Error())
	}
	propComplex := &data.ComplexObject{Metadata: "", Value: propData}
	tc.SetInput("properties", propComplex)

	act.Eval(tc)

	complexOutput := tc.GetOutput("Output")
	assert.NotNil(t, complexOutput)

	outputData := complexOutput.(*data.ComplexObject).Value
	dataBytes, err := json.Marshal(outputData)
	assert.Nil(t, err)
	assert.NotNil(t, dataBytes)
	fmt.Printf("\n%s\n", string(dataBytes))
	//check result attr
}

func TestEvalJsonBody(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile(connFile)
	if err != nil {
		panic(connFile + " not found")
	}
	var connection interface{}
	err = json.Unmarshal(connectionBytes, &connection)
	if err != nil {
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}
	cmap := connection.(map[string]interface{})
	connector := cmap["Connection"]
	settings := connector.(map[string]interface{})["settings"]
	cmap["settings"] = settings

	var messagejson = `{
		"FName":"Walter",
		"LName":"Matheou",
		"Age": 98,
		"EyeColr":"blue"
	}`

	var msgJSONMap map[string]interface{}
	err = json.Unmarshal([]byte(messagejson), &msgJSONMap)
	if err != nil {
		fmt.Print(err)
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}
	complexJSONObj := &data.ComplexObject{Metadata: "", Value: msgJSONMap}

	act, tc := setupActivity(t)
	//setup attrs
	tc.SetInput("Connection", connection.(map[string]interface{})["Connection"])
	tc.SetInput("valueType", "JSON")
	tc.SetInput("MessageJson", complexJSONObj)
	tc.SetInput("queue", "flogo.request")
	tc.SetInput("GenCorrelationID", false)
	tc.SetInput("messageType", "Datagram")

	mqmdData := make(map[string]interface{})
	err = json.Unmarshal([]byte(MQMDData), &mqmdData)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	complex := &data.ComplexObject{Metadata: "", Value: mqmdData}
	tc.SetInput("MQMD", complex)

	//setup props
	propData := make(map[string]interface{})
	err = json.Unmarshal([]byte(PropData), &propData)
	if err != nil {
		fmt.Printf("Error making props: %s", err)
		t.Errorf("Error: %s", err.Error())
	}
	propComplex := &data.ComplexObject{Metadata: "", Value: propData}
	tc.SetInput("properties", propComplex)

	act.Eval(tc)

	complexOutput := tc.GetOutput("Output")
	assert.NotNil(t, complexOutput)

	outputData := complexOutput.(*data.ComplexObject).Value
	dataBytes, err := json.Marshal(outputData)
	assert.Nil(t, err)
	assert.NotNil(t, dataBytes)
	fmt.Printf("\n%s\n", string(dataBytes))
	//check result attr
}

func TestEvalMulti(t *testing.T) {
	connectionBytes, err := ioutil.ReadFile(connFile)
	if err != nil {
		panic(connFile + " not found")
	}
	var connection interface{}
	err = json.Unmarshal(connectionBytes, &connection)
	if err != nil {
		t.Errorf("Deserialization of connection failed %s", err.Error())
		t.Fail()
	}

	cmap := connection.(map[string]interface{})
	cname := cmap["connectorName"]
	log.Debug("connection name is %s", cname)
	connector := cmap["Connection"]
	settings := connector.(map[string]interface{})["settings"]
	cmap["settings"] = settings

	act, tc := setupActivity(t)
	//setup attrs
	tc.SetInput("Connection", connection.(map[string]interface{})["Connection"])
	tc.SetInput("valueType", "String")
	tc.SetInput("MessageString", "hello world")
	tc.SetInput("queue", "flogo.complex")
	tc.SetInput("GenCorrelationID", false)
	tc.SetInput("messageType", "Datagram")

	mqmdData := make(map[string]interface{})
	err = json.Unmarshal([]byte(MQMDData), &mqmdData)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	complex := &data.ComplexObject{Metadata: "", Value: mqmdData}
	tc.SetInput("MQMD", complex)

	//setup props
	propData := make(map[string]interface{})
	err = json.Unmarshal([]byte(PropData), &propData)
	if err != nil {
		fmt.Printf("Error making props: %s", err)
		t.Errorf("Error: %s", err.Error())
	}
	propComplex := &data.ComplexObject{Metadata: "", Value: propData}
	tc.SetInput("properties", propComplex)

	//setup attrs
	for i := 0; i < 1000000; i++ {
		act.Eval(tc)

		complexOutput := tc.GetOutput("Output")
		assert.NotNil(t, complexOutput)

		outputData := complexOutput.(*data.ComplexObject).Value
		dataBytes, err := json.Marshal(outputData)
		assert.Nil(t, err)
		assert.NotNil(t, dataBytes)
		//fmt.Printf("\n%s\n", string(dataBytes))
		time.Sleep(1 * time.Millisecond)

	}
	//check result attr
}
