package flogosftp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"github.com/project-flogo/core/support/log"
)

const FLOGO_FILES_ROOT_DIR = "FLOGO_FILES_ROOT_DIR"

type (
	//fileInfo is the information about one file being transferred
	fileInfo struct {
		FileName      string `json:"File Name,omitempty"`
		NumberOfBytes any    `json:"Number of Bytes,omitempty"`
	}

	//Output is an aggregation of files transferred
	Output struct {
		BinaryData      string      `json:"Binary Data,omitempty"`
		ASCIIData       string      `json:"ASCII Data,omitempty"`
		FileTransferred []*fileInfo `json:"FileTransferred"`
	}

	// InputParam is a representation of activity's input parameters
	InputParam struct {
		RemoteFileName    string `json:"Remote File Name,omitempty"`
		LocalFileName     string `json:"Local File Name,omitempty"`
		BinaryData        string `json:"Binary Data,omitempty"`
		ASCIIData         string `json:"ASCII Data,omitempty"`
		RemoteDirectory   string `json:"Remote Directory,omitempty"`
		OldRemoteFileName string `json:"Old Remote File Name,omitempty"`
		NewRemoteFileName string `json:"New Remote File Name,omitempty"`
	}

	//FileMetadata is the metadata of the file
	FileMetadata struct {
		Name    string `json:"Name"`
		Size    int64  `json:"Size"`
		Mode    string `json:"Mode"`
		ModTime string `json:"ModTime"`
		IsDir   bool   `json:"IsDir"`
	}

	// ListOutput is an agrregation of FileMetadata
	ListOutput struct {
		FileMetadata []*FileMetadata `json:"FileMetadata"`
	}

	// OperationOutput is an output for rename, mkdir operations
	OperationOutput struct {
		NewRemoteFileName string `json:"New Remote File Name,omitempty"`
		RemoteDirectory   string `json:"Remote Directory,omitempty"`
	}
)

// FileTransferGetOperation downloads file from SFTP server
func FileTransferGetOperation(sc *sftp.Client, inputParams *InputParam, overwrite bool, activityName string) (*Output, error) {
	localFileName := inputParams.LocalFileName
	if strings.Contains(localFileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"Local File Name\"")
		return nil, GetError(DefaultError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(localFileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(DefaultError, activityName, err.Error())
		}
		localFileName = filepath.Join(rootDir, localFileName)
	}

	srcFile, err := sc.OpenFile(inputParams.RemoteFileName, (os.O_RDONLY))
	if err != nil {
		return nil, GetError(RemoteFileOpenError, activityName, err.Error())
	}
	defer srcFile.Close()

	var dstFile *os.File
	if overwrite {
		dstFile, err = os.Create(localFileName)
	} else {
		dstFile, err = os.OpenFile(localFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return nil, GetError(LocalFileOpenError, activityName, err.Error())
	}
	defer dstFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, GetError(GetOperationError, activityName, err.Error())
	}

	var fileInfo fileInfo
	fileInfo.FileName = inputParams.RemoteFileName
	fileInfo.NumberOfBytes = bytes

	var output Output
	output.FileTransferred = append(output.FileTransferred, &fileInfo)

	return &output, nil
}

// ProcessDataGetOperation downloads base64 binary encoded file content from SFTP server
func ProcessDataGetOperation(sc *sftp.Client, inputParams *InputParam, binary bool, activityName string) (*Output, error) {
	srcFile, err := sc.OpenFile(inputParams.RemoteFileName, (os.O_RDONLY))
	if err != nil {
		return nil, GetError(RemoteFileOpenError, activityName, err.Error())
	}
	defer srcFile.Close()

	// read the content of file hosted on sftp server
	dataBytes := new(bytes.Buffer)
	bytes, err := io.Copy(dataBytes, srcFile)
	//bytes, err := io.ReadAll(srcFile)
	if err != nil {
		return nil, GetError(GetOperationError, activityName, err.Error())
	}
	var fileInfo fileInfo
	fileInfo.FileName = inputParams.RemoteFileName
	fileInfo.NumberOfBytes = bytes

	var output Output
	if binary {
		output.BinaryData = base64.StdEncoding.EncodeToString(dataBytes.Bytes())
	} else {
		output.ASCIIData = dataBytes.String()
	}
	output.FileTransferred = append(output.FileTransferred, &fileInfo)

	return &output, nil
}

