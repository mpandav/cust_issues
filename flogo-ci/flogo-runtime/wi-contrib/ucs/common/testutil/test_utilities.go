package testutil

import (
	"os"
)

var testConnection = map[string]interface{}{
	"name":          "Benchmark Test",
	"description":   "Benchmark Test Connection",
	"url":           "https://localhost:44346/api",
	"connectorName": "Benchmark_0",
	"connectorProps": map[string]string{
		"Username": "ecole@tibco.com",
		"Password": "XXXXX",
	},
}

var testConnection1 = map[string]interface{}{
	"name":           "Benchmark Test",
	"description":    "Benchmark Test Connection",
	"url":            "https://localhost:44346/api",
	"connectorName":  "Benchmark_1",
	"connectorProps": map[string]string{},
	"connectionId":   "01F8Y17WQRWQJG3MJWK4TSACJX",
}

var testConnection2 = map[string]interface{}{
	"connectorName":      "Benchmark",
	"connectionId":       "01F8Y17WQRWQJG3MJWK4TSACJX",
	"instanceId":         "01FEET1V1B8GXDHPQF18X1YRHV",
	"providerPathPrefix": "/ucs/provider",
	"connectionName":     "Benchmark Test",
	"yukonclient":        nil,
	"settings":           nil,
}

var testConnection3 = map[string]interface{}{
	"name":               "Benchmark Test",
	"description":        "Benchmark Test Connection",
	"url":                "https://localhost:44346/api",
	"connectorName":      "Benchmark",
	"connectorProps":     map[string]string{},
	"connectionId":       "01F8Y17WQRWQJG3MJWK4TSACJX",
	"instanceID":         "01FE0RQ81JPA1AJ3DN2XH5NW0P",
	"providerPathPrefix": "/ucs/provider",
	"yukonclient":        nil,
	"settings":           nil,
}

func TestConnection() map[string]interface{} {
	conn := CopyMap(testConnection2)
	url, urlExists := os.LookupEnv("UCS_TEST_URL")
	if urlExists {
		conn["url"] = url
	}
	return conn
}

func CopyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = CopyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}
