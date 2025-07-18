package ucs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var lessthanequal LessThanEqual
var expectedLessThanEqualJSON interface{}

func TestLessThanEqualSuccessInt(t *testing.T) {
	prop := "Price"
	val := 123

	expectedLessThanEqualJSON = SimpleLookupCondition{Expr: "lte", Prop: "Price", Val: 123}

	equalJSON, err := lessthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanEqualJSON, equalJSON)
}

func TestLessThanEqualSuccessString(t *testing.T) {
	prop := "LastName"
	val := "Doe"

	expectedLessThanEqualJSON = SimpleLookupCondition{Expr: "lte", Prop: "LastName", Val: "Doe"}

	equalJSON, err := lessthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanEqualJSON, equalJSON)
}

func TestLessThanEqualSuccessFloat(t *testing.T) {
	prop := "Salary"
	val := 5644.74

	expectedLessThanEqualJSON = SimpleLookupCondition{Expr: "lte", Prop: "Salary", Val: 5644.74}

	equalJSON, err := lessthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanEqualJSON, equalJSON)
}

func TestLessThanEqualSuccessBool(t *testing.T) {
	prop := "emp_fulltime"
	val := true

	expectedLessThanEqualJSON = SimpleLookupCondition{Expr: "lte", Prop: "emp_fulltime", Val: true}

	equalJSON, err := lessthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanEqualJSON, equalJSON)
}

func TestLessThanEqualIntNegative(t *testing.T) {
	prop := "Price"
	val := 123

	expectedLessThanEqualJSON = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}

	equalJSON, err := lessthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedLessThanEqualJSON, equalJSON)
}
