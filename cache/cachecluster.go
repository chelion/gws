package cache
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"sort"
	"sync"
	"errors"
	"strconv"
	"sync/atomic"
	"github.com/chelion/gws/utils"
)
//CRC32 Consistent Hashing
var(
	CLUSTERCACHE_PARAM_ERR = errors.New("cluster cache param error")
	CLUSTERCACHE_NOSERVER_ERR = errors.New("can't find the the key cache server")
)

type cacheServerConfig struct{
	cache Cache
	addr string
	virtualNodeNum int32
	activeVirtualNodeNum int32
}

type CacheVirtualNode struct{
	csc *cacheServerConfig 
	name string
}

type CacheCluster struct{
	rwLock *sync.RWMutex
	sortIndex []uint32
	caches map[string]Cache
	cacheVirtualNodeMap map[uint32]*CacheVirtualNode
}

func (cacheCluster *CacheCluster)Len() int{
	return len(cacheCluster.sortIndex)
}

func (cacheCluster *CacheCluster)Less(i, j int) bool{
	return cacheCluster.sortIndex[i] < cacheCluster.sortIndex[j]
}

func (cacheCluster *CacheCluster)Swap(i, j int){
	tmp := cacheCluster.sortIndex[i]
	cacheCluster.sortIndex[i] = cacheCluster.sortIndex[j]
	cacheCluster.sortIndex[j] = tmp
}

func (cacheCluster *CacheCluster)sort(){
	tmpInt := make([]uint32,len(cacheCluster.cacheVirtualNodeMap))
	index := 0
	for key,_ := range cacheCluster.cacheVirtualNodeMap{
		tmpInt[index] = key
		index++
	}
	cacheCluster.sortIndex = tmpInt
	sort.Sort(cacheCluster)
}

func NewCacheCluster()(*CacheCluster,error){
	cacheCluster := &CacheCluster{rwLock:new(sync.RWMutex),caches:make(map[string]Cache),cacheVirtualNodeMap:make(map[uint32]*CacheVirtualNode,0)}
	return cacheCluster,nil
}

func DestroyCacheCluster(cacheCluster *CacheCluster)error{
	cacheCluster.rwLock.Lock()
	for _,cache := range cacheCluster.caches{
		cache.Stop()
	}
	cacheCluster.rwLock.Unlock()
	return nil
}

func (cacheCluster *CacheCluster)AddCacheServer(cache Cache,addr string,virtualNodeNum int32)(error){
	var i int32
	if nil == cache || virtualNodeNum <= 0{
		return CLUSTERCACHE_PARAM_ERR
	}
	cacheCluster.rwLock.Lock()
	err := cache.Start()
	if nil == err{
		csc := &cacheServerConfig{cache:cache,addr:addr,virtualNodeNum:virtualNodeNum,activeVirtualNodeNum:0}
		for i=0;i<virtualNodeNum;i++{
			atomic.AddInt32(&csc.activeVirtualNodeNum,1)
			virtualNodeName := addr+"#"+strconv.Itoa(int(i))
			csnKey := utils.CRC32(utils.String2Bytes(virtualNodeName))
			cacheCluster.cacheVirtualNodeMap[csnKey] = &CacheVirtualNode{name:virtualNodeName,csc:csc}
		}
	}
	cacheCluster.caches[addr] = cache
	cacheCluster.sort()
	cacheCluster.rwLock.Unlock()
	return nil
}

func (cacheCluster *CacheCluster)addCacheVirtualNode(v string,cacheVirualNode *CacheVirtualNode){
	csnKey := utils.CRC32(utils.String2Bytes(v))
	cacheCluster.rwLock.Lock()
	cacheCluster.cacheVirtualNodeMap[csnKey] = cacheVirualNode
	cacheCluster.sort()
	cacheCluster.rwLock.Unlock()
}

func (cacheCluster *CacheCluster)removeCacheVirtualNode(cacheVirualNode *CacheVirtualNode){
	csnKey := utils.CRC32(utils.String2Bytes(cacheVirualNode.name))
	cacheCluster.rwLock.Lock()
	if _,ok := cacheCluster.cacheVirtualNodeMap[csnKey];!ok{
		cacheCluster.rwLock.Unlock()
		return
	}
	delete(cacheCluster.cacheVirtualNodeMap,csnKey)
	if 0 == atomic.AddInt32(&(cacheVirualNode.csc.activeVirtualNodeNum),-1){
		if nil == cacheVirualNode.csc.cache.Stop(){
			delete(cacheCluster.caches,cacheVirualNode.csc.addr)
		}
	}
	cacheCluster.sort()
	cacheCluster.rwLock.Unlock()
}

