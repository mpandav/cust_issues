package utility

import (
	"fmt"
	"github.com/project-flogo/core/data/expression/function"
	"github.com/project-flogo/core/data/expression/script"
	"github.com/project-flogo/core/data/resolve"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var resolver = resolve.NewCompositeResolver(map[string]resolve.Resolver{"static": nil, ".": nil, "env": &resolve.EnvResolver{}})
var factory = script.NewExprFactory(resolver)

func init() {
	function.ResolveAliases()
}

var s = &RenderJson{}

func TestMain(m *testing.M) {
	function.ResolveAliases()
	retCode := m.Run()
	os.Exit(retCode)
}

func TestStaticFunc_RenderJsonStruct(t *testing.T) {
	type Person struct {
		FirstName, LastName string
	}
	person := &Person{FirstName: "Jon", LastName: "Snow"}
	final, err := s.Eval(person, false)
	assert.Nil(t, err)
	fmt.Println(final)
	assert.Equal(t, final, `{"FirstName":"Jon","LastName":"Snow"}`)
}

func TestStaticFunc_RenderJsonMap(t *testing.T) {
	person := make(map[string]string)
	person["FirstName"] = "Jon"
	person["LastName"] = "Snow"
	final, err := s.Eval(person, false)
	assert.Nil(t, err)
	fmt.Println(final)
	assert.Equal(t, final, `{"FirstName":"Jon","LastName":"Snow"}`)
}

func TestExpression(t *testing.T) {
	fun, err := factory.NewExpr(`utility.renderJSON("{}", false)`)
	assert.Nil(t, err)
	assert.NotNil(t, fun)
	v, err := fun.Eval(nil)
	assert.Nil(t, err)
	assert.NotNil(t, v)
	fmt.Println(v)
}
