package engine

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/property"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/httpservice"
)

type Variable struct {
	Name     string      `json:"name"`
	Type     string      `json:"datatype"`
	Value    interface{} `json:"value"`
	Readonly bool        `json:"readOnly,omitempty"`
}

var whiteListedVars = make(map[string]string)
var configRes = ConfigResponse{Config: &ConfigDetails{}}

type ConfigDetails struct {
	AppProperties    []*Variable `json:"appProperties"`
	SystemProperties []Variable  `json:"systemProperties"`
}

type ConfigResponse struct {
	Config *ConfigDetails `json:"configDetails"`
}

func init() {
	httpservice.RegisterHandler("/app/configuration", http.HandlerFunc(appConfigurationHandler))
	// whitelisted env vars
	whiteListedVars["FLOGO_RUNNER_QUEUE"] = ""
	whiteListedVars["FLOGO_RUNNER_WORKERS"] = ""
	whiteListedVars["FLOGO_HTTP_SERVICE_PORT"] = ""
	whiteListedVars["FLOGO_LOG_LEVEL"] = ""
	whiteListedVars["FLOGO_LOG_FORMAT"] = ""
	whiteListedVars["FLOGO_MAPPING_SKIP_MISSING"] = ""
	whiteListedVars["FLOGO_APP_METRICS_LOG_EMITTER_ENABLE"] = ""
	whiteListedVars["FLOGO_APP_METRICS_LOG_EMITTER_CONFIG"] = ""
	whiteListedVars["FLOGO_APP_DELAYED_STOP_INTERVAL"] = ""
	whiteListedVars["FLOGO_MAPPING_OMIT_NULLS"] = ""
	configRes.Config.SystemProperties = getSystemProperties()
}

func setAppVariables(appJson string) {
	// Fetch app properties from app json
	// This had to be done as OSS replaces all secrets with actual values before reading app into struct.
	// Once secret value is converted to actual value, there is no way to differentiate between regular string property vs secret.
	props := &struct {
		Properties []*Variable `json:"properties,omitempty"`
	}{}
	err := json.Unmarshal([]byte(appJson), props)
	if err != nil {
		log.RootLogger().Errorf("Failed to read app variables due to error %s", err.Error())
	}
	configRes.Config.AppProperties = props.Properties
}

// UnmarshalJSON This is executed when Unmarshal() called at line#56
func (a *Variable) UnmarshalJSON(props []byte) error {
	ser := &struct {
		Name  string      `json:"name"`
		Type  string      `json:"type"`
		Value interface{} `json:"value,omitempty"`
	}{}
	if err := json.Unmarshal(props, ser); err != nil {
		return err
	}
	a.Name = ser.Name
	a.Value = ser.Value
	switch ser.Type {
	case data.TypeString.String():
		val, ok := ser.Value.(string)
		if ok && strings.HasPrefix(val, "SECRET:") {
			a.Type = "password"
			a.Value = ""
		} else {
			if !ok {
				// Looks like non string value set to string field
				// Coerce value to string
				a.Value, _ = coerce.ToString(ser.Value)
			}
			a.Type = ser.Type
		}
	case data.TypeFloat64.String():
		a.Type = "number"
	default:
		a.Type = ser.Type
	}
	a.Readonly = false
	return nil
}

func appConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	log.RootLogger().Debugf("Returning current app configuration...")
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		if len(configRes.Config.AppProperties) > 0 {
			for _, v := range configRes.Config.AppProperties {
				if v.Type == "password" {
					// set empty value
					v.Value = ""
					continue
				}
				v.Value, _ = property.DefaultManager().GetProperty(v.Name)
			}
		}
		log.RootLogger().Debugf("Configuration: %+v", configRes)
		err := json.NewEncoder(w).Encode(configRes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getSystemProperties() []Variable {
	var engineVariables []Variable
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		name := pair[0]
		if strings.HasPrefix(name, "FLOGO_APP_PROPS") {
			eVar := Variable{}
			eVar.Name = name
			eVar.Type = "string"
			eVar.Value = pair[1]
			if name == "FLOGO_APP_PROPS_ENV" {
				eVar.Readonly = true
			}
			engineVariables = append(engineVariables, eVar)
		} else {
			_, whiteListed := whiteListedVars[name]
			if whiteListed || shouldBeIncluded(name) {
				log.RootLogger().Debugf("Variable [%s] whitelisted", name)
				eVar := Variable{}
				eVar.Name = name
				eVar.Value = pair[1]
				eVar.Type = "string"
				if name == "FLOGO_HTTP_SERVICE_PORT" {
					eVar.Readonly = true
				}
				engineVariables = append(engineVariables, eVar)
			}
		}
	}
	log.RootLogger().Debugf("Environment variables: %v", engineVariables)
	return engineVariables
}

func shouldBeIncluded(name string) bool {
	return strings.Contains(name, "AWS") || strings.Contains(name, "OTEL") || strings.Contains(name, "JAEGER")
}
