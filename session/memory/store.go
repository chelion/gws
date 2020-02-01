package memory
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	"time"
)

// session memory store

// new default memory store

type Store struct {
	keys	[]string
	sessionId string
	data      *MCacheCluster
	lastActiveTime int64
}

func NewMemoryStore(sessionId string,mcc *MCacheCluster) *Store {
	return  &Store{
		keys : []string{},
		sessionId : sessionId,
		data :mcc,
	}
}


// get data by key
func (s *Store) Get(key string) []byte {
	s.lastActiveTime = time.Now().Unix()
	return s.data.Get(s.sessionId+key)
}

// set data
func (s *Store) Set(key string, value []byte) {
	pkey := s.sessionId+key
	vlen := len(s.keys)
	if 0 == vlen{
		s.keys = append(s.keys,key)
	}else{
		i := 0
		for i=0;i<vlen;i++{
			if s.keys[i] == key{
				break
			}
		}
		if i != vlen{
			s.keys = append(s.keys,key)
		}
	}
	s.data.Set(pkey, value)
}

// delete data by key
func (s *Store) Delete(key string) {
	len := len(s.keys)
	if 0 == len{
		return
	}
	pkey := s.sessionId+key
	i := 0
	k := 0
	tmp := make([]string,len)
	for i=0;i<len;i++{
		if s.keys[i] == key{
			s.data.Delete(pkey)
		}else{
			tmp[k] = s.keys[i]
			k++
		}
	}
	s.keys = tmp
}

// get all data
func (s *Store) GetAll() map[string][]byte {
	len := len(s.keys)
	if 0 == len{
		s.lastActiveTime = time.Now().Unix()
		return nil
	}else{
		data := make(map[string][]byte)
		for i:=0;i<len;i++{
			data[s.keys[i]] = s.data.Get(s.sessionId+s.keys[i])
		}
		s.lastActiveTime = time.Now().Unix()
		return data
	}
}


// flush all data
func (s *Store) Flush() {
	len := len(s.keys)
	if 0 == len{
		return
	}else{
		for i:=0;i<len;i++{
			s.data.Delete(s.sessionId+s.keys[i])
		}
	}
	s.keys = []string{}
	return
}

// get session id
func (s *Store) GetSessionId() string {
	return s.sessionId
}


// save store
func (s *Store) Save() error {
	s.lastActiveTime = time.Now().Unix()
	return nil
}
