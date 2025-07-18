package bayeux

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	projectFlogoLogger "github.com/project-flogo/core/support/log"
)

// TriggerEvent describes an event received from Bayeaux Endpoint

type TriggerEvent struct {
	Data struct {
		Schema  string      `json:"schema"`
		Payload interface{} `json:"payload"`
		Event   struct {
			CreatedDate string `json:"createdDate,omitempty"`
			ReplayID    int    `json:"replayId"`
			Type        string `json:"type,omitempty"`
		} `json:"event"`
		Object json.RawMessage `json:"sobject,omitempty"`
	} `json:"data,omitempty"`
	Error      string `json:"error,omitemtpy"`
	ClientID   string `json:"clientId,omitempty"`
	Channel    string `json:"channel"`
	Successful bool   `json:"successful,omitempty"`
	Advice     struct {
		Interval  int    `json:"interval,omitempty"`
		Reconnect string `json:"reconnect,omitempty"`
	} `json:"advice,omitempty"`
}

func (t TriggerEvent) topic() string {
	s := strings.Replace(t.Channel, "/topic/", "", 1)
	return s
}

// Status is the state of success and subscribed channels
type Status struct {
	connected bool
	clientID  string
	channels  []string
}

type BayeuxHandshake []struct {
	Ext struct {
		Replay bool `json:"replay"`
	} `json:"ext"`
	MinimumVersion           string   `json:"minimumVersion"`
	ClientID                 string   `json:"clientId"`
	SupportedConnectionTypes []string `json:"supportedConnectionTypes"`
	Channel                  string   `json:"channel"`
	Version                  string   `json:"version"`
	Successful               bool     `json:"successful"`
}

type Subscription struct {
	ClientID     string `json:"clientId"`
	Channel      string `json:"channel"`
	Subscription string `json:"subscription"`
	Successful   bool   `json:"successful"`
	Error        string `json:"error"`
}

type Credentials struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	IssuedAt    int
	ID          string
	TokenType   string `json:"token_type"`
	Signature   string
	APIVersion  string
}

func (c Credentials) bayeuxUrl() string {
	return c.InstanceURL + "/cometd/" + c.APIVersion[1:]
}

type clientIDAndCookies struct {
	clientID string
	cookies  []*http.Cookie
}

// Bayeux struct allow for centralized storage of creds, ids, and cookies
type Bayeux struct {
	creds Credentials
	id    clientIDAndCookies
}

var wg sync.WaitGroup
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
var status = Status{false, "", []string{}}
var bayeuxLibraryLogger = projectFlogoLogger.ChildLogger(projectFlogoLogger.RootLogger(), "salesforce.vendor.bayeux")

// Call is the base function for making bayeux requests
func (b *Bayeux) call(body string, route string) (resp *http.Response, e error) {
	var jsonStr = []byte(body)
	req, err := http.NewRequest("POST", route, bytes.NewBuffer(jsonStr))
	if err != nil {
		logger.Fatalf("Bad Call request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", b.creds.AccessToken))
	// Per Stackexchange comment, passing back cookies is required though undocumented in Salesforce API
	// We were unable to get process working without passing cookies back to SF server.
	// SF Reference: https://developer.salesforce.com/docs/atlas.en-us.api_streaming.meta/api_streaming/intro_client_specs.htm
	for _, cookie := range b.id.cookies {
		req.AddCookie(cookie)
	}

	//bayeuxLibraryLogger.Debugf("REQUEST: %#v", req)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err == io.EOF {
		// Right way to handle EOF?
		bayeuxLibraryLogger.Debugf("Bad bayeuxCall io.EOF: %s\n", err)
		bayeuxLibraryLogger.Debugf("Bad bayeuxCall Response: %+v\n", resp)
	} else if err != nil {
		e = errors.New(fmt.Sprintf("Unknown error: %s", err))
		bayeuxLibraryLogger.Debugf("Bad unrecoverable Call: %s", err)
	}
	return resp, e
}

