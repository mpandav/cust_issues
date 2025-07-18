package common

import (
	"archive/zip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/howeyc/gopass"
	"github.com/magiconair/properties"
	"github.com/project-flogo/core/app"
	"github.com/project-flogo/core/app/propertyresolver"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"

	conAws "github.com/tibco/wi-contrib/connection/aws"
)

var defaultResolver = resolve.NewCompositeResolver(map[string]resolve.Resolver{
	".":        &resolve.ScopeResolver{},
	"env":      &resolve.EnvResolver{},
	"property": &property.Resolver{},
})

// EnvAppPropertyOverrideKey ...
const EnvAppPropertyJSON = "FLOGO_APP_PROPS_JSON"

const (
	ExecutableName      = "bootstrap"
	LambdaRuntime       = "provided.al2023"
	DEFAULT_CREDENTIALS = "Default Credentials"
)

var lambdaLog = log.ChildLogger(log.RootLogger(), "lambda-common")

// DeployFromCommandLine ...
func DeployFromCommandLine(target string, awsSecretKey string, appjson string, ref string, envConfig map[string]string) error {
	if target != "" {
		awsCon, err := GetAWSConnectionInfo(appjson, ref)
		if err != nil {
			log.RootLogger().Errorf("Get connection details failed, due to [%s]", err.Error())
			return fmt.Errorf("Get connection details failed, due to [%s]", err.Error())
		}

		if target == "lambda" {

			if awsCon.GetAuthenticationType() != DEFAULT_CREDENTIALS {
				key := awsSecretKey
				if key == "" {
					// Prompt and let user type secret.
					key = PasswordPrompt("Please type aws access key")
					if key == "" {
						fmt.Println("No aws access key typed")
						return fmt.Errorf("No aws access key typed")
					}
				}

				if key != awsCon.GetSecretKey() {
					fmt.Println("Mismatch aws secret access key, please retype secret access key")
					return fmt.Errorf("Mismatch aws secret access key, please retype secret access key")
				}
			}
			err = deployLambdaFunction(awsCon, appjson, ref, envConfig)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Unknow deploy target ", target)
			return fmt.Errorf("Unknow deploy target ")
		}

	}
	return nil
}

// Deploy ...
func Deploy(appjson string, ref string, envConfig map[string]string) error {
	awsCon, err := GetAWSConnectionInfo(appjson, ref)
	if err != nil {
		log.RootLogger().Errorf("Get connection details failed, due to [%s]", err.Error())
		return fmt.Errorf("Get connection details failed, due to [%s]", err.Error())
	}
	err = deployLambdaFunction(awsCon, appjson, ref, envConfig)
	if err != nil {
		return err
	}
	return nil
}

//func getAWSConnectionInfo(cfgJson string) (*session.Session, error) {
//	conn, err := getTriggerConnection(cfgJson)
//	if err != nil {
//		lambdaLog.Error(err.Error())
//		return nil, err
//	}
//	return conn, nil
//}

func deployLambdaFunction(awscon *conAws.Connection, appJSON string, ref string, envConfig map[string]string) error {
	lambdaLog.Info("Start to deploy lambda function..")
	flogoApp, err := engine.LoadAppConfig(appJSON, false)
	if err != nil {
		lambdaLog.Errorf("Get trigger from app failed, due to [%s]", err.Error())
		return fmt.Errorf("Get trigger from app failed, due to [%s]", err.Error())
	}

	functionName := GetFunctionName(flogoApp.Name)
	lambdaLog.Infof("Get Trigger info done, Starting create lambda function [%s]", functionName)

	sess := awscon.NewSession()

	execDir, err := MakeFunctionHandlerZip(ExecutableName)
	if err != nil {
		lambdaLog.Error(err)
		return err
	}

	// Set X-RAY tracing
	value, found := os.LookupEnv("FLOGO_AWS_XRAY_ENABLE")
	if found {
		envConfig["FLOGO_AWS_XRAY_ENABLE"] = value
	}

	result, err := CreateLambdaFunction(sess, path.Join(execDir, "handler.zip"), appJSON, flogoApp, functionName, ref, envConfig)
	if err != nil {
		lambdaLog.Errorf("Create lambda function failed, due to [%s]", err.Error())
		return fmt.Errorf("Create lambda function failed, due to [%s]", err.Error())
	}
	if result != nil {
		lambdaLog.Infof("Complete function creation [%s], the ARN is ===> %s", functionName, *result.FunctionArn)
	}

	if IsTCIEnv() {
		lambdaLog.Infof("Starting monitor lambda function [%s] logs", functionName)
		go CheckLog(10*time.Second, functionName, sess)

		//if isTCIEnv() {
		//	lambdaLog.Infof("Delete lambda function: %s", app.Name)
		//	if err := deleteLambdaFunction(sess, app.Name); err != nil {
		//		lambdaLog.Error(err)
		//	}
		//}

	}
	return nil
}

