package flogofile

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-flogo/core/support/log"
)

const FLOGO_FILES_ROOT_DIR = "FLOGO_FILES_ROOT_DIR"

type (
	// InputParam is a representation of activity's input parameters
	InputParam struct {
		FileName      string `json:"fileName"`
		Overwrite     bool   `json:"overwrite,omitempty"`
		TextContent   string `json:"textContent,omitempty"`
		BinaryContent string `json:"binaryContent,omitempty"`
		FromFileName  string `json:"fromFileName,omitempty"`
		ToFileName    string `json:"toFileName,omitempty"`
	}

	Output struct {
		FileMetadata *FileMetadata `json:"fileMetadata"`
	}

	FileMetadata struct {
		FullPath string `json:"fullPath"`
		Name     string `json:"name"`
		Size     int64  `json:"size"`
		Mode     string `json:"mode"`
		ModTime  string `json:"modTime"`
		IsDir    bool   `json:"isDir"`
	}

	// ListOutput is an agrregation of FileMetadata
	ListOutput struct {
		FileMetadata []*FileMetadata `json:"fileMetadata"`
	}

	ReadFileOutput struct {
		FileMetadata *FileMetadata `json:"fileMetadata"`
		FileContent  *FileContent  `json:"fileContent"`
	}

	FileContent struct {
		TextContent   string `json:"textContent,omitempty"`
		BinaryContent string `json:"binaryContent,omitempty"`
	}
)

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
	log.Debugf(GetMessage(ActivityInput, string(dataBytes)))

	err = json.Unmarshal(dataBytes, inputParams)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(Unmarshall, err.Error()))
	}

	return inputParams, nil
}

