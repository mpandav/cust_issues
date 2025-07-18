package metrics

import (
	"encoding/json"

	"github.com/project-flogo/core/engine"
	core "github.com/project-flogo/core/engine/event"
	"github.com/project-flogo/core/trigger"
	flow "github.com/project-flogo/flow/support/event"

	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/project-flogo/core/support/log"

	"github.com/tibco/wi-contrib/httpservice"
)

// Stats listener is disabled by default
var enabled = false

var statsLogger = log.ChildLogger(log.RootLogger(), "app-metrics-collector")
var flowMap map[string]*FlowMetrics
var triggerMap map[string]*TriggerMetrics
var instanceMap map[string]int64
var metricsListeners []AppMetricsListener

//  Event queue to collect events
var eventQueue = make(chan *core.Context, getQueueSize())
var eventTypes = []string{flow.FlowEventType, flow.TaskEventType, trigger.TriggerEventType}

const (
	EventQueueSizeKey    = "FLOGO_APP_METRICS_EVENT_QUEUE"
	EnableStatsListener  = "FLOGO_APP_METRICS"
	EnableMonListener    = "FLOGO_APP_MON_SERVICE_CONFIG"
	EnableHybridListener = "FLOGO_APP_MON_HYBRID_CONFIG"

	ListenerName = "app-metrics-collector"
)

// Stop channel to shutdown event handler GO routine
var stop = make(chan bool)

type AppMetricsListener interface {
	// Called when flow is started or finished
	FlowMetrics(flow *FlowMetrics, state flow.Status)
	// Called when task is started or finished
	TaskMetrics(task *TaskMetrics, state flow.Status)
}

type FlowEventListener struct {
}

type appMetrics struct {
	AppName    string            `json:"app_name"`
	AppVersion string            `json:"app_version"`
	Triggers   []*TriggerMetrics `json:"triggers,omitempty"`
	Flows      []*FlowMetrics    `json:"flows,omitempty"`
	Tasks      []*TaskMetrics    `json:"activities,omitempty"`
}

type commonStats struct {
	Created      uint32  `json:"started"`
	Completed    uint32  `json:"completed"`
	Failed       uint32  `json:"failed"`
	AvgExecTime  float64 `json:"avg_exec_time"`
	MinExecTime  float64 `json:"min_exec_time"`
	MaxExecTime  float64 `json:"max_exec_time"`
	totalTime    float64
	LastExecTime float64 `json:"-"`
	FlowName     string  `json:"flow_name"`
}
type FlowMetrics struct {
	commonStats
	tasks       map[string]*TaskMetrics
	instanceMap map[string]int64
	Activities  map[string]*TaskMetrics `json:"activities,omitempty"`
}
type TaskMetrics struct {
	commonStats
	TaskName string `json:"activity_name"`
}

type HandlerMetrics struct {
	Name      string            `json:"handler_name"`
	Config    map[string]string `json:"config,omitempty"`
	Created   uint32            `json:"started"`
	Completed uint32            `json:"completed"`
	Failed    uint32            `json:"failed"`
}

type TriggerMetrics struct {
	Name      string                     `json:"trigger_name"`
	Status    string                     `json:"status"`
	Created   uint32                     `json:"started"`
	Completed uint32                     `json:"completed"`
	Failed    uint32                     `json:"failed"`
	Handlers  map[string]*HandlerMetrics `json:"handlers,omitempty"`
}

func init() {
	httpservice.RegisterHandler("/app/metrics", statsHandler())
	httpservice.RegisterHandler("/app/metrics/flows", flowStatsHandler())
	httpservice.RegisterHandler("/app/metrics/triggers", flowStatsHandler())
	httpservice.RegisterHandler("/app/metrics/flow/", flowStatsHandler())
	if enableStatsListener() {
		// Listener is enabled through env var
		err := RegisterEventListener()
		if err != nil {
			statsLogger.Errorf("Failed to enable metrics collection due to error - '%v'", err)
		}
	}
}

func statsHandler() http.Handler {
	return http.HandlerFunc(enableDisableStatsCollection)
}
func flowStatsHandler() http.Handler {
	return http.HandlerFunc(getStats)
}

