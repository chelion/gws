package cache
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"io"
	"sync"
	"net"
	"bytes"
	"errors"
	"github.com/chelion/gws/utils"
)
const(
	CLIENTMINNUM = 10
	CLIENTMAXNUM = 30
	CACHEMAXSIZE = 64 * 1024 - 16 - 4 - 1
)

var (
	ErrFastCacheMiss = errors.New("fastcache: cache miss")
	ErrFastNotStored = errors.New("fastcache: item not stored")
	ErrFastCorrupt = errors.New("fastcache: corrupt get result read")
)

type FastItem struct {
	Key string
	Value []byte
}

type FastCache struct{
	clientPool *utils.ConnectPool
	config *CacheConfig
	serverAddr string
	netWork string
	lock *sync.RWMutex
	timeoutSec int64
	isStopped bool
}

type FastCacheClient struct{
	conn net.Conn
	cacheBuff []byte
	bytesBuff *bytes.Buffer
	isAlive bool
	timeout int64
	lastActiveTime int64
}


func (fcc *FastCacheClient)Open()(err error){
	return nil
}

func (fcc *FastCacheClient)Close()(err error){
	err = nil
	if nil != fcc.conn{
		err = fcc.conn.Close()
		if nil == err{
			fcc.conn = nil
		}
	}
	return 
}

func (fcc *FastCacheClient)IsAlive()(sta bool){
	if utils.GetNowUnixSec() - fcc.lastActiveTime > fcc.timeout {
		_,err := fcc.conn.Write([]byte("*PNG\r\n"))
		if nil != err{
			fcc.isAlive = false
			return fcc.isAlive 
		}
		fcc.cacheBuff=fcc.cacheBuff[0:]
		ilen, err := fcc.conn.Read(fcc.cacheBuff)
		if nil != err{
			fcc.isAlive = false
		}else{
			if 0 == bytes.Compare(fcc.cacheBuff[:ilen],[]byte("+PNG\r\n")){
				fcc.lastActiveTime = utils.GetNowUnixSec()
				fcc.isAlive = true
				return true
			}
		}
	}
	return fcc.isAlive
}

func NewFastCache(config *CacheConfig)(fcc *FastCache,err error){
	fcc = &FastCache{clientPool:nil,config:config,isStopped:true,serverAddr:config.ServerAddr,netWork:config.Network,lock:new(sync.RWMutex),timeoutSec:config.TimeoutSec}
	return fcc,nil
}

func (fastc *FastCache)Start()(err error){
	fastc.lock.Lock()
	defer fastc.lock.Unlock()
	if true == fastc.isStopped{
		clientPool,err := utils.NewConnectPool(fastc.config.ConnectMinNum,fastc.config.ConnectMaxNum,func ()(item utils.ConnectPoolItem,err error){
			rcc := &FastCacheClient{conn:nil}
			rcc.conn, err = net.Dial(fastc.netWork,fastc.serverAddr)
			if nil != err{
				return nil,CACHESERVER_ERR
			}
			rcc.isAlive = true
			rcc.timeout = fastc.timeoutSec
			rcc.cacheBuff = make([]byte,512)
			rcc.bytesBuff = new(bytes.Buffer)
			rcc.lastActiveTime = utils.GetNowUnixSec()
			return utils.ConnectPoolItem(rcc),nil
		})
		if nil != err{
			
			return err
		}
		fastc.clientPool = clientPool
		if nil == fastc.clientPool.Start(fastc.timeoutSec){
			fastc.isStopped = false
		}
		return err
	}
	return nil
}

func (fastc *FastCache)Stop()(err error){
	fastc.lock.Lock()
	defer fastc.lock.Unlock()
	if false == fastc.isStopped{
		err = fastc.clientPool.Stop()
		if utils.CONNECTPOOL_STOP_SUC == err{
			fastc.isStopped = true
		}
		return err
	}
	return nil
}


