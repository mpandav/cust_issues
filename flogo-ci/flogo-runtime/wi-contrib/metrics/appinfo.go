package metrics

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"

	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	contribEngine "github.com/tibco/wi-contrib/engine"
	"github.com/tibco/wi-contrib/httpservice"
)

var appInfoLogger = log.ChildLogger(log.RootLogger(), "app.appInfo")

const (
	FlogoHttpServicePort = "FLOGO_HTTP_SERVICE_PORT"
	MetricsPath          = "/app/metrics"
	AppType              = "flogo"
	AppInfoPath          = "/app/info"
	AppJSONExportPath    = "/app/export"
)

type AppInfo struct {
	AppID          string `json:"appId"`
	AppName        string `json:"appName"`
	AppType        string `json:"appType"`
	Description    string `json:"description"`
	InstanceID     string `json:"instanceId"`
	InstrumentURL  string `json:"instrumentUrl"`
	Version        string `json:"version"`
	FEVersion      string `json:"feVersion"`
	AppEndPoint    string `json:"appEndpoint"`
	ConfigEndPoint string `json:"configEndpoint"`
	LoggerEndPoint string `json:"loggerEndpoint"`
	SupportsRerun  bool   `json:"supportsRerun"`
}

var appDetails = &AppInfo{}

var isInit = false

func init() {
	httpservice.RegisterHandler(AppInfoPath, getAppInfo())
	httpservice.RegisterHandler(AppJSONExportPath, exportAppJSON())
}

func getFeVersion() string {
	feVersion, ok := contribEngine.GetSharedData("flogoProdVersion").(string)
	if !ok {
		feVersion = ""
	}

	return feVersion
}

func getInstrumentationURL() string {
	httpPort, appPortOk := os.LookupEnv(FlogoHttpServicePort)
	if !appPortOk {
		return ""
	}

	ip, err := GetIPV4()

	if err != nil {
		appInfoLogger.Errorf("Flogo app %s is not connected to the network or doesnt have a valid IP address. The app monitoring might fail ", engine.GetAppName())
	}

	instURL := "http://" + ip + ":" + httpPort + MetricsPath
	return instURL
}

func getAppEndPoint() string {
	httpPort, appPortOk := os.LookupEnv(FlogoHttpServicePort)
	if !appPortOk {
		return ""
	}

	ip, err := GetIPV4()

	if err != nil {
		appInfoLogger.Errorf("Flogo app %s is not connected to the network or doesnt have a valid IP address. The app monitoring might fail ", engine.GetAppName())
	}

	instURL := "http://" + ip + ":" + httpPort + AppInfoPath

	return instURL

}

func getConfigEndPoint() string {
	httpPort, appPortOk := os.LookupEnv(FlogoHttpServicePort)
	if !appPortOk {
		return ""
	}
	ip, err := GetIPV4()
	if err != nil {
		appInfoLogger.Errorf("Flogo app %s is not connected to the network or doesnt have a valid IP address. The app monitoring might fail ", engine.GetAppName())
	}

	configUrl := "http://" + ip + ":" + httpPort + "/app/configuration"
	return configUrl
}

func getLoggerEndPoint() string {
	httpPort, appPortOk := os.LookupEnv(FlogoHttpServicePort)
	if !appPortOk {
		return ""
	}
	ip, err := GetIPV4()
	if err != nil {
		appInfoLogger.Errorf("Flogo app %s is not connected to the network or doesnt have a valid IP address. The app monitoring might fail ", engine.GetAppName())
	}

	configUrl := "http://" + ip + ":" + httpPort + "/app/logger"
	return configUrl
}

func getAppInfo() http.Handler {
	return http.HandlerFunc(collectAppInfo)
}

func exportAppJSON() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var flogoJSONString string
		var err error
		flogoJSON := contribEngine.GetSharedData("flogoJSON")
		if flogoJSON == nil {
			// try to read the flogo.json from current dir
			appInfoLogger.Info("Cannot find app json in shared data map...trying to read from current dir")
			flogoJSONString, err = contribEngine.GetAppJsonFromCurrentDir()
			if flogoJSONString == "" || err != nil {
				w.WriteHeader(404)
				appInfoLogger.Errorf("Cannot get app JSON from shared data map or current dir Error: %s", err.Error())
				return
			}

		} else {
			appInfoLogger.Info("Found app json in shared data map")
			var ok bool
			flogoJSONString, ok = flogoJSON.(string)
			if !ok {
				w.WriteHeader(500)
				appInfoLogger.Error("Cannot convert app JSON to string")
				return
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write([]byte(flogoJSONString))
		//if err := json.NewEncoder(w).Encode(flogoJSON); err != nil {
		//	appInfoLogger.Error(err)
		//	w.WriteHeader(500)
		//}
	})
}

func collectAppInfo(writer http.ResponseWriter, request *http.Request) {
	initAppInfo()
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(200)

	if err := json.NewEncoder(writer).Encode(appDetails); err != nil {
		appInfoLogger.Error(err)
	}
}

func GetAppInfoLocal() *AppInfo {
	initAppInfo() 
	return appDetails
}

func initAppInfo() {
	if !isInit {
		appDetails.AppID = engine.GetAppName() + "-" + engine.GetAppVersion()
		appDetails.AppName = engine.GetAppName()
		appDetails.Version = engine.GetAppVersion()
		appDetails.FEVersion = getFeVersion()
		appDetails.AppType = AppType
		appDetails.Description = engine.GetAppName()
		appDetails.InstanceID, _ = os.Hostname()
		appDetails.InstrumentURL = getInstrumentationURL()
		appDetails.AppEndPoint = getAppEndPoint()
		appDetails.ConfigEndPoint = getConfigEndPoint()
		appDetails.LoggerEndPoint = getLoggerEndPoint()
		appDetails.SupportsRerun = true
		reqJSON, _ := json.Marshal(appDetails)
		appDetailsJSONString := string(reqJSON)
		appInfoLogger.Debug("AppInfo" + string(appDetailsJSONString))

		isInit = true
	}

}

func GetIPV4() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, intFace := range interfaces {
		if intFace.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if intFace.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		addrs, err := intFace.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("No network found on the Machine. Connect the Machine to the Local Network")
}
