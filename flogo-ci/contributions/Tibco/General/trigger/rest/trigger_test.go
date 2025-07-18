package rest

import (
	"context"
	"io/ioutil"
)

var jsonMetadata = getJsonMetadata()

func getJsonMetadata() string {
	jsonMetadataBytes, err := ioutil.ReadFile("trigger.json")
	if err != nil {
		panic("No Json Metadata found for trigger.json path")
	}
	return string(jsonMetadataBytes)
}

type TestRunner struct {
}

// Run implements action.Runner.Run
func (tr *TestRunner) Run(context context.Context, action action.Action, uri string, options interface{}) (code int, data interface{}, err error) {
	log.Debugf("Ran Action: %v", uri)
	return 200, nil, nil
}

func (tr *TestRunner) RunAction(ctx context.Context, act action.Action, options map[string]interface{}) (results map[string]*data.Attribute, err error) {
	return nil, nil
}

func (tr *TestRunner) Execute(ctx context.Context, act action.Action, inputs map[string]*data.Attribute) (results map[string]*data.Attribute, err error) {
	return nil, nil
}

/*func TestInitOk(t *testing.T) {

	config := &trigger.Config{}
	err := json.Unmarshal([]byte(testConfig), config)
	if err != nil {
		t.Fail()
	}

	var tConfigs []*trigger.Config
	tConfigs = append(tConfigs, config)

	t.Log(tConfigs[0].Id)

	runner := &TestRunner{}

	_, err = app.CreateTriggers(tConfigs, runner)

	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

//
func TestHandlerOk(t *testing.T) {

	config := &trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)

	action.RegisterFactory("", )

	var tConfigs []*trigger.Config
	tConfigs = append(tConfigs, config)

	runner := &TestRunner{}

	triggers, err := app.CreateTriggers(tConfigs, runner)

	if err != nil {
		t.Fail()
	}

	tgr := triggers["tibco-wi-rest"]
	if tgr == nil {
		t.Fail()
	}

	tgr.Start()
	defer tgr.Stop()

	uri := "http://127.0.0.1:8091/device/12345/reset"

	req, err := http.NewRequest("POST", uri, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Debug("response Status:", resp.Status)

	if resp.StatusCode >= 300 {
		t.Fail()
	}
}

func TestValidationNegative(t *testing.T) {
	// New  factory

	config := &trigger.Config{}
	json.Unmarshal([]byte(testConfigForValidation), config)

	tConfigs := make([]*trigger.Config, 1)
	tConfigs = append(tConfigs, config)

	runner := &TestRunner{}

	triggers, err := app.CreateTriggers(tConfigs, runner)

	if err != nil {
		t.Fail()
	}

	tgr := triggers["tibco-wi-rest"]
	if tgr == nil {
		t.Fail()
	}

	tgr.Start()
	defer tgr.Stop()

	uri := "http://127.0.0.1:8092/device/validation?age=11"

	body := `{"age": "xx"}`
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(body)))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Debug("response Status:", resp.Status)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fail()
	}
}*/
