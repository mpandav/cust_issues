package mailchimp

import (
	"encoding/json"
	"strings"
)


func InputObjectToStr(input interface{}) (string, error) {
	switch input.(type) {
	case string:
		return input.(string), nil
	default:
		b, err := json.Marshal(input)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
func ArrayToQueryParameters(params []string) string {
	return strings.Join(params, ",")
}
