package datetime

import (
	"fmt"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
)

// Deprecated
type FormatDatetime struct {
}

func init() {
	function.Register(&FormatDatetime{})
}

func (s *FormatDatetime) Name() string {
	return "formatDatetime"
}

func (s *FormatDatetime) GetCategory() string {
	return "datetime"
}

func (s *FormatDatetime) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeDateTime, data.TypeString}, false
}

func (s *FormatDatetime) Eval(params ...interface{}) (interface{}, error) {
	date, err := coerce.ToDateTime(params[0])
	if err != nil {
		return nil, fmt.Errorf("Format datetime first argument must be string")
	}
	f, err := coerce.ToString(params[1])
	if err != nil {
		return nil, fmt.Errorf("Format datetime second argument must be string")
	}
	return date.Format(format(f)), nil
}
