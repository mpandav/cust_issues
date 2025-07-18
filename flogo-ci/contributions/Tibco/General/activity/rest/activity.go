package rest

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	mxj "github.com/clbanning/mxj/v2"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/property"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/support/trace"
	"github.com/project-flogo/flow/instance"
	"github.com/tibco/flogo-general/src/app/General/connector/authorization"
	"github.com/tibco/wi-contrib/environment"
)

// RESTActivity is an Activity that is used to invoke a REST Operation
type RESTActivity struct {
	cachedClients sync.Map
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&RESTActivity{}, New)
}

// New creates new instance of RESTActivity
func New(ctx activity.InitContext) (activity.Activity, error) {
	return &RESTActivity{cachedClients: sync.Map{}}, nil
}

// Metadata returns the activity's metadata
func (a *RESTActivity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements activity.Activity.Eval - Invokes a REST Operation
func (a *RESTActivity) Eval(context activity.Context) (done bool, err error) {
	input := &Input{}

	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	method := input.Method

	parameters, err := GetParameter(context, input, context.Logger())
	if err != nil {
		context.Logger().Error(err)
		return false, err
	}

	var uri string
	if input.AsrEnabled {
		// Replace gsbc and app id if it comes from TCI
		if len(input.URI) > 0 && strings.Contains(input.URI, "gsbc") && strings.Contains(input.URI, "tci") {
			//TODO, need changes after enable cross-org supports
			subId := environment.GetTCISubscriptionId()
			uri = input.URI
			if len(subId) > 0 {
				if strings.HasPrefix(uri, "/") {
					uri = uri[1:]
				}
				urlparts := strings.Split(uri, "/")
				urlparts[1] = subId
				uri = strings.Join(urlparts, "/")
			}

			if serviceAppId, ok := GetServiceAppId(parameters.PathParams); ok {
				if len(serviceAppId) > 0 {
					if strings.HasPrefix(uri, "/") {
						uri = uri[1:]
					}
					urlparts := strings.Split(uri, "/")
					urlparts[3], _ = coerce.ToString(serviceAppId)
					uri = strings.Join(urlparts, "/")
				}
			}
			uri = environment.GetIntercomURL() + "/" + uri + input.ResourcePath
		} else {
			uri = environment.GetIntercomURL() + input.URI + input.ResourcePath
		}
	} else if input.ResourcePath != "" {
		uri = input.URI + input.ResourcePath
	} else {
		uri = input.URI
	}

	uri = buildURI(uri, parameters, context.Logger())

	key := context.ActivityHost().Name() + "-" + context.Name() + "-" + uri + "-" + method

	var client *http.Client
	if input.AsrEnabled {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second, // dial timeout
				KeepAlive: 0,               // disable keep-alives
				DualStack: true,
			}).DialContext,
			DisableKeepAlives: true,
		}
		client = &http.Client{Transport: tr}
	} else {
		cachedClient, ok := a.cachedClients.Load(key)
		if !ok {
			u, err := url.Parse(uri)
			if err != nil {
				return false, err
			}
			cachedClient, err = handleTransport(u.Scheme, input, context.Logger())
			if err != nil {
				return false, err
			}
			a.cachedClients.Store(key, cachedClient)
		}
		client = cachedClient.(*http.Client)

	}

	startTime := time.Now()
	var resp *http.Response
	if !input.AsrEnabled && input.Authorization {
		//Oauth2 code
		authConnection := input.AuthorizationConn.GetConnection().(*authorization.AuthorizationConnection)
		context.Logger().Infof("Request=====: ID: %s, Method: %s, URL: %s", context.Name()+"-"+context.ActivityHost().ID(), method, uri)
		if parameters.RequestType == "multipart/form-data" {
			resp, err = authConnection.SendRequest(client, method, uri, getHeaders(context, parameters, method), input.FormData, parameters.RequestType, input.Host, input.AsrEnabled, context.GetTracingContext(), context.Logger())
		} else {
			resp, err = authConnection.SendRequest(client, method, uri, getHeaders(context, parameters, method), input.Body, parameters.RequestType, input.Host, input.AsrEnabled, context.GetTracingContext(), context.Logger())
		}

	} else {
		//Just making sure are using same way as last release
		var reqBody io.Reader
		var multipartHeader string
		if parameters.RequestType == "multipart/form-data" && method != "GET" && method != "DELETE" {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if err != nil {
				return false, fmt.Errorf("error occured while marshaling data: %v", err.Error())
				//fmt.Errorf("HERE..", err.Error())
			}

			schemaTable := ToSchemaTable(input.MultipartForm)

			if schemaTable.Value == "" && input.FormData != nil {

				for k, v := range input.FormData {
					switch v.(type) {

					case map[string]interface{}:
						obj, _ := coerce.ToObject(v)
						b := new(bytes.Buffer)
						for key, value := range obj {
							fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
						}
						_ = writer.WriteField(k, b.String())

					default:
						fileString, _ := coerce.ToString(v)
						_ = writer.WriteField(k, fileString)
					}

				}

			} else {
				schemaTableValueMap, err := schemaTable.ToMap(input.FormData)
				if err != nil {
					return false, fmt.Errorf("Unable to read form data due to error: %v", err)
				}
				for k, v := range schemaTableValueMap {
					switch v.Type {
					case "filecontent":
						f := SchemaTableFile{
							Key: k, FileName: "tempfile", Value: v.Value, Required: v.Required,
						}
						err = readMultipartFormData(f, writer)
						if err != nil {
							return false, err
						}
					case "file":
						fileObj, _ := coerce.ToObject(v.Value)
						if v.Required && (fileObj == nil || len(fileObj) == 0) {
							return false, fmt.Errorf("Required form parameter [%s] is not configured", k)
						}
						fileName, _ := coerce.ToString(fileObj["filename"])
						contentType, _ := coerce.ToString(fileObj["content-type"])

						if fileName == "" {
							fileName = "tempfile"
						}
						f := SchemaTableFile{
							Key: k, FileName: fileName, Value: fileObj["content"], Required: v.Required, ContentType: contentType,
						}
						err = readMultipartFormData(f, writer)
						if err != nil {
							return false, err
						}

					case "files":
						fileArray, _ := coerce.ToArray(v.Value)
						if v.Required && len(fileArray) == 0 {
							return false, fmt.Errorf("Required form parameter [%s] is not configured", k)

						}
						for _, fileContent := range fileArray {
							fileObj, _ := coerce.ToObject(fileContent)
							fileName, _ := coerce.ToString(fileObj["filename"])
							if fileName == "" {
								fileName = "tempfile"
							}
							f := SchemaTableFile{
								Key: k, FileName: fileName, Value: fileObj["content"], Required: v.Required,
							}
							err = readMultipartFormData(f, writer)
							if err != nil {
								return false, err
							}
						}
					case "object":
						fileObj, _ := coerce.ToObject(v.Value)
						if v.Required && (fileObj == nil || len(fileObj) == 0) {
							return false, fmt.Errorf("Required form parameter [%s] is not configured", k)
						}
						b := new(bytes.Buffer)
						for key, value := range fileObj {
							fmt.Fprintf(b, "%s=\"%s\"\n", key, value)
						}
						_ = writer.WriteField(k, b.String())
					case "string":
						fileString, _ := coerce.ToString(v.Value)
						if v.Required && len(fileString) == 0 {
							return false, fmt.Errorf("Required form parameter [%s] is not configured", k)
						}
						_ = writer.WriteField(k, fileString)
					}
				}

			}

			err = writer.Close()
			if err != nil {
				return true, err
			}

			reqBody = body

			multipartHeader = writer.FormDataContentType()
		} else {
			reqBody, err = authorization.GetRequestBody(method, input.Body, parameters.RequestType, input.AsrEnabled, context.Logger())
			if err != nil {
				return false, err
			}
		}

		req, _ := http.NewRequest(method, uri, reqBody)
		if parameters.RequestType == "multipart/form-data" {
			req.Header.Set("Content-Type", multipartHeader)
		}

		if input.Host != "" {
			context.Logger().Infof("Overriding Host With: %s", input.Host)
			req.URL.Host = input.Host
			req.Host = input.Host
		}

		context.Logger().Infof("Request: ID: %s, Method: %s, URL: %s", context.Name()+"-"+context.ActivityHost().ID(), method, req.URL.String())
		handleHeaders(context, req, parameters, method)

		if input.AsrEnabled {
			//For cic2 we need set
			// 1.X-Atmosphere-Tenant-Id: tciapps
			// 2.X-Atmosphere-Subscription-Id
			req.Header.Set("X-Atmosphere-Tenant-Id", getTenantId(req.URL.String(), environment.GetTCISubscriptionId()))
			req.Header.Set("X-ATMOSPHERE-for-USER", environment.GetTCISubscriptionUName())
			req.Header.Set("X-Atmosphere-Subscription-Id", environment.GetTCISubscriptionId())
		}

		if context.GetTracingContext() != nil {
			_ = trace.GetTracer().Inject(context.GetTracingContext(), trace.HTTPHeaders, req)
		}

		logHeaders(context, req.Header)
		resp, err = client.Do(req)
	}

	if err != nil {
		if err2, ok := err.(*url.Error); ok {
			// Return retriable error
			return false, activity.NewRetriableError(err2.Error(), "", nil)
		}
		context.Logger().Errorf("Failed to send request due to error: %s", err.Error())
		return false, err
	}

	defer func() {
		if resp.Body != nil {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	responseTime := int64(time.Since(startTime) / time.Millisecond)
	context.Logger().Infof("Response: ID: %s, Status: %s, ResponseTime: %dms", context.Name()+"-"+context.ActivityHost().ID(), resp.Status, responseTime)

	if context.GetTracingContext() != nil {
		context.GetTracingContext().SetTag("http.status_code", resp.StatusCode)
	}

	context.Logger().Debugf("Rest invoke response status code [%d]", resp.StatusCode)

	output := &Output{}
	output.StatusCode = resp.StatusCode

	output.ResponseTime = responseTime

	context.Logger().Debug("Rest invoke response Status:", resp.Status)

	throwErr := context.(*instance.TaskInst).Task().ActivityConfig().GetOutput("throwError")
	var throwError = false
	if throwErr != nil {
		throwError = throwErr.(bool)
	}

	if throwError && resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("URL: %s, HTTP Response Status: %s", uri, resp.Status)
		if len(body) == 0 {
			context.Logger().Debugf("HTTP error %d: response body is empty", resp.StatusCode)
			return false, activity.NewError(msg, strconv.Itoa(resp.StatusCode), nil)
		} else {
			var jsonObj interface{}
			contentType := resp.Header.Get("Content-Type")
			if strings.Contains(contentType, "application/json") {
				if err := json.Unmarshal(body, &jsonObj); err == nil {
					pretty, _ := json.MarshalIndent(jsonObj, "", "  ")
					context.Logger().Debugf("HTTP error %d: %s\n", resp.StatusCode, string(pretty))
					return false, activity.NewError(msg, strconv.Itoa(resp.StatusCode), string(pretty))
				}
			} else {
				context.Logger().Debugf("HTTP error %d: %s\n", resp.StatusCode, string(body))
				return false, activity.NewError(msg, strconv.Itoa(resp.StatusCode), string(body))
			}
		}
	}

	var result interface{}

	//Response Headers
	var responseHeaders = map[string]interface{}{}
	context.Logger().Debug("Reading Response Header parameters")

	headersMap, err := LoadJsonSchemaFromOutput(context, "headers")
	if err != nil {
		return false, fmt.Errorf("failed to parse output headers due to error: %s", err.Error())
	}
	if headersMap != nil {
		headersConfig, err := ParseParams(headersMap)
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
	for key := range resp.Header {
		if responseHeaders[key] == nil {
			val := resp.Header[http.CanonicalHeaderKey(key)]
			responseHeaders[key] = strings.Join(val, ",")
		}
	}
	output.Headers = responseHeaders

	var responseCodes = map[string]interface{}{}

	if resp.Body != nil {
		var respBody io.ReadCloser
		// Check if content-encoding header is present
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			respBody, err = gzip.NewReader(resp.Body)
			if err != nil {
				return false, fmt.Errorf("Unable to decode response body due to error: [%s]", err.Error())
			}
			defer respBody.Close()
		case "deflate":
			respBody = flate.NewReader(resp.Body)
			defer respBody.Close()
		case "":
			// Content-Encoding header is empty, means either body is not encoded or is decoded by go
			respBody = resp.Body
			defer respBody.Close()
		default:
			return false, fmt.Errorf("Unsupported content encoding [%s]. Only gzip and deflate are supported", resp.Header.Get("Content-Encoding"))
		}
		contentType := resp.Header.Get("Content-Type")
		context.Logger().Debugf("Rest invoke response content-type: %s", contentType)

		confRC := context.(*instance.TaskInst).Task().ActivityConfig().GetOutput("configureResponseCodes")
		if confRC != nil {
			//this is for the apps having version >= 2.7
			confRC := confRC.(bool)
			if !confRC {
				//it is false that is user dont want to configure the response codes
				if parameters.ResponseType == "application/json" {
					if !strings.Contains(contentType, "application/json") {
						context.Logger().Warnf("expected content type [%s] but recieved [%s], attempting to convert response body to json", parameters.ResponseType, contentType)
					}
					if contentType == "application/xml" {
						body, err := ioutil.ReadAll(respBody)
						if err != nil {
							return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
						}
						jsonBytes, err := xmltoJSonConversion(body, context.Logger())
						if err != nil {
							return false, fmt.Errorf("unable to convert response to json due to error: [%s]", err.Error())

						}
						context.Logger().Debugf("Rest invoke  API Json converted response::", string(jsonBytes))
						json.Unmarshal(jsonBytes, &result)
						//output.ResponseBody = result
					} else {
						d := json.NewDecoder(respBody)
						d.UseNumber()
						err = d.Decode(&result)
						if err != nil {
							//Looks like invalid response
							return false, fmt.Errorf("unable to deserialize response to json due to error: [%s]", err.Error())
						}
					}
					output.ResponseBody = result
				} else if parameters.ResponseType == "application/xml" && parameters.ResponseOutput == "JSON Object" {
					if !strings.Contains(contentType, "application/xml") {
						context.Logger().Warnf("expect content type [%s] but recieved [%s], try to convert response body to json", parameters.ResponseType, contentType)
					}
					if contentType == "application/json" {
						d := json.NewDecoder(respBody)
						d.UseNumber()
						err = d.Decode(&result)
						if err != nil {
							//Looks like invalid response
							return false, fmt.Errorf("unable to deserialize response to json due to error: [%s]", err.Error())
						}
						//output.ResponseBody = result
					} else {
						body, err := ioutil.ReadAll(respBody)
						if err != nil {
							return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
						}
						context.Logger().Debugf("Rest invoke  API XML response::", string(body))

						jsonBytes, err := xmltoJSonConversion(body, context.Logger())
						if err != nil {
							return false, fmt.Errorf("unable to convert response to json due to error: [%s]", err.Error())
						}
						context.Logger().Debugf("Rest invoke  API Json converted response::", string(jsonBytes))
						json.Unmarshal(jsonBytes, &result)
					}

					output.ResponseBody = result
				} else if (parameters.ResponseType == "application/xml" && parameters.ResponseOutput == "XML String") || parameters.ResponseType == "text/plain" {

					respBody, err := ioutil.ReadAll(respBody)
					if err != nil {
						return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
					}

					if parameters.ResponseType == "application/xml" {
						if !strings.Contains(contentType, "application/xml") {
							context.Logger().Warnf("expect content type [%s] but got [%s], set all as text to output", parameters.ResponseType, contentType)
						}

					} else {
						if !strings.Contains(contentType, "text/plain") {
							context.Logger().Warnf("expect content type [%s] but got [%s], set all as text to output", parameters.ResponseType, contentType)
						}
					}
					//Just put text to body
					respData := make(map[string]string)
					respData["data"] = string(respBody)
					output.ResponseBody = respData
				} else {
					respBody, err := ioutil.ReadAll(respBody)
					if err != nil {
						return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
					}
					context.Logger().Debugf("Content type set to other, set all body to base64 encoded string")
					respData := make(map[string]string)
					respData["data"] = base64.StdEncoding.EncodeToString(respBody)
					output.ResponseBody = respData
				}
			} else {
				// if the user want response code to be configured
				codeMap, err := LoadJsonSchemaFromOutput(context, "responseCodes")
				if err != nil {
					return false, fmt.Errorf("error loading responseCodes output schema: %s", err.Error())
				}
				props := codeMap["properties"].(map[string]interface{})

				if strings.Contains(contentType, "application/json") {

					d := json.NewDecoder(respBody)
					d.UseNumber()
					err = d.Decode(&result)
				} else if strings.Contains(contentType, "application/xml") {

					body, err := ioutil.ReadAll(respBody)
					if err != nil {
						return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
					}
					context.Logger().Debugf("Rest invoke  API XML response::", string(body))
					jsonBytes, err := xmltoJSonConversion(body, context.Logger())
					if err != nil {
						return false, fmt.Errorf("unable to convert response to json due to error: [%s]", err.Error())

					}
					context.Logger().Debugf("Rest invoke  API Json converted response::", string(jsonBytes))
					json.Unmarshal(jsonBytes, &result)
				} else if strings.Contains(contentType, "text/plain") {
					respBody, err := ioutil.ReadAll(respBody)
					if err != nil {
						return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
					}
					result = string(respBody)
				} else {
					context.Logger().Debugf("Content type set to other, set all body to base64 encoded string", contentType)
					respBody, err := ioutil.ReadAll(respBody)
					if err != nil {
						return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
					}
					result = base64.StdEncoding.EncodeToString(respBody)

				}

				sc := output.StatusCode //always numeric
				var genericCodes = make(map[int]string)
				for k := range props {
					if strings.Contains(k, "_") {
						//Response Code have Headers
						headerCode, _ := strconv.Atoi(strings.Split(k, "_")[0])
						if resp.StatusCode == headerCode {
							headerCodeCount := props[k].(map[string]interface{})["properties"].(map[string]interface{})
							tempHeaderMap := make(map[string]interface{})
							for h := range headerCodeCount {
								value := resp.Header[http.CanonicalHeaderKey(h)]
								tempHeaderMap[h] = value[0]
							}
							responseCodes[strings.Split(k, "_")[0]+"_headers"] = tempHeaderMap
						}

					} else {
						convK, err := strconv.Atoi(k)
						if err == nil {
							if convK == sc {
								responseCodes[string(k)] = result
							} else {
								responseCodes[string(k)] = ""
							}
						} else {
							responseCodes[k] = ""
							genericCodes[len(genericCodes)] = k
						}
					}
				}

				if responseCodes[strconv.Itoa(sc)] == nil {
					generic := strconv.Itoa(sc)[:1] + "xx"
					for g := range genericCodes {
						if strings.EqualFold(genericCodes[g], generic) {
							responseCodes[genericCodes[g]] = result
						} else {
							responseCodes[genericCodes[g]] = ""
						}
					}
				}
				output.ResponseCodes = responseCodes
			}
		} else {
			//this for the apps below 2.7
			if parameters.ResponseType == "application/json" {
				if !strings.Contains(contentType, "application/json") {
					context.Logger().Warnf("expect content type [%s] but get [%s], try to convert response body to json", parameters.ResponseType, contentType)
				}

				d := json.NewDecoder(respBody)
				d.UseNumber()
				err = d.Decode(&result)

				if err != nil {
					//Looks like invalid response
					return false, fmt.Errorf("unable to deserialize response to json due to error: [%s]", err.Error())
				}
				output.ResponseBody = result
			} else if parameters.ResponseType == "text/plain" {
				if !strings.Contains(contentType, "text/plain") {
					context.Logger().Warnf("expect content type [%s] but got [%s], set all as text to output", parameters.ResponseType, contentType)
				}
				respBody, err := ioutil.ReadAll(respBody)
				if err != nil {
					return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
				}
				//Just put text to body
				respData := make(map[string]string)
				respData["data"] = string(respBody)
				output.ResponseBody = respData
			} else {
				respBody, err := ioutil.ReadAll(respBody)
				if err != nil {
					return false, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
				}
				context.Logger().Debugf("Content type set to other, set all body to base64 encoded string")
				respData := make(map[string]string)
				respData["data"] = base64.StdEncoding.EncodeToString(respBody)
				output.ResponseBody = respData
			}
		}
	}
	err = context.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("error setting output for RESTActivity [%s]: %s", context.Name(), err.Error())
	}

	/*if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		//1xx,4xx, 5xx errors
		return false, activity.NewError(resp.Status, strconv.Itoa(resp.StatusCode), nil)
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		//3xx errors
		return false, activity.NewError("Redirection not supported", strconv.Itoa(resp.StatusCode), nil)
	} else {
		//2xx
	}*/

	return true, nil

}

