package monitoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tibco/wi-contrib/metrics"

	"github.com/pkg/errors"
	"github.com/project-flogo/core/app"
	core "github.com/project-flogo/core/engine/event"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine"
)

const (
	MonConfig            = "FLOGO_APP_MON_SERVICE_CONFIG"
	FlogoHttpServicePort = "FLOGO_HTTP_SERVICE_PORT"
	MonPath              = "/v1/apps/register"
)

var statsLogger = log.ChildLogger(log.RootLogger(), "app.monitoring")

var eventTypes = []string{app.AppEventType}

type AppMetrics struct {
	appName    string
	appVersion string
}

type AppConfig struct {
	MonHost string   `json:"host"`
	MonPort string   `json:"port"`
	AppHost string   `json:"appHost"`
	AppPort string   `json:"appPort"`
	ApiKey  string   `json:"apiKey"`
	Tag     []string `json:"tags"`
}

var appMetrics = &AppMetrics{}

var appConfig = &AppConfig{}

type AppRequest struct {
	AppName    string   `json:"appName"`
	AppVersion string   `json:"appVersion"`
	Tags       []string `json:"tags"`
	HostName   string   `json:"appHost"`
	PortName   string   `json:"appPort"`
	FEVersion  string   `json:"feVersion"`
	HostIP     string   `json:"appHostIP"`
}

type AppEventListener struct {
}

func (ls *AppEventListener) HandleEvent(evt *core.Context) error {

	switch t := evt.GetEvent().(type) {
	case app.AppEvent:
		switch t.AppStatus() {
		case app.STARTED:
			statsLogger.Debug("App started. Registering App with the Monitoring Service")
			appMetrics.appName = t.AppName()
			appMetrics.appVersion = t.AppVersion()
			registerAppMonitoring()
		case app.STOPPING:
			statsLogger.Debug("App Stopped. De-registering App with the Monitoring Service")
			unregisterAppMonitoring()
		}
	}

	return nil
}

func init() {

	//Check for TCI Env. App Registration is not required in TCI Env.
	if isTCIEnv() {
		return
	}
	_, monFileOk := os.LookupEnv(MonConfig)

	_, appPortOk := os.LookupEnv(FlogoHttpServicePort)

	//If Mon File or Service Port is not set then no need to go ahead. Return.
	if !monFileOk || !appPortOk {
		statsLogger.Debug("No Monitoring Service configuration or HTTP Service port specified. Refer documentation for more details.")
		return
	}

	metrics.RegisterEventListener()

	// Start handler routine
	statsLogger.Debug("Monitoring Service configuration found. HTTP Service Port is set.")
	err := core.RegisterListener("appevent", &AppEventListener{}, eventTypes)
	if err != nil {
		statsLogger.Error(err)
		return
	}

}

func isTCIEnv() bool {
	_, ok := os.LookupEnv("TIBCO_INTERNAL_TCI_SUBSCRIPTION_ID")
	return ok
}

func unregisterAppMonitoring() {

	statsLogger.Debugf("Calling Monitoring Service at Host  %s  and Port %s ", appConfig.MonHost, appConfig.MonPort)
	monURL := "http://" + appConfig.MonHost + ":" + appConfig.MonPort + MonPath
	appRequest := prepareAppRequest(appConfig)

	callMonitoringService(http.MethodDelete, monURL, appRequest, false)
}

func registerAppMonitoring() {

	appConfig, err := readValuesFromJSON()
	if err != nil {
		statsLogger.Error(err)
		return
	} else {
		appRequest := prepareAppRequest(appConfig)
		statsLogger.Debugf("Calling Monitoring Service at Host  %s  and Port %s ", appConfig.MonHost, appConfig.MonPort)
		monURL := "http://" + appConfig.MonHost + ":" + appConfig.MonPort + MonPath

		callMonitoringService(http.MethodPost, monURL, appRequest, true)

	}

}

