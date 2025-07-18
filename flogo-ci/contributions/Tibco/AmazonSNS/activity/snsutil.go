package snsutil

import (
	"encoding/json"
	"fmt"

	"github.com/project-flogo/core/data/coerce"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

// DoPublishPlainText ...
func DoPublishPlainText(context activity.Context, snsSvc *sns.SNS, inputObj map[string]interface{}, log log.Logger) (*sns.PublishOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, err
	}
	request := &sns.PublishInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, err
	}
	log.Debugf("Input for activity [%s]: %s", context.Name(), request.GoString())
	err = request.Validate()
	if err != nil {
		return nil, fmt.Errorf("Error validating input data: %s", err.Error())
	}
	result, err := snsSvc.Publish(request)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// DoPublishCustom ...
func DoPublishCustom(context activity.Context, snsSvc *sns.SNS, inputObj map[string]interface{}, log log.Logger) (*sns.PublishOutput, error) {
	// STEP 1: Find message and coerce it into object
	msg, _ := coerce.ToObject(inputObj["Message"])
	// STEP 2: Find GCM like keys in message and convert them to string to get stringified msg
	for key, val := range msg {
		if _, ok := val.(string); !ok {
			stringifiedVal, err := stringify(val)
			if err != nil {
				return nil, err
			}
			msg[key] = stringifiedVal
		}
	}
	stringifiedMsg, err := stringify(msg)
	if err != nil {
		return nil, err
	}

	// Step 3: Remove the Message key from input, then map the input object to sns.PublishInput{}
	inputObj["Message"] = ""
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, err
	}
	request := &sns.PublishInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, err
	}

	// STEP 4: Now set the Message key in sns.PublishInput{}
	request.Message = &stringifiedMsg
	log.Debugf("Input for activity [%s]: %s", context.Name(), request.GoString())
	err = request.Validate()
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return nil, fmt.Errorf("Error validating input data: %s", err.Error())
	}

	result, err := snsSvc.Publish(request)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func stringify(m interface{}) (string, error) {
	mBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(mBytes), nil
}
