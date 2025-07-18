package dftrigger

import (
	dfcontext "context"
	"fmt"
	"time"

	datafactory "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"

	"github.com/project-flogo/core/activity"
	azdatafactory "github.com/tibco/wi-azdatafactory/src/app/AzureDataFactory"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {

	_ = activity.Register(&Activity{}, New)
}

// Activity is a stub for your Activity implementation
type Activity struct {
}

// New for Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil

}

// Metadata implements activity.Activity.Metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(context activity.Context) (done bool, err error) {
	log := context.Logger()
	log.Info("Executing Activity DataFactory-Trigger")
	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	connConfig, ok := input.Connection.GetConnection().(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("Failed to read connection configuration from connection manager")
	}
	operation := input.Operation
	if operation == "" {
		return false, activity.NewError("operation is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	subscriptionID := input.SubscriptionId
	resourceGroupName := input.ResourceGroup
	datafactoryName := input.DataFactories
	triggerName := input.DfTrigger

	tenantID := connConfig["tenantId"].(string)
	clientID := connConfig["clientID"].(string)
	userName := connConfig["userName"].(string)
	password := connConfig["password"].(string)

	var lastUpdatedAfter string
	var lastUpdatedBefore string

	inputMap, ok := input.Input.(map[string]interface{})
	paramMap := make(map[string]string)
	if ok && len(inputMap) > 0 {

		parameters := inputMap["parameters"]
		for k, v := range parameters.(map[string]interface{}) {
			paramMap[k] = fmt.Sprint(v)
		}
		if paramMap["resourceGroupName"] != "" {
			resourceGroupName = paramMap["resourceGroupName"]
		}
		if paramMap["factoryName"] != "" {
			datafactoryName = paramMap["factoryName"]
		}
		if paramMap["subscriptionId"] != "" {
			subscriptionID = paramMap["subscriptionId"]
		}
	}
	if subscriptionID == "" {
		return false, activity.NewError("subscriptionID is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	if resourceGroupName == "" {
		return false, activity.NewError("resourceGroupName is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	if datafactoryName == "" {
		return false, activity.NewError("datafactoryName is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	if operation != "Query Trigger Runs" {
		if paramMap["triggerName"] != "" {
			triggerName = paramMap["triggerName"]
		}
		if triggerName == "" {
			return false, activity.NewError("triggerName is not configured", "AZURE-DATAFACTORY-1004", nil)
		}
	} else {
		if paramMap["lastUpdatedBefore"] != "" {
			lastUpdatedBefore = paramMap["lastUpdatedBefore"]
		}
		if len(lastUpdatedBefore) <= 0 || lastUpdatedBefore == "" {
			return false, activity.NewError("lastUpdatedBefore is not configured", "AZURE-DATAFACTORY-1004", nil)
		}
		if paramMap["lastUpdatedAfter"] != "" {
			lastUpdatedAfter = paramMap["lastUpdatedAfter"]
		}
		if len(lastUpdatedAfter) <= 0 || lastUpdatedAfter == "" {
			return false, activity.NewError("lastUpdatedAfter is not configured", "AZURE-DATAFACTORY-1004", nil)
		}
	}
	var statusCode int
	ctx, cancel := dfcontext.WithTimeout(dfcontext.Background(), 30*time.Second)
	defer cancel()
	if err != nil {
		log.Error(err)
		return false, err
	}
	clientCred, err := azdatafactory.GetAzureClient(tenantID, clientID, userName, password)
	if err != nil {
		return false, err
	}
	triggerClient, err := datafactory.NewTriggersClient(subscriptionID, clientCred, nil)
	if err != nil {
		return false, err
	}
	triggerRunClient, err := datafactory.NewTriggerRunsClient(subscriptionID, clientCred, nil)
	if err != nil {
		return false, err
	}
	msgResponse := make(map[string]interface{})
	if operation == "Activate/Start Trigger" {
		_, err := triggerClient.BeginStart(ctx, resourceGroupName, datafactoryName, triggerName, nil)
		if err != nil {
			log.Error(err)
			return false, err
		}
		// TODO use result
		// log.Debug(result.())
		// Statuscode not present in result
		statusCode = 200
	} else if operation == "De-Activate/Stop Trigger" {
		_, err := triggerClient.BeginStop(ctx, resourceGroupName, datafactoryName, triggerName, nil)
		if err != nil {
			log.Error(err)
			return false, err
		}
		// TODO use result
		// log.Info(result.Done())
		statusCode = 200
	} else {
		filterParam := datafactory.RunFilterParameters{}
		filterParam.LastUpdatedAfter = &time.Time{}
		filterParam.LastUpdatedBefore = &time.Time{}
		//	str := "2018-10-04T00:00:00.371Z"
		beforetime, err := time.Parse(time.RFC3339, lastUpdatedBefore)
		if err != nil {
			log.Error(err)
			return false, err
		}
		aftertime, err := time.Parse(time.RFC3339, lastUpdatedAfter)
		if err != nil {
			log.Error(err)
			return false, err
		}
		*filterParam.LastUpdatedBefore = beforetime
		*filterParam.LastUpdatedAfter = aftertime
		result, err := triggerRunClient.QueryByFactory(ctx, resourceGroupName, datafactoryName, filterParam, nil)
		if err != nil {
			log.Error(err)
			return false, err
		}
		// if not error setting statuscode to 200
		statusCode = 200
		triggerRunsList := result.Value
		triggerList := make([]map[string]interface{}, len(triggerRunsList))
		for i, element := range triggerRunsList {
			props := make(map[string]string)
			triggeredpipelines := make(map[string]string)
			props["TriggerTime"] = *element.Properties["TriggerTime"]
			props["ScheduleTime"] = *element.Properties["ScheduleTime"]
			for k, v := range element.TriggeredPipelines {
				triggeredpipelines[k] = *v
			}
			triggerList[i] = map[string]interface{}{"triggerName": *element.TriggerName,
				"triggerType":         *element.TriggerType,
				"triggerRunId":        *element.TriggerRunID,
				"triggerRunTimestamp": *element.TriggerRunTimestamp,
				"message":             *element.Message,
				"status":              element.Status,
				"properties":          props,
				"triggeredpipelines":  triggeredpipelines,
			}
		}
		msgResponse["value"] = triggerList
	}
	log.Info("Received Response from azure datafactory backend")
	msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."
	msgResponse["statusCode"] = statusCode
	if operation != "Query Trigger Runs" {
		msgResponse["factoryName"] = datafactoryName
		msgResponse["triggerName"] = triggerName
	}
	output := &Output{}
	output.Output = msgResponse
	log.Debugf("Output is %s", output)
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	log.Info("Execution of Activity DataFactory-Trigger completed")
	return true, nil
}
