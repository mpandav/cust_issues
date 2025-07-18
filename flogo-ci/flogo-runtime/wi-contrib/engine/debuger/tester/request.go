package tester

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"

	appresolve "github.com/project-flogo/core/app/resolve"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/data/schema"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/engine/runner"
	coreSupport "github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
	"github.com/project-flogo/flow/instance"
	"github.com/project-flogo/flow/support"
)

const (
	RefFlow = "github.com/project-flogo/flow"
)

// RequestProcessor processes request objects and invokes the corresponding
// flow Manager methods
type RequestProcessor struct {
	runner action.Runner
	logger log.Logger
}

// NewRequestProcessor creates a new RequestProcessor
func NewRequestProcessor() *RequestProcessor {

	var rp RequestProcessor
	rp.runner = runner.NewPooled(&runner.PooledConfig{
		NumWorkers:    1,
		WorkQueueSize: 5,
	})
	//rp.runner = runner.NewDirect()
	//todo what logger should this use?
	rp.logger = log.RootLogger()

	return &rp
}

// StartFlow handles a StartRequest for a FlowInstance.  This will
// generate an ID for the new FlowInstance and queue a StartRequest.
func (rp *RequestProcessor) StartFlow(startRequest *StartRequest) (results map[string]interface{}, err error) {

	logger := rp.logger

	logger.Debugf("Tester starting flow")

	factory := action.GetFactory(RefFlow)
	settings := map[string]interface{}{"flowURI": startRequest.FlowURI}
	act, err := factory.New(&action.Config{Settings: settings})
	if err != nil {
		return nil, err
	}

	var inputs map[string]interface{}

	if len(startRequest.Attrs) > 0 {

		logger.Debugf("Starting with flow attrs: %#v", startRequest.Attrs)

		inputs = make(map[string]interface{}, len(startRequest.Attrs)+1)
		for name, value := range startRequest.Attrs {
			inputs[name] = value
		}
	} else if len(startRequest.Data) > 0 {

		logger.Debugf("Starting with flow attrs: %#v", startRequest.Data)

		inputs = make(map[string]interface{}, len(startRequest.Data)+1)

		for k, v := range startRequest.Data {
			//t, err := data.GetType(v)
			//if err != nil {
			//	t = data.TypeAny
			//}
			//attr, _ := data.NewAttribute(k, t, v)
			inputs[k] = v
		}
	} else {
		inputs = make(map[string]interface{}, 1)
	}

	execOptions := &instance.ExecOptions{Interceptor: startRequest.Interceptor, Patch: startRequest.Patch}
	ro := &instance.RunOptions{Op: instance.OpStart, ReturnID: true, FlowURI: startRequest.FlowURI, ExecOptions: execOptions}
	//attr, _ := data.NewAttribute("_run_options", data.TypeAny, ro)
	inputs["_run_options"] = ro

	return rp.runner.RunAction(trigger.NewContextWithEventId(context.Background(), "tester_event_id"), act, inputs)
}

func initFlowStart(config *app.Config) error {

	RegistryImport(config)

	properties := make(map[string]interface{}, len(config.Properties))
	for _, attr := range config.Properties {
		properties[attr.Name()] = attr.Value()
	}

	manger := property.NewManager(properties)
	property.SetDefaultManager(manger)

	resolver := resolve.NewCompositeResolver(map[string]resolve.Resolver{
		".":        &resolve.ScopeResolver{},
		"env":      &resolve.EnvResolver{},
		"property": &property.Resolver{},
		"loop":     &resolve.LoopResolver{},
	})

	appresolve.SetAppResolver(resolver)

	//for _, anImport := range config.Imports {
	//	matches := flogoImportPattern.FindStringSubmatch(anImport)
	//	err := registerImport(matches[1] + matches[3] + matches[5]) // alias + module path + relative import path
	//	if err != nil {
	//		log.RootLogger().Errorf("cannot register import '%s' : %v", anImport, err)
	//	}
	//}

	//function.ResolveAliases()

	// register schemas, assumes appropriate schema factories have been registered
	for id, def := range config.Schemas {
		_, err := schema.Register(id, def)
		if err != nil {
			return err
		}
	}

	schema.ResolveSchemas()
	for id, config := range config.Connections {
		_, err := connection.NewSharedManager(id, config)
		if err != nil {
			return err
		}
	}
	return nil
}

