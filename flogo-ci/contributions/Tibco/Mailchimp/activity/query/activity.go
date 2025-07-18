package query

import (
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/activity"
	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"
)

const (
	//fields
	CONNECTION = "Connection"
	RESOURCE   = "Resource"
	INPUT      = "input"

	//options
	CAMPAIGNS = "Campaigns"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&QueryActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &QueryActivity{}, nil
}

type QueryActivity struct {
}

func (*QueryActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval runtime execution
func (a *QueryActivity) Eval(context activity.Context) (done bool, err error) {
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

	inputData := input.Input

	query := &ApiQuery{ActivityInput: inputData, ApiToken: token, Log: actLogger}

	var apiRsp interface{}
	switch resource {
	case CAMPAIGNS:
		apiRsp, err = query.Campaigns()
		if err != nil {
			return false, fmt.Errorf("Fail to query campaigns, %s", err.Error())
		}
		break
	default:
	}

	actLogger.Debugf("==>Activity output: %s", string(apiRsp.([]byte)))

	objectResponse := make(map[string]interface{})
	json.Unmarshal(apiRsp.([]byte), &objectResponse)
	output := &Output{}
	output.Output = objectResponse
	err = context.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("Error setting output object, %s", err.Error())
	}
	actLogger.Info("Query operation completed!")
	return true, nil
}
