package cache

import(
	"fmt"
	"sync"
	"bytes"
	"time"
	"strconv"
	"testing"
	"encoding/gob"
	"github.com/google/uuid"
	"github.com/chelion/gws/utils"
)

func BenchmarkLocalCache(b *testing.B){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start benchmark local cache--------")
	cc,err := NewLocalCache(&LocalCacheConfig{512*1024*1024})
	if nil != err{
		b.Errorf("new local fail\n")
	}
	err = cc.Start()
	if nil != err{
		b.Errorf(err.Error())
	}else{
		fmt.Println("local cache init succes")
	}
	defer cc.Stop()
	wg := sync.WaitGroup{}
	b.N = 1000 * 1000
	wg.Add(1000)
	for i:=0;i<1000;i++{
		go func(i int){
			for j:=0;j<1000;j++{
				
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
	fmt.Println("--------end test LocalCache cache--------")
}
/*
func TestLocalCache(t *testing.T){
	type Student struct{
		Name string
		Age int
		Addr string
	}
	fmt.Println("--------start test local cache--------")
	cc,err := NewLocalCache(&LocalCacheConfig{256*1024*1024})
	if nil != err{
		t.Errorf("new local fail\n")
	}
	err = cc.Start()
	if nil != err{
		t.Errorf(err.Error())
	}else{
		fmt.Println("local cache init succes")
	}
	defer cc.Stop()
	student := &Student{Name:"yilin",Age:25,Addr:"shenzhen,china"}
	var result bytes.Buffer
    encoder := gob.NewEncoder(&result)
    encoder.Encode(student)
    userBytes := result.Bytes()
	err = cc.Set("hello",userBytes,0)
	if nil != err{
		t.Errorf(err.Error())
	}else{
		fmt.Println("set success")
	}
	gdata,err := cc.Get("hello")
	if nil == err{
		fmt.Print("get success->")
		var stu Student
		decoder := gob.NewDecoder(bytes.NewReader(gdata))
		decoder.Decode(&stu)
		fmt.Println(stu.Name,stu.Age,stu.Addr)
	}else{
		t.Errorf(err.Error())
	}
	sta,e0 := cc.Delete("hello")
	if nil != e0{
		t.Errorf(e0.Error())
	}else{
		if true == sta{
			fmt.Println("delete success")
		}
	}
	sta,e := cc.Exists("hello")
	if nil != e{
		t.Errorf(e.Error())
	}else{
		if true == sta{
			fmt.Println("exists")
		}else{
			fmt.Println("not exists")
		}
	}
	fmt.Println("--------end test memcache cache--------")
}
*/

func GenerateUid(userUid string)string{
	uuidv,_ := uuid.NewUUID()
	serviceEnd := utils.MD5("publiciot_"+strconv.FormatInt(time.Now().UnixNano(),10)+uuidv.String()+userUid+"_service")[8:24]
	return "S000-09-"+serviceEnd
}

func generateToken()([]byte){
	uuidv,_ := uuid.NewUUID()
	token := utils.MD5("publiciot_"+uuidv.String()+"_token")
	return utils.String2Bytes(token[8:24])
}
func BenchmarkLocalCacheM(b *testing.B){
	cc,err := NewLocalCache(&LocalCacheConfig{4096*1024*1024})
	if nil != err{
		b.Errorf("new local fail\n")
	}
	err = cc.Start()
	if nil != err{
		b.Errorf(err.Error())
	}else{
		fmt.Println("local cache init succes")
	}
	defer cc.Stop()
	b.N = 2000000
	for i := 0;i<2000000;i++{
		userUid := GenerateUid(strconv.FormatInt(int64(i),10))
		cc.Set(userUid,generateToken(),0)
	}
}