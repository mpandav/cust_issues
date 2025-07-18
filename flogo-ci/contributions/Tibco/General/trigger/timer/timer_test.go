package timer

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEval(t *testing.T) {
}

func TestRegistered(t *testing.T) {
	time := &TimerFactory{trigger.NewMetadata(jsonMetadata)}
	config := &trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	timer := time.New(config)
	if timer == nil {
		t.Error("Timer Trigger Not Registered")
		t.Fail()
		return
	}
}

func TestInit(t *testing.T) {
	time := &TimerFactory{trigger.NewMetadata(jsonMetadata)}
	config := &trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	timer := time.New(config)

	_, isNew := timer.(trigger.Initializable)

	if !isNew {
		runner := &TestRunner{}
		tgr, isOld := timer.(trigger.InitOld)
		if isOld {
			tgr.Init(runner)

		}
	}
}

func TestTimer(t *testing.T) {

	log.Debugf("TestTimer")
	timeFactory := &TimerFactory{trigger.NewMetadata(jsonMetadata)}
	config := &trigger.Config{}
	json.Unmarshal([]byte(testConfig), config)
	timer := timeFactory.New(config)

	timer.Start()
	<-time.After(time.Millisecond * 2000)
	defer timer.Stop()

	log.Debug("Test timer done")
}
