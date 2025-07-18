package authorization

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestEncodedUrlBodyFromObject(t *testing.T) {

	values := map[string]interface{}{"hello world": "This is from China", "Hello flogo": "This is from/flogo"}
	str := encodedUrlBodyFromObject(values)
	assert.Equal(t, "Hello+flogo=This+is+from%2Fflogo&hello+world=This+is+from+China", str)

	values = map[string]interface{}{url.QueryEscape("hello world"): url.QueryEscape("This is from China"), "Hello+flogo": "This+is from/flogo"}
	str2 := encodedUrlBodyFromObject(values)
	assert.Equal(t, "Hello+flogo=This+is+from%2Fflogo&hello+world=This+is+from+China", str2)
}

func TestEncodedUrlBodyFromArray(t *testing.T) {
	value := `[
	{
		"name":"hello world",
		"value":"This is from China"
	},
	{
		"name":"Hello flogo",
		"value":"This is from/flogo"
	}
]`

	var values []interface{}
	json.Unmarshal([]byte(value), &values)

	str, _ := encodedUrlBodyFromArray(values)
	assert.Equal(t, "Hello+flogo=This+is+from%2Fflogo&hello+world=This+is+from+China", str)

	value = `[
	{
		"name":"hello+world",
		"value":"This+is+from+China"
	},
	{
		"name":"Hello+flogo",
		"value":"This+is+from%2Fflogo"
	}
]`

	var values2 []interface{}
	json.Unmarshal([]byte(value), &values2)

	str, _ = encodedUrlBodyFromArray(values2)
	assert.Equal(t, "Hello+flogo=This+is+from%2Fflogo&hello+world=This+is+from+China", str)
}
