package jaeger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

const (
	tracerName           = "jaeger"
	otJaegerConfigEnvVar = "FLOGO_APP_MONITORING_OT_JAEGER"
)

var logger = log.ChildLogger(log.RootLogger(), "jaeger-tracer")
var enabled bool

func init() {
	if conf, ok := os.LookupEnv(otJaegerConfigEnvVar); ok {
		enabled, _ = coerce.ToBool(conf)
		if enabled {
			jt := &Tracer{}
			trace.RegisterTracer(jt)
		}
	}
}

// Tracer struct for FlogoTracer
type Tracer struct {
	tracer opentracing.Tracer
	closer io.Closer
}

// Start implements managed.Managed
func (jt *Tracer) Start() error {
	if enabled {
		// get config from jaeger env variables
		// Ideally, (AgentHost & AgentPort) OR Endpoint needs to be configured. Sampler config (type and param) must be set too to collect the traces.
		logger.Info("Jaeger tracing enabled")
		jaegerConfig, err := config.FromEnv()
		if err != nil {
			logger.Errorf("Failed to get jaeger configuration due to error: %s", err.Error())
			return err
		}

		if jaegerConfig.Sampler.Type == "" && jaegerConfig.Sampler.Param == 0 && jaegerConfig.Sampler.SamplingServerURL == "" {
			// Set defaults for sampler
			jaegerConfig.Sampler.Type = "const"
			jaegerConfig.Sampler.Param = 1
		}

		if jaegerConfig.ServiceName == "" {
			jaegerConfig.ServiceName = engine.GetAppName() + "-" + engine.GetAppVersion()
		}

		data, _ := json.Marshal(jaegerConfig)
		logger.Debugf("Jaeger configuration: %s", string(data))

		jt.tracer, jt.closer, err = jaegerConfig.NewTracer(config.Logger(jaeger.StdLogger))
		if err != nil {
			return fmt.Errorf("failed to create tracer due to error: %s", err.Error())
		}
	}
	return nil
}

// Stop implements managed.Managed
func (jt *Tracer) Stop() error {
	if enabled {
		enabled = false
		return jt.closer.Close()
	}
	return nil
}

// Name implements FlogoTracer.Configure
func (jt *Tracer) Name() string {
	return tracerName
}

// StartTrace starts a new trace or a child trace of parent
func (jt *Tracer) StartTrace(config trace.Config, parent trace.TracingContext) (trace.TracingContext, error) {
	// check if parent is spanContext
	var span opentracing.Span
	if parent == nil {
		// start new span
		span = jt.tracer.StartSpan(config.Operation, opentracing.Tags(config.Tags))
	} else if sc, ok := parent.TraceObject().(opentracing.SpanContext); ok {
		span = jt.tracer.StartSpan(config.Operation, opentracing.ChildOf(sc), opentracing.Tags(config.Tags))
	} else if s, ok := parent.TraceObject().(opentracing.Span); ok {
		span = jt.tracer.StartSpan(config.Operation, opentracing.ChildOf(s.Context()), opentracing.Tags(config.Tags))
	} else {
		return nil, fmt.Errorf("unknown tracing context while starting trace")
	}
	tc := TraceContext{jaegerCtx: span}
	return tc, nil
}

// FinishTrace implements trace.Tracer.FinishTrace()
func (jt *Tracer) FinishTrace(tContext trace.TracingContext, err error) error {
	if tContext.TraceObject() == nil {
		return fmt.Errorf("error finishing trace. Trace object is nil")
	}
	if s, ok := tContext.TraceObject().(opentracing.Span); ok {
		if err != nil {
			s.SetTag("error", true)
			s.LogFields(
				otlog.String("event", "error"),
				otlog.Error(err),
			)
		}
		s.Finish()
		return nil
	}
	return fmt.Errorf("error finishing trace. no span found")
}

// TraceContext struct implements trace.TracingContext
type TraceContext struct {
	// jaegerCtx could be opentracing.span or opentracing.SpanContext
	jaegerCtx interface{}
}

// TraceObject implements trace.TracingContext.TraceObject()
func (jc TraceContext) TraceObject() interface{} {
	if sc, ok := jc.jaegerCtx.(opentracing.SpanContext); ok {
		return sc
	}
	if s, ok := jc.jaegerCtx.(opentracing.Span); ok {
		return s
	}
	return nil
}

// SetTags implements trace.TracingContext.SetTags()
func (jc TraceContext) SetTags(tags map[string]interface{}) bool {
	if jc.TraceObject() == nil {
		logger.Warn("Unable to set tags as traceobject is nil")
		return false
	}
	span, ok := jc.TraceObject().(opentracing.Span)
	if ok {
		for k, v := range tags {
			span.SetTag(k, v)
		}
	}
	return ok
}

