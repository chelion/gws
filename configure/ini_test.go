package configure

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.


import(
	"fmt"
	"testing"
)

func TestIniConfigure(t *testing.T){
	fmt.Println("start ini configure test")
	cfg,err := NewIniConfigure("./test.ini")
	if nil != err{
		t.Errorf(err.Error())
	}
	err = cfg.Init()
	if nil != err{
		t.Errorf(err.Error())
	}
	defer cfg.DeInit()
	sectionsname := cfg.GetSectionsName()
	fmt.Println(sectionsname)
	sectiondata,e := cfg.GetSection("debug")
	if nil == e{
		fmt.Println(sectiondata)
	}
	sectiondata,e = cfg.GetSection("release")
	if nil == e{
		fmt.Println(sectiondata)
	}
	cfg.SetSection("release")
	addr,e := cfg.GetString("addr")
	if e != nil{
		t.Errorf(err.Error())
	}else{
		fmt.Println(addr)
	}
	port,e := cfg.GetInt("port",0)
	if e != nil{
		t.Errorf(err.Error())
	}else{
		fmt.Println(port)
	}
	uselink,e := cfg.GetBool("uselink",true)
	if e != nil{
		fmt.Println(e.Error())
	}else{
		fmt.Println(uselink)
	}
	fix,e := cfg.GetFloat("fix",0.0)
	if e != nil{
		t.Errorf(err.Error())
	}else{
		fmt.Println(fix)
	}
	fmt.Println("end ini configure test")
}