// MakeFunctionHandlerZip ...
func MakeFunctionHandlerZip(appName string) (string, error) {
	execDir, exeFile, err := getCurrectPath()
	if err != nil {
		return execDir, fmt.Errorf("Get executable dir failed, due to [%s]", err.Error())
	}

	if exeFile != appName {
		//Cp another file with app name
		// Read all content of src to data
		data, err := ioutil.ReadFile(exeFile)
		if err != nil {
			return execDir, err
		}
		// Write data to dst
		err = ioutil.WriteFile(appName, data, 0644)
		if err != nil {
			return execDir, err
		}
		if err := os.Chmod(appName, 0755); err != nil {
			return execDir, err
		}
	}
	defer func() {
		os.Remove(appName)
	}()

	lambdaLog.Infof("Currect dir %s, and executable file %s", execDir, exeFile)
	//For env var to override app property

	key := os.Getenv(EnvAppPropertyJSON)
	var zipFiles []string

	if fileExist(filepath.Join(execDir, "flogo.json")) {
		zipFiles = []string{appName, path.Join(execDir, "plugins"), filepath.Join(execDir, "flogo.json")}
	} else {
		zipFiles = []string{appName, path.Join(execDir, "plugins")}
	}

	if key != "" && strings.HasSuffix(key, ".json") {
		//Copy the file to current dir and make it into zip.
		if fileExist(key) {
			zipFiles = append(zipFiles, key)
		}
	}

	err = ZipFiles(zipFiles, path.Join(execDir, "handler.zip"))
	if err != nil {
		return execDir, fmt.Errorf("Zip executable file failed, due to [%s]", err.Error())
	}

	return execDir, nil
}

func getCurrectPath() (string, string, error) {
	//ex, err := os.Executable()
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", "", err
	}
	//exPath := filepath.Dir(ex)
	file := os.Args[0]
	file = file[2:] //remove ./ at beginning
	return path, file, nil
}

