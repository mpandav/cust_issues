package number

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type Int struct {
}

func init() {
	function.Register(&Int{})
}

func (s *Int) Name() string {
	return "int64"
}

func (s *Int) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeAny}, false
}

func (s *Int) GetCategory() string {
	return "number"
}

func (s *Int) Eval(in ...interface{}) (interface{}, error) {
	log.RootLogger().Debugf("Start Int64 function with parameters %s", in)
	return coerce.ToInt(in[0])
}
