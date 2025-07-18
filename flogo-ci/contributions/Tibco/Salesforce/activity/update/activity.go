package update

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
	_ = activity.Register(&UpdateActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &UpdateActivity{}, nil
}

type UpdateActivity struct {
}

func (*UpdateActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *UpdateActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debug("Executing Salesforce Update Activity")
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
	resBytes, err := update(context, objectName, inputBody, sscm)
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
	context.Logger().Info("Salesforce Update Activity successfully executed")
	return true, nil
}

func update(context activity.Context, objName string, input interface{}, sscm *sfconnection.SalesforceSharedConfigManager) ([]byte, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	//url := token.InstanceUrl + REQ_URI
	//REQ_URI = "/services/data/v48.0/composite/batch"
	url := token.InstanceUrl + "/services/data/" + apiVersion + "/composite/batch"

	reqData, err := covertInput(context, input, objName, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("Parse delete input error %s", err.Error())
	}

	//https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/requests_composite_batch.htm

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

	tempMap := make(map[string]interface{})

	switch content.(type) {
	case string:
		context.Logger().Debugf("Activity Input(string): %s ", content.(string))

		err := json.Unmarshal([]byte(content.(string)), &tempMap)
		if err != nil {
			return nil, fmt.Errorf("Can't construct update input [%s]", err.Error())
		}
	case map[string]interface{}:

		b, err := json.Marshal(content)
		if err != nil {
			return nil, fmt.Errorf("Can't marshal input data [%s] ", err.Error())
		}
		context.Logger().Debugf("Activity Input: %s ", string(b))
		tempMap = content.(map[string]interface{})

	default:
		context.Logger().Errorf("Invalid input: %+v ", content)
		return nil, fmt.Errorf("Invalid input %+v", content)
	}

	recordsObj := tempMap[objectName]
	b, err := json.Marshal(recordsObj)
	if err != nil {
		return nil, fmt.Errorf("Can't construct update input [%s]", err.Error())
	}

	err = json.Unmarshal(b, &input)

	var items []BatchItem
	for _, v := range input.Records {
		item := BatchItem{}
		item.Method = "PATCH"
		item.Url = "/services/data/" + apiVersion + "/sobjects/" + objectName + "/" + v["Id"].(string)

		delete(v, "Id")
		item.RichInput = v

		items = append(items, item)

	}

	reqData.BatchRequests = items

	return &reqData, nil
}

// Activity's input struct
/*
 {
  "Account": {
     "records":[
        {
         "Name": "123"
        }
      ]
    }
   }
*/

type ActivityInput struct {
	Records []map[string]interface{} `json:"records"`
}

// =========Salesforce request struct========
type Request struct {
	BatchRequests []BatchItem `json:"batchRequests"`
}

type BatchItem struct {
	Method    string                 `json:"method"`
	Url       string                 `json:"url"`
	RichInput map[string]interface{} `json:"richInput"`
}

// =============Response struct=============
type Response struct {
	HasErrors bool        `json:"hasErrors"`
	Results   []ResultObj `json:"results"`
}

type ResultObj struct {
	StatusCode int     `json:"statusCode"`
	ErrResult  []Error `json:"result,omitempty"`
}

type Error struct {
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
}
