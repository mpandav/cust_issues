package unittest

import (
	"fmt"
	"github.com/project-flogo/flow/support"
	"strings"
)

func (report *SuiteReport) addFlowFailureReport(test *TestCaseData, err error) {

	testReport := &TestReport{}
	testReport.Flow = test.FlowName
	testReport.Name = test.Name
	testReport.TestResult = &TestResult{}
	testReport.TestResult.FlowExecuted = false
	testReport.ActivityReport = []ActivityReport{}
	testReport.TestStatus = "Testcase failed to execute with error: " + err.Error()

	report.SuiteResult.TotalTests++
	report.SuiteResult.FailedTests++
	report.SuiteResult.ErrorFailed++
	report.TestReport = append(report.TestReport, *testReport)
}

func (report *SuiteReport) addFlowFailureReportIO(test *TestCaseData, err error, coverage *support.Coverage, interceptors []*support.TaskInterceptor) {

	dataMap := report.getSubFlowDataMap(test, interceptors, coverage)

	subFlowMap := make(map[string]map[string]string)
	for _, subFlowCoverage := range coverage.SubFlowCoverage {
		if val, ok := subFlowMap[subFlowCoverage.HostFlow]; ok {
			val[subFlowCoverage.SubFlowActivity] = subFlowCoverage.SubFlowName
		} else {
			subFlowMap[subFlowCoverage.HostFlow] = make(map[string]string)
			subFlowMap[subFlowCoverage.HostFlow][subFlowCoverage.SubFlowActivity] = subFlowCoverage.SubFlowName
		}

	}

	testReport := &TestReport{}
	testReport.Flow = test.FlowName
	testReport.Name = test.Name
	testReport.TestResult = &TestResult{}
	testReport.TestResult.FlowExecuted = false
	testReport.ActivityReport = []ActivityReport{}
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: []ActivityReport{},
		LinkReport:     []LinkReport{},
	}
	testReport.TestStatus = "Testcase failed to execute with error: " + err.Error()

	var activityReportList []ActivityReport
	var errorHandlerActivityReport []ActivityReport = make([]ActivityReport, 0)
	var linkReportList []LinkReport = make([]LinkReport, 0)
	var errorHandlerLinkReportList []LinkReport

	for _, activity := range coverage.ActivityCoverage {

		if activity.FlowName != test.FlowName {
			continue
		}
		activityReport := &ActivityReport{}
		activityReport.ActivityName = activity.ActivityName
		activityReport.Inputs = activity.Inputs
		activityReport.Outputs = &activity.Outputs
		activityReport.Error = activity.Error

		if activity != nil {
			if activity.IsMainFlow {
				activityReportList = append(activityReportList, *activityReport)
			} else {
				errorHandlerActivityReport = append(errorHandlerActivityReport, *activityReport)
			}
		} else {
			activityReportList = append(activityReportList, *activityReport)
		}
	}

	for _, link := range coverage.TransitionCoverage {

		if link.FlowName != test.FlowName {
			continue
		}
		linkReport := &LinkReport{}
		linkReport.LinkName = link.TransitionName
		linkReport.To = link.TransitionTo
		linkReport.From = link.TransitionFrom
		if link.IsMainFlow {
			linkReportList = append(linkReportList, *linkReport)
		} else {
			errorHandlerLinkReportList = append(errorHandlerLinkReportList, *linkReport)
		}

	}

	testReport.ActivityReport = activityReportList
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: errorHandlerActivityReport,
		LinkReport:     errorHandlerLinkReportList,
	}
	testReport.LinkReport = linkReportList

	for k, testReport := range dataMap {
		if val, ok := subFlowMap[k]; ok {
			subMap := make(map[string]interface{})
			for k1, v1 := range val {
				if val, ok := dataMap[v1]; ok {
					subMap[k1] = val
				}
			}
			testReport.SubFlow = subMap
		}
	}

	if _, ok := subFlowMap[test.FlowName]; ok {
		subFlow := subFlowMap[test.FlowName]
		subMap := make(map[string]interface{})
		for k, v := range subFlow {
			if val, ok := dataMap[v]; ok {
				subMap[k] = val
			}
		}
		testReport.SubFlow = subMap
	}
	
	report.TestReport = append(report.TestReport, *testReport)
	report.SuiteResult.TotalTests++
	report.SuiteResult.FailedTests++
	report.SuiteResult.ErrorFailed++
}

