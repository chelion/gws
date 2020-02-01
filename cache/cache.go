package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"errors"
)

var(
	CACHEMAXSIZE_OVER = errors.New("cache over max size")
	CACHECLIENT_NIL  = errors.New("cache client is nil")
	CACHEPARAM_ERR = errors.New("cache param is error")
	CACHECLIENT_ERR = errors.New("cache client is error")
	CACHESERVER_ERR = errors.New("cache server is error")
	CACHESTOP_SUC = errors.New("cache server is stop suc")
)

type CacheConfig struct{
	Network string
	ServerAddr string
	ConnectMinNum int64
	ConnectMaxNum int64
	TimeoutSec int64
}

type Cache interface{
	Start()(err error)//初始化启动缓存服务
	Stop()(err error)//停用缓存服务
	Get(key string)(value []byte,err error)//获取Key的值
	Set(key string,value []byte,expire int)(err error)//设置Key的值
	Delete(key string)(sta bool,err error)//删除指定Key的值
	Exists(key string)(sta bool,err error)//查看Key是否存在
}