func callMonitoringService(method string, monURL string, request *AppRequest, register bool) {

	var reqJson []byte = nil
	var err error
	var req *http.Request
	var response *http.Response

	reqJson, err = json.Marshal(request)
	buffer := bytes.NewBuffer(reqJson)

	client := &http.Client{Timeout: 1 * time.Minute}
	req, err = http.NewRequest(method, monURL, buffer)
	if err != nil {
		statsLogger.Errorf("Monitoring Service call request creation failed with error %s ", err.Error())
		return
	}

	if req != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Api-Key", appConfig.ApiKey)
	}

	statsLogger.Debugf("Created HTTP request for method %s. Calling Service", method)

	response, err = client.Do(req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "connection refused") {
			statsLogger.Errorf("Application monitoring is enabled but unable to connect to monitoring service. Ensure that monitoring service is running and values configured in [%s] environment variable are correct.", MonConfig)
		} else {
			statsLogger.Errorf("Monitoring Service call failed with error %s ", err.Error())
		}
		return
	}
	statsLogger.Debug(response)
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		if register {
			statsLogger.Infof("App %s registered successfully with Monitoring Service ", request.AppName)
		} else {
			statsLogger.Infof("App de-registered successfully with Monitoring Service ")

		}

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			statsLogger.Error(err)
		}
		bodyString := string(bodyBytes)
		statsLogger.Debugf("Response from Monitoring Service call %s ", bodyString)
	} else {
		if register {
			statsLogger.Errorf("App %s failed to register successfully with Monitoring Service with %d status code", request.AppName, response.StatusCode)
		} else {
			statsLogger.Errorf("App failed to de-registered successfully with Monitoring Service  with %d status code", response.StatusCode)
		}

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			statsLogger.Error(err)
		}
		statsLogger.Debugf("Response from Monitoring Service call %s ", string(bodyBytes))

	}
}

func prepareAppRequest(config *AppConfig) *AppRequest {

	feVersion, ok := engine.GetSharedData("flogoProdVersion").(string)
	if !ok {
		feVersion = ""
	}

	appPort := config.AppPort

	if appPort == "" {
		appPort, _ = os.LookupEnv(FlogoHttpServicePort)
	}

	appHost := config.AppHost
	if appHost == "" {
		appHost, _ = os.Hostname()

	}

	appHostIp, error := metrics.GetIPV4()

	if error != nil {
		statsLogger.Errorf("Flogo app %s is not connected to the network or doesnt have a valid IP address. The app monitoring might fail ", appMetrics.appName)
	}

	newReq := AppRequest{AppName: appMetrics.appName, AppVersion: appMetrics.appVersion, Tags: config.Tag, HostName: appHost, PortName: appPort, FEVersion: feVersion, HostIP: appHostIp}

	reqJson, _ := json.Marshal(newReq)
	reqJsonString := string(reqJson)
	statsLogger.Debugf("Sending App Request %s", reqJsonString)
	return &newReq

}

// Check if the Property FLOGO_APP_MON_CONFIG is set.
// If set then see if the value is directly provided as File or as a String and then parse accordingly.
func readValuesFromJSON() (*AppConfig, error) {

	monFile, monFileOk := os.LookupEnv(MonConfig)

	if !monFileOk {
		return nil, errors.New("Monitoring Service configuration not provided")
	}

	var monJson []byte
	var err error
	if monFile != "" {
		if strings.HasSuffix(monFile, ".json") {
			monJson, err = ioutil.ReadFile(monFile)
		} else if strings.HasPrefix(strings.TrimSpace(monFile), "{") {
			monJson = []byte(monFile)
		} else {
			errMsg := fmt.Sprintf("Invalid value set for %s variable. It must be a valid JSON file or key/value pair.", MonConfig)
			statsLogger.Error(errMsg)
		}
	}

	if err == nil {
		err := json.Unmarshal(monJson, appConfig)
		if err == nil {
			return appConfig, nil
		}
	}
	return nil, errors.New("Failed to load Monitoring Service configuration details.")
}
