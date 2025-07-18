package query

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/project-flogo/core/support/log"
)

// Convert JSON string to a map[string]interface{}
func jsonToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func TestConvertToQueryParams(t *testing.T) {
	// Define test cases as a table
	tests := []struct {
		name     string
		input    string // JSON input as string
		expected map[string][]string
	}{
		{
			name: "Basic_Types",
			input: `{
				"fields": ["field1", "field2"],
				"exclude_fields": ["exclude1", "exclude2"],
				"count": 10,
				"offset": 0,
				"type": "regular",
				"status": "active",
				"before_send_time": "2021-01-01T00:00:00Z",
				"since_send_time": "2021-02-01T00:00:00Z",
				"before_create_time": "2021-01-01T00:00:00Z",
				"since_create_time": "2021-02-01T00:00:00Z",
				"list_id": "list-1234",
				"folder_id": "folder-5678",
				"sort_field": "name",
				"sort_dir": "asc"
			}`,
			expected: map[string][]string{
				"fields":             {"field1", "field2"},
				"exclude_fields":     {"exclude1", "exclude2"},
				"count":              {"10"},
				"offset":             {"0"},
				"type":               {"regular"},
				"status":             {"active"},
				"before_send_time":   {"2021-01-01T00:00:00Z"},
				"since_send_time":    {"2021-02-01T00:00:00Z"},
				"before_create_time": {"2021-01-01T00:00:00Z"},
				"since_create_time":  {"2021-02-01T00:00:00Z"},
				"list_id":            {"list-1234"},
				"folder_id":          {"folder-5678"},
				"sort_field":         {"name"},
				"sort_dir":           {"asc"},
			},
		},
		{
			name:  "Empty_Slice",
			input: `{"fields": []}`,
			expected: map[string][]string{
				"fields": {},
			},
		},
		{
			name:     "Nil_Value",
			input:    `{"count": null}`,
			expected: map[string][]string{},
		},
		{
			name:  "Slice_with_Types_(Map)",
			input: `{"fields": [{"key": "value"}]}`,
			expected: map[string][]string{
				"fields": {"map[key:value]"},
			},
		},
		{
			name:  "Slice_with_Mixed_Types",
			input: `{"fields": [123, "field1", true, 45.67]}`,
			expected: map[string][]string{
				"fields": {"123", "field1", "true", "45.67"},
			},
		},
		{
			name: "Empty_String_Values",
			input: `{
				"type": "",
				"status": "",
				"before_send_time": "",
				"since_send_time": "",
				"before_create_time": "",
				"since_create_time": "",
				"list_id": "",
				"folder_id": "",
				"sort_field": "",
				"sort_dir": ""
			}`,
			expected: map[string][]string{
				"type":               {""},
				"status":             {""},
				"before_send_time":   {""},
				"since_send_time":    {""},
				"before_create_time": {""},
				"since_create_time":  {""},
				"list_id":            {""},
				"folder_id":          {""},
				"sort_field":         {""},
				"sort_dir":           {""},
			},
		},
		{
			name: "Missing_Optional_Params",
			input: `{
				"count": 20,
				"offset": 5,
				"status": "inactive"
			}`,
			expected: map[string][]string{
				"count":  {"20"},
				"offset": {"5"},
				"status": {"inactive"},
			},
		},
	}

	// Iterate over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert JSON string to map
			input, err := jsonToMap(tt.input)
			if err != nil {
				t.Fatalf("Failed to convert JSON to map: %v", err)
			}

			// Call the function to test
			output := ConvertToQueryParams(input, log.NewLogger("TEST LOG : "))

			// Compare the output with the expected result
			for key, expectedValue := range tt.expected {
				if fmt.Sprintf("%v", output[key]) != fmt.Sprintf("%v", expectedValue) {
					t.Errorf("Expected %v for key %s, got %v", expectedValue, key, output[key])
				}
			}
		})
	}
}
