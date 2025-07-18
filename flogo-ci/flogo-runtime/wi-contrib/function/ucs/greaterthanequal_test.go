package ucs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var greaterthanequal GreaterThanEqual
var expectedGreaterThanEqualJSON interface{}

func TestGreaterThanEqualSuccessInt(t *testing.T) {
	prop := "Price"
	val := 123

	expectedGreaterThanEqualJSON = SimpleLookupCondition{Expr: "gte", Prop: "Price", Val: 123}

	equalJSON, err := greaterthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanEqualJSON, equalJSON)
}

func TestGreaterThanEqualSuccessString(t *testing.T) {
	prop := "LastName"
	val := "Doe"

	expectedGreaterThanEqualJSON = SimpleLookupCondition{Expr: "gte", Prop: "LastName", Val: "Doe"}

	equalJSON, err := greaterthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanEqualJSON, equalJSON)
}

func TestGreaterThanEqualSuccessFloat(t *testing.T) {
	prop := "Salary"
	val := 5644.74

	expectedGreaterThanEqualJSON = SimpleLookupCondition{Expr: "gte", Prop: "Salary", Val: 5644.74}

	equalJSON, err := greaterthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanEqualJSON, equalJSON)
}

func TestGreaterThanEqualSuccessBool(t *testing.T) {
	prop := "emp_fulltime"
	val := true

	expectedGreaterThanEqualJSON = SimpleLookupCondition{Expr: "gte", Prop: "emp_fulltime", Val: true}

	equalJSON, err := greaterthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedGreaterThanEqualJSON, equalJSON)
}

func TestGreaterThanEqualIntNegative(t *testing.T) {
	prop := "Price"
	val := 123

	expectedGreaterThanEqualJSON = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}

	equalJSON, err := greaterthanequal.Eval(prop, val)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedGreaterThanEqualJSON, equalJSON)
}
