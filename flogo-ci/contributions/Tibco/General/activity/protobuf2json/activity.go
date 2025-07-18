package protobuf2json

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

var logger log.Logger

// mapMD will be initialized by shim at build time
var mapDSBytes map[string][]byte

// Activity ...
type Activity struct {
	settings *Settings
}

func init() {
	_ = activity.Register(&Activity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// New ...
func New(ctx activity.InitContext) (activity.Activity, error) {
	settings := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), settings, true)
	if err != nil {
		return nil, err
	}
	activity := &Activity{}
	logger = ctx.Logger()
	activity.settings = settings
	return activity, nil
}

// Eval ...
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {
	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}
	inputMsgBytes, err := base64.StdEncoding.DecodeString(input.ProtoMessage)
	if err != nil {
		return false, activity.NewError("base64: Unable to decode input message", "", err)
	}
	activityName := strings.ToLower(ctx.Name())
	flowName := strings.ToLower(ctx.ActivityHost().Name())
	key := fmt.Sprintf("%s/%s", flowName, activityName)
	protoFileName := a.settings.ProtoFile["filename"].(string)
	message, err := getMessageFromDSBytes(mapDSBytes[key], protoFileName, a.settings.MessageTypeName)
	if err != nil {
		return false, activity.NewError("Unable to get message descriptor from registry", "", err)
	}

	err = proto.Unmarshal(inputMsgBytes, message)
	if err != nil {
		return false, activity.NewError("proto: Unable to unmarshal input message", "", err)
	}
	marshalOpts := protojson.MarshalOptions{}
	if a.settings.IncludeDefaultValues {
		marshalOpts = protojson.MarshalOptions{EmitUnpopulated: true}
	}
	jsonMsgBytes, err := marshalOpts.Marshal(message)
	// jsonMsgBytes, err := protojson.Marshal(message)
	if err != nil {
		return false, activity.NewError("protojson: Unable to marshal message", "", err)
	}
	output := &Output{}
	err = json.Unmarshal(jsonMsgBytes, &output.JSONMessage)
	if err != nil {
		return false, activity.NewError("json: Unable to unmarshal message", "", err)
	}
	ctx.SetOutputObject(output)
	return true, nil
}

func getMessageFromDSBytes(dsBytes []byte, protoFileName, messageTypeName string) (*dynamicpb.Message, error) {
	descriptorSet := descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(dsBytes, &descriptorSet)
	if err != nil {
		return nil, err
	}
	registry, err := protodesc.NewFiles(&descriptorSet)
	if err != nil {
		return nil, err
	}
	fileDesc, err := registry.FindFileByPath(protoFileName)
	if err != nil {
		return nil, err
	}
	messageDescriptors := fileDesc.Messages()
	lastDot := strings.LastIndexByte(messageTypeName, '.')
	if lastDot != -1 {
		messageTypeName = messageTypeName[lastDot+1:]
	}
	md := messageDescriptors.ByName(protoreflect.Name(messageTypeName))
	return dynamicpb.NewMessage(md), nil
}
