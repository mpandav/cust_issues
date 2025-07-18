package rest

import (
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support/connection"
)

// Input struct for activity input
type Input struct {
	AsrEnabled           bool                   `md:"enableASR"`
	Authorization        bool                   `md:"authorization"`
	AuthorizationConn    connection.Manager     `md:"authorizationConn"`
	ServiceName          string                 `md:"serviceName"`
	ServiceMetadata      map[string]interface{} `md:"serviceMetadata"`
	ResourcePath         string                 `md:"resourcePath"`
	Method               string                 `md:"Method,required"`
	URI                  string                 `md:"Uri,required"`
	DisableKeepAlives    bool                   `md:"disableKeepAlives"`
	FollowRedirects      bool                   `md:"followRedirects"`
	Timeout              int                    `md:"Timeout"`
	RequestType          string                 `md:"requestType,required"`
	CertificateProvided  bool                   `md:"Use certificate for verification,required"`
	MutualAuth           bool                   `md:"mutualAuth"`
	ServerCertificate    string                 `md:"Server Certificate,required"`
	ClientCertificate    string                 `md:"Client Certificate"`
	ClientKey            string                 `md:"Client Key"`
	DisableSslValidation bool                   `md:"disableSslValidation"`
	Proxy                string                 `md:"proxy"`
	Host                 string                 `md:"host"`
	QueryParams          map[string]interface{} `md:"queryParams"`
	PathParams           map[string]interface{} `md:"pathParams"`
	Headers              map[string]interface{} `md:"headers"`
	Body                 interface{}            `md:"body"`
	FormData             map[string]interface{} `md:"multipartFormData"`
	MultipartForm        map[string]interface{} `md:"multipartForm"`
}

// ToMap conversion
func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"enableASR":                        i.AsrEnabled,
		"authorization":                    i.Authorization,
		"authorizationConn":                i.AuthorizationConn,
		"serviceName":                      i.ServiceName,
		"serviceMetadata":                  i.ServiceMetadata,
		"resourcePath":                     i.ResourcePath,
		"Method":                           i.Method,
		"Uri":                              i.URI,
		"disableKeepAlives":                i.DisableKeepAlives,
		"followRedirects":                  i.FollowRedirects,
		"Timeout":                          i.Timeout,
		"requestType":                      i.RequestType,
		"Use certificate for verification": i.CertificateProvided,
		"mautualAuth":                      i.MutualAuth,
		"Server Certificate":               i.ServerCertificate,
		"Client Certificate":               i.ClientCertificate,
		"Client Key":                       i.ClientKey,
		"disableSslValidation":             i.DisableSslValidation,
		"proxy":                            i.Proxy,
		"host":                             i.Host,
		"queryParams":                      i.QueryParams,
		"pathParams":                       i.PathParams,
		"headers":                          i.Headers,
		"body":                             i.Body,
		"multipartFormData":                i.FormData,
		"multipartForm":                    i.MultipartForm,
	}
}

