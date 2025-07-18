package rest

import (
	"github.com/project-flogo/core/data/coerce"
)

type Settings struct {
	Port             int    `md:"port"`
	APISpecUpdate    bool   `md:"APISpecUpdate"`
	Swagger          string `md:"swagger"`
	SwaggerVersion   string `md:"swaggerVersion"`
	SecureConnection string `md:"secureConnection"`
	ServerKey        string `md:"serverKey"`
	CaCertificate    string `md:"caCertificate"`
}

type HandlerSettings struct {
	APISpecPath      string `md:"APISpecPath"`
	Method           string `md:"Method"`
	Path             string `md:"Path"`
	OutputValidation bool   `md:"OutputValidation"`
}

// Output struct for trigger output
type Output struct {
	QueryParams map[string]interface{} `md:"queryParams"`
	PathParams  map[string]interface{} `md:"pathParams"`
	Headers     map[string]interface{} `md:"headers"`
	Body        interface{}            `md:"body"`
	Multipart   map[string]interface{} `md:"multipartFormData"`
	RequestURI  string                 `md:"requestURI"`
	Method      string                 `md:"method"`
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"queryParams":       o.QueryParams,
		"pathParams":        o.PathParams,
		"headers":           o.Headers,
		"body":              o.Body,
		"multipartFormData": o.Multipart,
		"requestURI":        o.RequestURI,
		"method":            o.Method,
	}
}

// FromMap conversion
func (o *Output) FromMap(values map[string]interface{}) error {

	var err error
	o.QueryParams, err = coerce.ToObject(values["queryParams"])
	if err != nil {
		return err
	}

	o.PathParams, err = coerce.ToObject(values["pathParams"])
	if err != nil {
		return err
	}

	o.Headers, err = coerce.ToObject(values["headers"])
	if err != nil {
		return err
	}

	o.Multipart, err = coerce.ToObject(values["multipartFormData"])
	if err != nil {
		return err
	}

	o.Body = values["body"]

	o.RequestURI, err = coerce.ToString(values["requestURI"])
	if err != nil {
		return err
	}

	o.Method, err = coerce.ToString(values["method"])
	if err != nil {
		return err
	}

	return nil
}

type Reply struct {
	Code                   int                    `md:"code"`
	ConfigureResponseCodes bool                   `md:"configureResponseCodes"`
	Data                   interface{}            `md:"data"`
	ResponseBody           map[string]interface{} `md:"responseBody"`
	ResponseCodesSchema    map[string]interface{} `md:"responseCodesSchema"`
	Message                string                 `md:"message"`
}

func (o *Reply) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":                   o.Code,
		"configureResponseCodes": o.ConfigureResponseCodes,
		"data":                   o.Data,
		"responseBody":           o.ResponseBody,
		"responseCodesSchema":    o.ResponseCodesSchema,
		"message":                o.Message,
	}
}

// FromMap conversion
func (o *Reply) FromMap(values map[string]interface{}) error {

	var err error
	o.Code, err = coerce.ToInt(values["code"])
	if err != nil {
		return err
	}

	o.ConfigureResponseCodes, err = coerce.ToBool(values["configureResponseCodes"])
	if err != nil {
		return err
	}

	o.ResponseBody, err = coerce.ToObject(values["responseBody"])
	if err != nil {
		return err
	}

	o.Data = values["data"]

	o.ResponseCodesSchema, err = coerce.ToObject(values["responseCodesSchema"])
	if err != nil {
		return err
	}

	o.Message, err = coerce.ToString(values["message"])
	if err != nil {
		return err
	}

	return nil
}
