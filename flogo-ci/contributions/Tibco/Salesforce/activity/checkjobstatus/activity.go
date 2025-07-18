package checkjobstatus

import (
	"encoding/json"
	"fmt"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&CheckJobStatusActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &CheckJobStatusActivity{}, nil
}

type CheckJobStatusActivity struct {
}

func (*CheckJobStatusActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *CheckJobStatusActivity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debug("Executing Salesforce CheckJobStatusActivity Activity")
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

	inputData := input.Input
	if inputData == nil {
		return false, fmt.Errorf("Input data is empty")
	}

	operation := input.Operation
	if operation == "" {
		return false, fmt.Errorf("Operation value is empty")
	}

	waitforcompletion := input.Waitforcompletion
	if waitforcompletion == "" {
		return false, fmt.Errorf("Waitforcompletion value is empty")
	}
	timeout := 0
	interval := 0
	if waitforcompletion == "Yes" {
		timeout = input.Timeout
		if timeout <= 0 {
			return false, fmt.Errorf("Timeout value is not valid")
		}
		interval = input.Interval
		if interval <= 0 {
			return false, fmt.Errorf("Interval value is not valid")
		}
		if timeout < interval {
			return false, fmt.Errorf("Timeout value should be greater than interval")
		}
	}

	resData, err := checkJobStatus(context, operation, waitforcompletion, timeout, interval, inputData, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}
	context.Logger().Debug("Operation success, set output")
	output := &Output{}
	finalResult := make(map[string]interface{})
	finalResult["JobInfo"] = resData
	output.Output = finalResult
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce CheckJobStatus Activity successfully executed")
	return true, nil
}

func checkJobStatus(context activity.Context, operation string, waitforcompletion string, timeout int, interval int, input interface{}, sscm *sfconnection.SalesforceSharedConfigManager) (map[string]interface{}, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	var inputMap map[string]interface{}
	inputMap = input.(map[string]interface{})
	jobID := ""
	url := ""
	q, ok := inputMap["jobId"]
	if ok {
		jobID = q.(string)
	}
	n := 1
	if waitforcompletion == "Yes" {
		n = timeout / interval
	}

	if operation == "query" {
		url = sscm.SalesforceToken.InstanceUrl + "/services/data/" + apiVersion + "/jobs/query/" + jobID
	}
	resData := make(map[string]interface{})
	for i := 0; i < n; i++ {
		resBody, err := salesforce.RestCall(sscm, "GET", url, nil, context.Logger())
		if err != nil {
			return nil, err
		}
		var result interface{}
		json.Unmarshal(resBody, &result)
		resData = result.(map[string]interface{})
		if resData["state"] == "JobComplete" {
			break
		}
	}

	return resData, nil
}

type Request struct {
	Records []map[string]interface{} `json:"records,omitempty"`
}

type Response struct {
	HasErrors bool                     `json:"hasErrors,omitempty"`
	Results   []map[string]interface{} `json:"results,omitempty"`
}
