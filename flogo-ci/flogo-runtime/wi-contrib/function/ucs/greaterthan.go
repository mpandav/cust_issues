package ucs

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type GreaterThan struct {
}

func init() {
	function.Register(&GreaterThan{})
}

func (gt *GreaterThan) Name() string {
	return "greaterThan"
}

func (gt *GreaterThan) GetCategory() string {
	return "ucs"
}

func (gt *GreaterThan) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString, data.TypeAny}, false
}

func (gt *GreaterThan) Eval(params ...interface{}) (interface{}, error) {
	var param string
	var value, jsonObject interface{}
	var err error

	log.RootLogger().Debug("UCS greaterThan function initiated ...")
	if len(params) < 2 {
		return nil, fmt.Errorf("Incorrect parameters provided for the ucs.greaterThan function")
	}

	paramInterface, err := GetParam(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.greaterThan parameters")
	}

	param = paramInterface.(string)

	value, err = GetValue(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.greaterThan parameters")
	}

	if param != "" && value != nil {
		jsonObject = SimpleLookupCondition{Expr: GREATER_THAN_EXPR, Prop: param, Val: value}
	}

	return jsonObject, nil
}