func (report *SuiteReport) addFlowFailureReportWithCoverage(test *TestCaseData, err error, coverage *support.Coverage) {

	testReport := &TestReport{}
	testReport.Flow = test.FlowName
	testReport.Name = test.Name
	testReport.TestResult = &TestResult{}
	testReport.TestResult.FlowExecuted = false
	testReport.ActivityReport = []ActivityReport{}
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: []ActivityReport{},
		LinkReport:     []LinkReport{},
	}
	testReport.TestStatus = "Testcase failed to execute with error: " + err.Error()

	var activityReportList []ActivityReport
	var errorHandlerActivityReport []ActivityReport = make([]ActivityReport, 0)
	var linkReportList []LinkReport = make([]LinkReport, 0)
	var errorHandlerLinkReportList []LinkReport

	for _, activity := range coverage.ActivityCoverage {

		if activity.FlowName != test.FlowName {
			continue
		}
		activityReport := &ActivityReport{}
		activityReport.ActivityName = activity.ActivityName
		activityReport.Inputs = activity.Inputs
		activityReport.Outputs = &activity.Outputs
		activityReport.Error = activity.Error

		if activity != nil {
			if activity.IsMainFlow {
				activityReportList = append(activityReportList, *activityReport)
			} else {
				errorHandlerActivityReport = append(errorHandlerActivityReport, *activityReport)
			}
		} else {
			activityReportList = append(activityReportList, *activityReport)
		}
	}

	for _, link := range coverage.TransitionCoverage {

		if link.FlowName != test.FlowName {
			continue
		}
		linkReport := &LinkReport{}
		linkReport.LinkName = link.TransitionName
		linkReport.To = link.TransitionTo
		linkReport.From = link.TransitionFrom
		if link.IsMainFlow {
			linkReportList = append(linkReportList, *linkReport)
		} else {
			errorHandlerLinkReportList = append(errorHandlerLinkReportList, *linkReport)
		}

	}

	testReport.ActivityReport = activityReportList
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: errorHandlerActivityReport,
		LinkReport:     errorHandlerLinkReportList,
	}
	testReport.LinkReport = linkReportList

	report.TestReport = append(report.TestReport, *testReport)
	report.SuiteResult.TotalTests++
	report.SuiteResult.FailedTests++
	report.SuiteResult.ErrorFailed++
}

