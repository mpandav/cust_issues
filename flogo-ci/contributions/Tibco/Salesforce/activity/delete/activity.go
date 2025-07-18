package delete

import (
	"encoding/json"
	"errors"
	"fmt"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&DeleteActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &DeleteActivity{}, nil
}

type DeleteActivity struct {
}

func (*DeleteActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *DeleteActivity) Eval(context activity.Context) (done bool, err error) {
	context.Logger().Debug("Executing Salesforce Delete Activity")
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

	resBytes, err := delete(context, objectName, inputBody, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}

	var result interface{}
	json.Unmarshal(resBytes, &result)

	context.Logger().Debugf("==>Activity Output: %s ", string(resBytes))

	context.Logger().Debug("Operation success, set output")
	output := &Output{}
	output.Output = result.(map[string]interface{})
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce Delete Activity successfully executed")
	return true, nil
}

func delete(context activity.Context, objName string, input interface{}, sscm *sfconnection.SalesforceSharedConfigManager) ([]byte, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	url := token.InstanceUrl + "/services/data/" + apiVersion + "/composite/batch"

	reqData, err := covertInput(context, input, objName, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("Parse delete input error %s", err.Error())
	}

	b, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	resBody, err := salesforce.RestCall(sscm, "POST", url, b, context.Logger())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func covertInput(context activity.Context, content interface{}, objectName string, apiVersion string) (*Request, error) {
	input := ActivityInput{}
	reqData := Request{}

	switch content.(type) {
	case string:
		context.Logger().Debugf("Activity Input(string): %s ", content.(string))

		inb := []byte(content.(string))
		err := json.Unmarshal(inb, &input)
		if err != nil {
			return nil, fmt.Errorf("Can't construct delete input [%s]", err.Error())
		}
	default:
		inb, err := json.Marshal(content)
		if err != nil {
			return nil, err
		}

		context.Logger().Debugf("Activity Input: %s ", string(inb))

		err = json.Unmarshal(inb, &input)
		if err != nil {
			return nil, fmt.Errorf("Can't construct delete input [%s]", err.Error())
		}
	}

	var items []BatchItem
	for _, v := range input.Data {
		if v.Id != "" {
			item := BatchItem{}
			item.Method = "DELETE"
			item.Url = "/services/data/" + apiVersion + "/sobjects/" + objectName + "/" + v.Id

			items = append(items, item)
		}
	}

	reqData.BatchRequests = items

	return &reqData, nil
}

// Activity's input struct
type ActivityInput struct {
	Data []IdObject `json:"data,omitempty"`
}

type IdObject struct {
	Id string `json:"id,omitempty"`
}

// Salesforce request body struct
type Request struct {
	BatchRequests []BatchItem `json:"batchRequests"`
}

type BatchItem struct {
	Method string `json:"method"`
	Url    string `json:"url"`
}
