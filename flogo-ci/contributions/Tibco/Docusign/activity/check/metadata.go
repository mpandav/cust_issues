package check

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

type Input struct {
	DocusignConnection connection.Manager `md:"docusignConnection, required"`
	EnvelopeID         string             `md:"envelopeId"`
}

type Output struct {
	Status string                 `md:"status"`
	Error  map[string]interface{} `md:"error"`
}

func (input *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"docusignConnection": input.DocusignConnection,
		"envelopeId":         input.EnvelopeID,
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

	return nil
}

func (output *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"status": output.Status,
		"error":  output.Error,
	}
}

func (output *Output) FromMap(values map[string]interface{}) error {
	var err error

	output.Status, err = coerce.ToString(values["status"])
	if err != nil {
		return err
	}

	output.Error, err = coerce.ToObject(values["error"])
	if err != nil {
		return err
	}

	return nil
}
