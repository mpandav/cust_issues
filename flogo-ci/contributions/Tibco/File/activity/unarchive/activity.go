package unarchive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-flogo/core/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

type FileUnarchiveActivity struct {
}

func init() {
	_ = activity.Register(&FileUnarchiveActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &FileUnarchiveActivity{}, nil
}

func (*FileUnarchiveActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (docActivity *FileUnarchiveActivity) Eval(context activity.Context) (done bool, err error) {
	logger := context.Logger()
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

	logger.Infof("unzipping at location : %s", destination)

	if archiveType == "zip" {
		r, err := zip.OpenReader(source)
		if err != nil {
			return false, err
		}
		defer r.Close()

		for _, f := range r.File {

			fpath := filepath.Join(destination, f.FileHeader.Name)
			basePath := filepath.Dir(fpath)

			if f.FileInfo().IsDir() {
				basePath = fpath
				err = createDestinationDirectory(context, basePath)
				if err != nil {
					return false, err
				}
				continue
			} else {
				err = createDestinationDirectory(context, basePath)
				if err != nil {
					return false, err
				}
			}

			//create file in current directory
			file, err := os.Create(fpath)
			if err != nil {
				cleanup(context, output)
				return false, err
			}
			defer file.Close()

			//open file in zip archive
			rc, err := f.Open()
			if err != nil {
				cleanup(context, output)
				return false, err
			}
			defer rc.Close()

			//copy file data to new file
			_, err = io.Copy(file, rc)
			if err != nil {
				cleanup(context, output)
				return false, err
			}
		}

	}

	output.Destination = destination
	err = context.SetOutputObject(output)
	if err != nil {
		cleanup(context, output)
		return false, err
	}

	return true, nil
}

func cleanup(context activity.Context, output *Output) {
	err := os.RemoveAll(output.Destination)
	if err != nil {
		context.Logger().Errorf("Error cleaning up file: %s", err.Error())
	}
}

func createDestinationDirectory(context activity.Context, destination string) error {
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		context.Logger().Debug("Directory does not exist.")
		err = os.MkdirAll(destination, 0755)
		if err != nil {
			context.Logger().Info("Error creating directory: %v\n", err)
			return err
		}
	} else {
		context.Logger().Debug("Directory exists.")
	}

	return nil
}
