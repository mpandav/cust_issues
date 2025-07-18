package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/engine/runner"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/flow/definition"
	"github.com/project-flogo/flow/instance"
	"github.com/project-flogo/flow/tester"
)

const (
	RefFlow = "github.com/project-flogo/flow"
)

func init() {
	Registry("flowtest", &testCommand{})
	Registry("-flowtest", &testCommand{})
	Registry("--flowtest", &testCommand{})

}

type testCommand struct {
	serverPort string
	testInput  string
	listFlows  bool
	genData    string
	testOutput string
}

func (b *testCommand) IsShimCommand() bool {
	return false
}

func (b *testCommand) Name() string {
	return "-test"
}

func (b *testCommand) Description() string {
	return "Enable test mode"
}

func (b *testCommand) Run(args []string, appJson string) error {
	flogoApp, err := engine.LoadAppConfig(appJson, false)
	if err != nil {
		return fmt.Errorf("Failed to create engine: %s", err.Error())
	}
	return b.testMode(flogoApp)
}

func (b *testCommand) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.testInput, "flowin", "", "Path to flow specific test data file")
	fs.BoolVar(&b.listFlows, "flows", false, "List all flows in the application")
	fs.StringVar(&b.genData, "flowdata", "", "Generate test data for given flow")
	fs.StringVar(&b.testOutput, "flowout", "", "Write flow output(if applicable) to the file")
}

func (b *testCommand) Parse() bool {
	return true
}

func (b *testCommand) PrintUsage() {
	execName := os.Args[0]
	usage := "Options:\n" +
		"        -flows      List all flows in the application \n" +
		"        -flowdata   Generate input test data file for given flow \n" +
		"        -flowin     Path of flow input test data file  \n" +
		"        -flowout    Write flow output(if applicable) to the specified file. If not specified, output will be printed on console. \n" +
		"Usage:\n" +
		"        " + execName + " -test -flows \n" +
		"        " + execName + " -test -flowdata TestFlow \n" +
		"        " + execName + " -test -flowin MyApp_TestFlow_input.json \n" +
		"        " + execName + " -test -flowin MyApp_TestFlow_input.json -flowout MyApp_TestFlow_output.json \n"
	fmt.Println(usage)
}

