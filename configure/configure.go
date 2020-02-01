package configure

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.


import(
	"os"
	"sync"
	"errors"
)
const(
	DEFAULT_SECTION = "DEFAULT"
)
var(
	NOSECTION_ERR = errors.New("no section error")
	PARAM_ERR = errors.New("param error")
	TYPE_ERR = errors.New("type is not you want")
	NOKEY_ERR = errors.New("this key is not exists error")
	PARSE_ERR = errors.New("configure parse error")
	FILE_NIL = errors.New("configure file is nil")
)

type Configure interface{
	Init()(err error)
	DeInit()(err error)
	Reload()(err error)
	SetSection(sectionName string)
	GetSection(sectionName string)(value map[string]interface{},err error)
	GetSectionsName()(sectionsName []string)
	GetInt(key string,defaultv int)(value int,err error)
	GetInt64(key string,defaultv int64)(value int64,err error)
	GetFloat(key string,defaultv float64)(value float64,err error)
	GetBool(key string,defaultv bool)(value bool,err error)
	GetString(key string)(value string,err error)
}

type ConfigureHandler struct{
	content	[]byte
	currentSectionName string
	sectionsName []string
	configData  map[string]map[string]interface{} 
	file *os.File
	filePath string
	hmutex *sync.Mutex
}

func (cfg *ConfigureHandler)SetSection(sectionName string){
	cfg.hmutex.Lock()
	cfg.currentSectionName = sectionName
	cfg.hmutex.Unlock()
}

func (cfg *ConfigureHandler)GetSection(sectionName string)(value map[string]interface{},err error){
	if _,ok := cfg.configData[sectionName];ok{
		return cfg.configData[sectionName],nil
	}
	return nil,NOSECTION_ERR
}

func (cfg *ConfigureHandler)GetSectionsName()(sectionsName[] string){
	return cfg.sectionsName
}