func getTenantId(url, subId string) string {
	urlPaths := strings.SplitAfter(url, subId)
	if len(urlPaths) > 1 {
		tenantLocator := strings.Split(urlPaths[1], "/")
		if len(tenantLocator) > 2 {
			switch tenantLocator[1] {
			case "tci":
				return "tciapps"
			case "tciapps":
				return "tciapps"
			case "tceuserapp":
				return "tceapps"
			case "tcam":
				return "TCAM"
			default:
				return "tciapps"
			}
		}
	}
	return "tciapps"
}

func xmltoJSonConversion(body []byte, log log.Logger) ([]byte, error) {
	s, err := mxj.NewMapXml(body)
	var emptyslice []byte
	if err != nil {
		log.Error(err)
		return emptyslice, fmt.Errorf("unable to read response body due to error: [%s]", err.Error())
	}
	dataMap := s.Old()
	jsonBytes, err := json.Marshal(dataMap)
	if err != nil {
		log.Error(err)
		return emptyslice, fmt.Errorf("unable to marshal  response body due to error: [%s]", err.Error())
	}
	return jsonBytes, nil
}

func buildURI(uri string, param *Parameters, log log.Logger) string {
	if param != nil {
		if param.PathParams != nil && len(param.PathParams) > 0 {
			uri = BuildURI(uri, param.PathParams)
		}

		if param.QueryParams != nil && len(param.QueryParams) > 0 {
			qp := url.Values{}
			for _, value := range param.QueryParams {
				if strings.ToLower(value.Type) == "object" {
					paramMap, _ := coerce.ToObject(value.Value)
					for k, v := range paramMap {
						qp.Add(k, fmt.Sprintf("%v", v))
					}
				} else {
					qp.Add(value.Name, value.ToString(log))
				}
			}
			uri = uri + "?" + qp.Encode()
		}

	}
	return uri
}

