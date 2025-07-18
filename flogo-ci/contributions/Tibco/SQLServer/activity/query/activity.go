/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

//Package query provides database query execution implementation for
//flogo WI SqlServer Connector
package query

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/engine/jsonschema"
	connection "github.com/tibco/wi-mssql/src/app/SQLServer/connector/connection"
)

// const (
// 	connectionProp    = "Connection"
// 	databaseURL       = "datbaseURL"
// 	host              = "host"
// 	port              = "port"
// 	user              = "username"
// 	password          = "password"
// 	databaseName      = "databaseName"
// 	inputProp         = "input"
// 	activityOutput    = "output"
// 	queryProperty     = "Query"
// 	queryNameProperty = "QueryName"
// 	fieldsProperty    = "Fields"
// 	outputProperty    = "Output"
// 	recordsProperty   = "records"
// )

//Acvitity is the structure for Activity Metadata
// type Acvitity struct {
// 	metadata *activity.Metadata
// }

// //NewActivity constructor for creating a New SqlServer activity
// func NewActivity(metadata *activity.Metadata) activity.Activity {
// 	return &Acvitity{metadata: metadata}
// }

// //Metadata returns SqlServer's Query activity's meteadata
// func (a *Acvitity) Metadata() *activity.Metadata {
// 	return a.metadata
// }

// var log = logger.GetLogger("sqlserver-query")

// //GetComplexValue safely get the object value
// func GetComplexValue(complexObject *data.ComplexObject) interface{} {
// 	if complexObject != nil {
// 		return complexObject.Value
// 	}
// 	return nil
// }

// //Eval handles the data processing in the activity
// func (a *Acvitity) Eval(context activity.Context) (done bool, err error) {
// 	log.Debug("Executing SqlServer Query Activity")

// 	connector := context.GetInput(connectionProp)
// 	if connector == nil {
// 		return false, fmt.Errorf("SqlServer connection is not configured")
// 	}

// 	connection, err := sqlserver.GetConnection(connector, log)
// 	if err != nil {
// 		return false, fmt.Errorf("Error getting SqlServer connection %s", err.Error())
// 	}
// 	_, err = connection.Login(log)
// 	if err != nil {
// 		return false, fmt.Errorf("Cannot Login %s", err.Error())
// 	}

// 	log.Debugf("Read SqlServer's connection information")

// 	query := context.GetInput(queryProperty)
// 	if query == nil {
// 		return false, fmt.Errorf("No SQL Query specified")
// 	}

// 	//extract input parameters, refactor into a method later
// 	inputData := GetComplexValue(context.GetInput(inputProp).(*data.ComplexObject))
// 	inputParams, err := getInputData(inputData)
// 	if err != nil {
// 		return false, fmt.Errorf("Failed to process input arguments, %s", err.Error())
// 	}

// 	result, err := connection.PreparedQuery(query.(string), inputParams, log)
// 	if err != nil {
// 		return false, fmt.Errorf("Failed to execute SQL Query, %s", err.Error())
// 	}

// 	//json schema validation
// 	outputSchema := context.GetInput(fieldsProperty)
// 	if outputSchema != nil {
// 		jsonBytes, err := json.Marshal(outputSchema)
// 		schema := string(jsonBytes)
// 		err = jsonschema.ValidateFromObject(schema, result)
// 		if err != nil {
// 			return false, fmt.Errorf("Schema validation error %s", err.Error())
// 		}
// 	}

// 	outputComplex := &data.ComplexObject{Metadata: "", Value: result}
// 	log.Debugf("SQL Query Output complex object %+v", outputComplex)
// 	context.SetOutput(outputProperty, outputComplex)
// 	return true, nil
// }

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {

	_ = activity.Register(&Activity{}, New)
}

// Activity is the structure for Activity Metadata
type Activity struct {
}

// New for Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil
}

// Metadata  returns SQLServer's Query activity's meteadata
func (*Activity) Metadata() *activity.Metadata {
	return activityMd
}

