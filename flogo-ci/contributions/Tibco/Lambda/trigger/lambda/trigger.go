package lambda

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"

	// Import the aws-lambda-go. Required for dep to pull on app create
	_ "github.com/aws/aws-lambda-go/lambda"
)

// log is the default package logger
var logger = log.ChildLogger(log.RootLogger(), "trigger-flogo-lambda")
var singleton *MyTrigger

func init() {
	_ = trigger.Register(&MyTrigger{}, &MyFactory{})
}

// Metadata ...
func (t *MyFactory) Metadata() *trigger.Metadata {
	return trigger.NewMetadata(&Settings{}, &Output{})
}

// New Creates a new trigger instance for a given id
func (t *MyFactory) New(config *trigger.Config) (trigger.Trigger, error) {
	singleton = &MyTrigger{metadata: t.metadata, config: config}
	return singleton, nil
}

// Initialize ...
func (t *MyTrigger) Initialize(ctx trigger.InitContext) error {
	t.handlers = ctx.GetHandlers()
	return nil
}

// Invoke ...
func Invoke() (interface{}, error) {

	logger.Info("Starting AWS Lambda Trigger")

	// Parse the flags
	flag.Parse()

	// Looking up the arguments
	evtArg := flag.Lookup("evt")
	var evt interface{}
	// Unmarshall evt
	if err := json.Unmarshal([]byte(evtArg.Value.String()), &evt); err != nil {
		return nil, err
	}

	logger.Debugf("Received evt: '%+v'\n", evt)

	// Get the context
	ctxArg := flag.Lookup("ctx")
	var lambdaCtx map[string]interface{}

	// Unmarshal ctx
	if err := json.Unmarshal([]byte(ctxArg.Value.String()), &lambdaCtx); err != nil {
		return nil, err
	}

	// logger.Debugf("Received ctx: '%+v'\n", lambdaCtx)
	logger.Debugf("Received lambdaCtx: '%+v'\n", lambdaCtx)

	//select handler, use 0th for now
	handler := singleton.handlers[0]

	// data := map[string]interface{}{
	// 	"Function":     lambdaCtx["Function"],
	// 	"Context":      lambdaCtx["Context"],
	// 	"Identity":     lambdaCtx["Identity"],
	// 	"ClientApp":    lambdaCtx["ClientApp"],
	// 	"EventPayload": evt,
	// }

	outputData := make(map[string]interface{})
	outputData["Function"] = lambdaCtx["Function"]
	outputData["Context"] = lambdaCtx["Context"]
	outputData["Identity"] = lambdaCtx["Identity"]
	outputData["ClientApp"] = lambdaCtx["ClientApp"]
	outputData["EventPayload"] = evt

	gContext := context.Background()
	if trace.Enabled() {
		logger.Debug("Tracing is enabled")
		carrier, ok := getLambdaCarrier(lambdaCtx)
		if ok {
			tracingContext, _ := trace.GetTracer().Extract(trace.Lambda, carrier)
			gContext = trace.AppendTracingContext(gContext, tracingContext)
		}
	}
	results, err := handler.Handle(gContext, outputData)

	rs, _ := json.Marshal(results)
	logger.Debugf("Handle results: %s \n", string(rs))

	var replyData interface{}

	if len(results) != 0 {
		dataAttr, ok := results["data"]
		if ok {
			switch dataAttr.(type) {
			case map[string]interface{}:
				complexData, err := coerce.ToObject(dataAttr)
				if err != nil {
					logger.Debugf("Lambda Trigger convert to complex object error: %s", err.Error())
					return nil, err
				}
				replyData = complexData
			default:
			}
		}
	}

	if err != nil {
		logger.Debugf("Lambda Trigger Error: %s", err.Error())
		return nil, err
	}

	rs, _ = json.Marshal(replyData)
	logger.Debugf("Reply data: %s \n", string(rs))

	return replyData, err
}

// Start ...
func (t *MyTrigger) Start() error {
	return nil
}

// Stop implements util.Managed.Stop
func (t *MyTrigger) Stop() error {
	return nil
}

func getLambdaCarrier(lambdaCtx map[string]interface{}) (map[string]string, bool) {
	m := make(map[string]string)
	const tracingContext string = "TracingContext"
	const xrayHTTPHeader string = "X-Amzn-Trace-Id"
	lc, ok := lambdaCtx["Context"].(map[string]interface{})
	if !ok {
		return m, false
	}
	if lc[tracingContext] != nil {
		tc, ok := lc[tracingContext].(map[string]interface{})
		if ok {
			m[xrayHTTPHeader] = tc[xrayHTTPHeader].(string)
			return m, true
		}
	}
	return m, false
}
