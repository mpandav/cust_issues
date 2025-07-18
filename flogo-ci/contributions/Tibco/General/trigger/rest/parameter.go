package rest

import (
	"encoding/json"
	"strings"

	"github.com/tibco/wi-contrib/engine/conversion"
)

type Parameter struct {
	Name      string `json:"parameterName"`
	Type      string `json:"type"`
	Repeating string `json:"repeating,omitempty"`
	Required  string `json:"required,omitempty"`
}

func ParseParams(paramSchema map[string]interface{}) ([]Parameter, error) {

	if paramSchema == nil {
		return nil, nil
	}

	var parameter []Parameter

	//Structure expected to be JSON schema like
	props := paramSchema["properties"].(map[string]interface{})
	for k, v := range props {
		param := &Parameter{}
		param.Name = k

		//if the k is required or not
		requiredKeys, ok := paramSchema["required"].([]interface{})
		if ok {
			for ran := range requiredKeys {
				if strings.EqualFold(k, requiredKeys[ran].(string)) {
					param.Required = "true"
					break
				}

			}
		}

		propValue := v.(map[string]interface{})
		for k1, v1 := range propValue {
			if k1 == "type" {
				if v1 != "array" {
					param.Repeating = "false"
					param.Type = v1.(string)
				}
			} else if k1 == "items" {
				param.Repeating = "true"
				items := v1.(map[string]interface{})
				s, err := conversion.ConvertToString(items["type"])
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

func LoadJsonSchemaFromMetadata(valueIN interface{}) (map[string]interface{}, error) {
	if valueIN != nil {
		complex := valueIN.(map[string]interface{})["value"]
		if complex != nil {
			params, err := convertToMap(complex)
			if err != nil {
				return nil, err
			}
			return params, nil
		}
	}
	return nil, nil
}

func convertToMap(data interface{}) (map[string]interface{}, error) {
	switch t := data.(type) {
	case string:
		if t != "" {
			m := map[string]interface{}{}
			err := json.Unmarshal([]byte(t), &m)
			if err != nil {
				return nil, err
			}
			return m, nil
		}
	case map[string]interface{}:
		return t, nil
	case interface{}:
		b, err := json.Marshal(t)
		if err != nil {
			return nil, err
		}
		m := map[string]interface{}{}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, err
		}
		return m, nil
	}

	return nil, nil
}
