package number

import (
	"fmt"
	"github.com/project-flogo/core/data/expression/function"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ln = &Len{}

func init() {
	function.ResolveAliases()
}

func TestLenSample(t *testing.T) {
	final, _ := ln.Eval("123")
	assert.Equal(t, int(3), final)
}

func TestLenExpression(t *testing.T) {
	fun, err := factory.NewExpr(`number.len("TIBCO NAME")`)
	assert.Nil(t, err)
	assert.NotNil(t, fun)
	v, err := fun.Eval(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v)
	fmt.Println(v)
}