// CreateLambdaFunction ...
func CreateLambdaFunction(sess *session.Session, zipFile string, appjson string, app *app.Config, funcName string, ref string, envConfig map[string]string) (*lambda.FunctionConfiguration, error) {

	roleName, err := getExeRoleName(appjson, ref)
	if err != nil {
		return nil, err
	}
	handlesZip, err := ioutil.ReadFile(zipFile)
	if err != nil {
		return nil, err
	}

	var roleARN string
	if strings.TrimSpace(roleName) == "" {
		roleName = "TIBCO_FlogoLambdaExecRole"
	}

	if strings.HasPrefix(strings.ToLower(roleName), "arn:") {
		roleARN = roleName
	} else {
		//Create New role
		lambdaLog.Infof("Execution role name [%s]", roleName)
		lambdaLog.Infof("Creating role name [%s]", roleName)
		role, err := createRole(roleName, sess)
		if err != nil {
			return nil, err
		}
		roleARN = role
	}
	lambdaLog.Infof("Role ARN [%s]", roleARN)

	description := app.Description
	if description == "" {
		description = "Function created by FLOGO"
	}
	input := &lambda.CreateFunctionInput{
		Code:         &lambda.FunctionCode{ZipFile: handlesZip},
		Description:  aws.String(description),
		FunctionName: aws.String(funcName),
		Handler:      aws.String(ExecutableName),
		Role:         aws.String(roleARN),
		Runtime:      aws.String(LambdaRuntime),
		Timeout:      aws.Int64(30),
	}

	svc := lambda.New(sess)
	envs := make(map[string]*string)
	key := os.Getenv(EnvAppPropertyJSON)
	if key != "" && strings.HasSuffix(key, ".json") {
		envs[EnvAppPropertyJSON] = aws.String(filepath.Base(key))
	} else if key != "" {
		envs[EnvAppPropertyJSON] = aws.String(key)
	}

	sum, err := getAppMd5Sum(app)
	if err != nil {
		return nil, err
	}
	if sum != "" {
		lambdaLog.Debugf("MDF %s", sum)
		envs["FLOGO_APP_MD5"] = aws.String(sum)
		lenv := lambda.Environment{
			Variables: updateEnv(envs, envConfig),
		}
		input.SetEnvironment(&lenv)
	}

	result, err := svc.CreateFunction(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				lambdaLog.Errorf(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				lambdaLog.Errorf(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				lambdaLog.Errorf(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeResourceConflictException:
				lambdaLog.Infof("Function with name %s already exists", funcName)
				//Check if env MD5 sum same, ignore if it is same
				fun, errget := GetFunctionConfiguration(funcName, svc)
				if errget != nil {
					return nil, errget
				}

				if roleARN != *fun.Role || len(envConfig) > 0 {
					//Update Role
					env := make(map[string]*string)
					if ok {
						env = fun.Environment.Variables
					}
					_, err = UpdateEnv(svc, funcName, updateEnv(env, envConfig), roleARN)
				}

				md, md5Err := getAppMd5Sum(app)
				if md5Err != nil {
					return nil, md5Err
				}

				funSum, ok := fun.Environment.Variables["FLOGO_APP_MD5"]
				if ok {
					if *funSum == string(md[:]) {
						lambdaLog.Infof("Function with name %s already exists and version matched", funcName)
						return nil, nil
					}
				}

				lambdaLog.Infof("Function with name %s already exists and version mismatched, update it...", funcName)
				//update function code

				result, err = UpdateLambdaFunction(svc, funcName, handlesZip)
				if err != nil {

				}
				env := make(map[string]*string)
				if ok {
					env = fun.Environment.Variables
				}
				env["FLOGO_APP_MD5"] = aws.String(md)

				key := os.Getenv(EnvAppPropertyJSON)
				if key != "" && strings.HasSuffix(key, ".json") {
					env[EnvAppPropertyJSON] = aws.String(filepath.Base(key))
				} else if key != "" {
					env[EnvAppPropertyJSON] = aws.String(key)
				}

				if !hasPropertyOverride() {
					//Delete app property override property
					delete(env, EnvAppPropertyJSON)
				}
				_, err = UpdateEnv(svc, funcName, updateEnv(env, envConfig), "")
			case lambda.ErrCodeTooManyRequestsException:
				lambdaLog.Errorf(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeCodeStorageExceededException:
				lambdaLog.Errorf(lambda.ErrCodeCodeStorageExceededException, aerr.Error())
			default:
				lambdaLog.Errorf(aerr.Error())
			}
		} else {
			lambdaLog.Errorf(err.Error())
		}
		return result, err
	}
	//no need to invoke
	//return InvoleFunction(*result.FunctionName, svc)
	return result, nil
}

func updateEnv(envs map[string]*string, envConfig map[string]string) map[string]*string {

	if len(envConfig) > 0 {
		if _, ok := envConfig[engine.EnvAppPropertyResolvers]; !ok {
			envs[engine.EnvAppPropertyResolvers] = aws.String("env")
			envs[propertyresolver.EnvAppPropertyEnvConfigKey] = aws.String("auto")
		}
		for k, v := range envConfig {
			envs[k] = aws.String(v)
		}
	}
	return envs
}

func deleteLambdaFunction(sess *session.Session, functionName string) error {
	svc := lambda.New(sess)
	input := &lambda.DeleteFunctionInput{FunctionName: aws.String(functionName)}
	_, err := svc.DeleteFunction(input)
	if err != nil {
		return fmt.Errorf("Delete function [%s] failed, due to [%s]", functionName, err.Error())
	}
	return nil
}

// UpdateEnv ...
func UpdateEnv(svc *lambda.Lambda, functionName string, envs map[string]*string, role string) (*lambda.FunctionConfiguration, error) {

	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	}
	updatedEnvs := make(map[string]*string)
	for k, v := range envs {
		updatedEnvs[k] = v
	}
	input.SetEnvironment(&lambda.Environment{Variables: updatedEnvs})
	if role != "" {
		input.SetRole(role)
	}
	result, err := svc.UpdateFunctionConfiguration(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				lambdaLog.Errorf(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				lambdaLog.Errorf(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				lambdaLog.Errorf(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeTooManyRequestsException:
				lambdaLog.Errorf(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeCodeStorageExceededException:
				lambdaLog.Errorf(lambda.ErrCodeCodeStorageExceededException, aerr.Error())
			default:
				lambdaLog.Errorf(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			lambdaLog.Errorf(err.Error())
		}

		return result, err
	}

	return result, nil
}

// UpdateLambdaFunction ...
func UpdateLambdaFunction(svc *lambda.Lambda, funcName string, zip []byte) (*lambda.FunctionConfiguration, error) {
	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(funcName),
		Publish:      aws.Bool(true),
		ZipFile:      zip,
	}

	result, err := svc.UpdateFunctionCode(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeServiceException:
				lambdaLog.Errorf(lambda.ErrCodeServiceException, aerr.Error())
			case lambda.ErrCodeResourceNotFoundException:
				lambdaLog.Errorf(lambda.ErrCodeResourceNotFoundException, aerr.Error())
			case lambda.ErrCodeInvalidParameterValueException:
				lambdaLog.Errorf(lambda.ErrCodeInvalidParameterValueException, aerr.Error())
			case lambda.ErrCodeTooManyRequestsException:
				lambdaLog.Errorf(lambda.ErrCodeTooManyRequestsException, aerr.Error())
			case lambda.ErrCodeCodeStorageExceededException:
				lambdaLog.Errorf(lambda.ErrCodeCodeStorageExceededException, aerr.Error())
			default:
				lambdaLog.Errorf(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			lambdaLog.Errorf(err.Error())
		}

		return result, err
	}

	return result, nil
}

// InvoleFunction ...
func InvoleFunction(functionName string, svc *lambda.Lambda) error {
	invokeInput := &lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: []byte("{}"), InvocationType: aws.String("Event"), LogType: aws.String("Tail")}
	_, err := svc.Invoke(invokeInput)
	if err != nil {
		return err
	}
	return nil
}

// ZipFiles ...
func ZipFiles(sources []string, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for _, source := range sources {
		info, err := os.Stat(source)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}

		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(source)
		}

		filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		})
	}

	return err
}

