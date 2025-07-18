package sendmail

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
)

var activityMd = activity.ToMetadata(&Input{})

type SendMailActivity struct {
	clients map[string]*smtp.Client
}

func init() {
	_ = activity.Register(&SendMailActivity{}, New)
}

// New creates new instance of SendMailActivity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &SendMailActivity{clients: make(map[string]*smtp.Client)}, nil
}

// Metadata returns the activity's metadata
func (a *SendMailActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Sends the Message
func (a *SendMailActivity) Eval(context activity.Context) (done bool, err error) {
	activityName := context.Name()

	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	if input.Server == "" {
		return false, activity.NewError(fmt.Sprintf("SMTP Server host must be configured for SendMail Activity [%s] in Flow[%s].", activityName, context.ActivityHost().Name()), "", nil)
	}

	if input.Port == "" {
		return false, activity.NewError(fmt.Sprintf("SMTP Server port must be configured for SendMail Activity[%s] in Flow[%s].", activityName, context.ActivityHost().Name()), "", nil)
	}

	sender := input.Sender
	userName := input.Username
	if sender == "" {
		sender = userName
		if sender == "" {
			return false, activity.NewError(fmt.Sprintf("Sender or UserName must be configured for SendMail Activity[%s] in Flow[%s].", activityName, context.ActivityHost().Name()), "", nil)
		}
	}

	recipients := input.Recipients
	ccRecipients := input.CcRecipients
	bccRecipients := input.BccRecipients
	if recipients == "" && ccRecipients == "" && bccRecipients == "" {
		return false, activity.NewError(fmt.Sprintf("One or more recipients must be configured for SendMail Activity[%s] in Flow[%s].", activityName, context.ActivityHost().Name()), "", nil)
	}

	var allRecepients []string
	if recipients != "" {
		for _, email := range strings.Split(recipients, ",") {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return false, activity.NewError(fmt.Sprintf("Invalid [To] email address[%s]. A valid email address must be configured for SendMail Activity[%s] in Flow[%s].", email, activityName, context.ActivityHost().Name()), "", nil)
			}
			allRecepients = append(allRecepients, email)
		}
	}

	if ccRecipients != "" {
		for _, email := range strings.Split(ccRecipients, ",") {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return false, activity.NewError(fmt.Sprintf("Invalid [Cc] email address[%s]. A valid email address must be configured for SendMail Activity[%s] in Flow[%s].", email, activityName, context.ActivityHost().Name()), "", nil)
			}
			allRecepients = append(allRecepients, email)
		}
	}

	if bccRecipients != "" {
		for _, email := range strings.Split(bccRecipients, ",") {
			_, err := mail.ParseAddress(email)
			if err != nil {
				return false, activity.NewError(fmt.Sprintf("Invalid [Bcc] email address[%s]. A valid email address must be configured for SendMail Activity[%s] in Flow[%s].", email, activityName, context.ActivityHost().Name()), "", nil)
			}
			allRecepients = append(allRecepients, email)
		}
	}

	bodyContentType := input.MessageContentType
	if bodyContentType == "" {
		bodyContentType = "text/plain"
	}

	if strings.Contains(bodyContentType, ";charset") == false {
		// Set charset to UTF
		bodyContentType = bodyContentType + ";charset=\"utf-8\""
	}

	email := bytes.NewBuffer(nil)

	defer func() { email = nil }()

	email.WriteString("From: " + sender + "\r\n")
	t := time.Now()
	email.WriteString("Date: " + t.Format(time.RFC1123Z) + "\r\n")

	if recipients != "" {
		email.WriteString("To: " + strings.TrimRight(recipients, ",") + "\r\n")
	}

	if ccRecipients != "" {
		email.WriteString("Cc: " + strings.TrimRight(recipients, ",") + "\r\n")
	}

	sbj := input.Subject
	email.WriteString("Subject: " + sbj + "\r\n")

	message := input.Message

	replyTo := input.ReplyTo

	if replyTo != "" {
		email.WriteString("Reply-To: " + replyTo + "\r\n")
	}

	if len(input.Attachments) > 0 {
		boundary := "**==tibcoflogo==**"
		email.WriteString("MIME-Version: 1.0\r\n")
		email.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
		email.WriteString("\r\n--" + boundary + "\r\n")

		if message != "" {
			email.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", bodyContentType))
			email.WriteString(message)
			email.WriteString("\r\n")
		}

		for i, attachment := range input.Attachments {
			file := attachment.File
			if file == "" {
				//TODO return error
			}

			var fileContents []byte
			fileName := attachment.FileName
			if strings.HasPrefix(file, "file://") {
				// Read file from disk
				filePath := file[7:]
				fileData, err := ioutil.ReadFile(filePath)
				if err != nil {
					return false, err
				}

				if fileName == "" {
					fileName = filepath.Base(filePath)
				}

				fileContents = fileData
			} else {
				fileContents = []byte(file)
			}

			if fileName == "" {
				fileName = "attachment_" + strconv.Itoa(i)
			}

			email.WriteString("\r\n\r\n--" + boundary + "\r\n")

			if attachment.Inline {
				email.WriteString("Content-Type: message/rfc822\r\n")
				email.WriteString("Content-Disposition: inline; filename=\"" + fileName + "\"\r\n\r\n")
				email.Write(fileContents)
			} else {
				ext := filepath.Ext(fileName)
				mimeType := mime.TypeByExtension(ext)
				if mimeType != "" {
					email.WriteString(fmt.Sprintf("Content-Type: %s\r\n", mimeType))
				} else {
					email.WriteString("Content-Type: application/octet-stream\r\n")
				}
				email.WriteString("Content-Transfer-Encoding: base64\r\n")
				email.WriteString("Content-Disposition: attachment;filename=\"" + fileName + "\"\r\n")
				if attachment.Base64EncodedContents {
					// Contents are already base64 encoded
					email.WriteString("\r\n" + string(fileContents))
				} else {
					email.WriteString("\r\n" + base64.StdEncoding.EncodeToString(fileContents))
				}
			}
		}
	} else {
		if message != "" {
			email.WriteString("MIME-Version: 1.0\r\n")
			email.WriteString(fmt.Sprintf("Content-Type: %s\r\n\r\n", bodyContentType))
			email.WriteString(message)
			email.WriteString("\r\n")
		}
	}

	password := input.Password
	connectionType := input.ConnectionType
	if connectionType == "" {
		connectionType = "TLS"
	}

	if strings.HasSuffix(password, "=") {
		pwdData, err := base64.StdEncoding.DecodeString(password)
		if err == nil {
			password = string(pwdData)
		}
	}

	context.Logger().Infof("Client Connection Type - [%s]", connectionType)

	if connectionType == "TLS" {
		//Use authentication
		err = smtp.SendMail(
			input.Server+":"+input.Port,
			getAuth(userName, password, input.Server),
			sender,
			allRecepients,
			email.Bytes(),
		)
		if err != nil {
			if err.Error() == "EOF" {
				context.Logger().Errorf("Failed to send email due to error - [%s]. Try SSL connection mode.", err.Error())
				return false, err
			} else {
				context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
				return false, err
			}
		}
	} else if connectionType == "SSL" {

		// TLS config
		tlsconfig := &tls.Config{
			ServerName: input.Server,
		}

		if input.ServerCert == "" {
			tlsconfig.InsecureSkipVerify = true
		} else {
			cert, err := getCert(input.ServerCert)
			if err != nil {
				context.Logger().Errorf("Failed to load certificate due to error - [%s]", err.Error())
				return false, err
			}
			if cert != nil {
				tlsconfig.RootCAs = cert
			} else {
				tlsconfig.InsecureSkipVerify = true
			}
		}

		conn, err := tls.Dial("tcp", input.Server+":"+input.Port, tlsconfig)
		if err != nil {
			tlsErr, ok := err.(tls.RecordHeaderError)
			if ok {
				context.Logger().Errorf("Failed to connect to the mail server due to error - [%s]. Use TLS connection type.", tlsErr.Error())
				return false, err
			} else {
				context.Logger().Errorf("Failed to connect to the mail server due to error - [%s]", err.Error())
				return false, err
			}
		}

		sslClient, err := smtp.NewClient(conn, input.Server)
		if err != nil {
			context.Logger().Errorf("Failed to create a mail client due error - [%s]", err.Error())
			return false, err
		}
		defer func() {
			_ = sslClient.Quit()
			_ = sslClient.Close()
		}()
		// Auth
		if ok, _ := sslClient.Extension("AUTH"); ok {
			if err = sslClient.Auth(getAuth(userName, password, input.Server)); err != nil {
				context.Logger().Errorf("Failed to authenticate with the mail server due to error - [%s]", err.Error())
				return false, err
			}
		}

		// To && From
		if err = sslClient.Mail(sender); err != nil {
			context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
			return false, err
		}

		for _, recp := range allRecepients {
			if recp == "" {
				continue
			}
			if err = sslClient.Rcpt(recp); err != nil {
				context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
				return false, err
			}
		}

		// Data
		w, err := sslClient.Data()
		if err != nil {
			context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
			return false, err
		}

		_, err = w.Write(email.Bytes())
		if err != nil {
			context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
			return false, err
		}

		err = w.Close()
		if err != nil {
			context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
			return false, err
		}

	} else {
		smtpClient, err := smtp.Dial(input.Server + ":" + input.Port)
		if err != nil {
			context.Logger().Errorf("Failed to connect to the mail server due to error - [%s].", err.Error())
			return false, err
		}

		defer func() {
			_ = smtpClient.Quit()
			_ = smtpClient.Close()
		}()

		if err = smtpClient.Mail(sender); err != nil {
			if strings.Contains(err.Error(), "STARTTLS") {
				context.Logger().Errorf("Failed to send email due to error - [%s]. Use TLS connection type.", err.Error())
			} else {
				context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
			}
			return false, err
		}
		for _, recp := range allRecepients {
			if recp == "" {
				continue
			}
			if err = smtpClient.Rcpt(recp); err != nil {
				context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
				return false, err
			}
		}
		// Send the email body.
		wc, err := smtpClient.Data()
		if err != nil {
			return false, err
		}
		defer wc.Close()
		strData := strings.Replace(email.String(), ",", ";", -1)
		nBuf := bytes.NewBufferString(strData)
		if _, err = nBuf.WriteTo(wc); err != nil {
			context.Logger().Errorf("Failed to send email due to error - [%s]", err.Error())
			return false, err
		}
	}

	context.Logger().Info("Mail successfully sent")

	return true, nil
}