func (report *SuiteReport) addFlowSuccessToReport(test *TestCaseData, interceptors []*support.TaskInterceptor, coverage *support.Coverage) {

	testReport := &TestReport{}
	testReport.Flow = test.FlowName
	testReport.Name = test.Name
	testReport.TestResult = &TestResult{}
	testReport.TestResult.FlowExecuted = true
	testReport.TestStatus = ""
	var flowOpReport *ActivityReport
	var activityReportList []ActivityReport
	var failAssertions []string
	var skippedAssertions []string
	var successAssertions []string
	var errorHandlerActivityReport []ActivityReport

	var activityMap = make(map[string]*support.ActivityCoverage)
	for _, activity := range coverage.ActivityCoverage {
		activityMap[activity.ActivityName] = activity
	}

	for _, interceptor := range interceptors {

		activityReport := &ActivityReport{}
		activityReport.ActivityName = strings.Replace(interceptor.ID, test.FlowName+"-", "", 1)

		if interceptor.Result == support.NotExecuted && interceptor.Outputs != nil {
			// Do not add to result mocked activities which are not executed.
			continue
		}

		flowOp := false
		if interceptor.ID == test.FlowName+"-"+"_flowOutput" {
			flowOp = true
			activityReport.ActivityName = "Flow Output"
		}

		switch interceptor.Result {
		case support.NotExecuted:
			activityReport.ActivityStatus = "not-executed"
		case support.Pass:
			activityReport.ActivityStatus = "pass"
		case support.Fail:
			activityReport.ActivityStatus = "fail"
			activityReport.Message = interceptor.Message
		case support.Mocked:
			activityReport.ActivityStatus = "mocked"
		}
		activityReport.Type = getType(interceptor.Type)

		var assertionList = make([]AssertionReport, 0)

		for _, assertion := range interceptor.Assertions {
			assertionReport := &AssertionReport{}
			testReport.TestResult.TotalAssertions++
			assertionReport.Name = assertion.Name
			assertionReport.Message = assertion.Message
			if assertion.Result > 0 {
				testReport.TestResult.ExecutedAssertion++
				if assertion.Result == 1 {
					assertionReport.Status = "pass"
					successAssertions = append(successAssertions, activityReport.ActivityName+":"+assertion.Name)
					testReport.TestResult.SuccessAssertions++
				} else if assertion.Result == 2 {
					assertionReport.Status = "fail"
					testReport.TestResult.FailedAssertions++
					failAssertions = append(failAssertions, activityReport.ActivityName+":"+assertion.Name)
				} else if assertion.Result == 4 {
					assertionReport.Status = "not-executed"
					assertionReport.Message = "Assertion not executed"
					testReport.TestResult.SkippedAssertions++
					skippedAssertions = append(skippedAssertions, activityReport.ActivityName+":"+assertion.Name)
				}

			} else {
				if interceptor.Result == support.Fail {
					assertionReport.Status = "not-executed"
					if assertion.Message == "" {
						assertionReport.Message = "Activity failed to execute"
					}
				} else {
					assertionReport.Status = "not-executed"
					if assertion.Message == "" {
						assertionReport.Message = "Activity not Executed"
					}
					testReport.TestResult.SkippedAssertions++
					skippedAssertions = append(skippedAssertions, activityReport.ActivityName+":"+assertion.Name)
				}

			}
			if assertion.EvalResult.ExpressionType != "" {
				assertionReport.ExpressionType = assertion.EvalResult.ExpressionType
				assertionReport.ExpressionEvaluated = assertion.EvalResult.ExpressionEvaluation
				assertionReport.Expression = assertion.Expression.(string)
			} else {
				assertionReport.ExpressionEvaluated = "Assertion not evaluated"
				assertionReport.Expression = assertion.Expression.(string)
				assertionReport.ExpressionType = "NA"
			}

			assertionList = append(assertionList, *assertionReport)
		}

		activityReport.AssertionReport = assertionList

		activity := activityMap[activityReport.ActivityName]
		if !flowOp {
			if activity != nil {
				if activity.IsMainFlow {
					activityReportList = append(activityReportList, *activityReport)
				} else {
					errorHandlerActivityReport = append(errorHandlerActivityReport, *activityReport)
				}
			} else {
				activityReportList = append(activityReportList, *activityReport)
			}
		} else {
			flowOpReport = activityReport
		}

	}
	if flowOpReport != nil {
		activityReportList = append(activityReportList, *flowOpReport)

	}
	testReport.ActivityReport = activityReportList
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: errorHandlerActivityReport,
	}
	report.SuiteResult.TotalTests++
	if testReport.TestResult.FailedAssertions > 0 || testReport.TestResult.SkippedAssertions > 0 {
		report.SuiteResult.FailedTests++
		report.SuiteResult.AssertionFailed++
	}
	report.TestReport = append(report.TestReport, *testReport)

	if testReport.TestResult.FailedAssertions > 0 || testReport.TestResult.SkippedAssertions > 0 {
		fmt.Printf(showError("\nTest case %s execution failed. %d out of %d assertions failed or skipped."), test.Name, testReport.TestResult.FailedAssertions+testReport.TestResult.SkippedAssertions, testReport.TestResult.TotalAssertions)
		if testReport.TestResult.SuccessAssertions > 0 {
			fmt.Printf(showSuccess("\nAssertions executed successfully \n %s"), successAssertions)
		}
		if testReport.TestResult.FailedAssertions > 0 {
			fmt.Printf(showError("\nFailed assertions \n %s"), failAssertions)
		}
		if testReport.TestResult.SkippedAssertions > 0 {
			fmt.Printf(showError("\nSkipped assertions \n %s\n"), skippedAssertions)
		}

	} else {
		fmt.Printf(showSuccess("\nTest case %s execution completed. All assertions passed"), test.Name)
		if testReport.TestResult.SuccessAssertions > 0 {
			fmt.Printf(showSuccess("\nAssertions executed successfully \n %s\n"), successAssertions)
		}
	}

}

