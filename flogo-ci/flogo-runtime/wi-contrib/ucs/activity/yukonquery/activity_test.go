package yukonquery

// // Imports
// import (
// 	"encoding/json"
// 	"fmt"
// 	"testing"

// 	"github.com/project-flogo/core/activity"
// 	"github.com/project-flogo/core/data/mapper"
// 	"github.com/project-flogo/core/data/resolve"
// 	"github.com/project-flogo/core/support/test"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/tibco/wi-contrib/ucs/common/testutil"
// 	connector "github.com/tibco/wi-contrib/ucs/connector/yukon"
// )

// // activityMetadata is the metadata of the activity as described in activity.json
// var activityMetadata *activity.Metadata

// // TestActivityRegistration checks whether the activity can be registered, and is registered in the engine
// func TestActivityRegistration(t *testing.T) {
// 	ref := activity.GetRef(&YukonQueryActivity{})
// 	act := activity.Get(ref)
// 	assert.NotNil(t, act)
// }

// func makeActivity(t *testing.T) activity.Activity {
// 	connFactory := &connector.YukonFactory{}
// 	connection, err := connFactory.NewManager(testutil.TestConnection())
// 	assert.Nil(t, err)
// 	settings := map[string]interface{}{"yukonConnection": connection}

// 	mf := mapper.NewFactory(resolve.GetBasicResolver())
// 	initContext := test.NewActivityInitContext(settings, mf)
// 	act, err := New(initContext)
// 	assert.Nil(t, err)
// 	return act
// }

// func testQuery(t *testing.T, query string, params map[string]interface{}) ([]interface{}, bool) {

// 	act := makeActivity(t)
// 	tc := test.NewActivityContext(act.Metadata())

// 	//setup attrs
// 	tc.SetInput("query", query)

// 	if len(params) > 0 {
// 		tc.SetInput("params", params)
// 	}

// 	act.Eval(tc)

// 	assert.NotNil(t, tc.GetOutput("eof"))
// 	eof := tc.GetOutput("eof").(bool)

// 	assert.NotNil(t, tc.GetOutput("results"))
// 	results := tc.GetOutput("results").([]interface{})

// 	return results, eof
// }

// func TestEvalSimpleSelectAll(t *testing.T) {
// 	results, eof := testQuery(t, "select * from entity2", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 250)
// }

// func TestEvalSimpleSelectTop10(t *testing.T) {

// 	results, eof := testQuery(t, "select top 10 * from entity2", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 10)
// }

// func TestEvalSimpleSelectSkip10(t *testing.T) {

// 	results, eof := testQuery(t, "select skip 10 * from entity2", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 250)

// 	firstResult := results[0].(map[string]interface{})
// 	firstIndex := firstResult["Index"].(float64)
// 	assert.True(t, firstIndex == 11)
// }

// func TestEvalSimpleSelectTop10Skip10(t *testing.T) {

// 	results, eof := testQuery(t, "select top 10 skip 10 * from entity2", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 10)

// 	firstResult := results[0].(map[string]interface{})
// 	firstIndex := firstResult["Index"].(float64)
// 	assert.True(t, firstIndex == 11)
// }

// func TestEvalSimpleSelect2Columns(t *testing.T) {

// 	results, eof := testQuery(t, "select index, prop1 from entity2", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 250)
// }

// func TestEvalSelectWithSimpleWhere(t *testing.T) {

// 	results, eof := testQuery(t, "select index, prop1 from entity2 where index < 10", map[string]interface{}{})
// 	assert.True(t, eof)
// 	assert.True(t, len(results) == 10)
// }

// func TestEvalSelectWithWhereWithAnd(t *testing.T) {

// 	results, eof := testQuery(t, "select * from entity2 where index < 10 and prop2 != 'xxxxxxx'", map[string]interface{}{})
// 	assert.True(t, eof)
// 	assert.True(t, len(results) == 0) // looks to be a benchmark connector issue?
// }

// func TestEvalSelectWithWhereWithAndMixedCase(t *testing.T) {

// 	results, eof := testQuery(t, "SELECT * FROM Entity2 WHERE Index < 10 AND Prop2 != 'xxxxxxx'", map[string]interface{}{})
// 	assert.True(t, eof)
// 	assert.True(t, len(results) == 0) // looks to be a benchmark connector issue?
// }

// func TestEvalSimpleSelectWithOrderBy(t *testing.T) {

// 	results, eof := testQuery(t, "select * from entity2 orderby index", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 250)

// 	firstResult := results[0].(map[string]interface{})
// 	firstIndex := firstResult["Index"].(float64)
// 	assert.True(t, firstIndex == 1) // benchmark does not support orderby
// }

// func TestEvalSimpleSelectWithOrderByAsc(t *testing.T) {

// 	results, eof := testQuery(t, "select * from entity2 orderby index asc", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 250)

// 	firstResult := results[0].(map[string]interface{})
// 	firstIndex := firstResult["Index"].(float64)
// 	assert.True(t, firstIndex == 1) // benchmark does not support orderby
// }

// func TestEvalSimpleSelectWithOrderByDesc(t *testing.T) {

// 	results, eof := testQuery(t, "select * from entity2 orderby index desc", map[string]interface{}{})
// 	assert.False(t, eof)
// 	assert.True(t, len(results) == 250)

// 	firstResult := results[0].(map[string]interface{})
// 	firstIndex := firstResult["Index"].(float64)
// 	assert.True(t, firstIndex == 1) // benchmark does not support orderby
// }

