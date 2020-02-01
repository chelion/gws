package utils
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"fmt"
	"sync"
	"time"
	"sync/atomic"
	"testing"
)
type TestConnect struct{

}
var cnt int64
func (ts *TestConnect)Open()(err error){
	fmt.Println("open")
	time.Sleep(1*time.Millisecond)
	return nil
}

func (fcc *TestConnect)Close()(err error){
	fmt.Println("close")
	return nil
}

func (fcc *TestConnect)IsAlive()(sta bool){
	if atomic.LoadInt64(&cnt) == 100{
		atomic.StoreInt64(&cnt,0) 
		fmt.Println("IsAlive false")
		return false
	}
	atomic.AddInt64(&cnt,1) 
	return true
}

func BenchmarkConnectPool(b *testing.B){
	atomic.StoreInt64(&cnt,0) 
	connectPool,_ := NewConnectPool(100,300,func ()(item ConnectPoolItem,err error){
		ts := &TestConnect{}
		return ConnectPoolItem(ts),nil
	})
	connectPool.Start(30)
	wg := sync.WaitGroup{}
	wg.Add(b.N)
	fmt.Println("----start test connect pool----")
	s := time.Now()
	for i:=0;i<b.N;i++{
		go func(i int){
				item,err := connectPool.Get()
				if nil == err{
					time.Sleep(5*time.Millisecond)
					connectPool.Put(item)
				}
				wg.Done()
		}(i)
	}
	wg.Wait()
	connectPool.Stop()
	elapsed := time.Since(s)
    fmt.Println("start test connect pool elapsed:", elapsed)
	fmt.Println("----end test connect pool----")
}
