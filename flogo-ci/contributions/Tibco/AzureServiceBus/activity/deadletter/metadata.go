package deadletter

import (
	"github.com/project-flogo/core/data/coerce"
)

type Input struct {
	DeadLetterReason      string `md:"reason"`
	DeadLetterDescription string `md:"description"`
}

type Output struct {
	Output map[string]interface{} `md:"output"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"reason":      i.DeadLetterReason,
		"description": i.DeadLetterDescription,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.DeadLetterReason, err = coerce.ToString(values["reason"])
	if err != nil {
		return err
	}
	i.DeadLetterDescription, err = coerce.ToString(values["description"])
	if err != nil {
		return err
	}
	return err
}
