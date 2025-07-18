package sqsreceivemessage

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

var triggerLog = log.ChildLogger(log.RootLogger(), "aws-trigger-sqsreceivemessage")
var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

type Factory struct {
}

func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}
	err := metadata.MapToStruct(config.Settings, s, true)

	if err != nil {
		return nil, err
	}

	cm, err := coerce.ToConnection(s.AWSConnection)

	if err != nil {
		return nil, err
	}

	session := cm.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if session.Config.Endpoint != nil {
		endpoint := *session.Config.Endpoint
		session.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}

	SQSSvc := sqs.New(session, endpointConfig)
	return &Trigger{name: config.Id, sqssvc: SQSSvc}, nil
}

type Trigger struct {
	SQSHandlers map[string]*SQSHandler
	sqssvc      *sqs.SQS
	name        string
}

type SQSHandler struct {
	sqssvc              *sqs.SQS
	handler             trigger.Handler
	settings            *HandlerSettings
	shutdown            chan bool
	triggerName         string
	receiveMessageInput sqs.ReceiveMessageInput
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {

	var err error
	t.SQSHandlers = make(map[string]*SQSHandler)
	for _, handler := range ctx.GetHandlers() {
		handlerSetting := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)
		if err != nil {
			return err
		}
		if handlerSetting.QueueURL == "" {
			return errors.New("Empty Queue URL")
		}

		tags := make(map[string]string)

		sqsHandler := &SQSHandler{}
		sqsHandler.settings = handlerSetting
		sqsHandler.handler = handler
		sqsHandler.shutdown = make(chan bool)
		sqsHandler.sqssvc = t.sqssvc
		sqsHandler.triggerName = t.name
		t.SQSHandlers[handlerSetting.QueueURL] = sqsHandler

		receiveMessageInput := &sqsHandler.receiveMessageInput
		receiveMessageInput.QueueUrl = aws.String(handlerSetting.QueueURL)

		/*attrsNames := handlerSetting.AttributeNames
		  if attrsNames != nil && len(attrsNames) > 0 {
		      //Add attribute names
		      attrs := make([]*string, len(attrsNames))
		      for i, v := range attrsNames {
		          attrInfo, _ := coerce.ToObject(v)
		          if attrInfo != nil && attrInfo["Name"] != nil {
		              attrs[i] = aws.String(attrInfo["Name"].(string))
		          }
		      }
		      receiveMessageInput.AttributeNames = attrs
		  }*/

		//Attributes for standard queue
		attrs := make([]*string, 5)
		attrs = append(attrs, aws.String("ApproximateReceiveCount"))
		attrs = append(attrs, aws.String("ApproximateFirstReceiveTimestamp"))
		attrs = append(attrs, aws.String("SenderId"))
		attrs = append(attrs, aws.String("SentTimestamp"))
		attrs = append(attrs, aws.String("SequenceNumber"))

		//Attributes for fifo queue
		if strings.HasSuffix(handlerSetting.QueueURL, ".fifo") {
			attrs = append(attrs, aws.String("MessageDeduplicationId"))
			attrs = append(attrs, aws.String("MessageGroupId"))
		}

		receiveMessageInput.AttributeNames = attrs

		attrsNames := handlerSetting.MessageAttributeNames
		if attrsNames != nil && len(attrsNames) > 0 {
			attrs := make([]*string, len(attrsNames))
			for i, v := range attrsNames {
				attrInfo, _ := coerce.ToObject(v)
				if attrInfo != nil && attrInfo["Name"] != nil {
					attrs[i] = aws.String(attrInfo["Name"].(string))
				}
			}
			receiveMessageInput.MessageAttributeNames = attrs
		}

		maxNumberOfMessages := handlerSetting.MaxNumberOfMessages
		if maxNumberOfMessages != 0 {
			receiveMessageInput.MaxNumberOfMessages = aws.Int64(int64(maxNumberOfMessages))
		}

		visibilityTimeout := handlerSetting.VisibilityTimeout
		if visibilityTimeout != 0 {
			receiveMessageInput.VisibilityTimeout = aws.Int64(int64(visibilityTimeout))
		}

		waitTimeSeconds := handlerSetting.WaitTimeSeconds
		if waitTimeSeconds != 0 {
			receiveMessageInput.WaitTimeSeconds = aws.Int64(int64(waitTimeSeconds))
		}

		receiveRequestAttemptId := handlerSetting.ReceiveRequestAttemptId
		if strings.HasSuffix(handlerSetting.QueueURL, ".fifo") && receiveRequestAttemptId != "" {
			receiveMessageInput.ReceiveRequestAttemptId = aws.String(receiveRequestAttemptId)
		}

		hc, ok := handler.(trigger.HandlerEventConfig)
		if ok {
			hc.SetDefaultEventData(tags)
		}
	}
	return err
}

