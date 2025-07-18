package ucs

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type LessThanEqual struct {
}

func init() {
	function.Register(&LessThanEqual{})
}

func (lte *LessThanEqual) Name() string {
	return "lessThanEqual"
}

func (lte *LessThanEqual) GetCategory() string {
	return "ucs"
}

func (lte *LessThanEqual) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString, data.TypeAny}, false
}

func (lte *LessThanEqual) Eval(params ...interface{}) (interface{}, error) {
	var param string
	var value, jsonObject interface{}
	var err error

	log.RootLogger().Debug("UCS lessThanEqual function initiated ...")
	if len(params) < 2 {
		return nil, fmt.Errorf("Incorrect parameters provided for the ucs.lessThanEqual function")
	}

	paramInterface, err := GetParam(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.lessThanEqual parameters")
	}

	param = paramInterface.(string)

	value, err = GetValue(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.lessThanEqual parameters")
	}

	if param != "" && value != nil {
		jsonObject = SimpleLookupCondition{Expr: LESS_THAN_EQUAL_EXPR, Prop: param, Val: value}
	}

	return jsonObject, nil
}
