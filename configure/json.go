package configure

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.


import(
	"os"
	"sync"
	"strings"
	"io/ioutil"
	"strconv"
	"encoding/json"
)

type JsonConfigure struct{
	ConfigureHandler
}

func (cfg *JsonConfigure)loadFile()(err error){
	if nil != cfg.file{
		cfg.hmutex.Lock()
		cfg.content,err = ioutil.ReadAll(cfg.file)
		cfg.hmutex.Unlock()
		if nil != err{
			return err
		}
		return nil
	}
	return FILE_NIL
}



func (cfg *JsonConfigure)parseMapData(sectionName string,m map[string]interface{}){
	if "" == sectionName{
		sectionName = DEFAULT_SECTION
	}else{
		if _,ok := cfg.configData[sectionName];!ok{
			cfg.configData[sectionName] = make(map[string]interface{})
			cfg.sectionsName = append(cfg.sectionsName,sectionName)
		}
	}
	for k, v := range m {
		k = strings.TrimSpace(k)
		k = strings.ToLower(k)
		switch v.(type) {
			default:{
				cfg.configData[sectionName][k] = v
			}
			case map[string]interface{}:{
				var data map[string]interface{}
				data = v.(map[string]interface{})
				cfg.parseMapData(k,data)
			}
			case map[string]string:{
				var data map[string]string
				data = v.(map[string]string)
				sectionName = k
				if _,ok := cfg.configData[k];!ok{
					cfg.configData[k] = make(map[string]interface{})
					cfg.sectionsName = append(cfg.sectionsName,sectionName)
				}
				for mk,mv := range data{
					cfg.configData[sectionName][mk] = mv
				}
			}
		}
	}
	return
}

func (cfg *JsonConfigure)parseFile()(err error){
	var data map[string]interface{}
	defer cfg.hmutex.Unlock()
	cfg.hmutex.Lock()
	if _,ok := cfg.configData[DEFAULT_SECTION];!ok{
		cfg.configData[DEFAULT_SECTION] = make(map[string]interface{})
		cfg.sectionsName = append(cfg.sectionsName,DEFAULT_SECTION)
	}
    if err := json.Unmarshal(cfg.content, &data); err == nil {
		cfg.parseMapData("",data)
		return nil
    }
	return err
}

func NewJsonConfigure(filePath string)(cfg *JsonConfigure,err error){
	cfg = &JsonConfigure{}
	cfg.ConfigureHandler = ConfigureHandler{currentSectionName:DEFAULT_SECTION,configData:make(map[string]map[string]interface{}),file:nil,filePath:filePath,hmutex:new(sync.Mutex)}
	file, err := os.Open(cfg.filePath)
	if nil != err{
		return nil,err
	}
	cfg.file = file
	return cfg,nil
}

func (cfg *JsonConfigure)Init()(err error){
	err = cfg.loadFile()
	if nil != err{
		return
	}
	err = cfg.parseFile()
	return err
}

func (cfg *JsonConfigure)DeInit()(err error){
	if nil != cfg.file{
		cfg.hmutex.Lock()
		cfg.file.Close()
		cfg.content = nil
		cfg.hmutex.Unlock()
	}
	return nil
}

func (cfg *JsonConfigure)Reload()(err error){
	err = cfg.loadFile()
	if nil != err{
		return 
	}
	return cfg.parseFile()
}

func (cfg *JsonConfigure)findData(key string)(data interface{},err error){
	if "" == cfg.currentSectionName{
		return "",NOSECTION_ERR
	}
	if _,ok := cfg.configData[cfg.currentSectionName];ok{
		if _,ok := cfg.configData[cfg.currentSectionName][key];ok{
			return cfg.configData[cfg.currentSectionName][key],nil
		}
	}
	return "",NOKEY_ERR
}

func (cfg *JsonConfigure)GetInt(key string,defaultv int)(value int,err error){
	if key == "" {
		return defaultv,PARAM_ERR
	}
	v,e := cfg.findData(strings.ToLower(key))
	if v != nil && e == nil{
		if v, ok := v.(int); ok {
			return int(v), nil
		}
		if v, ok := v.(int64); ok {
			return int(v), nil
		}
		if v, ok := v.(float64); ok {
			return int(v), nil
		}
		return defaultv,TYPE_ERR
	}
	return defaultv,e
}

func (cfg *JsonConfigure)GetInt64(key string,defaultv int64)(value int64,err error){
	if key == "" {
		return defaultv,PARAM_ERR
	}
	v,e := cfg.findData(strings.ToLower(key))
	if v != nil && e == nil{
		if v, ok := v.(int); ok {
			return int64(v), nil
		}
		if v, ok := v.(int64); ok {
			return v, nil
		}
		return defaultv,TYPE_ERR
	}
	return defaultv,e
}

func (cfg *JsonConfigure)GetFloat(key string,defaultv float64)(value float64,err error){
	if key == "" {
		return defaultv,PARAM_ERR
	}
	v,e := cfg.findData(strings.ToLower(key))
	if v != nil && e == nil{
		if v, ok := v.(float64); ok {
			return v, nil
		}
		if v, ok := v.(string); ok {
			fv,e := strconv.ParseFloat(v,64)
			if e != nil{
				return defaultv,e
			}
			return fv,nil
		}
		return defaultv,TYPE_ERR
	}
	return defaultv,e
}

func (cfg *JsonConfigure)GetBool(key string,defaultv bool)(value bool,err error){
	if key == "" {
		return defaultv,PARAM_ERR
	}
	v,e := cfg.findData(strings.ToLower(key))
	if v != nil && e == nil{
		if v, ok := v.(bool); ok {
			return v, nil
		}
		if v, ok := v.(string); ok {
			bv,e := strconv.ParseBool(v)
			if e != nil{
				return defaultv,e
			}
			return bv,nil
		}
		return defaultv,TYPE_ERR
	}
	return defaultv,e
}

func (cfg *JsonConfigure)GetString(key string)(value string,err error){
	if key == "" {
		return "",PARAM_ERR
	}
	v,e := cfg.findData(strings.ToLower(key))
	if v != nil && e == nil{
		if sv, ok := v.(string); ok {
			return sv, nil
		}
		return "",TYPE_ERR
	}
	return "",e
}