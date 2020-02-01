package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
import(
	//"fmt"
	"net"
	"sync"
	"bytes"
	//"errors"
	"github.com/chelion/gws/utils"
)

type RedisCache struct{
	clientPool *utils.ConnectPool
	config *CacheConfig
	serverAddr string
	netWork string
	lock *sync.RWMutex
	timeoutSec int64
	isStopped bool
}

type RedisCacheClient struct{
	conn net.Conn
	readBuff []byte
	bytesBuff *bytes.Buffer
	isAlive bool
	timeout int64
	lastActiveTime int64
}


func (rcc *RedisCacheClient)Open()(err error){
	return nil
}

func (rcc *RedisCacheClient)Close()(err error){
	err = nil
	if nil != mcc.conn{
		err = mcc.conn.Close()
		fmt.Println("close")
		if nil == err{
			mcc.conn = nil
		}
	}
	return 
}

func (rcc *RedisCacheClient)IsAlive()(sta bool){
	if utils.GetNowUnixSec() - mcc.lastActiveTime > mcc.timeout {
		_,err := mcc.conn.Write(versionCmd)
		if nil != err{
			mcc.isAlive = false
		}
		mcc.readBuff=mcc.readBuff[0:]
		_, err = mcc.conn.Read(mcc.readBuff)
		if nil != err{
			mcc.isAlive = false
		}
		if nil == err{
			mcc.lastActiveTime = utils.GetNowUnixSec()
			mcc.isAlive = true
			return true
		}
	}
	return mcc.isAlive
}


func NewRedisCache(config *CacheConfig)(redisc *RedisCache,err error){
	redisc = &MemCache{clientPool:nil,config:config,isStopped:true,serverAddr:config.ServerAddr,netWork:config.Network,lock:new(sync.RWMutex),timeoutSec:config.TimeoutSec}
	return redisc,nil
}

func (redisc *RedisCache)Start()(err error){
	fmt.Println("start---------------------------")
	redisc.lock.Lock()
	defer redisc.lock.Unlock()
	if true == redisc.isStopped{
		fmt.Println("start clientPool")
		clientPool,err := utils.NewConnectPool(redisc.config.ConnectMinNum,redisc.config.ConnectMaxNum,func ()(item utils.ConnectPoolItem,err error){
			mcc := &MemCacheClient{conn:nil}
			mcc.conn, err = net.Dial(redisc.netWork,redisc.serverAddr)
			if nil != err{
				return nil,CACHESERVER_ERR
			}
			mcc.isAlive = true
			mcc.timeout = redisc.timeoutSec
			mcc.readBuff = make([]byte,512)
			mcc.bytesBuff = new(bytes.Buffer)
			mcc.lastActiveTime = utils.GetNowUnixSec()
			return utils.ConnectPoolItem(mcc),nil
		})
		if nil != err{
			
			return err
		}
		redisc.clientPool = clientPool
		if nil == redisc.clientPool.Start(redisc.timeoutSec){
			redisc.isStopped = false
			fmt.Println("redisc.isActive")
		}
		fmt.Println("start------------end---------------",err)
		return err
	}
	return nil
}

func (redisc *RedisCache)Stop()(err error){
	redisc.lock.Lock()
	defer redisc.lock.Unlock()
	if false == redisc.isStopped{
		err = redisc.clientPool.Stop()
		if utils.CONNECTPOOL_STOP_SUC == err{
			redisc.isStopped = true
		}
		return err
	}
	return nil
}

func (redisc *RedisCache)Get(key string)(value []byte,err error){
	if false == checkKey(key){
		return nil,CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			mcc,ok := item.(*MemCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				getItem,err := getCmd(mcc,key)
				if nil == err{
					mcc.lastActiveTime = utils.GetNowUnixSec()
					if nil != getItem{
						return getItem.Value,nil
					}
					return nil,nil
				}else{
					if err == CACHECLIENT_ERR{
						mcc.isAlive = false
					}
				}
				return nil,err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == redisc.Stop(){
					return nil,CACHESTOP_SUC
				}
			}
			return nil,CACHESERVER_ERR
		}
	}
	redisc.lock.RUnlock()
	return nil,CACHECLIENT_NIL
}

func (redisc *RedisCache)Set(key string,value []byte,expire int)(err error){
	if nil == value || false == checkKey(key){
		return CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			mcc,ok := item.(*MemCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				newItem := &Item{Key:key,Value:value,Expiration:int32(expire)}
				err = setCmd(mcc,newItem)
				if err == CACHECLIENT_ERR{
					mcc.isAlive = false
				}else{
					mcc.lastActiveTime = utils.GetNowUnixSec()
				}
				return err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == redisc.Stop(){
					return CACHESTOP_SUC
				}
			}
			return CACHESERVER_ERR
		}
	}
	redisc.lock.RUnlock()
	return CACHECLIENT_NIL
}

func (redisc *RedisCache)Delete(key string)(sta bool,err error){
	if false == checkKey(key){
		return false,CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			mcc,ok := item.(*MemCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				err = deleteCmd(mcc,key)
				if nil == err{
					return true,nil
				}
				if err == CACHECLIENT_ERR{
					mcc.isAlive = false
				}else{
					mcc.lastActiveTime = utils.GetNowUnixSec()
				}
				return false,err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == redisc.Stop(){
					return false,CACHESTOP_SUC
				}
			}
			return false,CACHESERVER_ERR
		}
	}
	return false,CACHECLIENT_NIL
}
*/