//Eval handles the data processing in the activity
func (a *Activity) Eval(context activity.Context) (done bool, err error) {
	logger := context.Logger()
	logger.Info("Executing SQLServer Query Activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, fmt.Errorf("Cannot Login %s", err.Error())
	}
	sharedmanager := input.Connection.(*connection.SharedConfigManager)

	query := input.Query
	if query == "" {
		return false, fmt.Errorf("No SQL Query specified")
	}

	queryTimeout := input.QueryTimeout
	if queryTimeout < 0 {
		return false, fmt.Errorf("Query timeout cannot be a negative number")
	}

	if queryTimeout == 0 {
		logger.Debugf("Query timeout is set to 0 (unlimited)")
	} else {
		logger.Debugf("Query timeout set to %d secs", queryTimeout)
	}

	inputParams, err := getInputData(input.InputParams, logger)
	if err != nil {
		return false, fmt.Errorf("Failed to process input arguments, %s", err.Error())
	}

	result, err := sharedmanager.PreparedQuery(query, inputParams, logger, queryTimeout)
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if err == driver.ErrBadConn || strings.Contains(lowerErr, "connection refused") || strings.Contains(lowerErr, "network is unreachable") ||
			strings.Contains(lowerErr, "connection reset by peer") || strings.Contains(lowerErr, "dial tcp: lookup") ||
			strings.Contains(lowerErr, "connection timed out") || strings.Contains(lowerErr, "timedout") || strings.Contains(lowerErr, "time out") ||
			strings.Contains(lowerErr, "timed out") || strings.Contains(lowerErr, "net.Error") || strings.Contains(lowerErr, "i/o timeout") ||
			strings.Contains(lowerErr, "no such host") || strings.Contains(lowerErr, "broken pipe") {

			return false, activity.NewRetriableError(fmt.Sprintf("Failed to execute query [%s] on SQLServer server due to error - {%s}.", query, err.Error()), "sqlserver-query-4001", nil)
		}
		return false, fmt.Errorf("Failed to execute SQL Query, %s", err.Error())
	}
	//json schema validation
	outputSchema := input.Fields
	if outputSchema != nil {
		jsonBytes, err := json.Marshal(outputSchema)
		schema := string(jsonBytes)
		err = jsonschema.ValidateFromObject(schema, result)
		if err != nil {
			return false, fmt.Errorf("Schema validation error %s", err.Error())
		}
	}
	output := &Output{}
	output.Output = result
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	return true, nil
}

func getInputData(inputData interface{}, logger log.Logger) (inputParams *connection.Input, err error) {

	inputParams = &connection.Input{}

	if inputData == nil {
		return nil, fmt.Errorf("No input arguments for query specified")
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

func withOutEOLs(queryString string) (flattened string, err error) {
	regex, err := regexp.Compile("\n")
	if err != nil {
		return
	}

	queryString = regex.ReplaceAllString(queryString, " ")
	return
}

// func getInputData(inputData interface{}) (inputParams *sqlserver.Input, err error) {

// 	inputParams = &sqlserver.Input{}

// 	if inputData == nil {
// 		return nil, fmt.Errorf("No input arguments for query specified")
// 	}

// 	switch inputData.(type) {
// 	case string:
// 		logCache.Debugf("Input data content: %s", inputData.(string))
// 		tempMap := make(map[string]interface{})
// 		err := json.Unmarshal([]byte(inputData.(string)), &tempMap)
// 		if err != nil {
// 			return nil, fmt.Errorf("Cannot unmarshall data into parameters, %s", err.Error())
// 		}
// 		inputParams.Parameters = tempMap
// 	default:
// 		dataBytes, err := json.Marshal(inputData)
// 		logCache.Debugf("Input arguments data: %s", string(dataBytes))
// 		if err != nil {
// 			return nil, fmt.Errorf("Cannot deserialize inputData: %s", err.Error())
// 		}
// 		err = json.Unmarshal(dataBytes, inputParams)
// 		if err != nil {
// 			return nil, fmt.Errorf("Cannot convert deserialized data into input parameters, %s", err.Error())
// 		}
// 	}
// 	return
// }