func handleHeaders(ctx activity.Context, req *http.Request, param *Parameters, method string) {

	//if !confiuredResponseCode(ctx) {
	//	if param.ResponseType != "other" {
	//		req.Header.Set("Accept", param.ResponseType)
	//	}
	//}

	if param != nil && param.Headers != nil && len(param.Headers) > 0 {
		for _, v := range param.Headers {
			//Any input should oeverride exist header
			// To avoid canonicalization of header name, adding headers directly to the request header map instead of using Add/Set.
			req.Header[v.Name] = []string{v.ToString(ctx.Logger())}
		}
	}

	if method == authorization.MethodPOST || method == authorization.MethodPUT || method == authorization.MethodPATCH {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", param.RequestType)
		}
	}

}

func getHeaders(ctx activity.Context, param *Parameters, method string) http.Header {
	header := make(http.Header)
	//if !confiuredResponseCode(ctx) {
	//	if param.ResponseType != "other" {
	//		header.Set("Accept", param.ResponseType)
	//	}
	//}

	if param != nil && param.Headers != nil && len(param.Headers) > 0 {
		for _, v := range param.Headers {
			//Any input should oeverride exist header
			// To avoid canonicalization of header name, adding headers directly to the request header map instead of using Add/Set.
			header[v.Name] = []string{v.ToString(ctx.Logger())}
		}
	}

	if method == authorization.MethodPOST || method == authorization.MethodPUT || method == authorization.MethodPATCH {
		if header.Get("Content-Type") == "" {
			header.Set("Content-Type", param.RequestType)
		}
	}
	return header
}

