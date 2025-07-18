package eventflowcontrol

import (
	"errors"
	"fmt"
	"os"

	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/support/log"
)

var logger = log.ChildLogger(log.RootLogger(), "eventflowcontroller")

var registry = make(map[string]EventFlowController)

type EventFlowController interface {
	EventFlowControlEnabled() bool
}

func RegisterFlowController(name string, controller EventFlowController) error {
	_, found := registry[name]
	if found {
		msg := fmt.Sprintf("Evaluator with name '%s' already registered", name)
		logger.Errorf(msg)
		return errors.New(msg)
	}
	registry[name] = controller
	logger.Infof("Flow controller [%s] registered", name)
	_, set := os.LookupEnv(app.EnvKeyEnableFlowControl)
	if !set {
		_ = os.Setenv(app.EnvKeyEnableFlowControl, "true")
	}
	return nil
}

func FlowControlled() bool {
	isBusy := false
	for _, controller := range registry {
		isBusy = isBusy || controller.EventFlowControlEnabled()
	}
	return isBusy
}

func StartControl() error {
	if FlowControlled() {
		engineController := app.GetEventFlowController()
		if engineController != nil {
			err := engineController.StartControl()
			if err != nil {
				logger.Errorf("Failed to pause triggers. Error:%s", err.Error())
				return err
			}
		}
	}

	return nil
}

func ReleaseControl() error {
	if !FlowControlled() {
		engineController := app.GetEventFlowController()
		if engineController != nil {
			err := engineController.ReleaseControl()
			if err != nil {
				logger.Errorf("Failed to resume triggers. Error:%s", err.Error())
				return err
			}
		}
	}
	return nil
}
