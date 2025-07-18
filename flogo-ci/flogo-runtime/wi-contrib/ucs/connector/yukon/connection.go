package yukon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tibco/wi-contrib/environment"

	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	restClient "github.com/tibco/wi-contrib/ucs/common"
)

var logCache = log.ChildLogger(log.RootLogger(), "ucs.connection")
var factory = &YukonFactory{}

const UCS_PROVIDER_SRV_HEADER_NAME = "UCS_PROVIDER_SRV"

type YukonFactory struct {
}

type Settings struct {
	Action             string            `md:"action"`
	Name               string            `md:"name"`
	Description        string            `md:"description"`
	ConnectorName      string            `md:"connectorName"`
	URL                string            `md:"url"`
	ConnectorProps     map[string]string `md:"connectorProps"`
	ConnectionID       string            `md:"connectionId"`
	InstanceID         string            `md:"instanceId"`
	ProviderPathPrefix string            `md:"providerPathPrefix"`
}

type YukonSharedConfigManager struct {
	ConnectionID       string
	InstanceID         string
	ProviderPathPrefix string
	ConnectorName      string
	ConnectionName     string
	YukonClient        http.Client
	UCSProviderCookie  string
	Settings           *Settings
}

type OpenConnectionResponse struct {
	InstanceID         string `json:"instanceId"`
	ConnectionID       string `json:"connectionId"`
	ProviderPathPrefix string `json:"providerPathPrefix"`
}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

func (*YukonFactory) Type() string {
	return "yukon"
}

func (*YukonFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {
	// logger with logger name of connector name
	sharedConn := &YukonSharedConfigManager{}
	s := &Settings{}
	err := metadata.MapToStruct(settings, s, false)

	if err != nil {
		return nil, err
	}

	client, _ := restClient.GetHttpClient(logCache, 120)
	sharedConn, err = openConnection(client, s)
	if err != nil {
		return nil, err
	}

	sharedConn.YukonClient = client
	sharedConn.Settings = s

	return sharedConn, nil
}

func (m *YukonSharedConfigManager) Type() string {
	return "yukon"
}

func (m *YukonSharedConfigManager) GetConnection() interface{} {
	return m
}

func (m *YukonSharedConfigManager) GetHTTPClient() http.Client {
	return m.YukonClient
}

func (m *YukonSharedConfigManager) GetInstanceID() string {
	return m.InstanceID
}

func (m *YukonSharedConfigManager) GetConnectionID() string {
	return m.ConnectionID
}
func (m *YukonSharedConfigManager) GetProviderPathPrefix() string {
	return m.ProviderPathPrefix
}

func (m *YukonSharedConfigManager) GetSettings() interface{} {
	return m.Settings
}

func (m *YukonSharedConfigManager) ReleaseConnection(connection interface{}) {

}

func (m *YukonSharedConfigManager) Start() error {
	// fmt.Println("inside start")
	// client, _ := restClient.GetHttpClient(120)
	// sharedConn, err := openConnection(client, m.Settings)
	// if err != nil {
	// 	return err
	// }

	// sharedConn.YukonClient = client
	return nil
}

func (m *YukonSharedConfigManager) Stop() error {
	// logCache.Info("Cleaning up Connection")
	// // curl --location --request DELETE 'https://account.ucs.tcie.pro/ucs/provider/v1/instance/01FDQJMTD3SYPFJCMHK0WRQG3G?gsbc=01F6H55CTEQ2DRRSKANPCWK452' \
	// // 	--header 'Authorization: Bearer CIC~x4Mvn88_rcfhxIDrtRT0PT2e'

	// instanceID := m.InstanceID
	// intercomURL := environment.GetIntercomURL()
	// subUname := environment.GetTCISubscriptionUName()
	// // intercomURL := "https://account.ucs.tcie.pro"
	// subID := environment.GetTCISubscriptionId()
	// // subID := "01FRJR8Y5C7ANXFE1CPA5CMPRV"

	// logCache.Debugf("Intercom URL: %s", intercomURL)
	// logCache.Debugf("SubscriptionID: %s", subID)
	// logCache.Debugf("Subscription User: %s", subUname)

	// uri := intercomURL + "/" + "ucs/data/v1/instance/" + instanceID + "?gsbc=" + subID

	// logCache.Debugf("STOP Connection URI: %s", uri)
	// client, _ := restClient.GetHttpClient(120)

	// headers := make(map[string]string)
	// headers["X-Atmosphere-For-User"] = subUname
	// headers["X-Atmosphere-Tenant-Id"] = "ucs"
	// headers["X-Atmosphere-Subscription-Id"] = subID
	// headers["Content-Type"] = "application/json"
	// headers["Connection"] = "keep-alive"

	// resp, err := restClient.GetRestResponse(client, restClient.MethodDELETE, uri, headers, nil)
	// if err != nil {
	// 	logCache.Errorf("Error while closing the connection %s", err)
	// 	return err
	// }

	// if resp.Body != nil {
	// 	defer resp.Body.Close()
	// }

	return nil
}

func openConnection(client http.Client, setting *Settings) (*YukonSharedConfigManager, error) {
	sharedconn := &YukonSharedConfigManager{}
	if setting.ConnectionID == "" {
		return nil, fmt.Errorf("Connection ID is missing")
	}

	if setting.ConnectorName == "" {
		return nil, fmt.Errorf("Connection Name is missing")
	} else {
		sharedconn.ConnectorName = setting.ConnectorName
	}

	intercomURL := environment.GetIntercomURL()

	// intercomURL := "https://account.ucs.tcie.pro"
	subID := environment.GetTCISubscriptionId()
	subUname := environment.GetTCISubscriptionUName()

	uri := intercomURL + "/" + "ucs/data/v1/instance?gsbc=" + subID
	//query param = subid
	postBodyJSON, _ := json.Marshal(map[string]string{
		"connectionId": setting.ConnectionID,
	})
	postBody := bytes.NewBuffer(postBodyJSON)

	headers := make(map[string]string)
	headers["X-Atmosphere-For-User"] = subUname
	headers["X-Atmosphere-Tenant-Id"] = "ucs"
	headers["X-Atmosphere-Subscription-Id"] = subID
	headers["Content-Type"] = "application/json"
	headers["Connection"] = "keep-alive"

	resp, err := restClient.GetRestResponse(logCache, client, restClient.MethodPOST, uri, headers, postBody)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		logCache.Error(readerr)
		return nil, err
	}
	openConnectionResponse := OpenConnectionResponse{}
	jsonerr := json.Unmarshal(body, &openConnectionResponse)
	if jsonerr != nil {
		logCache.Error(jsonerr)
		return nil, err
	}

	var cookies []*http.Cookie
	var ucsProviderServerCookie string
	cookies = resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == UCS_PROVIDER_SRV_HEADER_NAME {
			ucsProviderServerCookie = cookie.Value
		}
	}

	sharedconn.InstanceID = openConnectionResponse.InstanceID
	sharedconn.ConnectionID = openConnectionResponse.ConnectionID
	sharedconn.ProviderPathPrefix = openConnectionResponse.ProviderPathPrefix
	sharedconn.UCSProviderCookie = ucsProviderServerCookie

	return sharedconn, nil
}
