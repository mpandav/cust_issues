package snowflake

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"
)

type (
	//record is the one row of the ResultSet retrieved from database after execution of SQL Query
	record map[string]interface{}

	//ResultSet is an aggregation of SQL Query data records fetched from the database
	ResultSet struct {
		Record []*record `json:"records"`
	}

	//InputOffset is to hold key and value of the input parameters along with offset
	InputOffset struct {
		offset   int
		keyfield string
		value    interface{}
	}

	//InputOffsets will hold multiple input params (offset, key and value)
	InputOffsets []InputOffset

	//Parameters would hold InputArrayType
	Parameters map[string]interface{}

	//Input is a representation of acitivity's input parametres
	Input struct {
		Parameters map[string]interface{}   `json:"parameters,omitempty"`
		Values     []map[string]interface{} `json:"values,omitempty"`
	}

	CommandInput struct {
		FileName           string      `json:"fileName,omitempty"`
		InternalStage      string      `json:"internalStage,omitempty"`
		Parallel           int64       `json:"parallel,omitempty"`
		Auto_Compress      bool        `json:"auto_compress,omitempty"`
		Source_Compression string      `json:"source_compression,omitempty"`
		Overwrite          bool        `json:"overwrite,omitempty"`
		TableName          string      `json:"tableName,omitempty"`
		StageName          string      `json:"stageName,omitempty"`
		Files              string      `json:"files,omitempty"`
		Pattern            string      `json:"pattern,omitempty"`
		FileType           string      `json:"type,omitempty"`
		Format_Name        string      `json:"format_Name,omitempty"`
		CopyOptions        CopyOptions `json:"copyOptions"`
		Validation_Mode    string      `json:"validation_mode,omitempty"`
		FormatTypeOptions  string      `json:"formatTypeOptions,omitempty"`
	}

	CopyOptions struct {
		On_error             string `json:"on_error,omitempty"`
		Size_Limit           int64  `json:"size_limit,omitempty"`
		Purge                bool   `json:"purge,omitempty"`
		Return_Failed_Only   bool   `json:"return_failed_only,omitempty"`
		Match_By_Column_Name string `json:"match_by_column_name,omitempty"`
		Enforce_Length       bool   `json:"enforce_length,omitempty"`
		TruncateColumns      bool   `json:"truncatecolumns,omitempty"`
		Force                bool   `json:"force,omitempty"`
		Load_Uncertain_Files bool   `json:"load_uncertain_files,omitempty"`
	}
)

func (o InputOffsets) Len() int           { return len(o) }
func (o InputOffsets) Less(i, j int) bool { return o[i].offset < o[j].offset }
func (o InputOffsets) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }

func generateInputWithOffset(inputParams map[string]interface{}, queryString string) (input InputOffsets) {
	queryStringCopy := queryString
	for keyfield, value := range inputParams {

		for strings.Index(queryStringCopy, "?"+keyfield) > 0 {
			offset := strings.Index(queryStringCopy, "?"+keyfield)
			io := InputOffset{offset, keyfield, value}
			input = append(input, io)
			queryStringCopy = strings.Replace(queryStringCopy, "?"+keyfield, "?", 1)
		}
	}
	return
}

// QueryHelper struct and methods on it are referred from mysql
// QueryHelper is a simple query parser for extraction of values from an insert statement
type QueryHelper struct {
	sqlString   string
	values      []string
	first       string
	last        string
	valuesToken string
}

// NewQueryHelper creates a new instance of QueryHelper
func NewQueryHelper(sql string) *QueryHelper {
	qh := &QueryHelper{
		sqlString:   sql,
		valuesToken: "VALUES",
	}
	return qh
}

// Compose reconstitutes the query and returns it
func (qp *QueryHelper) Compose() string {
	return qp.first + qp.valuesToken + " " + strings.Join(qp.values, ", ") + " " + qp.last
}

// ComposeWithValues reconstitutes the query with external values
func (qp *QueryHelper) ComposeWithValues(values []string) string {
	return qp.first + qp.valuesToken + " " + strings.Join(values, ", ") + " " + qp.last
}

