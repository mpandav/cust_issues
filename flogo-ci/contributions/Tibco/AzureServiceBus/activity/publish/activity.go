package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	azureservicebusconnection "github.com/tibco/wi-azservicebus/src/app/AzureServiceBus/connector/connection"
)

//Oss upgrade--

var activityMd = activity.ToMetadata(&Input{}, &Output{})
var azureServiceBusActivityLogger = log.ChildLogger(log.RootLogger(), "azureservicebus-activity")

type (
	// PublishRequest data structure
	PublishRequest struct {
		QueueName        string           `json:"queueName"`
		TopicName        string           `json:"topicName"`
		MessageString    string           `json:"messageString"`
		BrokerProperties BrokerProperties `json:"brokerProperties"`
	}
)
type (
	//BrokerProperties datastructure for storing BrokerProperties
	BrokerProperties struct {
		ContentType             string         `json:"ContentType"`
		CorrelationId           string         `json:"CorrelationId"`
		ForcePersistence        bool           `json:"ForcePersistence"`
		Label                   string         `json:"Label"`
		PartitionKey            string         `json:"PartitionKey"`
		ReplyTo                 string         `json:"ReplyTo"`
		ReplyToSessionId        string         `json:"ReplyToSessionId"`
		SessionId               string         `json:"SessionId"`
		To                      string         `json:"To"`
		TimeToLive              *time.Duration `json:"TimeToLive"`
		ScheduledEnqueueTimeUtc *time.Time     `json:"ScheduledEnqueueTimeUtc"`
	}
)

// PublishResponse datastructure for storing BrokerProperties
type PublishResponse struct {
	ResponseMessage string `json:"responseMessage"`
}

func init() {
	_ = activity.Register(&AzureServiceBusPublishActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &AzureServiceBusPublishActivity{}, nil
}

type AzureServiceBusPublishActivity struct {
}

func (*AzureServiceBusPublishActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *AzureServiceBusPublishActivity) Eval(ctx activity.Context) (done bool, err error) {

	ctx.Logger().Info("AzureServiceBus publish message Activity")

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}

	connection, _ := input.AzureServiceBusConnection.(*azureservicebusconnection.AzureServiceBusSharedConfigManager)
	if input.Input == nil {
		return false, activity.NewError(fmt.Sprintf("Input is required in publish activity for %s object", input.EntityName), "AZSERVICEBUS-PUBLISH-4015", nil)
	}

	var inputSchemaStr string
	//getting schema for custom properties
	if sIO, ok := ctx.(schema.HasSchemaIO); ok {
		inputSchema := sIO.GetInputSchema("customProperties")
		if inputSchema != nil {
			inputSchemaStr = inputSchema.Value()
		}
	}

	inputMap := make(map[string]interface{})
	if input.Input != nil {
		for k, v := range input.Input {
			inputMap[k] = v
		}
	}
	inputparamtersmap := make(map[string]interface{})
	for k, v := range inputMap["parameters"].(map[string]interface{}) {
		inputparamtersmap[k] = v
	}

	entityName := ""
	readresponseData := make(map[string]interface{})
	actualResponse := PublishResponse{}
	var readError error
	dataBytes, err := json.Marshal(inputMap["parameters"])
	if err != nil {
		azureServiceBusActivityLogger.Error(err)
	}
	publishInput := PublishRequest{}
	json.Unmarshal(dataBytes, &publishInput)
	reqmessage := azservicebus.Message{}
	if publishInput.BrokerProperties.PartitionKey != "" {
		reqmessage.PartitionKey = &publishInput.BrokerProperties.PartitionKey
	}
	if publishInput.BrokerProperties.ContentType != "" {
		reqmessage.ContentType = &publishInput.BrokerProperties.ContentType
	}
	if publishInput.BrokerProperties.CorrelationId != "" {
		reqmessage.CorrelationID = &publishInput.BrokerProperties.CorrelationId
	}
	if inputparamtersmap["messageString"] == nil {
		reqmessage.Body = []byte("")
	} else {
		reqmessage.Body = []byte(inputparamtersmap["messageString"].(string))
	}

	if publishInput.BrokerProperties.Label != "" {
		reqmessage.Subject = &publishInput.BrokerProperties.Label
	}
	if publishInput.BrokerProperties.ReplyTo != "" {
		reqmessage.ReplyTo = &publishInput.BrokerProperties.ReplyTo
	}
	if publishInput.BrokerProperties.To != "" {
		reqmessage.To = &publishInput.BrokerProperties.To
	}
	if publishInput.BrokerProperties.TimeToLive != nil {
		reqmessage.TimeToLive = publishInput.BrokerProperties.TimeToLive
	}
	// session support
	if publishInput.BrokerProperties.SessionId != "" {
		reqmessage.SessionID = &publishInput.BrokerProperties.SessionId
	}
	if publishInput.BrokerProperties.ScheduledEnqueueTimeUtc != nil {
		reqmessage.ScheduledEnqueueTime = publishInput.BrokerProperties.ScheduledEnqueueTimeUtc
	}

	//reading custom properties from input parameters
	if inputparamtersmap["customProperties"] != nil {
		customProperties := make(map[string]interface{})
		customProperties = inputparamtersmap["customProperties"].(map[string]interface{})
		reqmessage.ApplicationProperties = processCustomProperties(inputSchemaStr, customProperties)
		//azureServiceBusActivityLogger.Info(fmt.Printf("input map custom properties %#v \n", customProperties))
	}

	if trace.Enabled() {
		_ = trace.GetTracer().Inject(ctx.GetTracingContext(), trace.TextMap, reqmessage.ApplicationProperties)
	}

	if input.EntityType == "Queue" {
		if inputparamtersmap["queueName"] != nil && inputparamtersmap["queueName"].(string) != "" {
			entityName = inputparamtersmap["queueName"].(string)
		} else {
			return false, fmt.Errorf("Queue Name cannot be empty %s", "")
		}

	} else {
		if inputparamtersmap["topicName"] != nil && inputparamtersmap["topicName"].(string) != "" {
			entityName = inputparamtersmap["topicName"].(string)
		} else {
			return false, fmt.Errorf("Topic Name cannot be empty %s", "")
		}
	}
	//get or create connection from connection cache
	sender, err := connection.GetSenderConnection(entityName)
	if err != nil {
		return false, err
	}
	if input.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(input.Timeout))
		defer cancel()
		readError = sender.SendMessage(ctx, &reqmessage, nil)
	} else {
		ctx := context.Background()
		readError = sender.SendMessage(ctx, &reqmessage, nil)
	}

	if readError != nil {
		azureServiceBusActivityLogger.Errorf("Failed to send message to %q\n", entityName)
		return false, fmt.Errorf("Back-end invocation error: %s", readError.Error())
	} else {
		actualResponse.ResponseMessage += " /Published message to " + input.EntityType + " : " + entityName + " successfully / "
		databytes, _ := json.Marshal(actualResponse)
		err = json.Unmarshal(databytes, &readresponseData)
		//azureServiceBusActivityLogger.Info(readresponseData)
	}
	output := &Output{}
	output.Output = readresponseData
	err = ctx.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	ctx.Logger().Debug("AzureServicebus Publish Activity successfully executed")
	return true, nil

}

