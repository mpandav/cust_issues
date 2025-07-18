package google

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var connectionObject = `
{  
   "id":"12c237d0-2fbe-11e8-aab3-2714b6f3173f",
   "type":"flogo:connector",
   "version":"1.0.0",
   "name":"tibco-google",
   "inputMappings":{  

   },
   "outputMappings":{  

   },
   "title":"Google Connector",
   "description":"This is Google connector",
   "ref":"git.tibco.com/git/product/ipaas/wi-plugins.git/contributions/chargify/connector/chargify",
   "settings":[  
      {  
         "name":"name",
         "type":"string",
         "required":true,
         "display":{  
            "name":"Connection Name",
            "description":"Name of the connection"
         },
         "value":"mygoogle"
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
         "name":"seviceaccountkey",
         "type":"string",
         "required":true,
         "display":{  
            "name":"Service Account Key",
            "description":"Paste the contents of the service account key in this field.\nThis should have be a JSON formatted file",
            "visible":true,
            "type":"fileselector",
            "fileExtensions":[  
               ".json"
            ]
         },
		"value": "data:application/json;base64,ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAiZmxvZ28tMTIzNCIsCiAgInByaXZhdGVfa2V5X2lkIjogImY4MWRnNTk2ZjE5ZjMyYmY2OTk2NTJlNWVhMmI2YzRkZDIyODVkMTkiLAogICJwcml2YXRlX2tleSI6ICItLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS1cbk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQ2JYUUlUeWpleHozUktcbmdTZjA3V01HOVRuREFCdXVzekpQY20rYUV5cmZPTGlMeitIYzhzTzhhRmN4dXpQR1h5RHc0eVF0WnFBVUlHazZcbmM1NXV4Y3J0MGU5RFBaVk5OL2oxWnNzZzJWMjM5YkZhTEcrTzFsSjZnWVh6TzFaQzNiemtvb2lLVnhOZUZCY0Fcbm1Jb2xqSkxOTzhmdkYwejRpKzBBN293dEVVSU55VDZHeUxNL3kzUHBjTFdvSzQ4VWpHbzU3V3Vzc3VGanBmaGxcbmt1eDJ5bG14RlFuYVpRUVRxd29sZkQ5ZXphY0V4RnljME9vMWJBYlJQa3FPTTRTalFBbWJzYjhWUkRjR0hVclBcbi9BNlNzS2lTc0x2UE9oU1Y2OFNJSm5Oek83N2x0c3pUcjJQUjZ1ZExqTUZzaFFhWkNHZWhPMERYYlA1MUhVbWZcbmZBL3lzRG9EQWdNQkFBRUNnZ0VBR0VYVXNFRFVxTHdQb0NCRG5ObUZzaTJYNDZaZHJOS2tWcE03YW1mNk43dkZcbjRWb09JSlh4REx1RWUrbVNjamlramQzKzVmVDFwNDlVd1dRVTZaVlBnZzVkZ2pUWjRhR1FETThNKzAxYWZnWXRcbnVqZmRDZ1RrQjNXOTlyMWJnY0RmNFNNZmovV0F1aDhMWlBWd0IrUEpmN1VLVElsb1ppQitXN25wUHBWR3E1NTZcbkFoVytlVmZkMTVSanVnbFFNRXY5STltcVhhRFY2Q0xQblVDNXhFL2UydEJHeTc4Ky9QOHRUSkJsOEFaRkIwaG9cbjF5MEdjaGVkYWFzOTJmSmxHbm5nTFdtbU8zNzhhRDJzRFUreHZZLy9MYTNsRThwK2FDMS9VQ2hEdGNZUVhVWlBcbnduRWJjT25LemxyUmdjczVoUGRPc1lpdHYxWVkxQzJRWUU0MDVNV2t5UUtCZ1FET3VCMTlDeUk0M1R6ZzFtQW5cblMyOHVpZGl2Znc0WXpwUWRwckdQcUU0OWtTb3A5YlZoZHVVSXJob3ZpWGc3Wm1PVzQ1VDkxRGRHSDlZLzR2WDRcbmNJMWp3TDlSbnQ5MlVnRVFWcGlGZW01STFrU3ErTFIvTnFPekdvelQ1empNUzBqZFlOUjVGaWpGbjAxSGdiMW9cbldkYUxSU2E0MzdzOVd6VStSa2pHUGlteWxRS0JnUURBWnJDWktjYkhablNhNE5IczZJbXV6WmEzcFVBVjF5N3BcbjN0bUsxaE5ybysxVnAzcWkrenFDZkVRZkV3SnZtdkVWVjNwV3d4MDJlTzg5NjFERjlQZWpaZGNDUDVVdlZXeU5cbnpzZExRZDZ5R2llS1NGNWY3TmtGSEUxK2ZpUU1KYk45eE1xZVJvUWJoTURrYit5cDk5MzhjUmtURkswVElKamxcbjByS1lhZFJzTndLQmdRQ1NhZEZwQ1lPNXB1bEJqbFVZUDlPRnNOaXFwR0VGclBzM2JTT0NUb0RzRm04NHZQRTFcbkVSTHpiT3piRXBEMzhYTkVJZmtiTnozWEN5R2lxa3Z4SlRiZm1sdG5vaEZBS3FEYVE1dFBud0dSMFVGZG56Mm9cbmhMaTVXR3E2ZzZDMUFmV2Y1cjlXN0IwQXErMytZYVFYenRtb1Z0Z3dSVGJISkZ5M3VPdytqVFRYYVFLQmdBbHJcbnZjL3lITHFjeUs3Z3ZVYTFhREIzL3A1RmFDcnBtM0YyS1A3RVZyVVpsTUJ4Nys1VkVOdGN6RlVkTUN4WTBOOHpcbnBsampPdVgwNi9vRE1MUlF0Mk4zMUJ4WEVxMzdwOUlWd3VwcmNrVVVSTVZmbjhkZ3FJdTRoQTdpakU5UDlVYitcblFOR1pNRlRNbmtsUk5heG81NlM1d1BtUE5KNVFKVXh6a2EwbTJYRG5Bb0dBUkY0aXArbmtjZXh0SUZxMUtXNFhcbmRXUGRHUFVXMXpMRU1MSTYvMVZ3WXRBKyt0TXF4RTc2MzN1T0hlZWF5ek9KMnZkb25VaUQ0NmorVEUvWHlJaU5cblRHVVJIZ1BITEhNZVc0c1dKamVCTTZMYWdWOUlOZEN2eXBrMEl1L1VGLy95b2toNml2emFJbG9oWWZqc0lmUjlcbnBNZGEvRTNvWUxtZWJxSVRPTE1ncnJZPVxuLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLVxuIiwKICAiY2xpZW50X2VtYWlsIjogImRhdGFzdG9yZUBmbG9nby0xMjM0LmlhbS5nc2VydmljZWFjY291bnQuY29tIiwKICAiY2xpZW50X2lkIjogIjEwMTQyNzQxMjk4NjkzMTY0Njc5MSIsCiAgImF1dGhfdXJpIjogImh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbS9vL29hdXRoMi9hdXRoIiwKICAidG9rZW5fdXJpIjogImh0dHBzOi8vYWNjb3VudHMuZ29vZ2xlLmNvbS9vL29hdXRoMi90b2tlbiIsCiAgImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAogICJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9yb2JvdC92MS9tZXRhZGF0YS94NTA5L2RhdGFzdG9yZSU0MGJ3Y2UtMTg3NzE5LmlhbS5nc2VydmljZWFjY291bnQuY29tIgp9Cg=="
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
   "reply":[  

   ],
   "s3Prefix":"flogo",
   "key":"flogo/Chargify/connector/chargify/connector.json",
   "display":{  
      "description":"This is Chargify connector",
      "category":"Chargify",
      "visible":true,
      "smallIcon":"chargify.png"
   },
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
   "keyfield":"name",
   "isValid":true,
   "lastUpdatedTime":1521935403981,
   "createdTime":1521935403981,
   "user":"flogo",
   "subscriptionId":"flogo_sbsc",
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

	assert.Equal(t, conn.GetName(), "mygoogle")
	assert.Equal(t, conn.GetId(), "12c237d0-2fbe-11e8-aab3-2714b6f3173f")
	svcAccntKey := conn.GetServiceAccountKey()
    assert.NotNil(t, svcAccntKey)
	assert.Equal(t, svcAccntKey.GetProjectId(), "flogo-1234")

}
