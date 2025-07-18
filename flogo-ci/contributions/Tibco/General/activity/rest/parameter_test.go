package rest

var name = `{
    "headers": [
        {
            "name": "Content-Type",
            "value": "application/json"
        }
    ],
    "pathParams": [
        {
            "name": "ID",
            "value": 10001
        },
        {
            "name": "Name",
            "value": "dddddd"
        }
    ],
    "queryParams": [
        {
            "name": "query",
            "value": "queryString"
        }
    ]
}`

// func TestParseParameter(t *testing.T) {
// 	param, err := ParseParameters(name)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, param)
// 	assert.Equal(t, "10001", param.PathParams[0].ToString())
// }

// func TestBuildURI(t *testing.T) {
// 	param, err := ParseParameters(name)
// 	assert.Nil(t, err)
// 	url := buildURI("/pet/{ID}/{Name}", param)
// 	assert.Equal(t, "/pet/10001/dddddd?query=queryString", url)
// 	fmt.Println(url)
// }

// func TestBuildURI2(t *testing.T) {
// 	param, err := ParseParameters(name)
// 	assert.Nil(t, err)
// 	url := buildURI("/pet/{ID}", param)
// 	assert.Equal(t, "/pet/10001?query=queryString", url)
// 	fmt.Println(url)
// }
