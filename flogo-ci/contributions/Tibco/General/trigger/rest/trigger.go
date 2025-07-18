package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/CloudyKit/router"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/schema"
	engine2 "github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"
	"github.com/tibco/flogo-general/src/app/General/trigger/rest/cors"
	"github.com/tibco/wi-contrib/engine"
	"github.com/tibco/wi-contrib/engine/jsonschema"
	"github.com/tibco/wi-contrib/environment"
)

const (
	REST_CORS_PREFIX = "REST_TRIGGER"
	// SWAGGER_EP       = "/api/v2/swagger.json"
)

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

var validMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{}, &Reply{})

// RestTrigger REST trigger struct
type Trigger struct {
	server         *Server
	config         *trigger.Config
	name           string
	logger         log.Logger
	spec           map[string]interface{}
	basePathPrefix string
	// flowLimitLock  sync.RWMutex
	// flowControlled bool
}

type Factory struct {
}

func (t *Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New Creates a new trigger instance for a given id
func (t *Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	return &Trigger{name: config.Id, config: config}, nil
}

func (t *Trigger) Initialize(ctx trigger.InitContext) error {
	triggerLog := ctx.Logger()
	t.logger = triggerLog
	triggerLog.Debugf("In init, id '%s'", t.config.Id)

	router := router.New()
	s := &Settings{}
	err := metadata.MapToStruct(t.config.Settings, s, true)

	if err != nil {
		return err
	}

	if t.config.Settings == nil {
		panic(fmt.Sprintf("No Settings found for trigger '%s'", t.config.Id))
	}

	port := t.config.Settings["port"]
	if port == nil {
		panic(fmt.Sprintf("No Port found for trigger '%s' in settings", t.config.Id))
	}

	envPort, found := os.LookupEnv("PORT")
	if found && envPort != "" {
		triggerLog.Infof("Found PORT environment variable. Setting trigger port to '%s'.", envPort)
		port = envPort
	}

	tPort, _ := coerce.ToString(port)
	addr := ":" + tPort

	triggerLog.Infof("Name: %s, Port: %s", t.name, tPort)
	//Handle options for cors
	pathMap := make(map[string]string)
	for _, handler := range ctx.GetHandlers() {

		handlerSetting := &HandlerSettings{}

		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)

		if err != nil {
			return err
		}

		if handlerIsValid(handlerSetting) {
			path := handlerSetting.Path
			_, ok := pathMap[path]
			if !ok {
				pathMap[path] = path
				router.AddRoute("OPTIONS", replacePath(path), t.handleCorsPreflight)
			}
		} else {
			panic(fmt.Sprintf("Invalid handler: %v", handler))
		}
	}
	// Init handlers
	for _, handler := range ctx.GetHandlers() {

		handlerSetting := &HandlerSettings{}

		err := metadata.MapToStruct(handler.Settings(), handlerSetting, true)

		if err != nil {
			return err
		}
		if handlerIsValid(handlerSetting) {
			method := strings.ToUpper(handlerSetting.Method)
			path := handlerSetting.Path
			router.AddRoute(method, replacePath(path), newActionHandler(t, handler, handlerSetting))
			triggerLog.Infof("%s: Registered handler [Method: %s, Path: %s]", t.name, method, path)
		} else {
			panic(fmt.Sprintf("Invalid handler: %v", handler))
		}
	}
	if os.Getenv("FLOGO_EXPOSE_SWAGGER_EP") == "true" {

		// Register swagger endpoint only if env var is true
		// Default swagger endpoint is /api/v2/swagger.json

		swaggerEp := "/api/v2/swagger.json"
		if os.Getenv("FLOGO_SWAGGER_EP") != "" {
			swaggerEp = os.Getenv("FLOGO_SWAGGER_EP")
		}
		router.AddRoute("GET", swaggerEp, newSwagger2xHandler(t, swaggerEp))
		router.AddRoute("OPTIONS", swaggerEp, t.handleCorsPreflight)
		triggerLog.Infof("%s: Registered Swagger handler [Method: GET, Path: %s]", t.name, swaggerEp)
	}

	t.server = NewServer(addr, router)
	t.server.secureConnection, _ = coerce.ToBool(s.SecureConnection)
	if t.server.secureConnection && environment.IsTCIEnv() {
		// Skip secure connection when running in TCI
		triggerLog.Warn("Ignoring certificates configured for secure connection in TCI")
		t.server.secureConnection = false
	}
	if t.server.secureConnection == true {
		triggerLog.Info("Enabling secure connection...")
		t.server.serverKey, _ = coerce.ToString(s.ServerKey)
		t.server.caCertificate, _ = coerce.ToString(s.CaCertificate)

		if t.server.serverKey == "" || t.server.caCertificate == "" {
			return errors.New("Server Key and CA certificate must be configured for secure connection")
		}

		if strings.HasPrefix(t.server.serverKey, "file://") {
			// Its file
			fileName := t.server.serverKey[7:]
			serverKey, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}
			t.server.serverKey = string(serverKey)
		}

		if strings.HasPrefix(t.server.caCertificate, "file://") {
			// Its file
			fileName := t.server.caCertificate[7:]
			serverCert, err := ioutil.ReadFile(fileName)
			if err != nil {
				return err
			}
			t.server.caCertificate = string(serverCert)
		}
	}

	return nil
}

