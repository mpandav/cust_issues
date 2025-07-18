package modify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/flogo-odata/src/app/OData/connector/odata"
)

// ODataModifyActivity is an Activity that is used to create, update and delete a OData resource
type ODataModifyActivity struct {
	client *odata.AuthorizationConnection
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&ODataModifyActivity{}, New)
}

// New creates new instance of ODataModifyActivity
func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		ctx.Logger().Errorf("Failed to read activity settings due to error - %s", err.Error())
		return nil, err
	}
	client := s.ODataConnection.GetConnection().(*odata.AuthorizationConnection)
	return &ODataModifyActivity{client: client}, nil
}

// Metadata returns the activity's metadata
func (a *ODataModifyActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval - Invokes a OData service endpoint
func (a *ODataModifyActivity) Eval(context activity.Context) (done bool, err error) {
	input := &Input{}
	context.Logger().Info("Executing OData Modify Activity")

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	oDataConnection := a.client

	//fmt.Printf("oDataConnection: %+v\n ", oDataConnection)

	if oDataConnection == nil {
		return false, fmt.Errorf("OData Connection is not configured")
	}

	operation := input.Operation

	context.Logger().Info("Operation: ", operation)

	method := http.MethodPost

	if operation == "Update" {
		method = http.MethodPatch
	} else if operation == "Delete" {
		method = http.MethodDelete
	}

	reqType := input.RequestType

	uri := oDataConnection.RootURL + input.ODataURI

	context.Logger().Debug("oDataURI : ", uri)

	reqBody := input.RequestBody

	queryOpts, err := GetQueryOpts(context, input, context.Logger())
	if err != nil {
		context.Logger().Error(err)
		return false, err
	}

	uri = buildURI(uri, queryOpts, context.Logger())

	context.Logger().Info("Calling odata service endpoint: ", uri)

	var resp *http.Response

	client := &http.Client{}

	resp, err = oDataConnection.SendRequest(client, method, uri, getHeaders(context, queryOpts), reqBody, reqType, context.Logger())

	if err != nil {
		if err2, ok := err.(*url.Error); ok {
			// Return retriable error
			return false, activity.NewRetriableError(err2.Error(), "", nil)
		}
		context.Logger().Errorf("Failed to send odata request due to error: %s", err.Error())
		return false, err
	}

	defer func() {
		if resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	context.Logger().Info("OData modify activity response status: ", resp.Status)

	output := &Output{}

	//Response Headers
	var responseHeaders = map[string]interface{}{}
	context.Logger().Debug("Reading Response Header parameters")

	headersMap, err := LoadJsonSchemaFromOutput(context, "responseHeaders")
	if err != nil {
		return false, fmt.Errorf("failed to parse output headers due to error: %s", err.Error())
	}
	if headersMap != nil {
		headersConfig, err := ParseQueryOpts(headersMap)
		if err != nil {
			context.Logger().Error(err)
			return false, err
		}
		for _, hParam := range headersConfig {
			name := hParam.Name
			value := resp.Header[http.CanonicalHeaderKey(name)]
			if hParam.Required == "true" && value == nil {
				return false, fmt.Errorf("failed to process response. Required header [%s] value is empty", hParam.Name)
			}

			if value != nil {
				if hParam.Repeating == "true" {
					responseHeaders[name] = value
				} else {
					responseHeaders[name] = strings.Join(value, ",")
				}
			}
		}

	}
	output.ResponseHeaders = responseHeaders

	var result interface{}

	// Check if content-encoding header is present
	contentType := resp.Header.Get("Content-Type")
	context.Logger().Debugf("OData modify activity response content-type: %s", contentType)

	body, err := io.ReadAll(resp.Body)

	// context.Logger().Debugf("body: ", string(body))

	if err != nil {
		return false, fmt.Errorf("unable to deserialize response to json due to error: [%s]", err.Error())
	}

	// Check if response is JSON
	if strings.Contains(contentType, "application/json") {

		err = json.Unmarshal(body, &result)
		if err != nil {
			context.Logger().Errorf("Failed to unmarshal response body due to error: %s", err.Error())
		}
		output.ResponseBody = result

	} else {

		context.Logger().Debugf("content-type set to other, set all body to string")
		respData := make(map[string]string)
		respData["data"] = string(body)
		output.ResponseBody = respData

	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("error setting output for OData Modify Activity [%s]: %s", context.Name(), err.Error())
	}

	return true, nil

}

func buildURI(uri string, queryOpts *QueryOpts, log log.Logger) string {
	if queryOpts != nil {
		if queryOpts.Parameters != nil && len(queryOpts.Parameters) > 0 {
			uri = BuildURI(uri, queryOpts.Parameters)
		}

		if queryOpts.QueryOptions != nil && len(queryOpts.QueryOptions) > 0 {
			qp := url.Values{}
			for _, value := range queryOpts.QueryOptions {
				qp.Add(value.Name, value.ToString(log))
			}
			uri = uri + "?" + qp.Encode()
		}

	}
	return uri
}

// BuildURI util
func BuildURI(uri string, values []*TypedValue) string {
	for _, pp := range values {
		data, _ := coerce.ToString(pp.Value)
		switch pp.Type {
		case "string":
			uri = strings.Replace(uri, pp.Name, "'"+data+"'", -1)
		case "number":
			uri = strings.Replace(uri, pp.Name, data, -1)

		}
	}
	return uri
}

func getHeaders(ctx activity.Context, queryOpts *QueryOpts) http.Header {
	header := make(http.Header)

	if queryOpts != nil && queryOpts.Headers != nil && len(queryOpts.Headers) > 0 {
		for _, v := range queryOpts.Headers {
			//Any input should oeverride exist header
			// To avoid canonicalization of header name, adding headers directly to the request header map instead of using Add/Set.
			header[v.Name] = []string{v.ToString(ctx.Logger())}
		}
	}

	return header
}

func handleHeaders(ctx activity.Context, req *http.Request, queryOpts *QueryOpts) {

	if queryOpts != nil && queryOpts.Headers != nil && len(queryOpts.Headers) > 0 {
		for _, v := range queryOpts.Headers {
			//Any input should oeverride exist header
			// To avoid canonicalization of header name, adding headers directly to the request header map instead of using Add/Set.
			req.Header[v.Name] = []string{v.ToString(ctx.Logger())}
		}
	}

}
