package listactivity

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
	Output map[string]interface{} `md:"Output"`
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
		"Output": output.Output,
		"error":  output.Error,
	}
}

func (output *Output) FromMap(values map[string]interface{}) error {
	var err error

	output.Output, err = coerce.ToObject(values["Output"])
	if err != nil {
		return err
	}

	output.Error, err = coerce.ToObject(values["error"])
	if err != nil {
		return err
	}

	return nil
}
