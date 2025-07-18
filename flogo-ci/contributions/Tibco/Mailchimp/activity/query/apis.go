package query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/project-flogo/core/support/log"
	mailchimp "github.com/tibco/wi-mailchimp/src/app/Mailchimp/activity"
	mailchimpConn "github.com/tibco/wi-mailchimp/src/app/Mailchimp/connector/mailchimp"
)

const (
	campaignsPath = "/campaigns"
)

type ApiQuery struct {
	Log           log.Logger
	ActivityInput map[string]interface{}
	ApiToken      *mailchimpConn.Token
}

func (a *ApiQuery) DoQuery(url string) (interface{}, error) {

	queryParams := make(map[string][]string)
	if a.ActivityInput != nil {
		inByte, err := json.Marshal(a.ActivityInput)
		a.Log.Debugf("===>Activity input: %s", string(inByte))

		if err != nil {
			return nil, fmt.Errorf("fail to marshal input: %s", err.Error())
		}

		if string(inByte) == "\"{}\"" || string(inByte) == "{}" {
			a.Log.Debug("===>Empty value input for the call")

		} else {
			// not empty value
			paramMap := make(map[string]interface{})
			err = json.Unmarshal(inByte, &paramMap)
			if err != nil {
				return nil, fmt.Errorf("fail to unmarshal input %s", err.Error())
			}

			queryParams = ConvertToQueryParams(paramMap, a.Log)
		}

	} else {
		a.Log.Debug("===>No input for the call")
	}

	apiResponse, err := mailchimp.GetCall(url, a.ApiToken, queryParams, nil, a.Log)
	if err != nil {
		return nil, fmt.Errorf("fail to call Mailchimp API : %s", err.Error())
	}

	return apiResponse.Body, nil
}

func (a *ApiQuery) Campaigns() (interface{}, error) {
	url := a.ApiToken.APIEndpoint + mailchimp.API_VERSION_PATH + campaignsPath
	return a.DoQuery(url)
}

func ConvertToQueryParams(paramMap map[string]interface{}, logger log.Logger) map[string][]string {
	queryParams := make(map[string][]string)

	for k, v := range paramMap {
		if v == nil {
			continue // skip nil values
		}

		val := reflect.ValueOf(v)
		typ := val.Type()
		kind := typ.Kind()

		switch kind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			queryParams[k] = []string{strconv.FormatInt(val.Int(), 10)}

		case reflect.Float32, reflect.Float64:
			queryParams[k] = []string{strconv.FormatFloat(val.Float(), 'f', -1, 64)}

		case reflect.String:
			queryParams[k] = []string{val.String()}

		case reflect.Bool:
			queryParams[k] = []string{strconv.FormatBool(val.Bool())}

		case reflect.Slice:
			if val.Len() == 0 {
				// Handle empty slice explicitly
				queryParams[k] = []string{}
			} else {
				sliceStrings := []string{}
				for i := 0; i < val.Len(); i++ {
					elem := val.Index(i)
					switch elem.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						sliceStrings = append(sliceStrings, strconv.FormatInt(elem.Int(), 10))
					case reflect.String:
						sliceStrings = append(sliceStrings, elem.String())
					case reflect.Bool:
						sliceStrings = append(sliceStrings, strconv.FormatBool(elem.Bool()))
					case reflect.Float32, reflect.Float64:
						sliceStrings = append(sliceStrings, strconv.FormatFloat(elem.Float(), 'f', -1, 64))
					case reflect.Interface:
						// Handle interface{} elements by converting them to strings
						elemStr := fmt.Sprintf("%v", elem.Interface())
						sliceStrings = append(sliceStrings, elemStr)
					default:
						logger.Debugf("Skipping unsupported element type in slice for key %s: %s", k, elem.Kind())
					}
				}
				if len(sliceStrings) > 0 {
					queryParams[k] = sliceStrings
				}
			}

		default:
			logger.Debugf("Skipping unsupported type for key %s: %s", k, kind)
		}
	}

	return queryParams
}
