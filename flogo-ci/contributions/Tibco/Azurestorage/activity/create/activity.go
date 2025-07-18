package create

import (
	"bytes"
	ctx "context"
	b64 "encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
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

var activityMd = activity.ToMetadata(&Input{}, &Output{})
var activityLog = log.ChildLogger(log.RootLogger(), "azure-storage-create")

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
	operation    string
	service      string
	typeofUpload string
	path         string
}

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
		case "Create Share":
			_, err := shareClient.Create(bgConext, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to create share: %s", err.Error())
			}
		case "Create Directory":
			dirName := paramMap["directoryName"]
			if !(paramMap["directoryPath"] == "" || len(paramMap["directoryPath"]) < 1) {
				dirName = paramMap["directoryPath"] + "/" + dirName
			}
			_, err := shareClient.NewDirectoryClient(dirName).Create(bgConext, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to create directory: %s", err.Error())
			}
			msgResponse["directoryPath"] = paramMap["directoryPath"]
			msgResponse["directoryName"] = paramMap["directoryName"]
		case "Create File":
			data, _ := b64.StdEncoding.DecodeString(paramMap["fileContent"])

			dirClient := shareClient.NewDirectoryClient(paramMap["directoryPath"])
			fileClient := dirClient.NewFileClient(paramMap["fileName"])
			_, err = fileClient.Create(bgConext, int64(len(data)), nil)
			if err != nil {
				return false, fmt.Errorf("Failed to create file: %s", err.Error())
			}
			_, err = fileClient.UploadRange(bgConext, 0, commonutil.NewReadSeekCloserFromBytes(data), nil)
			if err != nil {
				return false, fmt.Errorf("Failed to write content : %s", err.Error())
			}
			msgResponse["directoryPath"] = paramMap["directoryPath"]
			msgResponse["fileName"] = paramMap["fileName"]
		}
		msgResponse["isSuccess"] = true
		msgResponse["shareName"] = paramMap["shareName"]
	} else if service == "Blob" {
		data, _ := b64.StdEncoding.DecodeString(paramMap["blobContent"])
		blobClient, err := mcon.GetBlobClient()
		if err != nil {
			return false, fmt.Errorf("Failed to create blob client: %s", err.Error())
		}
		if string(paramMap["relativePath"]) != "" {
			paramMap["blobName"] = paramMap["relativePath"] + "/" + paramMap["blobName"]
		}
		//TODO:: Confirm the bufferSize and maxBuffers
		bufferSize := (int64)(2 * 1024 * 1024) // Configure the size of the rotating buffers that are used when uploading
		maxBuffers := 3                        // Configure the number of rotating buffers that are used when uploading
		_, err = blobClient.UploadStream(bgConext, paramMap["containerName"], paramMap["blobName"], bytes.NewReader(data), &azblob.UploadStreamOptions{BlockSize: bufferSize, Concurrency: maxBuffers})
		if err != nil {
			return false, fmt.Errorf("Failed to upload blob: %s", err.Error())
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
