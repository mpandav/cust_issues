package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/core/trigger"
)

var triggerMd = trigger.NewMetadata(&Settings{}, &HandlerSettings{}, &Output{})

func init() {
	_ = trigger.Register(&Trigger{}, &Factory{})
}

type Factory struct {
}

// Metadata implements trigger.Factory.Metadata
func (*Factory) Metadata() *trigger.Metadata {
	return triggerMd
}

// New implements trigger.Factory.New
func (*Factory) New(config *trigger.Config) (trigger.Trigger, error) {
	s := &Settings{}
	err := metadata.MapToStruct(config.Settings, s, true)
	if err != nil {
		return nil, err
	}

	return &Trigger{id: config.Id, settings: s}, nil
}

// Trigger REST trigger struct
type Trigger struct {
	server     *http.Server
	settings   *Settings
	id         string
	logger     log.Logger
	mode       string
	handlerMap map[string]*HandlerInfo
}

type HandlerInfo struct {
	settings *HandlerSettings
	handler  trigger.Handler
}

const (
	ModeOpaque = "Proxy"
	ModeData   = "Data"
)

var CorrelationHeaderList = []string{"X-Atmosphere-Request-Id", "X-Request-ID", "X-Correlation-ID"}

func (t *Trigger) Initialize(ctx trigger.InitContext) (err error) {

	t.logger = ctx.Logger()

	t.handlerMap = make(map[string]*HandlerInfo)

	//data mode
	router := httprouter.New()
	addr := ":" + strconv.Itoa(t.settings.Port)
	processingMode := t.settings.Mode
	if processingMode == "" {
		processingMode = ModeOpaque
	}
	// Init handlers
	for _, handler := range ctx.GetHandlers() {
		handlerSettings := &HandlerSettings{}
		err := metadata.MapToStruct(handler.Settings(), handlerSettings, true)
		if err != nil {
			return err
		}
		if processingMode == ModeData {
			router.Handle(handlerSettings.Method, replacePath(handlerSettings.ContextPath), newActionHandler(t, handler))
			t.logger.Infof("Context path '%s' registered for data handler '%s' with '%s' method", handlerSettings.ContextPath, handler.Name(), handlerSettings.Method)
		} else if processingMode == ModeOpaque {
			handlerInfo := &HandlerInfo{}
			handlerInfo.settings = handlerSettings
			handlerInfo.handler = handler
			handlerSettings.ContextPath = normalize(handlerSettings.ContextPath)
			registered, match, _ := t.isRegistered(handlerSettings.ContextPath, handler.Name())
			if registered {
				return fmt.Errorf("Found two handlers with conflicting base paths - '%s' and '%s'. The base path must be unique and should not conflict with '/'.", handlerSettings.ContextPath, match)
			}
			t.handlerMap[handlerSettings.ContextPath] = handlerInfo
			t.logger.Infof("Context path '%s' registered for handler '%s'", handlerSettings.ContextPath, handler.Name())
		}

	}

	if processingMode == ModeOpaque {
		mux := http.NewServeMux()
		mux.HandleFunc("/", t.handlerFunction)
		t.server = &http.Server{Handler: mux}
	} else {
		t.server = &http.Server{
			Addr:    addr,
			Handler: router,
		}
	}
	return nil
}

func normalize(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return path
}

func (t *Trigger) isRegistered(bPath, hName string) (bool, string, error) {

	h, found := t.handlerMap["/"]
	if found {
		// generic base path is already registered
		return false, "", fmt.Errorf("Context path '%s' configured for handler '%s' conflicts with context path '/' configured for handler '%s'. Either use only '/' context path in the trigger or remove it.", bPath, hName, h.handler.Name())
	}
	if bPath == "/" && len(t.handlerMap) > 1 {
		return false, "", fmt.Errorf("Context path '/' configured for handler '%s' conflicts with context path(s) configured for other handler(s). Either use only '/' context path in the trigger or remove it.", hName)
	}

	for k, h := range t.handlerMap {
		regPath := path.Base(k)
		basePath := path.Base(bPath)
		if strings.EqualFold(regPath, basePath) {
			return false, "", fmt.Errorf("Context path '%s' configured for handler '%s' conflicts with context path '%s' configured for handler '%s'. Ensure base context paths are unique e.g. /a/b and /x.", bPath, hName, k, h.handler.Name())
		}
	}
	return false, "", nil
}

