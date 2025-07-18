package query

import (
	ctx "context"
	b64 "encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azfile/directory"
	srv "github.com/Azure/azure-sdk-for-go/sdk/storage/azfile/service"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	commonutil "github.com/tibco/wi-azstorage/src/app/Azurestorage"
	azstorage "github.com/tibco/wi-azstorage/src/app/Azurestorage/connector/connection"
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
var activityLog = log.ChildLogger(log.RootLogger(), "azure-storage-query")

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

type Entry struct {
	Name       string
	Properties interface{}
}

type ShareEntry struct {
	Name       string
	Properties ShareProperty
}

type ShareProperty struct {
	LastModified string `json:"Last-Modified"`
	ETag         string
	Quota        string
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
	objectResponse := make(map[string]interface{})
	serviceResponse := make(map[string]interface{})
	msgResponse := make(map[string]interface{})

	bgConext := ctx.Background()

	if service == "File" {
		shareClient, err := mcon.GetShareClient(paramMap)
		if err != nil {
			return false, fmt.Errorf("Failed to create share client: %s", err.Error())
		}

		switch operation {
		case "Get File":
			dirClient := shareClient.NewDirectoryClient(paramMap["directoryPath"])
			stream, err := dirClient.NewFileClient(paramMap["fileName"]).DownloadStream(bgConext, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to read file: %s", err.Error())
			}
			data, err := io.ReadAll(stream.Body)
			if err != nil {
				return false, fmt.Errorf("Failed to read data from file: %s", err.Error())
			}
			msgResponse["statusCode"] = 200
			msgResponse["fileContent"] = b64.StdEncoding.EncodeToString(data)
			msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."
			msgResponse["isSuccess"] = true
			msgResponse["shareName"] = paramMap["shareName"]
			msgResponse["directoryPath"] = paramMap["directoryPath"]
			msgResponse["fileName"] = paramMap["fileName"]
			objectResponse[service] = msgResponse
		case "List Directories and Files":
			dirClient := shareClient.NewDirectoryClient(paramMap["directoryPath"])

			marker := paramMap["nextmarker"]

			prefix := paramMap["prefix"]

			lstDirOpts := &directory.ListFilesAndDirectoriesOptions{
				Marker: &marker,
				Prefix: &prefix,
			}

			if paramMap["maxresults"] != "" {
				maxResults, err := strconv.Atoi(paramMap["maxresults"])
				if err != nil {
					return false, fmt.Errorf("maxresults must be a number: %s", err.Error())
				}
				maxResults32 := int32(maxResults)
				lstDirOpts.MaxResults = &maxResults32
			}

			// List directories and files
			pager := dirClient.NewListFilesAndDirectoriesPager(lstDirOpts)
			serviceResponse["EnumerationResults"] = make(map[string]interface{})
			enumerationResults := serviceResponse["EnumerationResults"].(map[string]interface{})
			enumerationResults["Entries"] = make(map[string]interface{})

			directories := make([]Entry, 0)

			files := make([]Entry, 0)
			if pager.More() {
				resp, err := pager.NextPage(bgConext)
				if err != nil {
					return false, fmt.Errorf("Failed to get next page: %s", err.Error())
				}
				for _, dir := range resp.Segment.Directories {
					directories = append(directories, Entry{
						Name:       *dir.Name,
						Properties: dir.Properties,
					})
				}
				for _, file := range resp.Segment.Files {
					files = append(files, Entry{
						Name:       *file.Name,
						Properties: file.Properties,
					})
				}
				enumerationResults["NextMarker"] = *resp.NextMarker
			}
			enumerationResults["Entries"].(map[string]interface{})["Directory"] = directories
			enumerationResults["Entries"].(map[string]interface{})["File"] = files
			objectResponse[service] = serviceResponse
		case "List Shares":
			shareServiceClient, err := mcon.GetShareServiceClient()
			if err != nil {
				return false, fmt.Errorf("Failed to create share service client: %s", err.Error())
			}
			marker := paramMap["nextmarker"]

			prefix := paramMap["prefix"]

			lstShareOpts := &srv.ListSharesOptions{
				Marker: &marker,
				Prefix: &prefix,
			}

			if paramMap["maxresults"] != "" {
				maxResults, err := strconv.Atoi(paramMap["maxresults"])
				if err != nil {
					return false, fmt.Errorf("maxresults must be a number: %s", err.Error())
				}
				maxResults32 := int32(maxResults)
				lstShareOpts.MaxResults = &maxResults32
			}

			pager := shareServiceClient.NewListSharesPager(lstShareOpts)

			serviceResponse["EnumerationResults"] = make(map[string]interface{})
			enumerationResults := serviceResponse["EnumerationResults"].(map[string]interface{})
			enumerationResults["Shares"] = make(map[string]interface{})
			shares := make([]ShareEntry, 0)
			if pager.More() {
				resp, err := pager.NextPage(bgConext)
				if err != nil {
					return false, fmt.Errorf("Failed to get next page: %s", err.Error())
				}
				for _, share := range resp.Shares {
					shares = append(shares, ShareEntry{

						Name: *share.Name,
						Properties: ShareProperty{
							//TODO:: check the format of lastmodified
							LastModified: share.Properties.LastModified.String(),
							ETag:         string(*share.Properties.ETag),
							Quota:        fmt.Sprintf("%v", *share.Properties.Quota),
						},
					})
				}
				enumerationResults["NextMarker"] = *resp.NextMarker

			}
			enumerationResults["Shares"] = shares
			objectResponse[service] = serviceResponse
		}
	} else if service == "Blob" {

		switch operation {
		case "Download Blob":
			blobClient, err := mcon.GetBlobClient()
			if err != nil {
				return false, fmt.Errorf("Failed to create blob client: %s", err.Error())
			}
			blobName := paramMap["blobName"]
			if string(paramMap["relativePath"]) != "" {
				blobName = paramMap["relativePath"] + "/" + paramMap["blobName"]
			}
			stream, err := blobClient.DownloadStream(bgConext, paramMap["containerName"], blobName, nil)
			if err != nil {
				return false, fmt.Errorf("Failed to download blob: %s", err.Error())
			}
			data, err := io.ReadAll(stream.Body)
			if err != nil {
				return false, fmt.Errorf("Failed to read data from blob: %s", err.Error())
			}
			msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."
			msgResponse["containerName"] = paramMap["containerName"]
			msgResponse["blobName"] = paramMap["blobName"]
			msgResponse["blobContent"] = b64.StdEncoding.EncodeToString(data)
		case "List Blobs":
			blobClient, err := mcon.GetBlobClient()
			if err != nil {
				return false, fmt.Errorf("Failed to create blob client: %s", err)
			}
			pager := blobClient.NewListBlobsFlatPager(paramMap["containerName"], nil)
			blobs := make([]container.BlobItem, 0)
			if pager.More() {
				resp, err := pager.NextPage(bgConext)
				if err != nil {
					return false, fmt.Errorf("Failed to get next page: %s", err.Error())
				}
				for _, blob := range resp.Segment.BlobItems {
					blobs = append(blobs, *blob)
				}
			}
			msgResponse["statusMessage"] = "Operation " + operation + " successfully completed."
			msgResponse["containerName"] = paramMap["containerName"]
			msgResponse["listofBlobs"] = blobs
		}
		objectResponse[service] = msgResponse
	}
	return objectResponse, nil
}
