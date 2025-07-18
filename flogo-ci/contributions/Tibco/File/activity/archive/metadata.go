package archive

import "github.com/project-flogo/core/data/coerce"

type Input struct {
	ArchiveType string `md:"archiveType,required"`
	Source      string `md:"sourcePath,required"`
	Destination string `md:"destinationFilePath,required"`
	// Compress    bool   `md:"compress"`
}

type Output struct {
	Destination string `md:"archivePath,required"`
}

func (input *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"archiveType": input.ArchiveType,
		// "compress":            input.Compress,
		"sourcePath":          input.Source,
		"destinationFilePath": input.Destination,
	}
}

func (input *Input) FromMap(values map[string]interface{}) error {
	var err error

	input.ArchiveType, err = coerce.ToString(values["archiveType"])
	if err != nil {
		return err
	}

	// input.Compress, err = coerce.ToBool(values["compress"])
	// if err != nil {
	// 	return err
	// }

	input.Source, err = coerce.ToString(values["sourcePath"])
	if err != nil {
		return err
	}

	input.Destination, err = coerce.ToString(values["destinationFilePath"])
	if err != nil {
		return err
	}

	return nil
}

func (output *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"archivePath": output.Destination,
	}
}

func (output *Output) FromMap(values map[string]interface{}) error {
	var err error

	output.Destination, err = coerce.ToString(values["archivePath"])
	if err != nil {
		return err
	}

	return nil
}
