package ucs

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type Equal struct {
}

func init() {
	function.Register(&Equal{})
}

func (eq *Equal) Name() string {
	return "equal"
}

func (eq *Equal) GetCategory() string {
	return "ucs"
}

func (eq *Equal) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString, data.TypeAny}, false
}

func (eq *Equal) Eval(params ...interface{}) (interface{}, error) {
	var param string
	var value, jsonObject interface{}
	var err error

	log.RootLogger().Debug("UCS equal function initiated ...")
	if len(params) < 2 {
		return nil, fmt.Errorf("Incorrect parameters provided for the ucs.equal function")
	}

	paramInterface, err := GetParam(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.equal parameters")
	}

	param = paramInterface.(string)

	value, err = GetValue(params)
	if err != nil {
		return nil, fmt.Errorf("Error occurred while processing ucs.equal parameters")
	}

	if param != "" && value != nil {
		jsonObject = SimpleLookupCondition{Expr: EQUAL_EXPR, Prop: param, Val: value}
	}

	return jsonObject, nil
}