// ComposeWithValuesForPartialInsert would be used to partial insert 16384 records
func (qp *QueryHelper) ComposeWithValuesForPartialInsert(values []string) string {
	return qp.first + qp.valuesToken + " " + strings.Join(values, ", ") + " ;"
}

// Decompose parses the SQL string to extract values from a SQL statement
func (qp *QueryHelper) Decompose() []string {
	//sql := `INSERT INTO distributors (did, name) values (1, 'Cheese', 9.99), (2, 'Bread', 1.99), (3, 'Milk', 2.99) `
	parts := strings.Split(qp.sqlString, "VALUES") //what if nested statement has a values too, not supporting that at the moment
	if len(parts) == 1 {
		parts = strings.Split(qp.sqlString, "values")
		qp.valuesToken = "values"
	} //should contain the values clause, since we are doing validation in the UI
	if len(parts) == 1 {
		//Values provided by an expression.  Hopefully all on one line..
		return nil
	}

	qp.first = parts[0]
	spart := parts[1]
	spartLength := len(spart)
	i := 0

	braketCount := 0
	for i < spartLength {
		ch := spart[i]
		i++
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			continue
		}
		if ch == '(' {
			braketCount = braketCount + 1
			position := i
			for i < len(spart) {
				ch = spart[i]
				//fmt.Print(string(ch))
				i++
				if ch == '(' {
					braketCount = braketCount + 1
				}
				if ch == ')' || ch == 0 {
					braketCount = braketCount - 1
					if braketCount == 0 {
						break
					}
				}
			}
			value := "(" + spart[position:i-1] + ")"
			qp.values = append(qp.values, value)
			if i == spartLength {
				break
			}
			ch = spart[i]
			i++
			for ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				ch = spart[i]
				i++
			}
			if ch != ',' {
				break
			}
		}
	}
	qp.last = spart[i-1:]
	return qp.values
}

// ExecutePreparedQuery allows querying Snowflake with parameters
func ExecutePreparedQuery(queryString string, inputObj *Input, db *sql.DB, ActivityName string, log log.Logger) (results *ResultSet, err error) {
	//regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
	preparedQuery := queryString
	inputParams := inputObj.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString, ActivityName, log)
	if err != nil {
		return nil, GetError(DefaultError, ActivityName, err.Error())
	}

	log.Debug(GetMessage(ExecuteSQLWithParam, preparedQuery, inputArgs))

	statement, err := db.Prepare(preparedQuery)
	if err != nil {
		return nil, GetError(QueryPreparationFailed, queryString, err.Error())
	}
	defer statement.Close()

	rows, err := statement.Query(inputArgs...)
	if err != nil {
		return nil, GetError(SQLExecutionFailed, ActivityName, ActivityName, err.Error())
	}

	if rows == nil {
		log.Info(GetMessage(NoRowsFound))
		return nil, nil
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, GetError(FailedGettingColumnInfo, ActivityName, err.Error())
	}

	count := len(columns)
	cols := make([]interface{}, count)
	args := make([]interface{}, count)
	for i := range cols {
		args[i] = &cols[i]
	}

	coltypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, GetError(FailedGettingColumnType, ActivityName, err.Error())
	}

	var resultSet ResultSet
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return nil, GetError(ScanningRowFailed, ActivityName, err.Error())
		}

		m := make(record)
		for i, value := range cols {
			valueType := fmt.Sprintf("%T", value)
			dbType := coltypes[i].DatabaseTypeName()
			m[columns[i]] = value

			if value != nil {
				switch dbType {
				case "FIXED", "REAL":
					if valueType == "string" {
						if nx, ok := strconv.ParseFloat(value.(string), 512); ok == nil {
							m[columns[i]] = nx
						}
					}

				case "TEXT":
					m[columns[i]] = value.(string)

				case "BOOLEAN":
					m[columns[i]] = (value == "1")

				case "DATE":
					if t, ok := value.(time.Time); ok {
						fmt.Println(t.String())
						m[columns[i]] = t.Format("2006-01-02")
					}

				case "TIME":
					//Since we do not know precision used by user for the column type, we would display what comes from db.
					//We are removing any trailing zero's from the nano seconds part of time.
					if t, ok := value.(time.Time); ok {
						timeValueWithDefaultPrecision := t.Format("15:04:05.000000000")
						timeValueSplitOnNanoSeconds := strings.Split(timeValueWithDefaultPrecision, ".")
						nanoSecondsString := timeValueSplitOnNanoSeconds[1]
						nanoSecondsString = strings.TrimRight(nanoSecondsString, "0")
						if !strings.EqualFold(nanoSecondsString, "") {
							// nanoSecondsInteger, err := strconv.ParseInt(nanoSecondsString, 10, 64)
							// if err != nil {
							// 	fmt.Println(err)
							// 	return nil, err
							// }
							// timeValueSplitOnNanoSeconds[0] = fmt.Sprintf("%s.%d", timeValueSplitOnNanoSeconds[0], nanoSecondsInteger)
							timeValueSplitOnNanoSeconds[0] = fmt.Sprintf("%s.%s", timeValueSplitOnNanoSeconds[0], nanoSecondsString)
							m[columns[i]] = timeValueSplitOnNanoSeconds[0]
						} else {
							m[columns[i]] = timeValueSplitOnNanoSeconds[0]
						}
					}

					// review this once UI fields are available and verify if this is required after mapping output to input
					/*case "TIMESTAMP_NTZ":
						//////check this!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
						//If I set it to as "15:04:05.000" it will show 000 for fields that we had not entered precision
						if t, ok := value.(time.Time); ok {
							m[columns[i]] = t.Format("2006-01-02 15:04:05.000")
						}

					case "TIMESTAMP_LTZ", "TIMESTAMP_TZ":
						//////check this!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
						//precision is not yet considered.
						if t, ok := value.(time.Time); ok {
							m[columns[i]] = t.Format("2006-01-02 15:04:05-0700")
						}*/
				}
			}
		}
		if len(m) > 0 {
			resultSet.Record = append(resultSet.Record, &m)
		}
	}
	if len(resultSet.Record) == 0 {
		log.Info(GetMessage(NoRowsFound))
	}
	return &resultSet, nil
}

