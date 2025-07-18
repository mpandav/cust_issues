package utility

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/support/log"
)

type RenderJson struct {
}

func init() {
	function.Register(&RenderJson{})
}

func (s *RenderJson) Name() string {
	return "renderJSON"
}

func (s *RenderJson) GetCategory() string {
	return "utility"
}

func (s *RenderJson) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeAny, data.TypeBool}, false
}

func (s *RenderJson) Eval(params ...interface{}) (interface{}, error) {
	object := params[0]
	printPretty, err := coerce.ToBool(params[1])
	if err != nil {
		return nil, fmt.Errorf("renderJSON function second parameter must be boolean")
	}
	log.RootLogger().Debugf("Start render-json function with parameters %s", object)
	var result string
	if printPretty {
		b, err := json.MarshalIndent(object, "", "    ")
		if err != nil {
			return "", err
		}
		result = string(b)
	} else {
		buffer := new(bytes.Buffer)
		json_bytes, err := json.Marshal(object)
		if err != nil {
			return "", err
		}
		err = json.Compact(buffer, json_bytes)
		if err != nil {
			return "", err
		}
		result = buffer.String()
	}

	log.RootLogger().Debugf("Done render-json function with result %s", result)
	return result, nil
}
