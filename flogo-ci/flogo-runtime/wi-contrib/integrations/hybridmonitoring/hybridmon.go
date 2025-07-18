package hybridmonitoring

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	core "github.com/project-flogo/core/engine/event"
	"github.com/project-flogo/core/support/log"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/tibco/wi-contrib/httpservice"
	"github.com/tibco/wi-contrib/integrations/hybridmonitoring/types"
	"github.com/tibco/wi-contrib/integrations/monitoring"
	"github.com/tibco/wi-contrib/metrics"
)

type AppEventListener struct {
}

type HybridMonManaged struct {
}

var statsLogger = log.ChildLogger(log.RootLogger(), "hybrid.monitoring")

var startedAt int64
var DefaultDispatchInterval types.AtomicNumeric
var successTotalLastPush types.AtomicNumeric
var failureTotalLastPush types.AtomicNumeric
var cpuLastPush types.AtomicNumeric
var memLastPush types.AtomicNumeric
var enabled bool

const (
	HybridAgentHost = "TCI_HYBRID_AGENT_HOST"
	HybridAgentPort = "TCI_HYBRID_AGENT_PORT"
	AppHostOverride = "FLOGO_APP_MON_INSTANCE_NAME"
)

var eventTypes = []string{app.AppEventType}

var agentConfig = &types.HybridAgentConfig{}

type HybridAgentConfig struct {
	HybridAgentHost string `json:"hybridAgentHost"`
	HybridAgentPort string `json:"hybridAgentPort"`
}

func InitAtomics() {
	startedAt = time.Now().Unix()
	DefaultDispatchInterval.Set(int(60))
	successTotalLastPush.Set(int64(0))
	failureTotalLastPush.Set(int64(0))
	cpuLastPush.Set(float32(0))
	memLastPush.Set(float32(0))
}

func init() {
	httpservice.RegisterHandler("/app/tci/metrics", enableTCIMetricsCollectorHandler())
	agentHost, agentHostOk := os.LookupEnv(HybridAgentHost)
	//If Mon File or Service Port is not set then no need to go ahead. Return.
	if !agentHostOk {
		//statsLogger.Debug("No Monitoring Service configuration or HTTP Service port specified. Refer documentation for more details.")
		return
	}

	_, metricsPortOk := os.LookupEnv(types.FlogoHttpServicePort)
	if !metricsPortOk {
		statsLogger.Errorf("Hybrid monitoring is enabled but metrics collection port is not set. Configure port using [%s].", monitoring.FlogoHttpServicePort)
		return
	}

	agentPort, agentPortOk := os.LookupEnv(HybridAgentPort)
	if !agentPortOk {
		statsLogger.Errorf("Hybrid monitoring is enabled but agent port is not set. Configure port using [%s].", HybridAgentPort)
		return
	}

	agentConfig.HybridAgentHost = agentHost
	agentConfig.HybridAgentPort = agentPort
	appHost, ok := os.LookupEnv(AppHostOverride)
	if ok {
		agentConfig.AppHost = appHost
	}

	// Collect app metrics
	_ = metrics.RegisterEventListener()
	InitAtomics()
	engine.LifeCycle(&HybridMonManaged{})
	err := core.RegisterListener("appevent", &AppEventListener{}, eventTypes)
	if err != nil {
		statsLogger.Error(err)
		return
	}
	enabled = true
}

func enableTCIMetricsCollectorHandler() http.Handler {
	return http.HandlerFunc(enableTCIMetricsCollector)
}