// ExecutePreparedUpdate allows querying Snowflake with parameters
func ExecutePreparedUpdate(queryString string, inputObj *Input, db *sql.DB, ActivityName string, log log.Logger) (results map[string]interface{}, err error) {
	//regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
	preparedQuery := queryString
	inputParams := inputObj.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString, ActivityName, log)
	if err != nil {
		return nil, GetError(DefaultError, ActivityName, err.Error())
	}

	log.Debug(GetMessage(ExecuteSQLWithParam, preparedQuery, inputArgs))
	result, err := db.Exec(preparedQuery, inputArgs...)
	if err != nil {
		return nil, GetError(SQLExecutionFailed, ActivityName, ActivityName, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}

// ExecutePreparedDelete allows querying Snowflake with parameters
func ExecutePreparedDelete(queryString string, inputObj *Input, db *sql.DB, ActivityName string, log log.Logger) (results map[string]interface{}, err error) {
	//regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
	preparedQuery := queryString
	inputParams := inputObj.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString, ActivityName, log)
	if err != nil {
		return nil, GetError(DefaultError, ActivityName, err.Error())
	}

	log.Debug(GetMessage(ExecuteSQLWithParam, preparedQuery, inputArgs))
	result, err := db.Exec(preparedQuery, inputArgs...)
	if err != nil {
		return nil, GetError(SQLExecutionFailed, ActivityName, ActivityName, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}

// ExecutePreparedMerge allows querying Snowflake with parameters
func ExecutePreparedMerge(queryString string, inputObj *Input, db *sql.DB, ActivityName string, log log.Logger) (results map[string]interface{}, err error) {
	//regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
	preparedQuery := queryString
	inputParams := inputObj.Parameters
	argCount := len(inputParams)
	inputArgs := make([]interface{}, argCount)

	inputArgs, preparedQuery, err = getQueryAndArgs(inputParams, queryString, ActivityName, log)
	if err != nil {
		return nil, GetError(DefaultError, ActivityName, err.Error())
	}

	log.Debug(GetMessage(ExecuteSQLWithParam, preparedQuery, inputArgs))
	result, err := db.Exec(preparedQuery, inputArgs...)
	if err != nil {
		return nil, GetError(SQLExecutionFailed, ActivityName, ActivityName, err.Error())
	}
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected
	return output, nil
}

func getQueryAndArgs(inputParams map[string]interface{}, query string, ActivityName string, log log.Logger) (args []interface{}, preparedQuery string, err error) {

	regexDate := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}\\+[0-9]{2}:[0-9]{2}$")
	var inputArgs []interface{}
	preparedQuery = query

	inputOffsets := generateInputWithOffset(inputParams, query)
	sort.Sort(inputOffsets)

	for i := 0; i < len(inputOffsets); i++ {
		io := inputOffsets[i]
		preparedQuery = strings.Replace(preparedQuery, "?"+string(io.keyfield), "?", 1)

		strValue, err := coerce.ToString(io.value)
		if err != nil {
			return nil, query, GetError(DefaultError, ActivityName, err.Error())
		}

		if regexDate.MatchString(strValue) == true {
			//Date
			t, err := time.Parse("2006-01-02+00:00", strValue)
			if err != nil {
				return nil, query, GetError(DefaultError, ActivityName, err.Error())
			}
			//YYYY-MM-DD
			strValue = t.Format("2006-01-02")
		}
		inputArgs = append(inputArgs, strValue)
	}

	return inputArgs, preparedQuery, nil
}

//Commenting this function for now. It would be used when we support blob type
// func getSchemaFromFields(fields interface{}) map[string]string {
// 	schema := map[string]string{}
// 	for _, fieldObject := range fields.([]interface{}) {
// 		if fieldName, ok := fieldObject.(map[string]interface{})["FieldName"]; ok {
// 			schema[fieldName.(string)] = fieldObject.(map[string]interface{})["Type"].(string)
// 		}
// 	}
// 	return schema
// }

// PreparedInsert - Code referred from mysql
// PreparedInsert allows querying database with named parameters
func PreparedInsert(queryString string, inputData *Input /*fields interface{}, */, db *sql.DB, ActivityName string, log log.Logger) (results map[string]interface{}, err error) {

	//schema := getSchemaFromFields(fields)
	prepared := queryString
	queryHelper := NewQueryHelper(queryString)
	queryValues := queryHelper.Decompose()

	queryValuesLength := len(queryValues)
	inputValuesLength := len(inputData.Values)
	if queryValuesLength > 1 && queryValuesLength != inputValuesLength && inputValuesLength != 0 {
		return nil, GetError(DefaultError, ActivityName, fmt.Errorf("input data values length does not match query input data values length, %v != %v", queryValuesLength, inputValuesLength).Error())
	}

	inputArgs := []interface{}{} // not very efficient better to size, but to find size you have to search, will optimize later
	replacedValues := []string{}
	var partialRowsAffected int64
	valuesCounter := 0

	//for logging purpose
	inputArgsForLogging := []interface{}{}

	//using a transaction object because we do not want a partial insert
	tx, err := db.Begin()

	var queryValue string
	if queryValuesLength > 0 {
		queryValue = queryValues[0]
	}
	// process the values clauses
	for inputValuesIndex, inputValues := range inputData.Values {
		//16384 is the maximum rows that can be inserted in a single insert statement (Snowflake driver limitation)
		//https://github.com/snowflakedb/gosnowflake/issues/253
		if valuesCounter == 16384 {
			if len(replacedValues) > 0 {
				prepared = queryHelper.ComposeWithValuesForPartialInsert(replacedValues)
			}
			result, err := tx.Exec(prepared, inputArgs...)
			if err != nil {
				log.Errorf("PreparedQuery got error: %s", err)
				tx.Rollback()
				return nil, GetError(DefaultError, ActivityName, err.Error())
			}
			rowsAffected, _ := result.RowsAffected()
			partialRowsAffected = partialRowsAffected + rowsAffected
			inputArgs = nil
			replacedValues = nil
			valuesCounter = 0
		}

		if queryValuesLength > 1 {
			queryValue = queryValues[inputValuesIndex]
		}
		regExp := regexp.MustCompile("\\?\\w*")
		matches := regExp.FindAllStringSubmatch(queryValue, -1)

		value := queryValue
		for _, match := range matches {
			parameter := strings.Split(match[0], "?")[1]
			substitution, ok := inputData.Parameters[parameter]
			if !ok {
				substitution, ok = inputValues[parameter]
				if !ok {
					return nil, GetError(DefaultError, ActivityName, fmt.Errorf("missing substitution for: %s", match[0]).Error())
				}
			}
			// replace the first occurance, as it is found
			value = strings.Replace(value, match[0], "?", 1)

			// TBD refactor into a mysql.Marshal function
			// parameterType, ok := schema[parameter]
			// if ok && parameterType == "BLOB" || parameterType == "IMAGE" {
			// 	substitution = decodeBlob(substitution.(string))
			// }
			// end refactor
			inputArgs = append(inputArgs, substitution)
			inputArgsForLogging = append(inputArgsForLogging, substitution)
		}
		replacedValues = append(replacedValues, value)

		valuesCounter++
	}
	//process parameters in values clauses
	if inputValuesLength == 0 { // just do the parameters
		for queryValuesIndex, queryValue := range queryValues {
			if queryValuesLength > 1 {
				queryValue = queryValues[queryValuesIndex]
			}
			regExp := regexp.MustCompile("\\?\\w*")
			matches := regExp.FindAllStringSubmatch(queryValue, -1)

			value := queryValue
			for _, match := range matches {
				parameter := strings.Split(match[0], "?")[1]
				substitution, ok := inputData.Parameters[parameter]
				if !ok {
					return nil, GetError(DefaultError, ActivityName, fmt.Errorf("missing parameter substitution for: %s", match[0]).Error())
				}
				// replace the first occurance, as it is found
				value = strings.Replace(value, match[0], "?", 1)
				//log.Debugf("prepared statement: %s", value)

				// TBD refactor into a mysql.Marshal function
				// parameterType, ok := schema[parameter]
				// if ok && parameterType == "BLOB" || parameterType == "IMAGE" {
				// 	substitution = decodeBlob(substitution.(string))
				// }
				// end refactor
				inputArgs = append(inputArgs, substitution)
				inputArgsForLogging = append(inputArgsForLogging, substitution)
			}
			replacedValues = append(replacedValues, value)
		}
	}
	if len(replacedValues) > 0 {
		prepared = queryHelper.ComposeWithValues(replacedValues)
	}

	//process parameters not in values clauses (for instance a select statement)
	r := regexp.MustCompile("\\?\\w*")
	matches := r.FindAllStringSubmatch(prepared, -1)
	for _, match := range matches {
		parameter := strings.Split(match[0], "?")[1]
		parameterValue := inputData.Parameters[parameter]
		if parameterValue == nil {
			continue
		}
		prepared = strings.Replace(prepared, match[0], "?", 1)
		//log.Debugf("prepared statement: %s", prepared)
		inputArgs = append(inputArgs, parameterValue)
		inputArgsForLogging = append(inputArgsForLogging, parameterValue)
	}

	log.Debug(GetMessage(ExecuteSQLWithParam, queryString, inputArgsForLogging))

	result, err := tx.Exec(prepared, inputArgs...)
	if err != nil {
		log.Errorf("PreparedQuery got error: %s", err)
		tx.Rollback()
		return nil, GetError(DefaultError, ActivityName, err.Error())
	}
	tx.Commit()
	output := make(map[string]interface{})
	rowsAffected, _ := result.RowsAffected()
	output["rowsAffected"] = rowsAffected + partialRowsAffected
	return output, nil
}

// GetInputData converts the input to snowflake.Input
func GetInputData(inputData interface{}, log log.Logger, ActivityName string) (inputParams *Input, err error) {
	inputParams = &Input{}

	if inputData == nil {
		return nil, GetError(SpecifyInput, ActivityName)
	}

	switch inputData.(type) {
	case string:
		log.Debug(GetMessage(ActivityInput, inputData.(string)))
		tempMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(inputData.(string)), &tempMap)
		if err != nil {
			return nil, GetError(DefaultError, ActivityName, err.Error())
		}
		inputParams.Parameters = tempMap
	default:
		dataBytes, err := json.Marshal(inputData)
		log.Debug(GetMessage(ActivityInput, string(dataBytes)))
		if err != nil {
			return nil, GetError(DefaultError, ActivityName, err.Error())
		}
		err = json.Unmarshal(dataBytes, inputParams)
		if err != nil {
			return nil, GetError(DefaultError, ActivityName, err.Error())
		}
	}
	return
}

