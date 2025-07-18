package string

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type String struct {
}

func init() {
	function.Register(&String{})
}

func (s *String) Name() string {
	return "tostring"
}

func (s *String) GetCategory() string {
	return "string"
}

func (s *String) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeAny}, false
}

func (s *String) Eval(d ...interface{}) (interface{}, error) {
	log.RootLogger().Debugf("Start String function with parameters %s", d[0])
	return coerce.ToString(d[0])
}
