package jsonschema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaValidation(t *testing.T) {
	schema := `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "properties": {
        "pet": {
            "properties": {
                "category": {
                    "properties": {
                        "id": {
                            "type": "integer"
                        },
                        "name": {
                            "type": "string"
                        }
                    },
                    "type": "object"
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "photoUrls": {
                    "items": {
                        "type": "string"
                    },
                    "type": "array"
                },
                "status": {
                    "type": "string"
                },
                "tags": {
                    "items": {
                        "properties": {
                            "id": {
                                "type": "integer"
                            },
                            "name": {
                                "type": "string"
                            }
                        },
                        "type": "object"
                    },
                    "type": "array"
                }
            },
            "required": [
                "id",
                "name"
            ],
            "type": "object"
        }
    },
    "required": [
        "pet"
    ],
    "type": "object"
}`

	jsonData := `{
  "pet": {
    "name": "lixingwang",
    "category": {
      "id": 5,
      "name": "hahaha"
    },
    "id": "5"
  }
}`

	ok, err := validate(schema, jsonData)
	if err != nil {
		fmt.Println(err)
	}
	assert.False(t, ok)
}
