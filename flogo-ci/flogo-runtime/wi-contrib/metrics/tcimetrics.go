package metrics

import (
	"encoding/json"
	"net/http"
	"sync"

	core "github.com/project-flogo/core/engine/event"
	"github.com/project-flogo/core/support/log"
	flow "github.com/project-flogo/flow/support/event"
	"github.com/tibco/wi-contrib/environment"
	"github.com/tibco/wi-contrib/httpservice"
)

var tciMatrixLogger = log.ChildLogger(log.RootLogger(), "tci-app.metrics")

type Response struct {
	AppType    string     `json:"appType"`
	SuccessNum int        `json:"successNum"`
	FailureNum int        `json:"failureNum"`
	lock       sync.Mutex `json:"-"`
}

var tcoFlowQueue = make(chan *core.Context, 10)

//var excutionLastPush *ExecutionLastPush
var response = &Response{AppType: "flogo", lock: sync.Mutex{}}

type tciFlowEventListener struct {
}

func (ls *tciFlowEventListener) HandleEvent(evt *core.Context) error {
	tcoFlowQueue <- evt
	return nil
}

func init() {
	if (environment.IsTCIEnv() || environment.IsEnvHybridMon()) && !environment.IsTesterEnv() {
		EnableTCIMetricsCollection()
	}
}

func EnableTCIMetricsCollection() {
	err := core.RegisterListener("tci-flow-metrics-collector", &tciFlowEventListener{}, []string{flow.FlowEventType})
	if err != nil {
		statsLogger.Errorf("Failed to enable tci-flow-metrics-collector due to error - '%v'", err)
	}
	go handleTCIFlowEvent()
	httpservice.RegisterHandler("/tci/app/metricsUpdate", metricsUpdate())
}

func handleTCIFlowEvent() {
	for {
		select {
		case event := <-tcoFlowQueue:
			switch t := event.GetEvent().(type) {
			case flow.FlowEvent:
				switch t.FlowStatus() {
				case flow.COMPLETED:
					response.lock.Lock()
					response.SuccessNum++
					response.lock.Unlock()
				case flow.FAILED:
					response.lock.Lock()
					response.FailureNum++
					response.lock.Unlock()
				}
			}
		}
	}
}

func metricsUpdate() http.Handler {
	return http.HandlerFunc(collectAppMatricsUpdate)
}

func LocalMetricsUpdate() *Response {
	return response
}

func collectAppMatricsUpdate(w http.ResponseWriter, req *http.Request) {
	//excutionLastPush.lock.Lock()
	//
	//matricEntry := &entry{
	//	Gsbc:       environment.GetTCISubscriptionId(),
	//	AppType:    "flogo",
	//	AppId:      environment.GetTCIAppId(),
	//	InstanceId: environment.GettCIContainerId(),
	//	TimeStamp:  time.Now().Unix(),
	//	ExecutionLastPush: &ExecutionLastPush{
	//		SuccessNum: excutionLastPush.SuccessNum,
	//		FailureNum: excutionLastPush.FailureNum,
	//	},
	//}
	//excutionLastPush.SuccessNum = 0
	//excutionLastPush.FailureNum = 0
	//excutionLastPush.lock.Unlock()
	//
	//entries := make([]*entry, 1)
	//entries[0] = matricEntry
	//response := &updateMatricsResponse{EventType: "metricsUpdate", Entries: entries}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(200)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		tciMatrixLogger.Error(err)
	}
	return
}
