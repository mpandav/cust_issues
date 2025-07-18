package opentelemetry

import (
	"github.com/project-flogo/core/support/log"
)

// Constants ...
const (
	OTelExporterEndpoint            = "FLOGO_OTEL_OTLP_ENDPOINT"
	OTelExporterTLSServerCert       = "FLOGO_OTEL_TLS_SERVER_CERT"
	OTelEnableTrace                 = "FLOGO_OTEL_TRACE"
	OTelEnableMetrics               = "FLOGO_OTEL_METRICS"
	OTelTraceAttributes             = "FLOGO_OTEL_TRACE_ATTRIBUTES"
	OTelMetricsAttributes           = "FLOGO_OTEL_METRICS_ATTRIBUTES"
	OTelExporterHeaders             = "FLOGO_OTEL_OTLP_HEADERS"
	OTelServiceName                 = "OTEL_SERVICE_NAME"
	TracerName                      = "opentelemetry"
	FlogoTagPrefix                  = "flogo."
	FlogoOTelTraceOTLPEndpointEnv   = "FLOGO_OTEL_TRACE_OTLP_ENDPOINT"
	FlogoOTelMetricsOTLPEndpointEnv = "FLOGO_OTEL_METRICS_OTLP_ENDPOINT"
)

var logger = log.ChildLogger(log.RootLogger(), "opentelemetry")
