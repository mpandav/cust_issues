package getqueryjobresults

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/project-flogo/core/activity"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&GetQueryJobResultsActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &GetQueryJobResultsActivity{}, nil
}

type GetQueryJobResultsActivity struct {
}

func (*GetQueryJobResultsActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *GetQueryJobResultsActivity) Eval(context activity.Context) (done bool, err error) {
	context.Logger().Debug("Executing Salesforce Bulk Query Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	sscm, _ := input.SalesforceConnection.(*sfconnection.SalesforceSharedConfigManager)

	justToChecktoken := sscm.SalesforceToken
	if justToChecktoken == nil {
		return false, fmt.Errorf("Get salesforce token from connection field error [%s]", err.Error())
	}
	objectName := input.ObjectName
	if objectName == "" {
		return false, errors.New("Object name is required")
	}
	context.Logger().Debugf("objectName is %s", objectName)

	inputData := input.Input

	if inputData == nil {
		return false, fmt.Errorf("Input data is empty")
	}

	queryObject, err := fetchRecordsAndClosejob(context, objectName, inputData, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}
	context.Logger().Debug("Salesforce Job Closed")
	// outschema := context.GetInput("outputSchema")
	// if outschema != nil && queryObject != nil {
	// 	schema := outschema.(string)
	// 	err = jsonschema.ValidateFromObject(schema, queryObject)
	// 	if err != nil {
	// 		return false, err
	// 	}
	// }

	context.Logger().Debug("Operation success, set output")
	output := &Output{}
	output.Output = queryObject
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce Query Job Results Activity successfully executed")
	return true, nil
}

func fetchRecordsAndClosejob(context activity.Context, objName string, input interface{}, sscm *sfconnection.SalesforceSharedConfigManager) (map[string]interface{}, error) {

	var inputMap map[string]interface{}
	inputMap = input.(map[string]interface{})
	queryJobID := ""
	locator := ""
	maxRecords := 0

	q, ok := inputMap["queryJobId"]
	if ok {
		queryJobID = q.(string)
	} else {
		return nil, fmt.Errorf("query job id is mandatory")
	}

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	l, ok := inputMap["locator"]
	if ok {
		locator = l.(string)
	}

	if locator == "" || locator == "null" {
		url := sscm.SalesforceToken.InstanceUrl + "/services/data/" + apiVersion + "/jobs/query/" + queryJobID
		resData := make(map[string]interface{})
		resBody, err := salesforce.RestCall(sscm, "GET", url, nil, context.Logger())
		if err != nil {
			return nil, err
		}
		var result interface{}
		json.Unmarshal(resBody, &result)
		resData = result.(map[string]interface{})
		obj := resData["object"].(string)
		if strings.EqualFold(obj, objName) == false {
			return nil, fmt.Errorf("Object name of job and selected object name in settings doesnot match")
		}

	}
	m, ok := inputMap["maxRecords"]
	if ok {
		maxRecords = int(m.(float64))
	}

	queryurl := sscm.SalesforceToken.InstanceUrl + "/services/data/" + apiVersion + "/jobs/query/" + queryJobID + "/results"
	if locator != "" && locator != "null" {
		queryurl = queryurl + "?locator=" + locator
		if maxRecords != 0 {
			queryurl = queryurl + "&maxRecords=" + strconv.Itoa(maxRecords)
		}
	} else if maxRecords != 0 {
		queryurl = queryurl + "?maxRecords=" + strconv.Itoa(maxRecords)
	}

	resBody, err := salesforce.RestCallForCSVResponse(sscm, "GET", queryurl, nil, context.Logger())
	if err != nil {
		return nil, err
	}
	if resBody == nil {
		return nil, nil
	}
	outputObj := make(map[string]interface{})
	outputObj[objName] = resBody
	if resBody["Locator"] == "null" || resBody["Locator"] == "" {
		closeJobURL := sscm.SalesforceToken.InstanceUrl + "/services/data/" + apiVersion + "/jobs/query/" + queryJobID
		salesforce.RestCall(sscm, "DELETE", closeJobURL, nil, context.Logger())
	}
	return outputObj, nil
}