func createRole(name string, sess *session.Session) (string, error) {
	role := `{
	 "Version": "2012-10-17",
	 "Statement": [{
	   "Sid": "",
	   "Effect": "Allow",
	   "Principal": {
	     "Service": "lambda.amazonaws.com"
	   },
	   "Action": "sts:AssumeRole"
	 }]
	}`

	input := &iam.CreateRoleInput{RoleName: aws.String(name), AssumeRolePolicyDocument: aws.String(role)}
	svc := iam.New(sess)
	output, err := svc.CreateRole(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeEntityAlreadyExistsException:
				output, err := svc.GetRole(&iam.GetRoleInput{RoleName: aws.String(name)})
				if err != nil {
					return "", err
				}
				return *output.Role.Arn, nil
			default:
				fmt.Println(aerr.Error())
			}
		}
		return "", err
	}

	<-time.After(10 * time.Second)
	attachedPolicy := &iam.AttachRolePolicyInput{PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"), RoleName: aws.String(name)}
	output3, err3 := svc.AttachRolePolicy(attachedPolicy)
	if err3 != nil {
		return "", err3
	}
	fmt.Println(output3.String())
	fmt.Println(*output.Role.Arn)
	return *output.Role.Arn, nil
}

// CheckLog ...
func CheckLog(d time.Duration, functionName string, sess *session.Session) {
	startTime := time.Now()
	for t := range time.Tick(d) {
		if PrintLogsToConsole(functionName, nil, sess, aws.TimeUnixMilli(startTime.UTC()), aws.TimeUnixMilli(t.UTC())) {
			startTime = t
		}
	}
}

// PrintLogsToConsole ...
func PrintLogsToConsole(functionName string, token *string, sess *session.Session, startTime, endtime int64) bool {
	svc := cloudwatchlogs.New(sess)

	logs, err := SearchLogs(svc, "/aws/lambda/"+functionName, "", startTime, endtime)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudwatchlogs.ErrCodeResourceNotFoundException:
				//Ignore just keep watching
			default:
				lambdaLog.Errorf(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			lambdaLog.Errorf(err.Error())
		}
	}

	for _, l := range logs {
		if len(l) > 0 && len(strings.TrimSpace(l)) > 0 {
			fmt.Println(strings.TrimSpace(l))
		}
	}
	return len(logs) > 1
}

