package connection

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// DBDiagnostic ..
type DBDiagnostic struct {
	State       string `json:"State"`
	NativeError string `json:"NativeError"`
	Messge      string `json:"Message"`
}

// ActionError ..
type ActionError struct {
	APIName string         `json:"APIName"`
	Diag    []DBDiagnostic `json:"Diag,omitempty"`
}

// New ..
func New(actionerr ActionError) error {
	return &ActionError{
		APIName: actionerr.APIName,
		Diag:    actionerr.Diag,
	}
}

func (acterr *ActionError) Error() string {
	out, err := json.Marshal(acterr)
	if err != nil {
		return "Error JSON Marshal Fialed"
	}
	return string(out)
}

const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_."
const terminators = " ;),<>+-*%/"
const starters = " =,(<>+-*%/"
const quotes = "\"'`"

func alphaOnly(s string, keyset string) bool {
	for _, char := range s {
		if !strings.Contains(keyset, strings.ToLower(string(char))) {
			return false
		}
	}
	return true
}

// QueryBasicCheck ..
func QueryBasicCheck(s string) error {
	if alphaOnly(string(s[0]), quotes) || alphaOnly(string(s[len(s)-1]), quotes) {
		var dignostic []DBDiagnostic
		QueryError := ActionError{
			APIName: "HandleQuery",
			Diag: append(dignostic, DBDiagnostic{
				State:       "42000",
				NativeError: "1064",
				Messge:      "Query Syntax Error: Query should not be quoted, syntax error at or near " + string(s[0]) + " index:0",
			}),
		}
		return New(QueryError)
	}
	return nil
}

func getSubstutionValue(paramValue interface{}, paramType string, param string) (string, error) {
	substitute := ""
	switch paramType {

	case "VARCHAR", "CHAR", "TIMESTAMP", "DATE", "LONGVARCHAR":
		s, ok := paramValue.(string)
		if !ok {
			return "", fmt.Errorf("error converting %v to string", paramValue)
		}
		substitute = "'" + s + "'"
	//
	default:
		substitute = fmt.Sprintf("%v", paramValue)
	}

	return substitute, nil
}
func EvaluateQuery(query string, inputData Input, paramTypes []string) (string, error) {
	re := regexp.MustCompile("\\n")
	query = re.ReplaceAllString(query, " ")
	re = regexp.MustCompile("\\t")
	query = re.ReplaceAllString(query, " ")
	query = strings.TrimSpace(query)
	if query[len(query)-1:] != ";" {
		query += ";"
	}
	param := ""

	dq, sq, bt := "\"", "'", "`"
	dqm, sqm, btm := false, false, false
	paramMarker := false
	paramCounter := 0
	prepared := query
	modifiedLength := 0 //"ad?da", 'as?as', = ?name where ;
	for i, val := range query {
		chr := string(val)
		if !dqm || !sqm || !btm {
			if chr == dq && string(query[i-1]) != "\\" && !sqm && !btm {
				dqm = !dqm
				continue
			}
			if chr == sq && string(query[i-1]) != "\\" && !dqm && !btm {
				sqm = !sqm
				continue
			}
			if chr == bt && string(query[i-1]) != "\\" && !dqm && !sqm {
				btm = !btm
				continue
			}
			if !dqm && !sqm && !btm {
				if chr == "?" && alphaOnly(string(query[i-1]), starters) {
					if paramMarker {
						paramMarker = false
						param = ""
						continue
					}
					paramMarker = true
					param = ""
					continue
				} //		 select *from users where name=?name and sal=?salary sarvrav; i = 25-5=20
				//prepared = select *from users where name='s' and sal=500 ;
				if paramMarker {
					if !alphaOnly(chr, alpha) && chr != "\n" {
						paramMarker = false
						if param == "" && string(query[i-1]) == "?" && alphaOnly(chr, terminators) {
							return "", fmt.Errorf("Parameters can not be unnamed, hint: ?paramname")
						}
						if param == "" && string(query[i-1]) == "?" && !alphaOnly(chr, terminators) {
							continue
						}
						if alphaOnly(chr, terminators) {
							paramLength := len(param)
							substituteStartIndex := i - paramLength - 1 - modifiedLength
							substitute, err := getSubstutionValue(inputData.Parameters[param], paramTypes[paramCounter], param)
							if err != nil {
								return "", err
							}
							prepared = prepared[:substituteStartIndex] + substitute + prepared[i-modifiedLength:]
							modifiedLength += paramLength - len(substitute) + 1

							paramCounter++
							param = ""
							continue
						}
					}
					param = param + chr
				}
			}
		}
	}

	return prepared, nil
}
