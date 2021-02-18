package ginsession

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

// memSession is the memory version of session
type memSession struct {
	id      string
	data    map[string]interface{}
	expired int
	rwLock  sync.RWMutex
}

func NewMemSession(id string) *memSession {
	return &memSession{
		id:   id,
		data: make(map[string]interface{}, 8),
	}
}

func (m *memSession) ID() string {
	return m.id
}

func (m *memSession) Set(key string, value interface{}) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	m.data[key] = value
}

func (m *memSession) Get(key string) (interface{}, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	value, ok := m.data[key]
	if !ok {
		err := fmt.Errorf("invalid Key")
		return value, err
	}
	return value, nil
}

func (m *memSession) Del(key string) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	delete(m.data, key)
}

func (m *memSession) SetExpired(expired int) {
	m.expired = expired
}

func (m *memSession) IsRedis() bool {
	return false
}

// Is not used now.
func (m *memSession) Save() error    { return nil }
func (m *memSession) Load() error    { return nil }
func (m *memSession) IsModify() bool { return false }

// MemSessionMgr is the manager of memSession
type MemSessionMgr struct {
	session map[string]Session
	rwLock  sync.RWMutex
}

func NewMemSessionMgr() *MemSessionMgr {
	return &MemSessionMgr{
		session: make(map[string]Session, 1024),
	}
}

// Init is not used now
func (m *MemSessionMgr) Init(addr string, options ...string) error { return nil }

func (m *MemSessionMgr) GetSession(sessionID string) (Session, error) {
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	sd, ok := m.session[sessionID]
	if !ok {
		err := fmt.Errorf("invalid session id : %s", sessionID)
		return sd, err
	}
	return sd, nil
}

func (m *MemSessionMgr) CreateSession() Session {
	sessionID := uuid.NewString()
	sd := NewMemSession(sessionID)
	m.session[sd.ID()] = sd
	return sd
}

func (m *MemSessionMgr) Clear(sessionID string) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	delete(m.session, sessionID)
}
