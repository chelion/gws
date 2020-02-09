package cache
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"fmt"
	"sync"
	"testing"
)

func BenchmarkCacheCluster(b *testing.B){
	addr := "127.0.0.1:11611"
	cc,err := NewFastCache(&CacheConfig{"tcp",addr,30,50,30})
	if nil != err{
		fmt.Println(err)
		return
	}
	cacheCluster := NewCacheCluster()
	err = cacheCluster.AddCacheServer(cc,addr,16)
	if nil != err{
		fmt.Println(err)
		return
	}
	err = cacheCluster.Set("hello",[]byte("world"),0)
	if nil != err{
		fmt.Println(err)
	}else{
		fmt.Println("set success->"+"hello")
	}
	wg := sync.WaitGroup{}
	wg.Add(30)
	for j:=0;j < 30;j++{
		go func(j int,wg *sync.WaitGroup){
			for i := 0; i < 50000; i++ {
				gdata,err := cacheCluster.Get("hello")
				if nil == err{
					_ = gdata
					//fmt.Println("get success->",string(gdata))
				}
			}
			wg.Done()
		}(j,&wg)
	}
	b.N = 30 * 50000
	wg.Wait()
}
