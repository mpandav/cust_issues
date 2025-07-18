package unittest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/project-flogo/core/action"
	enginerunner "github.com/project-flogo/core/engine/runner"
	"github.com/project-flogo/flow/instance"
	"github.com/project-flogo/flow/support"
	"github.com/project-flogo/flow/tester"
	"golang.org/x/net/context"

	"io/ioutil"
	"os"
	"time"
)

func (runner *TestRunner) runTests() {

	runner.testProgress = &TestProgress{}
	runner.testProgress.TestStats = &TestStats{}
	runner.testProgress.TestStats.TotalSuites = len(runner.testSuites.TestSuiteData)
	for _, suite := range runner.testSuites.TestSuiteData {
		runner.testProgress.TestStats.TotalTests = runner.testProgress.TestStats.TotalTests + len(suite.TestCases)
	}

	report := &Report{}
	report.SuiteReport = []SuiteReport{}
	runner.report = report

	if isTCIEnv() {
		runner.ticker = time.NewTicker(8 * time.Second)
		runner.tickerChan = make(chan bool, 1)
		go runner.startTicker()
	}

	for _, suit := range runner.testSuites.TestSuiteData {
		runner.runSuite(&suit)
	}

	if isTCIEnv() {
		runner.ticker.Stop()
		runner.tickerChan <- true
	}

}

func (runner *TestRunner) runSuite(suit *TestSuiteData) {

	suiteReport := SuiteReport{}
	suiteReport.Name = suit.Name

	if len(suit.TestCases) == 0 {
		fmt.Printf(showError("\nTest Suite %s has no test cases added. Test Suite execution will be skipped. \n\n"), suiteReport.Name)
		return

	}
	fmt.Printf(showInfo("\nRunning [%d] Test(s) from Test Suite [%s] \n\n"), len(suit.TestCases), suiteReport.Name)

	for _, test := range suit.TestCases {
		fmt.Printf(showInfo("\nRunning Test [%s] for Flow [%s] from Suite [%s] \n"), test.Name, test.FlowName, suit.Name)
		runner.runTest(&test, &suiteReport)
		runner.testProgress.TestStats.CompletedTests++
	}

	runner.report.SuiteReport = append(runner.report.SuiteReport, suiteReport)

	printSuiteResult(&suiteReport)
	runner.testProgress.TestStats.CompletedSuites++

}
func printSuiteResult(report *SuiteReport) {
	if report.SuiteResult.FailedTests > 0 {
		fmt.Printf(showError("\nTest suite %s executed. \nTests Run: %d \t Success: %d \t Failure: %d \t Error: %d  \n\n"), report.Name, report.SuiteResult.TotalTests, report.SuiteResult.TotalTests-report.SuiteResult.FailedTests, report.SuiteResult.AssertionFailed, report.SuiteResult.ErrorFailed)

	} else {
		fmt.Printf(showSuccess("\nTest suite %s executed. \nTests Run: %d \t Success: %d \t Failure: %d \t Error: %d  \n\n"), report.Name, report.SuiteResult.TotalTests, report.SuiteResult.TotalTests-report.SuiteResult.FailedTests, report.SuiteResult.AssertionFailed, report.SuiteResult.ErrorFailed)

	}
}

func (runner *TestRunner) runTest(test *TestCaseData, report *SuiteReport) {

	if !test.Valid {
		report.addFlowFailureReport(test, nil)
		fmt.Printf(showError("\nTest case %s execution skipped due to invalid flow inputs for flow %s."), test.Name, test.FlowName)
		return
	}

	if (len(test.Activities) == 0 || hasNoAssertion(test.Activities)) && !runner.collectIO {
		report.addFlowFailureReport(test, errors.New("Test case doesnt have any activity added with assertion(s)"))
		fmt.Printf(showError("\nTest case %s doesnt have any assertions. Test execution skipped."), test.Name)
		return
	} else {
		runner.playMode = true
		report.IsPlayMode = true
	}

	runner.runFlow(report, test)
}