// FileTransferPutOperation downloads file from SFTP server
func FileTransferPutOperation(sc *sftp.Client, inputParams *InputParam, overwrite bool, activityName string) (*Output, error) {
	localFileName := inputParams.LocalFileName
	if strings.Contains(localFileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"Local File Name\"")
		return nil, GetError(DefaultError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(localFileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(DefaultError, activityName, err.Error())
		}
		localFileName = filepath.Join(rootDir, localFileName)
	}

	srcFile, err := os.Open(localFileName)
	if err != nil {
		return nil, GetError(LocalFileOpenError, activityName, err.Error())
	}
	defer srcFile.Close()

	// Make remote directories recursion
	parent := filepath.Dir(inputParams.RemoteFileName)
	path := string(filepath.Separator)
	dirs := strings.Split(parent, path)
	for _, dir := range dirs {
		path = filepath.Join(path, dir)
		sc.Mkdir(path)
	}

	var dstFile *sftp.File
	if overwrite {
		dstFile, err = sc.OpenFile(inputParams.RemoteFileName, (os.O_TRUNC | os.O_CREATE | os.O_WRONLY))
	} else {
		dstFile, err = sc.OpenFile(inputParams.RemoteFileName, (os.O_APPEND | os.O_CREATE | os.O_WRONLY))
	}
	if err != nil {
		return nil, GetError(RemoteFileOpenError, activityName, err.Error())
	}
	defer dstFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return nil, GetError(PutOperationError, activityName, err.Error())
	}

	var fileInfo fileInfo
	fileInfo.FileName = inputParams.RemoteFileName
	fileInfo.NumberOfBytes = bytes

	var output Output
	output.FileTransferred = append(output.FileTransferred, &fileInfo)

	return &output, nil
}

// ProcessDataPutOperation downloads base64 binary encoded file content from SFTP server
func ProcessDataPutOperation(sc *sftp.Client, inputParams *InputParam, overwrite bool, binary bool, activityName string) (*Output, error) {
	var data []byte
	var err error
	if binary {
		data, err = base64.StdEncoding.DecodeString(inputParams.BinaryData)
		if err != nil {
			return nil, GetError(DecodeError, activityName, err.Error())
		}
	} else {
		data = []byte(inputParams.ASCIIData)
	}

	// Make remote directories recursion
	parent := filepath.Dir(inputParams.RemoteFileName)
	path := string(filepath.Separator)
	dirs := strings.Split(parent, path)
	for _, dir := range dirs {
		path = filepath.Join(path, dir)
		sc.Mkdir(path)
	}

	var dstFile *sftp.File
	if overwrite {
		dstFile, err = sc.OpenFile(inputParams.RemoteFileName, (os.O_TRUNC | os.O_CREATE | os.O_WRONLY))
	} else {
		dstFile, err = sc.OpenFile(inputParams.RemoteFileName, (os.O_APPEND | os.O_CREATE | os.O_WRONLY))
	}
	if err != nil {
		return nil, GetError(RemoteFileOpenError, activityName, err.Error())
	}
	defer dstFile.Close()

	bytes, err := dstFile.Write(data)
	if err != nil {
		return nil, GetError(PutOperationError, activityName, err.Error())
	}

	var fileInfo fileInfo
	fileInfo.FileName = inputParams.RemoteFileName
	fileInfo.NumberOfBytes = bytes

	var output Output
	output.FileTransferred = append(output.FileTransferred, &fileInfo)

	return &output, nil
}

// GetInputData converts the input to helper.Input
func GetInputData(inputData interface{}, log log.Logger) (inputParams *InputParam, err error) {
	inputParams = &InputParam{}
	if inputData == nil {
		return nil, fmt.Errorf(GetMessage(SpecifyInput))
	}

	//log input at debug level
	dataBytes, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(Deserialize, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, string(dataBytes)))

	err = json.Unmarshal(dataBytes, inputParams)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(Unmarshall, err.Error()))
	}

	return inputParams, nil
}

