package requestreply

import (
	"sync"

	"github.com/project-flogo/core/engine"
	"github.com/tibco/msg-ems-client-go/tibems"
)

type SessionCache struct {
	sessionMap map[string]*tibems.Session
	slock      *sync.RWMutex
}

func init() {
	sessionCache = &SessionCache{}
	sessionCache.slock = &sync.RWMutex{}
	sessionCache.sessionMap = make(map[string]*tibems.Session)
	engine.LifeCycle(sessionCache)
}

var sessionCache *SessionCache

func (sessionCache *SessionCache) GetSession(connection *tibems.Connection) (session *tibems.Session, err error) {
	sessionCache.slock.RLock()
	clientId, err := connection.GetClientID()
	if err != nil {
		sessionCache.slock.RUnlock()
		return nil, err
	}
	session, ok := sessionCache.sessionMap[clientId]
	if ok && session != nil {
		sessionCache.slock.RUnlock()
		return session, nil
	}
	sessionCache.slock.RUnlock()
	sessionCache.slock.Lock()
	defer sessionCache.slock.Unlock()
	session, err = connection.CreateSession(false, tibems.AckModeAutoAcknowledge)
	sessionCache.sessionMap[clientId] = session
	return session, err
}

func (sessionCache *SessionCache) Start() error {
	return nil
}

func (sessionCache *SessionCache) Stop() error {
	for _, session := range sessionCache.sessionMap {
		session.Close()
	}
	return nil
}
