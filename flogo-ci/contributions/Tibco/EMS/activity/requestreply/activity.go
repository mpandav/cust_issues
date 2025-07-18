package requestreply

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/tibco/msg-ems-client-go/tibems"
)

func init() {
	_ = activity.Register(&RequestreplyActivity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

type RequestreplyActivity struct {
	emsConnectionMgr connection.Manager
	producer         *tibems.MsgProducer
	destinationType  string
	replyDestination string
	requestTimeout   int64
	logger           log.Logger
}

type Properties struct {
	Property []Property `json:"property,omitempty"`
}

// Property
type Property struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	act := &RequestreplyActivity{}
	act.emsConnectionMgr = s.EMSConManager
	act.logger = ctx.Logger()

	//create producer
	connection := act.emsConnectionMgr.GetConnection().(*tibems.Connection)
	session, err := sessionCache.GetSession(connection)
	if err != nil {
		return nil, err
	}
	destinationObj, err := createDestination(s.DestinationType, s.Destination)
	if err != nil {
		return nil, err
	}
	act.producer, err = session.CreateProducer(destinationObj)
	if err != nil {
		return nil, err
	}
	if s.DeliveryDelay > 0 {
		act.producer.SetDeliveryDelay(s.DeliveryDelay)
	}

	act.destinationType = s.DestinationType
	act.replyDestination = s.ReplyToDestination
	act.requestTimeout = s.RequestTimeout
	return act, nil
}

func (*RequestreplyActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *RequestreplyActivity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	output := &Output{
		OutputHeaders:     make(map[string]interface{}),
		MessageProperties: make(map[string]interface{}),
	}

	//get connection and create seesion if not exists
	connection := a.emsConnectionMgr.GetConnection().(*tibems.Connection)
	session, err := sessionCache.GetSession(connection)
	if err != nil {
		return false, err
	}

	// create ems msg
	msg, err := tibems.CreateTextMsg()
	defer msg.Close()
	if err != nil {
		return false, err
	}
	err = msg.SetText(input.MessageBody)
	if err != nil {
		return false, err
	}
	if input.Headers["correlationID"] != nil {
		correlationID := input.Headers["correlationID"].(string)
		msg.SetCorrelationID(correlationID)
	}

	var replyToDestination *tibems.Destination
	if input.Headers["replyTo"] != nil && input.Headers["replyTo"].(string) != "" {
		replyTo := input.Headers["replyTo"].(string)
		replyToDestination, err = createDestination(a.destinationType, replyTo)
	} else if a.replyDestination != "" {
		//create destination acc a.destinationType
		replyToDestination, err = createDestination(a.destinationType, a.replyDestination)
	} else {
		//create temp destination to receive reply
		if a.destinationType == "Queue" {
			replyToDestination, err = session.CreateTemporaryQueue()
		} else {
			replyToDestination, err = session.CreateTemporaryTopic()
		}
	}
	if err == nil {
		msg.SetReplyTo(replyToDestination)
	} else {
		ctx.Logger().Errorf("Failed to set reply to destination %s", err.Error())
		return false, err
	}
	defer replyToDestination.Close()

	if trace.Enabled() {
		tracingHeader := make(map[string]string)
		_ = trace.GetTracer().Inject(ctx.GetTracingContext(), trace.TextMap, tracingHeader)
		for headerKey, headerValue := range tracingHeader {
			msg.SetStringProperty(headerKey, headerValue)
		}
	}

	sendOptions := &tibems.SendOptions{}
	if input.Headers["deliveryMode"] != nil {
		deliveryMode := strings.ToLower(input.Headers["deliveryMode"].(string))
		if deliveryMode == "reliable" {
			sendOptions.DeliveryMode = tibems.DeliveryModeReliable
		} else if deliveryMode == "nonpersistent" {
			sendOptions.DeliveryMode = tibems.DeliveryModeNonPersistent
		} else {
			sendOptions.DeliveryMode = tibems.DeliveryModePersistent
		}
	} else {
		sendOptions.DeliveryMode = tibems.DeliveryModePersistent
	}

	if input.Headers["expiration"] != nil {
		sendOptions.TimeToLive = int64(input.Headers["expiration"].(float64))
	} else {
		sendOptions.TimeToLive = 0
	}
	if input.Headers["priority"] != nil {
		sendOptions.Priority = int32(input.Headers["priority"].(float64))
	} else {
		sendOptions.Priority = 4
	}

	if input.Destination != "" {
		destinationObj, err := createDestination(a.destinationType, input.Destination)
		if err != nil {
			return false, err
		}
		defer destinationObj.Close()
		sendOptions.Destination = destinationObj
	}

	// Handle message properties
	properties := &Properties{}
	dataBytes, err := json.Marshal(input.MessageProperties)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(dataBytes, properties)
	if err != nil {
		return false, err
	}

	for _, prop := range properties.Property {
		propName := prop.Name
		propValue := prop.Value
		propType := prop.Type

		// If type is empty, default to "string"
		if propType == "" {
			propType = "string"
		}

		var setErr error
		switch strings.ToLower(propType) {
		case "string":
			setErr = msg.SetStringProperty(propName, propValue)
		case "boolean":
			boolValue, coerceErr := coerce.ToBool(propValue)
			if coerceErr != nil {
				setErr = fmt.Errorf("property %s: cannot coerce to boolean: %w", propName, coerceErr)
			} else {
				setErr = msg.SetBooleanProperty(propName, boolValue)
			}
		case "short":
			intValue, coerceErr := coerce.ToInt32(propValue)
			if coerceErr != nil {
				setErr = fmt.Errorf("property %s: cannot coerce to short: %w", propName, coerceErr)
			} else if intValue < math.MinInt16 || intValue > math.MaxInt16 {
				setErr = fmt.Errorf("short value out of range for property %s: %d", propName, intValue)
			} else {
				setErr = msg.SetShortProperty(propName, int16(intValue))
			}
		case "integer":
			intValue, coerceErr := coerce.ToInt32(propValue)
			if coerceErr != nil {
				setErr = fmt.Errorf("property %s: cannot coerce to integer: %w", propName, coerceErr)
			} else {
				setErr = msg.SetIntProperty(propName, intValue)
			}
		case "long":
			intValue, coerceErr := coerce.ToInt64(propValue)
			if coerceErr != nil {
				setErr = fmt.Errorf("property %s: cannot coerce to long: %w", propName, coerceErr)
			} else {
				setErr = msg.SetLongProperty(propName, intValue)
			}
		case "float":
			floatValue, coerceErr := coerce.ToFloat32(propValue)
			if coerceErr != nil {
				setErr = fmt.Errorf("property %s: cannot coerce to float: %w", propName, coerceErr)
			} else {
				setErr = msg.SetFloatProperty(propName, floatValue)
			}
		case "double":
			floatValue, coerceErr := coerce.ToFloat64(propValue)
			if coerceErr != nil {
				setErr = fmt.Errorf("property %s: cannot coerce to double: %w", propName, coerceErr)
			} else {
				setErr = msg.SetDoubleProperty(propName, floatValue)
			}
		default:
			ctx.Logger().Warnf("Incorrect type for property %s: defaulting to string", propName)
			setErr = msg.SetStringProperty(propName, propValue)
		}

		if setErr != nil {
			ctx.Logger().Errorf("Error setting message property %s: %v", propName, setErr)
		}
	}

	err = a.producer.Send(msg, sendOptions)
	if err != nil {
		return false, err
	}

	//create consumer to listen for reply
	selector := ""
	if cId, _ := msg.GetCorrelationID(); cId != "" {
		selector = "JMSCorrelationID='" + cId + "'"
	} else {
		messageID, _ := msg.GetMessageID()
		selector = "JMSCorrelationID='" + messageID + "'"
	}

	// create seperate session as JMS specification doesn't allow concurrent consumers
	consumerSession, err := connection.CreateSession(false, tibems.AckModeAutoAcknowledge)
	if err != nil {
		ctx.Logger().Error(err)
	}
	defer consumerSession.Close()

	consumer, err := consumerSession.CreateConsumer(replyToDestination, selector, false)
	if err != nil {
		ctx.Logger().Error(err)
		return false, err
	}
	defer consumer.Close()
	replyMessage, err := consumer.ReceiveTimeout(a.requestTimeout)
	if err != nil {
		if errors.Is(err, tibems.ErrExceededLimit) {
			ctx.Logger().Error("Timeout occurred while receiving reply ", err)
		} else {
			ctx.Logger().Error(err)
		}
		return false, err
	}
	// message headers
	msgDestination, err := replyMessage.GetDestination()
	if err == nil {
		output.OutputHeaders["destination"], _ = msgDestination.GetName()
	}

	replytodest, err := replyMessage.GetReplyTo()
	if err == nil && replytodest != nil {
		output.OutputHeaders["replyTo"], _ = replytodest.GetName()
	}

	replyDeliveryMode, err := replyMessage.GetDeliveryMode()
	if err == nil {
		if replyDeliveryMode == tibems.DeliveryModePersistent {
			output.OutputHeaders["deliveryMode"] = "persistent"
		} else if replyDeliveryMode == tibems.DeliveryModeNonPersistent {
			output.OutputHeaders["deliveryMode"] = "nonpersistent"
		} else {
			output.OutputHeaders["deliveryMode"] = "reliable"
		}
	}
	output.OutputHeaders["messageID"], _ = replyMessage.GetMessageID()
	output.OutputHeaders["timestamp"], _ = replyMessage.GetTimestamp()
	output.OutputHeaders["expiration"], _ = replyMessage.GetExpiration()
	output.OutputHeaders["reDelivered"], _ = replyMessage.GetRedelivered()
	output.OutputHeaders["priority"], _ = replyMessage.GetPriority()
	output.OutputHeaders["correlationID"], _ = replyMessage.GetCorrelationID()

	// Handle reply message properties
	getMessageProperties(ctx.Logger(), replyMessage, output)

	replyMessageText := replyMessage.(*tibems.TextMsg)
	output.Message, err = replyMessageText.GetText()
	if err != nil {
		ctx.Logger().Error(err)
	}

	err = ctx.SetOutputObject(output)
	if err != nil {
		ctx.Logger().Error(err)
	}
	return true, nil
}

// getMessageProperties retrieves message properties and adds them to the output
func getMessageProperties(logger log.Logger, message tibems.Message, out *Output) {
	messageEnumeration, err := message.GetPropertyNames()
	if err == nil {
		var properties []Property
		for {
			propertyName, err := messageEnumeration.GetNextName()
			if err != nil {
				break
			}

			propertyValue, err := message.GetStringProperty(propertyName)
			if err != nil {
				logger.Errorf("Error getting property value for %s: %v", propertyName, err)
				continue
			}

			// Get property type, default to STRING if not found
			propertyType, err := message.GetPropertyType(propertyName)
			if err != nil {
				propertyType = 9 // 9 corresponds to STRING
			}

			var propertyTypeString string
			switch propertyType {
			case 1:
				propertyTypeString = "boolean"
			case 4:
				propertyTypeString = "short"
			case 5:
				propertyTypeString = "integer"
			case 6:
				propertyTypeString = "long"
			case 7:
				propertyTypeString = "float"
			case 8:
				propertyTypeString = "double"
			case 9:
				propertyTypeString = "string"
			default:
				propertyTypeString = "string"
			}
			properties = append(properties, Property{
				Name:  propertyName,
				Type:  propertyTypeString,
				Value: propertyValue,
			})
		}
		out.MessageProperties["property"] = properties
	} else {
		logger.Errorf("Error getting property names: %v", err)
	}
}

func (a *RequestreplyActivity) Cleanup() (err error) {
	if a.producer != nil {
		a.producer.Close()
	}
	return nil
}

func createDestination(DestinationType string, destination string) (destinationObj *tibems.Destination, err error) {
	if DestinationType == "Queue" {
		destinationObj, err = tibems.CreateDestination(tibems.DestTypeQueue, destination)
		if err != nil {
			return nil, err
		}
	} else {
		destinationObj, err = tibems.CreateDestination(tibems.DestTypeTopic, destination)
		if err != nil {
			return nil, err
		}
	}
	return destinationObj, nil
}