func fastGetCmd(fcc *FastCacheClient,key string)(item *FastItem,err error){
	fcc.bytesBuff.Reset()
	fcc.bytesBuff.Write([]byte("*GET\r\n$"))
	fcc.bytesBuff.Write(utils.Int32ToBytes(int32(len(key))))
	fcc.bytesBuff.Write(utils.Int32ToBytes(0))
	fcc.bytesBuff.Write([]byte("\r\n"))
	fcc.bytesBuff.Write(utils.String2Bytes(key))
	fcc.bytesBuff.Write([]byte("\r\n"))
	_,err = fcc.conn.Write(fcc.bytesBuff.Bytes())
	if nil != err{
		return nil,CACHECLIENT_ERR
	}
	fcc.cacheBuff=fcc.cacheBuff[0:]
	ilen, err := fcc.conn.Read(fcc.cacheBuff)
	if nil != err{
		if (ilen == 0 && err == io.EOF) || bytes.Contains(utils.String2Bytes(err.Error()),[]byte("close")){
			return nil,CACHECLIENT_ERR
		}
	}
	if 7 <= ilen && fcc.cacheBuff[0] == '+'{
		dataLen := utils.BytesToInt32(fcc.cacheBuff[1:5])
		if 0 == dataLen{
			return nil,ErrFastCacheMiss
		}
		item = new(FastItem)
		item.Value = make([]byte, dataLen+2)
		valueReadLen := ilen-7
		copy(item.Value,fcc.cacheBuff[7:ilen])
		readLen := 0
		if int(dataLen+2) > valueReadLen{
			for{
				readLen, err = fcc.conn.Read(item.Value[valueReadLen:])
				if nil != err{
					return nil,CACHECLIENT_ERR
				}
				valueReadLen += readLen
				if valueReadLen == int(dataLen+2){
					break
				}
			}
		}
		if !bytes.HasSuffix(item.Value, []byte("\r\n")) {
			item.Value = nil
			return nil,ErrFastCorrupt
		}
		item.Value = item.Value[:dataLen]
		return item,nil
	}else{
		return nil,ErrFastCacheMiss
	}
}

