package unittest

import (
	"flag"
	"fmt"
	"github.com/tibco/wi-contrib/engine/cmd"
	"os"
	"path/filepath"
	"strings"
)

const (
	Flow = "github.com/project-flogo/flow"
)

func init() {
	cmd.Registry("test", &utCommand{})
	cmd.Registry("-test", &utCommand{})
	cmd.Registry("--test", &utCommand{})
	cmd.Registry("--ut", &utCommand{})

}

type utCommand struct {
	appJSON         string
	flogoTestJSON   string
	testSuites      string
	listTestSuites  string
	name            string
	output          string
	testCases       string
	testFlows       string
	collectIO       bool
	collectCoverage bool
}

func (b *utCommand) Name() string {
	return "-test"
}

func (b *utCommand) Description() string {
	return "Run Test Suite for the Flogo App"
}

func (b *utCommand) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.appJSON, "app", "", "Application Flogo JSON")
	fs.StringVar(&b.flogoTestJSON, "test-file", "", "Flogo Test JSON")
	fs.StringVar(&b.testSuites, "test-suites", "", "Test Suites")
	fs.StringVar(&b.testCases, "test-cases", "", "Test Cases")
	fs.StringVar(&b.testFlows, "test-flows", "", "Test Flows")
	fs.StringVar(&b.name, "result-filename", "", "Name of the test result file. If not provided, file with name <APPNAME> will be generated")
	fs.StringVar(&b.output, "output-dir", "", "Folder where test result file will be created. By default, binary will be created in current directory")
	fs.BoolVar(&b.collectIO, "test-preserve-io", false, "Collect Input & Output for executed activities")
	fs.BoolVar(&b.collectCoverage, "collect-coverage", false, "Collect Input & Output for executed activities")

}

func (b *utCommand) PrintUsage() {
	execName := os.Args[0]

	usage := "Command: \n" +
		"        " + execName + " --test --app <application json path> --test-file <test json path> --test-suites <test suite names> --output-dir < test result output directory>  --result-filename <output file name> \n" +
		"        Application json is optional. If application json is not provided it will take the embedded app in the binary.\n" +
		"        Test suites is optional. If test suites are not provided, tests will be executed for all test suites in the test file .\n" +
		"        Output directory is optional. If output directory is not provided it will store the test result in the working directory.\n" +
		"        Test result file is optional. If test result file name is not provided it will store as <App Name>.testresult\n" +
		"Usage:\n" +
		"        " + execName + " --test --app /Users/app/flogo.json --test-file /Users/app/ut.flogotest --test-suites \"suite1,suite2\" --output-dir /home/apps/tests --result-filename unitestapp\n"
	fmt.Println(usage)
}

func (b *utCommand) IsShimCommand() bool {
	return false
}

func (b *utCommand) Parse() bool {
	return true
}

func (b *utCommand) Run(args []string, appJson string) error {

	config := &UnitTestConfig{}

	if !isTCIEnv() && (b.output != "") {
		if b.output != "" {
			path := strings.TrimRight(b.output, "/")
			if !filepath.IsAbs(path) {
				return fmt.Errorf("Output location provided is invalid. Provide valid folder path")
			}
		}
	}

	config.collectIO = b.collectIO
	config.collectCoverage = b.collectCoverage
	config.output = b.output
	config.name = b.name
	err := config.loadFlogoJSON(appJson, b.appJSON)
	if err != nil {
		return err
	}

	if !isTCIEnv() && b.flogoTestJSON == "" {
		return fmt.Errorf("Test file not provided. Provide test file path with --testfile argument")
	}
	err = config.loadTestFile(b.flogoTestJSON)
	if err != nil {
		return err
	}

	config.loadTestSuites(b.testSuites)

	config.loadTestCases(b.testCases)

	config.loadTestFlows(b.testFlows)

	err = config.process()
	if err != nil {
		return err
	}

	return nil
}
