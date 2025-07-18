package string

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"os"
)

type DatetimeFormat struct {
}

func init() {
	function.Register(&DatetimeFormat{})
}

func (s *DatetimeFormat) Name() string {
	return "datetimeFormat"
}

func (s *DatetimeFormat) GetCategory() string {
	return "string"
}

func (s *DatetimeFormat) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{}, true
}

func (s *DatetimeFormat) Eval(d ...interface{}) (interface{}, error) {
	return GetDatetimeFormat(), nil
}

const (
	WI_DATETIME_FORMAT         string = "WI_DATETIME_FORMAT"
	WI_DATETIME_FORMAT_DEFAULT string = "2006-01-02T15:04:05-07:00"
)

func GetDatetimeFormat() string {
	format, ok := os.LookupEnv(WI_DATETIME_FORMAT)
	if ok && format != "" {
		return format
	}
	return WI_DATETIME_FORMAT_DEFAULT
}