func fastSetCmd(fcc *FastCacheClient,item *FastItem)(err error){
	fcc.bytesBuff.Reset()
	fcc.bytesBuff.Write([]byte("*SET\r\n$"))
	fcc.bytesBuff.Write(utils.Int32ToBytes(int32(len(item.Key))))
	fcc.bytesBuff.Write(utils.Int32ToBytes(int32(len(item.Value))))
	fcc.bytesBuff.Write([]byte("\r\n"))
	fcc.bytesBuff.Write(utils.String2Bytes(item.Key))
	fcc.bytesBuff.Write([]byte("\r\n"))
	fcc.bytesBuff.Write(item.Value)
	fcc.bytesBuff.Write([]byte("\r\n"))
	
	_,err = fcc.conn.Write(fcc.bytesBuff.Bytes())
	if nil != err{
		return CACHECLIENT_ERR
	}
	fcc.cacheBuff=fcc.cacheBuff[0:]
	ilen, err := fcc.conn.Read(fcc.cacheBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	fcc.bytesBuff.Reset()
	fcc.bytesBuff.Write(fcc.cacheBuff[0:ilen])
	line, err := fcc.bytesBuff.ReadBytes('\n')
	if err != nil {
		return err
	}
	if 0 == bytes.Compare(line,[]byte("+OK\r\n")) {
		return nil
	}
	return ErrFastNotStored
}

func fastDeleteExistsCmd(cmd []byte,fcc *FastCacheClient,key string)(err error){
	fcc.bytesBuff.Reset()
	fcc.bytesBuff.Write(cmd)
	fcc.bytesBuff.Write(utils.Int32ToBytes(int32(len(key))))
	fcc.bytesBuff.Write(utils.Int32ToBytes(0))
	fcc.bytesBuff.Write([]byte("\r\n"))
	fcc.bytesBuff.Write(utils.String2Bytes(key))
	fcc.bytesBuff.Write([]byte("\r\n"))
	_,err = fcc.conn.Write(fcc.bytesBuff.Bytes())
	if nil != err{
		return CACHECLIENT_ERR
	}
	fcc.cacheBuff=fcc.cacheBuff[0:]
	ilen, err := fcc.conn.Read(fcc.cacheBuff)
	if nil != err{
		return CACHECLIENT_ERR
	}
	fcc.bytesBuff.Reset()
	fcc.bytesBuff.Write(fcc.cacheBuff[0:ilen])
	line, err := fcc.bytesBuff.ReadBytes('\n')
	if err != nil {
		return err
	}
	if len(line) > 2 && line[0] == ':'{
		if line[1] == '1'{
			return nil
		}
	}
	return ErrMemCacheMiss
}


func (fastc *FastCache)Get(key string)(value []byte,err error){
	if 0 == len(key) || len(key) >CACHEMAXSIZE{
		return nil,CACHEPARAM_ERR
	}
	fastc.lock.RLock()
	if nil != fastc.clientPool{
		fastc.lock.RUnlock()
		item,err := fastc.clientPool.Get()
		if nil == err{
			fcc,ok := item.(*FastCacheClient)
			if ok{
				defer fastc.clientPool.Put(item)
				getItem,err := fastGetCmd(fcc,key)
				if err == CACHECLIENT_ERR{
					fcc.isAlive = false
				}else{
					fcc.lastActiveTime = utils.GetNowUnixSec()
					if nil != getItem{
						return getItem.Value,nil
					}
				}
				return nil,err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fastc.Stop(){
					return nil,CACHESTOP_SUC
				}
			}
			return nil,CACHESERVER_ERR
		}
	}
	fastc.lock.RUnlock()
	return nil,CACHECLIENT_NIL
}


func (fastc *FastCache)Set(key string,value []byte,expire uint32)(err error){
	if nil == value || 0 == len(value) || 0 == len(key) || 
		len(key) >CACHEMAXSIZE || len(value) > CACHEMAXSIZE || len(key)+len(value) > CACHEMAXSIZE{
		return CACHEPARAM_ERR
	}
	fastc.lock.RLock()
	if nil != fastc.clientPool{
		fastc.lock.RUnlock()
		item,err := fastc.clientPool.Get()
		if nil == err{
			rcc,ok := item.(*FastCacheClient)
			if ok{
				defer fastc.clientPool.Put(item)
				newItem := &FastItem{Key:key,Value:value}
				err = fastSetCmd(rcc,newItem)
				if err == CACHECLIENT_ERR{
					rcc.isAlive = false
				}else{
					rcc.lastActiveTime = utils.GetNowUnixSec()
				}
				return err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fastc.Stop(){
					return CACHESTOP_SUC
				}
			}
			return CACHESERVER_ERR
		}
	}
	fastc.lock.RUnlock()
	return CACHECLIENT_NIL
}

func (fastc *FastCache)Delete(key string)(sta bool,err error){
	if 0 == len(key) || len(key) >CACHEMAXSIZE{
		return false,CACHEPARAM_ERR
	}
	fastc.lock.RLock()
	if nil != fastc.clientPool{
		fastc.lock.RUnlock()
		item,err := fastc.clientPool.Get()
		if nil == err{
			fcc,ok := item.(*FastCacheClient)
			if ok{
				defer fastc.clientPool.Put(item)
				err = fastDeleteExistsCmd([]byte("*DEL\r\n$"),fcc,key)
				if nil == err{
					return true,nil
				}
				if err == CACHECLIENT_ERR{
					fcc.isAlive = false
				}else{
					fcc.lastActiveTime = utils.GetNowUnixSec()
				}
				return false,err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fastc.Stop(){
					return false,CACHESTOP_SUC
				}
			}
			return false,CACHESERVER_ERR
		}
	}
	return false,CACHECLIENT_NIL
}

func (fastc *FastCache)Exists(key string)(sta bool,err error){
	if 0 == len(key) || len(key) >CACHEMAXSIZE{
		return false,CACHEPARAM_ERR
	}
	fastc.lock.RLock()
	if nil != fastc.clientPool{
		fastc.lock.RUnlock()
		item,err := fastc.clientPool.Get()
		if nil == err{
			fcc,ok := item.(*FastCacheClient)
			if ok{
				defer fastc.clientPool.Put(item)
				err = fastDeleteExistsCmd([]byte("*EXS\r\n$"),fcc,key)
				if nil == err{
					return true,nil
				}
				if err == CACHECLIENT_ERR{
					fcc.isAlive = false
				}else{
					fcc.lastActiveTime = utils.GetNowUnixSec()
				}
				return false,err
			}
		}else{
			if utils.CONNECTPOOL_NEEDSTOP == err{
				if utils.CONNECTPOOL_STOP_SUC == fastc.Stop(){
					return false,CACHESTOP_SUC
				}
			}
			return false,CACHESERVER_ERR
		}
	}
	return false,CACHECLIENT_NIL
}