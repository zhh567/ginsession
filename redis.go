package ginsession

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
)

// redisSession is redis version of Session.
type redisSession struct {
	id         string
	data       map[string]interface{}
	modifyFlag bool
	expired    int
	rwLock     sync.RWMutex
	client     *redis.Client
}

func NewRedisSession(id string, client *redis.Client) *redisSession {
	return &redisSession{
		id:         id,
		data:       make(map[string]interface{}, 8),
		modifyFlag: true,
		client:     client,
	}
}

func (r *redisSession) ID() string {
	return r.id
}

// Load from redis.
func (r *redisSession) Load() error {
	data, err := r.client.Get(r.id).Bytes()
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	err = decoder.Decode(&r.data)
	if err != nil {
		return err
	}
	return err
}

func (r *redisSession) Get(key string) (interface{}, error) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	value, ok := r.data[key]
	if !ok {
		return nil, errors.New("invalid Key")
	}
	return value, nil
}

func (r *redisSession) Set(key string, value interface{}) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	r.data[key] = value
	r.modifyFlag = true
}

func (r *redisSession) Del(key string) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	delete(r.data, key)
	r.modifyFlag = true
}

// Save to redis.
func (r *redisSession) Save() error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	if !r.modifyFlag {
		return nil
	}
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(r.data)
	if err != nil {
		return err
	}
	r.client.Set(r.id, buf.Bytes(), time.Second*time.Duration(r.expired))
	r.modifyFlag = false
	return err
}

func (r *redisSession) SetExpired(expired int) {
	r.expired = expired
}

// IsModify is used to determine whether data in memory has charged.
func (r *redisSession) IsModify() bool {
	return r.modifyFlag
}

func (r *redisSession) IsRedis() bool {
	return true
}

// redisSessionMgr is the manager of redisSession
type redisSessionMgr struct {
	session map[string]Session
	rwLock  sync.RWMutex
	client  *redis.Client
}

// Init the connection to Redis.
func (r *redisSessionMgr) Init(addr string, options ...string) error {
	password := ""
	db := 0
	var err error

	if len(options) == 1 {
		password = options[0]
	}
	if len(options) == 2 {
		password = options[0]
		db, err = strconv.Atoi(options[1])
		if err != nil {
			return errors.New("invalid redis DB param")
		}
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	_, err = r.client.Ping().Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *redisSessionMgr) GetSession(sessionID string) (Session, error) {
	session := NewRedisSession(sessionID, r.client)
	err := session.Load()
	if err != nil {
		return nil, err
	}
	r.rwLock.RLock()
	r.session[sessionID] = session
	r.rwLock.RUnlock()
	return session, nil
}

func (r *redisSessionMgr) CreateSession() Session {
	sessionID := uuid.NewString()
	sd := NewRedisSession(sessionID, r.client)
	r.session[sd.ID()] = sd
	return sd
}

func (r *redisSessionMgr) Clear(sessionID string) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	if r.session[sessionID].IsModify() {
		_ = r.session[sessionID].Save()
	}
	delete(r.session, sessionID)
}

func NewRedisSessionMgr() *redisSessionMgr {
	return &redisSessionMgr{
		session: make(map[string]Session, 1024),
	}
}
