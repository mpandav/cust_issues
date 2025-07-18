package helper

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"

	//register oracle dialer
	godror "github.com/godror/godror"
)

type (
	//record is the one row of the ResultSet retrieved from database after execution of SQL Query
	record map[string]interface{}
	//ResultSet is an aggregation of SQL Query data records fetched from the database
	ResultSet struct {
		Record []*record `json:"records"`
	}

	//Input is a representation of activity's input parametres
	Input struct {
		Parameters map[string]interface{} `json:"parameters,omitempty"`
		//Values     map[string]interface{} `json:"values,omitempty"`
		Values []map[string]interface{} `json:"values,omitempty"`
	}
)

type FieldsInfoStruct struct {
	Fields []Param `json:"fields"`
	Ok     bool    `json:"ok"`
	Query  string  `json:"query"`
}
type Param struct {
	Direction  string `json:"Direction"`
	FieldName  string `json:"FieldName"`
	Type       string `json:"Type"`
	IsEditable bool   `json:"isEditable"`
}

// PreparedQuery allows querying database with named parameters
func PreparedQuery(db *sql.DB, queryString string, inputData *Input, log log.Logger) (results *ResultSet, err error) {
	log.Debugf(GetMessage(ExecutingQuery, queryString))

	//remove ';' from the query as the third party complains about it
	queryString = strings.TrimSuffix(strings.TrimSpace(queryString), ";")

	preparedQuery := queryString
	inputParams := inputData.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}
	log.Debugf(GetMessage(QueryWithParam, preparedQuery, inputArgs))

	statement, err := db.Prepare(preparedQuery)
	if err != nil {
		log.Warnf("query preparation failed: %s, %s", preparedQuery, err.Error())
		return nil, fmt.Errorf("query preparation failed: %s, %s", queryString, err.Error())
	}
	defer statement.Close()

	rows, err := statement.Query(inputArgs...)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(QueryFailed, err.Error()))
	}

	if rows == nil {
		log.Infof(GetMessage(NoRowsFound, preparedQuery))
		return nil, nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf(GetMessage(GettingInColumnInfo, err.Error()))
	}

	count := len(columns)
	cols := make([]interface{}, count)
	args := make([]interface{}, count)
	for i := range cols {
		args[i] = &cols[i]
	}

	coltypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf(GetMessage(GettingInColumnType, err.Error()))
	}

	var resultSet ResultSet
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return nil, fmt.Errorf(GetMessage(ScanningRow, err.Error()))
		}

		m := make(record)
		for i, b := range cols {
			dbType := coltypes[i].DatabaseTypeName()
			m[columns[i]] = b
			if b != nil {
				switch dbType {
				case "CHAR", "NCHAR", "NVARCHAR2", "LONG", "VARCHAR2":
					m[columns[i]] = b.(string)

				case "NUMBER", "FLOAT", "DOUBLE":
					var x godror.Number
					if reflect.TypeOf(b) == reflect.TypeOf(x) {
						if nx, ok := strconv.ParseFloat(string(b.(godror.Number)), 512); ok == nil {
							m[columns[i]] = nx
						}
					}

				case "DATE":
					if t, ok := b.(time.Time); ok {
						m[columns[i]] = t.Format("2006-01-02")
					}

				//Datatype "TIMESTAMP WITH TIMEZONE" do not need any explicit handling for now
				case "TIMESTAMP", "TIMESTAMP WITH LOCAL TIMEZONE":
					//Since we do not know precision used by user for the column type, we would display what comes from db.
					//We are removing any trailing zero's from the nano seconds part of time.
					if t, ok := b.(time.Time); ok {
						timeValueWithDefaultPrecision := t.Format("2006-01-02T15:04:05.000000000")
						timeValueSplitOnNanoSeconds := strings.Split(timeValueWithDefaultPrecision, ".")
						nanoSecondsString := timeValueSplitOnNanoSeconds[1]
						nanoSecondsString = strings.TrimRight(nanoSecondsString, "0")
						if !strings.EqualFold(nanoSecondsString, "") {
							/*nanoSecondsInteger, err := strconv.ParseInt(nanoSecondsString, 10, 64)
							if err != nil {
								fmt.Println(err)
								return nil, err
							}
							timeValueSplitOnNanoSeconds[0] = fmt.Sprintf("%s.%d", timeValueSplitOnNanoSeconds[0], nanoSecondsInteger)*/
							timeValueSplitOnNanoSeconds[0] = fmt.Sprintf("%s.%s", timeValueSplitOnNanoSeconds[0], nanoSecondsString)
							m[columns[i]] = timeValueSplitOnNanoSeconds[0]
						} else {
							m[columns[i]] = timeValueSplitOnNanoSeconds[0]
						}
					}

				case "INTERVAL YEAR TO MONTH":
					interval := b.(string)
					//commenting below code as godror version 0.9.0 returned interval year to month
					//in the form aybm where 'a' is year and 'b' is month and y and m are characters in
					//the result to signify year and month. In new godror version 0.15.0 it is returned
					//in form a-b where 'a' is year and 'b' is month. Hence we do not need to format.
					/*
						years := strings.Split(interval, "y")
						months := strings.Split(years[1], "m")
						m[columns[i]] = fmt.Sprintf("+%s-%s", years[0], months[0])
					*/
					m[columns[i]] = interval

				case "INTERVAL DAY TO SECOND":
					if t, ok := b.(time.Duration); ok {
						days := t / (24 * time.Hour)
						hours := t % (24 * time.Hour)
						minutes := hours % time.Hour
						seconds := math.Mod(minutes.Seconds(), 60)
						var secondsWithAccurateNanoseconds string
						timeValueSplitOnNanoSeconds := strings.Split(fmt.Sprintf("%.9f", seconds), ".")
						nanoSecondsString := timeValueSplitOnNanoSeconds[1]
						nanoSecondsString = strings.TrimRight(nanoSecondsString, "0")
						if !strings.EqualFold(nanoSecondsString, "") {
							secondsWithAccurateNanoseconds = fmt.Sprintf("%s.%s", timeValueSplitOnNanoSeconds[0], nanoSecondsString)
						} else {
							secondsWithAccurateNanoseconds = timeValueSplitOnNanoSeconds[0]
						}

						m[columns[i]] = fmt.Sprintf("+%d %d:%d:%s", days, hours/time.Hour, minutes/time.Minute, secondsWithAccurateNanoseconds)
					}
				}

				/*if x, ok := b.([]byte); ok == true {
					switch dbType {
					//case "BINARY_INTEGER", "INT", "MEDIUMINT", "SMALLINT", "TINYINT":
					//if nx, ok := strconv.ParseUint(string(x), 10, 64); ok == nil {
					//	m[columns[i]] = nx
					//}
					case "NUMBER", "FLOAT", "DOUBLE":
						if nx, ok := strconv.ParseFloat(string(x), 512); ok == nil {
							m[columns[i]] = nx
						}
					//case "BOOLEAN":
					//if nx, ok := strconv.ParseBool(string(x)); ok == nil {
					//	m[columns[i]] = nx
					//}
					case "TIMESTAMP WITH TIMEZONE", "TIMESTAMP", "TIMESTAMP WITH LOCAL TIMEZONE":
						//reference time: Mon Jan 2 15:04:05 -0700 MST 2006
						if nx, ok := time.Parse("2006-1-2 15:04:05", string(x)); ok == nil {
							m[columns[i]] = nx
						}
					case "INTERVAL YEAR TO MONTH":
						//reference time: Mon Jan 2 15:04:05 -0700 MST 2006
						if nx, ok := time.Parse("2006", string(x)); ok == nil {
							m[columns[i]] = nx
						}
					case "DATE":
						//reference time: Mon Jan 2 15:04:05 -0700 MST 2006
						if nx, ok := time.Parse("2006-1-2", string(x)); ok == nil {
							m[columns[i]] = nx
						}
					case "TIME":
						//reference time: Mon Jan 2 15:04:05 -0700 MST 2006
						if nx, ok := time.Parse("15:04:05", string(x)); ok == nil {
							m[columns[i]] = nx
						}
					case "CHAR", "NCHAR", "NVARCHAR2", "LONG", "VARCHAR2":
						m[columns[i]] = string(x)
					default:
						m[columns[i]] = b
					}
				}*/
			}
		}
		if len(m) > 0 {
			resultSet.Record = append(resultSet.Record, &m)
		}
	}
	if len(resultSet.Record) == 0 {
		log.Infof(GetMessage(NoRowsFound, preparedQuery))
	}
	return &resultSet, nil
}

