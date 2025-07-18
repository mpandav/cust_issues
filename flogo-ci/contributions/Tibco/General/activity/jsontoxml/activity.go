package jsontoxml

import (
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

	context.Logger().Debug("Executing JSONToXML activity")

	input := &Input{}
	err = context.GetInputObject(input)
	if err != nil {
		return false, err
	}

	if input.JsonString == "" {
		return false, errors.New("JSON string is empty. Provide valid input string.")
	}

	var jsonData = input.JsonString

	mapVal, merr := mxj.NewMapJson([]byte(jsonData))

	if merr != nil {
		return false, merr
	}

	_, merr = mapVal.Json(true)

	if merr != nil {
		return false, merr
	}

	xmlByte, merr := mapVal.Xml()
	if err != nil {
		return false, merr
	}

	output := &Output{}

	output.XmlString = string(xmlByte)

	err = context.SetOutputObject(output)
	if err != nil {
		return false, fmt.Errorf("Error setting output for Activity [%s]: %s", context.Name(), err.Error())
	}
	return true, nil
}
