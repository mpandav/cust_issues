package string

import (
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/expression/function"
	"os"
)

type TimeFormat struct {
}

func init() {
	function.Register(&TimeFormat{})
}

func (s *TimeFormat) Name() string {
	return "timeFormat"
}

func (s *TimeFormat) GetCategory() string {
	return "string"
}

func (s *TimeFormat) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{}, false
}

func (s *TimeFormat) Eval(d ...interface{}) (interface{}, error) {
	return GetTimeFormat(), nil
}

const (
	WI_TIME_FORMAT         string = "WI_TIME_FORMAT"
	WI_TIME_FORMAT_DEFAULT string = "15:04:05-07:00"
)

func GetTimeFormat() string {
	format, ok := os.LookupEnv(WI_TIME_FORMAT)
	if ok && format != "" {
		return format
	}
	return WI_TIME_FORMAT_DEFAULT
}