func PreparedQueryCALL(db *sql.DB, queryString string, inputData []map[string]interface{}, log log.Logger) (output map[string]interface{}, err error) {
	log.Debugf(GetMessage(ExecutingQuery, queryString))

	//remove ';' from the query as the third party complains about it
	tx, err := db.Begin()
	if err != nil {

		return
	}
	cursorkeyvalues := []string{}
	outputKeyValues := []string{}
	outputMap := make(map[string]interface{})
	cursorOutputMap := make(map[string]interface{})
	queryString = strings.TrimSuffix(strings.TrimSpace(queryString), ";")
	preparedQuery := queryString
	inputParams := make(map[string]interface{})
	param := make(map[string]interface{})
	outputValues := make([]string, len(inputData))
	rset := make([]driver.Rows, len(inputData))
	outputCount := 0
	cursorCount := 0
	// Iterating inputdata for prepare statement
	if len(inputData) > 0 {
		for _, name := range inputData {
			temp1 := name["Direction"].(string)
			if strings.ToUpper(temp1) == "IN" {
				param[name["FieldName"].(string)] = name["Value"]
			} else if strings.ToUpper(temp1) == "OUT" && strings.ToUpper(name["Type"].(string)) != "REFCURSOR" {
				// normal out parameter
				param[name["FieldName"].(string)] = sql.Out{Dest: &outputValues[outputCount]}
				outputMap[name["FieldName"].(string)] = ""
				outputKeyValues = append(outputKeyValues, name["FieldName"].(string))
				outputCount++
			} else if strings.ToUpper(temp1) == "OUT" && strings.ToUpper(name["Type"].(string)) == "REFCURSOR" {
				// cursor out parameter
				param[name["FieldName"].(string)] = sql.Out{Dest: &rset[cursorCount]}
				cursorkeyvalues = append(cursorkeyvalues, name["FieldName"].(string))
				cursorCount++
			} else if strings.ToUpper(temp1) == "INOUT" {
				//set input param value in the variable
				//add output param
				outputValues[outputCount] = name["Value"].(string)
				param[name["FieldName"].(string)] = sql.Out{Dest: &outputValues[outputCount], In: true}
				outputMap[name["FieldName"].(string)] = ""
				outputKeyValues = append(outputKeyValues, name["FieldName"].(string))
				outputCount++
			}
		}
	}

	resultset := [][]string{}
	inputParams["parameters"] = param
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)
	inputArgs, preparedQuery, err = getQueryAndArgs(param, queryString)
	log.Debugf(GetMessage(Preparedquery, preparedQuery))
	statement, _ := db.Prepare(preparedQuery)
	tx.Stmt(statement).Exec(inputArgs...)
	if cursorCount > 0 {
		for i := 0; i < cursorCount; i++ {
			// iterating over rset and scanning values
			col := rset[i].(driver.RowsColumnTypeScanType).Columns()
			dests1 := make([]driver.Value, len(col))
			tempResultset := []string{}
			for {
				if err := rset[i].Next(dests1); err != nil {
					if err == io.EOF {
						break
					}
					rset[i].Close()
					return nil, err
				}
				val, _ := coerce.ToString(dests1)
				tempResultset = append(tempResultset, val)

			}
			resultset = append(resultset, tempResultset)

		}
		index := 0
		for i := 0; i < len(cursorkeyvalues); i++ {
			cursorOutputMap[cursorkeyvalues[i]] = resultset[index]
			index++
		}

	} else {
		index := 0
		for i := 0; i < len(outputKeyValues); i++ {
			outputMap[outputKeyValues[i]] = outputValues[index]
			index++
		}
	}

	if err = tx.Commit(); err != nil {
		log.Debugf(GetMessage(TxcommitError, err))

	}

	for k, v := range cursorOutputMap {
		outputMap[k] = v
	}
	log.Debugf(GetMessage(OutputDataContent, outputMap))
	return outputMap, nil
}

