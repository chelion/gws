package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"net"
	"sync"
	"bytes"
	"errors"
	"strconv"
	"github.com/chelion/gws/utils"
)

var (
	ErrMemCacheMiss = errors.New("memcache: cache miss")
	ErrMemNotStored = errors.New("memcache: item not stored")
	ErrMemCorrupt = errors.New("memcache: corrupt get result read")
	ErrMemUnexpectedError = errors.New("memcache: unexpected line in get response")
)

type MemItem struct {
	Key string
	Value []byte
	Flags uint32
	Expiration uint32
	casid uint64
}

type MemCache struct{
	clientPool *utils.ConnectPool
	config *CacheConfig
	serverAddr string
	netWork string
	lock *sync.RWMutex
	timeoutSec int64
	isStopped bool
}

type MemCacheClient struct{
	conn net.Conn
	cacheBuff []byte
	bytesBuff *bytes.Buffer
	isAlive bool
	timeout int64
	lastActiveTime int64
}

func (mcc *MemCacheClient)Open()(err error){
	return nil
}

func (mcc *MemCacheClient)Close()(err error){
	err = nil
	if nil != mcc.conn{
		err = mcc.conn.Close()
		if nil == err{
			mcc.conn = nil
		}
	}
	return 
}

func (mcc *MemCacheClient)IsAlive()(sta bool){
	if utils.GetNowUnixSec() - mcc.lastActiveTime > mcc.timeout {
		_,err := mcc.conn.Write([]byte("version \r\n"))
		if nil != err{
			mcc.isAlive = false
		}
		mcc.cacheBuff=mcc.cacheBuff[0:]
		_, err = mcc.conn.Read(mcc.cacheBuff)
		if nil != err{
			mcc.isAlive = false
		}else{
			mcc.lastActiveTime = utils.GetNowUnixSec()
			mcc.isAlive = true
			return true
		}
	}
	return mcc.isAlive
}


func NewMemCache(config *CacheConfig)(memc *MemCache,err error){
	memc = &MemCache{clientPool:nil,config:config,isStopped:true,serverAddr:config.ServerAddr,netWork:config.Network,lock:new(sync.RWMutex),timeoutSec:config.TimeoutSec}
	return memc,nil
}

func (memc *MemCache)Start()(err error){
	memc.lock.Lock()
	defer memc.lock.Unlock()
	if true == memc.isStopped{
		clientPool,err := utils.NewConnectPool(memc.config.ConnectMinNum,memc.config.ConnectMaxNum,func ()(item utils.ConnectPoolItem,err error){
			mcc := &MemCacheClient{conn:nil}
			mcc.conn, err = net.Dial(memc.netWork,memc.serverAddr)
			if nil != err{
				return nil,CACHESERVER_ERR
			}
			mcc.isAlive = true
			mcc.timeout = memc.timeoutSec
			mcc.cacheBuff = make([]byte,512)
			mcc.bytesBuff = new(bytes.Buffer)
			mcc.lastActiveTime = utils.GetNowUnixSec()
			return utils.ConnectPoolItem(mcc),nil
		})
		if nil != err{
			
			return err
		}
		memc.clientPool = clientPool
		if nil == memc.clientPool.Start(memc.timeoutSec){
			memc.isStopped = false
		}
		return err
	}
	return nil
}

func (memc *MemCache)Stop()(err error){
	memc.lock.Lock()
	defer memc.lock.Unlock()
	if false == memc.isStopped{
		err = memc.clientPool.Stop()
		if utils.CONNECTPOOL_STOP_SUC == err{
			memc.isStopped = true
		}
		return err
	}
	return nil
}

func memCheckKey(key string) bool {
	if 0 == len(key) || len(key) > 250 {
		return false
	}
	for i := 0; i < len(key); i++ {
		if key[i] <= ' ' || key[i] == 0x7f {
			return false
		}
	}
	return true
}

func memScanGetResponseLine(line []byte, it *MemItem) (size int64, err error) {
	if len(line) < 2 && line[0] == 'V'{
		return -1,ErrMemUnexpectedError
	}
	params := bytes.Split(line[:len(line)-2],[]byte(" "))
	if len(params) >= 3{
		it.Key = utils.Bytes2String(params[0])
		if 3 == len(params){
			return strconv.ParseInt(utils.Bytes2String(params[2]), 10, 0)
		}
		return strconv.ParseInt(utils.Bytes2String(params[3]), 10, 0)
	}
	return -1,ErrMemUnexpectedError
}

func memGetCmd(mcc *MemCacheClient,key string)(item *MemItem,err error){
	_,err = mcc.conn.Write(utils.String2Bytes("get "+key+"\r\n"))
	if nil != err{
		return nil,CACHECLIENT_ERR
	}
	mcc.cacheBuff=mcc.cacheBuff[0:]
	ilen, err := mcc.conn.Read(mcc.cacheBuff)
	if nil != err{
		return nil,CACHECLIENT_ERR
	}
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write(mcc.cacheBuff[0:ilen])
	useLen := 0
	for {
		line, err := mcc.bytesBuff.ReadBytes('\n')
		useLen += len(line)
		if err != nil {
			return nil,err
		}
		if line[0] == 'E'{//"END\r\n"
			if ilen == useLen{
				return nil,ErrMemCacheMiss
			}
			continue
		}
		item = new(MemItem)
		size, err := memScanGetResponseLine(line, item)
		if err != nil {
			return nil,err
		}
		item.Value = make([]byte, size+2)
		valueReadLen := len(mcc.bytesBuff.Bytes())
		copy(item.Value,mcc.bytesBuff.Bytes())
		readLen := 0
		if size+2 > int64(valueReadLen){
			for{
				readLen, err = mcc.conn.Read(item.Value[valueReadLen:])
				if nil != err{
					return nil,CACHECLIENT_ERR
				}
				valueReadLen += readLen
				if int64(valueReadLen) == size+2{
					break
				}
			}
		}
		if !bytes.HasSuffix(item.Value, []byte("\r\n")) {
			item.Value = nil
			return nil,ErrMemCorrupt
		}
		item.Value = item.Value[:size]
		return item,nil
	}
}

