package tcm

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/TIBCOSoftware/eftl"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/engine"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/connection"
	"github.com/project-flogo/core/support/log"
	"github.com/tibco/wi-contrib/connection/generic"
)

var logCache = log.ChildLogger(log.RootLogger(), "messaging.connection")
var factory = &TCMFactory{}

type TCMFactory struct {
}

type Settings struct {
	Name              string `md:"name,required"`
	URL               string `md:"url,required"`
	AuthKey           string `md:"authKey,required"`
	ConnectionTimeout int    `md:"timeout"`
	AutoReconAttempt  int    `md:"autoReconnectAttempts"`
	AutoReconMaxDelay int    `md:"autoReconnectMaxDelay"`
	MaxPendingAcks    int32  `md:"maxPendingAcks"`
}

type retry struct {
	attempts          int64
	maxDelay          time.Duration
	autoReconAttempts int64
	err               error
	wg                sync.WaitGroup
}

type Connection interface {
	GetEFTLConnection() *eftl.Connection
	Reconnect(name string) error
	IsTrying() bool
	WaitingConnectionAvailable(name string)
}

type tcmConnection struct {
	conn     *eftl.Connection
	lock     sync.Mutex
	retrying *retry
}

func (*TCMFactory) Type() string {
	return "TCM"
}

func init() {
	err := connection.RegisterManagerFactory(factory)
	if err != nil {
		panic(err)
	}
}

func (*TCMFactory) NewManager(settings map[string]interface{}) (connection.Manager, error) {

	sharedConn := &TCMSharedConfigManager{lock: sync.Mutex{}}
	var err error
	s := &Settings{}

	err = metadata.MapToStruct(settings, s, false)
	if err != nil {
		return nil, err
	}

	if s.ConnectionTimeout == 0 {
		s.ConnectionTimeout = 10
	}
	if s.AutoReconAttempt == 0 {
		s.AutoReconAttempt = 25
	}

	if s.AutoReconMaxDelay == 0 {
		s.AutoReconMaxDelay = 5
	}

	sharedConn.settings = s

	return sharedConn, nil
}

func (k *TCMSharedConfigManager) newConnection(isPublisher bool) (*tcmConnection, error) {

	s := k.settings
	opts := &eftl.Options{
		Password:              s.AuthKey,
		Timeout:               (time.Duration(s.ConnectionTimeout) * time.Second),
		ClientID:              GetClientID(s),
		AutoReconnectAttempts: int64(s.AutoReconAttempt),
		AutoReconnectMaxDelay: (time.Duration(s.AutoReconMaxDelay) * time.Second),
		HandshakeTimeout:      (time.Duration(10) * time.Second),
		OnStateChange: func(c *eftl.Connection, state eftl.State) {
			logCache.Debugf("Connection(ClientId:%s) state is changed to [%s]", c.Options.ClientID, state.String())
		},
	}

	if s.MaxPendingAcks > 0 {
		opts.MaxPendingAcks = s.MaxPendingAcks
	}

	logCache.Debug("Connecting to TIBCO Cloud Messaging Service")

	var message = "| ClientId => " + opts.ClientID + "| Timeout => " + opts.Timeout.String() + " AutoReconnectAttempts => " + strconv.FormatInt(opts.AutoReconnectAttempts, 10) + " | AutoReconnectMaxDelay => " + opts.AutoReconnectMaxDelay.String() + " | HandshakeTimeout => " + opts.HandshakeTimeout.String()

	logCache.Debug(message)

	if isPublisher {
		logCache.Infof("Connecting to TIBCO Cloud Messaging Service with client id [%s] for publisher", opts.ClientID)
	} else {
		logCache.Infof("Connecting to TIBCO Cloud Messaging Service with client id [%s] for subscriber", opts.ClientID)
	}

	// connect to TIBCO Cloud Messaging
	errChan := make(chan error, 1)
	conn, err := eftl.Connect(s.URL, opts, errChan)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to connect to TIBCO Cloud Messaging service due to error - {%s}. Either connection parameters are invalid or messaging service is down.", err.Error()))
	}
	return &tcmConnection{
		conn: conn,
		lock: sync.Mutex{},
	}, nil
}

type TCMSharedConfigManager struct {
	subscribeConnection *tcmConnection
	publisherConnection *tcmConnection
	settings            *Settings
	lock                sync.Mutex
}

func (k *TCMSharedConfigManager) Type() string {
	return "Kafka"
}

func (k *TCMSharedConfigManager) GetConnection() interface{} {
	return nil
}

func (k *TCMSharedConfigManager) GetSubscribeConnection() (*tcmConnection, error) {
	if k.subscribeConnection != nil {
		return k.subscribeConnection, nil
	} else {
		k.lock.Lock()
		defer k.lock.Unlock()
		con, err := k.newConnection(false)
		if err != nil {
			return nil, err
		}
		k.subscribeConnection = con
		go handleConnectionErrors(con.GetEFTLConnection())
		return con, err
	}
}

func (k *TCMSharedConfigManager) GetPublisherConnection() (*tcmConnection, error) {
	k.lock.Lock()
	defer k.lock.Unlock()
	if k.publisherConnection != nil {
		return k.publisherConnection, nil
	} else {
		con, err := k.newConnection(true)
		if err != nil {
			logCache.Errorf("Unable to connect to TIBCO Cloud Messaging service: %s. Attempting reconnection...", err.Error())
			re := &retry{
				attempts:          0,
				maxDelay:          time.Duration(k.settings.AutoReconMaxDelay) * time.Second,
				autoReconAttempts: int64(k.settings.AutoReconAttempt),
				wg:                sync.WaitGroup{},
			}
			con, err = re.retryCreateConnection(k, true)
			if err != nil {
				return nil, err
			}
		}

		k.publisherConnection = con
		go handleConnectionErrors(con.GetEFTLConnection())
		return con, err
	}
}