func getQueryAndArgs(inputParams map[string]interface{}, query string) (args []interface{}, preparedQuery string, err error) {
	regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
	var inputArgs []interface{}
	preparedQuery = query

	regExp := regexp.MustCompile("\\?\\w*")
	matches := regExp.FindAllStringSubmatch(preparedQuery, -1)

	i := 0

	for _, match := range matches {
		parameter := strings.Split(match[0], "?")[1]
		substitution, ok := inputParams[parameter]
		if !ok {
			return nil, query, fmt.Errorf("missing substitution for: %s", match[0])
		}
		preparedQuery = strings.Replace(preparedQuery, match[0], ":"+strconv.Itoa(i+1), 1)

		strValue, err := coerce.ToString(substitution)
		if err != nil {
			return nil, query, fmt.Errorf(GetMessage(DefaultError, err.Error()))
		}

		if regexDate.MatchString(strValue) == true {
			//Date
			t, err := time.Parse("2006-01-02+00:00", strValue)
			if err != nil {
				return nil, query, fmt.Errorf(GetMessage(DefaultError, err.Error()))
			}
			//DD-MMM-YY
			strValue = t.Format("02-Jan-06")
		}

		i++
		inputArgs = append(inputArgs, substitution)
	}

	return inputArgs, preparedQuery, nil
}

