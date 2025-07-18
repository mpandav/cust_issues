package metrics

import (
	"encoding/json"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"os"
	"strconv"
	"strings"
	"time"
)

var logStatsLogger = log.ChildLogger(log.RootLogger(), "app.metrics")

const (
	EnableStatsLogEmitter = "FLOGO_APP_METRICS_LOG_EMITTER_ENABLE"
	StatLogEmmiterConfig  = "FLOGO_APP_METRICS_LOG_EMITTER_CONFIG"
)

type statConfig struct {
	Interval string   `json:"interval"`
	Type     []string `json:"type"`
}

func init() {

	if enableStatsLogListener() {
		// Listener is enabled through env var
		err := RegisterEventListener()
		if err != nil {
			statsLogger.Errorf("Failed to enable metrics collection due to error - '%v'", err)
		} else {
			if strings.EqualFold(os.Getenv(StatLogEmmiterConfig), "") {
				//means not this ENV is not provided
				statsLogger.Info("No App Metrics Configuration Provided, Setting by default at 30s Interval for flow and activity metrics")
				go ShowStatsAsLog("30s", []string{"activity", "flow"})
			} else {
				conf, err := statLogEmmiterConfig()
				if err != nil {
					logStatsLogger.Errorf("Error with Stat Configuration %s", err)
					go ShowStatsAsLog("30s", []string{"activity", "flow"})
				} else {
					iVal := conf.(statConfig).Interval
					attrType := conf.(statConfig).Type

					//only interval
					if !strings.EqualFold(strings.TrimSpace(iVal), "") && len(attrType) == 0 {
						statsLogger.Info("No types found setting it by default to [\"flow\",\"activity\"]")
						go ShowStatsAsLog(conf.(statConfig).Interval, []string{"activity", "flow"})
					}

					//only types
					if strings.EqualFold(strings.TrimSpace(iVal), "") && len(attrType) != 0 {
						statsLogger.Info("No Interval found setting it by default to 30 seconds")
						go ShowStatsAsLog("30s", conf.(statConfig).Type)
					}

					//interval and types
					if !strings.EqualFold(strings.TrimSpace(iVal), "") && len(attrType) != 0 {
						go ShowStatsAsLog(conf.(statConfig).Interval, conf.(statConfig).Type)
					}
				}
			}
		}
	}
}

func ShowStatsAsLog(interval string, types []string) {
	t, err := time.ParseDuration(interval)
	if err != nil {
		statsLogger.Errorf("Error in Parsing Interval %s", err)
		panic(err)
	}
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	for {
		select {
		case _ = <-ticker.C:
			logStatsLogger.Infof(getStatsForLogging(types))
		}
	}
}

type metrics struct {
	AppName    string         `json:"app_name"`
	AppVersion string         `json:"app_version"`
	HostName   string         `json:"host_name"`
	Flows      []*FlowMetrics `json:"flows,omitempty"`
	Tasks      []*TaskMetrics `json:"activities,omitempty"`
}

func contains(types []string, data string) bool {

	for i := range types {
		if strings.EqualFold(types[i], data) {
			return true
		}
	}
	return false
}

func getStatsForLogging(types []string) string {
	host, _ := os.Hostname()
	data := metrics{AppName: engine.GetAppName(), AppVersion: engine.GetAppVersion(), HostName: host}

	for _, flowEntry := range flowMap {
		if contains(types, "flow") {
			if contains(types, "activity") {
				activities := flowEntry.tasks
				for _, taskEntry := range activities {
					data.Tasks = append(data.Tasks, taskEntry)
				}
			}
			data.Flows = append(data.Flows, flowEntry)
		} else if contains(types, "activity") {
			activities := flowEntry.tasks
			for _, taskEntry := range activities {
				data.Tasks = append(data.Tasks, taskEntry)
			}
		}
	}
	str, err := json.Marshal(data)

	if err != nil {
		return "NO METRICS-ERROR"
	}
	return string(str)
}

func enableStatsLogListener() bool {
	enableListener := os.Getenv(EnableStatsLogEmitter)
	if len(enableListener) == 0 {
		return false
	}
	b, _ := strconv.ParseBool(enableListener)
	return b
}

func statLogEmmiterConfig() (interface{}, error) {
	configurations := os.Getenv(StatLogEmmiterConfig)
	config := statConfig{}
	err := json.Unmarshal([]byte(configurations), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