func ExecutePutCommand(inputData map[string]interface{}, inputParams *CommandInput, db *sql.DB, activityName string, log log.Logger) (*ResultSet, error) {
	query := fmt.Sprintf("PUT file://%s %s", inputParams.FileName, inputParams.InternalStage)
	if inputParams.Source_Compression != "" {
		query = query + fmt.Sprintf(" SOURCE_COMPRESSION = %s", inputParams.Source_Compression)
	}
	if inputParams.Parallel != 0 {
		query = query + fmt.Sprintf(" PARALLEL = %d", inputParams.Parallel)
	}
	if IsInputPresent(inputData, "auto_compress") {
		query = query + fmt.Sprintf(" AUTO_COMPRESS = %t", inputParams.Auto_Compress)
	}
	if IsInputPresent(inputData, "overwrite") {
		query = query + fmt.Sprintf(" OVERWRITE = %t", inputParams.Overwrite)
	}
	log.Debugf("Query: %s", query)
	// Execute the PUT command
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows == nil {
		log.Info(GetMessage(NoRowsFound))
		return nil, nil
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, GetError(FailedGettingColumnInfo, activityName, err.Error())
	}
	count := len(columns)
	cols := make([]interface{}, count)
	args := make([]interface{}, count)
	for i := range cols {
		args[i] = &cols[i]
	}
	coltypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, GetError(FailedGettingColumnType, activityName, err.Error())
	}
	var resultSet ResultSet
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return nil, GetError(ScanningRowFailed, activityName, err.Error())
		}
		m := make(record)
		for i, value := range cols {
			valueType := fmt.Sprintf("%T", value)
			dbType := coltypes[i].DatabaseTypeName()
			m[columns[i]] = value
			if value != nil {
				switch dbType {
				case "FIXED", "REAL":
					if valueType == "string" {
						if nx, ok := strconv.ParseFloat(value.(string), 512); ok == nil {
							m[columns[i]] = nx
						}
					}
				case "TEXT":
					m[columns[i]] = value.(string)
				}
			}
		}
		if len(m) > 0 {
			resultSet.Record = append(resultSet.Record, &m)
		}
	}
	if len(resultSet.Record) == 0 {
		log.Info(GetMessage(NoRowsFound))
	}
	return &resultSet, nil
}

