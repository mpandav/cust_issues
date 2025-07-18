package ucs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var lessthan LessThan
var expectedLessThanJSON interface{}

func TestLessThanSuccessInt(t *testing.T) {
	prop := "Price"
	val := 123

	expectedLessThanJSON = SimpleLookupCondition{Expr: "lt", Prop: "Price", Val: 123}

	equalJSON, err := lessthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanJSON, equalJSON)
}

func TestLessThanSuccessString(t *testing.T) {
	prop := "LastName"
	val := "Doe"

	expectedLessThanJSON = SimpleLookupCondition{Expr: "lt", Prop: "LastName", Val: "Doe"}

	equalJSON, err := lessthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanJSON, equalJSON)
}

func TestLessThanSuccessFloat(t *testing.T) {
	prop := "Salary"
	val := 5644.74

	expectedLessThanJSON = SimpleLookupCondition{Expr: "lt", Prop: "Salary", Val: 5644.74}

	equalJSON, err := lessthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanJSON, equalJSON)
}

func TestLessThanSuccessBool(t *testing.T) {
	prop := "emp_fulltime"
	val := true

	expectedLessThanJSON = SimpleLookupCondition{Expr: "lt", Prop: "emp_fulltime", Val: true}

	equalJSON, err := lessthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedLessThanJSON, equalJSON)
}

func TestLessThanIntNegative(t *testing.T) {
	prop := "Price"
	val := 123

	expectedLessThanJSON = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}

	equalJSON, err := lessthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedLessThanJSON, equalJSON)
}
