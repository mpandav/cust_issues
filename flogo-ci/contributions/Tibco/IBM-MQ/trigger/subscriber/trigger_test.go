package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/test"
	"github.com/project-flogo/core/trigger"
)

var jsonMetadata = getJsonMetadata()

func getJsonMetadata() string {
	jsonMetadataBytes, err := ioutil.ReadFile("trigger.json")
	if err != nil {
		panic("No Json Metadata found for trigger.json path")
	}
	return string(jsonMetadataBytes)
}
func getTriggerConfig() string {
	jsonMetadataBytes, err := ioutil.ReadFile("trigger_test.json")
	if err != nil {
		panic("No Json Metadata found in trigger_test.json")
	}
	return string(jsonMetadataBytes)
}

// func getTriggerConfigMulti() string {
// 	jsonMetadataBytes, err := ioutil.ReadFile("trigger_multi_test.json")
// 	if err != nil {
// 		panic("No Json Metadata found for trigger.json path")
// 	}
// 	return string(jsonMetadataBytes)
// }

type TestRunner struct {
}

// Run implements action.Runner.Run
func (tr *TestRunner) Run(context context.Context, action action.Action, uri string, options interface{}) (code int, data interface{}, err error) {
	return 0, nil, nil
}

func (tr *TestRunner) RunAction(ctx context.Context, act action.Action, options map[string]interface{}) (results map[string]*data.Attribute, err error) {
	return nil, nil
}

func (tr *TestRunner) Execute(ctx context.Context, act action.Action, inputs map[string]*data.Attribute) (results map[string]*data.Attribute, err error) {
	return nil, nil
}

const testConfig string = `{
                "id": "mytrigger",
                    "settings": {
                    "setting": "somevalue"
                },
                "handlers": [
                    {
                        "actionId": "test_action",
                        "settings": {
                            "handler_setting": "somevalue"
                        }
                    }
                ]
            }`

func TestInit(t *testing.T) {
	md := triggerMd
	config := &trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	f := md.New(config)

	_, isNew := f.(trigger.Initializable)

	if !isNew {
		runner := &TestRunner{}
		tgr, isOld := f.(trigger.InitOld)
		if isOld {
			tgr.Init(runner)

		}
	}
}

func TestStartListenPlain(t *testing.T) {
	log.SetLogLevel(log.DebugLevel)

	config := &trigger.Config{}
	triggerconfig := getTriggerConfig()
	err := json.Unmarshal([]byte(triggerconfig), config)
	if err != nil {
		fmt.Print(err)
		t.Errorf("Deserialization of config failed %s", err.Error())
		t.Fail()
		panic(err)
	}
	// New  factory
	f := &triggerMd
	count := 0
	actions := map[string]action.Action{"dummy": test.NewDummyAction(func() {
		fmt.Println("Dummy action received message")
		count++
	})}
	tgr, err := test.InitTrigger(f, config, actions)
	if err != nil {
		t.Errorf("Trigger initialization failed %s", err.Error())
		panic("Trigger initialization failed")

	}
	tgr.Start()
	// go send it some messages.
	time.Sleep(300000 * time.Millisecond)
	tgr.Stop()
}

// func TestStartListenMulti(t *testing.T) {
// 	log.SetLogLevel(logger.DebugLevel)

// 	config := &trigger.Config{}
// 	err := json.Unmarshal([]byte(getTriggerConfigMulti()), config)
// 	if err != nil {
// 		t.Errorf("Deserialization of config failed %s", err.Error())
// 		t.Fail()
// 	}
// 	// New  factory
// 	f := &MqTriggerFactory{trigger.NewMetadata(jsonMetadata)}
// 	count := 0
// 	actions := map[string]action.Action{"dummy": test.NewDummyAction(func() {
// 		fmt.Println("Dummy action received message")
// 		count++
// 	})}
// 	tgr, err := test.InitTrigger(f, config, actions)
// 	if err != nil {
// 		t.Errorf("Trigger initialization failed %s", err.Error())
// 		panic("Trigger initialization failed")

// 	}
// 	tgr.Start()
// 	// go send it some messages.
// 	time.Sleep(300000 * time.Millisecond)
// 	tgr.Stop()
// }
