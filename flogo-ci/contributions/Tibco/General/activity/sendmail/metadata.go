package sendmail

import (
	"encoding/json"

	"github.com/project-flogo/core/data/coerce"
)



type Input struct {
	Server             string `md:"Server,required"`
	Port               string    `md:"Port,required"`
	ConnectionType     string `md:"Connection Security"`
	ServerCert         string `md:"serverCertificate"`
	Username           string `md:"Username"`
	Password           string `md:"Password"`
	MessageContentType string `md:"message_content_type"`
	Sender             string `md:"sender"`
	Recipients         string `md:"recipients"`
	CcRecipients       string `md:"cc_recipients"`
	BccRecipients      string `md:"bcc_recipients"`
	ReplyTo            string `md:"reply_to"`
	Subject            string `md:"subject"`
	Message            string `md:"message"`
	Attachments        []attachment `md:"attachments"`
}

type attachment struct {
	File                  string
	FileName              string
	Inline                bool
	Base64EncodedContents bool
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Server":               i.Server,
		"Port":                 i.Port,
		"Connection Security":  i.ConnectionType,
		"serverCertificate": i.ServerCert,
		"Username":             i.Username,
		"Password":             i.Password,
		"message_content_type": i.MessageContentType,
		"sender":               i.Sender,
		"recipients":           i.Recipients,
		"cc_recipients":        i.CcRecipients,
		"bcc_recipients":       i.BccRecipients,
		"reply_to":             i.ReplyTo,
		"subject":              i.Subject,
		"message":              i.Message,
		"attachments":          i.Attachments,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.Server, err = coerce.ToString(values["Server"])
	if err != nil {
		return err
	}

	i.Port, err = coerce.ToString(values["Port"])
	if err != nil {
		return err
	}

	i.ConnectionType, err = coerce.ToString(values["Connection Security"])
	if err != nil {
		return err
	}

	i.Username, err = coerce.ToString(values["Username"])
	if err != nil {
		return err
	}

	i.Password, err = coerce.ToString(values["Password"])
	if err != nil {
		return err
	}

	i.MessageContentType, err = coerce.ToString(values["message_content_type"])
	if err != nil {
		return err
	}

	i.Sender, err = coerce.ToString(values["sender"])
	if err != nil {
		return err
	}

	i.Recipients, err = coerce.ToString(values["recipients"])
	if err != nil {
		return err
	}

	i.CcRecipients, err = coerce.ToString(values["cc_recipients"])
	if err != nil {
		return err
	}

	i.BccRecipients, err = coerce.ToString(values["bcc_recipients"])
	if err != nil {
		return err
	}

	i.ReplyTo, err = coerce.ToString(values["reply_to"])
	if err != nil {
		return err
	}

	i.Subject, err = coerce.ToString(values["subject"])
	if err != nil {
		return err
	}
	i.Message, _ = coerce.ToString(values["message"])

	attachments, _ := json.Marshal(values["attachments"])
	err = json.Unmarshal(attachments, &i.Attachments)
	if err != nil {
		return err
	}

	i.ServerCert, _ = coerce.ToString(values["serverCertificate"])
	return nil
}