// DeleteOperation deletes file from SFTP server
func DeleteOperation(sc *sftp.Client, inputParams *InputParam, activityName string, logger log.Logger) error {
	matches, err := sc.Glob(inputParams.RemoteFileName)
	if err != nil {
		msg := fmt.Sprintf("error in matching the pattern : %s", err.Error())
		return GetError(DefaultError, activityName, msg)
	}

	if len(matches) == 0 {
		// sc.Glob returns empty slice even when the connection is lost
		// so we need to check if the connection is alive
		if !IsConnectionAlive(sc) {
			msg := fmt.Sprintf("connection lost")
			return GetError(DefaultError, activityName, msg)
		} else {
			// if connection is alive, and no match found, show a warn message and return nil
			logger.Warnf("No match found for Input : %s", inputParams.RemoteFileName)
			return nil
		}
	}

	var errMsg string
	for _, fileName := range matches {
		err := sc.Remove(fileName)
		if err != nil {
			errMsg = errMsg + fmt.Sprintf("[FileName : %s , Error : %s.]", fileName, err.Error())
		}
	}
	if errMsg != "" {
		return GetError(DeleteOperationError, activityName, errMsg)
	}
	return nil
}

// ListOperation lists all the files in the remote directory
func ListOperation(sc *sftp.Client, inputParams *InputParam, activityName string) (*ListOutput, error) {
	fileInfos, err := sc.ReadDir(inputParams.RemoteDirectory)
	if err != nil {
		return nil, GetError(ListOperationError, activityName, inputParams.RemoteDirectory, err.Error())
	}
	var listOutput ListOutput
	for _, fileInfo := range fileInfos {
		var fileMetadata FileMetadata
		fileMetadata.Name = fileInfo.Name()
		fileMetadata.Size = fileInfo.Size()
		fileMetadata.Mode = fileInfo.Mode().String()
		fileMetadata.ModTime = fileInfo.ModTime().String()
		fileMetadata.IsDir = fileInfo.IsDir()

		listOutput.FileMetadata = append(listOutput.FileMetadata, &fileMetadata)
	}

	return &listOutput, nil
}

// MkdirOperation creates remote directory on the server
func MkdirOperation(sc *sftp.Client, inputParams *InputParam, activityName string) (*OperationOutput, error) {
	err := sc.Mkdir(inputParams.RemoteDirectory)
	if err != nil {
		return nil, GetError(MkdirOperationError, activityName, inputParams.RemoteDirectory, err.Error())
	}

	var output OperationOutput
	output.RemoteDirectory = inputParams.RemoteDirectory

	return &output, nil
}

// RenameOperation creates remote directory on the server
func RenameOperation(sc *sftp.Client, inputParams *InputParam, activityName string) (*OperationOutput, error) {
	err := sc.Rename(inputParams.OldRemoteFileName, inputParams.NewRemoteFileName)
	if err != nil {
		return nil, GetError(RenameOperationError, activityName, inputParams.OldRemoteFileName, err.Error())
	}

	var output OperationOutput
	output.NewRemoteFileName = inputParams.NewRemoteFileName

	return &output, nil
}

// IsConnectionAlive checks if the SFTP connection is alive
// this is a simple check that attempts to list the root directory
func IsConnectionAlive(sc *sftp.Client) bool {
	if sc == nil {
		return false
	}

	_, err := sc.Stat("/")
	return err == nil
}

// IsConnectionLost checks if the error indicates a lost connection
func IsConnectionLost(err error) bool {
	if err == nil {
		return false
	}
	// Check for common connection lost messages
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "connection lost") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection closed") ||
		strings.Contains(errMsg, "use of closed network connection") ||
		strings.Contains(errMsg, "EOF") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "i/o timeout") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "dial tcp") ||
		strings.Contains(errMsg, "handshake failed") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "client not connected")
}
