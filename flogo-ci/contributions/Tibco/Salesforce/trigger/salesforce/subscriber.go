package salesforce

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/project-flogo/core/trigger"

	sfconnection "github.com/tibco/wi-salesforce/src/app/Salesforce/connector"
	bay "github.com/zph/bayeux"
)

const REFRESH_NUMBER = 3

type Subscriber struct {
	topic                         PushTopic
	changeDataCapture             ChangeDataCapture
	platformEvent                 PlatformEvent
	salesforceSharedConfigManager *sfconnection.SalesforceSharedConfigManager
	handler                       trigger.Handler
	retryCount                    int
	topicId                       string
	replayID                      int
	subscriberType                string
	autoCreatePushTopic           bool
	bayeuxClient                  *bay.Bayeux
}

type PushTopic struct {
	Name                       string
	Query                      string
	ApiVersion                 float32
	NotifyForOperationCreate   bool
	NotifyForOperationUpdate   bool
	NotifyForOperationUndelete bool
	NotifyForOperationDelete   bool
	NotifyForFields            string
}
type ChangeDataCapture struct {
	Name string
}
type PlatformEvent struct {
	Name string
}

type EventData struct {
	Channel string      `json:"channel"`
	Event   EventType   `json:"event"`
	SObject interface{} `json:"sobject"`
}

type EventType struct {
	CreateDate string `json:"createDate"`
	ReplayId   int    `json:"replayId"`
	Type       string `json:"type"`
}

type TopicResponse struct {
	Id      string `json:"id"`
	Success bool   `json:"success"`
}

func (s *Subscriber) CreateTopic(topic string) (PushTopic, error) {
	pushTopic := PushTopic{}

	apiVersion := s.salesforceSharedConfigManager.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	url := fmt.Sprintf("%s/services/data/%s/sobjects/PushTopic", s.salesforceSharedConfigManager.SalesforceToken.InstanceUrl, apiVersion)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(topic)))

	req.Header.Add("Authorization", "Bearer "+s.salesforceSharedConfigManager.SalesforceToken.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return pushTopic, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return pushTopic, err
	}

	apiErrors := &ApiErrors{}
	if err := json.Unmarshal(body, apiErrors); err == nil {
		// Check if api error is valid
		if apiErrors.Validate() {
			if apiErrors.InvalidTokenErorr() {
				// token being refreshed
				if s.retryCount >= REFRESH_NUMBER {
					return pushTopic, errors.New("More than three times token refresh, stop it")
				}
				terr := s.RetryToRefreshToken()
				if terr != nil {
					return pushTopic, terr
				}
				_, err = s.CreateTopic(topic)
				if err != nil {
					return pushTopic, err
				}

			} else if apiErrors.DuplicateTopicError() {
				// existed already, not creating
				// TODO  may still has issue, need expose Topic Name to UI in future and let designtime create the topic firstly
				json.Unmarshal([]byte(topic), &pushTopic)
				logCache.Warnf("Topic with name %s already existing, reuse it", pushTopic.Name)
				return pushTopic, nil

			} else {
				return pushTopic, apiErrors
			}
		}
	}

	json.Unmarshal([]byte(topic), &pushTopic)

	if s.topicId == "" {
		topicResult := &TopicResponse{}
		json.Unmarshal(body, topicResult)
		s.topicId = topicResult.Id
	}

	return pushTopic, nil
}