func handleTransport(schema string, input *Input, log log.Logger) (*http.Client, error) {
	//Security Configuration
	var useSslcert = input.CertificateProvided
	var disableSslValidation = input.DisableSslValidation

	// if schema == "https" || schema == "HTTPS" {
	// 	disableSslValidation = !useSslcert
	// }
	mTLSConfig := &tls.Config{}
	mTLSConfig.InsecureSkipVerify = disableSslValidation

	transport := &http.Transport{TLSClientConfig: mTLSConfig}
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100
	transport.IdleConnTimeout = 5 * time.Second

	transport.DisableKeepAlives = input.DisableKeepAlives
	transport.Proxy = http.ProxyFromEnvironment

	// Set the proxy server to use, if supplied
	proxyValue := input.Proxy
	if len(proxyValue) > 0 {

		if !strings.Contains(proxyValue, "http://") && !strings.Contains(proxyValue, "https://") {
			log.Debug("Setting the http protocol....")
			proxyValue = "http://" + proxyValue
		}

		proxyURL, urlErr := url.Parse(proxyValue)
		if urlErr != nil {
			log.Debug("Error parsing proxy url:", urlErr)
			return nil, urlErr
		}

		log.Info("Setting proxy server:", proxyValue)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	var client *http.Client
	client = &http.Client{Transport: transport}
	if !input.FollowRedirects {
		client = &http.Client{Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}}
	}

	if input.Timeout > 0 {
		//Set the timeout only if it is positive
		log.Infof("Client timeout is set to '%d' milliseconds", input.Timeout)
		client.Timeout = time.Millisecond * time.Duration(input.Timeout)
	}

	if useSslcert {
		serverCert := input.ServerCertificate
		cCert := input.ClientCertificate
		cKey := input.ClientKey

		if serverCert == "" {
			return nil, activity.NewError("Server/CA certificate not configured for TLS", "", nil)
		}

		if input.MutualAuth && (cCert == "" || cKey == "") {
			return nil, activity.NewError("Client certificates not configured for mutual TLS", "", nil)
		}

		sCert, err := decodeCerts(serverCert)
		if err != nil {
			return nil, err
		}

		if sCert != nil {
			caCertPool := x509.NewCertPool()
			pemBlock, _ := pem.Decode(sCert)
			if pemBlock == nil {
				return nil, activity.NewError("Unsupported certificate found. It must be a valid PEM certificate.", "", nil)
			}
			serverCert, err1 := x509.ParseCertificate(pemBlock.Bytes)
			if err1 != nil {
				return nil, err1
			}
			caCertPool.AddCert(serverCert)
			mTLSConfig.RootCAs = caCertPool
		}

		if input.MutualAuth {
			log.Debug("mTLS enabled")
			//Mutual authentication enabled
			clientCert, err := decodeCerts(cCert)
			if err != nil {
				return nil, err
			}
			clientKey, err := decodeCerts(cKey)
			if err != nil {
				return nil, err
			}
			cert, err := tls.X509KeyPair(clientCert, clientKey)
			if err != nil {
				return nil, err
			}
			mTLSConfig.Certificates = []tls.Certificate{cert}
			mTLSConfig.BuildNameToCertificate()
			mTLSConfig.ClientAuth = 4
		}
	}
	return client, nil
}

