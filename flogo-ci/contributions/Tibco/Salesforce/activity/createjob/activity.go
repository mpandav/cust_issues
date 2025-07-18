package createjob

import (
	"encoding/json"
	"fmt"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&CreateJobActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &CreateJobActivity{}, nil
}

type CreateJobActivity struct {
}

func (*CreateJobActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *CreateJobActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debug("Executing Salesforce CreateJobActivity Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	context.Logger().Debugf("Input is %s", input)
	sscm, _ := input.SalesforceConnection.(*sfconnection.SalesforceSharedConfigManager)

	justToChecktoken := sscm.SalesforceToken
	if justToChecktoken == nil {
		return false, fmt.Errorf("Get salesforce token from connection field error [%s]", err.Error())
	}

	inputBody := input.Input
	if inputBody == nil {
		return false, fmt.Errorf("Input data is empty")
	}

	operation := input.Operation
	if operation == "" {
		return false, fmt.Errorf("operation value is empty")
	}

	// inputSchema := inputBody.(string)
	// inputMap := make(map[string]interface{})
	// err := json.Unmarshal([]byte(inputSchema), &inputMap)
	// if err != nil {
	// 	panic(err)
	// }

	resBytes, err := createJob(context, operation, inputBody, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}

	var result interface{}
	json.Unmarshal(resBytes, &result)

	//context.Logger().Debug("===>Activity Output: %s ", string(resBytes))

	context.Logger().Debug("Operation success, set output")
	output := &Output{}
	finalResult := make(map[string]interface{})
	finalResult["JobInfo"] = result.(map[string]interface{})
	output.Output = finalResult
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce CreateJob Activity successfully executed")
	return true, nil
}

func createJob(context activity.Context, operation string, input interface{}, sscm *sfconnection.SalesforceSharedConfigManager) ([]byte, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}
	url := sscm.SalesforceToken.InstanceUrl + "/services/data/" + apiVersion + "/jobs/query"

	inputMap := input.(map[string]interface{})
	parameters := inputMap["parameters"]
	paramMap := parameters.(map[string]interface{})
	paramMap["operation"] = operation
	body, err := json.Marshal(paramMap)
	if err != nil {
		return nil, err
	}

	resBody, err := salesforce.RestCall(sscm, "POST", url, body, context.Logger())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

type Request struct {
	Records []map[string]interface{} `json:"records,omitempty"`
}

type Response struct {
	HasErrors bool                     `json:"hasErrors,omitempty"`
	Results   []map[string]interface{} `json:"results,omitempty"`
}
