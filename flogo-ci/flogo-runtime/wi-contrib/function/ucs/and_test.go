package ucs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-flogo/core/data/coerce"
	"github.com/stretchr/testify/assert"
)

var and AND
var expectedAndJSON interface{}
var andleft, andright SimpleLookupCondition

func TestAndSuccessInt(t *testing.T) {
	andleft = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}
	andright = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 5000}

	expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleft, Right: andright}

	equalJSON, err := and.Eval(andleft, andright)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedAndJSON, equalJSON)
}

func TestAndSuccessString(t *testing.T) {
	andleft = SimpleLookupCondition{Expr: "eq", Prop: "Lastname", Val: "Morris"}
	andright = SimpleLookupCondition{Expr: "eq", Prop: "Firstname", Val: "James"}

	expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleft, Right: andright}

	equalJSON, err := and.Eval(andleft, andright)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedAndJSON, equalJSON)
}

func TestAndSuccessFloat(t *testing.T) {
	andleft = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 123.44}
	andright = SimpleLookupCondition{Expr: "eq", Prop: "OrderID", Val: 444.33}

	expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleft, Right: andright}

	equalJSON, err := and.Eval(andleft, andright)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedAndJSON, equalJSON)
}

func TestAndSuccessBool(t *testing.T) {
	andleft = SimpleLookupCondition{Expr: "eq", Prop: "isFullTime", Val: true}
	andright = SimpleLookupCondition{Expr: "eq", Prop: "isExempted", Val: false}

	expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleft, Right: andright}

	equalJSON, err := and.Eval(andleft, andright)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedAndJSON, equalJSON)
}

func TestAndIntNegative(t *testing.T) {
	andleft = SimpleLookupCondition{Expr: "eq", Prop: "isFullTime", Val: true}
	andright = SimpleLookupCondition{Expr: "eq", Prop: "isExempted", Val: false}

	expectedAndJSON = ComplexLookupCondition{Expr: "or", Left: andleft, Right: andright}

	equalJSON, err := and.Eval(andleft, andright)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedAndJSON, equalJSON)
}

func TestAndSuccessIntJSON(t *testing.T) {
	// andleft = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}
	// andright = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 5000}

	andleft, err := coerce.ToObject("{\"expr\":\"eq\",\"prop\":\"cr163_name\",\"val\":\"test-1\"}")
	andright, err := coerce.ToObject("{\"expr\":\"neq\",\"prop\":\"lastname\",\"val\":\"testlastname\"}")

	andleftJSONString, err := json.Marshal(andleft)
	if err != nil {
		t.Error("error")
	}

	andrightJSONString, err := json.Marshal(andright)
	if err != nil {
		t.Error("error")
	}
	var andleftSimple, andrightSimple SimpleLookupCondition
	json.Unmarshal(andleftJSONString, &andleftSimple)
	json.Unmarshal(andrightJSONString, &andrightSimple)

	expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleftSimple, Right: andrightSimple}

	andJSON, err := and.Eval(andleft, andright)
	assert.Nil(t, err)
	fmt.Println(andJSON)
	assert.Equal(t, expectedAndJSON, andJSON)
}

func TestAndSuccessNestedJSON(t *testing.T) {
	andleft := `ucs.and(ucs.or(ucs.equal("firstname", "John"), ucs.equal("lastname", "Doe")), ucs.equal("zipcode", 94566))`
	andright := `ucs.equals("firstname", "test")`

	andleftJSONString, err := json.Marshal(andleft)
	if err != nil {
		t.Error("error")
	}

	andrightJSONString, err := json.Marshal(andright)
	if err != nil {
		t.Error("error")
	}
	var andleftSimple, andrightSimple SimpleLookupCondition
	json.Unmarshal(andleftJSONString, &andleftSimple)
	json.Unmarshal(andrightJSONString, &andrightSimple)

	// expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleftSimple, Right: andrightSimple}
	errMessage := `Nested ucs.and() or ucs.or() expression found. Nested UCS functions not supported`

	andJSON, err := and.Eval(andleft, andright)
	assert.NotNil(t, err)
	assert.Equal(t, errMessage, err.Error())
	assert.Equal(t, nil, andJSON)
}

func TestAndSuccessNested1JSON(t *testing.T) {
	andleft := `ucs.and(ucs.and(ucs.equal("firstname", "John"), ucs.equal("lastname", "Doe")), ucs.equal("zipcode", 94566))`
	andright := `ucs.equals("firstname", "test")`

	andleftJSONString, err := json.Marshal(andleft)
	if err != nil {
		t.Error("error")
	}

	andrightJSONString, err := json.Marshal(andright)
	if err != nil {
		t.Error("error")
	}
	var andleftSimple, andrightSimple SimpleLookupCondition
	json.Unmarshal(andleftJSONString, &andleftSimple)
	json.Unmarshal(andrightJSONString, &andrightSimple)

	// expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleftSimple, Right: andrightSimple}
	errMessage := `Nested ucs.and() or ucs.or() expression found. Nested UCS functions not supported`

	andJSON, err := and.Eval(andleft, andright)
	assert.NotNil(t, err)
	assert.Equal(t, errMessage, err.Error())
	assert.Equal(t, nil, andJSON)
}