func replacePath(path string) string {
	path = strings.Replace(path, "}", "", -1)
	if strings.Contains(path, "*") {
		return strings.Replace(path, "{", "", -1)
	}
	return strings.Replace(path, "{", ":", -1)
}

func (t *Trigger) Start() error {
	t.logger.Infof("Starting %s...", t.name)
	err := t.server.Start(t.logger)
	if err != nil {
		t.logger.Errorf("Failed to start [%s] due to error - %s", t.name, err.Error())
		return err
	}
	t.logger.Infof("Started %s", t.name)
	return err
}

// Stop implements util.Managed.Stop
func (t *Trigger) Stop() error {
	t.logger.Infof("Stopping %s...", t.name)
	err := t.server.Stop()
	if err != nil {
		t.logger.Errorf("Failed to stop Trigger [%s] due to error - %s", t.name, err.Error())
		return err
	}
	t.logger.Infof("Stopped %s", t.name)
	return err
}

// func (t *Trigger) Pause() error {
// 	t.controlFlow()
// 	t.logger.Infof("Paused %s", t.name)
// 	return nil
// }

// func (t *Trigger) Resume() error {
// 	t.releaseFlowControl()
// 	t.logger.Infof("Resumed %s", t.name)
// 	return nil
// }

// Handles the cors preflight request
func (t *Trigger) handleCorsPreflight(w http.ResponseWriter, r *http.Request, _ router.Parameter) {

	t.logger.Debugf("Received [OPTIONS] request to CorsPreFlight: %+v", r)
	c := cors.New(REST_CORS_PREFIX, t.logger)
	c.HandlePreflight(w, r)
}

// IDResponse id response object
type IDResponse struct {
	ID string `json:"id"`
}

type Endpoint struct {
	Title   string                 `json:"title"`
	Swagger map[string]interface{} `json:"swagger"`
	Name    string                 `json:"name"`
}

type AppMd struct {
	Endpoints []*Endpoint `json:"endpoints"`
}

