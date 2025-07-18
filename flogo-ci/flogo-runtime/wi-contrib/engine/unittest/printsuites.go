package unittest

import (
	"flag"
	"fmt"
	"github.com/tibco/wi-contrib/engine/cmd"
	"os"
)

func init() {
	cmd.Registry("listsuites", &printSuiteCommand{})
	cmd.Registry("-listsuites", &printSuiteCommand{})
	cmd.Registry("--listsuites", &printSuiteCommand{})
}

type printSuiteCommand struct {
	testFile string
}

func (b *printSuiteCommand) Name() string {
	return "-listSuites"
}

func (b *printSuiteCommand) Description() string {
	return "List the Suites in the Test File"
}

func (b *printSuiteCommand) Run(args []string, appJson string) error {
	config := &UnitTestConfig{}

	if b.testFile == "" {
		return fmt.Errorf("Test file not provided. Provide test file path with --testfile argument")
	}
	err := config.loadTestFile(b.testFile)
	if err != nil {
		return err
	}

	testParser := &TestParser{}
	testParser.suiteDataJson = config.suiteDataJson
	testParser.runAllTests = true

	err = testParser.loadTestSuites()
	if err != nil {
		return err
	}

	printTestSuites(&testParser.testSuites.TestSuiteData)

	return nil
}

func printTestSuites(suites *[]TestSuiteData) {
	for _, suite := range *suites {
		fmt.Println(suite.Name)
	}
}
func (b *printSuiteCommand) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.testFile, "testfile", "", "Flogo Test JSON File")
}

func (b *printSuiteCommand) PrintUsage() {
	execName := os.Args[0]

	usage := "Command: \n" +
		"        " + execName + "--testfile <test json path> \n" +
		"Usage:\n" +
		"        " + execName + " --testfile /Users/app/ut.flogotest \n"
	fmt.Println(usage)
}

func (b *printSuiteCommand) IsShimCommand() bool {
	return false
}

func (b *printSuiteCommand) Parse() bool {
	return true
}
