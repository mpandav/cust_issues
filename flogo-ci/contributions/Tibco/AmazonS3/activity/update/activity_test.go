package update

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
	_ "github.com/tibco/flogo-aws/src/app/AWS/connector"
)

var activityMetadata *activity.Metadata

const (
	BucketName       = "awakchau-s3-gotest"
	OwnerCanonicalID = "1826704023be699bab6815fbeb8415f34d20430413f2e76cfdf5c91a24e3c578"
	ExistingObject   = "hello"
)

func getConnectionManager() interface{} {
	connectionBytes, err := ioutil.ReadFile("../connectionData.json")
	if err != nil {
		panic("connectionData.json file found")
	}
	var connectionObj map[string]interface{}
	json.Unmarshal(connectionBytes, &connectionObj)
	support.RegisterAlias("connection", "connector", "git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/AWS/connector")
	connmgr, _ := coerce.ToConnection(connectionObj)
	return connmgr
}

func setupActivity(t *testing.T) (*Activity, *test.TestActivityContext) {
	act := &Activity{}
	tc := test.NewActivityContext(act.Metadata())
	tc.SetInput("connection", getConnectionManager())
	return act, tc
}

func logOutput(t *testing.T, tc *test.TestActivityContext) {
	output := tc.GetOutput("output")
	assert.NotNil(t, output)
	outputBytes, err := json.Marshal(output)
	assert.Nil(t, err)
	tc.Logger().Info("output:", string(outputBytes))
}

func TestUpdateBucketACL(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	tc.SetInput("updateType", "acl")

	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
		"AccessControlPolicy": {
			"Grants": [ 
        {
					"Grantee": {
						"Type": "CanonicalUser",
						"ID": "` + OwnerCanonicalID + `"
					}, 
					"Permission": "FULL_CONTROL"
        },
				{
					"Grantee": {
						"Type": "Group",
						"URI": "http://acs.amazonaws.com/groups/s3/LogDelivery"
					},
					"Permission": "READ_ACP"
				}
			],
			"Owner": {
				"ID": "` + OwnerCanonicalID + `"
			}
		}
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to update bucket ACL due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestUpdateBucketACLShortHand(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	tc.SetInput("updateType", "acl")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
        "Bucket": "` + BucketName + `",
				"ACL": "private"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to update bucket ACL due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestUpdateBucketCORS(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	tc.SetInput("updateType", "cors")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
		"CORSConfiguration": {
			"CORSRules": [
				{
					"AllowedMethods": [ "GET" ],
					"AllowedOrigins": [ "*" ]
				}
			]
		}
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to update bucket CORS due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

// This test will fail as Principal is invalid
func TestUpdateBucketPolicy(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	tc.SetInput("updateType", "policy")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
    "Policy": "{  \"Version\":\"2012-10-17\",  \"Statement\":[    {      \"Sid\":\"AddPerm\",      \"Effect\":\"Allow\",      \"Principal\": {\"AWS\": [\"arn:aws:iam::336528382084:user/pukumar\"]},      \"Action\":[\"s3:GetObject\"],      \"Resource\":[\"arn:aws:s3:::` + BucketName + `/*\"]    }  ]}"
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	if assert.NotNil(t, err) {
		tc.Logger().Info(err.Error())
	}
}

func TestUpdateBucketVersioning(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	tc.SetInput("updateType", "versioning")

	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
    "VersioningConfiguration": {
			"Status": "Enabled"
		}
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to update bucket versioning due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestUpdateBucketWebsite(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "bucket")
	tc.SetInput("updateType", "website")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
    "WebsiteConfiguration": {
			"ErrorDocument": {
				"Key": "error.html"
			},
			"IndexDocument": {
				"Suffix": "index.html"
			}
		}
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to update bucket website due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestUpdateObjectACL(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "object")
	tc.SetInput("updateType", "acl")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
		"Key": "` + ExistingObject + `",
		"AccessControlPolicy": {
			"Grants": [
				{
					"Grantee": {
						"Type": "CanonicalUser", 
						"ID": "` + OwnerCanonicalID + `"
					}, 
					"Permission": "FULL_CONTROL"
				},
				{
					"Grantee": {
						"Type": "AmazonCustomerByEmail",
						"EmailAddress": "pukumar@tibco.com"
					},
					"Permission": "READ"
				}
			],
			"Owner": {
				"ID": "` + OwnerCanonicalID + `"
			}
		}
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Update Object ACL due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}

func TestUpdateObjectTaggin(t *testing.T) {
	act, tc := setupActivity(t)
	tc.SetInput("serviceName", "object")
	tc.SetInput("updateType", "tagging")
	var inputParams map[string]interface{}
	var inputJSON = []byte(`{
		"Bucket": "` + BucketName + `",
		"Key": "` + ExistingObject + `",
		"Tagging": {
			"TagSet": [
				{
					"Key": "tag1",
					"Value": "value1"
				},
				{
					"Key": "tag2",
					"Value": "value2"
				}
			]
		}
	}`)
	err := json.Unmarshal(inputJSON, &inputParams)
	assert.Nil(t, err)
	tc.SetInput("input", inputParams)
	_, err = act.Eval(tc)
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("Unable to Update Object Tagging due to %s", err.Error())
		t.Fail()
	}
	logOutput(t, tc)
}
