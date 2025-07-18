package aws

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-xray-sdk-go/header"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/aws/aws-xray-sdk-go/xraylog"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
)

const (
	tracerName         = "xray"
	otXrayConfigEnvVar = "FLOGO_AWS_XRAY_ENABLE"
	envMetadata        = "FLOGO_AWS_XRAY_METADATA"
	envAnnotations     = "FLOGO_AWS_XRAY_ANNOTATIONS"
	xrayHTTPHeader     = "X-Amzn-Trace-Id"
	metadataNamespace  = "flogo"
)

var rootLogger log.Logger
var tracerLogger log.Logger
var sdkLogger log.Logger
var enabled bool

// Tracer struct for FlogoTracer
type Tracer struct{}

func init() {
	if conf, ok := os.LookupEnv(otXrayConfigEnvVar); ok {
		enabled, _ = coerce.ToBool(conf)
		if enabled {
			rootLogger = log.ChildLogger(log.RootLogger(), "aws-xray")
			tracerLogger = log.ChildLogger(rootLogger, "tracer")
			sdkLogger = log.ChildLogger(rootLogger, "sdk")
			tracerLogger.Info("Initializing AWS Xray Tracing")
			xt := &Tracer{}
			trace.RegisterTracer(xt)
		}
	}
}

// Start implements managed.Managed
func (xt *Tracer) Start() error {
	tracerLogger.Info("AWS X-Ray tracing enabled")
	xl := xrayLogger{}
	xray.SetLogger(xl)
	return nil
}

// Stop implements managed.Managed
func (xt *Tracer) Stop() error {
	if enabled {
		tracerLogger.Info("AWS X-Ray tracing stopped")
		enabled = false
	}
	return nil
}

// Name implements FlogoTracer.Configure
func (xt *Tracer) Name() string {
	return tracerName
}

// StartTrace starts a new trace or a child trace of parent
func (xt *Tracer) StartTrace(config trace.Config, parent trace.TracingContext) (trace.TracingContext, error) {
	var ctx context.Context
	var segment *xray.Segment
	var isFlow bool
	var subsegmentName string

	activityName := config.Tags["task_name"]
	parentFlowID := config.Tags["parent_flow_id"]
	appNameAndVersion := getRootSegmentName()

	if parent == nil {
		tracerLogger.Debugf("Tracing context is nil, creating root segment with name [%s]", appNameAndVersion)
		ctx, _ = xray.BeginSegment(context.Background(), appNameAndVersion)
		subsegmentName, isFlow = config.Operation, true
	} else if tc, ok := parent.TraceObject().(TraceContext); ok {
		ctx = tc.ctx
		if parentFlowID != nil {
			// Case: subflow, got parent but invoking activity name is not present
			tracerLogger.Debugf("Subflow detected")
			subsegmentName = config.Operation
		} else if tc.isRoot {
			// Case: timer-plus-http, got parent but not activity yet
			tracerLogger.Debugf("HTTP trigger detected")
			subsegmentName, isFlow = config.Operation, true
		} else {
			// Case: normal ectivity
			subsegmentName = activityName.(string)
		}
	} else {
		return nil, fmt.Errorf("Unknown tracing context while starting trace")
	}
	tracerLogger.Debugf("Creating subsegment with name [%s]", subsegmentName)
	ctx, segment = xray.BeginSubsegment(ctx, subsegmentName)
	segment.AddAnnotation("flogo_flow_name", config.Operation)
	segment.AddMetadataToNamespace(metadataNamespace, "flow_name", config.Operation)
	if activityName != nil {
		segment.AddMetadataToNamespace(metadataNamespace, "activity_name", activityName.(string))
	}
	tc := TraceContext{ctx: ctx, isFlow: isFlow}
	return tc, nil
}

