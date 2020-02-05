package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.


import(
	"fmt"
	"net"
	"sync"
	"time"
	"bytes"
	"errors"
	"strconv"
	"github.com/chelion/gws/utils"
)

var (
	ErrRedisCacheMiss = errors.New("rediscache: cache miss")
	ErrRedisNotStored = errors.New("rediscache: item not stored")
	ErrRedisCorrupt = errors.New("rediscache: corrupt get result read")
)

type RedisCache struct{
	clientPool *utils.ConnectPool
	config *CacheConfig
	serverAddr string
	netWork string
	lock *sync.RWMutex
	timeoutSec int64
	isStopped bool
	stopChan chan struct{} 
}

type RedisCacheClient struct{
	conn net.Conn
	cacheBuff []byte
	bytesBuff *bytes.Buffer
	isAlive bool
	timeout int64
	lastActiveTime int64
}

type RedisItem struct {
	Key string
	Value []byte
	Expiration uint32
}

func (rcc *RedisCacheClient)Open()(err error){
	return nil
}

func (rcc *RedisCacheClient)Close()(err error){
	err = nil
	if nil != rcc.conn{
		err = rcc.conn.Close()
		fmt.Println("close")
		if nil == err{
			rcc.conn = nil
		}
	}
	return 
}

func (rcc *RedisCacheClient)IsAlive()(sta bool){
	if utils.GetNowUnixSec() - rcc.lastActiveTime > rcc.timeout {
		_,err := rcc.conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
		if nil != err{
			rcc.isAlive = false
			return rcc.isAlive
		}
		rcc.cacheBuff=rcc.cacheBuff[0:]
		ilen, err := rcc.conn.Read(rcc.cacheBuff)
		if nil != err{
			rcc.isAlive = false
		}else{
			if 0 == bytes.Compare(rcc.cacheBuff[:ilen],[]byte("+PONG\r\n")){
				fmt.Println("good ping")
				rcc.lastActiveTime = utils.GetNowUnixSec()
				rcc.isAlive = true
				return true
			}
		}
	}
	return rcc.isAlive
}


func NewRedisCache(config *CacheConfig)(redisc *RedisCache,err error){
	redisc = &RedisCache{clientPool:nil,config:config,isStopped:true,serverAddr:config.ServerAddr,netWork:config.Network,
		lock:new(sync.RWMutex),timeoutSec:config.TimeoutSec,stopChan:make(chan struct{})}
	return redisc,nil
}

func (redisc *RedisCache)tick(){
	var i int64
	ticker := time.NewTicker(time.Duration(redisc.timeoutSec)*time.Second)
	defer ticker.Stop()
	for{
		select{
			case <-ticker.C:{
				redisc.lock.Lock()
				if nil != redisc.clientPool{
					for i=0;i<redisc.clientPool.GetCurrentIdleNum();i++{
						item,err := redisc.clientPool.Get()
						if nil == err{
							rcc,ok := item.(*RedisCacheClient)
							if ok{
								rcc.IsAlive()
								redisc.clientPool.Put(item)
							}
						}
					}
				}
				redisc.lock.Unlock()
			}
			case <-redisc.stopChan:{
				return
			}
		}
	}
}

func (redisc *RedisCache)Start()(err error){
	fmt.Println("start---------------------------")
	redisc.lock.Lock()
	defer redisc.lock.Unlock()
	if true == redisc.isStopped{
		fmt.Println("start clientPool")
		clientPool,err := utils.NewConnectPool(redisc.config.ConnectMinNum,redisc.config.ConnectMaxNum,func ()(item utils.ConnectPoolItem,err error){
			rcc := &RedisCacheClient{conn:nil}
			rcc.conn, err = net.Dial(redisc.netWork,redisc.serverAddr)
			if nil != err{
				return nil,CACHESERVER_ERR
			}
			rcc.isAlive = true
			rcc.timeout = redisc.timeoutSec
			rcc.cacheBuff = make([]byte,512)
			rcc.bytesBuff = new(bytes.Buffer)
			rcc.lastActiveTime = utils.GetNowUnixSec()
			return utils.ConnectPoolItem(rcc),nil
		})
		if nil != err{
			
			return err
		}
		redisc.clientPool = clientPool
		if nil == redisc.clientPool.Start(redisc.timeoutSec){
			redisc.isStopped = false
			go redisc.tick()
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
			redisc.stopChan <- struct{}{}
		}
		return err
	}
	return nil
}

