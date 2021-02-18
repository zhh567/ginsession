package ginsession

import (
	"errors"
)

// Session is the interface of session.
// It's stored in memory or redis.
type Session interface {
	ID() string
	Load() error
	Get(string) (interface{}, error)
	Set(string, interface{})
	Del(string)
	Save() error
	SetExpired(int)
	IsModify() bool
	IsRedis() bool
}

// SessionMgr used to manage session.
type SessionMgr interface {
	Init(addr string, options ...string) error
	GetSession(string) (Session, error)
	CreateSession() Session
	Clear(string)
}

// CreateSessionMgr return session manager
func CreateSessionMgr(name string, addr string, options ...string) (SessionMgr, error) {
	var sm SessionMgr
	if name == "memory" {
		sm = NewMemSessionMgr()
	} else if name == "redis" {
		sm = NewRedisSessionMgr()
	} else {
		return nil, errors.New("only support 'memory' and 'redis'")
	}
	err := sm.Init(addr, options...)
	if err != nil {
		return nil, err
	}
	return sm, nil
}