// RFC 7807 compliant error response
type ErrorResponse struct {
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func newSwagger2xHandler(rt *Trigger, swaggerEp string) router.Handler {

	return func(w http.ResponseWriter, r *http.Request, ps router.Parameter) {

		errorRes := &ErrorResponse{}
		if rt.spec == nil {
			var appJSON []byte
			var err error
			var ok bool

			appJSONStr, ok := engine.GetSharedData("flogoJSON").(string)

			if ok {
				appJSON = []byte(appJSONStr)
			}

			if appJSON == nil || !ok {
				rt.logger.Debug("Reading flogo.json file")
				appJSON, err = os.ReadFile("./flogo.json")
				if appJSON == nil || err != nil {
					errorRes.Status = http.StatusInternalServerError
					errorRes.Detail = "App not found in the cache. Contact support."
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(errorRes)
					return
				}
			}

			md := &struct {
				Md *AppMd `json:"metadata,omitempty"`
			}{}
			err = json.Unmarshal(appJSON, md)
			if err != nil {
				rt.logger.Errorf("Failed to read metadata from the app. Error: %s", err.Error())
				errorRes.Status = http.StatusInternalServerError
				errorRes.Detail = "Error reading metadata from the app. Check app logs."
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(errorRes)
				return
			}

			if md == nil || md.Md == nil || len(md.Md.Endpoints) == 0 {
				rt.logger.Error("Required swagger metadata is missing in the app. App requires re-export from latest version.")
				errorRes.Status = http.StatusBadRequest
				errorRes.Detail = "Required swagger metadata is missing in the app. App requires re-export from latest version."
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(errorRes)
				return
			}

			for _, ep := range md.Md.Endpoints {
				if ep.Title == rt.name { //check Title for single and Name for multitrigger app
					rt.spec = ep.Swagger
					break
				} else if ep.Name == rt.name {
					rt.spec = ep.Swagger
				}
			}
			if environment.IsTCIEnv() {
				if len(md.Md.Endpoints) > 1 {
					rt.basePathPrefix = os.Getenv("TCI_APP_ID") + "/" + rt.name
				} else {
					rt.basePathPrefix = os.Getenv("TCI_APP_ID")
				}

			}

		}
		if rt.spec == nil {
			rt.logger.Error("Swagger spec is missing in the metadata. Contact support.")
			errorRes.Status = http.StatusBadRequest
			errorRes.Detail = "Swagger spec is missing in the metadata. Contact support."
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(errorRes)
			return
		}
		host := r.Host
		prefixHeader := r.Header["X-Forwarded-Prefix"]
		var basePath string
		if len(prefixHeader) > 0 {
			basePath = prefixHeader[0]
		} else if environment.IsTCIEnv() {
			basePath = strings.Replace(r.URL.Path, swaggerEp, "/", -1)
			basePath = basePath + rt.basePathPrefix
		} else {
			basePath = strings.Replace(r.URL.Path, swaggerEp, "/", -1)
		}
		rt.logger.Debugf("Swagger Host: %s", host)
		rt.logger.Debugf("Swagger BasePath: %s", basePath)
		_, ok := rt.spec["openapi"]
		if !ok {
			rt.spec["host"] = host
			rt.spec["basePath"] = basePath
			rt.spec["schemes"] = []string{"https", "http"}
		} else {
			servers := []map[string]string{
				{"url": "https://" + host + basePath},
				{"url": "http://" + host + basePath},
			}
			rt.spec["servers"] = servers
		}
		if rt.spec["info"] != nil {
			swaggerInfo, _ := rt.spec["info"].(map[string]interface{})
			if swaggerInfo != nil {
				swaggerInfo["title"] = engine2.GetAppName()
				swaggerInfo["version"] = engine2.GetAppVersion()
				rt.spec["info"] = swaggerInfo
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get(cors.ORIGIN_HEADER))
		err := json.NewEncoder(w).Encode(rt.spec)
		if err != nil {
			rt.logger.Errorf("Failed to return swagger spec. Error: %s", err.Error())
			errorRes.Status = http.StatusInternalServerError
			errorRes.Detail = "Failed to return swagger spec. Check app logs."
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(errorRes)
			return
		}
	}
}
func newActionHandler(rt *Trigger, handler trigger.Handler, handlerSetting *HandlerSettings) router.Handler {

	return func(w http.ResponseWriter, r *http.Request, ps router.Parameter) {

		// if flow control is enabled and we reach runner queue limit, return server busy
		// if rt.isFlowControlled() {
		// 	http.Error(w, "Server is busy", http.StatusServiceUnavailable)
		// 	return
		// }

		startTimestamp := time.Now()
		//triggerLog.Infof("REST Trigger: Received request for id '%s'", rt.config.Id)
		triggerLog := handler.Logger()
		triggerLog.Infof("REST Trigger: RequestId: %s", w.Header().Get("X-Request-Id"))
		output := &Output{}

		outputValidateVal := handlerSetting.OutputValidation
		outputValidate, err := coerce.ToBool(outputValidateVal)
		if err != nil {
			outputValidate = false
		}

		method := strings.ToUpper(handlerSetting.Method)
		output.Method = method

		//Path parameters
		pathParams := make(map[string]interface{})

		if ps.Len() > 0 {
			pathParamMetadata, _ := LoadJsonSchemaFromMetadata(handler.Schemas().Output["pathParams"])
			if pathParamMetadata != nil {
				definePathParam, _ := ParseParams(pathParamMetadata)
				if definePathParam != nil {
					for _, qParam := range definePathParam {
						if ps.ByName(qParam.Name) == "" && strings.EqualFold(qParam.Required, "true") {
							errMsg := fmt.Sprintf("Required path parameter [%s] is not set", qParam.Name)
							triggerLog.Error(errMsg)
							http.Error(w, errMsg, http.StatusBadRequest)
							return
						}

						qParamName := qParam.Name
						if strings.Contains(qParam.Name, "*") {
							qParamName = strings.Replace(qParam.Name, "*", "", -1)
						}
						if ps.ByName(qParamName) != "" {
							values, err := getValuewithType(qParam, []string{ps.ByName(qParamName)})
							if err != nil {
								errMsg := fmt.Sprintf("Fail to validate path parameter: %v", err)
								triggerLog.Error(errMsg)
								http.Error(w, errMsg, http.StatusBadRequest)
								return
							}
							pathParams[qParam.Name] = values[0]
						}

					}
					output.PathParams = pathParams
				}

			}
		}

		//Query parameters
		queryParams := make(map[string]interface{})

		queryMetadata, _ := LoadJsonSchemaFromMetadata(handler.Schemas().Output["queryParams"])
		if queryMetadata != nil {
			definedQueryParams, _ := ParseParams(queryMetadata)
			if definedQueryParams != nil {
				queryValues := r.URL.Query()
				queryParams = make(map[string]interface{}, len(definedQueryParams))
				for _, qParam := range definedQueryParams {
					value := queryValues[qParam.Name]
					if !notEmpty(value) && strings.EqualFold(qParam.Required, "true") {
						errMsg := fmt.Sprintf("Required query parameter [%s] is not set", qParam.Name)
						triggerLog.Error(errMsg)
						http.Error(w, errMsg, http.StatusBadRequest)
						return
					}

					if notEmpty(value) {
						values, err := getValuewithType(qParam, value)
						if err != nil {
							errMsg := fmt.Sprintf("Fail to validate query parameter: %v", err)
							triggerLog.Error(errMsg)
							http.Error(w, errMsg, http.StatusBadRequest)
							return
						}
						if qParam.Repeating == "false" {
							queryParams[qParam.Name] = values[0]
						} else {
							queryParams[qParam.Name] = values
						}

						triggerLog.Debugf("Query param: Name[%s], Value[%s]", qParam.Name, queryParams[qParam.Name])
					}

				}
				output.QueryParams = queryParams
			}

		}
		//Headers
		headerMetadata, _ := LoadJsonSchemaFromMetadata(handler.Schemas().Output["headers"])
		if headerMetadata != nil {
			definedHeaderParams, _ := ParseParams(headerMetadata)
			if definedHeaderParams != nil {
				headers := make(map[string]interface{}, len(definedHeaderParams))
				headerValues := r.Header
				for _, hParam := range definedHeaderParams {
					value := headerValues[http.CanonicalHeaderKey(hParam.Name)]
					if len(value) == 0 && hParam.Required == "true" {
						errMsg := fmt.Sprintf("Required header [%s] is not set", hParam.Name)
						triggerLog.Error(errMsg)
						http.Error(w, errMsg, http.StatusBadRequest)
						return
					}

					if len(value) > 0 {
						values, err := getValuewithType(hParam, value)
						if err != nil {
							errMsg := fmt.Sprintf("Fail to validate header parameter: %v", err)
							triggerLog.Error(errMsg)
							http.Error(w, errMsg, http.StatusBadRequest)
							return
						}
						if hParam.Repeating == "false" {
							headers[hParam.Name] = values[0]
						} else {
							headers[hParam.Name] = values
						}
						if strings.ToLower(hParam.Name) == "authorization" {
							triggerLog.Debugf("Header: Name[%s], Value[********]", hParam.Name)
						} else {
							triggerLog.Debugf("Header: Name[%s], Value[%s]", hParam.Name, headers[hParam.Name])
						}
					}
				}

				output.Headers = headers
			}
		}

		var handlerPath string
		handlerPath = handlerSetting.Path
		for param := range pathParams {
			valueString := ""
			switch pathParams[param].(type) {
			case int, float32, float64, bool:
				valueString = fmt.Sprint(pathParams[param])
			default:
				valueString = pathParams[param].(string)
			}
			handlerPath = strings.Replace(handlerPath, "{"+param+"}", valueString, -1)
		}
		qp := url.Values{}
		for param := range queryParams {
			valueString := ""
			switch queryParams[param].(type) {
			case int, float32, float64, bool:
				valueString = fmt.Sprint(queryParams[param])
				qp.Add(param, valueString)
			case []interface{}:
				qParams := queryParams[param].([]interface{})
				for _, val := range qParams {
					valueString, err = coerce.ToString(val)
					if err != nil {
						fmt.Println("Error while string coersion: ", err.Error())
					}
					qp.Add(param, valueString)
				}
			default:
				valueString = queryParams[param].(string)
				qp.Add(param, valueString)
			}
			//qp.Add(param, valueString)
		}

		if len(qp) > 0 {
			handlerPath = handlerPath + "?" + qp.Encode()
		}

		output.RequestURI = handlerPath

		c := cors.New(REST_CORS_PREFIX, triggerLog)
		c.WriteCorsActualRequestHeaders(w)

		if r.Header.Get("Content-Type") != "" && strings.Split(r.Header.Get("Content-Type"), ";")[0] == "multipart/form-data" {

			requestBody := make(map[string]interface{})

			if log.RootLogger().DebugEnabled() {
				triggerLog.Debugf("Rest trigger body Contains File Data")
			}

			if method == "POST" || method == "PUT" || method == "PATCH" {
				if outputValidate {
					err := doJsonSchemaValiation(handler.Schemas().Output["multipartFormData"], requestBody)
					if err != nil {
						errMsg := fmt.Sprintf("Fail to validate body: %v", err)
						triggerLog.Error(errMsg)
						http.Error(w, errMsg, http.StatusBadRequest)
						return
					}
				}

				err := r.ParseMultipartForm(10 << 20) //In memory Max 10 MB
				if err != nil {
					triggerLog.Error(err)
					return
				}

				for k, v := range r.MultipartForm.Value {
					requestBody[k] = v[0]
				}

				for k, v := range r.MultipartForm.File {

					if len(v) > 0 {
						// Multiple File has been uploaded with the same Key name

						var tempData interface{}

						for _, fileData := range v {
							f, err := fileData.Open()
							if err != nil {
								triggerLog.Error(err)
								return
							}

							fileBytes, err := ioutil.ReadAll(f)
							if err != nil {
								triggerLog.Error(err)
								return
							}

							//DATA UPLOAD ON DISK
							/*
								dataDir, err := os.Stat("./file-data")
								if err != nil {
									// dataDir does nor exist, create one
									err := os.MkdirAll("./file-data", os.ModePerm)
									if err != nil {
										triggerLog.Error("Error in creating Storage Dir for storing Files")
										return
									}
									dataDir, _ = os.Stat("./file-data")
								}

								dirName := "./"+dataDir.Name()

								tempFile, err := ioutil.TempFile(dirName, "upload-*")
								if err != nil {
									triggerLog.Error(err)
									return
								}
								tempFile.Write(fileBytes)
							*/

							if tempData == nil || len(tempData.([][]byte)) == 0 {
								tempData = [][]byte{fileBytes}
							} else {
								tempData = append(tempData.([][]byte), fileBytes)
							}
						}
						requestBody[k] = tempData
					} else {
						if len(v) != 0 {
							f, err := v[0].Open()
							if err != nil {
								triggerLog.Error(err)
								return
							}

							fileBytes, err := ioutil.ReadAll(f)
							if err != nil {
								triggerLog.Error(err)
								return
							}

							// DATA UPLOAD ON DISK
							/*
								dataDir, err := os.Stat("./file-data")
								if err != nil {
									// dataDir does nor exist, create one
									err := os.MkdirAll("./file-data", os.ModePerm)
									if err != nil {
										triggerLog.Error("Error in creating Storage Dir for storing Files")
										return
									}
								}

								tempFile, err := ioutil.TempFile("./"+dataDir.Name(), "upload-")
								if err != nil {
									triggerLog.Error(err)
									return
								}
								tempFile.Write(fileBytes)
							*/

							requestBody[k] = [][]byte{fileBytes}

						} else {
							triggerLog.Error("No Files selected for %s", k)
							return
						}
					}
				}

			}

			//triggerLog.Info("*****REQUIRED CHECKING******")
			//params := handler.Schemas().Output["multipartForm"].(map[string]interface{})["fe_metadata"]
			//triggerLog.Info(params)
			//triggerLog.Info("****REAL_REQUEST_BODY******")
			//triggerLog.Info(requestBody)

			//checking for the required param
			stringParams := handler.Schemas().Output["multipartForm"].(map[string]interface{})["fe_metadata"].(string)
			var params []map[string]interface{}
			err := json.Unmarshal([]byte(stringParams), &params)

			if err != nil {
				triggerLog.Error(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			for _, v := range params {
				//var x map[string]interface{}
				//err := json.Unmarshal([]byte(v.(string)),&x)
				//if err != nil {
				//	triggerLog.Error(err)
				//	http.Error(w, err.Error(), http.StatusBadRequest)
				//	return
				//}
				//triggerLog.Info("*****REQUIRED CHECKING******")
				//triggerLog.Info(v)

				if v["required"] == true || v["required"] == "true" {
					//triggerLog.Info("Required")
					if _, ok := requestBody[v["name"].(string)]; !ok {
						errMsg := fmt.Sprintf("Required Form [%s] is not set", v["name"].(string))
						triggerLog.Error(errMsg)
						http.Error(w, errMsg, http.StatusBadRequest)
						return
					}
				}
			}

			output.Multipart = requestBody

		} else if r.Header.Get("Content-Type") != "" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {

			requestBody := make(map[string]interface{})

			if method == "POST" || method == "PUT" || method == "PATCH" {
				err := r.ParseForm()
				if err != nil {
					triggerLog.Error(err)
					return
				}

				for k, v := range r.Form {
					requestBody[k] = v[0]
				}

				if outputValidate {
					err := doJsonSchemaValiation(handler.Schemas().Output["body"], requestBody)
					if err != nil {
						errMsg := fmt.Sprintf("Fail to validate body: %v", err)
						triggerLog.Error(errMsg)
						http.Error(w, errMsg, http.StatusBadRequest)
						return
					}
				}
			}
			output.Body = requestBody

		} else {
			requestBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				switch {
				case err == io.EOF:
					// empty body
					//todo should handler say if content is expected?
				case err != nil:
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}

			if log.RootLogger().DebugEnabled() {
				triggerLog.Debugf("Rest trigger body: %s", string(requestBody))
			}

			if outputValidate && (method == "POST" || method == "PUT" || method == "PATCH") {
				err := doJsonSchemaValiation(handler.Schemas().Output["body"], string(requestBody))
				if err != nil {
					errMsg := fmt.Sprintf("Fail to validate body: %v", err)
					triggerLog.Error(errMsg)
					http.Error(w, errMsg, http.StatusBadRequest)
					return
				}
			}
			output.Body = requestBody
		}

		gContext := context.Background()
		if trace.Enabled() {
			tracingContext, _ := trace.GetTracer().Extract(trace.HTTPHeaders, r)
			gContext = trace.AppendTracingContext(gContext, tracingContext)
		}

		tags := make(map[string]string, 2)
		tags["method"] = method
		tags["path"] = handlerSetting.Path
		evtContext := trigger.AppendEventDataToContext(gContext, tags)

		if w.Header().Get("X-Request-Id") != "" {
			evtContext = trigger.NewContextWithEventId(evtContext, w.Header().Get("X-Request-Id"))
		}

		results, err := handler.Handle(evtContext, output)

		var replyData interface{}
		var replyCode int

		if len(results) != 0 {

			// for serverless text message reply
			txtMsgAttr, ok := results["message"]
			if ok && txtMsgAttr != nil {
				attrValue := txtMsgAttr

				if attrValue != nil {
					complexV := attrValue
					if ok {
						replyData = complexV
					} else {
						replyData = attrValue
					}
				}
			}

			//handling responses
			headerData, ok := results["responseHeaders"]
			if ok && headerData != nil {
				triggerLog.Info("headerinfo: ", headerData)
			}

			dataAttr, ok := results["data"]
			if ok && dataAttr != nil {
				attrValue := dataAttr

				if attrValue != nil {
					complexV := attrValue
					if ok {
						replyData = complexV
					} else {
						replyData = attrValue
					}
				}
			}

			codeAttr, ok := results["code"]
			if ok {
				replyCode, _ = coerce.ToInt(codeAttr)

			}

			responseCodedata, ok := results["responseBody"]
			if ok && responseCodedata != nil {
				responseCodeValue := responseCodedata

				if responseCodeValue != nil {

					// This is the response code condition where the body will contain the response.
					mapvalue, ok1 := responseCodeValue.(map[string]interface{})
					if ok1 {

						//Handle Response Body
						if _, ok := mapvalue["body"]; ok {
							replyData = mapvalue["body"]
						} else {
							replyData = responseCodeValue
						}

						//Handle Response Headers
						if _, ok := mapvalue["headers"]; ok {
							if mapvalue["headers"] != nil {
								headers := mapvalue["headers"].(map[string]interface{})

								for key, value := range headers {
									if key != "Content-Length" {
										headerValue := fmt.Sprintf("%v", value)
										w.Header().Set(key, headerValue)
									}
								}
							}
						}
					} else {
						replyData = responseCodeValue
					}
				}
			}
		}

		if err != nil {
			triggerLog.Errorf("REST Trigger Error: %s for request body %s", err.Error(), fmt.Sprintf("%s", output.Body))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		triggerLog.Debugf("The reply http code is: %d", replyCode)
		if replyData != nil {
			if replyCode == 0 {
				replyCode = 200
			}

			switch t := replyData.(type) {
			case string:
				if len(t) > 0 {
					if t[0] == '{' || t[0] == '[' {
						w.Header().Set("Content-Type", "application/json; charset=UTF-8")
					} else {
						if w.Header().Get("Content-Type") == "" {
							// default to text/plain
							w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
						}
					}
				}
				//Comment as it has performance impact
				//var v interface{}
				//err := json.Unmarshal([]byte(t), &v)
				//if err != nil {
				//w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
				//} else {
				//	//Json
				//	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				//}
				w.WriteHeader(replyCode)
				_, err = w.Write([]byte(t))
				if err != nil {
					triggerLog.Error("Failed to send response due to error - %s", err.Error())
					if errors.Is(err, syscall.EPIPE) {
						// Looks like connection is closed
						triggerLog.Error("Looks like connection is closed while request is being processed by the Flogo engine. Check timeout configuration(s) of client and/or API Management service (if configured).")
					}

				}
				triggerLog.Debugf("REST Trigger: Total %d ms Taken for the RequestId: %s", int64(time.Since(startTimestamp)/time.Millisecond), w.Header().Get("X-Request-Id"))
				return
			default:
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(replyCode)
				en := json.NewEncoder(w)
				//en.SetIndent("", "  ")
				if err := en.Encode(replyData); err != nil {
					triggerLog.Error("Failed to send response due to error - %s", err.Error())
					if errors.Is(err, syscall.EPIPE) {
						// Looks like connection is closed
						triggerLog.Error("Looks like connection is closed while request is being processed by Flogo engine. Check timeout configuration(s) for client and/or API Management service (if configured).")
					}
				}
				triggerLog.Debugf("REST Trigger: Total %d ms Taken for the RequestId: %s", int64(time.Since(startTimestamp)/time.Millisecond), w.Header().Get("X-Request-Id"))
				return
			}
		}
		if replyCode > 0 {
			w.WriteHeader(replyCode)

		} else {
			w.WriteHeader(http.StatusOK)
		}

		if r.Body != nil {
			_, _ = io.Copy(ioutil.Discard, r.Body)
			r.Body.Close()
		}

		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}

		triggerLog.Debugf("REST Trigger: Total %d ms Taken for the RequestId: %s", int64(time.Since(startTimestamp)/time.Millisecond), w.Header().Get("X-Request-Id"))
	}
}

////////////////////////////////////////////////////////////////////////////////////////
// Utils

// For query and header, since go http treats query/header value as string,
// directly type check as validation
func getValuewithType(param Parameter, sv []string) ([]interface{}, error) {
	var values []interface{}
	switch param.Type { // json schema data type
	case "number":
		if param.Repeating == "false" {
			v, err := strconv.ParseFloat(sv[0], 64)
			if err != nil {
				return nil, fmt.Errorf("value %s is not a %s type", sv[0], param.Type)
			}
			values = append(values, v)
		} else {
			for _, item := range sv {
				hvalues := strings.Split(item, ",")
				if len(hvalues) > 0 {
					for _, val := range hvalues {
						v, err := strconv.ParseFloat(val, 64)
						if err != nil {
							return nil, fmt.Errorf("value %s is not a %s type", item, param.Type)
						}
						values = append(values, v)
					}
				} else {
					v, err := strconv.ParseFloat(item, 64)
					if err != nil {
						return nil, fmt.Errorf("value %s is not a %s type", item, param.Type)
					}
					values = append(values, v)
				}
			}
		}

	case "integer":
		if param.Repeating == "false" {
			v, err := strconv.ParseInt(sv[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("value %s is not a %s type", sv[0], param.Type)
			}
			values = append(values, v)
		} else {
			for _, item := range sv {
				hvalues := strings.Split(item, ",")
				if len(hvalues) > 0 {
					for _, val := range hvalues {
						v, err := strconv.ParseInt(val, 10, 64)
						if err != nil {
							return nil, fmt.Errorf("value %s is not a %s type", item, param.Type)
						}
						values = append(values, v)
					}
				} else {
					v, err := strconv.ParseInt(item, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("value %s is not a %s type", item, param.Type)
					}
					values = append(values, v)
				}
			}
		}

	case "boolean":
		if param.Repeating == "false" {
			v, err := strconv.ParseBool(sv[0])
			if err != nil {
				return nil, fmt.Errorf("value %s is not a %s type", sv[0], param.Type)
			}
			values = append(values, v)

		} else {
			for _, item := range sv {
				hvalues := strings.Split(item, ",")
				if len(hvalues) > 0 {
					for _, val := range hvalues {
						v, err := strconv.ParseBool(val)
						if err != nil {
							return nil, fmt.Errorf("value %s is not a %s type", item, param.Type)
						}
						values = append(values, v)
					}
				} else {
					v, err := strconv.ParseBool(item)
					if err != nil {
						return nil, fmt.Errorf("value %s is not a %s type", item, param.Type)
					}
					values = append(values, v)
				}
			}
		}
	case "string":
		if param.Repeating == "false" {
			v, err := coerce.ToString(sv[0])
			if err != nil {
				return nil, err
			}
			values = append(values, v)

		} else {
			for _, item := range sv {
				v, err := coerce.ToString(item)
				if err != nil {
					return nil, err
				}
				values = append(values, v)
			}
		}
	}
	return values, nil
}

func doJsonSchemaValiation(co interface{}, data interface{}) error {
	rawSchema, err := schema.FindOrCreate(co)
	if err != nil {
		return err
	}
	schemaString := rawSchema.Value()
	if schemaString != "" {
		switch data.(type) {
		case string:
			err := jsonschema.ForceValidate(schemaString, data.(string))
			return err
		default:
			// object
			err := jsonschema.ForceValidateFromObject(schemaString, data)
			return err
		}
	}
	return nil
}

func handlerIsValid(handler *HandlerSettings) bool {

	if handler.Method == "" {
		return false
	}

	if !stringInList(strings.ToUpper(handler.Method), validMethods) {
		return false
	}

	//validate path

	return true
}

func stringInList(str string, list []string) bool {
	for _, value := range list {
		if value == str {
			return true
		}
	}
	return false
}

func notEmpty(array []string) bool {
	if len(array) > 0 {
		if len(array) == 1 {
			if array[0] != "" && len(array[0]) > 0 {
				return true
			}
			return false
		} else {
			return true
		}
	}
	return false
}

// func (t *Trigger) isFlowControlled() bool {
// 	if app.EnableFlowControl() {
// 		t.flowLimitLock.RLock()
// 		defer t.flowLimitLock.RUnlock()
// 	}
// 	return t.flowControlled
// }

// func (t *Trigger) controlFlow() {
// 	t.flowLimitLock.Lock()
// 	defer t.flowLimitLock.Unlock()
// 	t.flowControlled = true
// }

// func (t *Trigger) releaseFlowControl() {
// 	t.flowLimitLock.Lock()
// 	defer t.flowLimitLock.Unlock()
// 	t.flowControlled = false
// }
