package opentelemetry

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"
	"time"

	flow "github.com/TIBCOSoftware/flogo-contrib/action/flow/event"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/flow/support/event"
	"github.com/tibco/wi-contrib/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	_ "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	metricApi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc/credentials"
)

const (
	ApplicationName           = attribute.Key("app_name")
	ApplicationVersion        = attribute.Key("app_version")
	FlowName                  = attribute.Key("flow_name")
	ActivityName              = attribute.Key("activity_name")
	State                     = attribute.Key("state")
	Environment               = attribute.Key("deployment_environment")
	ActivityExecutionsMetrics = "flogo_activity_executions_total"
	FlowExecutionsMetrics     = "flogo_flow_executions_total"
)

type OTMetricsCollector struct {
	meterProvider       *metric.MeterProvider
	activityExecCounter metricApi.Int64Counter
	flowExecCounter     metricApi.Int64Counter
	commonTags          []attribute.KeyValue
	hostName            string
}

func init() {
	enableMetrics, _ := coerce.ToBool(os.Getenv(OTelEnableMetrics))
	if enableMetrics {
		collector := &OTMetricsCollector{}
		engine.LifeCycle(collector)
	}
}

func (c *OTMetricsCollector) Start() error {
	ctx := context.Background()
	exporter, err := createMetricsExporter(ctx)
	if err != nil {
		logger.Errorf("Failed to enable OpenTelemetry metrics due to error - %s. Ensure OpenTelemetry configuration is correct.", err.Error())
		return err
	}
	serviceName := os.Getenv(OTelServiceName)
	if serviceName == "" {
		serviceName = engine.GetAppName()
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
		resource.WithHost(),
		resource.WithProcessPID(),
	)
	if err != nil {
		logger.Errorf("Failed to create resource due to error - %s", err.Error())
		return err
	}
	customTags := os.Getenv(OTelMetricsAttributes)
	if customTags != "" {
		logger.Infof("Custom metrics attributes: %s", customTags)
		tagPairs := strings.Split(customTags, ",")
		for _, pair := range tagPairs {
			if pair != "" {
				kv := strings.Split(pair, "=")
				if len(kv) == 2 {
					k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
					c.commonTags = append(c.commonTags, attribute.String(k, v))
				}
			}
		}
	}
	hostName, _ := os.Hostname()
	c.commonTags = append(c.commonTags, semconv.HostNameKey.String(hostName))
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(10*time.Second))), metric.WithResource(res))
	meter := meterProvider.Meter("flogo_" + engine.GetAppName() + "_" + engine.GetAppVersion())
	c.activityExecCounter, err = meter.Int64Counter(ActivityExecutionsMetrics, metricApi.WithDescription("Total number of times the activity is started, completed or failed"), metricApi.WithUnit("1"))
	if err != nil {
		return err
	}
	c.flowExecCounter, err = meter.Int64Counter(FlowExecutionsMetrics, metricApi.WithDescription("Total number of times the flow is started, completed or failed"), metricApi.WithUnit("1"))
	if err != nil {
		return err
	}
	c.meterProvider = meterProvider
	metrics.RegisterAppMetricsListener(c)
	logger.Info("OpenTelemetry metrics collector started")
	return nil
}

func (c *OTMetricsCollector) Stop() error {
	if c.meterProvider != nil {
		_ = c.meterProvider.Shutdown(context.Background())
	}
	return nil
}

func (c *OTMetricsCollector) FlowMetrics(flowEntry *metrics.FlowMetrics, state event.Status) {
	switch state {
	case flow.COMPLETED:
		{
			c.flowMetrics(flowEntry, flow.COMPLETED)
		}
	case flow.CANCELLED:
		{
			c.flowMetrics(flowEntry, flow.CANCELLED)
		}
	case flow.FAILED:
		{
			c.flowMetrics(flowEntry, flow.FAILED)
		}
	case flow.STARTED:
		{
			c.flowMetrics(flowEntry, flow.STARTED)
		}
	}
}

