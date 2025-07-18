package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/expression/function"
)

type GzipUncompress struct {
}

func init() {
	function.Register(&GzipUncompress{})
}

func (g *GzipUncompress) Name() string {
	return "gzipUncompress"
}

func (g *GzipUncompress) GetCategory() string {
	return "compression"
}

func (g *GzipUncompress) Sig() (paramTypes []data.Type, isVariadic bool) {
	return []data.Type{data.TypeString}, false
}

func (s *GzipUncompress) Eval(params ...interface{}) (interface{}, error) {

	inputStr, err := coerce.ToString(params[0])
	if err != nil {
		return nil, fmt.Errorf("gzipUncompress function argument must be string")
	}
	// log.RootLogger().Debugf("Decode base64 to string \"%s\"", params[0])
	bReader := bytes.NewReader([]byte(inputStr))
	gReader, err := gzip.NewReader(bReader)
	if err != nil {
		return "", err
	}
	defer gReader.Close()

	uncompressedStr, err := io.ReadAll(gReader)
	if err != nil {
		return "", err
	}

	return string(uncompressedStr), nil
}