// CreateFile creates the file
func CreateFile(inputParams *InputParam, createNonExistingDir bool, activityName string) (*Output, error) {
	fileName := inputParams.FileName
	overwrite := inputParams.Overwrite

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(CreateFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(CreateFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) && !overwrite {
		//throw error
		err := fmt.Errorf("[%s] already exists", fileName)
		return nil, GetError(CreateFileError, activityName, err.Error())
	}

	if createNonExistingDir {
		dirpath := filepath.Dir(fileName)
		if _, err := os.Stat(dirpath); err != nil {
			err := os.MkdirAll(dirpath, os.ModePerm)
			if err != nil {
				return nil, GetError(CreateFileError, activityName, err.Error())
			}
		}
	}

	dstFile, err := os.Create(fileName)
	if err != nil {
		return nil, GetError(CreateFileError, activityName, err.Error())
	}
	defer dstFile.Close()

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var output Output
	output.FileMetadata = &fileMetadata

	return &output, nil
}

// CreateDir creates the directory
func CreateDir(inputParams *InputParam, createNonExistingDir bool, activityName string) (*Output, error) {
	fileName := inputParams.FileName
	overwrite := inputParams.Overwrite

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(CreateFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(CreateFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) && !overwrite {
		//throw error when exists but overwrite is not true
		err := fmt.Errorf("[%s] already exists", fileName)
		return nil, GetError(CreateDirError, activityName, err.Error())
	} else if !errors.Is(err, os.ErrNotExist) && overwrite {
		//exists and overwrite is set true then delete existing
		err := os.Remove(fileName)
		if err != nil {
			err := fmt.Errorf("error in deleting the existing directory : %s", err.Error())
			return nil, GetError(CreateDirError, activityName, err.Error())
		}
	}

	if createNonExistingDir {
		err := os.MkdirAll(fileName, os.ModePerm)
		if err != nil {
			return nil, GetError(CreateDirError, activityName, err.Error())
		}
	} else {
		err := os.Mkdir(fileName, os.ModePerm)
		if err != nil {
			return nil, GetError(CreateDirError, activityName, err.Error())
		}
	}

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var output Output
	output.FileMetadata = &fileMetadata

	return &output, nil
}

// RemoveFile removes the file or directory
func RemoveFile(inputParams *InputParam, removeRecursive bool, activityName string) (*Output, error) {
	fileName := inputParams.FileName

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(RemoveFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(RemoveFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	//get file information before deleting the file
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var output Output
	output.FileMetadata = &fileMetadata

	if removeRecursive {
		//remove recursively
		err = os.RemoveAll(fileName)
	} else {
		err = os.Remove(fileName)
	}
	if err != nil {
		return nil, GetError(RemoveFileError, activityName, err.Error())
	}

	return &output, nil
}

// RenameFile renames the file or directory
func RenameFile(inputParams *InputParam, createNonExistingDir bool, activityName string) (*Output, error) {
	fromFileName := inputParams.FromFileName
	toFileName := inputParams.ToFileName
	overwrite := inputParams.Overwrite

	if strings.Contains(fromFileName, "..") || strings.Contains(toFileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fromFileName\", \"toFileName\"")
		return nil, GetError(RenameFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(RenameFileError, activityName, err.Error())
		}
		if !strings.HasPrefix(fromFileName, "/") {
			fromFileName = filepath.Join(rootDir, fromFileName)
		}
		if !strings.HasPrefix(toFileName, "/") {
			toFileName = filepath.Join(rootDir, toFileName)
		}
	}

	if _, err := os.Stat(toFileName); !errors.Is(err, os.ErrNotExist) && !overwrite {
		//throw error
		err := fmt.Errorf("[%s] already exists", toFileName)
		return nil, GetError(RenameFileError, activityName, err.Error())
	}

	if createNonExistingDir {
		dirpath := filepath.Dir(toFileName)
		if _, err := os.Stat(dirpath); err != nil {
			err := os.MkdirAll(dirpath, os.ModePerm)
			if err != nil {
				return nil, GetError(RenameFileError, activityName, err.Error())
			}
		}
	}

	err := os.Rename(fromFileName, toFileName)
	if err != nil {
		return nil, GetError(RenameFileError, activityName, err.Error())
	}

	//get metadata of new file
	fileInfo, err := os.Stat(toFileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(toFileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var output Output
	output.FileMetadata = &fileMetadata

	return &output, nil
}

// CopyFile copies the file or directory
func CopyFile(inputParams *InputParam, createNonExistingDir bool, includeSubDirectories bool, activityName string, logger log.Logger) error {
	fromFileName := inputParams.FromFileName
	toFileName := inputParams.ToFileName
	overwrite := inputParams.Overwrite

	if strings.Contains(fromFileName, "..") || strings.Contains(toFileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fromFileName\", \"toFileName\"")
		return GetError(CopyFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return GetError(CopyFileError, activityName, err.Error())
		}
		if !strings.HasPrefix(fromFileName, "/") {
			fromFileName = filepath.Join(rootDir, fromFileName)
		}
		if !strings.HasPrefix(toFileName, "/") {
			toFileName = filepath.Join(rootDir, toFileName)
		}
	}

	matches, err := filepath.Glob(fromFileName)
	if err != nil {
		msg := fmt.Sprintf("error in matching the pattern : %s", err.Error())
		return GetError(DefaultError, activityName, msg)
	}

	if len(matches) == 0 {
		logger.Warnf("No match found for Input : %s", fromFileName)
		return nil
	}

	for _, srcFileName := range matches {
		err := checkRequirements(srcFileName, toFileName, createNonExistingDir, len(matches), activityName, fromFileName)
		if err != nil {
			return GetError(CopyFileError, activityName, err.Error())
		}

		file_name := filepath.Base(srcFileName)

		var destFileName string
		if fileInfo, err := os.Stat(toFileName); errors.Is(err, os.ErrNotExist) {
			// entered toFileName is a non existing file, hence create with this name
			destFileName = toFileName
		} else {
			if fileInfo.IsDir() {
				// entered toFileName is a directory, copy filename same as fromFileName
				destFileName = filepath.Join(toFileName, file_name)

				//check if exists and decide based on overwrite flag
				if destFileInfo, err := os.Stat(destFileName); !errors.Is(err, os.ErrNotExist) && !overwrite {
					//throw error
					err := fmt.Errorf("[%s] already exists", destFileName)
					return GetError(CopyFileError, activityName, err.Error())
				} else if !errors.Is(err, os.ErrNotExist) && overwrite && destFileInfo.IsDir() {
					//directory exists and overwrite is set true then delete manually
					err := os.Remove(destFileName)
					if err != nil {
						err := fmt.Errorf("error in deleting the existing directory : %s", err.Error())
						return GetError(CopyFileError, activityName, err.Error())
					}
				}
			} else {
				// entered toFileName is a file
				destFileName = toFileName
				//check if exists and decide based on overwrite flag
				if fileInfo, err := os.Stat(destFileName); !errors.Is(err, os.ErrNotExist) && !overwrite && !fileInfo.IsDir() {
					//throw error
					err := fmt.Errorf("[%s] already exists", destFileName)
					return GetError(CopyFileError, activityName, err.Error())
				}
			}
		}

		//fmt.Println("Destination file name : ", destFileName)
		srcFileInfo, err := os.Stat(srcFileName)
		if err != nil {
			return GetError(FailedToGetFileInfo, activityName, err.Error())
		}

		if !srcFileInfo.IsDir() {
			//source is file
			err := copyFileOperation(srcFileName, destFileName, activityName)
			if err != nil {
				return GetError(CopyFileError, activityName, err.Error())
			}
		} else {
			err := copyDirOperation(srcFileName, destFileName, includeSubDirectories, activityName)
			if err != nil {
				return GetError(CopyFileError, activityName, err.Error())
			}
		}
	}

	return nil
}

func copyFileOperation(srcFileName string, destFileName string, activityName string) error {
	srcFile, err := os.Open(srcFileName)
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}
	defer srcFile.Close()

	destFile, err := os.Create(destFileName) // creates if file doesn't exist
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}

	err = destFile.Sync()
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}

	return nil
}

func copyDirOperation(srcFileName string, destFileName string, includeSubDirectories bool, activityName string) error {
	srcFileInfo, err := os.Stat(srcFileName)
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}

	err = os.Mkdir(destFileName, srcFileInfo.Mode())
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}

	srcChildren, err := os.ReadDir(srcFileName)
	if err != nil {
		return GetError(DefaultError, activityName, err.Error())
	}

	for _, srcChild := range srcChildren {
		srcPath := filepath.Join(srcFileName, srcChild.Name())
		destPath := filepath.Join(destFileName, srcChild.Name())
		if srcChild.IsDir() {
			if !includeSubDirectories {
				//if not recursive copy, skip directories
				continue
			}
			err := copyDirOperation(srcPath, destPath, includeSubDirectories, activityName)
			if err != nil {
				return GetError(DefaultError, activityName, err.Error())
			}
		} else {
			err := copyFileOperation(srcPath, destPath, activityName)
			if err != nil {
				return GetError(DefaultError, activityName, err.Error())
			}
		}
	}

	return nil
}

