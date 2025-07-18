package opentelemetry

// Implements go.opentelemetry.io/otel/propagation.TextMapCarrier
type MapCarrier map[string]string

func (mc MapCarrier) Get(key string) string {
	val, ok := mc[key]
	if ok {
		return val
	}
	return ""
}

func (mc MapCarrier) Set(key string, value string) {
	mc[key] = value
}

func (mc MapCarrier) Keys() []string {
	var keys []string
	for k := range mc {
		keys = append(keys, k)
	}
	return keys
}
