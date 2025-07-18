package prometheus

import (
	"os"
	"strconv"
	"strings"

	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	flow "github.com/project-flogo/flow/support/event"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tibco/wi-contrib/httpservice"
	"github.com/tibco/wi-contrib/metrics"
)

const (
	EnablePromListener = "FLOGO_APP_METRICS_PROMETHEUS"
	PrometheusLabels   = "FLOGO_APP_METRICS_PROMETHEUS_LABELS"
)

var (
	activityMetricsCount        *prometheus.GaugeVec
	activityMetricsDurationMsec *prometheus.GaugeVec
	flowMetricsCount            *prometheus.GaugeVec
	flowMetricsDurationMsec     *prometheus.GaugeVec

	statsLogger  = log.ChildLogger(log.RootLogger(), "prometheus-metrics-collector")
	commonLabels = make(map[string]string)
)

func init() {
	enableProm, _ := strconv.ParseBool(os.Getenv(EnablePromListener))
	if !enableProm {
		return
	}

	metrics.RegisterAppMetricsListener(&PrometheusMetricsCollector{})
	httpservice.RegisterHandler("/metrics", promhttp.Handler())

	customLabels := os.Getenv(PrometheusLabels)
	if customLabels != "" {
		statsLogger.Infof("Custom Prometheus labels [%s] are set", customLabels)
		labels := strings.Split(customLabels, ",")
		for _, labelPair := range labels {
			attrList := strings.Split(labelPair, "=")
			if len(attrList) == 2 {
				key := strings.TrimSpace(attrList[0])
				value := strings.TrimSpace(attrList[1])
				if key != "" && value != "" {
					commonLabels[key] = value
				}
			}
		}
	}

	activityLabelNames := []string{
		"ApplicationName", "ApplicationVersion", "FlowName", "ActivityName", "State",
	}
	flowLabelNames := []string{
		"ApplicationName", "ApplicationVersion", "FlowName", "State",
	}
	if len(commonLabels) > 0 {
		for k := range commonLabels {
			activityLabelNames = append(activityLabelNames, k)
			flowLabelNames = append(flowLabelNames, k)
		}
	}

	activityMetricsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flogo_activity_execution_count",
		Help: "Total number of times the activity is started, completed or failed",
	}, activityLabelNames)

	activityMetricsDurationMsec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flogo_activity_execution_duration_msec",
		Help: "Total time(in ms) taken by the activity for successful completion or failure",
	}, activityLabelNames)

	flowMetricsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flogo_flow_execution_count",
		Help: "Total number of times the flow is started, completed or failed",
	}, flowLabelNames)

	flowMetricsDurationMsec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "flogo_flow_execution_duration_msec",
		Help: "Total time(in ms) taken by the flow for successful completion or failure",
	}, flowLabelNames)

	prometheus.MustRegister(flowMetricsCount)
	prometheus.MustRegister(flowMetricsDurationMsec)
	prometheus.MustRegister(activityMetricsCount)
	prometheus.MustRegister(activityMetricsDurationMsec)

	statsLogger.Info("Prometheus metrics collector is enabled")
}

type PrometheusMetricsCollector struct{}

func buildFlowLabels(flowEntry *metrics.FlowMetrics, state string) prometheus.Labels {
	labels := prometheus.Labels{
		"ApplicationName":    engine.GetAppName(),
		"ApplicationVersion": engine.GetAppVersion(),
		"FlowName":           flowEntry.FlowName,
		"State":              state,
	}

	for k, v := range commonLabels {
		labels[k] = v
	}
	return labels
}

func buildTaskLabels(taskEntry *metrics.TaskMetrics, state string) prometheus.Labels {
	labels := prometheus.Labels{
		"ApplicationName":    engine.GetAppName(),
		"ApplicationVersion": engine.GetAppVersion(),
		"FlowName":           taskEntry.FlowName,
		"ActivityName":       taskEntry.TaskName,
		"State":              state,
	}
	for k, v := range commonLabels {
		labels[k] = v
	}
	return labels
}

func (prom *PrometheusMetricsCollector) FlowMetrics(flowEntry *metrics.FlowMetrics, state flow.Status) {
	switch state {
	case flow.COMPLETED:
		flowMetricsDurationMsec.With(buildFlowLabels(flowEntry, "Completed")).Set(flowEntry.LastExecTime)
		flowMetricsCount.With(buildFlowLabels(flowEntry, "Completed")).Set(float64(flowEntry.Completed))
	case flow.FAILED:
		flowMetricsDurationMsec.With(buildFlowLabels(flowEntry, "Failed")).Set(flowEntry.LastExecTime)
		flowMetricsCount.With(buildFlowLabels(flowEntry, "Failed")).Set(float64(flowEntry.Failed))
	case flow.STARTED:
		flowMetricsCount.With(buildFlowLabels(flowEntry, "Started")).Set(float64(flowEntry.Created))
	}
}

func (prom *PrometheusMetricsCollector) TaskMetrics(taskEntry *metrics.TaskMetrics, state flow.Status) {
	switch state {
	case flow.COMPLETED:
		activityMetricsDurationMsec.With(buildTaskLabels(taskEntry, "Completed")).Set(taskEntry.LastExecTime)
		activityMetricsCount.With(buildTaskLabels(taskEntry, "Completed")).Set(float64(taskEntry.Completed))
	case flow.FAILED:
		activityMetricsDurationMsec.With(buildTaskLabels(taskEntry, "Failed")).Set(taskEntry.LastExecTime)
		activityMetricsCount.With(buildTaskLabels(taskEntry, "Failed")).Set(float64(taskEntry.Failed))
	case flow.STARTED:
		activityMetricsCount.With(buildTaskLabels(taskEntry, "Started")).Set(float64(taskEntry.Created))
	}
}
