package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	lm "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/project-flogo/core/support/log"
	feeEngine "github.com/tibco/wi-contrib/engine"
	"github.com/tibco/wi-contrib/engine/cmd"
	"github.com/tibco/wi-contrib/engine/starter"
	common "github.com/tibco/wi-plugins/contributions/flogo-lambda/src/app/Lambda/trigger"
	fl "github.com/tibco/wi-plugins/contributions/flogo-lambda/src/app/Lambda/trigger/lambda"
)

const REF = "github.com/tibco/wi-plugins/contributions/flogo-lambda/src/app/Lambda/trigger/lambda"

func init() {
	starter.RegistryShimStarter("Lambda", &LambdaRunner{})
}

type LambdaRunner struct {
}

func (r *LambdaRunner) Init(args []string) error {
	return nil
}

func (r *LambdaRunner) Run(args []string) (int, error) {
	return Start()
}

func init() {
	cmd.Registry("deploy", &lambdaCommand{})
	cmd.Registry("-deploy", &lambdaCommand{})
	cmd.Registry("--deploy", &lambdaCommand{})

}

type lambdaCommand struct {
	awsSecretKey  string
	target        string
	envConfigFile string
	fs            *flag.FlagSet
}

func (b *lambdaCommand) Name() string {
	return "-deploy"
}

func (b *lambdaCommand) Description() string {
	return "Deploy Lambda function"
}

func (b *lambdaCommand) Run(args []string, appJson string) error {
	b.target = args[0]
	if err := b.fs.Parse(args[1:]); err != nil {
		return err
	}

	envMap := make(map[string]string)
	if len(b.envConfigFile) > 0 {
		var err error
		envMap, err = common.GetEnvConfig(b.envConfigFile)
		if err != nil {
			return err
		}
		if err := common.SetEnvVar(envMap); err != nil {
			return err
		}
	}
	//Create engine here for load properties and resolve connections
	_, err := feeEngine.CreateEngine(appJson)
	if err != nil {
		return err
	}
	return common.DeployFromCommandLine(b.target, b.awsSecretKey, appJson, REF, envMap)
}

func (b *lambdaCommand) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.awsSecretKey, "aws-access-key", "", "AWS secret access key")
	fs.StringVar(&b.envConfigFile, "env-config", "", "Environment variables uses to Lambda execution env")
	b.fs = fs
}

func (b *lambdaCommand) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("	executable [command]")
	fmt.Println("Commands:")
	fmt.Println("	-h							help")
	fmt.Println("	--deploy <lambda> 			deploy target")
	fmt.Println("	--aws-access-key <key>      AWS secret access key")
}

func (b *lambdaCommand) IsShimCommand() bool {
	return true
}

func (b *lambdaCommand) Parse() bool {
	return false
}

// Handle implements the Flogo Function handler
func Handle(ctx context.Context, evt json.RawMessage) (interface{}, error) {
	err := setupArgs(evt, &ctx)
	if err != nil {
		return nil, err
	}

	result, err := fl.Invoke()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func setupArgs(evt json.RawMessage, ctx *context.Context) error {
	// Setup environment argument
	evtJSON, err := json.Marshal(&evt)
	if err != nil {
		return err
	}

	evtFlag := flag.Lookup("evt")
	if evtFlag == nil {
		flag.String("evt", string(evtJSON), "Lambda Environment Arguments")
	} else {
		flag.Set("evt", string(evtJSON))
	}

	// Setup context argument
	ctxObj, _ := lambdacontext.FromContext(*ctx)
	// lambdaContext := map[string]interface{}{
	// 	"logStreamName":   lambdacontext.LogStreamName,
	// 	"logGroupName":    lambdacontext.LogGroupName,
	// 	"functionName":    lambdacontext.FunctionName,
	// 	"functionVersion": lambdacontext.FunctionVersion,
	// 	"awsRequestId":    ctxObj.AwsRequestID,
	// 	"memoryLimitInMB": lambdacontext.MemoryLimitInMB,
	// }

	eventContext := EventContext{
		Function: LFunction{
			LogGroup:  lambdacontext.LogGroupName,
			LogStream: lambdacontext.LogStreamName,
			Name:      lambdacontext.FunctionName,
			Version:   lambdacontext.FunctionVersion,
		},
		Identity:  ctxObj.Identity,
		ClientApp: ctxObj.ClientContext.Client,
		Context: LContext{
			AwsRequestID:   ctxObj.AwsRequestID,
			ARN:            ctxObj.InvokedFunctionArn,
			TracingContext: ctxObj.ClientContext.Custom,
		},
	}

	ctxJSON, err := json.Marshal(eventContext)
	if err != nil {
		return err
	}

	ctxFlag := flag.Lookup("ctx")
	if ctxFlag == nil {
		flag.String("ctx", string(ctxJSON), "Lambda Context Arguments")
	} else {
		flag.Set("ctx", string(ctxJSON))
	}

	return nil
}

type EventContext struct {
	Function  LFunction
	Identity  lambdacontext.CognitoIdentity
	ClientApp lambdacontext.ClientApplication
	Context   LContext
}

type LFunction struct {
	LogGroup  string
	LogStream string
	Name      string
	Version   string
}

type LContext struct {
	AwsRequestID   string
	ARN            string
	TracingContext map[string]string
}

func Start() (int, error) {
	// aws lambda reversed env
	runMode := os.Getenv("LAMBDA_TASK_ROOT")
	if runMode != "" {
		log.RootLogger().Info("In lambda env, try to start lambda")
		lm.Start(Handle)
	} else {
		if common.IsTCIEnv() {
			log.RootLogger().Info("In TCI env, starting lambda")
			return starter.WAIT_FOR_TERMINATE, common.Deploy(cfgJson, REF, make(map[string]string))
			log.RootLogger().Info("Created Lambda function")

		} else {
			log.RootLogger().Info("Nothing to do with lambda trigger.  please use deploy command to deploy lambda function")
			return starter.TERMINATE, nil
		}
	}
	return starter.WAIT_FOR_TERMINATE, nil
}