// FinishTrace implements trace.Tracer.FinishTrace()
func (xt *Tracer) FinishTrace(tContext trace.TracingContext, err error) error {
	if tContext.TraceObject() == nil {
		return fmt.Errorf("Error finishing trace. Trace object is nil")
	}
	tc, ok := tContext.TraceObject().(TraceContext)
	segment := xray.GetSegment(tc.ctx)
	if ok && segment != nil {
		segment.Close(err)
		if tc.isFlow {
			segment.ParentSegment.Close(err)
		}
		return nil
	}
	return fmt.Errorf("Error finishing trace. Invalid trace context")
}

// Inject implements trace.Tracer.Inject()
func (xt *Tracer) Inject(tctx trace.TracingContext, format trace.CarrierFormat, carrier interface{}) (err error) {
	if tctx.TraceObject() == nil {
		return fmt.Errorf("Failed to inject trace as trace object was nil")
	}
	tc, ok := tctx.TraceObject().(TraceContext)
	if !ok {
		return fmt.Errorf("Unknown type in trace object. Failed to inject")
	}
	switch format {
	case trace.Binary:
		tracerLogger.Warn("Unsupported carrier format for inject: BINARY")
	case trace.TextMap:
		tracerLogger.Warn("Unsupported carrier format for inject: TEXT MAP")
	case trace.HTTPHeaders:
		tracerLogger.Debugf("Injecting downstream HTTP header with name [%s] for xray tracing", xrayHTTPHeader)
		req, ok := carrier.(*http.Request)
		if !ok {
			return fmt.Errorf("Invalid type for data in tracer inject. Expected: %T, Received: %T", http.Request{}, carrier)
		}
		dh := xray.GetSegment(tc.ctx).DownstreamHeader()
		req.Header.Add(xrayHTTPHeader, dh.String())
	case trace.Lambda:
		tracerLogger.Debug("Injecting trace header via lambda context for xray tracing")
		lambdaClientContext, ok := carrier.(*lambdacontext.ClientContext)
		if !ok {
			return fmt.Errorf("Invalid type for data in tracer inject. Expected: %T, Received: %T", "", carrier)
		}
		dh := xray.GetSegment(tc.ctx).DownstreamHeader().String()
		lambdaClientContext.Custom = map[string]string{xrayHTTPHeader: dh}
	}
	return nil
}

// Extract implements trace.Tracer.Extract()
func (xt *Tracer) Extract(format trace.CarrierFormat, data interface{}) (ret trace.TracingContext, err error) {
	tc := &TraceContext{}
	switch format {
	case trace.Binary:
		tracerLogger.Warn("Unsupported carrier format for extract: BINARY")
	case trace.TextMap:
		tracerLogger.Warn("Unsupported carrier format for extract: TEXT MAP")
	case trace.HTTPHeaders:
		tracerLogger.Debug("Extracting trace headers from HTTP request")
		req, ok := data.(*http.Request)
		if !ok {
			return nil, fmt.Errorf("Invalid type for data in tracer extract. Expected: %T, Received: %T", http.Request{}, data)
		}
		headerString := req.Header.Get(xrayHTTPHeader)
		var seg *xray.Segment
		var ctx context.Context
		if headerString != "" {
			tracerLogger.Debug("Creating root segment from HTTP request context and headers")
			h := header.FromString(headerString)
			ctx, seg = xray.NewSegmentFromHeader(req.Context(), getRootSegmentName(), req, h)
		} else {
			// HTTP trigger was used...starting http trace from empty context
			ctx, seg = xray.BeginSegment(context.Background(), getRootSegmentName())
		}
		traceHTTP(seg, req)
		tc.ctx = ctx
		tc.isRoot = true
		return tc, nil
	case trace.Lambda:
		tracerLogger.Debug("Extracing trace headers from lambda context")
		m, ok := data.(map[string]string)
		if !ok {
			return nil, fmt.Errorf("Invalid type for data in tracer extract. Expected: %T, Received: %T", map[string]string{}, data)
		}
		h := header.FromString(m[xrayHTTPHeader])
		tc.ctx, _ = xray.BeginFacadeSegment(context.Background(), getRootSegmentName(), h)
		tc.isRoot = true
		return tc, nil
	}
	return nil, nil
}

