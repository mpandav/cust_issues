/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

// Package callprocedure provides Execution for stored procedure by invoking them
// flogo  TDV Connector
package callprocedure

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/alexbrainman/odbc/api"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/flogo-tdv/src/app/TDV/cgoutil/execute"
	"github.com/tibco/flogo-tdv/src/app/TDV/connector/connection"
	"github.com/tibco/wi-contrib/engine/jsonschema"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func init() {

	_ = activity.Register(&Activity{}, New)
}

// Activity is the structure for Activity Metadata
type Activity struct {
	mu                     *sync.RWMutex
	connMgr                *connection.TDVSharedConfigManager
	StatementHandle        api.SQLHSTMT
	isStatementInitialized bool
	isMetadataInitialized  bool
	callStatement          string
	inputParamMetadata     []execute.InputParamMetadata
	inputParamPositions    []int
	cursors                []string
	paramPosition          int
	parameterInfo          []execute.Parameter
}

// New for Activity
func New(ctx activity.InitContext) (activity.Activity, error) {
	// logCache := ctx.Logger()
	var statementHandle api.SQLHSTMT
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	connMgr, err := coerce.ToConnection(s.Connection)
	if err != nil {
		return nil, err
	}
	conn := connMgr.GetConnection().(*connection.TDVSharedConfigManager)

	return &Activity{connMgr: conn, StatementHandle: statementHandle, isMetadataInitialized: false, isStatementInitialized: false, mu: &sync.RWMutex{}}, nil
}

// Metadata  returns TDV's Query activity's meteadata
func (*Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval handles the data processing in the activity
func (a *Activity) Eval(context activity.Context) (done bool, err error) {
	logCache := context.Logger()
	logCache.Info("Executing TDV Call procedure Activity")
	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	//If Activity is running for first time Initialize the Necessary Input
	if !a.isMetadataInitialized {
		a.mu.Lock()
		defer a.mu.Unlock()
		a.callStatement, a.inputParamMetadata, a.inputParamPositions, a.cursors, a.paramPosition, err = input.GenerateQuery()
		if err != nil {
			return false, fmt.Errorf("could not generate query : %v", err)
		}
		a.isMetadataInitialized = true
	}

	//Connection--
	sharedmanager := a.connMgr

	if sharedmanager.GetCgoConnection() == nil {
		err = sharedmanager.SetCgoConnection()
		//If Set SQL Connection fails
		if err != nil {
			return false, err
		}
		a.isStatementInitialized = false

	} else if !sharedmanager.IsCGOConnectionAlive() {
		//If connection is not alive Release the Statement Handle and Allocate new
		connection.ReleaseStatmentHandle(a.StatementHandle)
		sharedmanager.ReleaseCgoConnection()
		logCache.Info("Recovering the Connection")
		err = sharedmanager.SetCgoConnection()
		//If Set SQL Connection fails
		if err != nil {
			return false, err
		}
		a.isStatementInitialized = false
	}
	if !a.isStatementInitialized {
		//TODO :
		a.StatementHandle, err = sharedmanager.GetPreparedStatement(a.callStatement)
		if err != nil {
			return false, fmt.Errorf("Prepairing Statment failed : %s", err.Error())
		}
		a.parameterInfo, err = execute.ExtractInputParameters(a.StatementHandle, a.inputParamPositions, logCache)
		if err != nil {
			//defer releaseHandle(stmt)
			return false, fmt.Errorf("Failed to extract Input parameters : %s", err.Error())
		}
	}

	logCache.Debug("Callstatement for given Procedure is : ", a.callStatement)

	inputParams, err := getInputData(input.InputParams, logCache)
	if err != nil {
		return false, fmt.Errorf("failed to read input arguments: %s", err.Error())
	}

	result, err := sharedmanager.ExecuteProcedure(a.StatementHandle, a.callStatement, inputParams, a.inputParamMetadata, a.inputParamPositions, a.paramPosition-1, a.cursors, a.parameterInfo, logCache)
	if err != nil {
		if err == driver.ErrBadConn || strings.Contains(strings.ToLower(strings.ToLower(err.Error())), "connection refused") || strings.Contains(strings.ToLower(err.Error()), "network is unreachable") ||
			strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") || strings.Contains(strings.ToLower(err.Error()), "dial tcp: lookup") ||
			strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "timedout") ||
			strings.Contains(strings.ToLower(err.Error()), "request timed out") || strings.Contains(strings.ToLower(err.Error()), "timed out") || strings.Contains(strings.ToLower(err.Error()), "net.Error") || strings.Contains(strings.ToLower(err.Error()), "i/o timeout") || strings.Contains(strings.ToLower(err.Error()), "connection is closed") || strings.Contains(strings.ToLower(err.Error()), "invalid cursor") {

			return false, activity.NewRetriableError(fmt.Sprintf("Failed to execute query [%s] on TDV server due to error - {%s}.", a.callStatement, err.Error()), "tdv-query-4001", nil)
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