func (b *testCommand) testMode(flogoApp *app.Config) error {

	function.ResolveAliases()
	//Remove Trigger part
	flogoApp.Triggers = nil
	e, err := engine.New(flogoApp)
	if err != nil {
		return fmt.Errorf("Failed to create engine: %s", err.Error())
	}

	if b.listFlows {
		if len(flogoApp.Resources) == 0 {
			return fmt.Errorf("No flows found in the application")
		}
		fmt.Println("Flows:")
		for _, v := range flogoApp.Resources {
			if strings.Index(v.ID, "flow:") > -1 {
				fmt.Println("   " + strings.Replace(v.ID, "flow:", "", -1))
			}
		}
	} else if b.genData != "" {
		flowDef := e.App().GetResource("flow:" + b.genData)
		if flowDef == nil {
			return fmt.Errorf("Flow - %s not found", b.genData)
		}

		flow, _ := flowDef.Object().(*definition.Definition)
		inputData := make(map[string]interface{}, 2)
		inputData["flowUri"] = "res://flow:" + b.genData

		dataAttr := make(map[string]interface{}, len(flow.Metadata().Input))
		for k, v := range flow.Metadata().Input {
			dataAttr[k] = getValue(v.Type())
		}
		inputData["data"] = dataAttr

		v, err := json.MarshalIndent(inputData, "", "    ")
		if err != nil {
			return fmt.Errorf("Failed to generate flow input data due to error - %s", err.Error())
		}
		fileName := flogoApp.Name + "_" + b.genData + "_input.json"
		fmt.Println("Generating test data file: " + fileName)
		err = ioutil.WriteFile(fileName, v, 0777)
		if err != nil {
			return fmt.Errorf("Failed to generate test data file due to error - %s", err.Error())
		}
		fmt.Println("Test data file successfully created at " + fileName)

	} else if b.serverPort != "" {
		//// Start REST Service
		//settings := make(map[string]string, 1)
		//settings["port"] = b.serverPort
		//serviceConfig := &support.ServiceConfig{
		//	Name:     "FlowTester",
		//	Enabled:  true,
		//	Settings: settings,
		//}
		//apiServer := tester.NewRestEngineTester(serviceConfig)
		//if apiServer == nil {
		//	return fmt.Errorf("Failed to create Test Engine")
		//}
		//
		////Start API Server
		//err := apiServer.Start()
		//if err != nil {
		//	return fmt.Errorf("Failed to start Test Engine due to error - %s", err.Error())
		//}
		//
		//fmt.Println("Engine API Server started")
		//apiList := "APIs:\n" +
		//	"   Start Flow : POST http://localhost:" + b.serverPort + "/flow/start"
		//fmt.Println(apiList)
		//
		//exitChan := setupSignalHandling()
		//fmt.Println("Engine API Server stopped")
		//code := <-exitChan
		//if code == 0 {
		//	return nil
		//}
		//
		//return nil

	} else if b.testInput != "" {

		testdataFile, err := ioutil.ReadFile(b.
			testInput)
		if err != nil {
			return fmt.Errorf("Failed to read test data file due to error - %s", err.Error())
		}
		inData := tester.StartRequest{}
		err = json.Unmarshal(testdataFile, &inData)
		if err != nil {
			fmt.Println("Invalid input")
			b.PrintUsage()
			return nil
		}

		//e.App().
		factory := action.GetFactory(RefFlow)
		factory.Initialize(e.App())
		act, _ := factory.New(&action.Config{Settings: map[string]interface{}{"flowURI": inData.FlowURI}})

		inputs := make(map[string]interface{})

		if len(inData.Data) > 0 {
			log.RootLogger().Debugf("Starting with flow attrs: %#v", inData.Data)
			for k, v := range inData.Data {
				inputs[k] = v
			}
		} else {
			inputs = make(map[string]interface{}, 1)
		}

		execOptions := &instance.ExecOptions{Interceptor: inData.Interceptor, Patch: inData.Patch}
		ro := &instance.RunOptions{Op: instance.OpStart, ReturnID: false, FlowURI: inData.FlowURI, ExecOptions: execOptions}
		inputs["_run_options"] = ro

		run := runner.NewDirect()
		results, err := run.RunAction(context.Background(), act, inputs)
		if err != nil {
			return fmt.Errorf("Flow [%s] execution failed due to error - %s", getFlowName(inData.FlowURI), err.Error())
		}

		fmt.Println("Flow execution successful")

		reply := make(map[string]interface{}, len(results))
		for k, v := range results {
			reply[k] = v
		}

		v, err := json.MarshalIndent(reply, "", "    ")
		if err != nil {
			return fmt.Errorf("Failed to generate flow output data due to error - %s", err.Error())
		}

		if b.testOutput != "" {
			err = ioutil.WriteFile(b.testOutput, v, 0777)
			if err != nil {
				return fmt.Errorf("Failed to generate test data file due to error - %s", err.Error())
			}
		} else {
			fmt.Println(string(v))
		}
	} else {
		// Print Usage
		b.PrintUsage()
	}
	return nil
}

func getFlowName(flowURI string) string {
	if strings.HasPrefix(flowURI, "res://flow:") {
		return flowURI[11:]
	}
	return flowURI
}

func getValue(dataType data.Type) interface{} {

	switch dataType {
	case data.TypeAny:
		return make(map[string]string)
	case data.TypeString:
		return ""
	case data.TypeInt:
		return 0
	case data.TypeFloat64:
		return 0
	case data.TypeInt64:
		return 0.0
	case data.TypeBool:
		return false
	case data.TypeObject:
		return make(map[string]string)
	case data.TypeArray:
		return make([]map[string]interface{}, 0)
	case data.TypeParams:
		return make([]map[string]string, 0)
	}

	return ""
}

func setupSignalHandling() chan int {

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int, 1)
	select {
	case s := <-signalChan:
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitChan <- 0
		default:
			log.RootLogger().Debug("Unknown signal.")
			exitChan <- 1
		}
	}
	return exitChan
}