func bitsUint32(a uint32)int{
    i := 0
    for a >0 {
        a /= 10
        i++
	}
	return i
}

func redisGetCmd(rcc *RedisCacheClient,key string)(item *RedisItem,err error){
	rcc.bytesBuff.Reset()
	rcc.bytesBuff.Write([]byte("*2\r\n$3\r\nGET\r\n$"))
	rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(len(key))))
	rcc.bytesBuff.Write([]byte("\r\n"))
	rcc.bytesBuff.Write(utils.String2Bytes(key))
	rcc.bytesBuff.Write([]byte("\r\n"))
	_,err = rcc.conn.Write(rcc.bytesBuff.Bytes())
	if nil != err{
		return nil,CACHECLIENT_ERR
	}
	rcc.cacheBuff=rcc.cacheBuff[0:]
	ilen, err := rcc.conn.Read(rcc.cacheBuff)
	if nil != err{
		return nil,CACHECLIENT_ERR
	}
	rcc.bytesBuff.Reset()
	rcc.bytesBuff.Write(rcc.cacheBuff[0:ilen])
	useLen := 0
	line, err := rcc.bytesBuff.ReadBytes('\n')
	useLen += len(line)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	if line[0] == '$'{
		dataLen,err := strconv.Atoi(utils.Bytes2String(line[1:useLen-2]))
		if nil != err || -1 == dataLen{
			return nil,ErrRedisCacheMiss
		}
		item = new(RedisItem)
		item.Value = make([]byte, dataLen+2)
		valueReadLen := len(rcc.bytesBuff.Bytes())
		copy(item.Value,rcc.bytesBuff.Bytes())
		readLen := 0
		if dataLen+2 > valueReadLen{
			for{
				readLen, err = rcc.conn.Read(item.Value[valueReadLen:])
				if nil != err{
					return nil,CACHECLIENT_ERR
				}
				valueReadLen += readLen
				if valueReadLen == dataLen+2{
					break
				}
			}
		}
		if !bytes.HasSuffix(item.Value, []byte("\r\n")) {
			item.Value = nil
			return nil,ErrRedisCorrupt
		}
		item.Value = item.Value[:dataLen]
		return item,nil
	}else{
		return nil,ErrRedisCacheMiss
	}
}

func redisSetCmd(rcc *RedisCacheClient,item *RedisItem)(err error){
	rcc.bytesBuff.Reset()
	if item.Expiration == 0{
		rcc.bytesBuff.Write([]byte("*3\r\n$3\r\nSET\r\n$"))
		rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(len(item.Key))))
		rcc.bytesBuff.Write([]byte("\r\n"))
		rcc.bytesBuff.Write(utils.String2Bytes(item.Key))
		rcc.bytesBuff.Write([]byte("\r\n$"))
		rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(len(item.Value))))
		rcc.bytesBuff.Write([]byte("\r\n"))
		rcc.bytesBuff.Write(item.Value)
		rcc.bytesBuff.Write([]byte("\r\n"))
	}else{
		rcc.bytesBuff.Write([]byte("*5\r\n$3\r\nSET\r\n$"))
		rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(len(item.Key))))
		rcc.bytesBuff.Write([]byte("\r\n"))
		rcc.bytesBuff.Write(utils.String2Bytes(item.Key))
		rcc.bytesBuff.Write([]byte("\r\n$"))
		rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(len(item.Value))))
		rcc.bytesBuff.Write([]byte("\r\n"))
		rcc.bytesBuff.Write(item.Value)
		rcc.bytesBuff.Write([]byte("\r\n$2\r\nEX\r\n$"))
		rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(bitsUint32(item.Expiration))))
		rcc.bytesBuff.Write([]byte("\r\n"))
		rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(int(item.Expiration))))
		rcc.bytesBuff.Write([]byte("\r\n"))
	}
	_,err = rcc.conn.Write(rcc.bytesBuff.Bytes())
	if nil != err{
		return CACHECLIENT_ERR
	}
	rcc.cacheBuff=rcc.cacheBuff[0:]
	ilen, err := rcc.conn.Read(rcc.cacheBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	rcc.bytesBuff.Reset()
	rcc.bytesBuff.Write(rcc.cacheBuff[0:ilen])
	line, err := rcc.bytesBuff.ReadBytes('\n')
	if err != nil {
		return err
	}
	if 0 == bytes.Compare(line,[]byte("+OK\r\n")) {
		return nil
	}
	return ErrRedisNotStored
}