func memSetCmd(mcc *MemCacheClient,item *MemItem)(err error){
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write([]byte("set "))
	mcc.bytesBuff.Write([]byte(item.Key))
	mcc.bytesBuff.Write([]byte(" "))
	mcc.bytesBuff.Write([]byte(strconv.Itoa(int(item.Flags))))
	mcc.bytesBuff.Write([]byte(" "))
	mcc.bytesBuff.Write([]byte(strconv.Itoa(int(item.Expiration))))
	mcc.bytesBuff.Write([]byte(" "))
	mcc.bytesBuff.Write([]byte(strconv.Itoa(len(item.Value))))
	mcc.bytesBuff.Write([]byte("\r\n"))
	mcc.bytesBuff.Write(item.Value)
	mcc.bytesBuff.Write([]byte("\r\n"))
	_,err = mcc.conn.Write(mcc.bytesBuff.Bytes())
	if nil != err{
		return CACHECLIENT_ERR
	}
	mcc.cacheBuff=mcc.cacheBuff[0:]
	ilen, err := mcc.conn.Read(mcc.cacheBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write(mcc.cacheBuff[0:ilen])
	useLen := 0
	for {
		line, err := mcc.bytesBuff.ReadBytes('\n')
		useLen += len(line)
		if err != nil {
			return err
		}
		if line[0] == 'S'{//"STORED\r\n"
			return nil
		}
		if line[0] == 'N'{//"NOT_STORED\r\n"
			return ErrMemNotStored
		}
		if line[0] == 'E'{//"END\r\n"
			if ilen == useLen{
				return ErrMemUnexpectedError
			}
			continue
		}
		return ErrMemUnexpectedError
	}
}


func memDeleteCmd(mcc *MemCacheClient,key string) error {
	_, err := mcc.conn.Write(utils.String2Bytes("delete "+key+"\r\n"))
	if err != nil {
		return CACHECLIENT_ERR
	}
	mcc.cacheBuff=mcc.cacheBuff[0:]
	ilen, err := mcc.conn.Read(mcc.cacheBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write(mcc.cacheBuff[0:ilen])
	useLen := 0
	for{
		line, err := mcc.bytesBuff.ReadBytes('\n')
		useLen += len(line)
		if err != nil {
			return err
		}
		if line[0] == 'D'{//"DELETED\r\n"
			return nil
		}
		if line[0] == 'N'{//"NOT_FOUND\r\n"
			return ErrMemCacheMiss
		}
		if line[0] == 'E'{//"END\r\n"
			if ilen == useLen{
				return ErrMemCacheMiss
			}
			continue
		}
		return ErrMemUnexpectedError
	}
}

func (memc *MemCache)Get(key string)(value []byte,err error){
	if false == memCheckKey(key){
		return nil,CACHEPARAM_ERR
	}
	memc.lock.RLock()
	if nil != memc.clientPool{
		memc.lock.RUnlock()
		item,err := memc.clientPool.Get()
		if nil == err{
			mcc,ok := item.(*MemCacheClient)
			if ok{
				defer memc.clientPool.Put(item)
				getItem,err := memGetCmd(mcc,key)
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
				if utils.CONNECTPOOL_STOP_SUC == memc.Stop(){
					return nil,CACHESTOP_SUC
				}
			}
			return nil,CACHESERVER_ERR
		}
	}
	memc.lock.RUnlock()
	return nil,CACHECLIENT_NIL
}

func (memc *MemCache)Set(key string,value []byte,expire uint32)(err error){
	if nil == value || false == memCheckKey(key){
		return CACHEPARAM_ERR
	}
	memc.lock.RLock()
	if nil != memc.clientPool{
		memc.lock.RUnlock()
		item,err := memc.clientPool.Get()
		if nil == err{
			mcc,ok := item.(*MemCacheClient)
			if ok{
				defer memc.clientPool.Put(item)
				newItem := &MemItem{Key:key,Value:value,Expiration:expire}
				err = memSetCmd(mcc,newItem)
				if err == CACHECLIENT_ERR{
					mcc.isAlive = false
				}else{
					mcc.lastActiveTime = utils.GetNowUnixSec()
				}
				return err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == memc.Stop(){
					return CACHESTOP_SUC
				}
			}
			return CACHESERVER_ERR
		}
	}
	memc.lock.RUnlock()
	return CACHECLIENT_NIL
}

func (memc *MemCache)Delete(key string)(sta bool,err error){
	if false == memCheckKey(key){
		return false,CACHEPARAM_ERR
	}
	memc.lock.RLock()
	if nil != memc.clientPool{
		memc.lock.RUnlock()
		item,err := memc.clientPool.Get()
		if nil == err{
			mcc,ok := item.(*MemCacheClient)
			if ok{
				defer memc.clientPool.Put(item)
				err = memDeleteCmd(mcc,key)
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
				if utils.CONNECTPOOL_STOP_SUC == memc.Stop(){
					return false,CACHESTOP_SUC
				}
			}
			return false,CACHESERVER_ERR
		}
	}
	return false,CACHECLIENT_NIL
}

func (memc *MemCache)Exists(key string)(sta bool,err error){
	item,err := memc.Get(key)
	if nil != item{
		return true,nil
	}
	return false,err
}