func (report *SuiteReport) addFlowSuccessToReportWithIO(test *TestCaseData, interceptors []*support.TaskInterceptor, coverage *support.Coverage) {

	dataMap := report.getSubFlowDataMap(test, interceptors, coverage)

	subFlowMap := make(map[string]map[string]string)
	for _, subFlowCoverage := range coverage.SubFlowCoverage {
		if val, ok := subFlowMap[subFlowCoverage.HostFlow]; ok {
			val[subFlowCoverage.SubFlowActivity] = subFlowCoverage.SubFlowName
		} else {
			subFlowMap[subFlowCoverage.HostFlow] = make(map[string]string)
			subFlowMap[subFlowCoverage.HostFlow][subFlowCoverage.SubFlowActivity] = subFlowCoverage.SubFlowName
		}

	}

	testReport := &TestReport{}
	testReport.Flow = test.FlowName
	testReport.Name = test.Name
	testReport.TestResult = &TestResult{}
	testReport.TestResult.FlowExecuted = true
	testReport.TestStatus = ""
	var flowOpReport *ActivityReport
	var activityReportList []ActivityReport
	var errorHandlerActivityReport []ActivityReport = make([]ActivityReport, 0)
	var linkReportList []LinkReport
	var errorHandlerLinkReportList []LinkReport = make([]LinkReport, 0)

	var failAssertions []string
	var skippedAssertions []string
	var successAssertions []string

	var interceptorMap = make(map[string]*support.TaskInterceptor)

	for _, interceptor := range interceptors {
		activityName := strings.Replace(interceptor.ID, test.FlowName+"-", "", 1)
		interceptorMap[activityName] = interceptor
	}

	for _, activity := range coverage.ActivityCoverage {

		if activity.FlowName != test.FlowName {
			continue
		}
		activityReport := &ActivityReport{}
		activityReport.ActivityName = activity.ActivityName
		activityReport.Inputs = activity.Inputs
		activityReport.Outputs = &activity.Outputs
		activityReport.Error = activity.Error

		interceptor, ok := interceptorMap[activity.ActivityName]
		if ok {

			failAssertions, skippedAssertions, successAssertions = report.processInterceptor(interceptor, activityReport, testReport, failAssertions, skippedAssertions, successAssertions)

		}
		if activity != nil {

			if activity.IsMainFlow {
				activityReportList = append(activityReportList, *activityReport)
			} else {
				errorHandlerActivityReport = append(errorHandlerActivityReport, *activityReport)
			}
		} else {
			activityReportList = append(activityReportList, *activityReport)
		}
	}

	for _, link := range coverage.TransitionCoverage {

		if link.FlowName != test.FlowName {
			continue
		}
		linkReport := &LinkReport{}
		linkReport.LinkName = link.TransitionName
		linkReport.To = link.TransitionTo
		linkReport.From = link.TransitionFrom
		if link.IsMainFlow {
			linkReportList = append(linkReportList, *linkReport)
		} else {
			errorHandlerLinkReportList = append(errorHandlerLinkReportList, *linkReport)
		}

	}

	flowOp, ok := interceptorMap["_flowOutput"]
	if ok {
		flowOpReport = &ActivityReport{}
		flowOpReport.ActivityName = "Flow Output"
		failAssertions, skippedAssertions, successAssertions = report.processInterceptor(flowOp, flowOpReport, testReport, failAssertions, skippedAssertions, successAssertions)

	}

	if flowOpReport != nil {
		activityReportList = append(activityReportList, *flowOpReport)
	}
	testReport.ActivityReport = activityReportList
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: errorHandlerActivityReport,
		LinkReport:     errorHandlerLinkReportList,
	}
	testReport.LinkReport = linkReportList
	report.SuiteResult.TotalTests++
	if testReport.TestResult.FailedAssertions > 0 || testReport.TestResult.SkippedAssertions > 0 {
		report.SuiteResult.FailedTests++
		report.SuiteResult.AssertionFailed++
	}

	for k, testReport := range dataMap {
		if val, ok := subFlowMap[k]; ok {
			subMap := make(map[string]interface{})
			for k1, v1 := range val {
				if val, ok := dataMap[v1]; ok {
					subMap[k1] = val
				}
			}
			testReport.SubFlow = subMap
		}
	}

	if _, ok := subFlowMap[test.FlowName]; ok {
		subFlow := subFlowMap[test.FlowName]
		subMap := make(map[string]interface{})
		for k, v := range subFlow {
			if val, ok := dataMap[v]; ok {
				subMap[k] = val
			}
		}
		testReport.SubFlow = subMap
	}

	report.TestReport = append(report.TestReport, *testReport)

	if testReport.TestResult.FailedAssertions > 0 || testReport.TestResult.SkippedAssertions > 0 {
		fmt.Printf(showError("\nTest case %s execution failed. %d out of %d assertions failed or skipped."), test.Name, testReport.TestResult.FailedAssertions+testReport.TestResult.SkippedAssertions, testReport.TestResult.TotalAssertions)
		if testReport.TestResult.SuccessAssertions > 0 {
			fmt.Printf(showSuccess("\nAssertions executed successfully \n %s"), successAssertions)
		}
		if testReport.TestResult.FailedAssertions > 0 {
			fmt.Printf(showError("\nFailed assertions \n %s"), failAssertions)
		}
		if testReport.TestResult.SkippedAssertions > 0 {
			fmt.Printf(showError("\nSkipped assertions \n %s\n"), skippedAssertions)
		}

	} else {
		fmt.Printf(showSuccess("\nTest case %s execution completed. All assertions passed"), test.Name)
		if testReport.TestResult.SuccessAssertions > 0 {
			fmt.Printf(showSuccess("\nAssertions executed successfully \n %s\n"), successAssertions)
		}
	}

}