func (s *Subscriber) ListenToPushTopic(action func(handler trigger.Handler, data interface{})) bool {
	logCache.Debugf("New streaming subscriber...")

	if os.Getenv("FLOGO_LOG_LEVEL") == "DEBUG" {
		os.Setenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG", "true")
		os.Setenv("SF_BAYEUX_DEBUG", "true")
	} else {
		os.Setenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG", "false")
		os.Setenv("SF_BAYEUX_DEBUG", "false")
	}

	//b := bay.Bayeux{}
	s.bayeuxClient = &bay.Bayeux{}
	//bClient := s.bayeuxClient
	token := s.salesforceSharedConfigManager.SalesforceToken
	tb, _ := json.Marshal(token)

	creds := bay.Credentials{APIVersion: s.salesforceSharedConfigManager.APIVersion}
	json.Unmarshal(tb, &creds)

	//logCache.Debug("Access token value is : ", s.token.AccessToken)
	signalChan := make(chan int)
	c, errChan := s.bayeuxClient.SubscribeToChannel(creds, s.subscriberType, s.topic.Name, s.replayID, signalChan) //Involves handshake, subscribe & connect calls

	if s.bayeuxClient.GetClientID() == "" {
		terr := s.RetryToRefreshToken()
		if terr != nil {
			logCache.Debugf("Failed to refresh access token due to %s", terr)
		}
		logCache.Infof("Starting Listen after 2 seconds")
		time.Sleep(2 * time.Second)
		s.ListenToPushTopic(action)
	}
	//logCache.Debugf("Handshake clientID is : ", s.bayeuxClient.GetClientID())
	// start long conn
	signalChan <- 1
	for {
		select {
		case e := <-c:

			//interface to struct conversion
			triggerEvent := bay.TriggerEvent{}
			jsonByte, _ := json.Marshal(e)
			json.Unmarshal(jsonByte, &triggerEvent)

			logCache.Debugf("Salesforce streaming message received: %+v", e)
			if triggerEvent.Channel == "/topic/"+s.topic.Name {
				bs, _ := json.Marshal(e)

				event := EventData{}

				etype := EventType{}
				etype.Type = triggerEvent.Data.Event.Type
				etype.CreateDate = triggerEvent.Data.Event.CreatedDate
				etype.ReplayId = triggerEvent.Data.Event.ReplayID

				event.Channel = triggerEvent.Channel
				event.Event = etype
				json.Unmarshal(triggerEvent.Data.Object, &event.SObject)

				//event struct to interface conversion
				bs, _ = json.Marshal(event)
				json.Unmarshal(bs, &e)

				logCache.Debugf("Salesforce event data on PushTopic [%s]: %s ", s.topic.Name, string(bs))

				// reconn, keep long conn open
				// logCache.Infof("Reconnect after success msg")
				// signalChan <- 1

				go action(s.handler, e)

			} else if triggerEvent.Channel == "/meta/connect" {

				if !triggerEvent.Successful {
					// re-handshake
					logCache.Infof("Received connect failure message: %+v, start rehandshake after 2 seconds", e)

					// halt the stale bayuex client
					signalChan <- -1

					time.Sleep(2 * time.Second)
					s.ListenToPushTopic(action)
				} else {
					// keep long conn
					signalChan <- 1

				}

			} else {
				logCache.Infof("Unknow message received, ignore: %+v ", e)
				signalChan <- 1

			}

		case unknowErr := <-errChan:
			logCache.Infof("Unknow error occurred: %s,  start rehandshake after 2 seconds", unknowErr)

			// halt the stale bayuex client
			signalChan <- -1

			time.Sleep(2 * time.Second)
			s.ListenToPushTopic(action)

		}
	}
}
func (s *Subscriber) ListenToChangeDataCapture(action func(handler trigger.Handler, data interface{})) bool {
	logCache.Debugf("New streaming subscriber...")

	if os.Getenv("FLOGO_LOG_LEVEL") == "DEBUG" {
		os.Setenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG", "true")
		os.Setenv("SF_BAYEUX_DEBUG", "true")
	} else {
		os.Setenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG", "false")
		os.Setenv("SF_BAYEUX_DEBUG", "false")
	}

	//b := bay.Bayeux{}
	s.bayeuxClient = &bay.Bayeux{}
	//bClient := s.bayeuxClient
	token := s.salesforceSharedConfigManager.SalesforceToken
	tb, _ := json.Marshal(token)

	creds := bay.Credentials{APIVersion: s.salesforceSharedConfigManager.APIVersion}
	json.Unmarshal(tb, &creds)

	//logCache.Debug("Access token value is : ", s.token.AccessToken)
	signalChan := make(chan int)
	c, errChan := s.bayeuxClient.SubscribeToChannel(creds, s.subscriberType, s.changeDataCapture.Name, s.replayID, signalChan) //Involves handshake, subscribe & connect calls

	if s.bayeuxClient.GetClientID() == "" {
		terr := s.RetryToRefreshToken()
		if terr != nil {
			logCache.Debugf("Failed to refresh access token due to %s", terr)
		}
		logCache.Infof("Starting Listen after 2 seconds")
		time.Sleep(2 * time.Second)
		s.ListenToChangeDataCapture(action)
	}
	//logCache.Debugf("Handshake clientID is : ", s.bayeuxClient.GetClientID())
	// start long conn
	signalChan <- 1
	for {
		select {
		case e := <-c:
			logCache.Debugf("Salesforce streaming message received: %+v", e)
			if e.(map[string]interface{})["channel"] == "/data/"+s.changeDataCapture.Name {
				bs, _ := json.Marshal(e)

				var event interface{}
				json.Unmarshal(bs, &event)

				logCache.Debugf("Salesforce event data on Change Data Capture [%s]: %s ", s.changeDataCapture.Name, string(bs))

				// reconn, keep long conn open
				// logCache.Infof("Reconnect after success msg")
				// signalChan <- 1

				go action(s.handler, event)

			} else if e.(map[string]interface{})["channel"] == "/meta/connect" {

				if e.(map[string]interface{})["successful"] == false {
					// re-handshake
					logCache.Infof("Received connect failure message: %+v, start rehandshake after 2 seconds", e)

					// halt the stale bayuex client
					signalChan <- -1

					time.Sleep(2 * time.Second)
					s.ListenToChangeDataCapture(action)
				} else {
					// keep long conn
					signalChan <- 1

				}

			} else {
				logCache.Infof("Unknow message received, ignore: %+v ", e)
				signalChan <- 1

			}

		case unknowErr := <-errChan:
			logCache.Infof("Unknow error occurred: %s,  start rehandshake after 2 seconds", unknowErr)

			// halt the stale bayuex client
			signalChan <- -1

			time.Sleep(2 * time.Second)
			s.ListenToChangeDataCapture(action)

		}
	}
}
func (s *Subscriber) ListenToPlatformEvent(action func(handler trigger.Handler, data interface{})) bool {
	logCache.Debugf("New streaming subscriber...")

	if os.Getenv("FLOGO_LOG_LEVEL") == "DEBUG" {
		os.Setenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG", "true")
		os.Setenv("SF_BAYEUX_DEBUG", "true")
	} else {
		os.Setenv("FLOGO_SF_CONNECTOR_BAYEUX_DEBUG", "false")
		os.Setenv("SF_BAYEUX_DEBUG", "false")
	}

	//b := bay.Bayeux{}
	s.bayeuxClient = &bay.Bayeux{}
	//bClient := s.bayeuxClient
	token := s.salesforceSharedConfigManager.SalesforceToken
	tb, _ := json.Marshal(token)

	creds := bay.Credentials{APIVersion: s.salesforceSharedConfigManager.APIVersion}
	json.Unmarshal(tb, &creds)
	//logCache.Debug("Access token value is : ", s.token.AccessToken)

	signalChan := make(chan int)

	c, errChan := s.bayeuxClient.SubscribeToChannel(creds, s.subscriberType, s.platformEvent.Name, s.replayID, signalChan) //Involves handshake, subscribe & connect calls

	if s.bayeuxClient.GetClientID() == "" {
		terr := s.RetryToRefreshToken()
		if terr != nil {
			logCache.Debugf("Failed to refresh access token due to %s", terr)
		}
		logCache.Infof("Starting Listen after 2 seconds")
		time.Sleep(2 * time.Second)
		s.ListenToPlatformEvent(action)
	}
	//logCache.Debugf("Handshake clientID is : ", s.bayeuxClient.GetClientID())
	// start long conn
	signalChan <- 1
	for {
		select {
		case e := <-c:
			logCache.Debugf("Salesforce streaming message received: %+v", e)
			if e.(map[string]interface{})["channel"] == "/event/"+s.platformEvent.Name {
				bs, _ := json.Marshal(e)

				var event interface{}
				json.Unmarshal(bs, &event)

				logCache.Debugf("Salesforce event data on Platform Event [%s]: %s ", s.platformEvent.Name, string(bs))

				// reconn, keep long conn open
				// logCache.Infof("Reconnect after success msg")
				// signalChan <- 1

				go action(s.handler, event)

			} else if e.(map[string]interface{})["channel"] == "/meta/connect" {

				if e.(map[string]interface{})["successful"] == false {
					// re-handshake
					logCache.Infof("Received connect failure message: %+v, start rehandshake after 2 seconds", e)

					// halt the stale bayuex client
					signalChan <- -1

					time.Sleep(2 * time.Second)
					s.ListenToPlatformEvent(action)
				} else {
					// keep long conn
					signalChan <- 1

				}

			} else {
				logCache.Infof("Unknow message received, ignore: %+v ", e)
				signalChan <- 1

			}

		case unknowErr := <-errChan:
			logCache.Infof("Unknow error occurred: %s,  start rehandshake after 2 seconds", unknowErr)

			// halt the stale bayuex client
			signalChan <- -1

			time.Sleep(2 * time.Second)
			s.ListenToPlatformEvent(action)

		}
	}
}
func (s *Subscriber) RetryToRefreshToken() error {

	logCache.Debug("Access token expired, do token refresh.....")
	var err error
	s.retryCount++
	if s.salesforceSharedConfigManager.AuthType == "OAuth 2.0 JWT Bearer Flow" {
		err = s.salesforceSharedConfigManager.DoRefreshTokenUsingJWT(logCache)
	} else {
		err = s.salesforceSharedConfigManager.DoRefreshToken(logCache)
	}
	if err != nil {
		logCache.Errorf("Token refresh error [%s]", err.Error())
		return fmt.Errorf("Token refresh error [%s]", err.Error())
	}
	return err
}
func (s *Subscriber) Stop() error {
	var err error
	logCache.Debug("Subscriber type : ", s.subscriberType)
	logCache.Debug("Auto Create PushTopic : ", s.autoCreatePushTopic)
	if s.subscriberType == "" || (s.subscriberType == "PushTopic" && s.autoCreatePushTopic == true) {
		logCache.Debugf("Deleting PushTopic")
		err = deleteTopicWithId(s.salesforceSharedConfigManager, s.topicId)
	} else if s.subscriberType == "PushTopic" {
		logCache.Debugf("Unsubscribing to PushTopic")
		err = s.bayeuxClient.Unsubscribe("topic", s.topic.Name)
	} else if s.subscriberType == "Change Data Capture" {
		logCache.Debugf("Unsubscribing to Change Data Capture")
		err = s.bayeuxClient.Unsubscribe("data", s.changeDataCapture.Name)
	} else {
		logCache.Debugf("Unsubscribing to Platform Event")
		err = s.bayeuxClient.Unsubscribe("event", s.platformEvent.Name)
	}
	return err
}
func deleteTopicWithId(sscm *sfconnection.SalesforceSharedConfigManager, id string) error {
	logCache.Debugf("Delete topic with id: %s", id)
	apiVersion := sscm.APIVersion
	if apiVersion == "" {
		apiVersion = sfconnection.DEFAULT_APIVERSION
	}

	deleteUrl := sscm.SalesforceToken.InstanceUrl + "/services/data/" + apiVersion + "/sobjects/PushTopic/" + id
	_, err := RestCall(sscm, "DELETE", deleteUrl, nil, logCache)

	return err
}
