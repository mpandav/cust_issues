package stateful

import (
	"context"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"github.com/project-flogo/flow/instance"
	"github.com/project-flogo/flow/support"
)

const (
	RefFlow = "github.com/project-flogo/flow"
)

// RequestProcessor processes request objects and invokes the corresponding
// flow Manager methods
type RequestProcessor struct {
	runner action.Runner
	logger log.Logger
}

type RestartActivityRequest struct {
	FlowInstanceId string                 `json:"flowInstanceId"`
	TaskStepId     string                 `json:"taskStepId"`
	TaskName       string                 `json:"taskName"`
	Inputs         map[string]interface{} `json:"inputs"`
}

type RestartFlowRequest struct {
	FlowInstanceId string                 `json:"flowInstanceId"`
	Inputs         map[string]interface{} `json:"inputs"`
	FlowURI        string                 `json:"flowURI"`
}

// RestartRequest describes a request for restarting a FlowInstance
// todo: can be merged into StartRequest
type RestartRequest struct {
	InitialState *instance.IndependentInstance `json:"initialState"`
	Data         map[string]interface{}        `json:"data"`
	Interceptor  *support.Interceptor          `json:"interceptor"`
	Patch        *support.Patch                `json:"patch"`
	ReturnID     bool                          `json:"returnId"`
}

// RestartFlow handles a RestartRequest for a FlowInstance.  This will
// generate an ID for the new FlowInstance and queue a RestartRequest.
func (rp *RequestProcessor) RestartActivity(restartRequest *RestartRequest, flowID string, taskId int) (results map[string]interface{}, err error) {

	logger := rp.logger

	logger.Debugf("Tester restarting flow")

	//todo share action, for now add flowUri to settings
	settings := map[string]interface{}{"flowURI": restartRequest.InitialState.FlowURI()}

	factory := action.GetFactory(RefFlow)
	act, err := factory.New(&action.Config{Settings: settings})
	if err != nil {
		return nil, err
	}

	inputs := make(map[string]interface{}, len(restartRequest.Data)+1)

	if restartRequest.Data != nil {

		logger.Debugf("Updating flow attrs: %v", restartRequest.Data)

		for k, v := range restartRequest.Data {
			//attr, _ := data.NewAttribute(k, data.TypeAny, v)
			inputs[k] = v
		}
	}

	execOptions := &instance.ExecOptions{Interceptor: restartRequest.Interceptor, Patch: restartRequest.Patch}
	ro := &instance.RunOptions{PreservedInstanceId: flowID, Op: instance.OpRestart, ReturnID: restartRequest.ReturnID, FlowURI: restartRequest.InitialState.FlowURI(), InitStepId: restartRequest.InitialState.StepID(), InitialState: restartRequest.InitialState, ExecOptions: execOptions, Rerun: true}
	//attr, _ := data.NewAttribute("_run_options", data.TypeAny, ro)
	ro.InitStepId = taskId

	inputs["_run_options"] = ro

	return rp.runner.RunAction(trigger.NewContextWithEventId(context.Background(), "replay_event_id"), act, inputs)
}

func (rp *RequestProcessor) RestartFlow(startRequest *RestartFlowRequest) (results map[string]interface{}, err error) {

	logger := rp.logger

	logger.Debugf("stateful starting flow")

	factory := action.GetFactory(RefFlow)

	settings := map[string]interface{}{"flowURI": startRequest.FlowURI}
	act, err := factory.New(&action.Config{Settings: settings})
	if err != nil {
		return nil, err
	}

	if startRequest.Inputs == nil {
		startRequest.Inputs = make(map[string]interface{})
	}

	logger.Infof("Flow instance executed from Execution History re-run option using Execution Id(Flow Instance Id): %s", startRequest.FlowInstanceId)
	execOptions := &instance.ExecOptions{}
	ro := &instance.RunOptions{Op: instance.OpStart, ReturnID: true, FlowURI: startRequest.FlowURI, ExecOptions: execOptions, OriginalInstanceId: startRequest.FlowInstanceId}
	startRequest.Inputs["_run_options"] = ro
	return rp.runner.RunAction(trigger.NewContextWithEventId(context.Background(), "replay_event_id"), act, startRequest.Inputs)
}
