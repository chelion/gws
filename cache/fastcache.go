package cache
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"sync"
	"net/rpc"
	"github.com/chelion/gws/utils"
)
const(
	CLIENTMINNUM = 10
	CLIENTMAXNUM = 30
	CACHEMAXSIZE = 64 * 1024 - 16 - 4 - 1
)

type Args struct{
	Key []byte
	Data []byte
	Expire int
}

var(
	isAliveData = Args{Key:nil,Data:nil,Expire:-1}
)

type FastCache struct{
	clientPool *utils.ConnectPool
	config *CacheConfig
	serverAddr string
	netWork string
	lock *sync.Mutex
	timeoutSec int64
	isStopped bool
}

type FastCacheClient struct{
	client *rpc.Client
	isAlive bool
	timeout int64
	lastActiveTime int64
	args Args
}

func (fcc *FastCacheClient)Open()(err error){
	return nil
}

func (fcc *FastCacheClient)Close()(err error){
	err = nil
	if nil != fcc.client{
		err = fcc.client.Close()
		if nil == err{
			fcc.client = nil
		}
	}
	return 
}

func (fcc *FastCacheClient)IsAlive()(sta bool){
	if nil != fcc.client{
		if utils.GetNowUnixSec() - fcc.lastActiveTime > fcc.timeout {
			sta := false
			err := (fcc.client).Call("FastCache.Ping",isAliveData, &sta)
			if nil == err{
				fcc.lastActiveTime = utils.GetNowUnixSec()
				fcc.isAlive = true
				return true
			}else{
				fcc.isAlive = false
			}
		}
		return fcc.isAlive
	}
	return false
}

func NewFastCache(config *CacheConfig)(fcc *FastCache,err error){
	fcc = &FastCache{clientPool:nil,config:config,isStopped:true,serverAddr:config.ServerAddr,netWork:config.Network,lock:new(sync.Mutex),timeoutSec:config.TimeoutSec}
	return fcc,nil
}

func (fcc *FastCache)Start()(err error){
	fcc.lock.Lock()
	defer fcc.lock.Unlock()
	if true == fcc.isStopped{
		clientPool,err := utils.NewConnectPool(fcc.config.ConnectMinNum,fcc.config.ConnectMaxNum,func ()(item utils.ConnectPoolItem,err error){
			fc := &FastCacheClient{client:nil}
			fc.client, err = rpc.Dial(fcc.netWork,fcc.serverAddr)
			if nil != err{
				return nil,CACHESERVER_ERR
			}
			fc.isAlive = true
			fc.args = Args{Key:make([]byte,0),Data:make([]byte,0),Expire:-1}
			fc.timeout = fcc.timeoutSec
			fc.lastActiveTime = utils.GetNowUnixSec()
			return utils.ConnectPoolItem(fc),nil
		})
		if nil != err{
			return err
		}
		fcc.clientPool = clientPool
		if nil == fcc.clientPool.Start(fcc.timeoutSec){
			fcc.isStopped = false
		}
		return err
	}
	return nil
}

func (fcc *FastCache)Stop()(err error){
	fcc.lock.Lock()
	defer fcc.lock.Unlock()
	if false == fcc.isStopped{
		err = fcc.clientPool.Stop()
		if utils.CONNECTPOOL_STOP_SUC == err{
			fcc.isStopped = true
		}
		return err
	}
	return nil
}

func (fcc *FastCache)Get(key string)(value []byte,err error){
	if 0 == len(key) || len(key) >CACHEMAXSIZE{
		return nil,CACHEPARAM_ERR
	}
	if nil != fcc.clientPool{
		item,err := fcc.clientPool.Get()
		if nil == err{
			fc,ok := item.(*FastCacheClient)
			if ok{
				defer fcc.clientPool.Put(item)
				fc.args.Key = []byte(key)
				fc.args.Data = nil
				fc.args.Expire = -1
				err =(fc.client).Call("FastCache.Get",fc.args, &value)
				if nil == err{
					fc.lastActiveTime = utils.GetNowUnixSec()
					return value,nil
				}else{
					fc.isAlive = false
					return nil,CACHECLIENT_ERR
				}
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fcc.Stop(){
					return nil,CACHESTOP_SUC
				}
			}
			return nil,CACHESERVER_ERR
		}
	}
	return nil,CACHECLIENT_NIL
}


func (fcc *FastCache)Set(key string,value []byte,expire int)(err error){
	var sta bool = false
	if nil == value || 0 == len(key) || len(key) >CACHEMAXSIZE || len(value) > CACHEMAXSIZE{
		return CACHEPARAM_ERR
	}
	if nil != fcc.clientPool{
		item,err := fcc.clientPool.Get()
		if nil == err{
			fc,ok := item.(*FastCacheClient)
			if ok{
				defer fcc.clientPool.Put(item)
				fc.args.Key = []byte(key)
				fc.args.Data = value
				fc.args.Expire = expire
				err = (fc.client).Call("FastCache.Set",fc.args, &sta)
				if nil != err{
					fc.isAlive = false
					return CACHECLIENT_ERR
				}
				fc.lastActiveTime = utils.GetNowUnixSec()
				return nil
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fcc.Stop(){
					return CACHESTOP_SUC
				}
			}
			return CACHESERVER_ERR
		}
	}
	return CACHECLIENT_NIL
}

func (fcc *FastCache)Delete(key string)(sta bool,err error){
	if 0 == len(key) || len(key) >CACHEMAXSIZE{
		return false,CACHEPARAM_ERR
	}
	if nil != fcc.clientPool{
		item,err := fcc.clientPool.Get()
		if nil == err{
			fc,ok := item.(*FastCacheClient)
			if ok{
				defer fcc.clientPool.Put(item)
				fc.args.Key = []byte(key)
				fc.args.Data = nil
				fc.args.Expire = -1
				err = (fc.client).Call("FastCache.Delete", fc.args, &sta)
				if nil == err{
					fc.lastActiveTime = utils.GetNowUnixSec()
					return sta,nil
				}else{
					fc.isAlive = false
					return false,CACHECLIENT_ERR
				}
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fcc.Stop(){
					return false,CACHESTOP_SUC
				}
			}
			return false,CACHESERVER_ERR
		}
	}
	return false,CACHECLIENT_NIL
}

func (fcc *FastCache)Exists(key string)(sta bool,err error){
	if 0 == len(key) || len(key) >CACHEMAXSIZE{
		return false,CACHEPARAM_ERR
	}
	if nil != fcc.clientPool{
		item,err := fcc.clientPool.Get()
		if nil == err{
			fc,ok := item.(*FastCacheClient)
			if ok{
				defer fcc.clientPool.Put(item)
				fc.args.Key = []byte(key)
				fc.args.Data = nil
				fc.args.Expire = -1
				err = (fc.client).Call("FastCache.Exists", fc.args, &sta)
				if nil == err{
					fc.lastActiveTime = utils.GetNowUnixSec()
					return sta,nil
				}else{
					fc.isAlive = false
					return false,CACHECLIENT_ERR
				}
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fcc.Stop(){
					return false,CACHESTOP_SUC
				}
			}
			return false,CACHESERVER_ERR
		}
	}
	return false,CACHECLIENT_NIL
}