// RestartFlow handles a RestartRequest for a FlowInstance.  This will
// generate an ID for the new FlowInstance and queue a RestartRequest.
func (rp *RequestProcessor) RestartFlow(restartRequest *RestartRequest) (results map[string]interface{}, err error) {

	logger := rp.logger

	logger.Debugf("Tester restarting flow")

	//todo share action, for now add flowUri to settings
	settings := map[string]interface{}{"flowURI": restartRequest.InitialState.FlowURI()}

	factory := action.GetFactory(RefFlow)
	act, err := factory.New(&action.Config{Settings: settings})
	if err != nil {
		return nil, err
	}

	inputs := make(map[string]interface{}, len(restartRequest.Data)+1)

	if restartRequest.Data != nil {

		logger.Debugf("Updating flow attrs: %v", restartRequest.Data)

		for k, v := range restartRequest.Data {
			//attr, _ := data.NewAttribute(k, data.TypeAny, v)
			inputs[k] = v
		}
	}

	execOptions := &instance.ExecOptions{Interceptor: restartRequest.Interceptor, Patch: restartRequest.Patch}
	ro := &instance.RunOptions{Op: instance.OpRestart, ReturnID: restartRequest.ReturnID, FlowURI: restartRequest.InitialState.FlowURI(), InitialState: restartRequest.InitialState, ExecOptions: execOptions}
	//attr, _ := data.NewAttribute("_run_options", data.TypeAny, ro)
	inputs["_run_options"] = ro

	return rp.runner.RunAction(trigger.NewContextWithEventId(context.Background(), "tester_event_id"), act, inputs)
}

// ResumeFlow handles a ResumeRequest for a FlowInstance.  This will
// queue a RestartRequest.
func (rp *RequestProcessor) ResumeFlow(resumeRequest *ResumeRequest) (results map[string]interface{}, err error) {

	logger := rp.logger

	logger.Debugf("Tester resuming flow")

	//todo share action, for now add flowUri to settings
	settings := map[string]interface{}{"flowURI": resumeRequest.State.FlowURI()}

	factory := action.GetFactory(RefFlow)
	act, err := factory.New(&action.Config{Settings: settings})
	if err != nil {
		return nil, err
	}

	inputs := make(map[string]interface{}, len(resumeRequest.Data)+1)

	if resumeRequest.Data != nil {

		logger.Debugf("Updating flow attrs: %v", resumeRequest.Data)

		for k, v := range resumeRequest.Data {
			//attr, _ := data.NewAttribute(k, data.TypeAny, v)
			inputs[k] = v
		}
	}

	execOptions := &instance.ExecOptions{Interceptor: resumeRequest.Interceptor, Patch: resumeRequest.Patch}
	ro := &instance.RunOptions{Op: instance.OpResume, ReturnID: resumeRequest.ReturnID, FlowURI: resumeRequest.State.FlowURI(), InitialState: resumeRequest.State, ExecOptions: execOptions}
	//attr, _ := data.NewAttribute("_run_options", data.TypeAny, ro)
	//inputs[attr.Name()] = attr
	inputs["_run_options"] = ro
	return rp.runner.RunAction(trigger.NewContextWithEventId(context.Background(), "tester_event_id"), act, inputs)
}

// StartRequest describes a request for starting a FlowInstance
type StartRequest struct {
	FlowURI     string                 `json:"flowUri"`
	App         *app.Config            `json:"app"`
	Data        map[string]interface{} `json:"data"`
	Attrs       map[string]interface{} `json:"attrs"`
	Interceptor *support.Interceptor   `json:"interceptor"`
	Patch       *support.Patch         `json:"patch"`
	ReplyTo     string                 `json:"replyTo"`
}

