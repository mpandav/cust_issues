package query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/support/log"
)

// Parameters ...
type QueryOpts struct {
	Headers      []*TypedValue `json:"headers"`
	Parameters   []*TypedValue `json:"pathParams"`
	QueryOptions []*TypedValue `json:"queryParams"`
	// RequestType    string
	// ResponseType   string
	// ResponseOutput string
}

// String ...
func (p *QueryOpts) String(log log.Logger) string {
	v, err := json.Marshal(p)
	if err != nil {
		log.Errorf("Parameter object to string err %s", err.Error())
		return ""
	}
	return string(v)
}

// TypedValue ....
type TypedValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

// Param ...
type Param struct {
	Name      string
	Type      string
	Required  string
	Repeating string
}

// ParseQueryOptions ...
func ParseQueryOpts(queryOptsSchema map[string]interface{}) ([]Param, error) {

	if queryOptsSchema == nil {
		return nil, nil
	}

	var parameter []Param

	//Structure expected to be JSON schema like

	props := queryOptsSchema["properties"].(map[string]interface{})

	for k, v := range props {
		param := &Param{}
		param.Name = k
		propValue := v.(map[string]interface{})
		for k1, v1 := range propValue {
			if k1 == "required" {
				param.Required = v1.(string)
			} else if k1 == "type" {
				if v1 != "array" {
					param.Repeating = "false"
					param.Type = v1.(string)
				}
			} else if k1 == "items" {
				param.Repeating = "true"
				items := v1.(map[string]interface{})
				s, err := coerce.ToString(items["type"])
				if err != nil {
					return nil, err
				}
				param.Type = s
			}
		}
		parameter = append(parameter, *param)
	}

	return parameter, nil
}

// ToString ...
func (t *TypedValue) ToString(log log.Logger) string {
	if t.Value != nil {
		v, err := coerce.ToString(t.Value)
		if err != nil {
			log.Error("Typed value %+v to string error %s", t, err.Error())
			return ""
		}
		return v
	}
	return ""
}

// GetQueryOpts ...
func GetQueryOpts(context activity.Context, input *Input, log log.Logger) (opts *QueryOpts, err error) {
	opts = &QueryOpts{}

	//Headers
	log.Debug("Reading request headers")
	headersMap, err := LoadJsonSchemaFromInput(context, "headers")
	if err != nil {
		return nil, fmt.Errorf("error loading headers input schema: %s", err.Error())
	}

	if headersMap != nil {
		headers, err := ParseQueryOpts(headersMap)
		if err != nil {
			return opts, err
		}

		if headers != nil {
			inputHeaders := input.Headers
			var typeValuesHeaders []*TypedValue
			for _, hParam := range headers {
				isRequired := hParam.Required
				paramName := hParam.Name
				if isRequired == "true" && inputHeaders[paramName] == nil {
					return nil, fmt.Errorf("Required header parameter [%s] is not configured", paramName)
				}
				if inputHeaders[paramName] != nil {
					if hParam.Repeating == "true" {
						val := inputHeaders[paramName]
						switch reflect.TypeOf(val).Kind() {
						case reflect.Slice:
							s := reflect.ValueOf(val)

							typeValue := &TypedValue{}
							typeValue.Name = paramName
							typeValue.Type = hParam.Type

							var multiVs []string
							for i := 0; i < s.Len(); i++ {
								sv, _ := coerce.ToString(s.Index(i).Interface())

								multiVs = append(multiVs, sv)
							}

							typeValue.Value = strings.Join(multiVs, ",")
							typeValuesHeaders = append(typeValuesHeaders, typeValue)
						}
					} else {
						typeValue := &TypedValue{}
						typeValue.Name = paramName
						typeValue.Value = inputHeaders[paramName]
						typeValue.Type = hParam.Type
						typeValuesHeaders = append(typeValuesHeaders, typeValue)
					}
					opts.Headers = typeValuesHeaders
				}
			}
		}
	}

	//Query Options
	log.Debug("Reading query options")
	queryOptionsMap, err := LoadJsonSchemaFromInput(context, "queryOptions")
	if err != nil {
		return nil, fmt.Errorf("error loading queryOptions input schema: %s", err.Error())
	}

	if queryOptionsMap != nil {
		queryParams, err := ParseQueryOpts(queryOptionsMap)
		if err != nil {
			return opts, err
		}

		if queryParams != nil {
			inputQueries := input.QueryOptions
			var typeValuesQueries []*TypedValue
			for _, qParam := range queryParams {
				isRequired := qParam.Required
				paramName := qParam.Name
				if isRequired == "true" && inputQueries[paramName] == nil {
					return nil, fmt.Errorf("required query option [%s] is not configured", paramName)
				}

				if inputQueries[paramName] != nil {

					typeValue := &TypedValue{}
					typeValue.Name = paramName
					typeValue.Value = inputQueries[paramName]
					typeValue.Type = qParam.Type
					typeValuesQueries = append(typeValuesQueries, typeValue)

					opts.QueryOptions = typeValuesQueries
				}
			}
		}

	}

	//Parameters
	log.Debug("Reading parameters aliases")
	parametersMap, err := LoadJsonSchemaFromInput(context, "parameters")
	if err != nil {
		return nil, fmt.Errorf("error loading parameters input schema: %s", err.Error())
	}
	if parametersMap != nil {
		pathParams, err := ParseQueryOpts(parametersMap)
		if err != nil {
			return opts, err
		}
		if pathParams != nil {
			inputPathParams := input.Parameters
			var typeValuesPath []*TypedValue
			for _, pParam := range pathParams {
				paramName := pParam.Name
				if pParam.Required == "true" && inputPathParams[paramName] == nil {
					return nil, fmt.Errorf("required parameter [%s] is not configured", paramName)
				}

				typeValue := &TypedValue{}
				typeValue.Name = paramName
				typeValue.Value = inputPathParams[paramName]
				typeValue.Type = pParam.Type
				typeValuesPath = append(typeValuesPath, typeValue)
				opts.Parameters = typeValuesPath

			}
		}
	}

	//fmt.Println(opts)

	return opts, err
}

// LoadJsonSchemaFromInput ...
func LoadJsonSchemaFromInput(context activity.Context, attributeName string) (map[string]interface{}, error) {
	var schemaModel schema.Schema
	if sIO, ok := context.(schema.HasSchemaIO); ok {
		schemaModel = sIO.GetInputSchema(attributeName)
		if schemaModel != nil {
			return coerce.ToObject(schemaModel.Value())
		}
	}
	return nil, nil
}

// LoadJsonSchemaFromOutput ...
func LoadJsonSchemaFromOutput(context activity.Context, attributeName string) (map[string]interface{}, error) {
	var schemaModel schema.Schema
	if sIO, ok := context.(schema.HasSchemaIO); ok {
		schemaModel = sIO.GetOutputSchema(attributeName)
		if schemaModel != nil {
			return coerce.ToObject(schemaModel.Value())
		}
	}
	return nil, nil
}
