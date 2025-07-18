package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

var logger = log.ChildLogger(log.RootLogger(), "file.activity.archive")

var activityMd = activity.ToMetadata(&Input{}, &Output{})

type FileArchiveActivity struct {
}

func init() {
	_ = activity.Register(&FileArchiveActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &FileArchiveActivity{}, nil
}

func (*FileArchiveActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (docActivity *FileArchiveActivity) Eval(context activity.Context) (done bool, err error) {
	logger.Info("Executing File Archive Activity")

	input := &Input{}
	output := &Output{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	if input.Source == "" {
		return false, fmt.Errorf("Source is required")
	} else if input.Destination == "" {
		return false, fmt.Errorf("Destination is required")
	} else if input.ArchiveType == "" {
		return false, fmt.Errorf("ArchiveType is required")
	}

	archiveType := input.ArchiveType
	source := input.Source
	destination := input.Destination

	if strings.Contains(source, "..") || strings.Contains(destination, "..") {
		errMessage := fmt.Sprintf("\"..\" is not allowed in the input \"source\", \"destination\"")
		return false, fmt.Errorf(errMessage)
	}

	rootDir := os.Getenv("FLOGO_FILES_ROOT_DIR")
	if rootDir != "" {
		err := os.MkdirAll(rootDir, os.ModePerm)
		if err != nil {
			return false, err
		}
		if !strings.HasPrefix(source, "/") {
			source = filepath.Join(rootDir, source)
		}
		if !strings.HasPrefix(destination, "/") {
			destination = filepath.Join(rootDir, destination)
		}
	}

	destination, err = validateDestination(destination, archiveType)
	if err != nil {
		return false, err
	}

	destPath := filepath.Dir(destination)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		err := os.MkdirAll(destPath, 0755)
		if err != nil {
			return false, err
		}
	}

	f, err := os.Create(destination)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if archiveType == "zip" {

		logger.Info("Creating zip archive")
		//Create zip writer
		w1 := zip.NewWriter(f)

		isDir, err := isDirectory(source)
		if err != nil {
			cleanup(context, output)
			return false, err
		}

		if isDir {
			logger.Debug("Source is a directory")
			err = filepath.Walk(source, func(path string, fInfo os.FileInfo, err error) error {

				// logger.Infof("Finfo name: %s", fInfo.Name())
				// logger.Info("*****************************************")
				// logger.Infof("source: %s", source)
				// logger.Infof("Path: %s", path)
				// logger.Info("*****************************************")

				if fInfo.Name() == filepath.Base(source) {
					return nil
				}

				err = addFileToZip(w1, path, source, fInfo)
				if err != nil {
					cleanup(context, output)
					return err
				}

				return nil
			})
			if err != nil {
				cleanup(context, output)
				return false, err
			}
		} else {
			logger.Debug("Source is a file")
			fInfo, err := os.Stat(source)
			if err != nil {
				cleanup(context, output)
				return false, err
			}
			err = addFileToZip(w1, source, filepath.Dir(source), fInfo)
			if err != nil {
				cleanup(context, output)
				return false, err
			}
		}

		logger.Debug("Closing zip archive")
		err = w1.Close()
		if err != nil {
			cleanup(context, output)
			return false, err
		}

	}

	output.Destination = destination

	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	logger.Info("Activity completed. File/folder successfully archived")
	return true, nil
}

func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist")
	}
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func validateDestination(destination, archiveType string) (string, error) {

	if archiveType == "zip" {
		extension := filepath.Ext(destination)
		if extension == "" {
			destination = strings.Trim(destination, ".")
			destination += ".zip"
		} else if extension != ".zip" {
			return "", fmt.Errorf("destination should have .zip extension")
		}
	}

	_, err := os.Stat(destination)
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("file already exists")
	}

	return destination, nil
}

func cleanup(context activity.Context, output *Output) {
	err := os.RemoveAll(output.Destination)
	if err != nil {
		context.Logger().Errorf("Error cleaning up file: %s", err.Error())
	}
}

func addFileToZip(zipWriter *zip.Writer, source, filePath string, fInfo fs.FileInfo) error {
	logger.Debug("Adding file to zip archive")
	// logger.Infof("Source in addfile funtion: %s", source)
	// logger.Infof("filePath in addfile funtion: %s", filePath)

	fHeader, err := zip.FileInfoHeader(fInfo)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(filePath, source)
	if err != nil {
		return err
	}
	fHeader.Name = relPath
	fHeader.Method = zip.Deflate

	if fInfo.IsDir() {
		dirPath := fmt.Sprintf("%s%c", relPath, os.PathSeparator)
		_, err := zipWriter.Create(dirPath)
		if err != nil {
			return err
		}
		return nil
	}

	fileHandle, err := os.Open(source)
	if err != nil {
		return err
	}

	defer fileHandle.Close()

	// create file in zip archive
	f, err := zipWriter.CreateHeader(fHeader)
	if err != nil {
		return err
	}

	//copy file data to zip archive
	_, err = io.Copy(f, fileHandle)
	if err != nil {
		return err
	}
	return nil
}
