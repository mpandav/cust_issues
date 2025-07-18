package update

import (
	ctx "context"
	b64 "encoding/base64"
	"fmt"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azfile/file"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	commonutil "github.com/tibco/wi-azstorage/src/app/Azurestorage"
	azstorage "github.com/tibco/wi-azstorage/src/app/Azurestorage/connector/connection"
)

const (
	ivConnection = "Connection"
	ivInput      = "input"
	ivService    = "service"
	ivOperation  = "operation"
	ovOutput     = "output"
	ovError      = "error"
)

// Activity is a stub for your Activity implementation
func init() {
	err := activity.Register(&Activity{}, New)
	if err != nil {
		log.RootLogger().Error(err)
	}
}

// New functioncommon
func New(ctx1 activity.InitContext) (activity.Activity, error) {
	act := &Activity{}
	return act, nil
}

// Activity is a stub for your Activity implementation
type Activity struct {
	operation string
	service   string
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})
var activityLog = log.ChildLogger(log.RootLogger(), "azure-storage-update")

// Metadata implements activity.Activity.Metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Cleanup method
func (a *Activity) Cleanup() error {

	return nil
}

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(context activity.Context) (done bool, err error) {
	activityLog.Debugf("Executing Activity [%s] ", context.Name())
	inputVal := &Input{}
	err = context.GetInputObject(inputVal)
	if err != nil {
		return false, fmt.Errorf("Error while getting input object: %s", err.Error())
	}
	mcon, _ := inputVal.Connection.(*azstorage.AzStorageSharedConfigManager)

	objectResponse, err := execute(inputVal)
	if err != nil {
		//Renew SAS token if expired
		err = commonutil.CheckForRenewal(err, activityLog, mcon)
		if err != nil {
			return false, err
		}
		objectResponse, err = execute(inputVal)
		if err != nil {
			return false, err
		}
	}

	context.SetOutput("output", objectResponse)
	activityLog.Debugf("Execution of Activity [%s]" + context.Name() + " completed")
	return true, nil
}

func execute(inputVal *Input) (interface{}, error) {
	service := inputVal.Service
	if service == "" {
		return false, activity.NewError("service is not configured", "AZURE-STORAGE-1003", nil)
	}

	operation := inputVal.Operation
	if operation == "" {
		return false, activity.NewError("operation is not configured", "AZURE-STORAGE-1004", nil)
	}
	mcon, _ := inputVal.Connection.(*azstorage.AzStorageSharedConfigManager)
	paramMap := make(map[string]string)
	if inputVal != nil {
		inputMap := inputVal.Input
		if inputMap["parameters"] != nil {
			parameters := inputMap["parameters"]
			for k, v := range parameters.(map[string]interface{}) {
				paramMap[k] = fmt.Sprint(v)

			}
		}
	}
	bgContext := ctx.Background()
	objectResponse := make(map[string]interface{})
	msgResponse := make(map[string]interface{})
	shareClient, err := mcon.GetShareClient(paramMap)
	if err != nil {
		return false, fmt.Errorf("Failed to create share client: %s", err.Error())
	}
	dirClient := shareClient.NewDirectoryClient(paramMap["directoryPath"])
	if err != nil {
		return false, fmt.Errorf("Failed to create directory: %s", err.Error())
	}

	var filelen int64 = 0
	fileClient := dirClient.NewFileClient(paramMap["fileName"])
	prop, err := fileClient.GetProperties(bgContext, nil)
	if err == nil {
		filelen = *prop.ContentLength
	}
	startRange := int64(0)
	endRange := int64(0)
	// if either of range is specified use it
	if len(paramMap["startRange"]) > 0 && len(paramMap["endRange"]) > 0 {
		startRange, err = strconv.ParseInt(paramMap["startRange"], 10, 64)
		if err != nil {
			return false, fmt.Errorf("Failed to parse startRange: %s", err.Error())
		}
		endRange, err = strconv.ParseInt(paramMap["endRange"], 10, 64)
		if err != nil {
			return false, fmt.Errorf("Failed to parse endRange: %s", err.Error())
		}
	}

	switch operation {
	case "Clear Content":
		// if range not specified clear entire content
		if len(paramMap["startRange"]) <= 0 || len(paramMap["endRange"]) <= 0 {
			startRange = 0
			endRange = filelen - 1
		}

		httpRange := file.HTTPRange{
			Offset: startRange,
			Count:  endRange - startRange + 1,
		}
		_, err = fileClient.ClearRange(bgContext, httpRange, nil)
		if err != nil {
			return false, fmt.Errorf("Failed to clear range : %s", err.Error())
		}

	case "Write Content":
		data, _ := b64.StdEncoding.DecodeString(paramMap["fileContent"])

		//if range not specified append at the end of file
		if len(paramMap["startRange"]) <= 0 || len(paramMap["endRange"]) <= 0 {
			contentLen := len(data)
			_, err := fileClient.Resize(bgContext, filelen+int64(contentLen), nil)
			if err != nil {
				return false, fmt.Errorf("Error resizing file : ", err)
			}
			startRange = filelen
		}

		_, err = fileClient.UploadRange(bgContext, startRange, commonutil.NewReadSeekCloserFromBytes(data), nil)
		if err != nil {
			return false, fmt.Errorf("Failed to write content : %s", err.Error())
		}

	}
	msgResponse["statusCode"] = 200
	msgResponse["isSuccess"] = true
	msgResponse["shareName"] = paramMap["shareName"]
	msgResponse["directoryPath"] = paramMap["directoryPath"]
	msgResponse["fileName"] = paramMap["fileName"]
	msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."
	objectResponse[service] = msgResponse
	return objectResponse, nil
}
