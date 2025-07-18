package get

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

type Input struct {
	DocusignConnection connection.Manager `md:"docusignConnection, required"`
	EnvelopeID         string             `md:"envelopeId"`
	DocumentID         string             `md:"documentId"`
	GetAllDocs         bool               `md:"getAllDocuments"`
}

type Output struct {
	FileContent string                 `md:"fileContent"`
	FileType    string                 `md:"fileType"`
	OutputType  string                 `md:"outputType"`
	Error       map[string]interface{} `md:"error"`
}

func (input *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"docusignConnection": input.DocusignConnection,
		"envelopeId":         input.EnvelopeID,
		"documentId":         input.DocumentID,
		"getAllDocuments":    input.GetAllDocs,
	}
}

func (input *Input) FromMap(values map[string]interface{}) error {
	var err error

	input.DocusignConnection, err = docusignconnection.GetSharedConfiguration(values["docusignConnection"])
	if err != nil {
		return err
	}

	input.EnvelopeID, err = coerce.ToString(values["envelopeId"])
	if err != nil {
		return err
	}

	input.DocumentID, err = coerce.ToString(values["documentId"])
	if err != nil {
		return err
	}

	input.GetAllDocs, err = coerce.ToBool(values["getAllDocuments"])
	if err != nil {
		return err
	}

	return nil
}

func (output *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"fileContent": output.FileContent,
		"fileType":    output.FileType,
		"outputType":  output.OutputType,
		"error":       output.Error,
	}
}

func (output *Output) FromMap(values map[string]interface{}) error {
	var err error

	output.FileContent, err = coerce.ToString(values["fileContent"])
	if err != nil {
		return err
	}

	output.FileType, err = coerce.ToString(values["fileType"])
	if err != nil {
		return err
	}

	output.OutputType, err = coerce.ToString(values["outputType"])
	if err != nil {
		return err
	}

	output.Error, err = coerce.ToObject(values["error"])
	if err != nil {
		return err
	}

	return nil
}
