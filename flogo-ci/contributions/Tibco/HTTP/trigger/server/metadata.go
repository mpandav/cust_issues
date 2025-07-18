package server

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Port int    `md:"port,required"` // The port to listen on
	Mode string `md:"processingMode"`
}

type HandlerSettings struct {
	ContextPath string `md:"contextPath"` // The base path
	Method      string `md:"reqMethod"`   // The HTTP method (ie. GET,POST,PUT,PATCH or DELETE)
}

type Output struct {
	ProxyData   map[string]interface{} `md:"proxyData"`   // The proxy mode
	PathParams  map[string]string      `md:"pathParams"`  // Path params
	QueryParams map[string]string      `md:"queryParams"` // The query parameters (e.g., 'id' in http://.../pet?id=someValue )
	Headers     map[string]string      `md:"headers"`     // The HTTP header parameters
	RequestBody interface{}            `md:"requestBody"` // The content of the request
	Method      string                 `md:"method"`      // The HTTP method used for the request
}

type Reply struct {
	StatusCode   int               `md:"statusCode"`   // The http code to reply with
	Headers      map[string]string `md:"headers"`      // The HTTP response headers
	ResponseBody interface{}       `md:"responseBody"` // The data to reply with
	Cookies      []interface{}     `md:"cookies"`      // "The response cookies, adds `Set-Cookie` headers"
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"proxyData":   o.ProxyData,
		"pathParams":  o.PathParams,
		"queryParams": o.QueryParams,
		"headers":     o.Headers,
		"method":      o.Method,
		"requestBody": o.RequestBody,
	}
}

func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.ProxyData, err = coerce.ToObject(values["proxyData"])
	if err != nil {
		return err
	}
	o.PathParams, err = coerce.ToParams(values["pathParams"])
	if err != nil {
		return err
	}
	o.QueryParams, err = coerce.ToParams(values["queryParams"])
	if err != nil {
		return err
	}
	o.Headers, err = coerce.ToParams(values["headers"])
	if err != nil {
		return err
	}
	o.Method, err = coerce.ToString(values["method"])
	if err != nil {
		return err
	}
	o.RequestBody = values["requestBody"]

	return nil
}

func (r *Reply) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"statusCode":   r.StatusCode,
		"headers":      r.Headers,
		"responseBody": r.ResponseBody,
		"cookies":      r.Cookies,
	}
}

func (r *Reply) FromMap(values map[string]interface{}) error {

	var err error
	r.StatusCode, err = coerce.ToInt(values["statusCode"])
	if err != nil {
		return err
	}

	r.Headers, err = coerce.ToParams(values["headers"])
	if err != nil {
		return err
	}

	r.ResponseBody = values["responseBody"]

	r.Cookies, err = coerce.ToArray(values["cookies"])
	if err != nil {
		return err
	}
	return nil
}
