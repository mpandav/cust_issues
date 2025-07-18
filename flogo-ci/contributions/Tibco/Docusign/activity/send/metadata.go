package send

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
	docusignconnection "github.com/tibco/wi-docusign/src/app/Docusign/connector/docusign"
)

type Input struct {
	DocusignConnection connection.Manager `md:"docusignConnection"`
	IsMultiDoc         bool               `md:"isMultiDoc"`
	Recipients         string             `md:"recipients"`
	SigningOrder       bool               `md:"signingInOrder"`
	FileName           string             `md:"fileName"`
	FileContent        string             `md:"fileContent"`
	Documents          []Document         `md:"documents"`
}

type Output struct {
	Envelope map[string]interface{} `md:"envelope"`
	Error    map[string]interface{} `md:"error"`
}

func (input *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"docusignConnection": input.DocusignConnection,
		"isMultiDoc":         input.Documents,
		"recipients":         input.Recipients,
		"signingInOrder":     input.SigningOrder,
		"fileName":           input.FileName,
		"fileContent":        input.FileContent,
		"documents":          input.Documents,
	}
}

// var logger = log.ChildLogger(log.RootLogger(), "connection.docusign")

func (input *Input) FromMap(values map[string]interface{}) error {
	var err error

	input.DocusignConnection, err = docusignconnection.GetSharedConfiguration(values["docusignConnection"])
	if err != nil {
		return err
	}

	input.IsMultiDoc, err = coerce.ToBool(values["isMultiDoc"])
	if err != nil {
		return err
	}

	input.Recipients, err = coerce.ToString(values["recipients"])
	if err != nil {
		return err
	}

	input.SigningOrder, err = coerce.ToBool(values["signingInOrder"])
	if err != nil {
		return err
	}

	input.FileName, err = coerce.ToString(values["fileName"])
	if err != nil {
		return err
	}

	input.FileContent, err = coerce.ToString(values["fileContent"])
	if err != nil {
		return err
	}

	docs, err := coerce.ToArray(values["documents"])
	if err != nil {
		return err
	}
	var input_doc Document
	var documentArr []Document
	for _, doc := range docs {
		// logger.Infof("docs: %#v", doc)
		// doc_json, err := json.Marshal(doc)
		// if err != nil {
		// 	return err
		// }
		// logger.Infof("doc_json: %#v", doc_json)

		// err = json.Unmarshal(doc_json, &input_doc)
		// if err != nil {
		// 	return err
		// }

		// logger.Infof("input_doc: %#v", doc.(map[string]interface{}))
		docMap := doc.(map[string]interface{})
		input_doc.Name, err = coerce.ToString(docMap["name"])
		if err != nil {
			return err
		}
		input_doc.Content, err = coerce.ToString(docMap["content"])
		if err != nil {
			return err
		}

		// logger.Infof("input_doc : %#v", input_doc)

		documentArr = append(documentArr, input_doc)
	}
	input.Documents = documentArr
	return nil
}

func (output *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"envelope": output.Envelope,
		"error":    output.Error,
	}
}

func (output *Output) FromMap(values map[string]interface{}) error {
	var err error

	output.Envelope, err = coerce.ToObject(values["envelope"])
	if err != nil {
		return err
	}

	output.Error, err = coerce.ToObject(values["error"])
	if err != nil {
		return err
	}

	return nil
}