func enableStatsListener() bool {

	// Monitoring should be enabled if Monitoring config is provided or if
	//FLOGO_APP_METRICS property is set to true.

	enableListener := os.Getenv(EnableStatsListener)

	if len(enableListener) == 0 {
		return false
	}
	b, _ := strconv.ParseBool(enableListener)
	return b
}
func getQueueSize() int {
	queueSize := os.Getenv(EventQueueSizeKey)
	if len(queueSize) > 0 {
		i, err := strconv.Atoi(queueSize)
		if err == nil {
			return i
		}
	}
	//The size should better be twice of engine one
	return 2 * core.GetEventQueueSize()
}

func (ls *FlowEventListener) Name() string {
	return ListenerName
}

func (ls *FlowEventListener) HandleEvent(evt *core.Context) error {

	if len(eventQueue) == cap(eventQueue) {
		statsLogger.Warnf("Event queue is full. For better performance, increase queue size using '%s'. Current queue size - %d", EventQueueSizeKey, cap(eventQueue))
	}
	eventQueue <- evt
	statsLogger.Debugf("Event is added to the queue. Current Queue Size: %d", len(eventQueue))

	return nil
}

func handleEvent() {
	for {
		select {
		case evt := <-eventQueue:
			switch t := evt.GetEvent().(type) {
			case flow.FlowEvent:
				switch t.FlowStatus() {
				case flow.STARTED:
					flowEntry, ok := flowMap[t.FlowName()]
					if !ok {
						flowEntry = &FlowMetrics{}
						flowEntry.FlowName = t.FlowName()
						flowEntry.tasks = make(map[string]*TaskMetrics)
						flowEntry.instanceMap = make(map[string]int64)
						flowEntry.Created++
						flowMap[t.FlowName()] = flowEntry
					} else {
						flowEntry.Created++
					}
					instanceMap[t.FlowID()] = t.Time().UnixNano()

					// Send FlowMetrics to listeners
					for i := range metricsListeners {
						metricsListeners[i].FlowMetrics(flowEntry, t.FlowStatus())
					}

				case flow.COMPLETED, flow.FAILED:
					flowEntry, ok := flowMap[t.FlowName()]
					if ok {
						if t.FlowStatus() == flow.FAILED {
							flowEntry.Failed++
							startTime, ok := instanceMap[t.FlowID()]
							if ok {
								flowEntry.LastExecTime = float64(t.Time().UnixNano()-startTime) / float64(time.Millisecond)
							}
						}

						if t.FlowStatus() == flow.COMPLETED {
							flowEntry.Completed++
							startTime, ok := instanceMap[t.FlowID()]
							if ok {
								timeDiff := float64(t.Time().UnixNano()-startTime) / float64(time.Millisecond)
								// Roundup number to 3 digit precision
								timeDiff = math.Ceil(timeDiff*1000) / 1000
								if flowEntry.MaxExecTime < timeDiff {
									flowEntry.MaxExecTime = timeDiff
								}
								if flowEntry.MinExecTime == 0 || flowEntry.MinExecTime > timeDiff {
									flowEntry.MinExecTime = timeDiff
								}

								if flowEntry.totalTime >= math.MaxFloat64 {
									// Reset value to 0
									flowEntry.totalTime = 0
								}
								flowEntry.totalTime = flowEntry.totalTime + timeDiff
								flowEntry.LastExecTime = timeDiff
							}
							avg := flowEntry.totalTime / float64(flowEntry.Completed)
							// Roundup number to 3 digit precision
							flowEntry.AvgExecTime = math.Ceil(avg*1000) / 1000
						}
						// Send FlowMetrics to listeners
						for i := range metricsListeners {
							metricsListeners[i].FlowMetrics(flowEntry, t.FlowStatus())
						}

						delete(instanceMap, t.FlowID())
					}
				}
			case flow.TaskEvent:
				switch t.TaskStatus() {
				case flow.STARTED:
					flowStats, ok := flowMap[t.FlowName()]
					if ok {
						te, ok := flowStats.tasks[t.TaskName()]
						if !ok {
							te = &TaskMetrics{}
							te.Created++
							te.TaskName = t.TaskName()
							te.FlowName = t.FlowName()
							flowStats.tasks[t.TaskName()] = te
						} else {
							te.Created++
						}
						flowStats.instanceMap[t.FlowID()] = t.Time().UnixNano()
						// Send TaskMetrics to listeners
						for i := range metricsListeners {
							metricsListeners[i].TaskMetrics(te, t.TaskStatus())
						}
					}

				case flow.COMPLETED, flow.FAILED:
					flowStats, ok := flowMap[t.FlowName()]
					if ok {
						taskEntry, ok := flowStats.tasks[t.TaskName()]
						if ok {
							if t.TaskStatus() == flow.FAILED {
								taskEntry.Failed++
								startTime, ok := flowStats.instanceMap[t.FlowID()]
								if ok {
									taskEntry.LastExecTime = float64(t.Time().UnixNano()-startTime) / float64(time.Millisecond)
								}
							}

							if t.TaskStatus() == flow.COMPLETED {
								taskEntry.Completed++
								startTime, ok := flowStats.instanceMap[t.FlowID()]
								if ok {
									timeDiff := float64(t.Time().UnixNano()-startTime) / float64(time.Millisecond)
									// Roundup number to 3 digit precision
									timeDiff = math.Ceil(timeDiff*1000) / 1000
									if taskEntry.MaxExecTime < timeDiff {
										taskEntry.MaxExecTime = timeDiff
									}

									if taskEntry.MinExecTime == 0 || taskEntry.MinExecTime > timeDiff {
										taskEntry.MinExecTime = timeDiff
									}

									if taskEntry.totalTime >= math.MaxInt64 {
										// Reset value to 0
										taskEntry.totalTime = 0
									}
									taskEntry.totalTime = taskEntry.totalTime + timeDiff
									taskEntry.LastExecTime = timeDiff
								}
								avg := taskEntry.totalTime / float64(taskEntry.Completed)
								taskEntry.AvgExecTime = math.Ceil(avg*1000) / 1000
							}

							// Send TaskMetrics to listeners
							for i := range metricsListeners {
								metricsListeners[i].TaskMetrics(taskEntry, t.TaskStatus())
							}

							delete(flowStats.instanceMap, t.FlowID())
						}
					}
				}
			case trigger.TriggerEvent:
				switch t.Status() {
				case trigger.INITIALIZING:
					_, ok := triggerMap[t.Name()]
					if !ok {
						triggerMap[t.Name()] = &TriggerMetrics{Name: t.Name(), Status: t.Status().String(), Handlers: make(map[string]*HandlerMetrics)}
					}
				case trigger.INITIALIZED, trigger.INIT_FAILED, trigger.STARTED, trigger.STOPPED, trigger.FAILED:
					tm, ok := triggerMap[t.Name()]
					if ok {
						tm.Status = t.Status().String()
					}
				}
			case trigger.HandlerEvent:
				switch t.Status() {
				case trigger.INITIALIZED:
					te, ok := triggerMap[t.TriggerName()]
					if ok {
						// This event expected to be published by the trigger implementation.
						hm := &HandlerMetrics{}
						hm.Name = t.HandlerName()
						hm.Config = t.Tags()
						te.Handlers[t.HandlerName()] = hm
					}
				case trigger.STARTED:
					te, ok := triggerMap[t.TriggerName()]
					if ok {
						he, found := te.Handlers[t.HandlerName()]
						if found {
							he.Created++
							te.Created++
						} else {
							// In case init event is not published by the trigger implementation
							hm := &HandlerMetrics{}
							hm.Name = t.HandlerName()
							hm.Config = t.Tags()
							te.Handlers[t.HandlerName()] = hm
							hm.Created++
							te.Created++
						}
					} else {
						if !ok {
							triggerMap[t.TriggerName()] = &TriggerMetrics{Name: t.TriggerName(), Status: t.Status().String(), Handlers: make(map[string]*HandlerMetrics)}
							te, _ := triggerMap[t.TriggerName()]
							hm := &HandlerMetrics{}
							hm.Name = t.HandlerName()
							hm.Config = t.Tags()
							te.Handlers[t.HandlerName()] = hm
							hm.Created++
							te.Created++
						}
					}
				case trigger.COMPLETED:
					te, ok := triggerMap[t.TriggerName()]
					if ok {
						he, found := te.Handlers[t.HandlerName()]
						if found {
							he.Completed++
							te.Completed++
						}
					}
				case trigger.FAILED:
					te, ok := triggerMap[t.TriggerName()]
					if ok {
						he, found := te.Handlers[t.HandlerName()]
						if found {
							he.Failed++
							te.Failed++
						}
					}
				}
			}
		case <-stop:
			statsLogger.Debugf("Shutting down event handler routine")
			return
		}
	}

}

