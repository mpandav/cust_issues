package ucs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var equal Equal
var expectedEqualJSON interface{}

func TestEqualSuccessInt(t *testing.T) {
	prop := "Price"
	val := 123

	expectedEqualJSON = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}

	equalJSON, err := equal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedEqualJSON, equalJSON)
}

func TestEqualSuccessString(t *testing.T) {
	prop := "LastName"
	val := "Doe"

	expectedEqualJSON = SimpleLookupCondition{Expr: "eq", Prop: "LastName", Val: "Doe"}

	equalJSON, err := equal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedEqualJSON, equalJSON)
}

func TestEqualSuccessFloat(t *testing.T) {
	prop := "Salary"
	val := 5644.74

	expectedEqualJSON = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 5644.74}

	equalJSON, err := equal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedEqualJSON, equalJSON)
}

func TestEqualSuccessBool(t *testing.T) {
	prop := "emp_fulltime"
	val := true

	expectedEqualJSON = SimpleLookupCondition{Expr: "eq", Prop: "emp_fulltime", Val: true}

	equalJSON, err := equal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedEqualJSON, equalJSON)
}

func TestEqualIntNegative(t *testing.T) {
	prop := "Price"
	val := 123

	expectedGreaterThanJSON = SimpleLookupCondition{Expr: "neq", Prop: "Price", Val: 123}

	equalJSON, err := greaterthan.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedGreaterThanJSON, equalJSON)
}