func binarySearch(sortedList []uint32, lookingFor uint32) int {//2分查找
	var lt int = 0
	var gt int = len(sortedList)-1
	cnt := 0
	for lt <= gt{
		mid := lt + (gt-lt)/2
		midValue := sortedList[mid]
		midrValue := midValue
		if mid >=1 {
			midrValue = sortedList[mid - 1]
		}
		if midValue >= lookingFor && midrValue <= lookingFor{
			return mid
		}else if midValue > lookingFor{
			gt = mid - 1
		}else{
			lt = mid + 1
		}
		cnt ++
	}
	return 0
}

func (cacheCluster *CacheCluster)findCacheVirtualNode(key string)(cacheVirtualNode *CacheVirtualNode){
	cacheVirtualNode = nil
	if 0 == len(cacheCluster.cacheVirtualNodeMap) || 0 == len(key){
		return
	}
	keyCRC32 := utils.CRC32(utils.String2Bytes(key))
	sortIndex := cacheCluster.sortIndex
	keyId := binarySearch(sortIndex,keyCRC32)
	cacheVirtualNode = cacheCluster.cacheVirtualNodeMap[sortIndex[keyId]]
	return
}

func (cacheCluster *CacheCluster)Get(key string)(value []byte,err error){
	cacheCluster.rwLock.RLock()
	cacheVirtualNode := cacheCluster.findCacheVirtualNode(key)
	if nil == cacheVirtualNode || nil == cacheVirtualNode.csc.cache{
		cacheCluster.rwLock.RUnlock()
		return nil,CLUSTERCACHE_NOSERVER_ERR
	}
	value,err = cacheVirtualNode.csc.cache.Get(key)
	if err == CACHESERVER_ERR{
		cacheCluster.rwLock.RUnlock()
		cacheCluster.removeCacheVirtualNode(cacheVirtualNode)
	}
	cacheCluster.rwLock.RUnlock()
	return
}

func (cacheCluster *CacheCluster)Set(key string,value []byte,expire uint32)(err error){
	cacheCluster.rwLock.RLock()
	cacheVirtualNode := cacheCluster.findCacheVirtualNode(key)
	if nil == cacheVirtualNode || nil == cacheVirtualNode.csc || nil == cacheVirtualNode.csc.cache{
		cacheCluster.rwLock.RUnlock()
		return CLUSTERCACHE_NOSERVER_ERR
	}
	err = cacheVirtualNode.csc.cache.Set(key,value,expire)
	if err == CACHESERVER_ERR{
		cacheCluster.rwLock.RUnlock()
		cacheCluster.removeCacheVirtualNode(cacheVirtualNode)
	}
	cacheCluster.rwLock.RUnlock()
	return
}

func (cacheCluster *CacheCluster)Delete(key string)(sta bool,err error){
	cacheCluster.rwLock.RLock()
	cacheVirtualNode := cacheCluster.findCacheVirtualNode(key)
	if nil == cacheVirtualNode || nil == cacheVirtualNode.csc.cache{
		cacheCluster.rwLock.RUnlock()
		return false,CLUSTERCACHE_NOSERVER_ERR
	}
	sta,err = cacheVirtualNode.csc.cache.Delete(key)
	if err == CACHESERVER_ERR{
		cacheCluster.rwLock.RUnlock()
		cacheCluster.removeCacheVirtualNode(cacheVirtualNode)
	}
	cacheCluster.rwLock.RUnlock()
	return
}

func (cacheCluster *CacheCluster)Exists(key string)(sta bool,err error){
	cacheCluster.rwLock.RLock()
	cacheVirtualNode := cacheCluster.findCacheVirtualNode(key)
	if nil == cacheVirtualNode || nil == cacheVirtualNode.csc.cache{
		cacheCluster.rwLock.RUnlock()
		return false,CLUSTERCACHE_NOSERVER_ERR
	}
	sta,err = cacheVirtualNode.csc.cache.Exists(key)
	if err == CACHESERVER_ERR{
		cacheCluster.rwLock.RUnlock()
		cacheCluster.removeCacheVirtualNode(cacheVirtualNode)
	}
	cacheCluster.rwLock.RUnlock()
	return
}
