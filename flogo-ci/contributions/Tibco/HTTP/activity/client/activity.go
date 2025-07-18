package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/ssl"
	"github.com/project-flogo/core/support/trace"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

const (
	methodPOST  = "POST"
	methodPUT   = "PUT"
	methodPATCH = "PATCH"
)

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}
	act := &Activity{settings: s}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("redirects not supported")
		},
	}
	transport := &http.Transport{}
	client.Transport = transport
	transport.ResponseHeaderTimeout = time.Second * time.Duration(120)
	logger := ctx.Logger()
	if !(strings.HasPrefix(s.Host, "http://") || strings.HasPrefix(s.Host, "https://")) {
		logger.Infof("Protocol is not set in the host endpoint. Defaulting to http...")
		s.Host = "http://" + s.Host
	}

	if strings.HasPrefix(s.Host, "https://") {
		//TODO Enable certificate configuration
		cfg := &ssl.Config{}
		cfg.SkipVerify = true
		cfg.UseSystemCert = true
		tlsConfig, err := ssl.NewClientTLSConfig(cfg)

		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsConfig
	}
	act.client = client
	return act, nil
}

// Activity is an activity that is used to invoke a REST Operation
type Activity struct {
	settings *Settings
	client   *http.Client
}

func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Invokes a REST Operation
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	output := &Output{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}

	logger := ctx.Logger()

	endpoint := a.settings.Host

	if a.settings.Timeout > 0 {
		//Set the timeout only if it is positive
		logger.Infof("Client timeout is set to '%d' milliseconds", a.settings.Timeout)
		a.client.Timeout = time.Millisecond * time.Duration(a.client.Timeout)
	}

	if a.settings.Mode == "Proxy" {

		basePath := input.BasePath
		if basePath != "" {
			if strings.HasSuffix(endpoint, "/") {
				// Remove trailing /
				endpoint = endpoint[:len(endpoint)-1]
			}
			if !strings.HasPrefix(basePath, "/") {
				// Prepend / to base path
				basePath = "/" + basePath
			}
			if strings.HasSuffix(basePath, "/") {
				// Remove trailing /
				basePath = basePath[:len(basePath)-1]
			}
			endpoint = endpoint + basePath
		}

		oi := input.OpaqueData
		serverReq, ok := oi["requestObject"].(*http.Request)
		if !ok {
			logger.Errorf("Either 'proxyData' field is not set or configured with incorrect object. Check mapping.")
			return false, activity.NewError("Invalid 'proxyData' value", "", "")
		}

		serverWriter, ok := oi["responseObject"].(http.ResponseWriter)
		if !ok {
			logger.Errorf("Either 'proxyData' is not set or configured to incorrect object. Check mapping.")
			return false, activity.NewError("Invalid 'proxyData' value", "", "")
		}

		requestUrl := singleJoiningSlash(endpoint, serverReq.URL.RequestURI())

		// clientRequest, err := http.NewRequestWithContext(serverReq.Context(), serverReq.Method, requestUrl, serverReq.Body)

		// if err != nil {
		// 	return false, err
		// }

		clientRequest := serverReq.Clone(serverReq.Context())
		clientRequest.RequestURI = ""
		clientRequest.URL, _ = url.Parse(requestUrl)
		clientRequest.Host = clientRequest.URL.Host

		//copy headers

		//copyHeader(clientRequest.Header, serverReq.Header)

		for key, value := range input.Headers {
			clientRequest.Header.Set(key, value)
		}

		//Remove headers
		removeHeaders(clientRequest.Header, input.ExcludeRequestHeaders)

		logger.Infof("Sending request '%s %s'", clientRequest.Method, clientRequest.URL.String())
		logger.Debugf("Sending request headers - %v", clientRequest.Header)

		resp, err := a.client.Do(clientRequest)
		if err != nil {
			return false, err
		}

		logger.Info("Response status:", resp.Status)

		//Remove headers
		removeHeaders(resp.Header, input.ExcludeResponseHeaders)

		logger.Debugf("Sending response headers - %v", resp.Header)

		//copy headers from client response to server
		copyHeader(serverWriter.Header(), resp.Header)

		//Add headers
		//addHeaders(serverWriter.Header(), input.AddResponseHeaders, logger)
		serverWriter.WriteHeader(resp.StatusCode)
		if resp.Body != nil {
			_, err = io.Copy(serverWriter, resp.Body)
			if err != nil {
				return false, err
			}
			resp.Body.Close()
		}

		err = ctx.SetOutputObject(output)
		if err != nil {
			return false, err
		}
	} else if a.settings.Mode == "Data" {

		//populate path param
		for key, value := range input.PathParams {
			endpoint = strings.Replace(endpoint, "{"+key+"}", value, -1)
		}

		//populate Query param
		if len(input.QueryParams) > 0 {
			qp := url.Values{}

			for key, value := range input.QueryParams {
				qp.Set(key, value)
			}

			endpoint = endpoint + "?" + qp.Encode()
		}

		logger.Infof("Sending request '%s %s'", input.Method, endpoint)

		var reqBody io.Reader

		method := input.Method

		if method == methodPOST || method == methodPUT || method == methodPATCH {

			if input.RequestBody != nil {
				if str, ok := input.RequestBody.(string); ok {
					reqBody = bytes.NewBuffer([]byte(str))
				} else {
					b, _ := json.Marshal(input.RequestBody)
					reqBody = bytes.NewBuffer([]byte(b))
				}
			}
		} else {
			reqBody = nil
		}

		req, err := http.NewRequest(method, endpoint, reqBody)
		if err != nil {
			return false, err
		}

		//populate headers
		for key, value := range input.Headers {
			req.Header.Set(key, value)
		}

		if ctx.GetTracingContext() != nil {
			_ = trace.GetTracer().Inject(ctx.GetTracingContext(), trace.HTTPHeaders, req)
		}

		resp, err := a.client.Do(req)
		if err != nil {
			return false, err
		}

		if resp == nil {
			logger.Debugf("Empty response")
			return true, nil
		}

		defer func() {
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		logger.Info("Response status:", resp.Status)

		respHeaders := make(map[string]string, len(resp.Header))

		for key := range resp.Header {
			respHeaders[key] = resp.Header.Get(key)
		}

		var cookies []interface{}

		for _, cookie := range resp.Header["Set-Cookie"] {
			cookies = append(cookies, cookie)
		}

		var responseBody interface{}

		// Check the HTTP Header Content-Type
		respContentType := resp.Header.Get("Content-Type")
		switch respContentType {
		case "application/json":
			d := json.NewDecoder(resp.Body)
			d.UseNumber()
			err = d.Decode(&responseBody)
			if err != nil {
				switch {
				case err == io.EOF:
					// empty body
				default:
					return false, err
				}
			}
		default:
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}

			responseBody = string(b)
		}

		logger.Debug("Response body:", responseBody)

		output := &Output{StatusCode: resp.StatusCode, Headers: respHeaders, ResponseBody: responseBody, Cookies: cookies}
		err = ctx.SetOutputObject(output)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func removeHeaders(header http.Header, headers string) {
	if headers == "" {
		return
	}
	headerList := strings.Split(headers, ",")
	for _, hName := range headerList {
		if strings.HasSuffix(hName, "*") {
			var matchingList []string
			hName = strings.ToLower(hName[:len(hName)-1])
			for name := range header {
				if strings.HasPrefix(strings.ToLower(name), hName) {
					matchingList = append(matchingList, name)
				}
			}
			for _, v := range matchingList {
				header.Del(v)
			}
		} else {
			header.Del(hName)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////
// Utils

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case b == "/":
		return a
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
