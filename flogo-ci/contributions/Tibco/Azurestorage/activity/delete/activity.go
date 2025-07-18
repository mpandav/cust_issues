package delete

import (
	ctx "context"
	"fmt"

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
	return &Activity{}, nil
}

// Activity is a stub for your Activity implementation
type Activity struct {
	operation string
	service   string
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})
var activityLog = log.ChildLogger(log.RootLogger(), "azure-storage-delete")

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
	activityLog.Debugf("Execution of Activity [%s] " + context.Name() + " completed")
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

	msgResponse := make(map[string]interface{})
	bgConext := ctx.Background()

	if service == "File" {
		shareClient, err := mcon.GetShareClient(paramMap)
		if err != nil {
			return false, fmt.Errorf("Failed to create share client: %s", err.Error())
		}
		switch operation {
		case "Delete Share":
			_, err := shareClient.Delete(bgConext, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to delete share: %s", err.Error())
			}
		case "Delete Directory":
			dirName := paramMap["directoryName"]
			if !(paramMap["directoryPath"] == "" || len(paramMap["directoryPath"]) < 1) {
				dirName = paramMap["directoryPath"] + "/" + dirName
			}
			_, err := shareClient.NewDirectoryClient(dirName).Delete(bgConext, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to delete directory: %s", err.Error())
			}
			msgResponse["directoryPath"] = paramMap["directoryPath"]
			msgResponse["directoryName"] = paramMap["directoryName"]
		case "Delete File":
			dirClient := shareClient.NewDirectoryClient(paramMap["directoryPath"])
			if err != nil {
				return false, fmt.Errorf("Failed to read directory client: %s", err.Error())
			}
			_, err := dirClient.NewFileClient(paramMap["fileName"]).Delete(bgConext, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to delete file: %s", err.Error())
			}
			msgResponse["directoryPath"] = paramMap["directoryPath"]
			msgResponse["fileName"] = paramMap["fileName"]
		}
		msgResponse["isSuccess"] = true
		msgResponse["shareName"] = paramMap["shareName"]
	} else if service == "Blob" {

		blobClient, err := mcon.GetBlobClient()
		if err != nil {
			return false, fmt.Errorf("Failed to create blob client: %s", err.Error())
		}
		blobName := paramMap["blobName"]
		if string(paramMap["relativePath"]) != "" {
			blobName = paramMap["relativePath"] + "/" + paramMap["blobName"]
		}
		_, err = blobClient.DeleteBlob(bgConext, paramMap["containerName"], blobName, nil)
		if err != nil {
			return false, fmt.Errorf("Failed to delete blob: %s", err.Error())
		}
		msgResponse["containerName"] = paramMap["containerName"]
		msgResponse["blobName"] = paramMap["blobName"]
		msgResponse["isSuccess"] = true
	}
	msgResponse["statusCode"] = 200
	msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."

	objectResponse := make(map[string]interface{})
	objectResponse[service] = msgResponse
	return objectResponse, nil
}