func getCert(serverCert string) (*x509.CertPool, error) {

	//if certificate comes from fileselctor it will be base64 encoded
	var encodedDataOfCert []byte
	var err error
	if strings.HasPrefix(serverCert, "{") {
		certObj, err := coerce.ToObject(serverCert)
		if err != nil {
			return nil, err
		}
		certRealValue, ok := certObj["content"].(string)
		if !ok || certRealValue == "" {
			return nil, nil
		}

		index := strings.IndexAny(certRealValue, ",")
		if index > -1 {
			certRealValue = certRealValue[index+1:]
		}
		encodedDataOfCert, err = base64.StdEncoding.DecodeString(certRealValue)
		if err != nil {
			return nil, fmt.Errorf("Invalid base64 encoded certificate value")
		}
	} else {
		encodedDataOfCert, err = base64.StdEncoding.DecodeString(serverCert)
		if err != nil {
			return nil, fmt.Errorf("Invalid base64 encoded certificate. Check override value configured to the application property.")
		}
	}

	caCertPool := x509.NewCertPool()
	pemBlock, _ := pem.Decode(encodedDataOfCert)
	if pemBlock == nil {
		return nil, fmt.Errorf("Unsupported certificate found. It must be a valid PEM certificate.")
	}
	serverCert1, err1 := x509.ParseCertificate(pemBlock.Bytes)
	if err1 != nil {
		return nil, err1
	}
	caCertPool.AddCert(serverCert1)
	return caCertPool, nil
}

func getRecipients(recipients string) []string {
	newRecipients := strings.Split(recipients, ",")
	var finalRecipients []string
	for _, rec := range newRecipients {
		if rec != "" {
			finalRecipients = append(finalRecipients, strings.TrimSpace(rec))
		}
	}
	return finalRecipients
}

func getAuth(userName, password, server string) smtp.Auth {
	if strings.Contains(server, "smtp.office365.com") {
		return office365Auth(userName, password)
	}
	return smtp.PlainAuth(
		"",
		userName,
		password,
		server,
	)
}

type loginAuth struct {
	username, password string
}

func office365Auth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	prompt := strings.TrimSpace(string(fromServer))
	switch prompt {
	case "Username:":
		return []byte(a.username), nil
	case "Password:":
		return []byte(a.password), nil
	default:
		return nil, errors.New("Unknown fromServer")
	}
}
