package unittest

import (
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/engine"
	"time"
)

type UnitTestConfig struct {
	flogoJson            string
	testJson             string
	testSuites           []string
	suiteSet             map[string]struct{}
	runAllTests          bool
	suiteDataJson        map[string]interface{}
	output               string
	name                 string
	specialTestCaseSuite []Suite
	specialTestFlowSuite Suite
	runTestSuites        bool
	runTestCases         bool
	runTestFlows         bool
	collectIO            bool
	collectCoverage      bool
}

type TestRunner struct {
	engine          *engine.Engine
	testSuites      *TestSuites
	opProgress      bool
	ticker          *time.Ticker
	tickerChan      chan bool
	tickerTime      time.Time
	testProgress    *TestProgress
	report          *Report
	factory         action.Factory
	collectIO       bool
	playMode        bool
	collectCoverage bool
}

type TestParser struct {
	testSuites           *TestSuites
	runAllTests          bool
	suiteDataJson        map[string]interface{}
	userSuiteSet         map[string]struct{}
	testInfo             *TestInfo
	specialTestCaseSuite []Suite
	specialTestFlowSuite Suite
	runTestSuites        bool
	runTestCases         bool
	runTestFlows         bool
	collectIO            bool
	collectCoverage      bool
}

type SuiteReportProcessor struct {
	report *SuiteReport
}

type TestFile struct {
	Variables Variables      `json:"variables"`
	Suits     []Suite        `json:"suits"`
	Tests     []TestCaseData `json:"tests"`
}
type Variables struct {
	Variable []string `json:"variable"`
}

type TestProgress struct {
	TestStats *TestStats `json:"progress"`
}

type TestStats struct {
	TotalSuites     int `json:"totalSuites"`
	CompletedSuites int `json:"completedSuites"`

	TotalTests     int `json:"totalTests"`
	CompletedTests int `json:"completedTests"`
}

type TestSuites struct {
	TestSuiteData []TestSuiteData
}

type TestSuiteData struct {
	Name      string         `json:"name"`
	Type      int            `json:"type"`
	TestCases []TestCaseData `json:"tests"`
}

type Suite struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Disabled bool     `json:"disabled"`
	Tests    []string `json:"tests"`
	Type     int      `json:"type"`
}

type TestInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	ModelVersion int    `json:"modelVersion"`
}

type TestCaseData struct {
	Name       string                 `json:"name"`
	Flow       string                 `json:"flow"`
	FlowName   string                 `json:"flowName"`
	FlowInput  map[string]interface{} `json:"flowInputs"`
	Activities []Activity             `json:"activity"`
	Valid      bool                   `json:"valid"`
}

type Activity struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Mock          map[string]interface{} `json:"mock"`
	Assertion     []AssertionUT          `json:"assertion"`
	SkipExecution bool                   `json:"skipExecution"`
	Type          int                    `json:"type"`
}

type AssertionUT struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Expression interface{} `json:"expression"`
	Type       int
}

type CompleteReport struct {
	Report *Report       `json:"report"`
	Result *ReportResult `json:"result"`
}

type ReportResult struct {
	TotalSuites  int `json:"totalSuites"`
	FailedSuites int `json:"failedSuites"`
}

type Report struct {
	SuiteReport []SuiteReport `json:"suites"`
}

type SuiteReport struct {
	Name         string       `json:"suiteName"`
	TestReport   []TestReport `json:"testCases"`
	SuiteResult  SuiteResult  `json:"suiteResult"`
	IsPlayMode   bool         `json:"-"`
	ModelVersion int          `json:"-"`
}

type SuiteResult struct {
	TotalTests      int `json:"totalTests"`
	FailedTests     int `json:"failedTests"`
	ErrorFailed     int `json:"errorFailed"`
	AssertionFailed int `json:"assertionFailed"`
}

type TestReport struct {
	Name                   string                 `json:"testName,omitempty"`
	Flow                   string                 `json:"flowName"`
	ActivityReport         []ActivityReport       `json:"activities"`
	LinkReport             []LinkReport           `json:"links,omitempty"`
	TestReportErrorHandler TestReportErrorHandler `json:"errorHandler,omitempty"`
	TestResult             *TestResult            `json:"testResult,omitempty"`
	TestStatus             string                 `json:"testStatus"`
	SubFlow                map[string]interface{} `json:"subFlow,omitempty"`
}

type TestReportErrorHandler struct {
	ActivityReport []ActivityReport `json:"activities"`
	LinkReport     []LinkReport     `json:"links,omitempty"`
}

type TestResult struct {
	TotalAssertions   int  `json:"totalAssertions"`
	ExecutedAssertion int  `json:"executedAssertions"`
	FailedAssertions  int  `json:"failedAssertions"`
	SkippedAssertions int  `json:"skippedAssertions"`
	SuccessAssertions int  `json:"successAssertions"`
	FlowExecuted      bool `json:"flowExecuted"`
}

type ActivityReport struct {
	ActivityName    string                 `json:"name"`
	AssertionReport []AssertionReport      `json:"assertionResult,omitempty"`
	ActivityStatus  string                 `json:"activityStatus,omitempty"`
	Message         string                 `json:"message,omitempty"`
	Type            string                 `json:"type,omitempty"`
	Inputs          map[string]interface{} `json:"input,omitempty"`
	Outputs         *interface{}           `json:"output,omitempty"`
	Error           map[string]interface{} `json:"error,omitempty"`
}

type LinkReport struct {
	LinkName string `json:"linkName,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
}

type AssertionReport struct {
	Name                string `json:"name"`
	Status              string `json:"status"`
	Message             string `json:"message"`
	ExpressionType      string `json:"expressionType"`
	Expression          string `json:"expression"`
	ExpressionEvaluated string `json:"expressionEvaluated"`
}
