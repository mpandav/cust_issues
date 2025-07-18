package s3util

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/log"
)

// UserMetadata stuct
type UserMetadata struct {
	Key   string `json:"Key" type:"string"`
	Value string `json:"Value" type:"string"`
}

// Metadata for Get Object
type Metadata struct {
	AcceptRanges            string         `json:"AcceptRanges" type:"string"`
	CacheControl            string         `json:"CacheControl" type:"string"`
	ContentDisposition      string         `json:"ContentDisposition" type:"string"`
	ContentEncoding         string         `json:"ContentEncoding" type:"string"`
	ContentLanguage         string         `json:"ContentLanguage" type:"string"`
	ContentLength           int64          `json:"ContentLength" type:"long"`
	ContentRange            string         `json:"ContentRange" type:"string"`
	ContentType             string         `json:"ContentType" type:"string"`
	DeleteMarker            bool           `json:"DeleteMarker" type:"boolean"`
	ETag                    string         `json:"ETag" type:"string"`
	Expiration              string         `json:"Expiration" type:"string"`
	Expires                 string         `json:"Expires" type:"string"`
	LastModified            time.Time      `json:"LastModified" type:"timestamp"`
	MissingMeta             int64          `json:"MissingMeta" type:"integer"`
	PartsCount              int64          `json:"PartsCount" type:"integer"`
	ReplicationStatus       string         `json:"ReplicationStatus" type:"string"`
	RequestCharged          string         `json:"RequestCharged" type:"string"`
	Restore                 string         `json:"Restore" type:"string"`
	SSECustomerAlgorithm    string         `json:"SSECustomerAlgorithm" type:"string"`
	SSECustomerKeyMD5       string         `json:"SSECustomerKeyMD5" type:"string"`
	SSEKMSKeyID             string         `json:"SSEKMSKeyID" type:"string"`
	ServerSideEncryption    string         `json:"ServerSideEncryption" type:"string"`
	StorageClass            string         `json:"StorageClass" type:"string"`
	TagCount                int64          `json:"TagCount" type:"integer"`
	UserMetadata            []UserMetadata `json:"UserMetadata" type:"array"`
	VersionID               string         `json:"VersionID" type:"string"`
	WebsiteRedirectLocation string         `json:"WebsiteRedirectLocation" type:"string"`
}

// Constants
const (
	// common configuration
	ConfConnection  = "connection"
	ConfServiceName = "serviceName"
	Input           = "input"
	Output          = "output"
	Error           = "error"

	constACL  = "ACL"
	constTags = "Tags"

	// mapping constants
	paramBucket              = "Bucket"
	paramDestinationFilePath = "DestinationFilePath"
	paramKey                 = "Key"
	paramMeta                = "Metadata"
	paramTextContent         = "TextContent"
	paramUserMetadata        = "UserMetadata"
	paramSourceFilePath      = "SourceFilePath"
	paramCopySource          = "CopySource"
	paramRequestPayer        = "RequestPayer"
	valueWrite               = "WRITE"
)

/************************* GET ACTIVITY ***********************************/

// GetObjectOutput builds output map for get
func GetObjectOutput(request *s3.GetObjectInput, result *s3.GetObjectOutput, destinationPath string, isText bool) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	output[paramBucket] = aws.StringValue(request.Bucket)
	output[paramKey] = aws.StringValue(request.Key)

	if !isText {
		output[paramDestinationFilePath] = destinationPath
	} else {
		b, err := ioutil.ReadAll(result.Body)
		if err == nil {
			output[paramTextContent] = string(b)
		} else {
			return nil, errors.New(GetMessage(FailedToConvertOutputToBytes, err.Error()))
		}
	}

	userMeta := []UserMetadata{}
	if len(result.Metadata) > 0 {
		for k, v := range result.Metadata {
			if v != nil {
				um := UserMetadata{Key: k, Value: *v}
				userMeta = append(userMeta, um)
			}
		}
	}

	metadataOutput := Metadata{
		AcceptRanges:            aws.StringValue(result.AcceptRanges),
		CacheControl:            aws.StringValue(result.CacheControl),
		ContentDisposition:      aws.StringValue(result.ContentDisposition),
		ContentEncoding:         aws.StringValue(result.ContentEncoding),
		ContentLanguage:         aws.StringValue(result.ContentLanguage),
		ContentLength:           aws.Int64Value(result.ContentLength),
		ContentRange:            aws.StringValue(result.ContentRange),
		ContentType:             aws.StringValue(result.ContentType),
		DeleteMarker:            aws.BoolValue(result.DeleteMarker),
		ETag:                    aws.StringValue(result.ETag),
		Expiration:              aws.StringValue(result.Expiration),
		Expires:                 aws.StringValue(result.Expires),
		LastModified:            aws.TimeValue(result.LastModified),
		MissingMeta:             aws.Int64Value(result.MissingMeta),
		PartsCount:              aws.Int64Value(result.PartsCount),
		ReplicationStatus:       aws.StringValue(result.ReplicationStatus),
		RequestCharged:          aws.StringValue(result.RequestCharged),
		Restore:                 aws.StringValue(result.Restore),
		SSECustomerAlgorithm:    aws.StringValue(result.SSECustomerAlgorithm),
		SSECustomerKeyMD5:       aws.StringValue(result.SSECustomerKeyMD5),
		SSEKMSKeyID:             aws.StringValue(result.SSEKMSKeyId),
		ServerSideEncryption:    aws.StringValue(result.ServerSideEncryption),
		StorageClass:            aws.StringValue(result.StorageClass),
		TagCount:                aws.Int64Value(result.TagCount),
		UserMetadata:            userMeta,
		VersionID:               aws.StringValue(result.VersionId),
		WebsiteRedirectLocation: aws.StringValue(result.WebsiteRedirectLocation),
	}

	output[paramMeta] = metadataOutput
	return output, nil
}

