package snsutil

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
)

// MAttributeValue ...
type MAttributeValue struct {
	BinaryValue []byte `json:"BinaryValue,omitempty"`
	DataType    string `json:"DataType"`
	StringValue string `json:"StringValue,omitempty"`
}

// MessageAttributeMap ...
type MessageAttributeMap struct {
	AttributeName string `json:"AttributeName,required"`
	AttributeType string `json:"AttributeType,required"`
}

// IsFatalError checks for AWS Error
func IsFatalError(err error) bool {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			errMap := make(map[string]interface{})
			errMap["code"] = awsErr.Code()
			errMap["message"] = awsErr.Message()
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				// A service error occured
				errMap["statusCode"] = reqErr.StatusCode()
				errMap["requestId"] = reqErr.RequestID()
			}
		}
		return true
	}
	return false
}

// ConvertMAttributes ...
func ConvertMAttributes(messageAttributeNames []MessageAttributeMap, attribValueMap map[string]interface{}) map[string]MAttributeValue {
	mAttributes := make(map[string]MAttributeValue)
	for _, mAttrib := range messageAttributeNames {
		if mAttrib.AttributeType == "Binary" {
			mAttributes[mAttrib.AttributeName] = MAttributeValue{
				DataType:    mAttrib.AttributeType,
				BinaryValue: []byte(attribValueMap[mAttrib.AttributeName].(string)),
			}
		} else {
			mAttributes[mAttrib.AttributeName] = MAttributeValue{
				DataType:    mAttrib.AttributeType,
				StringValue: attribValueMap[mAttrib.AttributeName].(string),
			}
		}
	}
	return mAttributes
}
