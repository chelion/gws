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

func TestFastCache(t *testing.T){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test Fastcache cache--------")
	cc,err := NewFastCache(&CacheConfig{"tcp","127.0.0.1:11811",5,10,50})
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
		fmt.Println(err)
		if err == CACHESTOP_SUC{
			cc.Start()
		}else{
			fmt.Println(err)
		}
	}
	cc.Exists("test25")
	fmt.Println("--------end test Fastcache cache--------")
}


func BenchmarkFastCache(b *testing.B) {
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test FastCache cache--------")
	cc,err := NewFastCache(&CacheConfig{"tcp","127.0.0.1:11811",1,5,50})
	if nil != err{
		b.Errorf("new FastCache fail\n")
		return
	}
	err = cc.Start()
	if nil != err{
		b.Errorf(err.Error())
		return
	}else{
		fmt.Println("FastCache cache init succes")
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
				}
			}else{
				//fmt.Println("set success")
			}
			/*
			gdata,err := cc.Get(name)
			if nil == err{
				if nil != gdata{
					var stu Student
					decoder := gob.NewDecoder(bytes.NewReader(gdata))
					decoder.Decode(&stu)
					//fmt.Println(stu.Name,stu.Age,stu.Addr)
				}else{
					//fmt.Println("get fail")
				}
			}else{
				if err == CACHESTOP_SUC{
					//fmt.Println(err)
					cc.Start()
				}else{
				//	fmt.Println(err)
				}
			}
			
			sta,e0 := cc.Delete(name)
			if nil != e0{
				fmt.Println("delete fail")
				if e0 == CACHESTOP_SUC{
					cc.Start()
				}
			}else{
				if true == sta{
					//fmt.Println("delete")
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
			}*/
				
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println("--------end test FastCache cache--------")
}