// SearchLogs ...
func SearchLogs(svc *cloudwatchlogs.CloudWatchLogs, group, stream string, start, end int64) ([]string, error) {
	var logs []string

	streams, err := listStreams(svc, group, stream)
	if err != nil {
		return logs, err
	}

	for _, s := range streams {
		if s.LastEventTimestamp != nil && *s.LastIngestionTime < start {
			continue
		}

		newLogs, err := GetEventLogs(svc, group, *s.LogStreamName, start, end)
		if err != nil {
			return logs, err
		}

		if len(newLogs) > 0 {
			logs = append(logs, newLogs...)
		}
	}

	return logs, nil
}

// Helper function to get Log Streams (with Token support).
func listStreams(svc *cloudwatchlogs.CloudWatchLogs, group, prefix string) ([]*cloudwatchlogs.LogStream, error) {
	var streams []*cloudwatchlogs.LogStream

	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
		//LogStreamNamePrefix: aws.String(prefix),
		Descending: aws.Bool(true),
	}

	for {
		resp, err := svc.DescribeLogStreams(params)
		if err != nil {
			return streams, err
		}

		for _, s := range resp.LogStreams {
			streams = append(streams, s)
		}

		if resp.NextToken == nil {
			return streams, nil
		}

		params.NextToken = resp.NextToken
	}
}

// GetEventLogs ...
func GetEventLogs(svc *cloudwatchlogs.CloudWatchLogs, group, stream string, from, to int64) ([]string, error) {
	var logs []string

	params := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		StartTime:     aws.Int64(from),
		EndTime:       aws.Int64(to),
	}

	for {
		resp, err := svc.GetLogEvents(params)
		if err != nil {
			return logs, err
		}

		if len(resp.Events) < 1 {
			return logs, nil
		}

		for _, e := range resp.Events {
			if *e.Message != "" {
				logs = append(logs, ParseToTCILog(strings.TrimSpace(*e.Message)))
			}
		}

		if resp.NextForwardToken == nil {
			return logs, nil
		}

		params.NextToken = resp.NextForwardToken
	}
}

// ParseToTCILog ...
func ParseToTCILog(msg string) string {
	//2018-05-08 19:59:00.443 DEBUG [basic-mapper] - Updated mapping def &{Type:3 Value:dddddddddd MapTo:['message']}
	logParts := strings.Split(msg, " ")
	if len(logParts) > 3 {
		level := logParts[2]
		return fmt.Sprintf("%s %-6s [%s] - %s\n", time.Now().Format(GetLogDateTimeFormat()), getLogLevel(log.ToLogLevel(strings.TrimSpace(level))), "flogo-lambda", msg)
	}
	return fmt.Sprintf("%s %-6s [%s] - %s\n", time.Now().Format(GetLogDateTimeFormat()), "DEBUG", "flogo-lambda", msg)
}

func getLogLevel(level log.Level) string {
	switch level {
	case log.DebugLevel:
		return "DEBUG"
	case log.InfoLevel:
		return "INFO"
	case log.WarnLevel:
		return "WARN"
	case log.ErrorLevel:
		return "ERROR"
	default:
		return "DEBUG"
	}
}

// GetFunctionConfiguration ...
func GetFunctionConfiguration(functionName string, svc *lambda.Lambda) (*lambda.FunctionConfiguration, error) {
	out, err := svc.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{FunctionName: aws.String(functionName)})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func getExeRoleName(appjson string, ref string) (string, error) {
	tri, err := GetTrigger(appjson, ref)
	if err != nil {
		return "", fmt.Errorf("Can't get trigger due to %s ", err.Error())
	}

	role := tri.Settings["ExecutionRoleName"]
	if role == nil {
		return "", nil
	}
	roleName, err := coerce.ToString(role)
	if err != nil {
		lambdaLog.Warnf("Can't coerce exec role to string, %s", err.Error())
		return "", nil
	}

	if roleName != "" && roleName[0] == '=' {
		roleName = roleName[1:]
		resolvedRole, err := defaultResolver.Resolve(roleName, nil)
		if err != nil {
			return "", fmt.Errorf("resolve app property [%s] failed, due to %s", roleName, err.Error())
		}
		return coerce.ToString(resolvedRole)
	}
	return roleName, nil
}

