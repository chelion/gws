package cache

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import(
	"fmt"
	"sync"
	"bytes"
	"strconv"
    "encoding/gob"
	"testing"
)


func TestRedisCache(t *testing.T){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test rediscache cache--------")
	cc,err := NewRedisCache(&CacheConfig{"tcp","127.0.0.1:6379",5,10,10})
	if nil != err{
		t.Errorf("new rediscache fail\n")
		return
	}
	err = cc.Start()
	if nil != err{
		t.Errorf(err.Error())
		return
	}else{
		fmt.Println("rediscache cache init succes")
	}
	defer cc.Stop()
	student := &Student{Name:"yilin",Age:20,Addr:"shenzhen,china"}
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	encoder.Encode(student)
	userBytes := result.Bytes()
	err = cc.Set("redistest",userBytes,0)
	if nil != err{
		if err == CACHESTOP_SUC{
			cc.Start()
			fmt.Println("set")
		}
	}else{
		fmt.Println("set success")
	}
	gdata,err := cc.Get("redistest")
	if nil == err{
		if nil != gdata{
			var stu Student
			decoder := gob.NewDecoder(bytes.NewReader(gdata))
			decoder.Decode(&stu)
			fmt.Println("get success")
			fmt.Println(stu.Name,stu.Age,stu.Addr)
		}else{
			fmt.Println("get fail",err.Error())
		}
	}else{
		if err == CACHESTOP_SUC{
			cc.Start()
		}else{
			fmt.Println(err)
		}
	}
	cc.Exists("test25")
	fmt.Println("--------end test rediscache cache--------")
}


func BenchmarkRedisCache(b *testing.B){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test rediscache cache--------")
	cc,err := NewRedisCache(&CacheConfig{"tcp","127.0.0.1:6379",5,10,10})
	if nil != err{
		b.Errorf("new rediscache fail\n")
		return
	}
	err = cc.Start()
	if nil != err{
		b.Errorf(err.Error())
		return
	}else{
		fmt.Println("rediscache cache init succes")
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
				fmt.Println("set fail")
				if err == CACHESTOP_SUC{
					cc.Start()
				//	fmt.Println("set")
				}
			}else{
				//fmt.Println("set success")
			}
			gdata,err := cc.Get(name)
			if nil == err{
				if nil != gdata{
					var stu Student
					decoder := gob.NewDecoder(bytes.NewReader(gdata))
					decoder.Decode(&stu)
					//fmt.Println(stu.Name,stu.Age,stu.Addr)
				}else{
					fmt.Println("get fail")
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
				//	fmt.Println("delete")
				}
			}
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println("--------end test rediscache cache--------")
}