func (report *SuiteReport) addFlowSuccessToReportWithCoverage(test *TestCaseData, interceptors []*support.TaskInterceptor, coverage *support.Coverage) {

	dataMap := report.getSubFlowDataMap(test, interceptors, coverage)

	subFlowMap := make(map[string]map[string]string)
	for _, subFlowCoverage := range coverage.SubFlowCoverage {
		if val, ok := subFlowMap[subFlowCoverage.HostFlow]; ok {
			val[subFlowCoverage.SubFlowActivity] = subFlowCoverage.SubFlowName
		} else {
			subFlowMap[subFlowCoverage.HostFlow] = make(map[string]string)
			subFlowMap[subFlowCoverage.HostFlow][subFlowCoverage.SubFlowActivity] = subFlowCoverage.SubFlowName
		}

	}

	testReport := &TestReport{}
	testReport.Flow = test.FlowName
	testReport.Name = test.Name
	testReport.TestResult = &TestResult{}
	testReport.TestResult.FlowExecuted = true
	testReport.TestStatus = ""
	var flowOpReport *ActivityReport
	var activityReportList []ActivityReport
	var errorHandlerActivityReport []ActivityReport = make([]ActivityReport, 0)
	var linkReportList []LinkReport
	var errorHandlerLinkReportList []LinkReport = make([]LinkReport, 0)

	var failAssertions []string
	var skippedAssertions []string
	var successAssertions []string

	var interceptorMap = make(map[string]*support.TaskInterceptor)

	for _, interceptor := range interceptors {
		activityName := strings.Replace(interceptor.ID, test.FlowName+"-", "", 1)
		interceptorMap[activityName] = interceptor
	}

	for _, activity := range coverage.ActivityCoverage {

		if activity.FlowName != test.FlowName {
			continue
		}
		activityReport := &ActivityReport{}
		activityReport.ActivityName = activity.ActivityName

		interceptor, ok := interceptorMap[activity.ActivityName]
		if ok {
			if interceptor.Result == support.NotExecuted && interceptor.Outputs != nil {
				// Do not add to result mocked activities which are not executed.
				continue
			}

			failAssertions, skippedAssertions, successAssertions = report.processInterceptor(interceptor, activityReport, testReport, failAssertions, skippedAssertions, successAssertions)

		}
		if activity != nil {

			if activity.IsMainFlow {
				activityReportList = append(activityReportList, *activityReport)
			} else {
				errorHandlerActivityReport = append(errorHandlerActivityReport, *activityReport)
			}
		} else {
			activityReportList = append(activityReportList, *activityReport)
		}
	}

	for _, link := range coverage.TransitionCoverage {

		if link.FlowName != test.FlowName {
			continue
		}
		linkReport := &LinkReport{}
		linkReport.LinkName = link.TransitionName
		linkReport.To = link.TransitionTo
		linkReport.From = link.TransitionFrom
		if link.IsMainFlow {
			linkReportList = append(linkReportList, *linkReport)
		} else {
			errorHandlerLinkReportList = append(errorHandlerLinkReportList, *linkReport)
		}

	}

	flowOp, ok := interceptorMap["_flowOutput"]
	if ok {
		flowOpReport = &ActivityReport{}
		flowOpReport.ActivityName = "Flow Output"
		failAssertions, skippedAssertions, successAssertions = report.processInterceptor(flowOp, flowOpReport, testReport, failAssertions, skippedAssertions, successAssertions)

	}

	if flowOpReport != nil {
		activityReportList = append(activityReportList, *flowOpReport)
	}
	testReport.ActivityReport = activityReportList
	testReport.TestReportErrorHandler = TestReportErrorHandler{
		ActivityReport: errorHandlerActivityReport,
		LinkReport:     errorHandlerLinkReportList,
	}
	testReport.LinkReport = linkReportList
	report.SuiteResult.TotalTests++
	if testReport.TestResult.FailedAssertions > 0 || testReport.TestResult.SkippedAssertions > 0 {
		report.SuiteResult.FailedTests++
		report.SuiteResult.AssertionFailed++
	}

	for k, testReport := range dataMap {
		if val, ok := subFlowMap[k]; ok {
			subMap := make(map[string]interface{})
			for k1, v1 := range val {
				if val, ok := dataMap[v1]; ok {
					subMap[k1] = val
				}
			}
			testReport.SubFlow = subMap
		}
	}

	if _, ok := subFlowMap[test.FlowName]; ok {
		subFlow := subFlowMap[test.FlowName]
		subMap := make(map[string]interface{})
		for k, v := range subFlow {
			if val, ok := dataMap[v]; ok {
				subMap[k] = val
			}
		}
		testReport.SubFlow = subMap
	}

	report.TestReport = append(report.TestReport, *testReport)

	if testReport.TestResult.FailedAssertions > 0 || testReport.TestResult.SkippedAssertions > 0 {
		fmt.Printf(showError("\nTest case %s execution failed. %d out of %d assertions failed or skipped."), test.Name, testReport.TestResult.FailedAssertions+testReport.TestResult.SkippedAssertions, testReport.TestResult.TotalAssertions)
		if testReport.TestResult.SuccessAssertions > 0 {
			fmt.Printf(showSuccess("\nAssertions executed successfully \n %s"), successAssertions)
		}
		if testReport.TestResult.FailedAssertions > 0 {
			fmt.Printf(showError("\nFailed assertions \n %s"), failAssertions)
		}
		if testReport.TestResult.SkippedAssertions > 0 {
			fmt.Printf(showError("\nSkipped assertions \n %s\n"), skippedAssertions)
		}

	} else {
		fmt.Printf(showSuccess("\nTest case %s execution completed. All assertions passed"), test.Name)
		if testReport.TestResult.SuccessAssertions > 0 {
			fmt.Printf(showSuccess("\nAssertions executed successfully \n %s\n"), successAssertions)
		}
	}

}

