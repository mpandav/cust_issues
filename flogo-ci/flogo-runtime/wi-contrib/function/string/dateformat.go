package string

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"os"
)

type DateFormat struct {
}

func init() {
	function.Register(&DateFormat{})
}

func (s *DateFormat) Name() string {
	return "dateFormat"
}

func (s *DateFormat) GetCategory() string {
	return "string"
}

func (s *DateFormat) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{}, true
}

func (s *DateFormat) Eval(d ...interface{}) (interface{}, error) {
	return GetDateFormat(), nil
}

const (
	WI_DATE_FORMAT         string = "WI_DATE_FORMAT"
	WI_DATE_FORMAT_DEFAULT string = "2006-01-02-07:00"
)

func GetDateFormat() string {
	format, ok := os.LookupEnv(WI_DATE_FORMAT)
	if ok && format != "" {
		return format
	}
	return WI_DATE_FORMAT_DEFAULT
}
