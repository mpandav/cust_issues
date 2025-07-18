package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var connectionObject = `
{  
   "id":"eec65890-a7a8-11e7-afde-87187d84888e",
   "type":"flogo:connector",
   "taskType":0,
   "version":"1.0.0",
   "name":"tibco-sqs",
   "title":"AWS SQS Connector",
   "description":"This is Amazon SQS connector",
   "ref":"github.com/TIBCOSoftware/tci-webintegrator/examples/AWS/connector/sqs",
   "settings":[  
      {  
         "name":"name",
         "type":"string",
         "required":true,
         "display":{  
            "name":"Connection Name",
            "description":"Name of the connection"
         },
         "value":"ads"
      },
      {  
         "name":"description",
         "type":"string",
         "display":{  
            "name":"Description",
            "description":"Connection description"
         },
         "value":""
      },
      {  
         "name":"accessKey",
         "type":"string",
         "required":true,
         "display":{  
            "name":"Access Key ID",
            "description":"AWS Access key ID for the user",
            "type":"password"
         },
         "value":"acacacaca"
      },
      {  
         "name":"secretKey",
         "type":"string",
         "required":true,
         "display":{  
            "name":"Secrete Access Key",
            "description":"AWS Secrete Access Key for the user",
            "type":"password"
         },
         "value":"axaxxaxaxaxaxaxaxxaxaxaxaxaxaxxa"
      },
      {  
         "name":"region",
         "type":"string",
         "required":true,
         "display":{  
            "name":"Region",
            "description":"Name of the region where SQS service is running"
         },
         "value":"us-west-2"
      }
   ],
   "outputs":[  

   ],
   "inputs":[  

   ],
   "handler":{  
      "settings":[  

      ]
   },
   "display":{  
      "description":"This is Amazon SQS connector",
      "category":"AWS",
      "visible":true,
      "smallIcon":"sqs.png"
   },
   "inputMappings":{  

   },
   "s3Prefix":"javycr6edqvqndtt3gkt5nt4ufneemrm",
   "actions":[  
      {  
         "name":"Connect",
         "display":{  
            "readonly":false,
            "valid":true,
            "visible":true
         }
      }
   ],
   "key":"javycr6edqvqndtt3gkt5nt4ufneemrm/AWS/connector/sqs/connector.json",
   "keyfield":"name",
   "isValid":true,
   "lastUpdatedTime":1506972966041,
   "createdTime":1506972966041,
   "user":"javycr6edqvqndtt3gkt5nt4ufneemrm",
   "connectorName":" ",
   "connectorDescription":" "
}
`

func TestConnection(t *testing.T) {
	conn, err := NewConnection(connectionObject)
	if conn == nil {
		t.Error(err)
		t.Fail()
		return
	}

	assert.Equal(t, conn.GetName(), "ads")
	assert.Equal(t, conn.GetId(), "eec65890-a7a8-11e7-afde-87187d84888e")
	assert.Equal(t, conn.GetAccessKey(), "acacacaca")
	assert.Equal(t, conn.GetRegion(), "us-west-2")
	assert.Equal(t, conn.GetSecretKey(), "axaxxaxaxaxaxaxaxxaxaxaxaxaxaxxa")
}

//func TestAssumeRole(t *testing.T) {
//	conf := aws.NewConfig()
//	conf.Credentials = credentials.NewStaticCredentials("AKIAR6ZF5NCSDGODFHOX", "+fiCwwI7pofEu78wy9sZZnzOUF+EbbDP6NTXmM53", "")
//	conf.Region = aws.String("us-east-1")
//
//	s := session.Must(session.NewSession(conf))
//	input := &sts.AssumeRoleInput{
//		DurationSeconds: aws.Int64(int64(900)),
//		RoleArn:         aws.String("arn:aws:iam::752037795865:role/flogo-assume-role"),
//		RoleSessionName: aws.String("flogo-assume-role-test"),
//		ExternalId:      aws.String("flogo-assume-role-test"),
//	}
//
//	clent := sts.New(s)
//
//	ou, err := clent.AssumeRole(input)
//	fmt.Println(err)
//	fmt.Println(ou)
//	if err != nil {
//		if aerr, ok := err.(awserr.Error); ok {
//			switch aerr.Code() {
//			case sts.ErrCodeMalformedPolicyDocumentException:
//				fmt.Println(sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
//			case sts.ErrCodePackedPolicyTooLargeException:
//				fmt.Println(sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
//			case sts.ErrCodeRegionDisabledException:
//				fmt.Println(sts.ErrCodeRegionDisabledException, aerr.Error())
//			default:
//				fmt.Println(aerr.Error())
//			}
//		} else {
//
//			fmt.Println(err.Error())
//		}
//		return
//	}
//
//}
//
//func TestLambdaAssumeRole(t *testing.T) {
//
//	con := &Connection{awsConnection: &awsConnection{}}
//	con.accessKey = "AKIAR6ZF5NCSDGODFHOX"
//	con.secretKey = "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"
//	con.region = "us-east-1"
//	con.roleArn = "arn:aws:iam::752037795865:role/flogo-assume-role"
//	con.roleSessionName = "lambdaFlogoSessionName"
//	con.externalID = "flogo-assume-role-test"
//	con.expirationDuration = 1000 * time.Second
//	con.assumeRole = true
//
//	ld := lambda.New(con.NewSession())
//	output, err := ld.ListFunctions(&lambda.ListFunctionsInput{})
//	fmt.Println(err)
//	fmt.Println(output.Functions)
//
//	for _, k := range output.Functions {
//		fmt.Print(k.FunctionName)
//	}
//
//}