func ExecuteCopyIntoCommand(inputData map[string]interface{}, inputParams *CommandInput, db *sql.DB, activityName string, log log.Logger) (*ResultSet, error) {
	query := fmt.Sprintf("COPY INTO %s FROM %s", inputParams.TableName, inputParams.StageName)
	if inputParams.Files != "" {
		query = query + fmt.Sprintf(" FILES = %s", inputParams.Files)
	}
	if inputParams.Pattern != "" {
		query = query + fmt.Sprintf(" PATTERN = %s", inputParams.Pattern)
	}

	if inputParams.FileType != "" || inputParams.Format_Name != "" {
		query += " FILE_FORMAT = ("
		if inputParams.FileType != "" {
			query += fmt.Sprintf("TYPE = %s", inputParams.FileType)
		}
		if inputParams.Format_Name != "" {
			if inputParams.FileType != "" {
				query += " "
			}
			query += fmt.Sprintf("FORMAT_NAME = %s", inputParams.Format_Name)
		}
		if inputParams.FormatTypeOptions != "" {
			result := make(map[string]interface{})
			json.Unmarshal([]byte(inputParams.FormatTypeOptions), &result)

			query += ", "
			for key, value := range result {
				switch v := value.(type) {
				case string:
					query += fmt.Sprintf("%s = '%s', ", key, v)
				case int:
					query += fmt.Sprintf("%s = %d, ", key, v)
				case bool:
					query += fmt.Sprintf("%s = %t, ", key, v)
				default:
					log.Warnf("Unsupported type for option %s: %T", key, value)
				}
			}
			query = query[:len(query)-2]
		}
		query += ")"
	}

	if inputParams.CopyOptions.On_error != "" {
		query = query + fmt.Sprintf(" ON_ERROR = %s", inputParams.CopyOptions.On_error)
	}
	if inputParams.CopyOptions.Size_Limit != 0 {
		query = query + fmt.Sprintf(" SIZE_LIMIT = %d", inputParams.CopyOptions.Size_Limit)
	}
	if inputParams.CopyOptions.Match_By_Column_Name != "" {
		query = query + fmt.Sprintf(" MATCH_BY_COLUMN_NAME = %s", inputParams.CopyOptions.Match_By_Column_Name)
	}
	if IsInputPresent(inputData, "purge") {
		query = query + fmt.Sprintf(" PURGE = %t", inputParams.CopyOptions.Purge)
	}
	if IsInputPresent(inputData, "return_failed_only") {
		query = query + fmt.Sprintf(" RETURN_FAILED_ONLY = %t", inputParams.CopyOptions.Return_Failed_Only)
	}
	if IsInputPresent(inputData, "enforce_length") {
		query = query + fmt.Sprintf(" ENFORCE_LENGTH = %t", inputParams.CopyOptions.Enforce_Length)
	}
	if IsInputPresent(inputData, "truncatecolumns") {
		query = query + fmt.Sprintf(" TRUNCATECOLUMNS = %t", inputParams.CopyOptions.TruncateColumns)
	}
	if IsInputPresent(inputData, "force") {
		query = query + fmt.Sprintf(" FORCE = %t", inputParams.CopyOptions.Force)
	}
	if IsInputPresent(inputData, "load_uncertain_files") {
		query = query + fmt.Sprintf(" LOAD_UNCERTAIN_FILES = %t", inputParams.CopyOptions.Load_Uncertain_Files)
	}
	if inputParams.Validation_Mode != "" {
		query = query + fmt.Sprintf(" VALIDATION_MODE = %s", inputParams.Validation_Mode)
	}
	log.Debugf("Query: %s", query)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows == nil {
		log.Info(GetMessage(NoRowsFound))
		return nil, nil
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, GetError(FailedGettingColumnInfo, activityName, err.Error())
	}

	count := len(columns)
	cols := make([]interface{}, count)
	args := make([]interface{}, count)
	for i := range cols {
		args[i] = &cols[i]
	}

	coltypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, GetError(FailedGettingColumnType, activityName, err.Error())
	}

	var resultSet ResultSet
	for rows.Next() {
		if err := rows.Scan(args...); err != nil {
			return nil, GetError(ScanningRowFailed, activityName, err.Error())
		}

		m := make(record)
		for i, value := range cols {
			valueType := fmt.Sprintf("%T", value)
			dbType := coltypes[i].DatabaseTypeName()
			m[columns[i]] = value

			if value != nil {
				switch dbType {
				case "FIXED", "REAL":
					if valueType == "string" {
						if nx, ok := strconv.ParseFloat(value.(string), 512); ok == nil {
							m[columns[i]] = nx
						}
					}

				case "TEXT":
					m[columns[i]] = value.(string)
				}
			}
		}
		if len(m) > 0 {
			resultSet.Record = append(resultSet.Record, &m)
		}
	}
	if len(resultSet.Record) == 0 {
		log.Info(GetMessage(NoRowsFound))
	}
	return &resultSet, nil
}

// GetCommandInput converts the input to CommandInput
func GetCommandInput(inputData interface{}, log log.Logger) (inputParams *CommandInput, err error) {
	inputParams = &CommandInput{}
	if inputData == nil {
		return nil, fmt.Errorf(GetMessage(SpecifyInput))
	}

	//log input at debug level
	dataBytes, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}
	log.Debugf(GetMessage(ActivityInput, string(dataBytes)))

	err = json.Unmarshal(dataBytes, inputParams)
	if err != nil {
		return nil, fmt.Errorf(GetMessage(DefaultError, err.Error()))
	}

	return inputParams, nil
}

func IsInputPresent(inputData map[string]interface{}, key string) bool {
	for k, v := range inputData {
		switch v.(type) {
		case map[string]interface{}:
			flag := IsInputPresent(v.(map[string]interface{}), key)
			return flag
		default:
			if k == key {
				return true
			}
		}
	}
	return false
}
