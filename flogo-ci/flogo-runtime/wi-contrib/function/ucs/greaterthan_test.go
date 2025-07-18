package ucs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var greaterthan GreaterThan
var expectedGreaterThanJSON interface{}

func TestGreaterThanSuccessInt(t *testing.T) {
	prop := "Price"
	val := 123

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "gt", Prop: "Price", Val: 123}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanJSON, equalJSON)
}

func TestGreaterThanSuccessString(t *testing.T) {
	prop := "LastName"
	val := "Doe"

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "gt", Prop: "LastName", Val: "Doe"}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanJSON, equalJSON)
}

func TestGreaterThanSuccessFloat(t *testing.T) {
	prop := "Salary"
	val := 5644.74

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "gt", Prop: "Salary", Val: 5644.74}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanJSON, equalJSON)
}

func TestGreaterThanSuccessBool(t *testing.T) {
	prop := "emp_fulltime"
	val := true

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "gt", Prop: "emp_fulltime", Val: true}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanJSON, equalJSON)
}

func TestGreaterThanIntNegative(t *testing.T) {
	prop := "Price"
	val := 123

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedGreaterThanJSON, equalJSON)
}
