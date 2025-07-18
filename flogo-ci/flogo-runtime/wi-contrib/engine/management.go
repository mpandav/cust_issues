package engine

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/managed"
	"github.com/tibco/wi-contrib/engine/reconfigure/dynamicprops"
	"github.com/tibco/wi-contrib/environment"
	"github.com/tibco/wi-contrib/httpservice"
)

const (
	MANAGEMENT_PORT_KEY = "FLOGO_HTTP_SERVICE_PORT"
)

var engineManager EngineManager

type APIReply struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func (m *Mangement) pingHandler(w http.ResponseWriter, r *http.Request) {
	m.createPingCallTracker()
	triggerInfoList := m.em.engine.App().TriggerStatuses()
	if len(triggerInfoList) > 0 {
		for _, triggerInfo := range triggerInfoList {
			if triggerInfo.Status == managed.StatusFailed {
				log.RootLogger().Errorf("Health check failed. Trigger:%s, Error:%s", triggerInfo.Name, triggerInfo.Error)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (m *Mangement) reconfigureHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		reply := APIReply{}
		if !property.IsPropertyReconfigureEnabled() {
			reply.Success = false
			reply.Message = fmt.Sprintf("Dynamic app property reconfiguration is not enabled. To enabled this set the environment variable %s to 'true'.", property.EnvAppPropertyReconfigure)
			log.RootLogger().Error(reply.Message)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(reply)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			reply.Success = false
			reply.Message = fmt.Sprintf("Failed to read request body. Error:%v", err)
			log.RootLogger().Error(reply.Message)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(reply)
			return
		}

		data := make(map[string]interface{})
		//Reconfigure the app properties for dynamic resolver only when when body is not empty
		if len(body) > 0 {
			//convert body to map[string]interface{}

			err = json.Unmarshal(body, &data)
			if err != nil {
				reply.Success = false
				reply.Message = fmt.Sprintf("Failed to unmarshal request body. Error:%v", err)
				log.RootLogger().Error(reply.Message)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(reply)
				return
			}

		}
		dynamicAppPropResolver := property.GetExternalPropertyResolver(dynamicprops.ResolverName).(*dynamicprops.DynamicAppPropResolver)
		if dynamicAppPropResolver != nil {
			// updating the Dynamic props resolver's store with new mappings
			dynamicAppPropResolver.UpdateStore(data)
		}

		err = GetEngineManager().ReconfigureApp()
		if err != nil {
			reply.Success = false
			reply.Message = fmt.Sprintf("Failed to reconfigure app. Error:%v", err)
			log.RootLogger().Error(reply.Message)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(reply)
			return
		}
		reply.Success = true
		reply.Message = "App successfully reconfigured"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(reply)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *Mangement) loggerHandler(w http.ResponseWriter, r *http.Request) {
	reply := APIReply{}
	switch r.Method {
	case http.MethodGet:
		logLevel := os.Getenv(log.EnvKeyLogLevel)
		if logLevel == "" {
			logLevel = "INFO"
		}
		data := make(map[string]string)
		data["level"] = logLevel
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(data)
		return
	case http.MethodPut:
		requestData := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			reply.Success = false
			reply.Message = fmt.Sprintf("Failed to process request body. Error:%v", err)
			_ = json.NewEncoder(w).Encode(reply)
			return
		}

		level, _ := requestData["level"].(string)
		if level == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			reply.Success = false
			reply.Message = "Log level is not set in the body"
			_ = json.NewEncoder(w).Encode(reply)
			return
		}
		toUpper := strings.ToUpper(level)
		switch toUpper {
		case "INFO", "DEBUG", "ERROR", "WARN", "TRACE":
			msg := "Log level set to '" + toUpper + "'"
			if toUpper == "WARN" || toUpper == "ERROR" {
				// Log before restricted log level is set
				log.RootLogger().Info(msg)
			}
			_ = os.Setenv(log.EnvKeyLogLevel, toUpper)
			logLevel := log.ToLogLevel(toUpper)
			log.SetLogLevel(log.RootLogger(), logLevel)
			if toUpper == "INFO" || toUpper == "DEBUG" || toUpper == "TRACE" {
				// Log after log level is set
				log.RootLogger().Info(msg)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			reply.Success = true
			reply.Message = msg
			_ = json.NewEncoder(w).Encode(reply)
			return
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			reply.Success = false
			reply.Message = fmt.Sprintf("Invalid log level [%s]. Supported Log Levels: [INFO, DEBUG, ERROR, WARN]", level)
			_ = json.NewEncoder(w).Encode(reply)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *Mangement) startServer() {
	port := getPort()
	if port != "" {
		mux := http.NewServeMux()
		//DEPRECATED
		mux.Handle("/ping", http.HandlerFunc(m.pingHandler))
		mux.Handle("/app/healthy", http.HandlerFunc(m.pingHandler))
		mux.Handle("/app/logger", http.HandlerFunc(m.loggerHandler))
		mux.Handle("/app/config/refresh", http.HandlerFunc(m.reconfigureHandler))
		if httpservice.HasHandler() {
			for name, h := range httpservice.AllHandlers() {
				mux.Handle(name, h)
			}
		}
		port := getPort()
		log.RootLogger().Infof("Management Service started successfully on Port[%s]", port)
		err := http.ListenAndServe(":"+port, mux)
		if err != nil {
			log.RootLogger().Errorf("Failed to management service on Port[%s] due to Error:{%s}", port, err.Error())
			panic(err.Error())
		}
	}
}

type Mangement struct {
	em EngineManager
	//flogoEngine engine.Engine
	ticker     *time.Ticker
	pingCalled bool
}

type EngineManager struct {
	engine engine.Engine
}

func GetEngineManager() EngineManager {
	return engineManager
}

func (em EngineManager) ReconfigureApp() error {
	return em.engine.App().Reconfigure()
}

func NewManagement(e engine.Engine) *Mangement {
	engineManager = EngineManager{engine: e}
	return &Mangement{em: engineManager}
}

func (m *Mangement) Start(engine engine.Engine) {
	go m.startServer()
}

// https://jira.tibco.com/browse/FLOGO-8917
func (m *Mangement) createPingCallTracker() {
	if environment.IsTCIEnv() {
		m.pingCalled = true
		if m.ticker == nil {
			// Starting ticker at 10.5 seconds interval. For TCI apps, ping endpoint is called at 10 seconds interval.
			m.ticker = time.NewTicker(10500 * time.Millisecond)
			go func() {
				for {
					select {
					case <-m.ticker.C:
						if m.pingCalled == false {
							// Ping endpoint is not called in last 10 seconds
							log.RootLogger().Error("Health-check endpoint is not called by infra")
						}
						// reset the flag
						m.pingCalled = false
					}
				}
			}()
		}
	}
}

func getPort() string {
	return os.Getenv(MANAGEMENT_PORT_KEY)
}
