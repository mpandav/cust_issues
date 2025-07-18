package unittest

import (
	"encoding/json"
	"fmt"
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/log"
	feengine "github.com/tibco/wi-contrib/engine"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (config *UnitTestConfig) process() error {
	_ = os.Setenv("TEST_MODE", "true")

	testRunner := &TestRunner{}

	testParser := &TestParser{}
	testParser.suiteDataJson = config.suiteDataJson
	testParser.userSuiteSet = config.suiteSet
	testParser.runAllTests = config.runAllTests
	testParser.runTestSuites = config.runTestSuites
	testParser.runTestCases = config.runTestCases
	testParser.runTestFlows = config.runTestFlows
	testParser.specialTestCaseSuite = config.specialTestCaseSuite
	testParser.specialTestFlowSuite = config.specialTestFlowSuite

	if config.collectIO || os.Getenv(FLOGO_UT_PRESERVE_IO) == "true" {
		testParser.collectIO = true
		testRunner.collectIO = true
	}
	if config.collectCoverage {
		testParser.collectCoverage = true
		testRunner.collectCoverage = true
	}

	err := testParser.loadTestSuites()
	if err != nil {
		return err
	}
	err = testParser.verifyTestSuites()
	if err != nil {
		return err
	}
	testRunner.testSuites = testParser.testSuites

	flogoEngine, err := config.createFlogoEngine()
	if err != nil {
		config.writeEngineError(err.Error())
		return err
	}

	config.compareEngineAndTestsVersion(testParser.testInfo, flogoEngine)

	testRunner.engine = &flogoEngine

	factory := action.GetFactory(Flow)
	err = factory.Initialize(flogoEngine.App())
	if err != nil {
		return err
	}
	testRunner.factory = factory

	testRunner.runTests()
	config.writeResult(flogoEngine, testRunner.report)
	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
	println("     TIBCO Flogo® Runtime Test Execution Completed")
	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

	return nil
}

func (config *UnitTestConfig) compareEngineAndTestsVersion(testInfo *TestInfo, engine engine.Engine) {
	if engine.App().Name() != testInfo.Name {
		fmt.Printf(showInfo("Warning! App Name [%s] in application json and App name [%s] in test suite  do not match. Test run may not be successful for the changes made in the current version\n"), engine.App().Name(), testInfo.Name)
		return
	}
	if engine.App().Version() != testInfo.Version {
		fmt.Printf(showInfo("Warning! App version [%s] in application json and App version [%s] in test suite  do not match. Test run may not be successful for the changes made in the current version\n"), engine.App().Version(), testInfo.Version)

	}
}

func (config *UnitTestConfig) writeEngineError(message string) {
	if !isTCIEnv() {
		return
	}
	jsonStr := "{\"error\":\"" + message + "\"}"
	fileName := "output.json"
	os.Remove(fileName)
	_ = ioutil.WriteFile(fileName, []byte(jsonStr), 777)
}

func (config *UnitTestConfig) writeResult(flogoEngine engine.Engine, report *Report) {

	var op []byte
	var err error
	if isTCIEnv() {
		op, err = json.MarshalIndent(report, "", "    ")
	} else {
		completeReport := getCompleteReport(report)
		op, err = json.MarshalIndent(completeReport, "", "    ")
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	fileName := "output.json"
	if !isTCIEnv() {
		fileName = config.getTestResultFilePath(flogoEngine.App().Name().(string))
	}
	fmt.Printf(showInfo("Printing result to file %s \n"), fileName)
	os.Remove(fileName)
	err = ioutil.WriteFile(fileName, op, 777)

}

func (config *UnitTestConfig) getTestResultFilePath(appName string) string {
	path := ""
	if config.output != "" || config.name != "" {
		if config.output != "" {
			path = strings.TrimRight(config.output, "/") + "/"
		}
		if config.name != "" {
			path = path + config.name + ".testresult"
		} else {
			path = path + appName + ".testresult"
		}
	} else {
		path = appName + ".testresult"
	}

	return path
}

func getCompleteReport(report *Report) *CompleteReport {
	completeReport := CompleteReport{Report: report}
	result := &ReportResult{}
	for _, key := range report.SuiteReport {
		result.TotalSuites++
		if key.SuiteResult.FailedTests > 0 {
			result.FailedSuites++
		}
		completeReport.Result = result
	}

	if result.FailedSuites > 0 {
		fmt.Printf(showError("Suite Result\nSuites Run: %d\tSuccess: %d\tFailure: %d\n"), result.TotalSuites, result.TotalSuites-result.FailedSuites, result.FailedSuites)
	} else {
		fmt.Printf(showSuccess("Suite Result\nSuites Run: %d\tSuccess: %d\tFailure: %d\n"), result.TotalSuites, result.TotalSuites-result.FailedSuites, result.FailedSuites)
	}
	return &completeReport

}

var engineConfig = `{
  "type": "flogo:engine",
  "actionSettings": {
    "github.com/project-flogo/flow": {
      "stateRecordingMode": "debugger"
    }
  },
  "services": [
    {
      "name": "stateful",
      "ref": "github.com/tibco/wi-contrib/engine/unittest",
      "enabled": true,
      "settings": {
      }
    }
  ]
}`

func (config *UnitTestConfig) createFlogoEngine() (engine.Engine, error) {

	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
	println("     Starting TIBCO Flogo® Runtime in Test Mode")
	println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

	flogoApp, err := engine.LoadAppConfig(config.flogoJson, false)

	if err != nil {
		return nil, fmt.Errorf("Failed to create engine with error : %s", err.Error())
	}
	function.ResolveAliases()
	flogoApp.Triggers = nil

	e, err := engine.New(flogoApp, engine.ConfigOption("", false))
	if err != nil {
		return nil, fmt.Errorf("Failed to create engine with error: %s", err.Error())
	}

	err = feengine.StartEngine(e)
	if err != nil {
		log.RootLogger().Error(err)
		os.Exit(1)
	}

	return e, nil
}

func (config *UnitTestConfig) loadTestSuites(testSuites string) {

	if testSuites == "" {
		config.runAllTests = true
		return
	}

	var userSuites []string
	userSuites = strings.Split(testSuites, ",")
	config.testSuites = userSuites

	suiteSet := make(map[string]struct{})
	var Empty struct{}
	for _, key := range userSuites {
		suiteSet[key] = Empty
	}
	config.suiteSet = suiteSet
	config.runTestSuites = true

}

func (config *UnitTestConfig) loadTestCases(testCases string) {

	if testCases == "" {
		return
	}
	config.runAllTests = false
	var userTestCases []string
	userTestCases = strings.Split(testCases, ",")

	var userTestSuites []Suite = make([]Suite, 0)
	testCasesSuite := Suite{
		ID:       "Selected-Test-Cases",
		Name:     "Selected Test Cases",
		Disabled: false,
		Tests:    make([]string, 0),
		Type:     TestCase,
	}
	tests := config.suiteDataJson["tests"].(map[string]interface{})
	var userTestsMap = make(map[string][]string)
	for _, testKey := range userTestCases {
		if strings.Contains(testKey, "/") {
			s := strings.Split(testKey, "/")
			usuite := s[0]
			ucase := s[1]
			_, ok := userTestsMap[usuite]
			if ok {
				userTestsMap[usuite] = append(userTestsMap[usuite], ucase)
			} else {
				userTestsMap[usuite] = []string{ucase}
			}
		} else {
			_, ok := userTestsMap["_orphan"]
			if ok {
				userTestsMap["_orphan"] = append(userTestsMap["_orphan"], testKey)
			} else {
				userTestsMap["_orphan"] = []string{testKey}
			}
		}
	}

	if len(userTestsMap["_orphan"]) > 0 {
		for _, testKey := range userTestsMap["_orphan"] {
			for key, _ := range tests {
				if key == testKey {
					testCasesSuite.Tests = append(testCasesSuite.Tests, key)
				}
			}
		}
		userTestSuites = append(userTestSuites, testCasesSuite)
	}

	for key, _ := range userTestsMap {
		if key == "_orphan" {
			continue
		}

		testCasesSuite = Suite{
			ID:       key,
			Name:     key,
			Disabled: false,
			Tests:    make([]string, 0),
			Type:     TestCase,
		}
		for _, testKey := range userTestsMap[key] {
			for key, _ := range tests {
				if key == testKey {
					testCasesSuite.Tests = append(testCasesSuite.Tests, key)
				}
			}
		}
		userTestSuites = append(userTestSuites, testCasesSuite)
	}

	config.specialTestCaseSuite = userTestSuites
	config.runTestCases = true
}

func (config *UnitTestConfig) loadTestFlows(testFlows string) {

	if testFlows == "" {
		return
	}
	config.runAllTests = false

	var userFlows []string
	userFlows = strings.Split(testFlows, ",")

	testFlowSuite := Suite{
		ID:       "Selected-Flows",
		Name:     "Selected Flows",
		Disabled: false,
		Tests:    make([]string, 0),
		Type:     TestFlow,
	}
	tests := config.suiteDataJson["tests"].(map[string]interface{})

	for _, flowKey := range userFlows {
		for key, s := range tests {
			testsJson := s.(map[string]interface{})
			flowName := testsJson["flowName"].(string)
			if flowName == flowKey {
				testFlowSuite.Tests = append(testFlowSuite.Tests, key)
			}
		}
	}

	config.specialTestFlowSuite = testFlowSuite
	config.runTestFlows = true
}

func (config *UnitTestConfig) loadTestFile(testFile string) error {
	if isTCIEnv() {
		err := config.loadTestFileForTCI()
		if err != nil {
			return err
		}
	} else {
		err := config.loadTestFileFromPath(testFile)
		if err != nil {
			return err
		}
	}

	var suiteDataJson map[string]interface{}
	if err := json.Unmarshal([]byte(config.testJson), &suiteDataJson); err != nil {
		return fmt.Errorf("failed to parse test file with error: %s", err.Error())
	}
	config.suiteDataJson = suiteDataJson
	return nil
}

func (config *UnitTestConfig) loadTestFileForTCI() error {
	content, err := ioutil.ReadFile("ut.flogotest")
	if err != nil {
		return fmt.Errorf("Failed to read unit testfile [%s]", "ut.flogotest")
	}
	config.testJson = string(content)
	return nil
}

func (config *UnitTestConfig) loadTestFileFromPath(testFile string) error {

	if filepath.Ext(testFile) != ".flogotest" {
		return fmt.Errorf("The test file should of [.flogotest] extension. The provided file has different extension [%s]", filepath.Ext(testFile))
	}
	content, err := ioutil.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("Failed to read test file  from path [%s] with error: %s", testFile, err.Error())
	}
	config.testJson = string(content)
	return nil
}

func (config *UnitTestConfig) loadFlogoJSON(appJsonEmbedded string, appJSONPath string) error {
	if isTCIEnv() {
		err := config.loadFlogoJSONforTCI()
		if err != nil {
			return err
		}
	} else {
		err := config.loadFlogoJSONforOnPrem(appJsonEmbedded, appJSONPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (config *UnitTestConfig) loadFlogoJSONforTCI() error {
	content, err := ioutil.ReadFile("flogo.json")
	if err != nil {
		return fmt.Errorf("Read application json file [%s]", "flogo.json")
	}
	config.flogoJson = string(content)
	return nil
}

func (config *UnitTestConfig) loadFlogoJSONforOnPrem(appJsonEmbedded string, appJSONPath string) error {

	if appJSONPath != "" {
		err := config.loadFlogoJSONFromPath(appJSONPath)
		if err != nil {
			return err
		}
		return nil
	}
	if appJsonEmbedded == "" {
		return fmt.Errorf("Failed to read application json embedded in the app.")
	}
	config.flogoJson = appJsonEmbedded
	return nil
}

func (config *UnitTestConfig) loadFlogoJSONFromPath(appJSONPath string) error {
	content, err := ioutil.ReadFile(appJSONPath)
	if err != nil {
		return fmt.Errorf("Failed to read application json file from path [%s] with error: %s", appJSONPath, err.Error())
	}
	config.flogoJson = string(content)
	return nil
}
