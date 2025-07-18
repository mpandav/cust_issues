package ucs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-flogo/core/data/coerce"
)

const (
	EQUAL_EXPR              = "eq"
	NOT_EQUAL_EXPR          = "neq"
	LESS_THAN_EXPR          = "lt"
	LESS_THAN_EQUAL_EXPR    = "lte"
	GREATER_THAN_EXPR       = "gt"
	GREATER_THAN_EQUAL_EXPR = "gte"
	OR_EXPR                 = "or"
	AND_EXPR                = "and"
)

type SimpleLookupCondition struct {
	Expr string      `json:"expr,required"`
	Prop string      `json:"prop,required"`
	Val  interface{} `json:"val,required"`
}

type ComplexLookupCondition struct {
	Expr  string                `json:"expr,required"`
	Left  SimpleLookupCondition `json:"left,omitempty"`
	Right SimpleLookupCondition `json:"right,omitempty"`
}

func GetParam(params []interface{}) (interface{}, error) {
	var err error

	switch param := params[0].(type) {
	case string:
		{
			param, err = coerce.ToString(param)
			if err != nil {
				return nil, fmt.Errorf("ucs function first parameter [%+v] must be string", param)
			}
			return param, nil
		}
	case SimpleLookupCondition:
		{
			return param, nil
		}
	default:
		return nil, fmt.Errorf("Error reading ucs function first parameter")
	}
}

func GetValue(params []interface{}) (interface{}, error) {
	var value interface{}
	switch val := params[1].(type) {
	case int:
		value = val

	case int32:
		value = val

	case int64:
		value = val

	case float32:
		value = val

	case float64:
		value = val

	case string:
		value = val

	case bool:
		value = val

	case nil:
		return nil, fmt.Errorf("ucs.equal function second parameter must not be nil")

	default:
		return nil, fmt.Errorf("ucs.equal function second parameter is unknown")
	}
	return value, nil
}

func GetSimpleConditionParam(inputParams []interface{}) ([]interface{}, error) {
	var params []interface{}
	var simpleParam SimpleLookupCondition

	for _, param := range inputParams {
		switch param := param.(type) {
		case string:
			if strings.Contains(param, "ucs.or(") || strings.Contains(param, "ucs.and(") {
				return nil, fmt.Errorf("Nested ucs.and() or ucs.or() expression found. Nested UCS functions not supported")
			}
		case SimpleLookupCondition:
			params = append(params, param)
		default:
			paramMap, ok := param.(map[string]interface{})
			if ok {
				paramByte, err := json.Marshal(paramMap)
				if err != nil {
					return nil, fmt.Errorf("Error reading ucs function parameter")
				}
				json.Unmarshal(paramByte, &simpleParam)
				if simpleParam.Expr == "or" || simpleParam.Expr == "and" {
					return nil, fmt.Errorf("Nested ucs.and() or ucs.or() expression found. Nested UCS functions not supported")
				}
				params = append(params, simpleParam)
			}
		}
	}
	return params, nil
}
