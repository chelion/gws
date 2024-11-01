package utils

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	CONNECTPOOL_PARAM_ERR   = errors.New("connect pool param is error")
	CONNECTPOOL_STOPPED_ERR = errors.New("connect pool already stopped")
	CONNECTPOOL_STOP_SUC    = errors.New("connect pool is stop suc")
	CONNECTPOOL_NEEDSTOP    = errors.New("connect pool need stop")
)

type ConnectPoolItem interface {
	Open() (err error)
	Close() (err error)
	IsAlive() (sta bool)
}

type Factory func() (connectPoolItem ConnectPoolItem, err error)

type ConnectPool struct {
	maxNum           int64
	idleNum          int64
	minNum           int64
	curNum           int64
	factory          Factory
	lock             *sync.RWMutex
	isStoped         bool
	failFlag         bool
	connectPoolItems chan ConnectPoolItem
}

//New a connect pool

func NewConnectPool(minNum int64, maxNum int64, factory Factory) (connectPool *ConnectPool, err error) {
	if minNum <= 0 || maxNum < minNum || nil == factory {
		return nil, CONNECTPOOL_PARAM_ERR
	}
	connectPool = &ConnectPool{minNum: minNum, failFlag: false, maxNum: maxNum, curNum: 0, idleNum: 0, factory: factory,
		isStoped: true, lock: new(sync.RWMutex)}
	return connectPool, nil
}

func (connectPool *ConnectPool) Start(timeoutSec int64) error { //超时后就返回错误
	var i int64
	var cnt int64 = 0
	connectPool.lock.Lock()
	if true == connectPool.isStoped {
		connectPool.connectPoolItems = make(chan ConnectPoolItem, connectPool.maxNum)
		connectPool.isStoped = false
	} else {
		connectPool.lock.Unlock()
		return nil
	}
	cnt = 0
	for i = 0; i < connectPool.minNum; i++ {
		for {
			item, e := connectPool.factory()
			if e != nil {
				if cnt == timeoutSec {
					for {
						select {
						case connectPoolItem, ok := <-connectPool.connectPoolItems:
							{
								if ok {
									connectPoolItem.Close()
								}
							}
						default:
							{
								close(connectPool.connectPoolItems)
								connectPool.lock.Unlock()
								return e
							}
						}
					}
				} else {
					time.Sleep(1 * time.Second)
					cnt++
					continue
				}
			} else {
				if nil == item.Open() {
					connectPool.connectPoolItems <- item
					break
				}
			}
		}
	}
	connectPool.failFlag = false
	atomic.StoreInt64(&connectPool.idleNum, connectPool.minNum)
	atomic.StoreInt64(&connectPool.curNum, connectPool.minNum)
	connectPool.lock.Unlock()
	return nil
}

func (connectPool *ConnectPool) Stop() (err error) {
	connectPool.lock.Lock()
	if connectPool.isStoped {
		connectPool.lock.Unlock()
		return CONNECTPOOL_STOPPED_ERR
	}
	for {
		select {
		case connectPoolItem, ok := <-connectPool.connectPoolItems:
			{
				if ok {
					if nil != connectPoolItem {
						connectPoolItem.Close()
					}
				}
			}
		default:
			{
				close(connectPool.connectPoolItems)
				connectPool.isStoped = true
				connectPool.lock.Unlock()
				return CONNECTPOOL_STOP_SUC
			}
		}
	}
}

func (connectPool *ConnectPool) newConnectPoolItem() {
	if atomic.LoadInt64(&connectPool.curNum) < connectPool.maxNum {
		atomic.AddInt64(&connectPool.curNum, 1)
		item, e := connectPool.factory()
		if e == nil {
			if nil == item.Open() {
				connectPool.lock.Lock()
				if connectPool.isStoped {
					item.Close()
				} else {
					atomic.AddInt64(&connectPool.idleNum, 1)
					connectPool.connectPoolItems <- item
				}
				connectPool.lock.Unlock()
			}
		} else {
			connectPool.lock.Lock()
			atomic.AddInt64(&connectPool.curNum, -1)
			if false == connectPool.failFlag {
				connectPool.failFlag = true
				connectPool.connectPoolItems <- nil
			}
			connectPool.lock.Unlock()
		}
	}
}

func (connectPool *ConnectPool) GetCurrentIdleNum() int64 {
	var idleNum int64
	connectPool.lock.Lock()
	if connectPool.isStoped {
		connectPool.lock.Unlock()
		return 0
	}
	idleNum = atomic.LoadInt64(&connectPool.idleNum)
	connectPool.lock.Unlock()
	return idleNum
}

func (connectPool *ConnectPool) Get() (ConnectPoolItem, error) {
	connectPoolItem, ok := <-connectPool.connectPoolItems
	if !ok {
		return nil, CONNECTPOOL_STOPPED_ERR
	}
	if nil == connectPoolItem {
		return nil, CONNECTPOOL_NEEDSTOP
	}
	atomic.AddInt64(&connectPool.idleNum, -1)
	if true == connectPoolItem.IsAlive() {
		if 0 >= atomic.LoadInt64(&connectPool.idleNum) { //idel is 0,need create
			if atomic.LoadInt64(&connectPool.curNum) < connectPool.maxNum {
				connectPool.lock.RLock()
				if connectPool.isStoped {
					connectPool.lock.RUnlock()
					return nil, CONNECTPOOL_STOPPED_ERR
				}
				connectPool.lock.RUnlock()
				go connectPool.newConnectPoolItem()
			}
		}
		return connectPoolItem, nil
	} else {
		connectPoolItem.Close()
		atomic.AddInt64(&connectPool.curNum, -1)
		if atomic.LoadInt64(&connectPool.curNum) < connectPool.maxNum {
			go connectPool.newConnectPoolItem()
		}
	}
	return nil, CONNECTPOOL_STOPPED_ERR
}

func (connectPool *ConnectPool) Put(connectPoolItem ConnectPoolItem) (err error) {
	connectPool.lock.Lock()
	if connectPool.isStoped {
		connectPoolItem.Close()
		connectPool.lock.Unlock()
		return CONNECTPOOL_STOPPED_ERR
	}
	atomic.AddInt64(&connectPool.idleNum, 1)
	connectPool.connectPoolItems <- connectPoolItem
	connectPool.lock.Unlock()
	return nil
}
