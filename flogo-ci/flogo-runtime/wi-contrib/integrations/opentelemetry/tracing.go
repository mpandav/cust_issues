package opentelemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/engine/secret"
	"github.com/project-flogo/core/support/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
)

// FlogoOTelTracer ...
type FlogoOTelTracer struct {
	otelTracer     oteltrace.Tracer
	tracerProvider *sdktrace.TracerProvider
	endPoint       string
	tags           []attribute.KeyValue
}

// FlogoTracingContext ...
type FlogoTracingContext struct {
	ctx context.Context
}

const (
	OtelSpan = "FLOGO_OTEL_SPAN_KIND"
)

func init() {
	enableTrace, _ := coerce.ToBool(os.Getenv(OTelEnableTrace))
	if enableTrace {
		endpoint, ok := os.LookupEnv(FlogoOTelTraceOTLPEndpointEnv)
		if !ok {
			endpoint = os.Getenv(OTelExporterEndpoint)
		}
		if endpoint == "" {
			logger.Error("OpenTelemetry tracing feature is enabled but OpenTelemetry collector endpoint is not set")
			return
		}
		jt := &FlogoOTelTracer{endPoint: endpoint}
		err := trace.RegisterTracer(jt)
		if err != nil {
			logger.Errorf("Failed to register tracer with Flogo engine due to error - %s", err.Error())
			return
		}
	}
}

// Start ...
func (fgTracer *FlogoOTelTracer) Start() error {
	ctx := context.Background()

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

	exporter, err := createExporter(ctx, fgTracer.endPoint)
	if err != nil {
		logger.Errorf("Failed to enable OpenTelemetry tracing due to error - %s. Ensure OpenTelemetry configuration is correct.", err.Error())
		//TODO: Currently engine does not fail when error returned from Start().
		// For now stopping engine from starting since issue with tracer
		panic(err.Error())
		//return err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetErrorHandler(FlogoErrorHandler{})
	fgTracer.otelTracer = otel.Tracer(TracerName)
	customTags := os.Getenv(OTelTraceAttributes)
	if customTags != "" {
		tagPairs := strings.Split(customTags, ",")
		for _, pair := range tagPairs {
			if pair != "" {
				kv := strings.Split(pair, "=")
				if len(kv) == 2 {
					k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
					fgTracer.tags = append(fgTracer.tags, attribute.String(k, v))
				}
			}
		}
	}
	fgTracer.tags = append(fgTracer.tags, semconv.ServiceVersionKey.String(engine.GetAppVersion()))
	fgTracer.tracerProvider = tracerProvider
	return nil
}

// FlogoErrorHandler ...
type FlogoErrorHandler struct {
}

// Handle ...
func (f FlogoErrorHandler) Handle(err error) {
	s, ok := status.FromError(err)
	if err == context.DeadlineExceeded || (ok && s.Code() == gcodes.Unavailable) {
		logger.Errorf("OpenTelemetry collector endpoint is not reachable. If issue persist, check value set to '%s' as well as check collector receiver configuration. If TLS is enabled for the receiver, ensure '%s' is set.", OTelExporterEndpoint, OTelExporterTLSServerCert)
		return
	}
	logger.Error(err.Error())
}

func createExporter(ctx context.Context, endpoint string) (*otlptrace.Exporter, error) {
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
	epURL, _ := url.Parse(endpoint)
	if epURL != nil && (epURL.Scheme == "https" || epURL.Scheme == "http") {
		options := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(epURL.Host),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
			otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
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
			options = append(options, otlptracehttp.WithTLSClientConfig(tlsConfig))
		} else {
			options = append(options, otlptracehttp.WithInsecure())
		}

		if epURL != nil && epURL.Path != "" {
			options = append(options, otlptracehttp.WithURLPath(epURL.Path))
		}
		return otlptracehttp.New(ctx, options...)
	}
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithCompressor("gzip"),
		otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
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
		options = append(options, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		options = append(options, otlptracegrpc.WithInsecure())
	}
	return otlptracegrpc.New(ctx, options...)

}