func (s *StartRequest) UnmarshalJSON(d []byte) error {

	appStartRequest := struct {
		FlowURI     string                 `json:"flowUri"`
		FlogoJson   interface{}            `json:"flogoJson"`
		Data        map[string]interface{} `json:"data"`
		Attrs       map[string]interface{} `json:"attrs"`
		Interceptor *support.Interceptor   `json:"interceptor"`
		Patch       *support.Patch         `json:"patch"`
		ReplyTo     string                 `json:"replyTo"`
	}{}

	err := json.Unmarshal(d, &appStartRequest)
	if err != nil {
		return nil
	}

	s.FlowURI = appStartRequest.FlowURI
	s.Attrs = appStartRequest.Attrs
	s.Data = appStartRequest.Data
	s.Interceptor = appStartRequest.Interceptor
	s.ReplyTo = appStartRequest.ReplyTo
	s.Patch = appStartRequest.Patch
	str, _ := coerce.ToString(appStartRequest.FlogoJson)
	app, err := engine.LoadAppConfig(str, false)
	if err != nil {
		return err
	}
	s.App = app
	return nil
}

// RestartRequest describes a request for restarting a FlowInstance
// todo: can be merged into StartRequest
type RestartRequest struct {
	InitialState *instance.IndependentInstance `json:"initialState"`
	Data         map[string]interface{}        `json:"data"`
	Interceptor  *support.Interceptor          `json:"interceptor"`
	Patch        *support.Patch                `json:"patch"`
	ReturnID     bool                          `json:"returnId"`
}

// ResumeRequest describes a request for resuming a FlowInstance
//todo: Data for resume request should be directed to waiting task
type ResumeRequest struct {
	State       *instance.IndependentInstance `json:"state"`
	Data        map[string]interface{}        `json:"data"`
	Interceptor *support.Interceptor          `json:"interceptor"`
	Patch       *support.Patch                `json:"patch"`
	ReturnID    bool                          `json:"returnId"`
}

var flogoImportPattern = regexp.MustCompile(`^(([^ ]*)[ ]+)?([^@:]*)@?([^:]*)?:?(.*)?$`) // extract import path even if there is an alias and/or a version

func RegistryImport(config *app.Config) {
	for _, anImport := range config.Imports {
		matches := flogoImportPattern.FindStringSubmatch(anImport)
		registerImport(matches[1] + matches[3] + matches[5]) // alias + module path + relative import path
	}
	function.ResolveAliases()
}

func registerImport(anImport string) error {

	parts := strings.Split(anImport, " ")

	var alias string
	var ref string
	numParts := len(parts)
	if numParts == 1 {
		ref = parts[0]
		alias = path.Base(ref)
	} else if numParts == 2 {
		alias = parts[0]
		ref = parts[1]
	} else {
		return fmt.Errorf("invalid import %s", anImport)
	}

	if alias == "" || ref == "" {
		return fmt.Errorf("invalid import %s", anImport)
	}

	ct := getContribType(ref)
	if ct == "other" {
		log.RootLogger().Debugf("Added Non-Contribution Import: %s", ref)
		return nil
		//return fmt.Errorf("invalid import, contribution '%s' not registered", anImport)
	}

	log.RootLogger().Debugf("Registering type alias '%s' for %s [%s]", alias, ct, ref)

	err := coreSupport.RegisterAlias(ct, alias, ref)
	if err != nil {
		return err
	}

	if ct == "function" {
		function.SetPackageAlias(ref, alias)
	}

	return nil
}

func getContribType(ref string) string {
	if activity.Get(ref) != nil {
		return "activity"
	} else if action.GetFactory(ref) != nil {
		return "action"
	} else if trigger.GetFactory(ref) != nil {
		return "trigger"
	} else if function.IsFunctionPackage(ref) {
		return "function"
	} else if connection.GetManagerFactory(ref) != nil {
		return "connection"
	} else {
		return "other"
	}
}