func redisDeleteExistsCmd(cmd []byte,rcc *RedisCacheClient,key string)(err error){
	rcc.bytesBuff.Reset()
	rcc.bytesBuff.Write(cmd)
	rcc.bytesBuff.Write(utils.String2Bytes(strconv.Itoa(len(key))))
	rcc.bytesBuff.Write([]byte("\r\n"))
	rcc.bytesBuff.Write(utils.String2Bytes(key))
	rcc.bytesBuff.Write([]byte("\r\n"))
	_,err = rcc.conn.Write(rcc.bytesBuff.Bytes())
	if nil != err{
		return CACHECLIENT_ERR
	}
	rcc.cacheBuff=rcc.cacheBuff[0:]
	ilen, err := rcc.conn.Read(rcc.cacheBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	rcc.bytesBuff.Reset()
	rcc.bytesBuff.Write(rcc.cacheBuff[0:ilen])
	line, err := rcc.bytesBuff.ReadBytes('\n')
	if err != nil {
		return err
	}
	if line[0] == ':' && len(line) > 2{
		if line[1] == '1'{
			return nil
		}
	}
	return ErrMemCacheMiss
}

func (redisc *RedisCache)Get(key string)(value []byte,err error){
	if 0 == len(key){
		return nil,CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			rcc,ok := item.(*RedisCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				getItem,err := redisGetCmd(rcc,key)
				if err == CACHECLIENT_ERR{
					rcc.isAlive = false
				}else{
					rcc.lastActiveTime = utils.GetNowUnixSec()
					if nil != getItem{
						return getItem.Value,nil
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

func (redisc *RedisCache)Set(key string,value []byte,expire uint32)(err error){
	if nil == value || 0 == len(key) || 0 == len(value){
		return CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			rcc,ok := item.(*RedisCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				newItem := &RedisItem{Key:key,Value:value,Expiration:expire}
				err = redisSetCmd(rcc,newItem)
				if err == CACHECLIENT_ERR{
					rcc.isAlive = false
				}else{
					rcc.lastActiveTime = utils.GetNowUnixSec()
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
	if 0 == len(key){
		return false,CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			rcc,ok := item.(*RedisCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				err = redisDeleteExistsCmd([]byte("*2\r\n$3\r\nDEL\r\n$"),rcc,key)
				if nil == err{
					return true,nil
				}
				if err == CACHECLIENT_ERR{
					rcc.isAlive = false
				}else{
					rcc.lastActiveTime = utils.GetNowUnixSec()
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

func (redisc *RedisCache)Exists(key string)(sta bool,err error){
	if 0 == len(key){
		return false,CACHEPARAM_ERR
	}
	redisc.lock.RLock()
	if nil != redisc.clientPool{
		redisc.lock.RUnlock()
		item,err := redisc.clientPool.Get()
		if nil == err{
			rcc,ok := item.(*RedisCacheClient)
			if ok{
				defer redisc.clientPool.Put(item)
				err = redisDeleteExistsCmd([]byte("*2\r\n$6\r\nEXISTS\r\n$"),rcc,key)
				if nil == err{
					return true,nil
				}
				if err == CACHECLIENT_ERR{
					rcc.isAlive = false
				}else{
					rcc.lastActiveTime = utils.GetNowUnixSec()
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