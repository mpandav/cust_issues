package yukon

import (
	"testing"

	"github.com/project-flogo/core/activity"
) // Imports

// activityMetadata is the metadata of the activity as described in activity.json
var activityMetadata *activity.Metadata

func TestConnectionValid(t *testing.T) {
	// _, err := factory.NewManager(testutil.TestConnection())
	// assert.Nil(t, err)
}

func TestYukonOpenConnectionValid(t *testing.T) {
	// sharedconn, err := factory.NewManager(testutil.TestConnection())
	// assert.Nil(t, err)
	// assert.NotNil(t, sharedconn.GetConnection())
	// m := sharedconn.GetConnection().(*YukonSharedConfigManager)
	// assert.NotNil(t, m)
	// m.Start()
}

func TestYukonCloseConnectionValid(t *testing.T) {

}