func decodeCerts(certVal string) ([]byte, error) {
	if certVal == "" {
		return nil, fmt.Errorf("Certificate not configured")
	}

	//if certificate comes from fileselctor it will be base64 encoded
	if strings.HasPrefix(certVal, "{") {
		certObj, err := coerce.ToObject(certVal)
		if err == nil {
			certRealValue, ok := certObj["content"].(string)
			if !ok || certRealValue == "" {
				return nil, fmt.Errorf("Invalid certificate value")
			}

			index := strings.IndexAny(certRealValue, ",")
			if index > -1 {
				certRealValue = certRealValue[index+1:]
			}

			encodedDataOfCert, err := base64.StdEncoding.DecodeString(certRealValue)
			if err != nil {
				return nil, fmt.Errorf("Invalid base64 encoded certificate value")
			}
			return []byte(encodedDataOfCert), nil
		}
		return nil, err
	}

	//if certificate is read from k8s secret then it's in original format
	if strings.HasPrefix(certVal, "-----") {
		return []byte(certVal), nil
	}

	//if the certificate is defined as application property that points to a file
	if strings.HasPrefix(certVal, "file://") {
		// app property pointing to a file
		fileName := certVal[7:]
		return os.ReadFile(fileName)
	}

	//if certificate is in base64 endoded format
	encodedDataOfCert, err := base64.StdEncoding.DecodeString(certVal)
	if err != nil {
		return nil, fmt.Errorf("Invalid base64 encoded certificate. Check override value configured to the application property.")
	}
	return []byte(encodedDataOfCert), nil
}

