package xmltojson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	mxj "github.com/clbanning/mxj/v2"
	"github.com/project-flogo/core/activity"
)

// XMLToJSONActivity is an activity which converts XML string into JSON Object.
type Activity struct {
}

var activityMd = activity.ToMetadata(&Input{}, &Output{})

func init() {
	_ = activity.Register(&Activity{})
}

func New(ctx activity.InitContext) (activity.Activity, error) {
	return &Activity{}, nil
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

func (a *Activity) Eval(context activity.Context) (done bool, err error) {

	context.Logger().Debug("Executing XMLToJSON activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	if input.XmlString == "" {
		return false, errors.New("XML string is empty. Provide valid input string.")
	}

	var xmlData = input.XmlString

	var jsonByte []byte

	// Preserve ordering of XML elements
	if input.Ordered {
		msv2, err := mxj.NewMapXmlSeq([]byte(xmlData), input.TypeCast)
		if err != nil {
			return false, err
		}

		jsonByte, err = mapseqToJson(msv2, false)
		if err != nil {
			return false, err
		}

	} else {
		mv, err := mxj.NewMapXml([]byte(xmlData), input.TypeCast)
		if err != nil {
			return false, err
		}
		jsonByte, err = mv.Json(false)
		if err != nil {
			return false, err
		}
	}

	output := &Output{}

	err = json.Unmarshal(jsonByte, &output.JsonObject)
	if err != nil {
		context.Logger().Error(err)
		return false, activity.NewError("Failed to parse JSON data", "", nil)
	}

	err = context.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("Error setting output for Activity [%s]: %s", context.Name(), err.Error())
	}
	return true, nil
}

func mapseqToJson(msv mxj.MapSeq, safeEncoding ...bool) ([]byte, error) {
	var s bool
	if len(safeEncoding) == 1 {
		s = safeEncoding[0]
	}

	b, err := json.Marshal(msv)

	if !s {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}
