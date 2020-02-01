package memory
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"github.com/chelion/gws/cache"
)


type MCacheCluster struct{
	cacheCluster *cache.CacheCluster
}

// New slice number ccMap
func NewCCMap(cacheConfig []*MCacheConfig) *MCacheCluster {
	cacheCluster,err := cache.NewCacheCluster()
	if nil != err{
		panic(err)
	}
	for i:=0;i<len(cacheConfig);i++{
		err = cacheCluster.AddCacheServer(cacheConfig[i].Cache,cacheConfig[i].Addr,cacheConfig[i].VirtualNodeNum)
		if nil != err{
			panic(err)
		}
	}
	return &MCacheCluster{cacheCluster:cacheCluster}
}


// key is exist
func (mcc *MCacheCluster) IsExist(key string) bool {
	sta,_:= mcc.cacheCluster.Exists(key)
	return sta
}

// set key value
func (mcc *MCacheCluster) Set(key string, value []byte) {
	mcc.cacheCluster.Set(key,value,0)
}

// get by key
func (mcc *MCacheCluster) Get(key string) []byte {
	v,_ := mcc.cacheCluster.Get(key)
	return v
}

// delete by key
func (mcc *MCacheCluster) Delete(key string) {
	mcc.cacheCluster.Delete(key)
}

// update by key
// if key exist, update value
func (mcc *MCacheCluster) Update(key string, value []byte) {
	mcc.cacheCluster.Set(key,value,0)
}

// replace
// if key exist, update value.
// if key not exist, insert value.
func (mcc *MCacheCluster) Replace(key string, value []byte) {
	mcc.cacheCluster.Set(key,value,0)
}
