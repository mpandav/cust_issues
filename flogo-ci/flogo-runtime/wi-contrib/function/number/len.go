package number

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/log"

	"github.com/project-flogo/core/data/expression/function"
)

type Len struct {
}

func init() {
	function.Register(&Len{})
}

func (s *Len) Name() string {
	return "len"
}

func (s *Len) GetCategory() string {
	return "number"
}

func (s *Len) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeAny}, false
}

func (s *Len) Eval(str ...interface{}) (interface{}, error) {
	log.RootLogger().Debugf("Start len function with parameters %s", str)
	switch t := str[0].(type) {
	case string:
		return len(t), nil
	case []interface{}:
		return len(t), nil
	default:
		s, err := coerce.ToString(str)
		if err != nil {
			log.RootLogger().Errorf("Convert %+v to string error %s", str, err.Error())
		} else {
			return len(s), nil
		}
	}
	return 0, nil
}
