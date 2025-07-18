package client

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Host    string `md:"host,required"` // The URI of the service to invoke
	Mode    string `md:"mode,required,allowed(Proxy,Data)"`
	Timeout int    `md:"timeout"`
}

type Input struct {
	BasePath               string                 `md:"contextPath"`
	OpaqueData             map[string]interface{} `md:"proxyData"`
	AddRequestHeaders      map[string]string      `md:"addRequestHeaders"`
	AddResponseHeaders     map[string]string      `md:"addResponseHeaders"`
	ExcludeRequestHeaders  string                 `md:"excludeRequestHeaders"`
	ExcludeResponseHeaders string                 `md:"excludeResponseHeaders"`
	Method                 string                 `md:"method"`
	PathParams             map[string]string      `md:"pathParams"`
	QueryParams            map[string]string      `md:"queryParams"`
	Headers                map[string]string      `md:"headers"`
	RequestBody            interface{}            `md:"requestBody"`
}

func (i *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"contextPath":            i.BasePath,
		"proxyData":              i.OpaqueData,
		"addRequestHeaders":      i.AddRequestHeaders,
		"addResponseHeaders":     i.AddResponseHeaders,
		"excludeRequestHeaders":  i.ExcludeRequestHeaders,
		"excludeResponseHeaders": i.ExcludeResponseHeaders,
		"method":                 i.Method,
		"pathParams":             i.PathParams,
		"queryParams":            i.QueryParams,
		"headers":                i.Headers,
		"requestBody":            i.RequestBody,
	}
}

func (i *Input) FromMap(values map[string]interface{}) error {
	var err error
	i.BasePath, err = coerce.ToString(values["contextPath"])
	if err != nil {
		return err
	}
	i.OpaqueData, err = coerce.ToObject(values["proxyData"])
	if err != nil {
		return err
	}
	i.AddRequestHeaders, err = coerce.ToParams(values["addRequestHeaders"])
	if err != nil {
		return err
	}
	i.AddResponseHeaders, err = coerce.ToParams(values["addResponseHeaders"])
	if err != nil {
		return err
	}
	i.ExcludeRequestHeaders, err = coerce.ToString(values["excludeRequestHeaders"])
	if err != nil {
		return err
	}
	i.ExcludeResponseHeaders, err = coerce.ToString(values["excludeResponseHeaders"])
	if err != nil {
		return err
	}
	i.Method, err = coerce.ToString(values["method"])
	if err != nil {
		return err
	}
	i.PathParams, err = coerce.ToParams(values["pathParams"])
	if err != nil {
		return err
	}
	i.QueryParams, err = coerce.ToParams(values["queryParams"])
	if err != nil {
		return err
	}
	i.Headers, err = coerce.ToParams(values["headers"])
	if err != nil {
		return err
	}
	i.RequestBody, err = coerce.ToAny(values["requestBody"])
	if err != nil {
		return err
	}
	return nil
}

type Output struct {
	Headers      map[string]string `md:"headers"`
	ResponseBody interface{}       `md:"responseBody"`
	StatusCode   int               `md:"statusCode"`
	Cookies      []interface{}     `md:"cookies"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"headers":      o.Headers,
		"responseBody": o.ResponseBody,
		"statusCode":   o.StatusCode,
		"cookies":      o.Cookies,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error
	o.Headers, err = coerce.ToParams(values["headers"])
	if err != nil {
		return err
	}
	o.ResponseBody, err = coerce.ToAny(values["responseBody"])
	if err != nil {
		return err
	}
	o.StatusCode, err = coerce.ToInt(values["statusCode"])
	if err != nil {
		return err
	}
	o.Cookies, err = coerce.ToArray(values["cookies"])
	if err != nil {
		return err
	}
	return nil
}
