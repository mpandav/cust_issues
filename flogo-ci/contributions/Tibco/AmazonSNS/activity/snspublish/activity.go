package snspublish

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	snsutil "github.com/tibco/flogo-aws-sns/src/app/AmazonSNS/activity"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// Constants for activity
const (
	ActivityName = "SNSPublish"
	// Configuration constants

	// Mapping parameters
	MessageStructure = "MessageStructure"

	// Miscellaneous
	MessageTypePlainText = "plainText"
	MessageTypeCustom    = "custom"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	act := &Activity{}
	return act, nil
}

// Activity The S3 Get Activity
type Activity struct {
	metadata *activity.Metadata
}

// Metadata returns activity metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	logger.Info(fmt.Sprintf("Executing activity AmazonSNS %s - [%s]", ActivityName, ctx.Name()))

	snsSession := input.Connection.GetConnection().(*session.Session)
	endpointConfig := &aws.Config{}
	if snsSession.Config.Endpoint != nil {
		endpoint := *snsSession.Config.Endpoint
		snsSession.Config.Endpoint = nil
		endpointConfig.Endpoint = aws.String(endpoint)
	}
	snsSvc := sns.New(snsSession, endpointConfig)

	messageType := input.MessageType
	var outputObject interface{}

	if len(input.MessageAttributeNames) > 0 {
		attribValueMap := input.Input["MessageAttributes"].(map[string]interface{})
		mAttributes := snsutil.ConvertMAttributes(input.MessageAttributeNames, attribValueMap)
		input.Input["MessageAttributes"] = mAttributes
	}

	switch messageType {
	case MessageTypePlainText:
		input.Input[MessageStructure] = ""
		outputObject, err = snsutil.DoPublishPlainText(ctx, snsSvc, input.Input, logger)
		logger.Debug(ctx.Name(), outputObject)
		if snsutil.IsFatalError(err) {
			return false, activity.NewError(err.Error(), "", err)
		}
	case MessageTypeCustom:
		input.Input[MessageStructure] = "json"
		outputObject, err = snsutil.DoPublishCustom(ctx, snsSvc, input.Input, logger)
		logger.Debug(ctx.Name(), outputObject)
		if snsutil.IsFatalError(err) {
			return false, activity.NewError(err.Error(), "", err)
		}
	}
	// Setting the output
	outputObjectCoerced, err := coerce.ToObject(outputObject)
	if err != nil {
		return false, err
	}
	output := &Output{
		Output: outputObjectCoerced,
	}
	ctx.SetOutputObject(output)
	return true, nil
}