func (b *Bayeux) getClientID() error {
	handshake := `{"channel": "/meta/handshake", "supportedConnectionTypes": ["long-polling"], "version": "1.0"}`
	//var id clientIDAndCookies
	// Stub out clientIDAndCookies for first bayeuxCall
	resp, err := b.call(handshake, b.creds.bayeuxUrl())
	if err != nil {
		logger.Fatalf("Cannot get client id %s", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var h BayeuxHandshake
	if err := decoder.Decode(&h); err == io.EOF {
		logger.Fatal(err)
	} else if err != nil {
		logger.Fatal(err)
	}
	clientIDAndCookiesCreds := clientIDAndCookies{h[0].ClientID, resp.Cookies()}
	b.id = clientIDAndCookiesCreds
	bayeuxLibraryLogger.Debug("Credential clietID is : ", clientIDAndCookiesCreds.clientID)
	if clientIDAndCookiesCreds.clientID == "" {
		bayeuxLibraryLogger.Debug("ClientID is not present in handshake response")
		return nil
	}
	return nil
}

// ReplayAll replay for past 24 hrs
const ReplayAll = -2

// ReplayNone start playing events at current moment
const ReplayNone = -1

// Replay accepts the following values
// Value
// -2: replay all events from past 24 hrs
// -1: start at current
// >= 0: start from this event number
type Replay struct {
	Value int
}

func (b *Bayeux) subscribe(eType string, eName string, replay Replay) Subscription {
	handshake := fmt.Sprintf(`{
								"channel": "/meta/subscribe",
								"subscription": "/%s/%s",
								"clientId": "%s",
								"ext": {
									"replay": {"/%s/%s": %d}
									}
								}`, eType, eName, b.id.clientID, eType, eName, replay.Value)
	if replay.Value == 0 {
		handshake = fmt.Sprintf(`{
								"channel": "/meta/subscribe",
								"subscription": "/%s/%s",
								"clientId": "%s"
							}`, eType, eName, b.id.clientID)
	}

	if os.Getenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG") == "true" {
		bayeuxLibraryLogger.Debugf("Bayeux subscribing data: %s", handshake)
	}

	resp, err := b.call(handshake, b.creds.bayeuxUrl())
	if err != nil {
		logger.Fatalf("Cannot subscribe %s", err)
	}

	defer resp.Body.Close()
	if os.Getenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG") == "true" {
		bayeuxLibraryLogger.Debugf("Response: %+v", resp)
		// // Read the content
		var b []byte
		if resp.Body != nil {
			b, _ = ioutil.ReadAll(resp.Body)
		}
		// Restore the io.ReadCloser to its original state
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		// Use the content
		s := string(b)
		bayeuxLibraryLogger.Debug("Response Body: ", s)
	}

	if resp.StatusCode > 299 {
		logger.Fatalf("Received non 2XX response: HTTP_CODE %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	var h []Subscription
	if err := decoder.Decode(&h); err == io.EOF {
		logger.Fatal(err)
	} else if err != nil {
		logger.Fatal(err)
	}
	if h[0].Successful == false && strings.HasPrefix(h[0].Error, "400::") {
		logger.Fatal(h[0].Error)
	}
	sub := h[0]
	status.connected = sub.Successful
	status.clientID = sub.ClientID
	status.channels = append(status.channels, eName)
	bayeuxLibraryLogger.Debug("Established connection(s): %+v", status)
	// bayeuxLibraryLogger.Debug("Subscription : ", sub)
	return sub
}

func (b *Bayeux) connect(signChan chan int) (chan interface{}, chan string) {
	out := make(chan interface{})
	errChan := make(chan string)
	go func() {
		// TODO: add stop chan to bring this thing to halt
		for {
			signal := <-signChan
			switch signal {
			case -1:
				// halt
				bayeuxLibraryLogger.Error("Received signal = -1")
				bayeuxLibraryLogger.Error("Halt long connection for the stale bayeux client")
			case 1:
				bayeuxLibraryLogger.Debug("Received signal = 1")
				postBody := fmt.Sprintf(`{"channel": "/meta/connect", "connectionType": "long-polling", "clientId": "%s"} `, b.id.clientID)
				resp, err := b.call(postBody, b.creds.bayeuxUrl())
				if err != nil {
					bayeuxLibraryLogger.Errorf("Cannot connect to bayeux %s", err)
					bayeuxLibraryLogger.Error("Trying again...")
					errChan <- err.Error()
				} else {
					defer resp.Body.Close()
					if os.Getenv("SF_BAYEUX_DEBUG") == "true" {
						// // Read the content
						var b []byte
						if resp.Body != nil {
							b, _ = ioutil.ReadAll(resp.Body)
						}
						// Restore the io.ReadCloser to its original state
						resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
						// Use the content
						s := string(b)
						bayeuxLibraryLogger.Debug("Response Status : ", resp.Status)
						bayeuxLibraryLogger.Debug("Response Body : ", s)
						if s == "" {
							bayeuxLibraryLogger.Error("Response Body is empty")
							errChan <- "Empty response body"
						}
					}
					var x []interface{}
					decoder := json.NewDecoder(resp.Body)
					if err := decoder.Decode(&x); err != nil && err != io.EOF {
						//logger.Fatal(err)
						bayeuxLibraryLogger.Errorf("Unable to handle event due to error - %s", err.Error())
						errChan <- err.Error()
					} else {
						for _, e := range x {
							out <- e
						}
					}
				}
			}
		}
	}()
	return out, errChan
}

func GetSalesforceCredentials() Credentials {
	route := "https://login.salesforce.com/services/oauth2/token"
	clientID := mustGetEnv("SALESFORCE_CONSUMER_KEY")
	clientSecret := mustGetEnv("SALESFORCE_CONSUMER_SECRET")
	username := mustGetEnv("SALESFORCE_USER")
	password := mustGetEnv("SALESFORCE_PASSWORD")
	params := url.Values{"grant_type": {"password"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"username":      {username},
		"password":      {password}}
	res, err := http.PostForm(route, params)
	if err != nil {
		logger.Fatal(err)
	}
	decoder := json.NewDecoder(res.Body)
	var creds Credentials
	if err := decoder.Decode(&creds); err == io.EOF {
		logger.Fatal(err)
	} else if err != nil {
		logger.Fatal(err)
	} else if creds.AccessToken == "" {
		logger.Fatalf("Unable to fetch access token. Check credentials in environmental variables")
	}
	return creds
}

func mustGetEnv(s string) string {
	r := os.Getenv(s)
	if r == "" {
		panic(fmt.Sprintf("Could not fetch key %s", s))
	}
	return r
}

func (b *Bayeux) SetCreds(creds Credentials) {
	b.creds = creds
}

func (b *Bayeux) GetClientIDAndCookies() clientIDAndCookies {
	return b.id
}

func (b *Bayeux) GetClientID() string {
	return b.id.clientID
}
func (b *Bayeux) Unsubscribe(eType string, eName string) error {
	postBody := fmt.Sprintf(`{ "channel": "/meta/unsubscribe", "subscription": "/%s/%s", "clientId": "%s" }`, eType, eName, b.id.clientID)

	resp, err := b.call(postBody, b.creds.bayeuxUrl())
	if err != nil {
		bayeuxLibraryLogger.Debugf("Request failed due to error %s", err)
		return err
	}
	defer resp.Body.Close()
	bayeuxLibraryLogger.Debug("Response Status : ", resp.Status)
	var by []byte
	if resp.Body != nil {
		by, _ = ioutil.ReadAll(resp.Body)
	}
	// Restore the io.ReadCloser to its original state
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(by))
	// Print the content
	// bayeuxLibraryLogger.Debug("Response Body : %s\n", string(by))
	return nil
}

func (b *Bayeux) SubscribeToChannel(creds Credentials, subscriberType string, eName string, replayID int, signalChan chan int) (chan interface{}, chan string) {
	b.creds = creds
	err := b.getClientID()
	if err != nil {
		log.Fatal("Unable to get bayeux ClientId")
	}
	if b.id.clientID == "" {
		bayeuxLibraryLogger.Debug("Failed to get handshake ClientID")
		return nil, nil
	}
	r := Replay{}
	if replayID != 0 {
		bayeuxLibraryLogger.Debug("ReplayID field is present : ", replayID)
		r.Value = replayID
	} else {
		bayeuxLibraryLogger.Debug("ReplayID field is absent : ", replayID)
		r.Value = ReplayNone
	}
	bayeuxLibraryLogger.Debug("handshake() completed")

	if subscriberType == "PushTopic" {
		b.subscribe("topic", eName, r)
	} else if subscriberType == "Change Data Capture" {
		b.subscribe("data", eName, r)
	} else if subscriberType == "Platform Event" {
		b.subscribe("event", eName, r)
	}
	bayeuxLibraryLogger.Debug("subscribe() completed")

	c, errChan := b.connect(signalChan)
	bayeuxLibraryLogger.Debug("connect() completed")

	wg.Add(1)
	return c, errChan
}
