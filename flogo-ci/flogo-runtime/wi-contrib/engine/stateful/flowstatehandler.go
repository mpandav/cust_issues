package stateful

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tibco/wi-contrib/environment"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/engine/runner"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/service"
	"github.com/project-flogo/flow/instance"
	"github.com/project-flogo/flow/state"
	"github.com/project-flogo/flow/support"
	"github.com/tibco/wi-contrib/httpservice"
)

const (
	FlowStateManagerServiceEndpoint = "FLOGO_FLOW_SM_ENDPOINT"
	EnableFlowStatePersistence      = "FLOGO_FLOW_STATE_PERSISTENCE"
	EnableAPIAsyncInvocation        = "FLOGO_FLOW_STATE_ASYNC_INVOCATION"
	MAX_RETRY_COUNT                 = 3
	DO_RETRY                        = true
	ASYNC_CALLING_HEADER            = "Async-Calling"
)

func init() {
	_ = service.RegisterFactory(&FlowStateHandlerFactory{})
}

type FlowStateHandlerFactory struct {
}

// StateRecorder is an implementation of StateRecorder service
// that can access flows via URI
type FlowStateHandler struct {
	host             string
	logger           log.Logger
	enabled          bool
	client           *http.Client
	requestProcessor *RequestProcessor
	asyncAPICall     bool
	subId            string
	sigMap           map[string]chan struct{} // for supporting serial calls of flowstart and flowdone with unblocking app engine
}

func (s FlowStateHandlerFactory) NewService(_ *service.Config) (service.Service, error) {
	enabled, _ := coerce.ToBool(os.Getenv(EnableFlowStatePersistence))
	// asyncAPICall, _ := coerce.ToBool(os.Getenv(EnableAPIAsyncInvocation))
	flowStateHandler := &FlowStateHandler{}
	flowStateHandler.sigMap = make(map[string]chan struct{})
	flowStateHandler.logger = log.ChildLogger(log.RootLogger(), "flowstate.handler")
	asyncAPICall, present := os.LookupEnv(EnableAPIAsyncInvocation)
	if environment.IsTCIEnv() {
		if present {
			flowStateHandler.asyncAPICall, _ = coerce.ToBool(asyncAPICall)
		} else {
			flowStateHandler.asyncAPICall = true
		}

		flowStateHandler.host = environment.GetIntercomURL() + "/gsbc/" + strings.ToUpper(environment.GetTCISubscriptionId()) + "/tci/smserviceflogo-" + strings.ToLower(environment.GetTCISubscriptionId()) + "-system/state_manager_flogo"
		flowStateHandler.subId = strings.ToUpper(environment.GetTCISubscriptionId())

		// fmt.Println("######## Host is " + flowStateHandler.host)
		os.Setenv(support.UserName, strings.ToLower(environment.GetTCISubscriptionId()))
	} else {
		flowStateHandler.asyncAPICall, _ = coerce.ToBool(asyncAPICall)
		flowStateHandler.host = parseEndpoint(os.Getenv(FlowStateManagerServiceEndpoint))
		flowStateHandler.enabled = enabled && flowStateHandler.host != ""
	}
	//flowStateHandler.asyncAPICall = asyncAPICall
	flowStateHandler.requestProcessor = &RequestProcessor{logger: flowStateHandler.logger, runner: runner.NewDirect()}
	httpservice.RegisterHandler("/app/state", http.HandlerFunc(flowStateHandler.stateRecorder))
	httpservice.RegisterHandler("/app/state/instance/restart", http.HandlerFunc(flowStateHandler.RestartFlowFromBeginning))
	httpservice.RegisterHandler("/app/state/task/restart", http.HandlerFunc(flowStateHandler.RestartFlowFromActivity))
	return flowStateHandler, nil
}

func parseEndpoint(ep string) string {
	if strings.HasSuffix(ep, "/") {
		return ep[:len(ep)-1]
	}
	return ep
}

func (sr *FlowStateHandler) Name() string {
	return "FlowStateRecorder"
}

type recorderBody struct {
	StateManagerEndpoint string `json:"sm_endpoint"`
}

type response struct {
	Msg string `json:"msg"`
}

