package unittest

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/project-flogo/flow/support"
)

func (parser *TestParser) loadTestSuites() error {

	parser.testSuites = &TestSuites{}
	testInfo := &TestInfo{}
	tests, ok := parser.suiteDataJson["app"].(map[string]interface{})
	err := mapstructure.Decode(tests, testInfo)
	if err != nil {
		return fmt.Errorf("failed to find app info in the tests file")
	}
	parser.testInfo = testInfo
	parser.testInfo.ModelVersion = getUnitTestModel(parser.suiteDataJson["model"])

	suitesStr := "suites"
	if parser.testInfo.ModelVersion == V1 || parser.testInfo.ModelVersion == V2Internal {
		suitesStr = "suits"
	}
	suites, ok := parser.suiteDataJson[suitesStr].(map[string]interface{})
	defaultSuiteReq := false
	defaultSuite := Suite{}
	if (!ok || len(suites) == 0) && !parser.runTestCases && !parser.runTestFlows {
		defaultSuite = Suite{
			ID:       "Default",
			Name:     "Default",
			Disabled: false,
			Tests:    make([]string, 0),
			Type:     TestFlow,
		}
		parser.createDefaultSuite(&defaultSuite)
		defaultSuiteReq = true

	}
	parser.testSuites.TestSuiteData = []TestSuiteData{}

	if parser.runTestCases {
		for _, s := range parser.specialTestCaseSuite {
			parser.loadSingleSuite(&s)
		}
	}
	if parser.runTestFlows {
		parser.loadSingleSuite(&parser.specialTestFlowSuite)
	}
	if defaultSuiteReq {
		parser.loadSingleSuite(&defaultSuite)
	}

	for _, s := range suites {

		singleSuiteJson := s.(map[string]interface{})
		suit := &Suite{}
		suit.Type = TestSuite
		err := mapstructure.Decode(singleSuiteJson, suit)

		if err != nil {
			return fmt.Errorf("failed to load suite from the test file")
		}

		if !parser.runAllTests {
			_, ok := parser.userSuiteSet[suit.Name]
			if !ok {
				continue
			}
		}

		//if suit.Disabled {
		//	fmt.Printf(showInfo("Test Suite %s is disabled and its execution will be skipped"), suit.Name)
		//	continue
		//}
		err = parser.loadSingleSuite(suit)
		if err != nil {
			return err
		}
	}

	fmt.Printf(showInfo("Loaded  %d Test Suite(s) from Test file \n"), len(parser.testSuites.TestSuiteData))

	return nil
}

func (parser *TestParser) createDefaultSuite(suit *Suite) error {

	testsuite := &TestSuiteData{}
	testsuite.Name = suit.Name
	tests := parser.suiteDataJson["tests"].(map[string]interface{})

	testsuite.TestCases = []TestCaseData{}

	for key, _ := range tests {
		suit.Tests = append(suit.Tests, key)
	}

	return nil
}

func (parser *TestParser) loadSingleSuite(suit *Suite) error {

	testsuite := &TestSuiteData{}
	testsuite.Name = suit.Name
	tests := parser.suiteDataJson["tests"].(map[string]interface{})

	testsuite.TestCases = []TestCaseData{}

	// Iterate over the suit list in the suites element array.
	for _, testKey := range suit.Tests {

		// For each tests name in the suites element array search for it in the tests element
		for key, s := range tests {
			// If test with the name is found then load the TestCaseData Data.
			if key == testKey {
				testsJson := s.(map[string]interface{})
				err := parser.loadTestData(testsuite, testsJson)
				if err != nil {
					return err
				}
			}
		}
	}

	parser.testSuites.TestSuiteData = append(parser.testSuites.TestSuiteData, *testsuite)

	return nil
}

func (parser *TestParser) verifyTestSuites() error {
	if parser.runAllTests {
		fmt.Println("Running Tests for all test suites in the test file")
		return nil
	}

	if !parser.runTestSuites {
		return nil
	}

	suiteSet := make(map[string]struct{})
	for _, suite := range parser.testSuites.TestSuiteData {
		var Empty struct{}
		suiteSet[suite.Name] = Empty
	}
	var inValidSuites []string
	for key := range parser.userSuiteSet {
		_, ok := suiteSet[key]
		if !ok {
			inValidSuites = append(inValidSuites, key)
		}
	}
	if len(inValidSuites) > 0 {
		return fmt.Errorf("Test Suites %s provided as arguments not part of test file", inValidSuites)
	}

	return nil
}