func getTLSConfig(allowInsecure bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	serverCert := os.Getenv(OTelExporterTLSServerCert)
	if serverCert == "" {
		if allowInsecure {
			logger.Warnf("Skipping server certificate verification. This is not recommended in production. If TLS is enabled on the receiver side, configure server certificate using '%s'", OTelExporterTLSServerCert)
			tlsConfig.InsecureSkipVerify = true
		} else {
			return nil, nil
		}
	} else if strings.HasPrefix(serverCert, "file://") {
		logger.Infof("'%s' is set to '%s'", OTelExporterTLSServerCert, serverCert)
		certFile := serverCert[7:]
		caCertPEM, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, err
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCertPEM)
		if !ok {
			return nil, errors.New("Failed to parse root certificate. ")
		}
		tlsConfig.RootCAs = roots
	} else if strings.HasPrefix(serverCert, "SECRET:") {
		logger.Infof("'%s' is set to encrypted value", OTelExporterTLSServerCert)
		encodedValue := string(serverCert[7:])
		value, err := secret.GetSecretValueHandler().DecodeValue(encodedValue)
		if err != nil {
			return nil, err
		}
		caCertPEM, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return nil, err
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCertPEM)
		if !ok {
			return nil, errors.New("Failed to parse root PEM certificate. Likely certificate value is not properly encrypted. Refer documentation for correct usage.")
		}
		tlsConfig.RootCAs = roots
	} else {
		logger.Infof("'%s' is set to base64 encoded value", OTelExporterTLSServerCert)
		caCertPEM, err := base64.StdEncoding.DecodeString(serverCert)
		if err != nil {
			return nil, err
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCertPEM)
		if !ok {
			return nil, errors.New("Failed to parse root PEM certificate. Certificate value is not a valid base64 encoded string. Ensure that certificate value is properly encoded. If value is encrypted, then add 'SECRET:' prefix to the value.")
		}
		tlsConfig.RootCAs = roots
	}
	return tlsConfig, nil
}

// Stop ...
func (fgTracer *FlogoOTelTracer) Stop() error {
	if fgTracer.tracerProvider != nil {
		_ = fgTracer.tracerProvider.Shutdown(context.Background())
		fgTracer.tracerProvider = nil
	}

	if fgTracer.otelTracer != nil {
		fgTracer.otelTracer = nil
	}

	return nil
}

// Name ...
func (fgTracer *FlogoOTelTracer) Name() string {
	return TracerName
}

// Extract ...
func (fgTracer *FlogoOTelTracer) Extract(format trace.CarrierFormat, carrier interface{}) (trace.TracingContext, error) {
	switch format {
	case trace.HTTPHeaders:
		r, ok := carrier.(*http.Request)
		if !ok {
			return &FlogoTracingContext{}, fmt.Errorf("Invalid type for data in tracer extract. Expected: %T, Received: %T", http.Request{}, carrier)
		}
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		return &FlogoTracingContext{ctx: ctx}, nil
	case trace.TextMap:
		data, ok := carrier.(map[string]string)
		if !ok {
			return &FlogoTracingContext{}, fmt.Errorf("Invalid type for data in tracer extract. Expected: map[string]string, Received: %T", carrier)
		}
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), MapCarrier(data))
		return &FlogoTracingContext{ctx: ctx}, nil
	}

	return nil, nil
}

// Inject ...
func (fgTracer *FlogoOTelTracer) Inject(tCtx trace.TracingContext, format trace.CarrierFormat, carrierData interface{}) error {
	switch format {
	case trace.HTTPHeaders:
		r, ok := carrierData.(*http.Request)
		if !ok {
			return fmt.Errorf("Invalid type for data in tracer inject. Expected: %T, Received: %T", http.Request{}, carrierData)
		}
		fc := tCtx.(*FlogoTracingContext)
		span := oteltrace.SpanFromContext(fc.ctx)
		if span != nil {
			span.SetAttributes(semconv.HTTPMethodKey.String(r.Method), semconv.HTTPURLKey.String(r.URL.String()))
		}
		otel.GetTextMapPropagator().Inject(fc.ctx, propagation.HeaderCarrier(r.Header))
	case trace.TextMap:
		data, ok := carrierData.(map[string]string)
		if !ok {
			return fmt.Errorf("Invalid type for data in tracer inject. Expected: map[string]string, Received: %T", carrierData)
		}
		fc := tCtx.(*FlogoTracingContext)
		otel.GetTextMapPropagator().Inject(fc.ctx, MapCarrier(data))
	}
	return nil
}

