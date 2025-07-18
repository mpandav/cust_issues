package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&CreateActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &CreateActivity{}, nil
}

type CreateActivity struct {
}

func (*CreateActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *CreateActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debug("Executing Salesforce Create Activity")
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
	context.Logger().Debugf("objectName is %s", objectName)

	inputBody := input.Input
	if inputBody == nil {
		return false, fmt.Errorf("Input data is empty")
	}

	resBytes, err := create(context, objectName, inputBody, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}

	var result interface{}
	json.Unmarshal(resBytes, &result)

	//context.Logger().Debug("===>Activity Output: %s ", string(resBytes))

	context.Logger().Debug("Operation success, set output")
	output := &Output{}
	output.Output = result.(map[string]interface{})
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce Create Activity successfully executed")
	return true, nil
}

func create(context activity.Context, objName string, input interface{}, sscm *sfconnection.SalesforceSharedConfigManager) ([]byte, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	url := token.InstanceUrl + "/services/data/" + apiVersion + "/composite/tree/" + objName

	data, dataErr := covertInput(context, input, objName)
	if dataErr != nil {
		return nil, dataErr
	}
	//https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/dome_composite_sobject_tree_flat.htm#topic-title

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resBody, err := salesforce.RestCall(sscm, "POST", url, b, context.Logger())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

/*
	 Input:
		{
		  "Account": {
		    "records":[
		      {
			"Name": "123"
		      }
		    ]
		  }
		}

	 Request:
		{
		  "records":[
		      {
			"attributes": {"type": "Account", "referenceId": "1"}
			"Name": "123"
		      }
		   ]
		}
*/
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
	for i, v := range reqData.Records {

		attributes := make(map[string]interface{})
		attributes["type"] = objName
		attributes["referenceId"] = strconv.Itoa(i)
		v["attributes"] = attributes

	}

	return reqData, nil
}

type Request struct {
	Records []map[string]interface{} `json:"records,omitempty"`
}

type Response struct {
	HasErrors bool                     `json:"hasErrors,omitempty"`
	Results   []map[string]interface{} `json:"results,omitempty"`
}
