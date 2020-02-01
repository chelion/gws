package memory
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	"time"
	"sync"
	"errors"
	"reflect"
	"github.com/chelion/gws/utils"
	"github.com/chelion/gws/session"
	"github.com/cespare/xxhash"
)

// session memory provider

const ProviderName = "memory"

type Provider struct {
	mcc			*MCacheCluster
	storeLock	*sync.RWMutex
	storeMap 	map[uint64]*Store
	config      *Config
	maxLifeTime int64
}

// new memory provider
func NewProvider() *Provider {

	return &Provider{
		storeLock: new(sync.RWMutex),
		storeMap: make(map[uint64]*Store),
		mcc:	nil,
		maxLifeTime: 0,
	}
}

// init provider config
func (mp *Provider) Init(lifeTime int64, config session.ProviderConfig) error {
	if config.Name() != ProviderName {
		return errors.New("session memory provider init error, config must memory config")
	}
	vc := reflect.ValueOf(config)
	mc := vc.Interface().(*Config)
	mp.config = mc
	mp.mcc = NewCCMap(mp.config.CacheConfig)
	mp.maxLifeTime = lifeTime
	return nil
}

// need gc
func (mp *Provider) NeedGC() bool {
	return true
}

// session garbage collection
func (mp *Provider) GC() {
	mp.storeLock.RLock()
	storeMap := mp.storeMap
	mp.storeLock.RUnlock()
	for sessionId, value := range storeMap{
		if time.Now().Unix() >= value.lastActiveTime+mp.maxLifeTime {
			// destroy session sessionId
			if store,ok := storeMap[sessionId];ok{
				store.Flush()
				mp.storeLock.Lock()
				delete(mp.storeMap,sessionId)
				mp.storeLock.Unlock()
			}
		}
	}
}

// read session store by session id
func (mp *Provider) ReadStore(sessionId string) (session.SessionStore, error) {
	sssum64 := xxhash.Sum64(utils.String2Bytes(sessionId))
	mp.storeLock.RLock()
	if memStore,ok := mp.storeMap[sssum64];ok{
		mp.storeLock.RUnlock()
		return memStore,nil
	}
	mp.storeLock.RUnlock()
	newMemStore := NewMemoryStore(sessionId,mp.mcc)
	mp.storeLock.Lock()
	mp.storeMap[sssum64] = newMemStore
	mp.storeLock.Unlock()
	return newMemStore, nil
}

// regenerate session
func (mp *Provider) Regenerate(oldSessionId string, sessionId string) (session.SessionStore, error) {
	var nsssum64 uint64
	mp.storeLock.RLock()
	osssum64 := xxhash.Sum64(utils.String2Bytes(oldSessionId))
	memStoreInter := mp.storeMap[osssum64]
	mp.storeLock.RUnlock()
	if memStoreInter != nil {
		newMemStore := NewMemoryStore(sessionId,mp.mcc)
		newMemStore.Set(sessionId, memStoreInter.Get(oldSessionId))
		// delete old session store
		nsssum64 = xxhash.Sum64(utils.String2Bytes(sessionId))
		mp.storeLock.Lock()
		delete(mp.storeMap,osssum64)
		mp.storeMap[nsssum64] = newMemStore
		mp.storeLock.Unlock()
		return newMemStore, nil
	}

	memStore := NewMemoryStore(sessionId,mp.mcc)
	mp.storeLock.Lock()
	mp.storeMap[nsssum64] = memStore
	mp.storeLock.Unlock()
	return memStore, nil
}

// destroy session by sessionId
func (mp *Provider) Destroy(sessionId string) error {
	sssum64 := xxhash.Sum64(utils.String2Bytes(sessionId))
	mp.storeLock.RLock()
	if store,ok := mp.storeMap[sssum64];ok{
		mp.storeLock.RUnlock()
		store.Flush()
		mp.storeLock.Lock()
		delete(mp.storeMap,sssum64)
		mp.storeLock.Unlock()
		return nil
	}
	mp.storeLock.RUnlock()
	return nil
}

// session values count
func (mp *Provider) Count() int {
	return len(mp.storeMap)
}

// register session provider
func init() {
	session.Register(ProviderName, NewProvider())
}
