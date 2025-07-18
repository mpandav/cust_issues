package upsert

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/project-flogo/core/activity"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&UpsertActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &UpsertActivity{}, nil
}

type UpsertActivity struct {
}

func (*UpsertActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *UpsertActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debug("Executing Salesforce Upsert Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	sscm, _ := input.SalesforceConnection.(*sfconnection.SalesforceSharedConfigManager)

	objectName := input.ObjectName
	if objectName == "" {
		return false, errors.New("Object name is required")
	}
	context.Logger().Debugf("objectName: %s", objectName)

	inputBody := input.Input
	if inputBody == nil {
		return false, fmt.Errorf("Input data is empty")
	}
	externalIdFieldName := input.ExternalIdFieldName
	context.Logger().Debug("externalIdFieldName: ", externalIdFieldName)
	allOrNone := input.AllOrNone
	context.Logger().Debug("allOrNone: ", allOrNone)

	resBytes, err := upsert(context, objectName, inputBody, externalIdFieldName, allOrNone, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}

	var result interface{}
	json.Unmarshal(resBytes, &result)

	//context.Logger().Debug("===>Activity Output: %s ", string(resBytes))

	context.Logger().Debug("Operation success, set output")
	objectResponse := make(map[string]interface{})
	objectResponse["response"] = result
	output := &Output{}
	output.Output = objectResponse
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce Upsert Activity successfully executed")
	return true, nil
}

func upsert(context activity.Context, objName string, input map[string]interface{}, externalIdFieldName string, allOrNone bool, sscm *sfconnection.SalesforceSharedConfigManager) ([]byte, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}

	url := token.InstanceUrl + "/services/data/" + sscm.APIVersion + "/composite/sobjects/" + objName + "/" + externalIdFieldName

	data, dataErr := covertInput(context, input, objName)
	// fmt.Printf("+++%v", data)
	data.AllOrNone = allOrNone

	if dataErr != nil {
		return nil, dataErr
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resBody, err := salesforce.RestCall(sscm, "PATCH", url, b, context.Logger())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func covertInput(context activity.Context, content interface{}, objName string) (*Request, error) {
	var recordsObj interface{}

	switch content.(type) {
	case string:
		context.Logger().Debugf("Activity Input(string): %s ", content.(string))

		tempMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(content.(string)), &tempMap)
		if err != nil {
			return nil, err
		}
		recordsObj = tempMap[objName]
	default:
		inb, err := json.Marshal(content)
		if err != nil {
			return nil, err
		}
		context.Logger().Debugf("Activity Input: %s ", string(inb))

		inputMap := content.(map[string]interface{})
		recordsObj = inputMap[objName]
	}

	b, err := json.Marshal(recordsObj)
	if err != nil {
		return nil, err
	}

	reqData := &Request{}
	json.Unmarshal(b, reqData)
	for _, v := range reqData.Records {

		attributes := make(map[string]interface{})
		attributes["type"] = objName
		v["attributes"] = attributes

	}
	return reqData, nil
}

type Request struct {
	Records   []map[string]interface{} `json:"records,omitempty"`
	AllOrNone bool                     `json:"allOrNone,omitempty"`
}

type Response struct {
	Results []map[string]interface{} `json:"results,omitempty"`
}