// DoListObjects action
func DoListObjects(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.ListObjectsOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.ListObjectsInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.ListObjects(request)
	return result, err
}

// DownloadFile downloads file to destinationPath
func DownloadFile(downloader *s3manager.Downloader, request *s3.GetObjectInput, destinationPath string) (int64, error) {
	// Create a new temporary file
	f, err := os.Create(filepath.Join(destinationPath, aws.StringValue(request.Key)))
	if err != nil {
		return 0, err
	}
	// Download the file to disk
	n, err := downloader.Download(f, request)
	if err != nil {
		return 0, err
	}
	return n, err
}

/************************* PUT ACTIVITY ***********************************/

// DoCreateBucket action
func DoCreateBucket(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.CreateBucketOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.CreateBucketInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.CreateBucket(request)
	return result, err
}

// DoCopyObject action
func DoCopyObject(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, isPreserveACL bool, log log.Logger) (*s3.CopyObjectOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.CopyObjectInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	// do user metadata
	userMeta := make(map[string]*string)
	b, _ := json.Marshal(inputObj[paramUserMetadata])
	var um []UserMetadata
	err = json.Unmarshal(b, &um)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	if len(um) > 0 {
		for i, m := range um {
			userMeta[m.Key] = &um[i].Value
		}
		request.SetMetadata(userMeta)
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.CopyObject(request)
	if isPreserveACL {
		err = DoPreserveObjectACL(s3Svc, inputObj, result)
		if err != nil {
			return nil, err
		}
	}
	return result, err
}

// DoPreserveObjectACL action
func DoPreserveObjectACL(s3Svc *s3.S3, inputObj map[string]interface{}, copyResult *s3.CopyObjectOutput) error {
	copySource := inputObj[paramCopySource].(string)
	var requestPayer string
	if inputObj[paramRequestPayer] != nil {
		requestPayer = inputObj[paramRequestPayer].(string)
	}
	idx := strings.Index(copySource, "/")
	if idx != -1 {
		srcBucket := copySource[0:idx]
		srcObject := copySource[idx+1 : len(copySource)]
		getObjACLInput := &s3.GetObjectAclInput{
			Bucket:       aws.String(srcBucket),
			Key:          aws.String(srcObject),
			RequestPayer: GetAwsString(requestPayer),
			VersionId:    copyResult.CopySourceVersionId,
		}
		objectACL, err := s3Svc.GetObjectAcl(getObjACLInput)
		if err != nil {
			return errors.New(GetMessage(ErrorGettingObjectProperty, constACL, srcObject, srcBucket, err.Error()))
		}
		tgtBucket := inputObj[paramBucket].(string)
		tgtObject := inputObj[paramKey].(string)
		putObjectACLInput := &s3.PutObjectAclInput{
			AccessControlPolicy: &s3.AccessControlPolicy{
				Grants: objectACL.Grants,
				Owner:  objectACL.Owner,
			},
			Bucket:       aws.String(tgtBucket),
			Key:          aws.String(tgtObject),
			RequestPayer: GetAwsString(requestPayer),
			VersionId:    copyResult.VersionId,
		}
		_, err = s3Svc.PutObjectAcl(putObjectACLInput)
		if err != nil {
			return errors.New(GetMessage(ErrorUpdatingObjectACL, err.Error()))
		}
	}
	return nil
}

// DoUploadObject action
func DoUploadObject(context activity.Context, s3Svc *s3.S3, uploader *s3manager.Uploader, inputObj map[string]interface{}, isText bool, log log.Logger) (*s3.PutObjectOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutObjectInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}

	if isText {
		textContent := inputObj[paramTextContent].(string)
		request.SetBody(strings.NewReader(textContent))
	} else {
		reader, err := os.Open(inputObj[paramSourceFilePath].(string))
		if err != nil {
			return nil, errors.New(GetMessage(UploadError, err.Error()))
		}
		defer reader.Close()
		request.SetBody(reader)
	}

	// do user metadata
	userMeta := make(map[string]*string)
	b, _ := json.Marshal(inputObj[paramUserMetadata])
	var um []UserMetadata
	err = json.Unmarshal(b, &um)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	if len(um) > 0 {
		for i, m := range um {
			userMeta[m.Key] = &um[i].Value
		}
		request.SetMetadata(userMeta)
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutObject(request)
	return result, err
}

/************************* DELETE ACTIVITY ***********************************/

// DoDeleteBucket action
func DoDeleteBucket(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.DeleteBucketOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.DeleteBucketInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.DeleteBucket(request)
	return result, err
}

// DoDeleteObject action
func DoDeleteObject(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.DeleteObjectOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.DeleteObjectInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.DeleteObject(request)
	return result, err
}

/************************* UPDATE ACTIVITY ***********************************/

// DoUpdateBucketACL action
func DoUpdateBucketACL(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutBucketAclOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutBucketAclInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	// Commenting this to be in line with AWS behaviour when updating ACL
	// // save request ACLS to map, to compare with existing values
	// rACLMap := make(map[string]*string)
	// if request.AccessControlPolicy != nil {
	// 	for _, grant := range request.AccessControlPolicy.Grants {
	// 		mKey := aws.StringValue(grant.Grantee.URI) + aws.StringValue(grant.Grantee.ID)
	// 		rACLMap[mKey] = grant.Permission
	// 	}
	// }
	// // FLOGO-2068
	// existingACLInput := &s3.GetBucketAclInput{
	// 	Bucket: request.Bucket,
	// }
	// existingACL, err := s3Svc.GetBucketAcl(existingACLInput)
	// if err != nil {
	// 	return nil, errors.New(GetMessage(ErrorGettingBucketProperty, constACL, aws.StringValue(request.Bucket), err.Error()))
	// }
	// if existingACL != nil {
	// 	log.Debug(GetMessage(ExistingProperty, constACL, existingACL.GoString()))
	// 	for _, grant := range existingACL.Grants {
	// 		eKey := aws.StringValue(grant.Grantee.URI) + aws.StringValue(grant.Grantee.ID)
	// 		if _, exists := rACLMap[eKey]; !exists {
	// 			// FGAZS3-47
	// 			if request.ACL == nil {
	// 				if request.AccessControlPolicy != nil {
	// 					if request.AccessControlPolicy.Grants != nil {
	// 						request.AccessControlPolicy.Grants = append(request.AccessControlPolicy.Grants, grant)
	// 					}
	// 				} else {
	// 					existingAccessControlPolicy := &s3.AccessControlPolicy{
	// 						Grants: existingACL.Grants,
	// 						Owner:  existingACL.Owner,
	// 					}
	// 					request.SetAccessControlPolicy(existingAccessControlPolicy)
	// 					break
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutBucketAcl(request)
	return result, err
}

// DoUpdateBucketCORS action
func DoUpdateBucketCORS(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutBucketCorsOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutBucketCorsInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutBucketCors(request)
	return result, err
}

// DoUpdateBucketPolicy action
func DoUpdateBucketPolicy(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutBucketPolicyOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutBucketPolicyInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutBucketPolicy(request)
	return result, err
}

// DoUpdateBucketVersioning action
func DoUpdateBucketVersioning(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutBucketVersioningOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutBucketVersioningInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutBucketVersioning(request)
	return result, err
}

// DoUpdateBucketWebsite action
func DoUpdateBucketWebsite(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutBucketWebsiteOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutBucketWebsiteInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutBucketWebsite(request)
	return result, err
}

// DoUpdateObjectACL action
func DoUpdateObjectACL(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutObjectAclOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutObjectAclInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	// // Commenting this to be in line with AWS behaviour when updating ACL
	// // save request ACLS to map, to compare with existing values
	// rACLMap := make(map[string]*string)
	// if request.AccessControlPolicy != nil { // FLOGO-2028
	// 	for _, grant := range request.AccessControlPolicy.Grants {
	// 		if grant != nil && aws.StringValue(grant.Permission) == valueWrite {
	// 			return nil, errors.New(GetMessage(InvalidWritePermissionForObject))
	// 		}
	// 		mKey := aws.StringValue(grant.Grantee.URI) + aws.StringValue(grant.Grantee.ID)
	// 		rACLMap[mKey] = grant.Permission
	// 	}
	// }
	// // FLOGO-2068
	// existingACLInput := &s3.GetObjectAclInput{
	// 	Bucket:       request.Bucket,
	// 	Key:          request.Key,
	// 	RequestPayer: request.RequestPayer,
	// 	VersionId:    request.VersionId,
	// }
	// existingACL, err := s3Svc.GetObjectAcl(existingACLInput)
	// if err != nil {
	// 	return nil, errors.New(GetMessage(ErrorGettingObjectProperty, constACL, aws.StringValue(request.Key), aws.StringValue(request.Bucket), err.Error()))
	// }
	// // if uri or id of existing acl is same as uri or id of acl in request, existing acl will be replaced by acl in request
	// if existingACL != nil {
	// 	log.Debug(GetMessage(ExistingProperty, constACL, existingACL.GoString()))
	// 	for _, grant := range existingACL.Grants {
	// 		eKey := aws.StringValue(grant.Grantee.URI) + aws.StringValue(grant.Grantee.ID)
	// 		if _, exists := rACLMap[eKey]; !exists {
	// 			request.AccessControlPolicy.Grants = append(request.AccessControlPolicy.Grants, grant)
	// 		}
	// 	}
	// }
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutObjectAcl(request)
	return result, err
}

// DoUpdateObjectTagging action
func DoUpdateObjectTagging(context activity.Context, s3Svc *s3.S3, inputObj map[string]interface{}, log log.Logger) (*s3.PutObjectTaggingOutput, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutObjectTaggingInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return nil, errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	// save request tags to map, to compare with existing values
	rTagMap := make(map[string]*string)
	for _, rTag := range request.Tagging.TagSet {
		rTagMap[aws.StringValue(rTag.Key)] = rTag.Value
	}
	// FLOGO-2068
	existingTagsInput := &s3.GetObjectTaggingInput{
		Bucket:    request.Bucket,
		Key:       request.Key,
		VersionId: request.VersionId,
	}
	existingTags, err := s3Svc.GetObjectTagging(existingTagsInput)
	if err != nil {
		return nil, errors.New(GetMessage(ErrorGettingObjectProperty, constTags, aws.StringValue(request.Key), aws.StringValue(request.Bucket), err.Error()))
	}
	if existingTags != nil {
		log.Debug(GetMessage(ExistingProperty, constTags, existingTags.GoString()))
		for _, tag := range existingTags.TagSet {
			if _, exists := rTagMap[aws.StringValue(tag.Key)]; !exists {
				request.Tagging.TagSet = append(request.Tagging.TagSet, tag)
			}
		}
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	err = request.Validate()
	if err != nil {
		return nil, errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	result, err := s3Svc.PutObjectTagging(request)
	return result, err
}

// DoGeneratePresignedURLGET action
func DoGeneratePresignedURLGET(context activity.Context, s3Svc *s3.S3, expirationTimeSec int64, inputObj map[string]interface{}, log log.Logger) (string, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return "", errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.GetObjectInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return "", errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	err = request.Validate()
	if err != nil {
		return "", errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	req, _ := s3Svc.GetObjectRequest(request)
	urlStr, err := req.Presign(time.Duration(expirationTimeSec) * time.Second)
	if err != nil {
		return "", errors.New(GetMessage(FailedInExecution, err.Error()))
	}
	return urlStr, err
}

// DoGeneratePresignedURLPUT action
func DoGeneratePresignedURLPUT(context activity.Context, s3Svc *s3.S3, expirationTimeSec int64, inputObj map[string]interface{}, log log.Logger) (string, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return "", errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.PutObjectInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return "", errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	err = request.Validate()
	if err != nil {
		return "", errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	req, _ := s3Svc.PutObjectRequest(request)
	urlStr, err := req.Presign(time.Duration(expirationTimeSec) * time.Second)
	if err != nil {
		return "", errors.New(GetMessage(FailedInExecution, err.Error()))
	}
	return urlStr, err
}

// DoGeneratePresignedURLDELETE action
func DoGeneratePresignedURLDELETE(context activity.Context, s3Svc *s3.S3, expirationTimeSec int64, inputObj map[string]interface{}, log log.Logger) (string, error) {
	reqBytes, err := json.Marshal(inputObj)
	if err != nil {
		return "", errors.New(GetMessage(FailedToConvertInputToBytes, err.Error()))
	}
	request := &s3.DeleteObjectInput{}
	err = json.Unmarshal(reqBytes, request)
	if err != nil {
		return "", errors.New(GetMessage(FailedToParseInputData, err.Error()))
	}
	err = request.Validate()
	if err != nil {
		return "", errors.New(GetMessage(FailedToValidateInputData, err.Error()))
	}
	log.Debug(GetMessage(ActivityInput, context.Name(), request.GoString()))
	req, _ := s3Svc.DeleteObjectRequest(request)
	urlStr, err := req.Presign(time.Duration(expirationTimeSec) * time.Second)
	if err != nil {
		return "", errors.New(GetMessage(FailedInExecution, err.Error()))
	}
	return urlStr, err
}