func enableDisableStatsCollection(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodDelete:

		if !enabled {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// TODO Handle error
		core.UnRegisterListener(ListenerName, eventTypes)
		statsLogger.Info("Metrics collection is successfully stopped ")
		// stop goroutine
		stop <- true
		flowMap = nil
		instanceMap = nil
		w.WriteHeader(http.StatusOK)
		enabled = false
	case http.MethodGet:
		if !enabled {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data := appMetrics{AppName: engine.GetAppName(), AppVersion: engine.GetAppVersion()}
		for _, flowEntry := range flowMap {
			flowEntry.Activities = flowEntry.tasks
			data.Flows = append(data.Flows, flowEntry)
		}
		for _, tEntry := range triggerMap {
			data.Triggers = append(data.Triggers, tEntry)
		}

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			statsLogger.Errorf("Failed to return flow stats due to error - '%v'", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		if !enabled {
			err := RegisterEventListener()
			if err != nil {
				statsLogger.Errorf("Failed to enable metrics collection due to error - '%v'", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusConflict)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getStats(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		if !enabled {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		basePath, err := url.PathUnescape(req.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		data := appMetrics{AppName: engine.GetAppName(), AppVersion: engine.GetAppVersion()}
		if basePath == "/app/metrics/flows" {
			for _, flowEntry := range flowMap {
				// Reset activities
				flowEntry.Activities = nil
				data.Flows = append(data.Flows, flowEntry)
			}
			err := json.NewEncoder(w).Encode(data)
			if err != nil {
				statsLogger.Errorf("Failed to return flow stats due to error - '%v'", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if basePath == "/app/metrics/triggers" {
			for _, tEntry := range triggerMap {
				data.Triggers = append(data.Triggers, tEntry)
			}
			err := json.NewEncoder(w).Encode(data)
			if err != nil {
				statsLogger.Errorf("Failed to return flow stats due to error - '%v'", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			s := strings.Split(basePath, "/")
			len1 := len(s)
			if len1 == 5 {
				// return stats for single flow
				te, ok := flowMap[s[4]]
				if !ok {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				data.Flows = append(data.Flows, te)
				err := json.NewEncoder(w).Encode(data)
				if err != nil {
					statsLogger.Errorf("Failed to return flow stats due to error - '%v'", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else if len1 == 6 {

				if s[5] == "activities" {
					fe, ok := flowMap[s[4]]
					if !ok {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					for _, taskEntry := range fe.tasks {
						data.Tasks = append(data.Tasks, taskEntry)
					}
					err := json.NewEncoder(w).Encode(data)
					if err != nil {
						statsLogger.Errorf("Failed to return flow stats due to error - '%v'", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			w.WriteHeader(http.StatusOK)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func RegisterEventListener() error {
	if !enabled {
		flowMap = make(map[string]*FlowMetrics)
		triggerMap = make(map[string]*TriggerMetrics)
		instanceMap = make(map[string]int64)
		err := core.RegisterListener(ListenerName, &FlowEventListener{}, eventTypes)
		if err != nil {
			return err
		}
		// Start handler routine
		go handleEvent()
		statsLogger.Info("Metrics collection is successfully started")
		enabled = true
	}
	return nil
}

func RegisterAppMetricsListener(listener AppMetricsListener) {
	metricsListeners = append(metricsListeners, listener)
	if !enabled {
		// Enable stats collection
		err := RegisterEventListener()
		if err != nil {
			statsLogger.Errorf("Failed to enable metrics collection due to error - '%v'", err)
			panic("")
		}
	}
}
