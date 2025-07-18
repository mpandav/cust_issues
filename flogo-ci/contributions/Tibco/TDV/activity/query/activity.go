/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

// Package query provides database query execution implementation for
// flogo  TDV Connector
package query

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/flogo-tdv/src/app/TDV/connector/connection"
	"github.com/tibco/wi-contrib/engine/jsonschema"
)

// TDV Connection Properties
/***const (
	ConnectionProp    = "Connection"
	InputProp         = "input"
	QueryProperty     = "Query"
	QueryNameProperty = "QueryName"
	FieldsProperty    = "Fields"
	OutputProperty    = "Output"
	RecordsProperty   = "records"
)***/

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

// Metadata  returns TDV's Query activity's meteadata
func (*Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval handles the data processing in the activity
func (a *Activity) Eval(context activity.Context) (done bool, err error) {
	logCache := context.Logger()
	logCache.Info("Executing TDV Query Activity")
	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	sharedmanager := input.Connection.(*connection.TDVSharedConfigManager)
	if sharedmanager.IsSQLConnectionNil() {
		err = sharedmanager.SetSQLConnection()
		//If Set SQL Connection fails
		if err != nil {
			return false, err
		}
	} else if !sharedmanager.IsSQLConnectionAlive() {
		logCache.Info("Recovering the Connection")
		err = sharedmanager.SetSQLConnection()
		//If Set SQL Connection fails
		if err != nil {
			return false, err
		}
	}
	inputParamDatatypes := []string{}
	for _, v := range input.Fields {
		m, ok := v.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("error fetching metadata: %v", err)
		}
		if m["Parameter"].(bool) {
			inputParamDatatypes = append(inputParamDatatypes, m["Type"].(string))
		}
	}
	query := input.Query
	if query == "" {
		return false, fmt.Errorf("missing schema SQL statement")
	}

	inputParams, err := getInputData(input.InputParams, logCache)
	if err != nil {
		return false, fmt.Errorf("failed to read input arguments: %s", err.Error())
	}

	result, err := sharedmanager.PreparedQuery(query, inputParams, inputParamDatatypes, logCache)

	if err != nil {
		if err == driver.ErrBadConn || strings.Contains(strings.ToLower(strings.ToLower(err.Error())), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
			strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") || strings.Contains(strings.ToLower(err.Error()), "dial tcp: lookup") ||
			strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "timedout") ||
			strings.Contains(strings.ToLower(err.Error()), "request timed out") || strings.Contains(strings.ToLower(err.Error()), "timed out") || strings.Contains(strings.ToLower(err.Error()), "net.Error") || strings.Contains(strings.ToLower(err.Error()), "i/o timeout") || strings.Contains(strings.ToLower(err.Error()), "connection is closed") || strings.Contains(strings.ToLower(err.Error()), "invalid cursor") {

			return false, activity.NewRetriableError(fmt.Sprintf("Failed to execute query [%s] on TDV server due to error - {%s}.", query, err.Error()), "tdv-query-4001", nil)
		}
		return false, fmt.Errorf("query execution failed: %s", err.Error())
	}

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

func getInputData(inputData interface{}, logCache log.Logger) (inputParams *connection.Input, err error) {
	inputParams = &connection.Input{}

	if inputData == nil {
		return nil, fmt.Errorf("missing input arguments")
	}

	switch inputData.(type) {
	case string:
		logCache.Debugf("Input data content: %s", inputData.(string))
		tempMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(inputData.(string)), &tempMap)
		if err != nil {
			return nil, fmt.Errorf("string parameter read error: %s", err.Error())
		}
		inputParams.Parameters = tempMap
	default:
		dataBytes, err := json.Marshal(inputData)
		logCache.Debugf("input arguments data: %s", string(dataBytes))
		if err != nil {
			return nil, fmt.Errorf("input data read failed: %s", err.Error())
		}
		err = json.Unmarshal(dataBytes, inputParams)
		if err != nil {
			return nil, fmt.Errorf("complex parameters read error:, %s", err.Error())
		}
	}
	return
}
