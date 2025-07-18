package conversion

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/project-flogo/core/data/coerce"
)

func ConvertToParam(value interface{}) (map[string]string, error) {
	paramMap := map[string]string{}

	if value == nil {
		return paramMap, nil
	}

	switch t := value.(type) {
	case string:
		err := json.Unmarshal([]byte(t), &paramMap)
		if err != nil {
			return nil, err
		}
		return paramMap, nil
	default:
		return coerce.ToParams(value)
	}
	return paramMap, nil
}

func ConvertToInterface(value interface{}) (interface{}, error) {
	var paramMap interface{}

	if value == nil {
		return paramMap, nil
	}

	switch t := value.(type) {
	case string:
		err := json.Unmarshal([]byte(t), &paramMap)
		if err != nil {
			return nil, err
		}
		return paramMap, nil
	default:
		return value, nil
	}
	return paramMap, nil
}

func ConvertToObbject(value interface{}) (map[string]interface{}, error) {
	paramMap := map[string]interface{}{}

	if value == nil {
		return paramMap, nil
	}

	switch t := value.(type) {
	case string:
		err := json.Unmarshal([]byte(t), &paramMap)
		if err != nil {
			return nil, err
		}
		return paramMap, nil
	case map[string]interface{}:
		return coerce.ToObject(value)
	default:
		v, err := json.Marshal(t)
		if err != nil {
			return paramMap, err
		}
		err = json.Unmarshal(v, &paramMap)
		if err != nil {
			return nil, err
		}
		return paramMap, nil
	}
	return paramMap, nil
}

func ConvertToInt64(value interface{}) (int64, error) {
	switch t := value.(type) {
	case string:
		return strconv.ParseInt(t, 10, 64)
	case int:
		return int64(t), nil
	case int64:
		return t, nil
	case float64:
		return int64(t), nil
	default:
		v, err := json.Marshal(value)
		if err != nil {
			return 0, err
		}
		i, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, err
		}
		return i, nil
	}
}

func ConvertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case *int64:
		return int(*v), nil
	case *string:
		str := *v
		return strconv.Atoi(str)
	case *int:
		return *v, nil
	default:
		return coerce.ToInt(value)
	}
}

func ConvertToString(value interface{}) (string, error) {
	if value != nil {
		v, err := ConvertToBytes(value)
		if err != nil {
			return "", err
		}
		return string(v), nil
	}

	return "", nil
}

func ConvertToBytes(value interface{}) ([]byte, error) {
	if value != nil {
		switch t := value.(type) {
		case []byte:
			return t, nil
		case string:
			return []byte(t), nil
		default:
			return json.Marshal(value)
		}
	}
	return nil, fmt.Errorf("Cannot convert nil to []byte")
}