func (c *OTMetricsCollector) TaskMetrics(taskEntry *metrics.TaskMetrics, state event.Status) {
	switch state {
	case flow.COMPLETED:
		{
			c.activityMetrics(taskEntry, flow.COMPLETED)
		}
	case flow.CANCELLED:
		{
			c.activityMetrics(taskEntry, flow.CANCELLED)
		}
	case flow.FAILED:
		{
			c.activityMetrics(taskEntry, flow.FAILED)
		}
	case flow.STARTED:
		{
			c.activityMetrics(taskEntry, flow.STARTED)
		}
	}
}

func (c *OTMetricsCollector) activityMetrics(taskEntry *metrics.TaskMetrics, state string) {
	ctx := context.Background()
	var addOptions []metricApi.AddOption
	opt := metricApi.WithAttributes(
		ApplicationName.String(engine.GetAppName()),
		ApplicationVersion.String(engine.GetAppVersion()),
		FlowName.String(taskEntry.FlowName),
		ActivityName.String(taskEntry.TaskName),
		State.String(strings.ToLower(state)),
		Environment.String(engine.GetEnvName()),
	)

	addOptions = append(addOptions, opt)
	if len(c.commonTags) > 0 {
		addOptions = append(addOptions, metricApi.WithAttributes(c.commonTags...))
	}
	c.activityExecCounter.Add(ctx, 1, addOptions...)
}

func (c *OTMetricsCollector) flowMetrics(taskEntry *metrics.FlowMetrics, state string) {
	ctx := context.Background()
	var addOptions []metricApi.AddOption
	opt := metricApi.WithAttributes(
		ApplicationName.String(engine.GetAppName()),
		ApplicationVersion.String(engine.GetAppVersion()),
		FlowName.String(taskEntry.FlowName),
		State.String(strings.ToLower(state)),
		Environment.String(engine.GetEnvName()),
	)

	addOptions = append(addOptions, opt)
	if len(c.commonTags) > 0 {
		addOptions = append(addOptions, metricApi.WithAttributes(c.commonTags...))
	}
	c.flowExecCounter.Add(ctx, 1, addOptions...)
}

func createMetricsExporter(ctx context.Context) (metric.Exporter, error) {
	headers := make(map[string]string)
	customHeaders := os.Getenv(OTelExporterHeaders)
	if customHeaders != "" {
		tagPairs := strings.Split(customHeaders, ",")
		for _, pair := range tagPairs {
			if pair != "" {
				kv := strings.Split(pair, "=")
				if len(kv) == 2 {
					k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
					headers[k] = v
				}
			}
		}
	}
	endPoint, ok := os.LookupEnv(FlogoOTelMetricsOTLPEndpointEnv)
	if !ok {
		endPoint = os.Getenv(OTelExporterEndpoint)
	}
	if endPoint == "" {
		return nil, errors.New("OpenTelemetry metrics feature is enabled but OpenTelemetry collector endpoint is not set")
	}
	epURL, _ := url.Parse(endPoint)
	if epURL != nil && (epURL.Scheme == "https" || epURL.Scheme == "http") {
		options := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(epURL.Host),
			otlpmetrichttp.WithHeaders(headers),
			otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
			otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
				Enabled:         true,
				InitialInterval: 10 * time.Second,
				MaxInterval:     60 * time.Second,
				MaxElapsedTime:  2 * time.Minute,
			}),
		}
		tlsConfig, err := getTLSConfig(false)
		if err != nil {
			return nil, err
		} else if tlsConfig != nil {
			options = append(options, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
		} else {
			options = append(options, otlpmetrichttp.WithInsecure())
		}

		if epURL != nil && epURL.Path != "" {
			options = append(options, otlpmetrichttp.WithURLPath(epURL.Path))
		}
		return otlpmetrichttp.New(ctx, options...)
	}
	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endPoint),
		otlpmetricgrpc.WithHeaders(headers),
		otlpmetricgrpc.WithCompressor("gzip"),
		otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         true,
			InitialInterval: 10 * time.Second,
			MaxInterval:     60 * time.Second,
			MaxElapsedTime:  2 * time.Minute,
		}),
	}
	tlsConfig, err := getTLSConfig(false)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		options = append(options, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		options = append(options, otlpmetricgrpc.WithInsecure())
	}
	return otlpmetricgrpc.New(ctx, options...)
}