// PreparedUpdateOrDelete allows querying Snowflake with parameters
func PreparedUpdateOrDelete(db *sql.DB, queryString string, inputData *Input, log log.Logger) (results map[string]interface{}, err error) {
	log.Debugf(GetMessage(ExecutingQuery, queryString))

	//remove ';' from the query as the third party complains about it
	queryString = strings.TrimSuffix(strings.TrimSpace(queryString), ";")

	preparedQuery := queryString
	inputParams := inputData.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}
	log.Debugf(GetMessage(QueryWithParam, preparedQuery, inputArgs))

	result, err := db.Exec(preparedQuery, inputArgs...)
	if err != nil {
		log.Warnf("Query execution failed: %s, %s", preparedQuery, err.Error())
		return nil, fmt.Errorf("Query execution failed: %s, %s", queryString, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}

//Below method contains logic to support insert query (and not INSERT ALL)
/*func (con *Connection) PreparedInsert(queryString string, inputData *Input, log logger.Logger) (results map[string]interface{}, err error) {
	log.Debugf(GetMessage(ExecutingQuery, queryString))

	//remove ';' from the query as the third party complains about it
	queryString = strings.TrimSuffix(strings.TrimSpace(queryString), ";")

	//preparedQuery := queryString

	inputValues := inputData.Values
	inputParams := inputData.Parameters
	totalInput := make(map[string]interface{})

	for k, v := range inputValues {
		totalInput[k] = v
	}

	for k, v := range inputParams {
		totalInput[k] = v
	}

	inputArgs, preparedQuery, err := getQueryAndArgs(totalInput, queryString, log)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}
	log.Debugf(GetMessage(QueryWithParam, preparedQuery, inputArgs))

	result, err := con.DB.Exec(preparedQuery, inputArgs...)
	if err != nil {
		log.Warnf("Query execution failed: %s, %s", preparedQuery, err.Error())
		return nil, fmt.Errorf("Query execution failed: %s, %s", queryString, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}*/

func getValue(key string, value interface{}, insertQueryWithValues string) (insertQueryWithReplacedValues string, err error) {
	var strValue string
	switch reflect.ValueOf(value).Kind() {
	case reflect.Float64:
		strValue, err = coerce.ToString(value)
		if err != nil {
			return "", err
		}
		break

	case reflect.String:
		strValue, err = coerce.ToString(value)
		if err != nil {
			return "", err
		}

		//Support date time format
		regexDateTime := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
		regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")

		if regexDateTime.MatchString(strValue) == true {
			//DateTime
			t, err := time.Parse(time.RFC3339, strValue)
			if err != nil {
				return "", err
			}
			//DD-MMM-YY HH:mm:ss AM/PM
			strValue = t.Format("02-JAN-06 03:04:05 PM")
		} else if regexDate.MatchString(strValue) == true {
			//Date
			t, err := time.Parse("2006-01-02+00:00", strValue)
			if err != nil {
				return "", err
			}
			//DD-MMM-YY
			strValue = t.Format("02-Jan-06")
		}

		strValue = "'" + strValue + "'"
		break

	default:
		strValue, err = coerce.ToString(value)
		if err != nil {
			return "", err
		}
		strValue = "'" + strValue + "'"
		break
	}
	insertQueryWithValues = strings.Replace(insertQueryWithValues, "?"+key, strValue, 1)
	return insertQueryWithValues, nil
}

// PreparedInsert allows inserting into database with select clause and also allows insert all clause
func PreparedInsert(db *sql.DB, queryString string, inputData *Input, log log.Logger) (results map[string]interface{}, err error) {
	log.Debugf(GetMessage(ExecutingQuery, queryString))

	//remove ';' from the query as the third party complains about it
	queryString = strings.TrimSuffix(strings.TrimSpace(queryString), ";")

	preparedQuery := queryString
	inputParams := inputData.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}
	log.Debugf(GetMessage(QueryWithParam, preparedQuery, inputArgs))

	result, err := db.Exec(preparedQuery, inputArgs...)
	if err != nil {
		log.Warnf("Query execution failed: %s, %s", preparedQuery, err.Error())
		return nil, fmt.Errorf("Query execution failed: %s, %s", queryString, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}

// GetInputData converts the input to helper.Input
func GetInputData(inputData interface{}, log log.Logger) (inputParams *Input, err error) {
	inputParams = &Input{}

	if inputData == nil {
		return nil, fmt.Errorf(GetMessage(SpecifyInput))
	}

	switch inputData.(type) {
	case string:
		log.Debug(GetMessage(InputDataContent, inputData.(string)))
		tempMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(inputData.(string)), &tempMap)
		if err != nil {
			return nil, fmt.Errorf(GetMessage(Unmarshall, err.Error()))
		}
		inputParams.Parameters = tempMap
	default:
		dataBytes, err := json.Marshal(inputData)
		if err != nil {
			return nil, fmt.Errorf(GetMessage(Deserialize, err.Error()))
		}
		log.Debug(GetMessage(InputDataContent, string(dataBytes)))
		err = json.Unmarshal(dataBytes, inputParams)
		if err != nil {
			return nil, fmt.Errorf(GetMessage(Unmarshall, err.Error()))
		}
	}
	return
}

func GetInputDataCall(inputData interface{}, inputData2 string, log log.Logger) (inputParam []map[string]interface{}, err error) {

	inputDataNewadd := &Input{}

	bytes := []byte(inputData2)

	// Unmarshal string into structs.
	var params FieldsInfoStruct
	err1 := json.Unmarshal(bytes, &params)
	if err1 != nil {
		return nil, fmt.Errorf(GetMessage(Unmarshall, err.Error()))
	}

	fieldinfo := params.Fields

	listofmap := make([]map[string]interface{}, len(fieldinfo))

	dataBytes, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(Deserialize, err.Error()))
	}

	err = json.Unmarshal(dataBytes, inputDataNewadd)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(Unmarshall, err.Error()))
	}

	input := inputDataNewadd.Parameters

	for i, data := range fieldinfo {
		listofmap[i] = make(map[string]interface{})
		listofmap[i]["Direction"] = data.Direction
		listofmap[i]["FieldName"] = data.FieldName
		listofmap[i]["Type"] = data.Type
		listofmap[i]["Value"] = "0"

	}

	for i, mapval := range listofmap {
		if mapval["Direction"] == "IN" || mapval["Direction"] == "INOUT" {
			for inputkey, inputvalue := range input {
				if inputkey == mapval["FieldName"] {
					mapval["Value"] = inputvalue
				}
			}
		}

		listofmap[i] = mapval

	}

	return listofmap, nil
}

