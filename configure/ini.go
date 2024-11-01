package configure

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

type IniConfigure struct {
	ConfigureHandler
}

func (cfg *IniConfigure) loadFile() (err error) {
	if nil != cfg.file {
		cfg.hmutex.Lock()
		cfg.content, err = ioutil.ReadAll(cfg.file)
		cfg.hmutex.Unlock()
		if nil != err {
			return err
		}
		return nil
	}
	return FILE_NIL
}

func (cfg *IniConfigure) parseFile() (err error) {
	var linecount int
	var sectionName string = DEFAULT_SECTION
	defer cfg.hmutex.Unlock()
	cfg.hmutex.Lock()
	buf := bufio.NewReader(bytes.NewBuffer(cfg.content))
	mask, err := buf.Peek(3)
	if err == nil && len(mask) >= 3 &&
		mask[0] == 239 && mask[1] == 187 && mask[2] == 191 {
		buf.Read(mask)
	}
	if _, ok := cfg.configData[sectionName]; !ok {
		cfg.configData[sectionName] = make(map[string]interface{})
		cfg.sectionsName = append(cfg.sectionsName, sectionName)
	}
	for {
		linecount++
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\r")
		linelen := len(line)
		if err != nil {
			if err != io.EOF {
				return err
			}
			if 0 == linelen {
				break
			}
		}
		switch {
		case linelen == 0:
			{ //行长度为0
				continue
			}
		case line[0] == '#' || line[0] == ';':
			{ //注释
				continue
			}
		case line[0] == '[' && line[linelen-1] == ']':
			{ //新的Section
				sectionName = strings.TrimSpace(line[1 : linelen-1])
				if len(sectionName) == 0 || "" == sectionName {
					sectionName = DEFAULT_SECTION
				}
				if _, ok := cfg.configData[sectionName]; !ok {
					cfg.configData[sectionName] = make(map[string]interface{})
					cfg.sectionsName = append(cfg.sectionsName, sectionName)
				}
				continue
			}
		default:
			{
				if linelen <= 2 || !strings.Contains(line, "=") {
					continue
				}
				lineinfo := bytes.SplitN([]byte(line), []byte("="), 2)
				key := bytes.Trim(lineinfo[0], " ")
				value := bytes.Trim(lineinfo[1], " ")
				keylen := len(key)
				valuelen := len(value)
				if 0 == keylen || 0 == valuelen {
					return errors.New("parse error on line:" + strconv.Itoa(linecount))
				}
				keyquote := ""
				if key[0] == byte('"') && key[keylen-1] == byte('"') {
					if keylen >= 6 && string(key[0:3]) == `"""` && string(key[(keylen-3):]) == `"""` {
						keyquote = `"""`
					} else {
						keyquote = `"`
					}
				} else if key[0] == byte('`') && key[keylen-1] == byte('`') {
					keyquote = "`"
				}
				if keyquote != "" {
					lenkq := len(keyquote)
					key = key[lenkq:(keylen - lenkq)]
				}
				valuequote := ""
				if value[0] == byte('"') && value[valuelen-1] == byte('"') {
					if valuelen >= 6 && string(value[0:3]) == `"""` && string(value[(valuelen-3):]) == `"""` {
						valuequote = `"""`
					} else {
						valuequote = `"`
					}
				} else if value[0] == byte('`') && value[valuelen-1] == byte('`') {
					valuequote = "`"
				}

				if valuequote != "" {
					lenvq := len(valuequote)
					value = value[lenvq:(valuelen - lenvq)]
				}
				if 0 == len(key) {
					return errors.New("parse error on line:" + strconv.Itoa(linecount))
				}
				keystr := strings.ToLower(string(key))
				keystr = strings.Replace(keystr, " ", "", -1)
				cfg.configData[sectionName][keystr] = string(value)
			}
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

func NewIniConfigure(filePath string) (cfg *IniConfigure, err error) {
	cfg = &IniConfigure{}
	cfg.ConfigureHandler = ConfigureHandler{currentSectionName: DEFAULT_SECTION, configData: make(map[string]map[string]interface{}), file: nil, filePath: filePath, hmutex: new(sync.Mutex), sectionsName: make([]string, 1)}
	file, err := os.Open(cfg.filePath)
	if nil != err {
		return nil, err
	}
	cfg.file = file
	return cfg, nil
}

func (cfg *IniConfigure) Init() (err error) {
	err = cfg.loadFile()
	if nil != err {
		return
	}
	err = cfg.parseFile()
	return err
}

func (cfg *IniConfigure) DeInit() (err error) {
	if nil != cfg.file {
		cfg.hmutex.Lock()
		cfg.file.Close()
		cfg.content = nil
		cfg.hmutex.Unlock()
	}
	return nil
}

func (cfg *IniConfigure) Reload() (err error) {
	err = cfg.loadFile()
	if nil != err {
		return
	}
	return cfg.parseFile()
}

func (cfg *IniConfigure) findData(key string) (data string, err error) {
	if "" == cfg.currentSectionName {
		return "", NOSECTION_ERR
	}
	fmt.Println("find Data,sectionName", cfg.currentSectionName)
	if _, ok := cfg.configData[cfg.currentSectionName]; ok {
		if _, ok := cfg.configData[cfg.currentSectionName][key]; ok {
			return cfg.configData[cfg.currentSectionName][key].(string), nil
		}
	}
	return "", errors.New("this key is not exists error->" + key)
}

func (cfg *IniConfigure) GetInt(key string, defaultv int) (value int, err error) {
	if key == "" {
		return defaultv, PARAM_ERR
	}
	vstr, e := cfg.findData(strings.ToLower(key))
	if nil == e {
		v, e := strconv.ParseInt(vstr, 10, 64)
		if e != nil {
			return defaultv, e
		}
		return int(v), nil
	}
	return defaultv, e
}

func (cfg *IniConfigure) GetInt64(key string, defaultv int64) (value int64, err error) {
	if key == "" {
		return defaultv, PARAM_ERR
	}
	vstr, e := cfg.findData(strings.ToLower(key))
	if nil == e {
		v, e := strconv.ParseInt(vstr, 10, 64)
		if e != nil {
			return defaultv, e
		}
		return v, nil
	}
	return defaultv, e
}

func (cfg *IniConfigure) GetFloat(key string, defaultv float64) (value float64, err error) {
	if key == "" {
		return defaultv, PARAM_ERR
	}
	vstr, e := cfg.findData(strings.ToLower(key))
	if nil == e {
		v, e := strconv.ParseFloat(vstr, 64)
		if e != nil {
			return defaultv, e
		}
		return v, nil
	}
	return defaultv, e
}

func (cfg *IniConfigure) GetBool(key string, defaultv bool) (value bool, err error) {
	if key == "" {
		return defaultv, PARAM_ERR
	}
	vstr, e := cfg.findData(strings.ToLower(key))
	if nil == e {
		v, e := strconv.ParseBool(strings.ToLower(vstr))
		if e != nil {
			return defaultv, e
		}
		return v, nil
	}
	return defaultv, e
}

func (cfg *IniConfigure) GetString(key string) (value string, err error) {
	if key == "" {
		return "", PARAM_ERR
	}
	vstr, e := cfg.findData(strings.ToLower(key))
	if nil == e {
		return vstr, nil
	}
	return "", e
}
