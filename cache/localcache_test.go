package cache

import(
	"fmt"
	"bytes"
	"testing"
	"encoding/gob"
	"github.com/google/uuid"
)

func BenchmarkLocalCache(b *testing.B){
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
	for i:=0;i<1000000;i++{
		id, _ := uuid.NewUUID()
		cc.Set(id.String(),[]byte("--------start benchmark local cache--------"),0)
	}
	b.N = 1000000
}

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
