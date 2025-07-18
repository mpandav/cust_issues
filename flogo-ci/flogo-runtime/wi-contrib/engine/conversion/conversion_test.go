package conversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToInt64(t *testing.T) {
	s := "34"
	i, err := ConvertToInt64(s)
	assert.Nil(t, err)
	assert.Equal(t, int64(34), i)
}

func TestConvertToString(t *testing.T) {
	s := 34
	i, err := ConvertToString(s)
	assert.Nil(t, err)
	assert.Equal(t, "34", i)
}