func (a *AzureServiceBusPublishActivity) Cleanup() error {
	return nil
}

// process/typecast custom properties of a msg
func processCustomProperties(inputSchema string, customProperties map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	if inputSchema != "" {
		var value map[string]interface{}
		err := json.Unmarshal([]byte(inputSchema), &value)
		if err != nil {
			azureServiceBusActivityLogger.Errorf("Unable to unmarshal custom properties schema  %s", err.Error())
			return res
		}
		propertiesSchema := value["properties"].(map[string]interface{})

		for k, v := range customProperties {
			if propertiesSchema[k] != nil {
				switch propertiesSchema[k].(map[string]interface{})["type"] {
				case "number":
					data, err := strconv.ParseFloat(fmt.Sprint(v), 64)
					if err != nil {
						azureServiceBusActivityLogger.Warnf(k+" expected type is number but received value %s", fmt.Sprint(v))
						res[k] = fmt.Sprint(v)
						continue
					}
					res[k] = data
				case "integer":
					data, err := strconv.ParseInt(fmt.Sprint(v), 10, 64)
					if err != nil {
						azureServiceBusActivityLogger.Warnf(k+" expected type is number but received value %s", fmt.Sprint(v))
						res[k] = fmt.Sprint(v)
						continue
					}
					res[k] = data
				case "boolean":
					data, err := strconv.ParseBool(fmt.Sprint(v))
					if err != nil {
						azureServiceBusActivityLogger.Warnf(k+" expected type is number but received value %s", fmt.Sprint(v))
						res[k] = fmt.Sprint(v)
						continue
					}
					res[k] = data
				default:
					res[k] = fmt.Sprint(v)
				}
			}
		}

		//logCache.Info("custom obj %#v", res)
	}
	return res
}