// PreparedInsertWithValues allows inserting into database with values clause
func PreparedInsertWithValues(db *sql.DB, queryString string, inputData *Input, log log.Logger) (results map[string]interface{}, err error) {
	//log.Debugf(GetMessage(ExecutingQuery, queryString))
	//remove ';' from the query as the third party complains about it
	queryString = strings.TrimSuffix(strings.TrimSpace(queryString), ";")
	insertIntoStmt := strings.SplitN(queryString, " ", 2) //remove insert keyword

	inputObjVal := inputData.Values
	inputParamsData := inputData.Parameters
	insertAllQuery := "INSERT ALL"
	insertQueryWithValues := ""
	selectDual := " SELECT * FROM dual"

	for key := range inputObjVal {
		insertQueryWithValue := " " + insertIntoStmt[1]
		inputObjData := inputObjVal[key]
		inputObj, err := coerce.ToObject(inputObjData)
		if err != nil {
			return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
		}

		if len(inputObj) > 0 {

			regExp := regexp.MustCompile("\\?\\w*")
			matches := regExp.FindAllStringSubmatch(insertQueryWithValue, -1)

			for _, match := range matches {
				parameter := strings.Split(match[0], "?")[1]
				substitution, ok := inputObj[parameter]
				if !ok {
					substitution, ok = inputParamsData[parameter]
					if !ok {
						return nil, fmt.Errorf("missing substitution for: %s", match[0])
					}
				}
				insertQueryWithValue, err = getValue(parameter, substitution, insertQueryWithValue)
				if err != nil {
					return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
				}
			}
		}
		insertQueryWithValues += insertQueryWithValue
	}

	inputParams, err := coerce.ToObject(inputParamsData)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}

	if len(inputObjVal) == 0 {
		if len(inputParams) > 0 {
			//if query has all params and no value at all
			if insertQueryWithValues == "" {
				insertQueryWithValues = " " + insertIntoStmt[1]
			}

			//replace params in entire query

			regExp := regexp.MustCompile("\\?\\w*")
			matches := regExp.FindAllStringSubmatch(insertQueryWithValues, -1)

			for _, match := range matches {
				parameter := strings.Split(match[0], "?")[1]
				substitution, ok := inputParams[parameter]
				if !ok {
					substitution, ok = inputParamsData[parameter]
					if !ok {
						return nil, fmt.Errorf("missing substitution for: %s", match[0])
					}
				}
				insertQueryWithValues, err = getValue(parameter, substitution, insertQueryWithValues)
				if err != nil {
					return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
				}
			}
		}
	}

	//if query has no value no params, then simply copy such string
	if insertQueryWithValues == "" {
		insertQueryWithValues = " " + insertIntoStmt[1]
	}

	query := insertAllQuery + insertQueryWithValues + selectDual
	log.Debugf("Executing Query: %s", query)

	result, err := db.Exec(query)
	if err != nil {
		log.Warnf("Query execution failed: %s, %s", query, err.Error())
		return nil, fmt.Errorf("Query execution failed: %s, %s", query, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}
