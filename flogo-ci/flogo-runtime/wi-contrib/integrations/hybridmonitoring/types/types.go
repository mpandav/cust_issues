package types

type AppEventType string

const (
	ENGINE_START   AppEventType = "engineStart"
	ENGINE_STOP    AppEventType = "engineStop"
	METRICS_UPDATE AppEventType = "metricsUpdate"

	MonConfig                   = "FLOGO_APP_MON_HYBRID_CONFIG"
	FlogoHttpServicePort        = "FLOGO_HTTP_SERVICE_PORT"
	AGENTPATH            string = "/v1/delta/appengines"
	AGENT_REGISTRATION   string = "/v1/apps"

	TOP_CMD_LINUX = "top -b -n1 | grep -i cpu | head -1 |  awk ' {print $2  \" \" $4}'"

	TOP_CMD_MAC = "top -l1 | grep -i cpu | head -1 |  awk ' {print $3 \" \" $5}'"
)

type AppEngineMetrics struct {
	EventType AppEventType             `json:"eventType"`
	Entries   []*AppEngineMetricsEntry `json:"entries"`
}

type AppEngineMetricsEntry struct {
	Gsbc                   string           `json:"gsbc"`
	AppType                string           `json:"appType"`
	AppId                  string           `json:"appId"`
	InstanceId             string           `json:"instanceId"`
	TimeStamp              int64            `json:"timestamp"`
	ExecutionSinceLastPush *AppExecutionSum `json:"executionSinceLastPush"`
	ExecutionTotal         *AppExecutionSum `json:"executionTotal"`
	Uptime                 float32          `json:"uptime"`
	Cpu                    float32          `json:"cpu"`
	Mem                    float32          `json:"memory"`
	UsedMemory             float32          `json:usedMemory"`
	TotalMemory            float32          `json:totalMemory"`
}

type AppExecutionSum struct {
	SuccessNum int64 `json:"successNum"`
	FailureNum int64 `json:"failureNum"`
}

type EngineMetricsEntry struct {
	Cpu         float32 `json:"cpu"`
	Mem         float32 `json:"mem"`
	UsedMemory  float32 `json:usedMemory"`
	TotalMemory float32 `json:totalMemory"`
}

type EngineUpdateResponse struct {
	NextPushInterval int `json:"nextPushInterval"`
}

type AppEngineSum struct {
	AppType    string `json:"appType"`
	SuccessNum int64  `json:"successNum"`
	FailureNum int64  `json:"failureNum"`
}

type ErrorResponse struct {
	Code        int    `json:"code"`
	ErrorMsg    string `json:"errMsg"`
	ErrorDetail string `json:"errDetail"`
}

type HybridAgentConfig struct {
	HybridAgentHost string `json:"host"`
	HybridAgentPort string `json:"port"`
	AppHost         string `json:"appHost"`
}
type RegistratrationRequest struct {
	AppID          string `json:"appId"`
	AppType        string `json:"appType"`
	AppName        string `json:"appName"`
	Description    string `json:"description"`
	InstanceID     string `json:"instanceId"`
	Version        string `json:"version"`
	InstrumentURL  string `json:"instrumentUrl"`
	AppEndpoint    string `json:"appEndpoint"`
	ConfigEndpoint string `json:"configEndpoint"`
	LoggerEndpoint string `json:"loggerEndpoint"`
}

type RegistrationResponse struct {
	Msg string `json:"msg"`
}