// BuildURI util
func BuildURI(uri string, values []*TypedValue) string {
	for _, pp := range values {
		data, _ := coerce.ToString(pp.Value)
		uri = strings.Replace(uri, "{"+pp.Name+"}", data, -1)
	}
	return uri
}

func GetServiceAppId(values []*TypedValue) (string, bool) {
	for _, pp := range values {
		data, _ := coerce.ToString(pp.Value)
		if pp.Name == "serviceAppId" {
			return data, true
		}
	}
	return "", false
}

func confiuredResponseCode(context activity.Context) bool {
	t, ok := context.(*instance.TaskInst)
	if ok && t != nil {
		if t.Task() != nil && t.Task().ActivityConfig() != nil {
			yes, _ := coerce.ToBool(t.Task().ActivityConfig().GetOutput("configureResponseCodes"))
			return yes
		}
	}
	return false
}

func getAppProperty(name string) interface{} {
	manager := property.DefaultManager()
	property, ok := manager.GetProperty(name)
	if ok {
		return property
	}
	return nil
}

func logHeaders(context activity.Context, h http.Header) {
	hCopy := make(map[string][]string, len(h))
	for k, v := range h {
		hCopy[k] = v
	}
	if _, ok := hCopy[http.CanonicalHeaderKey("Authorization")]; ok {
		hCopy["Authorization"] = []string{"********"}
	}
	context.Logger().Debugf("Request Headers: %+v", hCopy)
}