func (t *Trigger) Start() error {
	for _, handler := range t.SQSHandlers {
		go handler.start()
	}
	return nil
}

func (t *Trigger) Stop() error {
	for _, handler := range t.SQSHandlers {
		handler.shutdown <- true
	}
	return nil
}

func (h *SQSHandler) start() {
	for {
		select {
		case <-h.shutdown:
			triggerLog.Debugf("Stopping receiver for Queue [%s] for trigger [%s]", h.settings.QueueURL, h.triggerName)
			return
		default:
			sqsSvc := h.sqssvc
			response, err := sqsSvc.ReceiveMessage(&h.receiveMessageInput)
			if err != nil {
				triggerLog.Errorf("Trigger [%s] failed to receive message for Queue [%s] due to error - {%v}", h.triggerName, h.settings.QueueURL, err)
			}

			deleteMsgs := h.settings.DeleteMessage

			//Set Message details in the output
			msgs := make([]map[string]interface{}, len(response.Messages))
			if len(response.Messages) > 0 {
				for i, msg := range response.Messages {
					if deleteMsgs {
						deleteMsgInput := &sqs.DeleteMessageInput{}
						deleteMsgInput.SetQueueUrl(h.settings.QueueURL)
						deleteMsgInput.SetReceiptHandle(*msg.ReceiptHandle)
						_, err := sqsSvc.DeleteMessage(deleteMsgInput)
						if err != nil {
							triggerLog.Errorf("Failed to delete received message from SQS due to error:%s", err)
						}
					}
					msgs[i] = make(map[string]interface{})
					//read attributes
					if len(msg.Attributes) > 0 {
						msgs[i]["Attribute"] = make(map[string]string, len(msg.Attributes))
						attrs := msgs[i]["Attribute"].(map[string]string)
						for k, v := range msg.Attributes {
							attrs[k] = *v
						}
					}
					//read message attributes
					if len(msg.MessageAttributes) > 0 {
						msgs[i]["MessageAttributes"] = make(map[string]string, len(msg.MessageAttributes))
						attrs := msgs[i]["MessageAttributes"].(map[string]string)
						for k, v := range msg.MessageAttributes {
							attrs[k] = *v.StringValue
						}
						msgs[i]["MD5OfMessageAttributes"] = *msg.MD5OfMessageAttributes
					}

					if msg.Body != nil {
						msgs[i]["Body"] = *msg.Body
						msgs[i]["MD5OfBody"] = *msg.MD5OfBody
					}
					if msg.MessageId != nil {
						msgs[i]["MessageId"] = *msg.MessageId
					}
					msgs[i]["ReceiptHandle"] = *msg.ReceiptHandle
				}
				output := &Output{}
				output.Message = msgs
				triggerLog.Debugf("Message received for Queue [%s] by trigger [%s]", h.settings.QueueURL, h.triggerName)
				_, err1 := h.handler.Handle(context.Background(), output)
				if err1 != nil {
					triggerLog.Errorf("Trigger [%s] failed to execute action for Queue [%s] due to error - {%v}", h.triggerName, h.settings.QueueURL, err)
				}
			}

		}
	}
}