func (sr *FlowStateHandler) stateRecorder(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodOptions:
		handleOption(w)
		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		sr.logger.Info("Disabling flow state persistence....")
		if !sr.enabled {
			sr.logger.Info("Ignoring request as Flow state persistence is not enabled for this application")
			w.WriteHeader(http.StatusNotModified)
			return
		}
		responseBody := &response{}
		responseBody.Msg = "Flow state persistence is now disabled"
		sr.enabled = false
		sr.logger.Info(responseBody.Msg)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(responseBody)
	case http.MethodPost:
		responseBody := &response{}
		if !sr.enabled {
			sr.logger.Info("Enabling flow state persistence....")
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				responseBody.Msg = "Failed enable feature due to incorrect configuration. Check application instance logs."
				sr.logger.Errorf("Failed to process request body due to error - %s", err.Error())
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(responseBody)
				return
			}
			defer req.Body.Close()
			if body != nil && len(body) > 0 {
				var rBody recorderBody
				err = json.Unmarshal(body, &rBody)
				if err != nil {
					responseBody.Msg = "Failed enable feature due to incorrect configuration. Check application instance logs."
					sr.logger.Errorf("Failed to process request body due to error - %s", err.Error())
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(responseBody)
					return
				}
				if rBody.StateManagerEndpoint != "" && sr.host == "" {
					// Set State Manager Endpoint
					sr.logger.Infof("Flow state manager endpoint set to %s", rBody.StateManagerEndpoint)
					sr.host = parseEndpoint(rBody.StateManagerEndpoint)
				}
			}

			// Check host value
			if sr.host == "" {
				responseBody.Msg = "Flow state manager endpoint is not configured for the application. This feature remain disabled. To enable this feature, you must set '" + FlowStateManagerServiceEndpoint + "' for this application. Refer docs for more details."
				sr.logger.Errorf(responseBody.Msg)
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(responseBody)
				return
			}
			sr.enabled = true
			responseBody.Msg = "Flow state persistence is enabled for the application"
			sr.logger.Info(responseBody.Msg)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(responseBody)
			return
		}
		w.WriteHeader(http.StatusNotModified)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// Start implements util.Managed.Start()
func (sr *FlowStateHandler) Start() error {
	// no-op

	sr.client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxConnsPerHost:     100,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}
	if environment.IsTCIEnv() {
		res, err := sr.CheckPersistenceStatus()
		if err != nil {
			//TODO do we need to retry and shouldn't add latency to app performance
		}
		if strings.EqualFold(res, "true") {
			sr.enabled = true
		}
		if sr.enabled {
			sr.logger.Info("Flow state persistence is enabled")
			if sr.asyncAPICall {
				sr.logger.Info("Asynchronous calling of the Flow State Manager is enabled")
			}
		}
	}
	return nil
}

func (sr *FlowStateHandler) CheckPersistenceStatus() (string, error) {
	uri := sr.host + "/v1/app/state/" + support.GetAppName()
	// set header username
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil { //TODO
		return "", err
	}
	req.Header.Set("username", support.GetUserName())

	sr.AddHeaders(req)

	resp, err := sr.client.Do(req)
	var result string
	err1 := json.NewDecoder(resp.Body).Decode(&result)
	if err1 != nil {
		return "", err1
	}
	return result, nil
}

func (sr *FlowStateHandler) AddHeaders(req *http.Request) {

	req.Header.Set("X-ATMOSPHERE-for-USER", "any")
	req.Header.Set("X-Atmosphere-Subscription-Id", sr.subId)
	req.Header.Set("X-Atmosphere-Tenant-Id", "tciapps")

}

// Stop implements util.Managed.Stop()
func (sr *FlowStateHandler) Stop() error {
	// no-op
	if sr.client != nil {
		sr.client.CloseIdleConnections()
		sr.client = nil
	}
	return nil
}

func (sr *FlowStateHandler) RecordStart(state1 *state.FlowState) error {
	state := *state1
	if !sr.enabled {
		return nil
	}
	if sr.asyncAPICall { // making sync to avoid overwriting issue of flowstart with flowend data
		sig := make(chan struct{})
		sr.sigMap[state.FlowInstanceId] = sig
		go sr.RecordStartSyncAsync(state)
	} else {
		sr.RecordStartSyncAsync(state)
	}
	return nil
}

