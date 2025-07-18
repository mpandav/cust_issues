package unarchive

import "github.com/project-flogo/core/data/coerce"

type Input struct {
	ArchiveType string `md:"archiveType,required"`
	Source      string `md:"sourceFilePath,required"`
	Destination string `md:"destinationPath,required"`
}

type Output struct {
	Destination string `md:"extractedPath,required"`
}

func (input *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"archiveType":     input.ArchiveType,
		"sourceFilePath":  input.Source,
		"destinationPath": input.Destination,
	}
}

func (input *Input) FromMap(values map[string]interface{}) error {
	var err error

	input.ArchiveType, err = coerce.ToString(values["archiveType"])
	if err != nil {
		return err
	}

	input.Source, err = coerce.ToString(values["sourceFilePath"])
	if err != nil {
		return err
	}

	input.Destination, err = coerce.ToString(values["destinationPath"])
	if err != nil {
		return err
	}

	return nil
}

func (output *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"extractedPath": output.Destination,
	}
}

func (output *Output) FromMap(values map[string]interface{}) error {
	var err error

	output.Destination, err = coerce.ToString(values["extractedPath"])
	if err != nil {
		return err
	}

	return nil
}