func checkRequirements(srcFileName string, destFileName string, createNonExistingDir bool, matchCount int, activityName string, fromFileName string) error {
	if matchCount == 1 {
		srcFileInfo, err := os.Stat(srcFileName)
		if err != nil {
			return GetError(FailedToGetFileInfo, activityName, err.Error())
		}

		if destFileInfo, err := os.Stat(destFileName); errors.Is(err, os.ErrNotExist) {
			err := createIntermediateDirectories(destFileName, createNonExistingDir, activityName)
			if err != nil {
				return GetError(DefaultError, activityName, err.Error())
			}
			if srcFileInfo.IsDir() {
				err := os.Mkdir(destFileName, os.ModePerm)
				if err != nil {
					return GetError(DefaultError, activityName, err.Error())
				}
			}
		} else {
			// deal with single source problem
			// cannot copy a directory to a file (file exists)
			if srcFileInfo.IsDir() && !destFileInfo.IsDir() {
				msg := fmt.Sprintf("Cannot copy a directory [%s] to an existing file [%s]", srcFileName, destFileName)
				return GetError(DefaultError, activityName, msg)
			}
		}
	} else {
		// can't copy more than 1 file to a single file
		if destFileInfo, err := os.Stat(destFileName); !errors.Is(err, os.ErrNotExist) {
			//destination exists but is a file then throw error
			if !destFileInfo.IsDir() {
				msg := fmt.Sprintf("Cannot copy multiple source files [%s] to an existing file [%s]", fromFileName, destFileName)
				return GetError(DefaultError, activityName, msg)
			}
		} else {
			// create a directory out of the last part
			destFileName = filepath.Join(destFileName, filepath.Base(destFileName))

			//destination does not exist then create non-existing directories based on flag
			err := createIntermediateDirectories(destFileName, createNonExistingDir, activityName)
			if err != nil {
				return GetError(DefaultError, activityName, err.Error())
			}
		}
	}

	return nil
}

