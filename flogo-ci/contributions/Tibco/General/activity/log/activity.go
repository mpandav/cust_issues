package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
)

const loggerName = "flogo.general.activity.log"

var logger log.Logger

type Activity struct {
}

var activityMd = activity.ToMetadata(&Input{})

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

func init() {
	_ = activity.Register(&Activity{}, New)

}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *Activity) Eval(context activity.Context) (done bool, err error) {
	if logger == nil {
		if log.CtxLoggingEnabled() {
			var fields []log.Field
			fields = append(fields, log.FieldString("activity.name", context.Name()))
			fields = append(fields, log.FieldString("flow.name", context.ActivityHost().Name()))
			fields = append(fields, log.FieldString("flow.id", context.ActivityHost().ID()))
			fields = append(fields, log.FieldString("app.name", engine.GetAppName()))
			fields = append(fields, log.FieldString("app.version", engine.GetAppVersion()))
			if engine.GetEnvName() != "" {
				fields = append(fields, log.FieldString("deployment.environment", engine.GetEnvName()))
			}

			if context.GetTracingContext() != nil {
				fields = append(fields, log.FieldString("trace.id", context.GetTracingContext().TraceID()))
				fields = append(fields, log.FieldString("span.id", context.GetTracingContext().SpanID()))
			}
			logger = log.NewLoggerWithFields(loggerName, fields...)
		} else {
			logger = log.NewLogger(loggerName)
		}
		flogoLogActivityLogLevel := ""
		if v := os.Getenv("FLOGO_LOGACTIVITY_LOG_LEVEL"); v != "" {
			flogoLogActivityLogLevel = v
		}
		switch flogoLogActivityLogLevel {
		case "DEBUG":
			log.SetLogLevel(logger, log.DebugLevel)
		case "INFO":
			log.SetLogLevel(logger, log.InfoLevel)
		case "WARN":
			log.SetLogLevel(logger, log.WarnLevel)
		case "ERROR":
			log.SetLogLevel(logger, log.ErrorLevel)
		default:
			log.SetLogLevel(logger, log.DebugLevel)
		}
	}

	//mv := context.GetInput(ivMessage)
	activityName := context.Name()
	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	msg := input.Message

	if input.FlowInfo {
		msg = fmt.Sprintf("%s. FlowInstanceID [%s], Flow [%s], Activity [%s].", msg,
			context.ActivityHost().ID(), context.ActivityHost().Name(), activityName)
	}
	lLevel := strings.ToUpper(input.LogLevel)

	lLevelPriority := strings.ToUpper(input.LogLevelInput)
	//Prioritize logLevel from input over dropdown
	if lLevelPriority != "" {
		lLevel = lLevelPriority
	}

	switch lLevel {
	case "INFO":
		logger.Info(msg)
	case "DEBUG":
		logger.Debug(msg)
	case "ERROR":
		logger.Error(msg)
	case "WARN":
		logger.Warn(msg)
	default:
		return false, activity.NewError(fmt.Sprintf("Invalid Log level [%s] configured. Valid values=[INFO, DEBUG, ERROR, WARN].", lLevel), "", nil)
	}
	return true, nil
}