func replacePath(path string) string {
	path = strings.Replace(path, "}", "", -1)
	return strings.Replace(path, "{", ":", -1)
}

func (t *Trigger) Start() error {
	if t.server.Addr == "" {
		t.server.Addr = ":http"
	}
	if t.settings.Mode == ModeOpaque {
		go t.startHttpServer()
	} else {
		go func() {

			t.logger.Infof("HTTP server started on port - %d", t.settings.Port)

			if err := t.server.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					t.logger.Error(err)
				}
			}
		}()
	}
	return nil
}

func (t *Trigger) startHttpServer() {
	port, err := coerce.ToString(t.settings.Port)
	if err != nil {
		t.logger.Errorf("Invalid port number - %d", t.settings.Port)
		panic(err.Error())
	}
	t.logger.Infof("HTTP server started on port - %s", port)
	l, err := net.Listen("tcp4", ":"+port)
	if err != nil {
		t.logger.Errorf("Failed to start server due to error - %s", err.Error())
		panic(err.Error())
	}
	err = t.server.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		t.logger.Errorf("Failed to start server due to error - %s", err.Error())
		panic(err.Error())
	}
}

// Stop implements util.Managed.Stop
func (t *Trigger) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := t.server.Shutdown(ctx)
	return err
}

func (t *Trigger) handlerFunction(w http.ResponseWriter, r *http.Request) {
	matchingPath, handlerInfo := t.findMatchingHandler(r.URL.Path)
	if handlerInfo.handler == nil {
		t.logger.Errorf("Handler not found for '%s'. Check trigger configuration.", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	urlPath := r.URL.Path
	handlerInfo.handler.Logger().Debugf("Handler '%s' is processing request '%s %s'", handlerInfo.handler.Name(), r.Method, r.URL.RequestURI())
	filterCICHeaders(r.Header, handlerInfo.handler.Logger())
	handlerInfo.handler.Logger().Debugf("Headers: %v", r.Header)

	removeBasePath(r, matchingPath)
	output := &Output{}
	output.ProxyData = make(map[string]interface{})
	output.ProxyData["requestObject"] = r
	output.ProxyData["responseObject"] = w
	eventId := getRequestIdFromRequest(r)
	ctx := context.Background()
	if eventId != "" {
		ctx = trigger.NewContextWithEventId(ctx, eventId)
	}
	_, err := handlerInfo.handler.Handle(ctx, output)
	if err != nil {
		t.logger.Errorf("Failed to process request for '%s' due to error '%s'.", urlPath, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func newActionHandler(t *Trigger, handler trigger.Handler) httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		urlPath := r.URL.Path
		out := &Output{}
		out.Method = r.Method

		filterCICHeaders(r.Header, t.logger)
		t.logger.Debugf("Headers: %v", r.Header)

		out.PathParams = make(map[string]string, len(ps))
		for _, param := range ps {
			out.PathParams[param.Key] = param.Value
		}

		queryValues := r.URL.Query()
		out.QueryParams = make(map[string]string, len(queryValues))
		out.Headers = make(map[string]string, len(r.Header))

		for key, value := range r.Header {
			out.Headers[key] = strings.Join(value, ",")
		}

		for key, value := range queryValues {
			out.QueryParams[key] = strings.Join(value, ",")
		}

		// Check the HTTP Header Content-Type
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/x-www-form-urlencoded":
			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				t.logger.Debugf("Error reading body: %s", err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			s := buf.String()
			m, err := url.ParseQuery(s)
			if err != nil {
				t.logger.Debugf("Error parsing query string: %s", err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			content := make(map[string]interface{}, 0)
			for key, val := range m {
				if len(val) == 1 {
					content[key] = val[0]
				} else {
					content[key] = val[0]
				}
			}

			out.RequestBody = content
		case "application/json":
			var content interface{}
			err := json.NewDecoder(r.Body).Decode(&content)
			if err != nil {
				switch {
				case err == io.EOF:
					// empty body
					//todo what should handler say if content is expected?
				default:
					t.logger.Debugf("Error parsing json body: %s", err.Error())
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
			out.RequestBody = content
		default:
			if strings.Contains(contentType, "multipart/form-data") {
				// need to still extract the body, only handling the multipart data for now...

				if err := r.ParseMultipartForm(32); err != nil {
					t.logger.Debugf("Error parsing multipart form: %s", err.Error())
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				var files []map[string]interface{}

				for key, fh := range r.MultipartForm.File {
					for _, header := range fh {

						fileDetails, err := getFileDetails(key, header)
						if err != nil {
							t.logger.Debugf("Error getting attached file details: %s", err.Error())
							http.Error(w, err.Error(), http.StatusBadRequest)
							return
						}

						files = append(files, fileDetails)
					}
				}

				// The content output from the trigger
				content := map[string]interface{}{
					"body":  nil,
					"files": files,
				}
				out.RequestBody = content
			} else {
				reqBody, err := ioutil.ReadAll(r.Body)
				if err != nil {
					t.logger.Errorf("Failed to process request for '%s' due to error '%s'.", urlPath, err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				out.RequestBody = string(reqBody)
			}
		}

		//adding tracing context
		gContext := context.Background()
		if trace.Enabled() {
			tracingContext, _ := trace.GetTracer().Extract(trace.HTTPHeaders, r)
			gContext = trace.AppendTracingContext(gContext, tracingContext)
			tags := make(map[string]string, 2)
			tags["method"] = r.Method
			tags["path"] = r.URL.Path
			gContext = trigger.AppendEventDataToContext(gContext, tags)
		}

		if w.Header().Get("X-Request-Id") != "" {
			gContext = trigger.NewContextWithEventId(gContext, w.Header().Get("X-Request-Id"))
		}

		results, err := handler.Handle(gContext, out)
		if err != nil {
			t.logger.Debugf("Error handling request: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if t.logger.TraceEnabled() {
			t.logger.Tracef("Action Results: %#v", results)
		}

		/// populate response from trigger
		reply := &Reply{}
		err = reply.FromMap(results)
		if err != nil {
			t.logger.Debugf("Error mapping results: %s", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(reply.Cookies) > 0 {
			if t.logger.TraceEnabled() {
				t.logger.Tracef("Adding Cookies")
			}

			err := addCookies(w, reply.Cookies)
			if err != nil {
				t.logger.Debugf("Error handling request: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// add response headers
		if len(reply.Headers) > 0 {
			if t.logger.TraceEnabled() {
				t.logger.Tracef("Adding Headers")
			}
			for key, value := range reply.Headers {
				w.Header().Set(key, value)
			}
		}

		//add response statuscode
		w.WriteHeader(reply.StatusCode)
		t.logger.Debugf("The reply http code is: %d", reply.StatusCode)

		if reply.ResponseBody != nil {
			bytes, err := coerce.ToBytes(reply.ResponseBody)
			if err != nil {
				t.logger.Debugf("Error handling request: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if w.Header().Get("Content-Type") == "" {
				w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
			}
			_, err = w.Write(bytes)
			if err != nil {
				t.logger.Debugf("Error writing body: %s", err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
}

func removeBasePath(r *http.Request, path1 string) {
	if path1 == "/" {
		return
	}
	if r.URL.Path+"/" == path1 {
		// remove trailing / for case like /foo
		path1 = path1[:len(path1)-1]
	}
	r.URL.Path = strings.Replace(r.URL.Path, path1, "/", 1)
}

func filterCICHeaders(header http.Header, logger log.Logger) {
	var headerList []string
	for h := range header {
		hName := strings.ToLower(h)
		// Collect all headers start with X-Atmosphere
		if strings.HasPrefix(hName, "x-atmosphere") {
			headerList = append(headerList, hName)
		}
	}

	// Remove headers from request
	logger.Debug("Filtering CIC headers...")
	for _, v := range headerList {
		header.Del(v)
	}
}

func getRequestIdFromRequest(r *http.Request) string {
	for i := range CorrelationHeaderList {
		id := r.Header.Get(CorrelationHeaderList[i])
		if id != "" {
			return id
		}
	}
	return ""
}

func (t *Trigger) findMatchingHandler(uri string) (string, *HandlerInfo) {
	hInfo := &HandlerInfo{}
	if !strings.HasSuffix(uri, "/") {
		uri = uri + "/"
	}
	for path, h := range t.handlerMap {
		if path == "/" || strings.HasPrefix(uri, path) {
			return path, h
		}
	}
	return "", hInfo
}