func createIntermediateDirectories(fileName string, createNonExistingDir bool, activityName string) error {
	dirpath := filepath.Dir(fileName)
	if _, err := os.Stat(dirpath); errors.Is(err, os.ErrNotExist) {
		if !createNonExistingDir {
			msg := fmt.Sprintf("Cannot create missing path [%s]", dirpath)
			return GetError(DefaultError, activityName, msg)
		} else {
			err := os.MkdirAll(dirpath, os.ModePerm)
			if err != nil {
				return GetError(DefaultError, activityName, err.Error())
			}
		}
	}
	return nil
}

// ReadTextFile read the text content
// Maintaining this as separate code as there might be more encoding support for text content in future
func ReadTextFile(inputParams *InputParam, compress string, activityName string) (*ReadFileOutput, error) {
	fileName := inputParams.FileName

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(ReadFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(ReadFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, GetError(ReadFileError, activityName, err.Error())
	}

	if compress == "GUnZip" {
		bReader := bytes.NewReader(data)
		gReader, err := gzip.NewReader(bReader)
		if err != nil {
			return nil, GetError(ReadFileError, activityName, err.Error())
		}
		defer gReader.Close()

		data, err = io.ReadAll(gReader)
		if err != nil {
			return nil, GetError(ReadFileError, activityName, err.Error())
		}
	}

	content := string(data) //text content

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var fileContent FileContent
	fileContent.TextContent = content

	var readFileOutput ReadFileOutput
	readFileOutput.FileMetadata = &fileMetadata
	readFileOutput.FileContent = &fileContent

	return &readFileOutput, nil
}

// ReadBinaryFile read the binary content
func ReadBinaryFile(inputParams *InputParam, compress string, activityName string) (*ReadFileOutput, error) {
	fileName := inputParams.FileName

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(ReadFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(ReadFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	//read content of file (with default encoding UTF-8)
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, GetError(ReadFileError, activityName, err.Error())
	}

	if compress == "GUnZip" {
		bReader := bytes.NewReader(data)
		gReader, err := gzip.NewReader(bReader)
		if err != nil {
			return nil, GetError(ReadFileError, activityName, err.Error())
		}
		defer gReader.Close()

		data, err = io.ReadAll(gReader)
		if err != nil {
			return nil, GetError(ReadFileError, activityName, err.Error())
		}
	}

	content := base64.StdEncoding.EncodeToString(data) //binary content

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var fileContent FileContent
	fileContent.BinaryContent = content

	var readFileOutput ReadFileOutput
	readFileOutput.FileMetadata = &fileMetadata
	readFileOutput.FileContent = &fileContent

	return &readFileOutput, nil
}

// WriteTextFile writes the text content to a file
// Maintaining this as separate code as there might be more encoding support for text content in future
func WriteTextFile(inputParams *InputParam, createNonExistingDir bool, compress string, activityName string) (*Output, error) {
	fileName := inputParams.FileName
	overwrite := inputParams.Overwrite
	textContent := inputParams.TextContent

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(WriteFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(WriteFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	if createNonExistingDir {
		dirpath := filepath.Dir(fileName)
		if _, err := os.Stat(dirpath); err != nil {
			err := os.MkdirAll(dirpath, os.ModePerm)
			if err != nil {
				return nil, GetError(WriteFileError, activityName, err.Error())
			}
		}
	}

	if compress == "GZip" {
		var inputBuffer bytes.Buffer
		gWriter := gzip.NewWriter(&inputBuffer)
		_, err := gWriter.Write([]byte(textContent))
		if err != nil {
			return nil, GetError(WriteFileError, activityName, err.Error())
		}
		if err := gWriter.Close(); err != nil {
			return nil, GetError(WriteFileError, activityName, err.Error())
		}
		textContent = inputBuffer.String()
	}

	var file *os.File
	var err error
	if overwrite {
		file, err = os.Create(fileName)
	} else {
		file, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return nil, GetError(WriteFileError, activityName, err.Error())
	}
	defer file.Close()

	_, err = file.Write([]byte(textContent))
	if err != nil {
		return nil, GetError(WriteFileError, activityName, err.Error())
	}

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var output Output
	output.FileMetadata = &fileMetadata

	return &output, nil
}

// WriteBinaryFile writes the binary content to a file
func WriteBinaryFile(inputParams *InputParam, createNonExistingDir bool, compress string, activityName string) (*Output, error) {
	fileName := inputParams.FileName
	overwrite := inputParams.Overwrite
	binaryContent := inputParams.BinaryContent

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(WriteFileError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(WriteFileError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	if createNonExistingDir {
		dirpath := filepath.Dir(fileName)
		if _, err := os.Stat(dirpath); err != nil {
			err := os.MkdirAll(dirpath, os.ModePerm)
			if err != nil {
				return nil, GetError(WriteFileError, activityName, err.Error())
			}
		}
	}

	var file *os.File
	var err error
	if overwrite {
		file, err = os.Create(fileName)
	} else {
		file, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return nil, GetError(WriteFileError, activityName, err.Error())
	}
	defer file.Close()

	data, err := base64.StdEncoding.DecodeString(binaryContent)
	if err != nil {
		return nil, GetError(WriteFileError, activityName, err.Error())
	}

	if compress == "GZip" {
		var inputBuffer bytes.Buffer
		gWriter := gzip.NewWriter(&inputBuffer)
		_, err := gWriter.Write(data)
		if err != nil {
			return nil, GetError(WriteFileError, activityName, err.Error())
		}
		if err := gWriter.Close(); err != nil {
			return nil, GetError(WriteFileError, activityName, err.Error())
		}
		data = inputBuffer.Bytes()
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, GetError(WriteFileError, activityName, err.Error())
	}

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
	}

	var fileMetadata FileMetadata
	absolutePath, err := filepath.Abs(fileName)
	if err == nil {
		fileMetadata.FullPath = absolutePath
	}
	fileMetadata.Name = fileInfo.Name()
	fileMetadata.Size = fileInfo.Size()
	fileMetadata.Mode = fileInfo.Mode().String()
	fileMetadata.ModTime = fileInfo.ModTime().String()
	fileMetadata.IsDir = fileInfo.IsDir()

	var output Output
	output.FileMetadata = &fileMetadata

	return &output, nil
}

// ListFiles lists all the files or directories in the specified directory
func ListFiles(inputParams *InputParam, mode string, activityName string) (*ListOutput, error) {
	fileName := inputParams.FileName

	if strings.Contains(fileName, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"fileName\"")
		return nil, GetError(DefaultError, activityName, errMessage)
	}

	rootDir := os.Getenv(FLOGO_FILES_ROOT_DIR)
	if rootDir != "" && !strings.HasPrefix(fileName, "/") {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return nil, GetError(DefaultError, activityName, err.Error())
		}
		fileName = filepath.Join(rootDir, fileName)
	}

	matches, err := filepath.Glob(fileName)
	if err != nil {
		msg := fmt.Sprintf("error in matching the pattern : %s", err.Error())
		return nil, GetError(DefaultError, activityName, msg)
	}

	var listOutput ListOutput

	for _, fileName := range matches {
		fileInfo, err := os.Stat(fileName)
		if err != nil {
			return nil, GetError(FailedToGetFileInfo, activityName, err.Error())
		}

		var fileMetadata FileMetadata
		absolutePath, err := filepath.Abs(fileName)
		if err == nil {
			fileMetadata.FullPath = absolutePath
		}
		fileMetadata.Name = fileInfo.Name()
		fileMetadata.Size = fileInfo.Size()
		fileMetadata.Mode = fileInfo.Mode().String()
		fileMetadata.ModTime = fileInfo.ModTime().String()
		fileMetadata.IsDir = fileInfo.IsDir()

		if accept(mode, fileMetadata.IsDir) {
			listOutput.FileMetadata = append(listOutput.FileMetadata, &fileMetadata)
		}
	}

	return &listOutput, nil
}

func accept(mode string, isDir bool) bool {
	return (mode == "Only Directories" && isDir) || (mode == "Only Files" && !isDir) || (mode == "Files and Directories")
}