// FromMap conversion
func (i *Input) FromMap(values map[string]interface{}) error {
	var err error

	i.AsrEnabled, err = coerce.ToBool(values["enableASR"])
	if err != nil {
		return err
	}

	i.Authorization, err = coerce.ToBool(values["authorization"])
	if err != nil {
		return err
	}

	if values["authorizationConn"] != nil {
		i.AuthorizationConn, err = coerce.ToConnection(values["authorizationConn"])
		if err != nil {
			return err
		}
	}

	i.ServiceName, err = coerce.ToString(values["serviceName"])
	if err != nil {
		return err
	}

	i.ServiceMetadata, err = coerce.ToObject(values["serviceMetadata"])
	if err != nil {
		return err
	}

	i.ResourcePath, err = coerce.ToString(values["resourcePath"])
	if err != nil {
		return err
	}

	i.Method, err = coerce.ToString(values["Method"])
	if err != nil {
		return err
	}

	i.URI, err = coerce.ToString(values["Uri"])
	if err != nil {
		return err
	}

	i.DisableKeepAlives, err = coerce.ToBool(values["disableKeepAlives"])
	if err != nil {
		return err
	}

	i.FollowRedirects, err = coerce.ToBool(values["followRedirects"])
	if err != nil {
		return err
	}

	i.Timeout, err = coerce.ToInt(values["Timeout"])
	if err != nil {
		return err
	}

	i.RequestType, err = coerce.ToString(values["requestType"])
	if err != nil {
		return err
	}

	i.CertificateProvided, err = coerce.ToBool(values["Use certificate for verification"])
	if err != nil {
		return err
	}
	i.MutualAuth, _ = coerce.ToBool(values["mutualAuth"])

	i.ServerCertificate, err = coerce.ToString(values["Server Certificate"])
	if err != nil {
		return err
	}

	i.ClientCertificate, err = coerce.ToString(values["Client Certificate"])
	if err != nil {
		return err
	}

	i.ClientKey, err = coerce.ToString(values["Client Key"])
	if err != nil {
		return err
	}

	i.DisableSslValidation, err = coerce.ToBool(values["disableSSLVerification"])
	if err != nil {
		return err
	}

	i.Proxy, err = coerce.ToString(values["proxy"])
	if err != nil {
		return err
	}

	i.Host, err = coerce.ToString(values["host"])
	if err != nil {
		return err
	}

	i.QueryParams, err = coerce.ToObject(values["queryParams"])
	if err != nil {
		return err
	}

	i.PathParams, err = coerce.ToObject(values["pathParams"])
	if err != nil {
		return err
	}

	i.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}

	i.FormData, err = coerce.ToObject(values["multipartFormData"])
	if err != nil {
		return err
	}

	i.Body = values["body"]

	i.MultipartForm, err = coerce.ToObject(values["multipartForm"])
	if err != nil {
		return err
	}

	return nil
}

// Output struct for activity output
type Output struct {
	StatusCode             int                    `md:"statusCode"`
	ConfigureResponseCodes bool                   `md:"configureResponseCodes"`
	ThrowError             bool                   `md:"throwError"`
	ResponseType           string                 `md:"responseType"`
	ResponseBody           interface{}            `md:"responseBody"`
	ResponseTime           int64                  `md:"responseTimeInMilliSec"`
	ResponseCodes          map[string]interface{} `md:"responseCodes"`
	ResponseCodesSchema    map[string]interface{} `md:"responseCodesSchema"`
	Headers                map[string]interface{} `md:"headers"`
}

// ToMap conversion
func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"statusCode":             o.StatusCode,
		"configureResponseCodes": o.ConfigureResponseCodes,
		"throwError":             o.ThrowError,
		"responseType":           o.ResponseType,
		"responseBody":           o.ResponseBody,
		"responseTimeInMilliSec": o.ResponseTime,
		"responseCodes":          o.ResponseCodes,
		"responseCodesSchema":    o.ResponseCodesSchema,
		"headers":                o.Headers,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.StatusCode, err = coerce.ToInt(values["statusCode"])
	if err != nil {
		return err
	}

	o.ConfigureResponseCodes, err = coerce.ToBool(values["configureResponseCodes"])
	if err != nil {
		return err
	}

	o.ThrowError, err = coerce.ToBool(values["throwError"])
	if err != nil {
		return err
	}

	o.ResponseType, err = coerce.ToString(values["responseType"])
	if err != nil {
		return err
	}

	o.ResponseBody = values["responseBody"]
	if err != nil {
		return err
	}

	o.ResponseTime, err = coerce.ToInt64(values["responseTimeInMilliSec"])
	if err != nil {
		return err
	}

	o.ResponseCodes, err = coerce.ToObject(values["responseCodes"])
	if err != nil {
		return err
	}

	o.ResponseCodesSchema, err = coerce.ToObject(values["responseCodesSchema"])
	if err != nil {
		return err
	}

	o.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}

	return nil
}