// func TestEvalNoQuery(t *testing.T) {

// 	act := makeActivity(t)
// 	tc := test.NewActivityContext(act.Metadata())

// 	// Specify the input values for the activity
// 	tc.SetInput("yukonConnection", testutil.TestConnection())
// 	tc.SetInput("query", "")

// 	// Execute the activity
// 	done, err := act.Eval(tc)
// 	assert.False(t, done)
// 	assert.NotNil(t, err)
// }

// func TestEvalBadTableName(t *testing.T) {

// 	act := makeActivity(t)
// 	tc := test.NewActivityContext(act.Metadata())

// 	// Specify the input values for the activity
// 	tc.SetInput("yukonConnection", testutil.TestConnection())
// 	tc.SetInput("query", "select * from BadTableName")

// 	// Execute the activity
// 	done, err := act.Eval(tc)
// 	assert.False(t, done)
// 	assert.NotNil(t, err)
// }

// func TestBasicParseQuery(t *testing.T) {

// 	// basic select * query
// 	basicquery, err := parseQuery("select * from entity2", nil)
// 	assert.Nil(t, err)
// 	jsonquery, err := json.Marshal(basicquery)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(string(jsonquery))
// }
// func TestComplexParseQuery(t *testing.T) {

// 	// basic select * query
// 	basicquery, err := parseQuery("Select top  100  skip  100  index , prop1 from  entity2 where index < 5 OR prop1 == 'xxxxx'  orderby index   desc  ", nil)
// 	assert.Nil(t, err)
// 	jsonquery, err := json.Marshal(basicquery)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(string(jsonquery))
// }
// func TestParseQuery(t *testing.T) {

// 	// messy select * query
// 	_, err := parseQuery(" select  *  from  entity2 ", nil)
// 	assert.Nil(t, err)

// 	// select column query
// 	_, err = parseQuery("select index from entity2", nil)
// 	assert.Nil(t, err)

// 	// select columns query
// 	_, err = parseQuery("select index, prop1 from entity2", nil)
// 	assert.Nil(t, err)

// 	// select * query with top
// 	_, err = parseQuery("select top 100 * from entity2", nil)
// 	assert.Nil(t, err)

// 	// select * query with skip first
// 	_, err = parseQuery("select skip 100 * from entity2", nil)
// 	assert.Nil(t, err)

// 	// select * query with skip last
// 	_, err = parseQuery("select * from entity2 skip 100", nil)
// 	assert.Nil(t, err)

// 	// select * query with top and skip
// 	_, err = parseQuery("select top 100 skip 100 * from entity2", nil)
// 	assert.Nil(t, err)

// 	// select * query with where
// 	_, err = parseQuery("select * from entity2 where index < 5", nil)
// 	assert.Nil(t, err)

// 	// select * query with where with and and or
// 	_, err = parseQuery("select * from entity2 where index < 5 or prop1 == 'xxxxx'", nil)
// 	assert.Nil(t, err)

// 	// select * query with orderby
// 	_, err = parseQuery("select * from entity2 orderby index", nil)
// 	assert.Nil(t, err)

// 	// select * query with orderby asc
// 	_, err = parseQuery("select * from entity2 orderby index asc", nil)
// 	assert.Nil(t, err)

// 	// select * query with orderby desc
// 	_, err = parseQuery("select * from entity2 orderby index desc", nil)
// 	assert.Nil(t, err)

// 	// messy big fat pig query
// 	_, err = parseQuery(" Select top  100  skip  100  index , prop1 from  entity2 where index < 5 or prop1 == 'xxxxx'  orderby index   desc  ", nil)
// 	assert.Nil(t, err)

// 	// blank query
// 	_, err = parseQuery("", nil)
// 	assert.NotNil(t, err)

// 	// only select
// 	_, err = parseQuery("select", nil)
// 	assert.NotNil(t, err)

// 	// no columns
// 	_, err = parseQuery("select from entity2", nil)
// 	assert.NotNil(t, err)

// 	// no from
// 	_, err = parseQuery("select *", nil)
// 	assert.NotNil(t, err)

// 	// no table
// 	_, err = parseQuery("select * from", nil)
// 	assert.NotNil(t, err)

// 	// no where values
// 	_, err = parseQuery("select * from entity2 where", nil)
// 	assert.NotNil(t, err)

// 	// no orderby values
// 	_, err = parseQuery("select * from entity2 orderby", nil)
// 	assert.NotNil(t, err)
// }

// func TestBuildWherePart(t *testing.T) {

// 	// valid
// 	_, err := buildWherePart("a", "=", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "==", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "!=", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "<>", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", ">", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "<", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", ">=", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "<=", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "!>", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "!<", "b", "")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "=", "b", "and")
// 	assert.Nil(t, err)

// 	_, err = buildWherePart("a", "=", "b", "or")
// 	assert.Nil(t, err)

// 	// invalid
// 	_, err = buildWherePart("", "", "", "")
// 	assert.NotNil(t, err)

// 	_, err = buildWherePart("a", "", "", "")
// 	assert.NotNil(t, err)

// 	_, err = buildWherePart("a", "=", "", "")
// 	assert.NotNil(t, err)

// 	_, err = buildWherePart("a", "", "b", "")
// 	assert.NotNil(t, err)

// 	_, err = buildWherePart("a", "??", "b", "")
// 	assert.NotNil(t, err)

// 	_, err = buildWherePart("a", "=", "b", "??")
// 	assert.NotNil(t, err)
// }