func (sr *FlowStateHandler) RecordStartSyncAsync(state state.FlowState) {
	defer func() { // signal by closing channel
		if sig, exist := sr.sigMap[state.FlowInstanceId]; exist {
			close(sig)
		}
	}()

	uri := sr.host + "/v1/instances/start"
	sr.logger.Debugf("POST record start: %s\n", uri)
	jsonReq, _ := json.Marshal(state)
	sr.logger.Debug("JSON: ", string(jsonReq))
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
	if err != nil {
		sr.logger.Warnf("unable to record flow start: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ATMOSPHERE-for-USER", environment.GetTCISubscriptionUName())
	req.Header.Set("X-Atmosphere-Subscription-Id", environment.GetTCISubscriptionId())
	sr.AddHeaders(req)

	sr.logger.Debug(req.Header.Get("X-Atmosphere-Subscription-Id"))
	sr.logger.Debug(req.Header.Get("X-ATMOSPHERE-for-USER"))

	// req.Header.Set(ASYNC_CALLING_HEADER, strconv.FormatBool(sr.asyncAPICall))
	err1 := sr.ConnectionRetry(req)
	if err1 != nil {
		sr.logger.Warnf("unable to record flow start: %v", err1)
		return
	}
}

func (sr *FlowStateHandler) ConnectionRetry(req *http.Request) error {
	var reTryCounter int = 0
	for {
		resp, err := sr.client.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				sr.logger.Error("Error occured while connecting to Flow State Manager : ", (err.Error()))
				if DO_RETRY && reTryCounter < MAX_RETRY_COUNT {
					dur := time.Duration(reTryCounter*2) * time.Second
					time.Sleep(dur)
					reTryCounter++
					sr.logger.Debugf("Retry attempt %d while connecting to Flow State Manager  : ", reTryCounter)
					continue
				} else {
					if !DO_RETRY {
						sr.logger.Info("Connection Retry not enabled")
					} else {
						sr.logger.Infof("Connection Max Retry attempt %d exhausted with err : %s", MAX_RETRY_COUNT, err)
					}
					return err
				}
			} else {
				sr.logger.Errorf("Connection error while connecting to Flow State Manager is:  %s", err)
				return err
			}
		}
		defer resp.Body.Close()
		defer io.Copy(ioutil.Discard, resp.Body)
		sr.logger.Debug("Response Status:", resp.Status)
		break
	}
	return nil
}

func (sr *FlowStateHandler) ConnectionRetryWithResponse(req *http.Request) (resp *http.Response, err error) {
	var reTryCounter int = 0
	for {
		resp, err = sr.client.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				sr.logger.Errorf("Error occured while connecting to Flow State Manager : ", (err.Error()))
				if DO_RETRY && reTryCounter < MAX_RETRY_COUNT {
					dur := time.Duration(reTryCounter*2) * time.Second
					time.Sleep(dur)
					reTryCounter++
					sr.logger.Debugf("Retry attempt %d while connecting to Flow State Manager  : ", reTryCounter)
					continue
				} else {
					if !DO_RETRY {
						sr.logger.Info("Connection Retry not enabled")
					} else {
						sr.logger.Infof("Connection Max Retry attempt %d exhausted with err : %s", MAX_RETRY_COUNT, err)
					}
					return nil, err
				}
			} else {
				sr.logger.Errorf("Connection error while connecting to Flow State Manager is:  %s", err)
				return nil, err
			}
		}
		// defer resp.Body.Close()
		sr.logger.Debug("Response Status:", resp.Status)
		break
	}
	return resp, nil
}

// RecordSnapshot implements instance.FlowStateHandler.RecordSnapshot
func (sr *FlowStateHandler) RecordSnapshot(snapshot *state.Snapshot) error {
	if !sr.enabled {
		return nil
	}

	uri := sr.host + "/v1/instances/snapshot"

	sr.logger.Debugf("POST Snapshot: %s\n", uri)

	jsonReq, _ := json.Marshal(snapshot)

	sr.logger.Debug("JSON: ", string(jsonReq))

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	sr.AddHeaders(req)

	req.Header.Set(ASYNC_CALLING_HEADER, strconv.FormatBool(sr.asyncAPICall))
	return sr.ConnectionRetry(req)
}

// RecordStep implements instance.FlowStateHandler.RecordStep
func (sr *FlowStateHandler) RecordStep(step *state.Step) error {
	if !sr.enabled {
		return nil
	}
	if sr.asyncAPICall {
		go sr.RecordStepSyncAsynch(step)
	} else {
		sr.RecordStepSyncAsynch(step)
	}
	return nil
}

