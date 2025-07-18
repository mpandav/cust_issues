package array

import (
	"fmt"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
)

func init() {
	_ = function.Register(&fnForeach{})
}

type fnForeach struct {
}

// Name returns the name of the function
func (fnForeach) Name() string {
	return "forEach"
}

// Sig returns the function signature
func (fnForeach) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeAny}, true
}

// Eval executes the function
func (fnForeach) Eval(params ...interface{}) (interface{}, error) {
	if len(params) <= 0 {
		return nil, nil
	}

	switch len(params) {
	case 1, 2:
		return params[0], nil
	case 3:
		//Need filter here, not support.
		return nil, fmt.Errorf("array.foreach() function does not support third arguments when use it in expression, please remove third argument")
	default:
		return nil, fmt.Errorf("too many arguments for function array.foreach(), expected 0 to 3 but got [%d]", len(params))
	}

}