func hasNoAssertion(activities []Activity) bool {
	for _, activity := range activities {
		if len(activity.Assertion) > 0 {
			return false
		}
	}
	return true
}

func (runner *TestRunner) runFlow(report *SuiteReport, test *TestCaseData) {

	factory := runner.factory
	inputs := generateFlowInput(test)
	interceptor := generateTaskInterceptor(test)
	coverage := &support.Coverage{
		ActivityCoverage:   make([]*support.ActivityCoverage, 0),
		TransitionCoverage: make([]*support.TransitionCoverage, 0),
		SubFlowCoverage:    make([]*support.SubFlowCoverage, 0),
	}

	request := tester.StartRequest{}
	request.Interceptor = &support.Interceptor{TaskInterceptors: interceptor, Coverage: coverage, CollectIO: runner.collectIO}

	act, _ := factory.New(&action.Config{Settings: map[string]interface{}{"flowURI": test.Flow}})
	execOptions := &instance.ExecOptions{Interceptor: request.Interceptor, Patch: request.Patch}
	ro := &instance.RunOptions{Op: instance.OpStart, ReturnID: false, FlowURI: test.Flow, ExecOptions: execOptions}
	inputs["_run_options"] = ro
	run := enginerunner.NewDirect()
	_, err := run.RunAction(context.Background(), act, inputs)

	if err != nil {
		fmt.Printf(showError("\nTest case %s execution failed. Flow %s linked to test case failed to execute with error %s."), test.Name, test.FlowName, err.Error())
		if runner.collectIO {
			report.addFlowFailureReportIO(test, err, coverage, interceptor)
		} else if runner.collectCoverage {
			report.addFlowFailureReportWithCoverage(test, err, coverage)
		} else {
			report.addFlowFailureReport(test, err)
		}
	} else {
		if runner.collectIO {
			report.addFlowSuccessToReportWithIO(test, interceptor, coverage)
		} else if runner.collectCoverage {
			report.addFlowSuccessToReportWithCoverage(test, interceptor, coverage)
		} else {
			report.addFlowSuccessToReport(test, interceptor, coverage)
		}
	}
}

func generateFlowInput(test *TestCaseData) map[string]interface{} {
	inputs := make(map[string]interface{})
	if len(test.FlowInput) > 0 {
		for k, v := range test.FlowInput {
			inputs[k] = v
		}
	} else {
		inputs = make(map[string]interface{}, 1)
	}
	return inputs
}

func generateTaskInterceptor(test *TestCaseData) []*support.TaskInterceptor {

	activities := test.Activities
	tasks := []*support.TaskInterceptor{}

	for _, activity := range activities {
		interceptor := &support.TaskInterceptor{}
		interceptor.ID = test.FlowName + "-" + activity.ID
		interceptor.Type = activity.Type
		interceptor.Skip = false
		if activity.Mock != nil {
			interceptor.Outputs = activity.Mock
		}
		interceptor.SkipExecution = activity.SkipExecution
		if activity.Assertion != nil && len(activity.Assertion) > 0 {
			interceptor.Assertions = []support.Assertion{}
			for _, assertion := range activity.Assertion {
				assertionut := &support.Assertion{}
				assertionut.ID = assertion.ID
				assertionut.Name = assertion.Name
				assertionut.Type = assertion.Type
				assertionut.Expression = assertion.Expression
				interceptor.Assertions = append(interceptor.Assertions, *assertionut)
			}
		}
		tasks = append(tasks, interceptor)
	}

	return tasks
}

func (runner *TestRunner) startTicker() {

	for {
		// Select statement
		select {
		// Case statement
		case <-runner.tickerChan:
			return

		// Case to print current time
		case tm := <-runner.ticker.C:
			op, _ := json.Marshal(runner.testProgress)
			os.Remove("output.json")
			ioutil.WriteFile("output.json", op, 777)
			runner.tickerTime = tm
		}
	}
}