func (report *SuiteReport) getSubFlowDataMap(test *TestCaseData, interceptors []*support.TaskInterceptor, coverage *support.Coverage) map[string]*TestReport {
	subFlowList := make(map[string]*TestReport)

	for _, activity := range coverage.ActivityCoverage {
		if activity == nil {
			continue
		}
		if activity.FlowName == test.FlowName {
			continue
		}
		activityReport := &ActivityReport{}
		activityReport.ActivityName = activity.ActivityName
		activityReport.Inputs = activity.Inputs
		activityReport.Outputs = &activity.Outputs
		activityReport.Error = activity.Error

		val, ok := subFlowList[activity.FlowName]
		if ok {
			if activity.IsMainFlow {
				val.ActivityReport = append(val.ActivityReport, *activityReport)
			} else {
				val.TestReportErrorHandler.ActivityReport = append(val.TestReportErrorHandler.ActivityReport, *activityReport)
			}
		} else {
			val = &TestReport{
				Flow:           activity.FlowName,
				ActivityReport: make([]ActivityReport, 0),
				LinkReport:     make([]LinkReport, 0),
				TestReportErrorHandler: TestReportErrorHandler{
					ActivityReport: make([]ActivityReport, 0),
					LinkReport:     make([]LinkReport, 0),
				},
				SubFlow: make(map[string]interface{}),
			}
			if activity.IsMainFlow {
				val.ActivityReport = append(val.ActivityReport, *activityReport)
			} else {
				val.TestReportErrorHandler.ActivityReport = append(val.TestReportErrorHandler.ActivityReport, *activityReport)
			}
			subFlowList[activity.FlowName] = val
		}
	}

	for _, link := range coverage.TransitionCoverage {

		if link.FlowName == test.FlowName {
			continue
		}
		linkReport := &LinkReport{}
		linkReport.LinkName = link.TransitionName
		linkReport.To = link.TransitionTo
		linkReport.From = link.TransitionFrom

		val, ok := subFlowList[link.FlowName]
		if ok {
			if link.IsMainFlow {
				val.LinkReport = append(val.LinkReport, *linkReport)
			} else {
				val.TestReportErrorHandler.LinkReport = append(val.TestReportErrorHandler.LinkReport, *linkReport)
			}
		}
	}

	return subFlowList

}