func (sr *FlowStateHandler) RecordStepSyncAsynch(step *state.Step) {
	uri := sr.host + "/v1/instances/steps"
	sr.logger.Debugf("POST Step: %s\n", uri)
	jsonReq, _ := json.Marshal(step)
	sr.logger.Debug("JSON: ", string(jsonReq))
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
	if err != nil {
		sr.logger.Warnf("unable to record step: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	sr.AddHeaders(req)

	req.Header.Set(ASYNC_CALLING_HEADER, strconv.FormatBool(sr.asyncAPICall))
	err1 := sr.ConnectionRetry(req)
	if err1 != nil {
		sr.logger.Warnf("unable to record step: %v", err1)
		return
	}
}

func (sr *FlowStateHandler) RecordDone(state1 *state.FlowState) error {
	state := *state1
	if !sr.enabled {
		return nil
	}
	if sr.asyncAPICall {
		go sr.RecordDoneSyncAsync(state) // using blocking channel to avoid overwriting issue of flowstart with flowend data
	} else {
		sr.RecordDoneSyncAsync(state)
	}
	return nil
}

func (sr *FlowStateHandler) RecordDoneSyncAsync(state state.FlowState) {
	sigchan := sr.sigMap[state.FlowInstanceId]
	defer delete(sr.sigMap, state.FlowInstanceId) // remove the flowstate signal entry from map as its objective done
	uri := sr.host + "/v1/instances/end"
	sr.logger.Debugf("POST record start: %s\n", uri)
	jsonReq, _ := json.Marshal(state)
	sr.logger.Debug("JSON: ", string(jsonReq))
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(jsonReq))
	if err != nil {
		sr.logger.Warnf("unable to record flowstate done: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	sr.AddHeaders(req)

	// req.Header.Set(ASYNC_CALLING_HEADER, strconv.FormatBool(sr.asyncAPICall))
	if sigchan != nil {
		<-sigchan // blocking to complete its counter part start call
	}
	err1 := sr.ConnectionRetry(req)
	if err != nil {
		sr.logger.Warnf("unable to record flowstate done : %v", err1)
		return
	}
}

func (sr *FlowStateHandler) FlowInstanceSnapshotById(flowId string) (string, error) {

	uri := sr.host + "/v1/instances/" + flowId + "/details"

	sr.logger.Debugf("Get flow [%s] details: %s", flowId)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("username", support.GetUserName())

	sr.AddHeaders(req)

	resp, err := sr.ConnectionRetryWithResponse(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	sr.logger.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		//todo return error
	}

	info := &state.FlowInfo{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, info)
	if err != nil {
		return "", err
	}
	return info.FlowURI, nil
}

func (sr *FlowStateHandler) TaskSnapshotByStepId(flowId string, stepId string) (*state.Snapshot, error) {

	uri := sr.host + "/v1/instances/" + flowId + "/snapshot/" + stepId

	sr.logger.Debugf("Get flow [%s] snapshot data by step id [%s]", flowId, stepId)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("username", support.GetUserName())

	sr.AddHeaders(req)

	resp, err := sr.ConnectionRetryWithResponse(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sr.logger.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		//todo return error
	}

	snapshot := &state.Snapshot{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, snapshot)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (sr *FlowStateHandler) DeleteStepRecordsFromStepId(flowId string, stepId string) (int, error) {
	uri := sr.host + "/v1/instances/" + flowId + "/step/" + stepId

	sr.logger.Debugf("Delete step records for flow [%s] from step id [%s]", flowId, stepId)

	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return 0, err
	}
	// req.Header.Set("Content-Type", "application/json")
	req.Header.Set("username", support.GetUserName())

	sr.AddHeaders(req)

	resp, err := sr.ConnectionRetryWithResponse(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	sr.logger.Info("response Status:", resp.Status)
	return resp.StatusCode, nil
}

func handleOption(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Headers", "Origin")
	w.Header().Add("Access-Control-Allow-Headers", "X-Requested-With")
	w.Header().Add("Access-Control-Allow-Headers", "Accept")
	w.Header().Add("Access-Control-Allow-Headers", "Accept-Language")
	w.Header().Set("Content-Type", "application/json")
}

// RestartFlow restarts a Flow Instance (POST "/flow/restart").
//
// To post a restart flow, try this at a shell:
// $ curl -H "Content-Type: application/json" -X POST -d '{...}' http://localhost:8080/flowstate/restart
func (sr *FlowStateHandler) RestartFlowFromBeginning(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		handleOption(w)
		w.WriteHeader(http.StatusOK)
	case http.MethodPost:

		w.Header().Add("Access-Control-Allow-Origin", "*")

		req := &RestartFlowRequest{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		flowState, err := sr.GetFlowDetails(req.FlowInstanceId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.FlowURI = flowState.FlowURI
		req.Inputs = flowState.FlowInputs

		results, err := sr.requestProcessor.RestartFlow(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		idAttr, ok := results["id"]
		if ok {
			idResponse := &instance.IDResponse{ID: idAttr.(string)}
			sr.logger.Debugf("Restarted Instance [ID:%s] for %s", idResponse.ID, req.FlowURI)

			encoder := json.NewEncoder(w)
			err := encoder.Encode(idResponse)
			if err != nil {
				sr.logger.Errorf("Unable to encode response: %v", err)
			}
		} else {
			sr.logger.Debugf("Restarted Instance %s", req.FlowURI)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			encoder := json.NewEncoder(w)
			err := encoder.Encode(results)
			if err != nil {
				sr.logger.Errorf("Unable to encode response: %v", err)
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (sr *FlowStateHandler) RestartFlowFromActivity(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodOptions:
		handleOption(w)
		w.WriteHeader(http.StatusOK)
	case http.MethodPost:
		w.Header().Add("Access-Control-Allow-Origin", "*")
		req := &RestartActivityRequest{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reStartRequest, err := sr.getRestartRequestForStepId(req.FlowInstanceId, req.TaskStepId, req.TaskName, req.Inputs)
		if err != nil {
			http.Error(w, fmt.Sprintf("Faield getting date from state server: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// clean old records
		id, err := sr.DeleteStepRecordsFromStepId(req.FlowInstanceId, req.TaskStepId)
		if err != nil {
			sr.logger.Debugf("Failed to delete step records for step ID [%d] due to error - %s.", id, err.Error())
		}
		stepId, _ := coerce.ToInt(req.TaskStepId)
		results, err := sr.requestProcessor.RestartActivity(reStartRequest, req.FlowInstanceId, stepId)
		if err != nil {
			sr.logger.Errorf("Failed to restart activity due to error - %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sr.logger.Debugf("Restarted Instance %s", req.FlowInstanceId)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		encoder := json.NewEncoder(w)
		err = encoder.Encode(results)
		if err != nil {
			sr.logger.Errorf("Unable to encode response: %v", err)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (sr *FlowStateHandler) getRestartRequestForStepId(flowId, stepId, taskName string, inputs map[string]interface{}) (*RestartRequest, error) {

	snapshot0Data, err := sr.TaskSnapshotByStepId(flowId, "0")
	if err != nil {
		return nil, err
	}
	stepIDInt, _ := strconv.Atoi(stepId)
	if len(snapshot0Data.Tasks) > 0 {
		if stepIDInt >= 1 {
			stepIDInt = stepIDInt - 1
		}
	}
	snapshot, err := sr.TaskSnapshotByStepId(flowId, strconv.Itoa(stepIDInt))
	if err != nil {
		return nil, err
	}

	snapshotData := &state.Snapshot{
		SnapshotBase: &state.SnapshotBase{
			FlowURI: snapshot.FlowURI,
			Status:  snapshot.Status,
			Attrs:   snapshot.Attrs,
			Tasks:   []*state.Task{{Id: taskName, Status: 20}},
			Links:   nil,
		},
		Id:        flowId,
		WorkQueue: []*state.WorkItem{{ID: 1, SubflowId: 0, TaskId: taskName}},
		Subflows:  snapshot.Subflows,
	}

	snapshotData.Status = 100
	restartReqStruct := struct {
		InitialState interface{}            `json:"initialState"`
		Data         map[string]interface{} `json:"data"`
		Interceptor  *support.Interceptor   `json:"interceptor"`
		Patch        *support.Patch         `json:"patch"`
		ReturnID     bool                   `json:"returnId"`
	}{
		InitialState: snapshotData,
		Data:         snapshotData.Attrs,
		Patch:        nil,
		Interceptor: &support.Interceptor{TaskInterceptors: []*support.TaskInterceptor{&support.TaskInterceptor{ID: taskName,
			Skip:    false,
			Inputs:  inputs,
			Outputs: nil}}},
		ReturnID: true,
	}

	restartReq := &RestartRequest{}
	v, _ := json.Marshal(restartReqStruct)
	_ = json.Unmarshal(v, restartReq)
	return restartReq, nil
}

func (sr *FlowStateHandler) GetFlowURI(flowId string) (string, error) {

	uri := sr.host + "/v1/instances/" + flowId + "/details"

	sr.logger.Debugf("Get flow [%s] details: %s", flowId)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	sr.AddHeaders(req)

	req.Header.Set("username", support.GetUserName())
	// client := &http.Client{}
	resp, err := sr.ConnectionRetryWithResponse(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	sr.logger.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		//todo return error
	}

	info := &state.FlowInfo{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, info)
	if err != nil {
		return "", err
	}
	return info.FlowURI, nil
}

func (sr *FlowStateHandler) GetFlowDetails(flowId string) (*state.FlowInfo, error) {

	uri := sr.host + "/v1/instances/" + flowId + "/details"

	sr.logger.Debugf("Get flow [%s] details: %s", flowId)

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	sr.AddHeaders(req)

	req.Header.Set("username", support.GetUserName())
	// client := &http.Client{}
	resp, err := sr.ConnectionRetryWithResponse(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sr.logger.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		//todo return error
	}

	info := &state.FlowInfo{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
