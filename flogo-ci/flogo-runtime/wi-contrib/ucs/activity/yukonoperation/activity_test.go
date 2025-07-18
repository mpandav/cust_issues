package yukonoperation

// Imports
import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/mapper"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"

	"github.com/tibco/wi-contrib/ucs/common/testutil"
	connector "github.com/tibco/wi-contrib/ucs/connector/yukon"
)

type ActivityInputs struct {
	DataObject      string             `json:"dataObject"`
	Action          string             `json:"action"`
	Input           []Input            `json:"input"`
	YukonConnection connection.Manager `json:"connection"`
}

type ActivityOutputs struct {
	Action     string    `json:"action"`
	DataObject string    `json:"dataObject"`
	Results    []Results `json:"results"`
}

var yukonConnectionManager = connector.YukonSharedConfigManager{
	ConnectionID:       "01F8Y17WQRWQJG3MJWK4TSACJX",
	InstanceID:         "01FE0RQ81JPA1AJ3DN2XH5NW0P",
	ProviderPathPrefix: "/ucs/provider",
	ConnectorName:      "MicrosoftCrm",
	ConnectionName:     "",
	YukonClient:        http.Client{Timeout: time.Duration(120) * time.Second},
	Settings:           nil,
}

var input = Input{}

// var TestCreateRequest = ActivityInputs{
// 	DataObject: "Entity2",
// 	Action:     "Create",
// 	Input: []Input{
// 		0: {LookupCondition: "test", InputData: make(map[string]interface{})},
// 	},
// 	YukonConnection: &yukonConnectionManager,
// }

// var TestBadCreateRequest = ActivityInputs{
// 	DataObject: "Entity2",
// 	Action:     "Create",
// 	InputData: []map[string]interface{}{
// 		0: {"BadField": "a value"},
// 	},
// }

// var TestUpdateRequest = ActivityInputs{
// 	DataObject: "Entity2",
// 	Action:     "Update",
// 	InputData: []map[string]interface{}{
// 		0: {
// 			"Prop1": "a value",
// 			"Prop2": "another value"},
// 	},
// }

// var TestDeleteRequest = ActivityInputs{
// 	DataObject: "Entity2",
// 	Action:     "Delete",
// 	InputData:  nil,
// }

// activityMetadata is the metadata of the activity as described in activity.json
var activityMetadata *activity.Metadata

// TestActivityRegistration checks whether the activity can be registered, and is registered in the engine
func TestActivityRegistration(t *testing.T) {
	ref := activity.GetRef(&YukonOpActivity{})
	act := activity.Get(ref)
	assert.NotNil(t, act)
}

func makeActivity(t *testing.T) activity.Activity {
	connFactory := &connector.YukonFactory{}
	connection, err := connFactory.NewManager(testutil.TestConnection())
	if connection == nil {
		fmt.Println("connection is nil")
	}
	yukonConnection := connection.GetConnection().(*connector.YukonSharedConfigManager)
	assert.Nil(t, err)
	settings := &Settings{YukonConnection: yukonConnection}

	mf := mapper.NewFactory(resolve.GetBasicResolver())
	initContext := test.NewActivityInitContext(settings, mf)
	act, err := New(initContext)
	assert.Nil(t, err)
	return act
}

func testOperation(t *testing.T, connection map[string]interface{}, activityInputs ActivityInputs) (*ActivityOutputs, error) {
	act := makeActivity(t)
	tc := test.NewActivityContext(act.Metadata())

	tc.SetInput("dataObject", activityInputs.DataObject)
	tc.SetInput("action", activityInputs.Action)
	tc.SetInput("input", activityInputs.Input)
	tc.SetInput("connection", activityInputs.YukonConnection)

	// Execute the activity
	_, err := act.Eval(tc)

	if err == nil {
		results, ok := tc.GetOutput("results").([]Results)
		if !ok {
			return nil, err
		}
		outputs := ActivityOutputs{
			Action:     tc.GetOutput("action").(string),
			DataObject: tc.GetOutput("dataObject").(string),
			Results:    results,
		}

		return &outputs, nil
	}

	outputs := ActivityOutputs{}
	return &outputs, err
}

// working test case commented since during build build agent cannot resolve environment variables needed for connecting to provider
// func TestConnectionOpSimpleCreate(t *testing.T) {
// 	outputs, err := testOperation(t, testutil.TestConnection(), TestCreateRequest)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, outputs.Action)
// 	assert.NotNil(t, outputs.DataObject)
// 	assert.NotNil(t, outputs.Results)
// 	v := reflect.ValueOf(outputs.Results[0])
// 	//resultValues := make([]Results, v.NumField())
// 	// for i := 0; i < v.NumField(); i++ {
// 	objAffected := v.Field(1).Interface()
// 	outputDataResult := v.Field(2).Interface()
// 	successValue := v.Field(3).Interface()
// 	fmt.Println("Object Affected: ", objAffected)
// 	fmt.Println("Success: ", successValue)
// 	assert.NotNil(t, objAffected)
// 	assert.NotNil(t, outputDataResult)
// 	assert.Equal(t, true, successValue)
// }