// StartTrace ...
func (fgTracer *FlogoOTelTracer) StartTrace(config trace.Config, parent trace.TracingContext) (trace.TracingContext, error) {
	kvs := make([]attribute.KeyValue, len(config.Tags))
	for k, v := range config.Tags {
		// Coerce value to string
		stringVal, _ := coerce.ToString(v)
		name := strings.ReplaceAll(k, "_", ".")
		if !strings.HasPrefix(name, FlogoTagPrefix) {
			name = FlogoTagPrefix + name
		}
		kvs = append(kvs, attribute.String(name, stringVal))
	}
	if engine.GetEnvName() != "" {
		kvs = append(kvs, attribute.String("deployment.environment", engine.GetEnvName()))
	}
	if len(fgTracer.tags) > 0 {
		// Custom tags added
		kvs = append(kvs, fgTracer.tags...)
	}

	flogoContext := &FlogoTracingContext{}
	ctx := context.Background()
	if parent != nil {
		// Use parent context to create child span
		parentFlogoContext := parent.(*FlogoTracingContext)
		ctx = parentFlogoContext.ctx
	}

	// Add SpanKindClient to the SpanStartOption instead of default value of SpanKindInternal

	opts := []oteltrace.SpanStartOption{oteltrace.WithAttributes(kvs...)}

	spanClient := oteltrace.SpanKindInternal
	otelspan, ok := os.LookupEnv(OtelSpan)
	if ok {
		otelspan = strings.ToLower(otelspan)
		switch otelspan {
		case "server":
			spanClient = oteltrace.SpanKindServer
		case "client":
			spanClient = oteltrace.SpanKindClient
		case "producer":
			spanClient = oteltrace.SpanKindProducer
		case "consumer":
			spanClient = oteltrace.SpanKindConsumer
		default:
			spanClient = oteltrace.SpanKindInternal
		}
	}
	opts = append(opts, oteltrace.WithSpanKind(spanClient))
	flogoContext.ctx, _ = fgTracer.otelTracer.Start(ctx, config.Operation, opts...)
	return flogoContext, nil
}

// FinishTrace ...
func (fgTracer *FlogoOTelTracer) FinishTrace(tContext trace.TracingContext, err error) error {
	if tContext == nil {
		return fmt.Errorf("Failed to complete span. Trace object is nil")
	}
	fc := tContext.(*FlogoTracingContext)
	span := oteltrace.SpanFromContext(fc.ctx)
	if span != nil {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		span.End()
	}
	return nil
}

// TraceObject ...
func (fc *FlogoTracingContext) TraceObject() interface{} {
	return fc.ctx
}

// SetTags ...
func (fc *FlogoTracingContext) SetTags(tags map[string]interface{}) bool {
	logger.Debugf("Setting tags")
	span := oteltrace.SpanFromContext(fc.ctx)
	if span != nil {
		for k, v := range tags {
			// Coerce value to string
			stringVal, _ := coerce.ToString(v)
			name := strings.ReplaceAll(k, "_", ".")
			if !strings.HasPrefix(name, FlogoTagPrefix) {
				name = FlogoTagPrefix + name
			}
			span.SetAttributes(attribute.String(name, stringVal))
		}
		return true
	}
	return false
}

// SetTag ...
func (fc *FlogoTracingContext) SetTag(tagKey string, tagValue interface{}) bool {
	logger.Debugf("Setting tag")
	strValue, err := coerce.ToString(tagValue)
	name := strings.ReplaceAll(tagKey, "_", ".")
	if !strings.HasPrefix(name, FlogoTagPrefix) {
		name = FlogoTagPrefix + name
	}
	span := oteltrace.SpanFromContext(fc.ctx)
	if err == nil && span != nil {
		span.SetAttributes(attribute.String(name, strValue))
		return true
	}
	return false
}

// LogKV ...
func (fc *FlogoTracingContext) LogKV(map[string]interface{}) bool {
	return true
}

// TraceID implements trace.TracingContext.TraceID()
func (fc *FlogoTracingContext) TraceID() string {
	if fc == nil || fc.ctx == nil {
		return ""
	}
	return oteltrace.SpanFromContext(fc.ctx).SpanContext().TraceID().String()
}

// SpanID implements trace.TracingContext.SpanID()
func (fc *FlogoTracingContext) SpanID() string {
	if fc == nil || fc.ctx == nil {
		return ""
	}
	return oteltrace.SpanFromContext(fc.ctx).SpanContext().SpanID().String()
}
