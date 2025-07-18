package yukonoperation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"

	"github.com/tibco/wi-contrib/environment"
	restClient "github.com/tibco/wi-contrib/ucs/common"
	"github.com/tibco/wi-contrib/ucs/connector/yukon"
)

func init() {
	_ = activity.Register(&YukonOpActivity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

const UCS_PROVIDER_SRV_HEADER_NAME = "UCS_PROVIDER_SRV"

type YukonOpResponse struct {
	Action     string                   `json:"action"`
	DataObject string                   `json:"dataObject"`
	Results    []map[string]interface{} `json:"results"`
}
type Error struct {
	Details string `json:"details"`
	Message string `json:"message"`
	Number  int    `json:"number"`
}
type Results struct {
	Error           map[string]interface{} `json:"error"`
	ObjectsAffected int                    `json:"objectsAffected"`
	OutputData      map[string]interface{} `json:"outputData"`
	Success         bool                   `json:"success"`
}

// YukonOpActivity describes the metadata of the activity as found in the activity.json file
type YukonOpActivity struct {
	activityLogger    log.Logger
	connectionManager *yukon.YukonSharedConfigManager
	action            string
}

type ComplexLookupCondition struct {
	Expr  string                `json:"expr,required"`
	Left  SimpleLookupCondition `json:"left,omitempty"`
	Right SimpleLookupCondition `json:"right,omitempty"`
}

type SimpleLookupCondition struct {
	Expr string      `json:"expr,required"`
	Prop string      `json:"prop,required"`
	Val  interface{} `json:"val,required"`
}

// New will instantiate a new Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	connectionManager := s.YukonConnection.GetConnection().(*yukon.YukonSharedConfigManager)
	logCache := log.ChildLogger(ctx.Logger(), strings.ToLower(connectionManager.ConnectorName)+".activity.operation")

	if s.YukonConnection == nil {
		logCache.Error("Error occurred while reading Connection")
	}

	// flogo..activity.operation
	if connectionManager == nil {
		logCache.Error("Error occurred while reading Connection")
	}

	action := s.Action
	activity := &YukonOpActivity{activityLogger: logCache, connectionManager: connectionManager, action: action}
	return activity, nil
}

// Metadata will return the metadata of the Activity
func (a *YukonOpActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval executes the activity
func (a *YukonOpActivity) Eval(context activity.Context) (done bool, err error) {
	input := &Input{}

	yukonconnection := a.connectionManager
	if a.action != "" {
		input.Action = a.action
	}
	if yukonconnection != nil {
		a.connectionManager.Settings = yukonconnection.GetSettings().(*yukon.Settings)
	}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	var lookupCondition interface{}
	for _, input := range input.Input {
		lookupCondition = input.LookupCondition
	}

	if a.activityLogger == nil {
		a.activityLogger = log.ChildLogger(context.Logger(), strings.ToLower(a.connectionManager.ConnectorName)+".activity."+strings.ToLower(input.Action))
	}

	if input.DataObject == "" {
		return false, activity.NewError("Data object not provided", "UCS-2005", nil)
	}

	if lookupCondition == nil {
		a.activityLogger.Debug("No lookup condition provided")
	}

	if len(input.Input) <= 0 {
		return false, activity.NewError("Input not provided", "UCS-2007", nil)
	}

	if lookupCondition != nil {
		var complexLookupCondition ComplexLookupCondition
		var simpleLookupCondition SimpleLookupCondition
		var lookupconditionBytes []byte
		// queryDetails.From = input.DataObject
		lookupconditionBytes, err = json.Marshal(lookupCondition)
		if err != nil {
			return false, activity.NewError("Error while reading lookup conditions", "UCS-2006", nil)
		}

		json.Unmarshal(lookupconditionBytes, &simpleLookupCondition)

		var finalCondition interface{}
		if simpleLookupCondition.Expr != "" {
			expr := strings.ToLower(simpleLookupCondition.Expr)
			if expr == "and" || expr == "or" {
				json.Unmarshal(lookupconditionBytes, &complexLookupCondition)
				finalCondition = complexLookupCondition
			} else {
				json.Unmarshal(lookupconditionBytes, &simpleLookupCondition)
				finalCondition = simpleLookupCondition
			}
		} else {
			return false, activity.NewError("Expression of lookup condition is empty", "UCS-2006", nil)
		}

		inputDatas := input.Input
		for index, inputData := range inputDatas {
			inputData.LookupCondition = finalCondition
			inputDatas[index] = inputData
		}
		input.Input = inputDatas
	}

	b, err := json.Marshal(input.Input)
	if err != nil {
		return false, err
	}
	a.activityLogger.Debugf("Activity Input: %s", string(b))

	opResponse, err := a.executeOperation(input)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Error executing activityname: %s operation: %s - %s ", context.Name(), input.Action, err.Error()), "Yukon-2008", nil)
	}
	resultCount := len(opResponse.Results)
	if resultCount != 1 {
		return false, activity.NewError(fmt.Sprintf("Error executing Operation: %s - Operation returned %d results", input.Action, resultCount), "Yukon-2009", nil)
	}

	activityOutput := &Output{}
	activityOutput.Output = OutputDetails{Action: opResponse.Action, DataObject: opResponse.DataObject, Results: opResponse.Results}
	err = context.SetOutputObject(activityOutput)
	if err != nil {
		return false, activity.NewError("Error setting output for Activity [%s]: %s", context.Name(), err.Error())
	}
	if input.Action != "" {
		a.activityLogger.Info(input.Action + "activity successfully executed")
	}

	return true, nil
}

func (a *YukonOpActivity) executeOperation(input *Input) (*YukonOpResponse, error) {
	activityLogger := a.activityLogger
	baseURL := environment.GetIntercomURL()
	subID := environment.GetTCISubscriptionId()
	subUname := environment.GetTCISubscriptionUName()
	var action string

	// baseURL := "https://account.ucs.tcie.pro"
	// subID:= "01F6H55CTEQ2DRRSKANPCWK452"

	yukonconnection := a.connectionManager

	ucscookievalue := yukonconnection.UCSProviderCookie
	if ucscookievalue == "" {
		return nil, activity.NewError("Error occurred while reading connection in activity", "UCS-2011", nil)
	}

	if yukonconnection == nil {
		activityLogger.Debug("yukon connection is nil in execute operation")
		return nil, activity.NewError("Error occurred while reading connection in activity", "UCS-2011", nil)
	}

	instanceID := yukonconnection.InstanceID
	providerPathPrefix := yukonconnection.ProviderPathPrefix
	action = a.action

	if instanceID == "" || providerPathPrefix == "" {
		return nil, activity.NewError(fmt.Sprintf("Error executing Operation: %s. Missing InstanceID or Provider Path Prefix", action), "Yukon-2011", nil)
	}

	uri := baseURL + fmt.Sprintf("%s/v1/instance/%s/actions/%s/execute?gsbc=%s",
		yukonconnection.ProviderPathPrefix, yukonconnection.InstanceID, action, subID)

	headers := make(map[string]string)
	headers["X-Atmosphere-For-User"] = subUname
	headers["X-Atmosphere-Tenant-Id"] = "ucs"
	headers["X-Atmosphere-Subscription-Id"] = subID
	headers["Content-Type"] = "application/json"
	headers["Connection"] = "keep-alive"
	headers["Accept"] = "application-json"
	headers["Cookie"] = UCS_PROVIDER_SRV_HEADER_NAME + "=" + ucscookievalue

	jsonValue, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	resp, err := restClient.GetRestResponse(activityLogger, yukonconnection.YukonClient, restClient.MethodPOST, uri, headers, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}

	var opResponse YukonOpResponse
	err = json.NewDecoder(resp.Body).Decode(&opResponse)
	if err != nil {
		return nil, err
	}

	return &opResponse, nil
}
