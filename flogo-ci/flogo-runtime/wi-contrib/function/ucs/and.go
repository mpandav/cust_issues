package ucs

import (
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type AND struct {
}

func init() {
	function.Register(&AND{})
}

func (and *AND) Name() string {
	return "and"
}

func (and *AND) GetCategory() string {
	return "ucs"
}

func (and *AND) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeObject, data.TypeObject}, false
}

func (and *AND) Eval(params ...interface{}) (interface{}, error) {
	var param1, param2 SimpleLookupCondition
	var jsonObject interface{}
	var err error

	log.RootLogger().Debug("UCS AND function initiated ...")
	if len(params) < 2 {
		return nil, fmt.Errorf("Incorrect parameters provided for the ucs.and function")
	}

	paramInterface, err := GetSimpleConditionParam(params)
	if err != nil {
		return nil, err
	}

	if len(paramInterface) != 2 {
		return nil, fmt.Errorf("Error occurred while processing ucs.and parameters")
	}

	param1 = paramInterface[0].(SimpleLookupCondition)
	param2 = paramInterface[1].(SimpleLookupCondition)

	if param1.Expr != "" && param2.Expr != "" {
		jsonObject = ComplexLookupCondition{Expr: AND_EXPR, Left: param1, Right: param2}
	}

	return jsonObject, nil
}
