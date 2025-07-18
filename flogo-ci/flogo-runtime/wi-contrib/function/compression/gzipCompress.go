package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
)

type GzipCompress struct {
}

func init() {
	function.Register(&GzipCompress{})
}

func (g *GzipCompress) Name() string {
	return "gzipCompress"
}

func (g *GzipCompress) GetCategory() string {
	return "compression"
}

func (g *GzipCompress) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString}, false
}

func (s *GzipCompress) Eval(params ...interface{}) (interface{}, error) {

	inputStr, err := coerce.ToString(params[0])
	if err != nil {
		return nil, fmt.Errorf("gzipCompress function argument must be string")
	}
	// log.RootLogger().Debugf("Decode base64 to string \"%s\"", params[0])

	var inputBuffer bytes.Buffer
	gWriter := gzip.NewWriter(&inputBuffer)
	_, err = gWriter.Write([]byte(inputStr))
	if err != nil {
		return "", err
	}
	if err := gWriter.Close(); err != nil {
		return "", err
	}
	return inputBuffer.String(), nil
}
