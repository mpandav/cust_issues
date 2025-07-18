package query

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/jsonschema"
	"github.com/tibco/wi-mysql/src/app/MySQL/connector/connection"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// MysqlQueryActivity is the structure for Activity Metadata
type MysqlQueryActivity struct {
	settings *Settings
}

func init() {
	_ = activity.Register(&MysqlQueryActivity{}, New)
}

// New constructor for creating a New MySQL activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &MysqlQueryActivity{}, nil
}

// Metadata returns MySQL's Query activity's meteadata
func (a *MysqlQueryActivity) Metadata() *activity.Metadata {
	return activityMd
}

//GetComplexValue safely get the object value
// func GetComplexValue(complexObject *data.ComplexObject) interface{} {
// 	if complexObject != nil {
// 		return complexObject.Value
// 	}
// 	return nil
// }

// Eval handles the data processing in the activity
func (a *MysqlQueryActivity) Eval(context activity.Context) (done bool, err error) {
	logger := context.Logger()
	logger.Info("Executing MySQL Query Activity")
	actinputs := &Input{}
	err = context.GetInputObject(actinputs)
	if err != nil {
		return false, fmt.Errorf("MySQL connection is not configured")
	}

	sharedmanager := actinputs.Connection.(*connection.SharedConfigManager)

	if actinputs.Query == "" {
		return false, fmt.Errorf("No Runtime SQL Query specified")
	}

	//extract input parameters, refactor into a method later
	inputParams, err := getInputData(actinputs.Input, logger)
	if err != nil {
		return false, fmt.Errorf("Failed to process input arguments, %s", err.Error())
	}

	result, err := sharedmanager.PreparedQuery(actinputs.Query, inputParams, logger)
	if err != nil {
		// logger.Debugf("^^^^^^ Error occurred while connecting to the MySQL server: %s", err.Error())
		if err == driver.ErrBadConn || err == mysql.ErrInvalidConn || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "network is unreachable") ||
			strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "dial tcp: lookup") ||
			strings.Contains(err.Error(), "connection timed out") || strings.Contains(err.Error(), "timedout") || strings.Contains(err.Error(), "time out") ||
			strings.Contains(err.Error(), "timed out") || strings.Contains(err.Error(), "net.Error") || strings.Contains(err.Error(), "i/o timeout") {

			return false, activity.NewRetriableError(fmt.Sprintf("Failed to execute query [%s] on MySQL server due to error - {%s}.", actinputs.Query, err.Error()), "mysql-query-4001", nil)
		}
		return false, fmt.Errorf("query execution failed: %s", err.Error())
	}

	//json schema validation
	outputSchema := actinputs.Fields
	if outputSchema != nil {
		jsonBytes, err := json.Marshal(outputSchema)
		schema := string(jsonBytes)
		err = jsonschema.ValidateFromObject(schema, result)
		if err != nil {
			return false, fmt.Errorf("Schema validation error %s", err.Error())
		}
	}
	actoutput := &Output{}
	actoutput.Output = result
	err = context.SetOutputObject(actoutput)
	if err != nil {
		return false, err
	}

	return true, nil
}

func withOutEOLs(queryString string) (flattened string, err error) {
	regex, err := regexp.Compile("\n")
	if err != nil {
		return
	}

	queryString = regex.ReplaceAllString(queryString, " ")
	return
}

func getInputData(inputData interface{}, logger log.Logger) (inputParams *connection.Input, err error) {
	inputParams = &connection.Input{}
	if inputData == nil {

		inputData = []byte(`{
			"parameters": {
			}
		}`)
		//return nil, fmt.Errorf("No input arguments for query specified")
	}

	switch inputData.(type) {
	case string:
		logger.Debugf("Input data content: %s", inputData.(string))
		tempMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(inputData.(string)), &tempMap)
		if err != nil {
			return nil, fmt.Errorf("Cannot unmarshall data into parameters, %s", err.Error())
		}
		inputParams.Parameters = tempMap
	default:
		dataBytes, err := json.Marshal(inputData)
		logger.Debugf("Input arguments data: %s", string(dataBytes))
		if err != nil {
			return nil, fmt.Errorf("Cannot deserialize inputData: %s", err.Error())
		}
		err = json.Unmarshal(dataBytes, inputParams)
		if err != nil {
			return nil, fmt.Errorf("Cannot convert deserialized data into input parameters, %s", err.Error())
		}
	}
	return
}
