package invokesync

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/trace"

	"github.com/aws/aws-sdk-go/service/lambda"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil
}

// Activity is a App Activity implementation
type Activity struct {
	// TODO: ask Tracy "why Mutex here"
	sync.Mutex
	metadata *activity.Metadata
}

// Metadata returns activity metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activity.ToMetadata(&Input{}, &Output{})
}

// Eval implements activity.Activity.Eval
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	logger := ctx.Logger()
	logger.Infof("Executing activity Lambda Invoke - [%s]", ctx.Name())
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	sess := input.Connection.GetConnection().(*session.Session)
	lambdaClient := lambda.New(sess)

	if len(input.LambdaARN) > 0 {
		input.ARN = input.LambdaARN
	}

	if len(input.ARN) == 0 {
		return false, activity.NewError("Function Name or ARN is required", "", nil)
	}

	payloadBytes, err := json.Marshal(input.Payload)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Can't get payload from input %s", err.Error()), "", err)
	}
	logger.Debugf("Lambda invoke payload: %s", string(payloadBytes))
	clientContext := &lambdacontext.ClientContext{}
	if ctx.GetTracingContext() != nil && trace.Enabled() {
		_ = trace.GetTracer().Inject(ctx.GetTracingContext(), trace.Lambda, clientContext)
	}
	clientContextEncoded, err := encodeClientContext(clientContext)
	if err != nil {
		return false, activity.NewError(fmt.Sprintf("Unable to encode client context: %s", err.Error()), "", err)
	}
	invokeOutput, err := lambdaClient.Invoke(&lambda.InvokeInput{
		ClientContext: &clientContextEncoded,
		FunctionName:  &input.ARN,
		Payload:       payloadBytes,
	})

	if err != nil {
		logger.Error(err)
		return false, activity.NewError(fmt.Sprintf("Unable to invoke lambda function %s", err.Error()), "", err)
	}
	logger.Infof("Lambda invoke output status code [%d]", *invokeOutput.StatusCode)

	payloadCoerced, err := coerce.ToObject(invokeOutput.Payload)

	output := &Output{
		Status: *invokeOutput.StatusCode,
		Result: payloadCoerced,
	}
	return true, ctx.SetOutputObject(output)
}

func encodeClientContext(clientContext *lambdacontext.ClientContext) (string, error) {
	ccbytes, err := json.Marshal(clientContext)
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString(ccbytes), nil
}
