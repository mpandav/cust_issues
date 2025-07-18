package ucs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var notequal NotEqual
var expectedNotEqualJSON interface{}

func TestNotEqualSuccessInt(t *testing.T) {
	prop := "Price"
	val := 123

	expectedNotEqualJSON = SimpleLookupCondition{Expr: "neq", Prop: "Price", Val: 123}

	equalJSON, err := notequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedNotEqualJSON, equalJSON)
}

func TestNotEqualSuccessString(t *testing.T) {
	prop := "LastName"
	val := "Doe"

	expectedNotEqualJSON = SimpleLookupCondition{Expr: "neq", Prop: "LastName", Val: "Doe"}

	equalJSON, err := notequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedNotEqualJSON, equalJSON)
}

func TestNotEqualSuccessFloat(t *testing.T) {
	prop := "Salary"
	val := 5644.74

	expectedNotEqualJSON = SimpleLookupCondition{Expr: "neq", Prop: "Salary", Val: 5644.74}

	equalJSON, err := notequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedNotEqualJSON, equalJSON)
}

func TestNotEqualSuccessBool(t *testing.T) {
	prop := "emp_fulltime"
	val := true

	expectedNotEqualJSON = SimpleLookupCondition{Expr: "neq", Prop: "emp_fulltime", Val: true}

	equalJSON, err := notequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedNotEqualJSON, equalJSON)
}

func TestNotEqualIntNegative(t *testing.T) {
	prop := "Price"
	val := 123

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedGreaterThanJSON, equalJSON)
}