func handleConnectionErrors(conn *eftl.Connection) {
	for {
		select {
		case err := <-conn.ErrorChan:
			logCache.Errorf("Connection error [%s]. Reconnecting to TIBCO Cloud Messaging service", err.Error())
		}
	}
}

func (t *tcmConnection) GetEFTLConnection() *eftl.Connection {
	return t.conn
}

func (t *tcmConnection) Reconnect(name string) error {
	//There is another one tryng the reconnect
	t.lock.Lock()
	if t.retrying != nil {
		//wait here
		ret := t.retrying
		t.lock.Unlock()
		logCache.Infof("[%s] waiting for other task's connection reconnect finish.....", name)
		//Let all gorountine or activity to wait here
		ret.wg.Wait()
		return ret.err
	}

	logCache.Infof("[%s] Starting connection reconnect", name)
	defer logCache.Infof("[%s] Reconnected done", name)

	re := &retry{
		attempts:          1,
		maxDelay:          t.conn.Options.AutoReconnectMaxDelay,
		autoReconAttempts: t.conn.Options.AutoReconnectAttempts,
		wg:                sync.WaitGroup{},
	}
	t.retrying = re
	re.wg.Add(1)
	//Release to let others wait at begining
	t.lock.Unlock()

	t.conn.Disconnect()
	var err error
	err = re.retry(err, t.conn.Reconnect)
	t.lock.Lock()
	re.wg.Done()
	t.retrying = nil
	t.lock.Unlock()
	return err
}

func (t *tcmConnection) IsTrying() bool {
	return t.retrying != nil
}

func (t *tcmConnection) WaitingConnectionAvailable(name string) {
	if t.retrying != nil {
		logCache.Infof("[%s] Waiting for reconnect done", name)
		t.retrying.wg.Wait()
	}
}

func (k *TCMSharedConfigManager) ReleaseConnection(connection interface{}) {
}

func (r *retry) retry(err error, fn2 func() error) error {
	if r.attempts < r.autoReconAttempts {
		// exponential backoff truncated to max delay
		dur := time.Duration(r.attempts*2) * time.Second
		if dur > r.maxDelay {
			dur = r.maxDelay
		}
		time.Sleep(dur)
		logCache.Infof("TCM Connection retry attempts [%d]", r.attempts)
		r.attempts++
		if e := fn2(); e != nil {
			return r.retry(e, fn2)
		}
	} else {
		return err
	}

	return nil
}

func (r *retry) retryCreateConnection(k *TCMSharedConfigManager, isPub bool) (*tcmConnection, error) {
	if r.attempts < r.autoReconAttempts {
		// exponential backoff truncated to max delay
		dur := time.Duration(r.attempts*2) * time.Second
		if dur > r.maxDelay {
			dur = r.maxDelay
		}
		r.attempts++

		time.Sleep(dur)
		logCache.Infof("Connect to TCM retry attempts [%d]", r.attempts)
		con, e := k.newConnection(isPub)
		if e != nil {
			logCache.Warnf("Connect to TCM server error: %s", e)
			return r.retryCreateConnection(k, isPub)
		} else {
			return con, nil
		}
	} else {
		return nil, fmt.Errorf("not able to establish connect to TCM server after [%d] tries", r.autoReconAttempts)
	}
}

func GetSharedConfiguration(conn interface{}) (connection.Manager, error) {

	var cManager connection.Manager
	var err error
	_, ok := conn.(map[string]interface{})
	if ok {
		cManager, err = handleLegacyConnection(conn)
	} else {
		cManager, err = coerce.ToConnection(conn)
	}

	if err != nil {
		return nil, err
	}

	return cManager, nil
}

func handleLegacyConnection(conn interface{}) (connection.Manager, error) {

	connectionObject, _ := coerce.ToObject(conn)
	if connectionObject == nil {
		return nil, errors.New("Connection object is nil")
	}

	id := connectionObject["id"].(string)

	cManager := connection.GetManager(id)
	if cManager == nil {

		connObject, err := generic.NewConnection(connectionObject)
		if err != nil {
			return nil, err
		}

		cManager, err = factory.NewManager(connObject.Settings())
		if err != nil {
			return nil, err
		}
		// Ignore error for concurrent issue for ylopo
		connection.RegisterManager(id, cManager)
	}
	return cManager, nil

}

func (k *TCMSharedConfigManager) Start() error {
	return nil
}

func (k *TCMSharedConfigManager) Stop() error {
	logCache.Info("Cleaning up Connection")
	if k.subscribeConnection != nil {
		k.subscribeConnection.GetEFTLConnection().Disconnect()
	}

	if k.publisherConnection != nil {
		k.publisherConnection.GetEFTLConnection().Disconnect()
	}

	return nil
}

func GetClientID(setting *Settings) string {
	uuid := string(time.Now().UnixNano())
	uuidGenerator, err := support.NewGenerator()
	if err == nil {
		uuid = uuidGenerator.NextAsString()
	}
	return getTCIAppID() + engine.GetAppName() + "-" + setting.Name + "-" + uuid + "-auto"
}

func getTCIAppID() string {
	appID := os.Getenv("TIBCO_INTERNAL_SERVICE_NAME")
	if len(appID) > 0 {
		return appID + "-"
	}
	return ""
}
