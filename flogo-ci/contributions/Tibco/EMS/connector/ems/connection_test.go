package ems

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var oneWaySSLConnectionJSON = []byte(`{
	"name": "ems_test",
	"authenticationMode": "SSL",
	"serverUrl": "ssl://10.233.80.11:7223",
	"caCert": "file://C:/Users/RohiniPrabhakar.Goha/Downloads/rootCA.crt.pem",
	"noVerifyHostname": false,
	"enablemTLS": false,
	"reconnectCount": 4,
	"reconnectDelay": 500,
	"reconnectTimeout": 0
}`)

var twoWaySSLConnectionJSON = []byte(`{
	"name": "ems_test",
	"authenticationMode": "SSL",
	"serverUrl": "ssl://10.233.80.11:7223",
	"clientCert": "file://C:/Users/RohiniPrabhakar.Goha/Downloads/client.crt.pem",
	"clientKey": "file://C:/Users/RohiniPrabhakar.Goha/Downloads/client.key.pem",
	"caCert": "file://C:/Users/RohiniPrabhakar.Goha/Downloads/rootCA.crt.pem",
	"noVerifyHostname": false,
	"enablemTLS": true,
	"reconnectCount": 4,
	"reconnectDelay": 500,
	"reconnectTimeout": 0
}`)

func TestEMSConnectionOneWaySSL(t *testing.T) {
	conn := make(map[string]interface{})

	err := json.Unmarshal(oneWaySSLConnectionJSON, &conn)
	if err != nil {
		fmt.Println("JSON Parsing Error:", err)
		os.Exit(1)
	}

	emsFactory := &EMSFactory{}
	manager, err := emsFactory.NewManager(conn)
	if err != nil {
		fmt.Println("Failed to create EMS Manager:", err)
		t.Fail()
	}

	emsManager, ok := manager.(*EmsSharedConfigManager)
	if !ok {
		fmt.Println("Failed to assert EMS manager type")
		t.Fail()
	}

	err = emsManager.Start()
	if err != nil {
		fmt.Println("Failed to start EMS connection:", err)
		t.Fail()
	}

	fmt.Println("EMS Connection (One-Way SSL) Started Successfully!")

	err = emsManager.Stop()
	if err != nil {
		fmt.Println("Failed to stop EMS connection:", err)
		t.Fail()
	}

	fmt.Println("EMS Connection Stopped Successfully!")
}

func TestEMSConnectionTwoWaySSL(t *testing.T) {
	conn := make(map[string]interface{})

	err := json.Unmarshal(twoWaySSLConnectionJSON, &conn)
	if err != nil {
		fmt.Println("JSON Parsing Error:", err)
		os.Exit(1)
	}

	emsFactory := &EMSFactory{}
	manager, err := emsFactory.NewManager(conn)
	if err != nil {
		fmt.Println("Failed to create EMS Manager:", err)
		t.Fail()
	}

	emsManager, ok := manager.(*EmsSharedConfigManager)
	if !ok {
		fmt.Println("Failed to assert EMS manager type")
		t.Fail()
	}

	err = emsManager.Start()
	if err != nil {
		fmt.Println("Failed to start EMS connection:", err)
		t.Fail()
	}

	fmt.Println("EMS Connection (Two-Way SSL) Started Successfully!")

	err = emsManager.Stop()
	if err != nil {
		fmt.Println("Failed to stop EMS connection:", err)
		t.Fail()
	}

	fmt.Println("EMS Connection Stopped Successfully!")
}