func readMultipartFormData(f SchemaTableFile, writer *multipart.Writer) error {
	switch value := f.Value.(type) {
	case [][]uint8:
		if f.Required && len(value) == 0 {
			return fmt.Errorf("Required form parameter [%s] is not configured", f.Key)
		}
		for _, fileData := range value {
			part, err := createPart(writer, f.Key, f.ContentType, f.FileName)
			if err != nil {
				return err
			}
			reader := bytes.NewReader([]byte(fileData))
			_, err = io.Copy(part, reader)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		if f.Required && len(value) == 0 {
			return fmt.Errorf("Required form parameter [%s] is not configured", f.Key)
		}
		for _, fileData := range value {
			part, err := createPart(writer, f.Key, f.ContentType, f.FileName)
			if err != nil {
				return err
			}
			reader := bytes.NewReader([]byte(fileData.(string)))
			_, err = io.Copy(part, reader)
			if err != nil {
				return err
			}
		}
	default:
		fileData, _ := coerce.ToString(value)
		if f.Required && len(fileData) == 0 {
			return fmt.Errorf("Required form parameter [%s] is not configured", f.Key)
		}
		part, err := createPart(writer, f.Key, f.ContentType, f.FileName)
		if err != nil {
			return err
		}
		reader := bytes.NewReader([]byte(fileData))
		_, err = io.Copy(part, reader)
		if err != nil {
			return err
		}
	}
	return nil
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func createPart(writer *multipart.Writer, fieldName, contentType, fileName string) (io.Writer, error) {
	if contentType != "" {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(fieldName), escapeQuotes(fileName)))
		h.Set("Content-Type", contentType)
		return writer.CreatePart(h)
	}
	return writer.CreateFormFile(fieldName, fileName)
}