// TraceContext struct implements trace.TracingContext
type TraceContext struct {
	ctx    context.Context
	isFlow bool
	isRoot bool
}

// TraceObject implements trace.TracingContext.TraceObject()
func (tc TraceContext) TraceObject() interface{} {
	return tc
}

// SetTags implements trace.TracingContext.SetTags()
func (tc TraceContext) SetTags(tags map[string]interface{}) bool {
	tracerLogger.Debugf("Setting trace tags(metadata): %+v", tags)
	if tc.TraceObject() == nil {
		tracerLogger.Warn("Unable to set tags as trace object is nil")
		return false
	}
	traceObj, ok := tc.TraceObject().(TraceContext)
	if ok {
		for k, v := range tags {
			xray.GetSegment(traceObj.ctx).AddMetadataToNamespace(metadataNamespace, k, v)
		}
	}
	return ok
}

// SetTag implements trace.TracingContext.SetTag()
func (tc TraceContext) SetTag(tagKey string, tagValue interface{}) bool {
	tracerLogger.Debugf("Setting tag with key: [%s] and value: [%s]", tagKey, tagValue)
	if tc.TraceObject() == nil {
		tracerLogger.Warn("Unable to set tag as trace object is nil")
		return false
	}
	traceObj, ok := tc.TraceObject().(TraceContext)
	if ok {
		xray.GetSegment(traceObj.ctx).AddMetadataToNamespace(metadataNamespace, tagKey, tagValue)
	}
	return ok
}

// LogKV implements trace.TracingContext.LogKV()
func (tc TraceContext) LogKV(kvs map[string]interface{}) bool {
	return true
}

// TraceID implements trace.TracingContext.TraceID()
func (tc TraceContext) TraceID() string {
	if tc.ctx == nil {
		return ""
	}
	segment := xray.GetSegment(tc.ctx)
	if segment == nil {
		return ""
	}
	if segment.TraceID != "" {
		return segment.TraceID
	}
	parent := segment.ParentSegment
	if parent != nil && parent.TraceID != "" {
		return parent.TraceID
	}
	return ""
}

// SpanID implements trace.TracingContext.SpanID()
func (tc TraceContext) SpanID() string {
	if tc.ctx == nil {
		return ""
	}
	segment := xray.GetSegment(tc.ctx)
	if segment == nil {
		return ""
	}
	return segment.ID
}

func getRootSegmentName() string {
	return fmt.Sprintf("%s-v%s", engine.GetAppName(), engine.GetAppVersion())
}

func traceHTTP(seg *xray.Segment, r *http.Request) {
	seg.Lock()
	defer seg.Unlock()
	scheme := "https://"
	if r.TLS == nil {
		scheme = "http://"
	}
	seg.GetHTTP().GetRequest().Method = r.Method
	seg.GetHTTP().GetRequest().URL = scheme + r.Host + r.URL.Path
	seg.GetHTTP().GetRequest().ClientIP, seg.GetHTTP().GetRequest().XForwardedFor = clientIP(r)
	seg.GetHTTP().GetRequest().UserAgent = r.UserAgent()
}

func clientIP(r *http.Request) (string, bool) {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0]), true
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr, false
	}
	return ip, false
}

type xrayLogger struct{}

func (xrayLogger) Log(level xraylog.LogLevel, msg fmt.Stringer) {
	switch level {
	case xraylog.LogLevelDebug:
		// sdkLogger.Debug(msg.String())
	case xraylog.LogLevelInfo:
		sdkLogger.Info(msg.String())
	case xraylog.LogLevelWarn:
		sdkLogger.Warn(msg.String())
	case xraylog.LogLevelError:
		sdkLogger.Error(msg.String())
	default:
		sdkLogger.Warnf("Unexpected logLevel [%d] %s", level, msg.String())
	}
}