// TestCaseData Data contains the Flow, Flowinput, FlowOutout and activities
func (parser *TestParser) loadTestData(testsuite *TestSuiteData, testsJson map[string]interface{}) error {
	test := TestCaseData{}
	test.Activities = []Activity{}
	for key := range testsJson {
		switch key {
		case "name":
			test.Name = testsJson["name"].(string)
		case "flowId":
			test.FlowName = testsJson["flowName"].(string)
			// Here the flow:Name is coming from the Studio
			test.Flow = "res://" + testsJson["flowId"].(string)
		case "flowInputs":
			input, ok := testsJson["flowInputs"].(map[string]interface{})
			if !ok {
				test.Valid = false
				continue
			} else {
				test.FlowInput = input
			}
		case "flowOutputs":
			activity := Activity{}
			activity.Name = "_flowOutput"
			activity.ID = "_flowOutput"
			assertionData := testsJson["flowOutputs"].(map[string]interface{})
			executionType := parser.findFlowOPType(assertionData)
			switch executionType {
			case -1:
				assertionArray, _ := parser.loadAssertions(assertionData)
				activity.Assertion = assertionArray
				activity.Type = support.AssertionActivity
			case support.AssertionActivity, support.AssertionException:
				assertionsJson := assertionData["assertions"]
				if assertionsJson != nil {
					assertionData := assertionsJson.(map[string]interface{})
					assertionArray, _ := parser.loadAssertions(assertionData)
					activity.Assertion = assertionArray
					activity.Type = executionType
				}
			}
			test.Activities = append(test.Activities, activity)

		case "activities":
			actsArray := testsJson["activities"].(map[string]interface{})
			for actKey := range actsArray {
				activity := Activity{}
				activity.Name = actKey
				activity.ID = actKey
				activityJson := actsArray[actKey].(map[string]interface{})

				executionType := parser.findExecutionType(activityJson)

				activity.Type = executionType
				switch executionType {
				case support.AssertionActivity, support.AssertionException:
					assertionsJson := activityJson["assertions"]
					if assertionsJson != nil {
						assertionData := assertionsJson.(map[string]interface{})
						assertionArray, _ := parser.loadAssertions(assertionData)
						activity.Assertion = assertionArray
					}
				case support.MockActivity, support.MockException:
					activity.SkipExecution = true
					mockJson := activityJson["mock"]

					if mockJson != nil {
						mockData, ok := mockJson.(map[string]interface{})
						if ok {
							activity.Mock = mockData
						} else {
							mockDataArray, ok := mockJson.([]interface{})
							if ok {
								if len(mockDataArray) > 0 {
									mockData, ok := mockDataArray[0].(map[string]interface{})
									if ok {
										activity.Mock = mockData
									}
								}
							}
						}
					}

				case support.SkipActivity:
					activity.SkipExecution = true
				}

				test.Activities = append(test.Activities, activity)
			}
		}
	}
	test.Valid = true
	testsuite.TestCases = append(testsuite.TestCases, test)
	return nil
}

func (parser *TestParser) findFlowOPType(assertionJson map[string]interface{}) int {
	executionType := -1
	if parser.testInfo.ModelVersion == V3 {
		executionTypeStr := assertionJson["type"].(string)
		switch executionTypeStr {
		case "ASSERT_ON_OP":
			executionType = support.AssertionActivity
		case "ASSERT_ON_ERR":
			executionType = support.AssertionException
		}
	}
	return executionType
}

func (parser *TestParser) findExecutionType(activityJson map[string]interface{}) int {
	executionType := -1
	if parser.testInfo.ModelVersion == V2 || parser.testInfo.ModelVersion == V3 {
		executionTypeStr := activityJson["type"].(string)
		switch executionTypeStr {
		case "ASSERT_ON_OP":
			executionType = support.AssertionActivity
		case "ASSERT_ON_ERR":
			executionType = support.AssertionException
		case "MOCK_ON_OP":
			executionType = support.MockActivity
		case "MOCK_ON_ERR":
			executionType = support.MockException
		case "SKIP_ACTIVITY":
			executionType = support.SkipActivity
		}
	} else {
		if activityJson["isMock"] != nil {
			isMock, ok := activityJson["isMock"].(bool)
			if ok && isMock {
				executionType = support.MockActivity
			}
		}

		if activityJson["isSkipped"] != nil {
			isSkipped, ok := activityJson["isSkipped"].(bool)
			if ok && isSkipped {
				executionType = support.SkipActivity
			}
		}

		if executionType != support.MockActivity && executionType != support.SkipActivity {
			assertionsJson := activityJson["assertions"]
			if assertionsJson != nil {
				assertionData := assertionsJson.(map[string]interface{})
				if len(assertionData) > 0 {
					executionType = support.AssertionActivity
				}
			}
		}
	}

	return executionType
}

func (parser *TestParser) loadAssertions(assertionData map[string]interface{}) ([]AssertionUT, error) {
	assertionArray := []AssertionUT{}
	for assertId, assertionSingleRaw := range assertionData {
		assertion := AssertionUT{}
		assertion.ID = assertId
		assertionSingle := assertionSingleRaw.(map[string]interface{})
		assertion.Name = assertionSingle["name"].(string)
		if assertionSingle["valueAssertion"] != nil {
			assertion.Expression = assertionSingle["valueAssertion"]
			assertion.Type = 1
		} else if assertionSingle["activityAssertion"] != nil {
			assertion.Expression = assertionSingle["activityAssertion"]
			assertion.Type = 2
		}
		assertionArray = append(assertionArray, assertion)

	}
	return assertionArray, nil
}
