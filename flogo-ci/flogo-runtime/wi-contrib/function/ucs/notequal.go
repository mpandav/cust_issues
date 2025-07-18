package ucs

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type NotEqual struct {
}

func init() {
	function.Register(&NotEqual{})
}

func (neq *NotEqual) Name() string {
	return "notEqual"
}

func (neq *NotEqual) GetCategory() string {
	return "ucs"
}

func (neq *NotEqual) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString, data.TypeAny}, false
}

func (neq *NotEqual) Eval(params ...interface{}) (interface{}, error) {
	var param string
	var value, jsonObject interface{}
	var err error

	log.RootLogger().Debug("UCS not equal function initiated ...")
	if len(params) < 2 {
		return nil, fmt.Errorf("Incorrect parameters provided for the ucs.notEqual function")
	}

	paramInterface, err := GetParam(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.notEqual parameters")
	}

	param = paramInterface.(string)

	value, err = GetValue(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.notEqual parameters")
	}

	if param != "" && value != nil {
		jsonObject = SimpleLookupCondition{Expr: NOT_EQUAL_EXPR, Prop: param, Val: value}
	}

	return jsonObject, nil
}
