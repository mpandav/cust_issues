package create

import (
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/activity"
	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"
)

const (
	CONNECTION = "Connection"
	RESOURCE   = "Resource"
	LIST_ID    = "ListId"
	INPUT      = "input"

	RESOURCE_LIST   = "List"
	RESOURCE_MEMBER = "Member"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&CreateActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &CreateActivity{}, nil
}

type CreateActivity struct {
}

func (*CreateActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval runtime execution
func (a *CreateActivity) Eval(context activity.Context) (done bool, err error) {
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
	if inputData == nil {
		return false, fmt.Errorf("%s", "Input data can't be empty")
	}

	create := &ApiCreate{ActivityInput: inputData, ApiToken: token, Log: actLogger}

	var apiRsp interface{}
	switch resource {
	case RESOURCE_LIST:
		apiRsp, err = create.List()
		if err != nil {
			return false, fmt.Errorf("Fail to create list, %s", err.Error())
		}
	case RESOURCE_MEMBER:
		listId := input.ListId
		if listId == "" {
			return false, fmt.Errorf("List id can't be empty for member creation")
		}
		apiRsp, err = create.Member(listId)
		if err != nil {
			return false, fmt.Errorf("Fail to create member in list %s due to, %s", listId, err.Error())
		}
	default:
		return false, fmt.Errorf("Resource %s is not supported", resource)
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

	actLogger.Info("Create operation completed!")
	return true, nil
}
