package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/project-flogo/core/activity"

	salesforce "github.com/tibco/wi-salesforce/src/app/Salesforce/activity"
	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
)

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&QueryActivity{}, New)
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &QueryActivity{}, nil
}

type QueryActivity struct {
}

func (*QueryActivity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *QueryActivity) Eval(context activity.Context) (done bool, err error) {
	context.Logger().Debug("Executing Salesforce Query Activity")
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}
	sscm, _ := input.SalesforceConnection.(*sfconnection.SalesforceSharedConfigManager)

	objectName := input.ObjectName
	if objectName == "" {
		return false, errors.New("Object name is required")
	}
	context.Logger().Debugf("objectName is %s", objectName)

	queryType := input.QueryType
	context.Logger().Debugf("queryType is %s", queryType)

	sql := input.Query
	if sql == "" {
		return false, fmt.Errorf("Input query string is empty")
	}

	queryObject, err := query(context, objectName, queryType, sql, sscm)
	if err != nil {
		return false, err // err here has been set with code
	}

	context.Logger().Debug("Operation success, set output")
	output := &Output{}
	output.Output = queryObject
	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("Salesforce Query Activity successfully executed")
	return true, nil
}

func query(context activity.Context, objectName, queryType, sql string, sscm *sfconnection.SalesforceSharedConfigManager) (map[string]interface{}, error) {

	token := sscm.SalesforceToken
	if token == nil {
		return nil, fmt.Errorf("Salesforce token is not configured for connection %s", sscm.ConnectionName)
	}
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	//sql = strings.Replace(sql, " ", "+", -1)
	sql = url.QueryEscape(sql)
	queryurl := ""

	if queryType == "" || queryType == "Query" {
		queryurl = token.InstanceUrl + "/services/data/" + apiVersion + "/query?q=" + sql
	} else {
		queryurl = token.InstanceUrl + "/services/data/" + apiVersion + "/queryAll?q=" + sql
	}
	//u, err := url.Parse(queryurl)
	//u.RawQuery = u.Query().Encode()

	resBody, err := salesforce.RestCall(sscm, "GET", queryurl, nil, context.Logger())
	if err != nil {
		return nil, err
	}

	response := &Response{}
	err = json.Unmarshal(resBody, response)
	if err != nil {
		return nil, fmt.Errorf("Fail to unmarshal query response: %+v", err)
	}

	var records []map[string]interface{}
	tmpRecords := ConvertResponseToRecords(response)

	innerObj := make(map[string]interface{})
	//innerobj := Object{}
	innerObj["done"] = true
	innerObj["totalSize"] = response.TotalSize
	records = append(records, tmpRecords...)

	if !response.Done && response.NextRecordsaUrl != "" {
		context.Logger().Debugf("Total records is %d, do query more...", response.TotalSize)
		var moreRecords []map[string]interface{}
		moreRecords, err = QueryMore(context, response.NextRecordsaUrl, moreRecords, sscm)
		if err != nil {
			return nil, fmt.Errorf("Fail to do Salesforce query more: %s", err.Error())
		}

		records = append(records, moreRecords...)
	}

	innerObj["records"] = records

	outputObj := make(map[string]interface{})
	outputObj[objectName] = innerObj

	return outputObj, nil
}

// QueryMore nextRecordsUrl: /services/data/v52.0/query/01g900000CL3t1UAQR-4000"
func QueryMore(context activity.Context, nexRecordstUrl string, records []map[string]interface{}, sscm *sfconnection.SalesforceSharedConfigManager) ([]map[string]interface{}, error) {

	context.Logger().Debugf("Do query more for next records with url %s", nexRecordstUrl)

	reqUrl := sscm.SalesforceToken.InstanceUrl + nexRecordstUrl
	resBody, err := salesforce.RestCall(sscm, "GET", reqUrl, nil, context.Logger())
	if err != nil {
		return nil, err
	}
	response := &Response{}
	err = json.Unmarshal(resBody, response)
	if err != nil {
		return nil, fmt.Errorf("Fail to unmarshal query more response: %s", err.Error())
	}

	tmp := ConvertResponseToRecords(response)

	records = append(records, tmp...)

	if !response.Done && response.NextRecordsaUrl != "" {
		return QueryMore(context, response.NextRecordsaUrl, records, sscm)
	}

	return records, nil

}

func ConvertResponseToRecords(response *Response) []map[string]interface{} {

	// innerObj := make(map[string]interface{})
	// //innerobj := Object{}
	// innerObj["done"] = response.Done
	// innerObj["totalSize"] = response.TotalSize

	tmpRecords := make([]map[string]interface{}, len(response.Records))
	for i, value := range response.Records {
		if value != nil {
			tmpMap := map[string]interface{}{}
			for k, v := range value {
				if strings.EqualFold("attributes", k) {
					//ignore
				} else if v != nil {
					switch v.(type) {
					case string:
						tmpMap[k] = v
					case map[string]interface{}:
						m := v.(map[string]interface{})
						tmp := map[string]interface{}{}
						for vk, vv := range m {
							if vv != nil {
								tmp[vk] = vv
							}
						}
						tmpMap[k] = tmp
					default:
						tmpMap[k] = v

					}
				}
			}
			tmpRecords[i] = tmpMap
		}
	}

	// innerObj["records"] = tmpRecords

	// outputObj := make(map[string]interface{})
	// outputObj[objectName] = innerObj

	return tmpRecords

}

type Response struct {
	Done            bool                     `json:"done, omitempty"`
	Records         []map[string]interface{} `json:"records,omitempty"`
	TotalSize       int64                    `json:"totalSize,omitempty"`
	NextRecordsaUrl string                   `json:"nextRecordsUrl"`
}