func fileExist(props string) bool {
	_, err := os.Stat(props)
	return !os.IsNotExist(err)
}

// IsTCIEnv ...
func IsTCIEnv() bool {
	_, ok := os.LookupEnv("TIBCO_INTERNAL_TCI_SUBSCRIPTION_ID")
	return ok
}

// GetFunctionName ...
func GetFunctionName(appName string) string {
	if IsTCIEnv() {
		appID := os.Getenv("TIBCO_INTERNAL_SERVICE_NAME")
		return appName + "-" + appID
	}
	return appName
}

// PasswordPrompt ...
func PasswordPrompt(prompt string) string {
	fmt.Printf(prompt + ":")
	password, _ := gopass.GetPasswdMasked()
	s := string(password[0:])
	return s
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return
}

// GetAWSConnectionInfo ...
func GetAWSConnectionInfo(appJSON string, triggerRef string) (*conAws.Connection, error) {
	conn, err := getTriggerConnection(appJSON, triggerRef)
	if err != nil {
		lambdaLog.Error(err.Error())
		return nil, err
	}
	return conn, nil
}

func getTriggerConnection(appJSON string, ref string) (*conAws.Connection, error) {
	tri, err := GetTrigger(appJSON, ref)
	if err != nil {
		return nil, fmt.Errorf("Can't get trigger due to %s ", err.Error())
	}
	var conn *conAws.Connection
	connRef, ok := tri.Settings["ConnectionName"].(string)
	if ok {
		conn, err = resolveConnString(connRef, appJSON)
	} else {
		conn, err = conAws.NewConnection(tri.Settings["ConnectionName"])
	}
	if err != nil {
		return nil, fmt.Errorf("Can't get connection due to %s ", err.Error())
	}
	return conn, nil
}

// GetTrigger ...
func GetTrigger(appJSON string, ref string) (*trigger.Config, error) {

	flogoApp, err := engine.LoadAppConfig(appJSON, false)
	if err != nil {
		return nil, err
	}
	if len(flogoApp.Triggers) <= 0 {
		return nil, fmt.Errorf("No trigger found in the app")
	}

	for _, tri := range flogoApp.Triggers {
		triRef := tri.Ref
		if strings.HasPrefix(triRef, "#") {
			triRef, _ = GetRef(GetAllImports(flogoApp.Imports), triRef)
		}
		if strings.EqualFold(triRef, ref) {
			return tri, nil
		}
	}

	return nil, fmt.Errorf("Not dound trigger with ref [%s]", ref)
}

func resolveConnString(t string, appJSON string) (*conAws.Connection, error) {
	var id string
	if strings.HasPrefix(t, "conn://") {
		id = t[7:]
	} else {
		return nil, fmt.Errorf("Unable to resolve connection string: %+v", id)
	}
	flogoApp, err := engine.LoadAppConfig(appJSON, false)
	if err != nil {
		return nil, err
	}
	connConfig := flogoApp.Connections[id]
	if connConfig == nil {
		return nil, fmt.Errorf("Unable to find connection object with id: %+v", id)
	}
	connObj := getConnObjectFromConfig(*connConfig, id)
	conn, err := conAws.NewConnection(connObj)
	return conn, err
}

func getConnObjectFromConfig(config connection.Config, id string) map[string]interface{} {
	m := make(map[string]interface{})
	m["id"] = id
	m["name"] = config.Settings["name"].(string)
	m["ref"] = config.Ref
	settings, _ := resolveTriggerSettings(config.Settings)
	m["settings"] = settings
	return m
}

func resolveTriggerSettings(settings map[string]interface{}) ([]interface{}, error) {
	var settingsArray []interface{}
	for k, v := range settings {
		finalVal, err := resolveProperties(v)
		if err != nil {
			return nil, err
		}
		settingsArray = append(settingsArray, map[string]interface{}{"name": k, "value": finalVal})
	}
	return settingsArray, nil
}

func resolveProperties(val interface{}) (interface{}, error) {
	strVal, ok := val.(string)
	if ok && len(strVal) > 0 && strVal[0] == '=' {
		return defaultResolver.Resolve(strVal[1:], nil)
	}
	return val, nil
}

