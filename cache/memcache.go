package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"fmt"
	"net"
	"sync"
	"bytes"
	"errors"
	"github.com/chelion/gws/utils"
)

var (
	versionCmd		= []byte("version \r\n")
	crlf            = []byte("\r\n")
	space           = []byte(" ")
	resultOK        = []byte("OK\r\n")
	resultStored    = []byte("STORED\r\n")
	resultNotStored = []byte("NOT_STORED\r\n")
	resultExists    = []byte("EXISTS\r\n")
	resultNotFound  = []byte("NOT_FOUND\r\n")
	resultDeleted   = []byte("DELETED\r\n")
	resultEnd       = []byte("END\r\n")
	resultOk        = []byte("OK\r\n")
	resultTouched   = []byte("TOUCHED\r\n")

	ErrCacheMiss = errors.New("memcache: cache miss")

	ErrCASConflict = errors.New("memcache: compare-and-swap conflict")

	ErrNotStored = errors.New("memcache: item not stored")

	ErrServerError = errors.New("memcache: server error")

	ErrMalformedKey = errors.New("malformed: key is too long or contains invalid characters")
)

type Item struct {
	Key string
	Value []byte
	Flags uint32
	Expiration int32
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
	readBuff []byte
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
		fmt.Println("close")
		if nil == err{
			mcc.conn = nil
		}
	}
	return 
}

