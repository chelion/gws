package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
import(
	"fmt"
	"sync"
//	"bytes"
   // "encoding/gob"
	"testing"
)

func TestFastcacheCache(t *testing.T){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test Fastcache cache--------")
	cc,err := NewFastCache(&CacheConfig{"tcp","127.0.0.1:11611",100,500,30})
	if nil != err{
		t.Errorf("new Fastcache fail\n")
		return
	}
	err = cc.Start()
	if nil != err{
		t.Errorf(err.Error())
	}else{
		fmt.Println("Fastcache cache init succes")
	}
	defer cc.Stop()
	wg := sync.WaitGroup{}
	
	wg.Add(500)
	for i:=0;i<500;i++{
		go func(i int){
			for j:=0;j<100;j++{
				student := &Student{Name:"yilin",Age:i,Addr:"shenzhen,china"}
				var result bytes.Buffer
				encoder := gob.NewEncoder(&result)
				encoder.Encode(student)
				userBytes := result.Bytes()
				err = cc.Set("hello",userBytes,0)
				if nil != err{
					//t.Errorf(err.Error())
					if err == CACHESTOP_SUC{
						cc.Start()
					}
				}else{
					fmt.Println("set success")
				}
				
				gdata,err := cc.Get("hello")
				if nil == err && nil != gdata{
					fmt.Print("get success->")
					var stu Student
					decoder := gob.NewDecoder(bytes.NewReader(gdata))
					decoder.Decode(&stu)
					fmt.Println(stu.Name,stu.Age,stu.Addr)
				}else{
					//t.Errorf(err.Error())
					if err == CACHESTOP_SUC{
						cc.Start()
					}
				}
				sta,e0 := cc.Delete("hello")
				if nil != e0{
				//	t.Errorf(e0.Error())
					if e0 == CACHESTOP_SUC{
						cc.Start()
					}
				}else{
					if true == sta{
						fmt.Println("delete success")
					}
				}
				sta,e := cc.Exists("hello")
				if nil != e{
					if e == CACHESTOP_SUC{
						cc.Start()
					}
				}else{
					if true == sta{
						fmt.Println("exists")
					}else{
						fmt.Println("not exists")
					}
				}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println("--------end test Fastcache cache--------")
}


func BenchmarkFastcache(b *testing.B) {
	fmt.Println("--------start test Fastcache cache Benchmark--------")
	cc,err := NewFastCache(&CacheConfig{"tcp","127.0.0.1:11611",30,50,30})
	if nil != err{
		b.Errorf("new Fastcache fail\n")
	}
	err = cc.Start()
	if nil != err{
		b.Errorf(err.Error())
		return
	}else{
		fmt.Println("Fastcache cache init succes")
	}
	err = cc.Set("hello",[]byte("world"),0)
	if nil != err{
		b.Errorf(err.Error())
	}else{
		fmt.Println("set success->"+"hello")
	}
	wg := sync.WaitGroup{}
	wg.Add(30)
	for j:=0;j < 30;j++{
		go func(j int,wg *sync.WaitGroup){
			for i := 0; i < 5000; i++ {
				gdata,err := cc.Get("hello")
				if nil == err{
					_= gdata
					//fmt.Println("get success->",string(gdata))
				}	
			}
			wg.Done()
		}(j,&wg)
	}
	b.N = 30 * 5000
	wg.Wait()
	cc.Stop()
	fmt.Println("--------end test Fastcache cache Benchmark--------")
}
*/