func getAppMd5Sum(app *app.Config) (string, error) {
	//Cache only flowmapping + flow + setting for connection change
	triggers := make([]*trigger.Config, len(app.Triggers))
	for index, trr := range app.Triggers {
		handlers := make([]*trigger.HandlerConfig, len(trr.Handlers))
		for i, h := range trr.Handlers {
			handlers[i] = &trigger.HandlerConfig{Actions: h.Actions}
		}
		triggers[index] = &trigger.Config{Handlers: handlers, Settings: trr.Settings}
	}

	newApp := struct {
		Resource    interface{}
		Trigger     interface{}
		AppProperty interface{}
	}{
		Resource: app.Resources,
		Trigger:  triggers,
	}

	key := os.Getenv(EnvAppPropertyJSON)
	if key != "" && strings.HasSuffix(key, ".json") {
		if fileExist(key) {
			appPro := make(map[string]interface{})
			b, err := ioutil.ReadFile(key)
			if err != nil {
				return "", err
			}
			err = json.Unmarshal(b, &appPro)
			if err != nil {
				return "", err
			}
			newApp.AppProperty = appPro
		}
	} else if key != "" {
		newApp.AppProperty = key
	}

	bys, err := json.Marshal(newApp)
	if err != nil {
		if err != nil {
			return "", err
		}
	}
	sum := fmt.Sprintf("%x", md5.Sum(bys))
	return sum, nil
}

// Import ...
type Import struct {
	Alias   string
	Import  string
	Version string
}

// GetAllImports ...
func GetAllImports(imports []string) []*Import {
	var importStruct []*Import

	for _, ref := range imports {
		if len(ref) > 0 {
			ref = strings.TrimSpace(ref)
			var alias string
			var version string

			if strings.Index(ref, " ") > 0 {
				alias = strings.TrimSpace(ref[:strings.Index(ref, " ")])
				ref = strings.TrimSpace(ref[strings.Index(ref, " ")+1:])
			}

			if strings.Index(ref, "@") > 0 {
				version = ref[strings.Index(ref, "@")+1:]
				ref = ref[:strings.Index(ref, "@")]
			}
			imp := &Import{Alias: alias, Import: ref, Version: version}
			importStruct = append(importStruct, imp)
		}
	}
	return importStruct
}

// GetRef ...
func GetRef(imports []*Import, refStr string) (string, error) {
	if strings.HasPrefix(refStr, "#") {
		refStr = refStr[1:]
	}

	for _, im := range imports {
		if refStr == im.Alias {
			return im.Import, nil
		}
		ref := im.Import
		if refStr == filepath.Base(ref) {
			return ref, nil
		}
	}

	return "", fmt.Errorf("cannot found import ref [%s] from %+v", refStr, imports)
}

// GetEnvConfig ...
func GetEnvConfig(configPath string) (map[string]string, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Env-Config file [%s] not found", configPath)
	}
	var envMap map[string]string
	if strings.HasSuffix(configPath, ".properties") {
		p, err := properties.LoadFile(configPath, properties.UTF8)
		if err != nil {
			return nil, fmt.Errorf("Load property file [%s] error : %s", configPath, err.Error())
		}
		envMap = p.Map()
	} else if strings.HasSuffix(configPath, ".json") {
		var jsonMap map[string]interface{}
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("Load json file [%s] error : %s", configPath, err.Error())
		}

		err = json.Unmarshal(content, &jsonMap)
		if err != nil {
			return nil, fmt.Errorf("Load env config json [%s] error : %s", configPath, err.Error())
		}
		envMap, err = coerce.ToParams(jsonMap)
		if err != nil {
			return nil, fmt.Errorf("Convert [%+v] to key value string error : %s", jsonMap, err.Error())
		}
	}

	//ENv Key cannot have . convert it to underline to same with env property resolver
	newEnvMap := make(map[string]string)
	for k, v := range envMap {
		k = strings.Replace(k, ".", "_", -1)
		newEnvMap[k] = v
	}

	return newEnvMap, nil
}

// SetEnvVar ...
func SetEnvVar(envs map[string]string) error {
	for k, v := range envs {
		os.Setenv(k, v)
	}
	return nil
}

func hasPropertyOverride() bool {
	key := os.Getenv(EnvAppPropertyJSON)
	if len(key) > 0 {
		return true
	}
	return false
}

func GetLogDateTimeFormat() string {
	logLevelEnv := os.Getenv(log.EnvKeyLogDateFormat)
	if len(logLevelEnv) > 0 {
		return logLevelEnv
	}
	return log.DefaultLogDateFormat
}
