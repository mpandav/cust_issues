package eventflowcontrol

import (
	"os"
	"strconv"
	"sync"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	core "github.com/project-flogo/core/engine/event"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

var eventBasedFlowControlLogger = log.ChildLogger(logger, "events")

type EventBasedFlowController struct {
	enableFlowControl  bool
	totalEvents        int
	created, completed *SafeCounter
	name               string
}

type SafeCounter struct {
	mu    *sync.RWMutex
	count int
}

func (s *SafeCounter) Inc() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count++
}

func (s *SafeCounter) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count = 0
}

func (s *SafeCounter) Get() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.count
}

const (
	EventBasedFlowControl          = "FLOGO_FLOW_CONTROL_EVENTS"
	EventBasedFlowControlMaxEvents = "FLOGO_FLOW_CONTROL_MAX_EVENTS"
)

func init() {
	enabled := os.Getenv(EventBasedFlowControl)
	if enabled != "" {
		set, err := strconv.ParseBool(enabled)
		if err != nil {
			eventBasedFlowControlLogger.Errorf("Invalid configuration [%s=%s]. The value must be either true or false.", EventBasedFlowControl, enabled)
			return
		}
		if set {
			eventBasedFlowController := &EventBasedFlowController{
				enableFlowControl: false,
				name:              "EventBasedFlowController",
				created: &SafeCounter{
					mu: &sync.RWMutex{},
				},
				completed: &SafeCounter{
					mu: &sync.RWMutex{},
				},
			}
			eventBasedFlowController.totalEvents = getTotalEvents()
			eventBasedFlowControlLogger.Infof("The flow control is be enabled when total events received are equal or greater than the threshold [%d]. It is disabled when all events are processed.", eventBasedFlowController.totalEvents)
			_ = RegisterFlowController("events", eventBasedFlowController)
			err := core.RegisterListener(eventBasedFlowController.name, eventBasedFlowController, []string{trigger.TriggerEventType})
			if err != nil {
				eventBasedFlowControlLogger.Errorf("Failed to register event listener. Error:%s. Active instance based flow control feature is not enabled.", err.Error())
			}
		}
	}
}

func (i *EventBasedFlowController) Name() string {
	return i.name
}

func getTotalEvents() int {
	runnerType := engine.GetRunnerType()
	if runnerType == engine.ValueRunnerTypePooled {
		return engine.GetRunnerQueueSize()
	}

	toInt, _ := coerce.ToInt(os.Getenv(EventBasedFlowControlMaxEvents))
	if toInt > 0 {
		return toInt
	}
	return 500
}

func (i *EventBasedFlowController) HandleEvent(evt *core.Context) error {
	switch t := evt.GetEvent().(type) {
	case trigger.HandlerEvent:
		switch t.Status() {
		case trigger.STARTED:
			i.created.Inc()
			val := i.created.Get()
			if val >= i.totalEvents && !i.enableFlowControl {
				eventBasedFlowControlLogger.Infof("Total events received [%d] are equal or higher than the threshold [%d]. Enabling flow control.", val, i.totalEvents)
				i.enableFlowControl = true
				_ = StartControl()
			}
		case trigger.COMPLETED, trigger.FAILED:
			i.completed.Inc()
			val := i.created.Get() - i.completed.Get()
			if val == 0 && i.enableFlowControl {
				eventBasedFlowControlLogger.Infof("All received events are processed. Disabling flow control.")
				i.enableFlowControl = false
				i.created.Reset()
				i.completed.Reset()
				_ = ReleaseControl()
			}
		}
	}
	return nil
}

func (i *EventBasedFlowController) EventFlowControlEnabled() bool {
	return i.enableFlowControl
}
