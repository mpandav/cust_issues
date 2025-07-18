package sendmessage

import (
	"encoding/json"
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
	_ = activity.Register(&SendMessageActivity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

type SendMessageActivity struct {
	emsConnectionMgr connection.Manager
	producer         *tibems.MsgProducer
	destinationType  string
	deliveryDelay    int64
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

	act := &SendMessageActivity{}
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
	act.deliveryDelay = s.DeliveryDelay
	act.destinationType = s.DestinationType
	return act, nil
}

func (*SendMessageActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval
func (a *SendMessageActivity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	output := &Output{}

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
	if input.Headers["replyTo"] != nil {
		replyTo := input.Headers["replyTo"].(string)
		replyDestinationObj, err := createDestination(a.destinationType, replyTo)
		defer replyDestinationObj.Close()
		if err == nil {
			msg.SetReplyTo(replyDestinationObj)
		} else {
			ctx.Logger().Errorf("Failed to set reply destionation %s", err.Error())
		}
	}
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
			a.logger.Errorf("Error setting message property %s: %v", propName, setErr)
		}
	}

	err = a.producer.Send(msg, sendOptions)
	if err != nil {
		return false, err
	}

	output.MessageId, err = msg.GetMessageID()
	if err != nil {
		return false, err
	}

	err = ctx.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	return true, nil

}

func (a *SendMessageActivity) Cleanup() (err error) {
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
