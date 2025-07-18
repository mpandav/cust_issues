package log

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	LogLevel      string `md:"Log Level"`
	Message       string `md:"message"`
	FlowInfo      bool   `md:"flowInfo"`
	LogLevelInput string `md:"logLevel"`
}

const (
	ivLogLevel      = "Log Level"
	ivMessage       = "message"
	ivFlowInfo      = "flowInfo"
	ivLogLevelInput = "logLevel"
)

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		ivLogLevel:      i.LogLevel,
		ivMessage:       i.Message,
		ivFlowInfo:      i.FlowInfo,
		ivLogLevelInput: i.LogLevelInput,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	i.LogLevel, _ = coerce.ToString(values[ivLogLevel])
	if i.LogLevel == "" {
		i.LogLevel = "INFO"
	}
	i.Message, _ = coerce.ToString(values[ivMessage])
	i.FlowInfo, _ = coerce.ToBool(values[ivFlowInfo])
	i.LogLevelInput, _ = coerce.ToString(values[ivLogLevelInput])
	return nil
}
