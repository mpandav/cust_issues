package action

import (
	"encoding/json"
	"fmt"

	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"

	"github.com/project-flogo/core/activity"
)

const (
	//Fields
	CONNECTION = "Connection"
	RESOURCE   = "Resource"
	INPUT      = "input"

	FIELD_ACTION = "Action"

	//Options
	VALUE_CAMPAIGNS = "Campaigns"
	VALUE_SEND      = "Send"
	VALUE_SCHEDULE  = "Schedule"
	VALUE_TEST      = "Test"

	CAMPAIGNS = "Campaigns"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&ActionActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &ActionActivity{}, nil
}

type ActionActivity struct {
}

func (*ActionActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval runtime execution
func (a *ActionActivity) Eval(context activity.Context) (done bool, err error) {
	actLogger := context.Logger()
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	conn := input.MailchimpConnection.(*mailchimpConn.MailchimpConnectionManager)
	token := conn.Token

	resource := input.Resource
	if resource == "" {
		return false, fmt.Errorf("%s", "Resource file can't be empty")
	}

	actionName := input.Action
	if actionName == "" {
		return false, fmt.Errorf("%s", "Action name can't be empty")
	}

	inputData := input.Input
	if inputData == nil {
		return false, fmt.Errorf("%s", "Input can't be empty")
	}

	// Marshal map into JSON
	jsonData, err := json.Marshal(inputData)
	if err != nil {
		fmt.Println("Error marshaling map:", err)
		return
	}

	requestInput := &RequestInput{}
	json.Unmarshal([]byte(jsonData), requestInput)

	action := &ApiAction{Data: requestInput.Data, CampaignId: requestInput.CampaiginId, ApiToken: token, Log: actLogger}

	actLogger.Infof("Mailchimp action: %s %s", actionName, resource)
	var resp interface{}

	switch resource {
	case CAMPAIGNS:
		switch actionName {
		case VALUE_SEND:
			resp, err = action.CampaignSend()
			if err != nil {
				return false, fmt.Errorf("Fail to send campaign %s due to, %s", requestInput.CampaiginId, err.Error())
			}
		case VALUE_SCHEDULE:
			resp, err = action.CampaignSchedule()
			if err != nil {
				return false, fmt.Errorf("Fail to schedule campaign %s due to, %s", requestInput.CampaiginId, err.Error())
			}
		case VALUE_TEST:
			resp, err = action.CampaginTest()
			if err != nil {
				return false, fmt.Errorf("Fail to test campaign %s due to, %s", requestInput.CampaiginId, err.Error())
			}
		default:
			return false, fmt.Errorf("Invalid action %s for resource %s", actionName, resource)
		}

	default:
		return false, fmt.Errorf("Invalid resource %s", resource)
	}

	objectResponse := make(map[string]interface{})
	// response for Action apis is empty response with status 204
	objectResponse["success"] = resp
	output := &Output{}
	output.Output = objectResponse
	err = context.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("Error setting output object, %s", err.Error())
	}

	actLogger.Info("Mailchimp action reqeust completed!")
	return true, nil
}

type RequestInput struct {
	CampaiginId string      `json:"campaign_id"`
	Data        interface{} `json:"data, omitempty"`
}

type RequestOutput struct {
	Success bool `json: "success"`
}
