package ucs

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type LessThan struct {
}

func init() {
	function.Register(&LessThan{})
}

func (lt *LessThan) Name() string {
	return "lessThan"
}

func (lt *LessThan) GetCategory() string {
	return "ucs"
}

func (lt *LessThan) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString, data.TypeAny}, false
}

func (lt *LessThan) Eval(params ...interface{}) (interface{}, error) {
	var param string
	var value, jsonObject interface{}
	var err error

	log.RootLogger().Debug("UCS lessThan function initiated ...")
	if len(params) < 2 {
		return nil, fmt.Errorf("Incorrect parameters provided for the ucs.lessThan function")
	}

	paramInterface, err := GetParam(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.lessThan parameters")
	}

	param = paramInterface.(string)

	value, err = GetValue(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.lessThan parameters")
	}

	if param != "" && value != nil {
		jsonObject = SimpleLookupCondition{Expr: LESS_THAN_EXPR, Prop: param, Val: value}
	}

	return jsonObject, nil
}
