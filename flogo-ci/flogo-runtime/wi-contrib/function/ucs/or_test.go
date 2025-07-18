package ucs

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-flogo/core/data/coerce"
	"github.com/stretchr/testify/assert"
)

var or OR
var expectedOrJSON interface{}
var left, right SimpleLookupCondition

func TestORSuccessInt(t *testing.T) {
	left = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}
	right = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 5000}

	expectedOrJSON = ComplexLookupCondition{Expr: "or", Left: left, Right: right}

	equalJSON, err := or.Eval(left, right)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedOrJSON, equalJSON)
}

func TestORSuccessString(t *testing.T) {
	left = SimpleLookupCondition{Expr: "eq", Prop: "Lastname", Val: "Morris"}
	right = SimpleLookupCondition{Expr: "eq", Prop: "Firstname", Val: "James"}

	expectedOrJSON = ComplexLookupCondition{Expr: "or", Left: left, Right: right}

	equalJSON, err := or.Eval(left, right)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedOrJSON, equalJSON)
}

func TestORSuccessFloat(t *testing.T) {
	left = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 123.44}
	right = SimpleLookupCondition{Expr: "eq", Prop: "OrderID", Val: 444.33}

	expectedOrJSON = ComplexLookupCondition{Expr: "or", Left: left, Right: right}

	equalJSON, err := or.Eval(left, right)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedOrJSON, equalJSON)
}

func TestORSuccessBool(t *testing.T) {
	left = SimpleLookupCondition{Expr: "eq", Prop: "isFullTime", Val: true}
	right = SimpleLookupCondition{Expr: "eq", Prop: "isExempted", Val: false}

	expectedOrJSON = ComplexLookupCondition{Expr: "or", Left: left, Right: right}

	equalJSON, err := or.Eval(left, right)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.Equal(t, expectedOrJSON, equalJSON)
}

func TestORIntNegative(t *testing.T) {
	left = SimpleLookupCondition{Expr: "eq", Prop: "isFullTime", Val: true}
	right = SimpleLookupCondition{Expr: "eq", Prop: "isExempted", Val: false}

	expectedOrJSON = ComplexLookupCondition{Expr: "and", Left: left, Right: right}

	equalJSON, err := or.Eval(left, right)
	assert.Nil(t, err)
	fmt.Println(equalJSON)
	assert.NotEqual(t, expectedOrJSON, equalJSON)
}

func TestORSuccessIntJSON(t *testing.T) {
	// andleft = SimpleLookupCondition{Expr: "eq", Prop: "Price", Val: 123}
	// andright = SimpleLookupCondition{Expr: "eq", Prop: "Salary", Val: 5000}

	orleft, err := coerce.ToObject("{\"expr\":\"eq\",\"prop\":\"cr163_name\",\"val\":\"test-1\"}")
	orright, err := coerce.ToObject("{\"expr\":\"neq\",\"prop\":\"lastname\",\"val\":\"testlastname\"}")

	orleftJSONString, err := json.Marshal(orleft)
	if err != nil {
		t.Error("error")
	}

	orrightJSONString, err := json.Marshal(orright)
	if err != nil {
		t.Error("error")
	}
	var orleftSimple, orrightSimple SimpleLookupCondition
	json.Unmarshal(orleftJSONString, &orleftSimple)
	json.Unmarshal(orrightJSONString, &orrightSimple)

	expectedOrJSON = ComplexLookupCondition{Expr: "or", Left: orleftSimple, Right: orrightSimple}

	orJSON, err := or.Eval(orleft, orright)
	assert.Nil(t, err)
	fmt.Println(orJSON)
	assert.Equal(t, expectedOrJSON, orJSON)
}

func TestOrSuccessNestedJSON(t *testing.T) {
	orleft := `ucs.or(ucs.and(ucs.equal("firstname", "John"), ucs.equal("lastname", "Doe")), ucs.equal("zipcode", 94566))`
	orright := `ucs.equals("firstname", "test")`

	orleftJSONString, err := json.Marshal(orleft)
	if err != nil {
		t.Error("error")
	}

	orrightJSONString, err := json.Marshal(orright)
	if err != nil {
		t.Error("error")
	}
	var orleftSimple, orrightSimple SimpleLookupCondition
	json.Unmarshal(orleftJSONString, &orleftSimple)
	json.Unmarshal(orrightJSONString, &orrightSimple)

	// expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleftSimple, Right: andrightSimple}
	errMessage := `Nested ucs.and() or ucs.or() expression found. Nested UCS functions not supported`

	orJSON, err := or.Eval(orleft, orright)
	assert.NotNil(t, err)
	assert.Equal(t, errMessage, err.Error())
	assert.Equal(t, nil, orJSON)
}

func TestOrSuccessNested1JSON(t *testing.T) {
	orleft := `ucs.or(ucs.or(ucs.equal("firstname", "John"), ucs.equal("lastname", "Doe")), ucs.equal("zipcode", 94566))`
	orright := `ucs.equals("firstname", "test")`

	orleftJSONString, err := json.Marshal(orleft)
	if err != nil {
		t.Error("error")
	}

	orrightJSONString, err := json.Marshal(orright)
	if err != nil {
		t.Error("error")
	}
	var orleftSimple, orrightSimple SimpleLookupCondition
	json.Unmarshal(orleftJSONString, &orleftSimple)
	json.Unmarshal(orrightJSONString, &orrightSimple)

	// expectedAndJSON = ComplexLookupCondition{Expr: "and", Left: andleftSimple, Right: andrightSimple}
	errMessage := `Nested ucs.and() or ucs.or() expression found. Nested UCS functions not supported`

	orJSON, err := or.Eval(orleft, orright)
	assert.NotNil(t, err)
	assert.Equal(t, errMessage, err.Error())
	assert.Equal(t, nil, orJSON)
}