func (report *SuiteReport) processInterceptor(interceptor *support.TaskInterceptor, activityReport *ActivityReport, testReport *TestReport, failAssertions []string, skippedAssertions []string, successAssertions []string) ([]string, []string, []string) {
	switch interceptor.Result {
	case support.NotExecuted:
		activityReport.ActivityStatus = "not-executed"
	case support.Pass:
		activityReport.ActivityStatus = "pass"
	case support.Fail:
		activityReport.ActivityStatus = "fail"
		activityReport.Message = interceptor.Message
	case support.Mocked:
		activityReport.ActivityStatus = "mocked"
	}
	activityReport.Type = getType(interceptor.Type)

	var assertionList = make([]AssertionReport, 0)

	for _, assertion := range interceptor.Assertions {
		assertionReport := &AssertionReport{}
		testReport.TestResult.TotalAssertions++
		assertionReport.Name = assertion.Name
		assertionReport.Message = assertion.Message
		if assertion.Result > 0 {
			testReport.TestResult.ExecutedAssertion++
			if assertion.Result == 1 {
				assertionReport.Status = "pass"
				testReport.TestResult.SuccessAssertions++
				successAssertions = append(successAssertions, activityReport.ActivityName+":"+assertion.Name)
			} else if assertion.Result == 2 {
				assertionReport.Status = "fail"
				testReport.TestResult.FailedAssertions++
				failAssertions = append(failAssertions, activityReport.ActivityName+":"+assertion.Name)
			} else if assertion.Result == 4 {
				assertionReport.Status = "not-executed"
				assertionReport.Message = "Assertion not executed"
				testReport.TestResult.SkippedAssertions++
				skippedAssertions = append(skippedAssertions, activityReport.ActivityName+":"+assertion.Name)
			}

		} else {
			if interceptor.Result == support.Fail {
				assertionReport.Status = "not-executed"
				if assertion.Message == "" {
					assertionReport.Message = "Activity failed to execute"
				}
			} else {
				assertionReport.Status = "not-executed"
				if assertion.Message == "" {
					assertionReport.Message = "Activity not Executed"
				}
				testReport.TestResult.SkippedAssertions++
				skippedAssertions = append(skippedAssertions, activityReport.ActivityName+":"+assertion.Name)
			}

		}
		if assertion.EvalResult.ExpressionType != "" {
			assertionReport.ExpressionType = assertion.EvalResult.ExpressionType
			assertionReport.ExpressionEvaluated = assertion.EvalResult.ExpressionEvaluation
			assertionReport.Expression = assertion.Expression.(string)
		} else {
			assertionReport.ExpressionEvaluated = "Assertion not evaluated"
			assertionReport.Expression = assertion.Expression.(string)
			assertionReport.ExpressionType = "NA"
		}
		assertionList = append(assertionList, *assertionReport)
	}

	activityReport.AssertionReport = assertionList
	return failAssertions, skippedAssertions, successAssertions
}

func getType(assertionType int) string {
	switch assertionType {
	case support.AssertionActivity:
		return "Assert On Outputs"
	case support.AssertionException:
		return "Assert On Error"
	case support.SkipActivity:
		return "Skip Execution"
	case support.MockActivity:
		return "Mock Outputs"
	case support.MockException:
		return "Mock Error"
	}
	return ""
}