func (mcc *MemCacheClient)IsAlive()(sta bool){
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


func NewMemCache(config *CacheConfig)(memc *MemCache,err error){
	memc = &MemCache{clientPool:nil,config:config,isStopped:true,serverAddr:config.ServerAddr,netWork:config.Network,lock:new(sync.RWMutex),timeoutSec:config.TimeoutSec}
	return memc,nil
}

func (memc *MemCache)Start()(err error){
	fmt.Println("start---------------------------")
	memc.lock.Lock()
	defer memc.lock.Unlock()
	if true == memc.isStopped{
		fmt.Println("start clientPool")
		clientPool,err := utils.NewConnectPool(memc.config.ConnectMinNum,memc.config.ConnectMaxNum,func ()(item utils.ConnectPoolItem,err error){
			mcc := &MemCacheClient{conn:nil}
			mcc.conn, err = net.Dial(memc.netWork,memc.serverAddr)
			if nil != err{
				return nil,CACHESERVER_ERR
			}
			mcc.isAlive = true
			mcc.timeout = memc.timeoutSec
			mcc.readBuff = make([]byte,512)
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
			fmt.Println("memc.isActive")
		}
		fmt.Println("start------------end---------------",err)
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

func checkKey(key string) bool {
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

func scanGetResponseLine(line []byte, it *Item) (size int, err error) {
	pattern := "VALUE %s %d %d %d\r\n"
	dest := []interface{}{&it.Key, &it.Flags, &size, &it.casid}
	if bytes.Count(line, space) == 3 {
		pattern = "VALUE %s %d %d\r\n"
		dest = dest[:3]
	}
	n, err := fmt.Sscanf(string(line), pattern, dest...)
	if err != nil || n != len(dest) {
		return -1, fmt.Errorf("memcache: unexpected line in get response: %q", line)
	}
	return size, nil
}

func getCmd(mcc *MemCacheClient,key string)(item *Item,err error){
	_,err = mcc.conn.Write(utils.String2Bytes("get "+key+"\r\n"))
	if nil != err{
		fmt.Println(err)
		return nil,CACHECLIENT_ERR
	}
	mcc.readBuff=mcc.readBuff[0:]
	ilen, err := mcc.conn.Read(mcc.readBuff)
	if nil != err{
		fmt.Println(err)
		return nil,CACHECLIENT_ERR
	}
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write(mcc.readBuff[0:ilen])
	useLen := 0
	for {
		line, err := mcc.bytesBuff.ReadBytes('\n')
		useLen += len(line)
		if err != nil {
			fmt.Println(err)
			return nil,err
		}
		if bytes.Equal(line, resultNotFound) {
			return nil,ErrCacheMiss
		}
		if bytes.Equal(line, resultEnd) {
			if ilen == useLen{
				return nil,ErrCacheMiss
			}
			continue
		}
		item = new(Item)
		size, err := scanGetResponseLine(line, item)
		if err != nil {
			return nil,err
		}
		item.Value = make([]byte, size+2)
		valueReadLen := len(mcc.bytesBuff.Bytes())
		copy(item.Value,mcc.bytesBuff.Bytes())
		readLen := 0
		if size+2 > valueReadLen{
			for{
				readLen, err = mcc.conn.Read(item.Value[valueReadLen:])
				if nil != err{
					return nil,CACHECLIENT_ERR
				}
				valueReadLen += readLen
				if valueReadLen == size+2{
					break
				}
			}
		}
		if !bytes.HasSuffix(item.Value, crlf) {
			item.Value = nil
			return nil,errors.New("memcache: corrupt get result read")
		}
		item.Value = item.Value[:size]
		return item,nil
	}
}

func setCmd(mcc *MemCacheClient,item *Item)(err error){
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write([]byte(fmt.Sprintf("set %s %d %d %d\r\n",
		item.Key, item.Flags, item.Expiration, len(item.Value))))
	
	mcc.bytesBuff.Write(item.Value)
	mcc.bytesBuff.Write(crlf)
	_,err = mcc.conn.Write(mcc.bytesBuff.Bytes())
	if nil != err{
		return CACHECLIENT_ERR
	}
	mcc.readBuff=mcc.readBuff[0:]
	ilen, err := mcc.conn.Read(mcc.readBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write(mcc.readBuff[0:ilen])
	useLen := 0
	for {
		line, err := mcc.bytesBuff.ReadBytes('\n')
		useLen += len(line)
		if err != nil {
			return err
		}
		if bytes.Equal(line, resultEnd) {
			if ilen == useLen{
				return ErrCacheMiss
			}
			continue
		}
		switch {
			case bytes.Equal(line, resultStored):
				return nil
			case bytes.Equal(line, resultNotStored):
				return ErrNotStored
			case bytes.Equal(line, resultExists):
				return ErrCASConflict
			case bytes.Equal(line, resultNotFound):
				return ErrCacheMiss
		}
		return fmt.Errorf("memcache: unexpected response line from set: %q",string(line))
	}
}


func deleteCmd(mcc *MemCacheClient,key string) error {
	_, err := mcc.conn.Write(utils.String2Bytes("delete "+key+"\r\n"))
	if err != nil {
		return CACHECLIENT_ERR
	}
	mcc.readBuff=mcc.readBuff[0:]
	ilen, err := mcc.conn.Read(mcc.readBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	mcc.bytesBuff.Reset()
	mcc.bytesBuff.Write(mcc.readBuff[0:ilen])
	useLen := 0
	for{
		line, err := mcc.bytesBuff.ReadBytes('\n')
		useLen += len(line)
		if err != nil {
			return err
		}
		if bytes.Equal(line, resultEnd) {
			if ilen == useLen{
				return ErrCacheMiss
			}
			continue
		}
		switch {
			case bytes.Equal(line, resultOK):
				return nil
			case bytes.Equal(line, resultDeleted):
				return nil
			case bytes.Equal(line, resultNotStored):
				return ErrNotStored
			case bytes.Equal(line, resultExists):
				return ErrCASConflict
			case bytes.Equal(line, resultNotFound):
				return ErrCacheMiss
		}
		return fmt.Errorf("memcache: unexpected response line: %q", string(line))
	}
}

func (memc *MemCache)Get(key string)(value []byte,err error){
	if false == checkKey(key){
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

func (memc *MemCache)Set(key string,value []byte,expire int)(err error){
	if nil == value || false == checkKey(key){
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
	if false == checkKey(key){
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
	/*
	if false == checkKey(key){
		return false,CACHEPARAM_ERR
	}

	return false,CACHECLIENT_NIL*/
	return true,nil
}