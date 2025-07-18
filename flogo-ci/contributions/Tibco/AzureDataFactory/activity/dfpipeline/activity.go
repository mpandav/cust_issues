package dfpipeline

import (
	dfcontext "context"
	"fmt"
	"time"

	datafactory "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory"

	"github.com/project-flogo/core/activity"
	azdatafactory "github.com/tibco/wi-azdatafactory/src/app/AzureDataFactory"
)

//var re = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

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
	log.Info("Executing Activity  dfpipeline")

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
	pipelineName := input.DfPipeline

	// var accessToken string
	tenantID := connConfig["tenantId"].(string)
	clientID := connConfig["clientID"].(string)
	userName := connConfig["userName"].(string)
	password := connConfig["password"].(string)

	var lastUpdatedAfter string
	var lastUpdatedBefore string
	var pipelineRunID string
	pipelineParameters := make(map[string]interface{})
	inputMap, ok := input.Input.(map[string]interface{})
	paramMap := make(map[string]string)
	if ok && len(inputMap) > 0 {

		parameters := inputMap["parameters"]
		for k, v := range parameters.(map[string]interface{}) {
			if k == "pipelineParameters" && operation == "Run Once" && v != nil {
				var ok bool
				pipelineParameters, ok = v.(map[string]interface{})
				if !ok {
					log.Debugf("Err getting the pipeline parameters from flow input. Executing Run Once without Pipeline Parameters")
					pipelineParameters = nil
				}
				continue
			}
			paramMap[k] = fmt.Sprint(v)
		}
		log.Debugf("ParamMap: %v", paramMap)
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
		return false, activity.NewError("subscriptionId is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	if resourceGroupName == "" {
		return false, activity.NewError("resourceGroupName is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	if datafactoryName == "" {
		return false, activity.NewError("datafactoryName is not configured", "AZURE-DATAFACTORY-1004", nil)
	}
	log.Debugf("Operation Selected: %v", operation)
	log.Debugf("ResourceGroupName: %v, SubscriptionID: %v, DataFactoryName: %v", resourceGroupName, subscriptionID, datafactoryName)

	//set operation specific parameters
	switch operation {
	case "Run Once":
		{
			if paramMap["pipelineName"] != "" {
				pipelineName = paramMap["pipelineName"]
			}
			if pipelineName == "" {
				return false, activity.NewError("pipelineName is not configured", "AZURE-DATAFACTORY-1004", nil)
			}
		}
	case "Query Runs":
		{
			if paramMap["pipelineName"] != "" {
				pipelineName = paramMap["pipelineName"]
			}
			if pipelineName == "" {
				log.Debug("Received empty string as pipelineName ")
			}
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
	pipelineRunClient, err := datafactory.NewPipelineRunsClient(subscriptionID, clientCred, nil)
	if err != nil {
		return false, err
	}
	pipelineClient, err := datafactory.NewPipelinesClient(subscriptionID, clientCred, nil)
	if err != nil {
		return false, err
	}
	msgResponse := make(map[string]interface{})
	switch operation {

	case "Cancel Run":
		{
			log.Debug("Executing Cancel Run..")
			if paramMap["pipelineRunId"] != "" {
				pipelineRunID = paramMap["pipelineRunId"]
			}
			if len(pipelineRunID) <= 0 || pipelineRunID == "" {
				return false, activity.NewError("pipelineRunId is not configured", "AZURE-DATAFACTORY-1004", nil)
			}
			result, err := pipelineRunClient.Cancel(ctx, resourceGroupName, datafactoryName, pipelineRunID, nil)
			if err != nil {
				log.Error(err)
				return false, err
			}
			log.Debug("Result of pipelineRunClient.Cancel is : ", result)
			//  Statuscode not present in result setting to 200 if not error
			statusCode = 200
			msgResponse["pipelineRunId"] = paramMap["pipelineRunId"]
		}

	case "Run Once":
		{
			log.Debug("Executing Run Once..")
			areValuesOverriden := paramMap["pipelineName"] != "" || paramMap["subscriptionId"] != "" || paramMap["resourceGroupName"] != "" || paramMap["factoryName"] != ""
			log.Debugf("Pipeline Parameters before filter: %v", pipelineParameters)
			pipelineParameters, err = getFilteredPipelineParameters(ctx, pipelineClient, resourceGroupName, datafactoryName, pipelineName, pipelineParameters, areValuesOverriden)
			if err != nil {
				return false, err
			}
			// Created Parameters inside options
			opts := datafactory.PipelinesClientCreateRunOptions{
				Parameters: pipelineParameters,
			}
			log.Debugf("Pipeline Parameters: %v", pipelineParameters)

			result, err := pipelineClient.CreateRun(ctx, resourceGroupName, datafactoryName, pipelineName, &opts)
			if err != nil {
				log.Error(err.Error())
				return false, err
			}
			// Statuscode not present in result setting to 200 if not error
			statusCode = 200
			msgResponse["pipelineRunId"] = *result.RunID
		}
	case "Query Runs":
		{
			log.Debug("Executing Query Pipeline Runs..")

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

			filterParam := datafactory.RunFilterParameters{
				LastUpdatedAfter:  &aftertime,
				LastUpdatedBefore: &beforetime,
			}
			if pipelineName != "" {
				runQueryFilterOpts := datafactory.RunQueryFilter{
					Operand:  &datafactory.PossibleRunQueryFilterOperandValues()[5],
					Operator: &datafactory.PossibleRunQueryFilterOperatorValues()[0],
					Values:   []*string{&pipelineName},
				}
				filterParam.Filters = append(filterParam.Filters, &runQueryFilterOpts)
			}
			result, err := pipelineRunClient.QueryByFactory(ctx, resourceGroupName, datafactoryName, filterParam, nil)
			if err != nil {
				log.Error(err)
				return false, err
			}
			// Statuscode not present in result setting to 200 if not error
			statusCode = 200
			// TODO  Result has fields result.PipelineRunsQueryResponse, result.ContinuationToken

			pipelineRunsList := result.Value
			pipelineRunsListMap := make([]map[string]interface{}, len(pipelineRunsList))
			for i, element := range pipelineRunsList {
				invokedBy := make(map[string]string)
				parameters := make(map[string]string)
				invokedBy["id"] = *element.InvokedBy.ID
				invokedBy["name"] = *element.InvokedBy.Name
				for k, v := range element.Parameters {
					parameters[k] = *v
				}
				var durationInMs int32
				var runEnd string
				var lastUpdated string
				if element.DurationInMs == nil {
					durationInMs = 0
				} else {
					durationInMs = (*element.DurationInMs)
				}
				if element.RunEnd == nil {
					runEnd = ""
				} else {
					runEnd = (*element.RunEnd).String()
				}
				if element.LastUpdated == nil {
					lastUpdated = ""
				} else {
					lastUpdated = (*element.LastUpdated).String()
				}

				pipelineRunsListMap[i] = map[string]interface{}{"runId": *element.RunID,
					"pipelineName": *element.PipelineName,
					"runStart":     *element.RunStart,
					"durationInMs": durationInMs,
					"runEnd":       runEnd,
					"status":       element.Status,
					"message":      *element.Message,
					"lastUpdated":  lastUpdated,
					"invokedBy":    invokedBy,
					"parameters":   parameters,
				}
			}
			msgResponse["value"] = pipelineRunsListMap
		}
	default:
		return false, fmt.Errorf("Invalid operation selected : %v", operation)

	}
	msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."
	//  StatusCode
	msgResponse["statusCode"] = statusCode
	if operation != "Query Runs" {
		msgResponse["factoryName"] = datafactoryName
		msgResponse["pipelineName"] = pipelineName
	}
	output := &Output{}
	output.Output = msgResponse
	log.Debugf("Output is %s", output)
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	log.Info("Executing Activity  dfpipeline completed")
	return true, nil
}

/*
Refer: WIAZDF-88
Datafactory Pipeline values are picked either from Settings or Input tab.
Values set through Settings are overriden using Input tab values
Pipeline parameters are set at design time based on Settings since it is not possible to set the pipeline parameters based on overriden values at runtime.
Hence, this is just an extra support we are providing where,
let
Pipeline1: params{ p1, p2, p3} => set at design time using Settings
Pipeline2: params{ p1, p2, p7, p8} => parameters of the overriden pipeline using Input Tab
we pick only matching params from the Pipeline1 and set others to default value i.e Pipeline2 will become
Pipeline2: params{ p1, p2, p7, p8} => {p1, p2, p7(default Value), p8(default value)}
If there is no default value, we just return error
*/
func getFilteredPipelineParameters(ctx dfcontext.Context, pipelineClient *datafactory.PipelinesClient, resourceGroupName string, factoryName string, pipelineName string, pipelineParameters map[string]interface{}, areValuesOverriden bool) (map[string]interface{}, error) {

	oPipelineParameters := make(map[string]interface{})

	pipelineResource, err := pipelineClient.Get(ctx, resourceGroupName, factoryName, pipelineName, nil)
	if err != nil || pipelineResource.PipelineResource.Properties == nil {
		return nil, fmt.Errorf("error fetching Pipeline Resource, %v", err)
	}

	for k, v := range pipelineResource.PipelineResource.Properties.Parameters {
		val, found := pipelineParameters[k]
		// isTypeMatched := true
		if found && areValuesOverriden {
			// isTypeMatched = pipelineParameterTypes[k] == string(v.Type)
			// if isTypeMatched {
			oPipelineParameters[k] = val
			// }
		}
		// if v.DefaultValue == nil && (!found || (found && !isTypeMatched)) {
		if v.DefaultValue == nil && (!found) {
			return nil, fmt.Errorf("Parameter %v defined for the pipeline %v is neither configured nor have a default value. Ensure parameter is set before triggering pipeline.", k, pipelineName)
		}
	}
	if areValuesOverriden {
		return oPipelineParameters, nil
	}
	return pipelineParameters, nil
}

// Not using Metadata from UI
// func getPipelineParamsMetadataFromInput(inputMetadata string) (map[string]string, error) {
// }
