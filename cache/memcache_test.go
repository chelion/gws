package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

/*
import(
	"fmt"
	"sync"
	"bytes"
	"strconv"
    "encoding/gob"
	"testing"
)


func BenchmarkMemCache(b *testing.B){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test memcache cache--------")
	cc,err := NewMemCache(&CacheConfig{"tcp","127.0.0.1:11211",5,10,10})
	if nil != err{
		b.Errorf("new memcache fail\n")
		return
	}
	err = cc.Start()
	if nil != err{
		b.Errorf(err.Error())
		return
	}else{
		fmt.Println("memcache cache init succes")
	}
	defer cc.Stop()
	wg := sync.WaitGroup{}
	b.N = 1000 * 100
	wg.Add(1000)
	for i:=0;i<1000;i++{
		go func(i int){
			for j:=0;j<100;j++{
				
			name := "yilin"+strconv.Itoa(i)
			student := &Student{Name:name,Age:i,Addr:"shenzhen,china"}
			var result bytes.Buffer
			encoder := gob.NewEncoder(&result)
			encoder.Encode(student)
			userBytes := result.Bytes()
			err = cc.Set(name,userBytes,0)
			if nil != err{
				if err == CACHESTOP_SUC{
					cc.Start()
					fmt.Println("set")
				}
			}else{
				fmt.Println("set success")
			}
			gdata,err := cc.Get(name)
			if nil == err{
				if nil != gdata{
					var stu Student
					decoder := gob.NewDecoder(bytes.NewReader(gdata))
					decoder.Decode(&stu)
					fmt.Println(stu.Name,stu.Age,stu.Addr)
				}else{
					fmt.Println("get")
				}
			}else{
				if err == CACHESTOP_SUC{
					cc.Start()
				}else{
					fmt.Println(err)
				}
			}
			
			sta,e0 := cc.Delete(name)
			if nil != e0{
				if e0 == CACHESTOP_SUC{
					cc.Start()
				}
			}else{
				if true == sta{
					fmt.Println("delete")
				}
			}
			}
			
			sta,e := cc.Exists("hello")
			if nil != e{
				b.Errorf(e.Error())
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
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println("--------end test memcache cache--------")
}
*/