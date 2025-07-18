package reply

import (
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
)

type ReplyActivity struct {
}

var activityMd = activity.ToMetadata(&Input{})

func init() {
	_ = activity.Register(&ReplyActivity{}, New)
}

// New creates new instance of RESTActivity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &ReplyActivity{}, nil
}

// Metadata returns the activity's metadata
func (a *ReplyActivity) Metadata() *activity.Metadata {
	return activityMd
}

type Input struct {
	Code    int                    `md:"code"`
	Message string                 `md:"message"`
	Data    map[string]interface{} `md:"data"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":    i.Code,
		"message": i.Message,
		"data":    i.Data,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	i.Message, _ = coerce.ToString(values["message"])
	i.Code, _ = coerce.ToInt(values["code"])
	i.Data, _ = coerce.ToObject(values["data"])
	return nil
}

// Eval implements api.Activity.Eval - Invokes a REST Operation
func (a *ReplyActivity) Eval(context activity.Context) (done bool, err error) {

	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	code := input.Code

	var replyData interface{}

	replyData = input.Data

	context.Logger().Infof("Reply data: %s", replyData)

	if code == 500 {

		message := input.Message

		if message != "" {
			replyData = message
		}
	}

	actionCtx := context.ActivityHost()

	//todo support replying with error
	if replyData != nil {
		actionCtx.Reply(nil, nil)
	} else {
		actionCtx.Reply(replyData.(map[string]interface{}), nil)
	}

	return true, nil
}