func enableTCIMetricsCollector(w http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		if !enabled {
			var hybridAgent *HybridAgentConfig
			err := json.NewDecoder(request.Body).Decode(&hybridAgent)
			if err != nil {
				statsLogger.Errorf("Failed to read Hybrid Agent configuration due to error: %s", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			agentConfig.HybridAgentHost = hybridAgent.HybridAgentHost
			agentConfig.HybridAgentPort = hybridAgent.HybridAgentPort
			// Enable app metrics collection
			_ = metrics.RegisterEventListener()
			InitAtomics()
			engine.LifeCycle(&HybridMonManaged{})
			regApp := true
			if request.URL.Query().Get("registerApp") != "" {
				regApp, _ = coerce.ToBool(request.URL.Query().Get("registerApp"))
			}
			err = registerAppMonitoring(regApp)
			if err == nil {
				go startMetricsPollerDispatcher()
			}
			// Enabled TCI execution stats collection
			metrics.EnableTCIMetricsCollection()
			enabled = true
		}
		statsLogger.Info("TCI Metrics collection is enabled")
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (ls *AppEventListener) HandleEvent(evt *core.Context) error {

	switch t := evt.GetEvent().(type) {
	case app.AppEvent:
		switch t.AppStatus() {
		case app.STARTED:
			err := registerAppMonitoring(true)
			if err == nil {
				go startMetricsPollerDispatcher()
			}
		}
	}

	return nil
}

func (ls *HybridMonManaged) Start() error {
	return nil
}

func (ls *HybridMonManaged) Stop() error {
	statsLogger.Debug("Application Stopped. De-registering application....")
	unregisterAppMonitoring()
	return nil

}

func registerAppMonitoring(registerApp bool) error {
	var err error
	if registerApp {
		err = register()
	}
	if err == nil {
		DispatchMetrics(types.ENGINE_START)
	}
	return err
}

func unregisterAppMonitoring() {
	DispatchMetrics(types.ENGINE_STOP)
}

func register() error {

	statsLogger.Debug("Registering application")

	appInfo := metrics.GetAppInfoLocal()

	regRequest := &types.RegistratrationRequest{
		AppID:          appInfo.AppID,
		AppType:        appInfo.AppType,
		AppName:        appInfo.AppName,
		Description:    appInfo.Description,
		InstanceID:     appInfo.InstanceID,
		Version:        appInfo.Version,
		InstrumentURL:  appInfo.InstrumentURL,
		AppEndpoint:    appInfo.AppEndPoint,
		ConfigEndpoint: appInfo.ConfigEndPoint,
		LoggerEndpoint: appInfo.LoggerEndPoint,
	}

	if agentConfig.AppHost != "" {
		regRequest.InstanceID = agentConfig.AppHost
	}

	_, err := registerApp(regRequest)

	if err != nil {
		statsLogger.Errorf("Application registration failed due to error - %v", err)
		//os.Exit(1)
	} else {
		statsLogger.Info("Application registered successfully")
	}

	return err

}

func startMetricsPollerDispatcher() {
	for {
		time.Sleep(time.Duration(DefaultDispatchInterval.Get().(int)) * time.Second)
		// send engine metrics update event
		DispatchMetrics(types.METRICS_UPDATE)
	}
}

func DispatchMetrics(eventType types.AppEventType) {

	ef := CollectAppEngineMetrics()
	mf := CollectExecMetrics()

	successNumDelta := ef.SuccessNum - successTotalLastPush.Get().(int64)
	failureNumDelta := ef.FailureNum - failureTotalLastPush.Get().(int64)

	appInfo := metrics.GetAppInfoLocal()
	appType := appInfo.AppType
	appId := appInfo.AppID
	instanceId := appInfo.InstanceID
	if agentConfig.AppHost != "" {
		instanceId = agentConfig.AppHost
	}
	appExecutionDelta := &types.AppExecutionSum{
		SuccessNum: successNumDelta,
		FailureNum: failureNumDelta,
	}
	appExecutionTotal := &types.AppExecutionSum{
		SuccessNum: ef.SuccessNum,
		FailureNum: ef.FailureNum,
	}
	uptime := float32(time.Now().Unix()-startedAt) / 60

	mEntry := &types.AppEngineMetricsEntry{
		Gsbc:                   "",
		AppType:                appType,
		AppId:                  appId,
		InstanceId:             instanceId,
		TimeStamp:              time.Now().Unix(),
		ExecutionSinceLastPush: appExecutionDelta,
		ExecutionTotal:         appExecutionTotal,
		Uptime:                 uptime,
		Cpu:                    mf.Cpu,
		Mem:                    mf.Mem,
		UsedMemory:             mf.UsedMemory,
		TotalMemory:            mf.TotalMemory,
	}
	appMetrics := &types.AppEngineMetrics{
		EventType: eventType,
		Entries:   []*types.AppEngineMetricsEntry{mEntry},
	}

	b, _ := json.Marshal(appMetrics)
	statsLogger.Debugf("Push app metrics to metrics server: %s\n", string(b))
	nextPushInterval, err := dispatchAppEngineMetrics(appMetrics)

	if err != nil {
		statsLogger.Warn(err.Error())
		return
	}

	successTotalLastPush.Set(ef.SuccessNum)
	failureTotalLastPush.Set(ef.FailureNum)
	cpuLastPush.Set(mf.Cpu)
	memLastPush.Set(mf.Mem)

	DefaultDispatchInterval.Set(nextPushInterval)

}

func CollectAppEngineMetrics() *types.AppEngineSum {
	response := metrics.LocalMetricsUpdate()
	engineSum := &types.AppEngineSum{}
	engineSum.AppType = response.AppType
	engineSum.FailureNum = int64(response.FailureNum)
	engineSum.SuccessNum = int64(response.SuccessNum)
	return engineSum

}
func CollectExecMetrics() *types.EngineMetricsEntry {

	execSum := &types.EngineMetricsEntry{}
	memory, _ := mem.VirtualMemory()
	totalCPU := getCPU()
	execSum.Mem = float32(memory.UsedPercent)
	execSum.Cpu = totalCPU
	execSum.UsedMemory = float32(memory.Used)
	execSum.TotalMemory = float32(memory.Total)
	return execSum
}

func getCPU() float32 {
	var stderr bytes.Buffer
	command := ""
	if runtime.GOOS == "linux" {
		command = types.TOP_CMD_LINUX
	} else if runtime.GOOS == "darwin" {
		command = types.TOP_CMD_MAC
	} else {
		getCPUFromLibrary()
	}
	cmd := exec.Command("sh", "-c", command)
	cmd.Stderr = &stderr
	topResponse, err := cmd.Output()
	top := string(topResponse)
	statsLogger.Debug(top)
	statsLogger.Debug(stderr.String())
	if err != nil {
		statsLogger.Error("Failed to execute command", err)
		getCPUFromLibrary()
	} else {
		if top == "" {
			return getCPUFromLibrary()
		}
		var totalCPU float32 = 0.0
		totalCPU, err = extractCPU(top)
		if err != nil {
			return getCPUFromLibrary()
		}
		statsLogger.Debug(totalCPU)
		return totalCPU
	}
	return 0
}