// SetTag implements trace.TracingContext.SetTag()
func (jc TraceContext) SetTag(tagKey string, tagValue interface{}) bool {
	if jc.TraceObject() == nil {
		logger.Warn("Unable to set tag as traceobject is nil")
		return false
	}
	span, ok := jc.TraceObject().(opentracing.Span)
	if ok {
		span.SetTag(tagKey, tagValue)
	}
	return ok
}

// LogKV implements trace.TracingContext.LogKV()
func (jc TraceContext) LogKV(kvs map[string]interface{}) bool {
	if jc.TraceObject() == nil {
		logger.Warn("Unable to log key value as traceobject is nil")
		return false
	}
	span, ok := jc.TraceObject().(opentracing.Span)
	if ok {
		span.LogKV(kvs)
	}
	return ok
}

// TraceID implements trace.TracingContext.TraceID()
func (jc TraceContext) TraceID() string {
	var jsc jaeger.SpanContext
	if sc, ok := jc.jaegerCtx.(opentracing.SpanContext); ok {
		jsc = sc.(jaeger.SpanContext)
	}
	if s, ok := jc.jaegerCtx.(opentracing.Span); ok {
		jsc = s.Context().(jaeger.SpanContext)
	}
	return jsc.TraceID().String()
}

// SpanID implements trace.TracingContext.SpanID()
func (jc TraceContext) SpanID() string {
	var jsc jaeger.SpanContext
	if sc, ok := jc.jaegerCtx.(opentracing.SpanContext); ok {
		jsc = sc.(jaeger.SpanContext)
	}
	if s, ok := jc.jaegerCtx.(opentracing.Span); ok {
		jsc = s.Context().(jaeger.SpanContext)
	}
	return jsc.SpanID().String()
}

// Inject implements trace.Tracer.Inject()
func (jt *Tracer) Inject(tctx trace.TracingContext, format trace.CarrierFormat, carrier interface{}) (err error) {
	if tctx.TraceObject() == nil {
		return fmt.Errorf("failed to inject trace as trace object was nil")
	}

	var spanCtx opentracing.SpanContext
	if sc, ok := tctx.TraceObject().(opentracing.SpanContext); ok {
		spanCtx = sc
	} else if s, ok := tctx.TraceObject().(opentracing.Span); ok {
		spanCtx = s.Context()
	} else {
		return fmt.Errorf("unknown type in trace object. failed to inject")
	}

	switch format {
	case trace.Binary:
		req, er := coerce.ToBytes(carrier)
		if er != nil {
			return fmt.Errorf("invalid type for data in tracer inject. Expected: %T, Received: %T", http.Request{}, carrier)
		}
		err = jt.tracer.Inject(spanCtx, opentracing.Binary, req)
	case trace.TextMap:
		req, ok := carrier.(map[string]string)
		if !ok {
			return fmt.Errorf("invalid type for data in tracer inject. Expected: %T, Received: %T", http.Request{}, carrier)
		}
		err = jt.tracer.Inject(spanCtx, opentracing.TextMap, opentracing.TextMapCarrier(req))
	case trace.HTTPHeaders:
		req, ok := carrier.(*http.Request)
		if !ok {
			return fmt.Errorf("invalid type for data in tracer inject. Expected: %T, Received: %T", http.Request{}, carrier)
		}
		if s, ok := tctx.TraceObject().(opentracing.Span); ok {
			s.SetTag("http.method", req.Method)
			s.SetTag("http.url", req.URL.String())
		}
		err = jt.tracer.Inject(spanCtx, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}
	if err != nil {
		return fmt.Errorf("failed to inject trace into request due to error: %s", err.Error())
	}
	// todo add any standard tags?
	return nil
}

// Extract implements trace.Tracer.Extract()
func (jt *Tracer) Extract(format trace.CarrierFormat, data interface{}) (ret trace.TracingContext, err error) {
	var spanCtx opentracing.SpanContext
	switch format {
	case trace.Binary:
		req, er := coerce.ToBytes(data)
		if er != nil {
			return TraceContext{}, fmt.Errorf("invalid type for data in tracer extract. Expected: %T, Received: %T", bytes.Buffer{}, data)
		}
		spanCtx, err = jt.tracer.Extract(opentracing.Binary, req)
	case trace.TextMap:
		req, ok := data.(map[string]string)
		if !ok {
			return TraceContext{}, fmt.Errorf("invalid type for data in tracer extract. Expected: %T, Received: %T", map[string]string{}, data)
		}
		spanCtx, err = jt.tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(req))
	case trace.HTTPHeaders:
		req, ok := data.(*http.Request)
		if !ok {
			return TraceContext{}, fmt.Errorf("invalid type for data in tracer extract. Expected: %T, Received: %T", http.Request{}, data)
		}
		spanCtx, err = jt.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	}
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return TraceContext{}, fmt.Errorf("failed to extract trace from request due to error: %s", err.Error())
	}
	if spanCtx != nil {
		return TraceContext{jaegerCtx: spanCtx}, nil
	}
	return nil, nil
}
