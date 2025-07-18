package parsejson

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/project-flogo/core/data/schema"
	"github.com/tibco/wi-contrib/engine/jsonschema"

	"github.com/project-flogo/core/activity"
)

// ParseJSONActivity is an activity that parses JSON string into JSON Object.
type Activity struct {
	metadata *activity.Metadata
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

func init() {
	_ = activity.Register(&Activity{})
}

func (a *Activity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Info("Executing ParseJSON activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	var jData []byte

	jsonData := strings.TrimSpace(input.JsonString)
	if strings.HasPrefix(jsonData, "{") || strings.HasPrefix(jsonData, "[") {
		jData = []byte(jsonData)
	} else if strings.HasPrefix(jsonData, "file:") {
		//Read data from file
		filePath := jsonData[7:]
		context.Logger().Infof("Reading file from [%s]", filePath)
		jData, err = ioutil.ReadFile(filePath)
		if err != nil {
			context.Logger().Error(err)
			return false, activity.NewError("Failed read JSON data from file. Ensure that file exists and path is prefixed with 'file://'.", "", nil)
		}
	} else {
		return false, activity.NewError("Invalid input. It must be a valid JSON string or a file path prefixed with 'file://'.", "", nil)
	}

	output := &Output{}

	err = json.Unmarshal(jData, &output.JsonObject)
	if err != nil {
		context.Logger().Error(err)
		return false, activity.NewError("Failed to parse JSON data", "", nil)
	}

	if input.Validate {

		if sIO, ok := context.(schema.HasSchemaIO); ok {
			s := sIO.GetOutputSchema(ovJSONObject)
			if s != nil {
				jSchema := s.Value()
				context.Logger().Debugf("Object schema for validation: %s", jSchema)
				err := jsonschema.ForceValidateFromObject(jSchema, output.JsonObject)
				if err != nil {
					context.Logger().Error(err)
					return false, activity.NewError("Output validation failed. JSON data does not match with the schema", "", nil)
				}
			}
		}

	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, err
	}
	context.Logger().Info("ParseJSON activity completed")
	return true, nil
}
