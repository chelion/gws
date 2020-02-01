package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"fmt"
	"log"
	"os"
	"time"
	"sync"
)

type FileLog struct{
	LogHandler
	file *os.File
	filename string
	hmutex *sync.Mutex
	split bool
	exit chan bool
}

func NewFileLog(filename string,split bool,date bool)(filelog *FileLog,err error){
	filepath := filename
	if true == split{
		t := time.Now()
		date := fmt.Sprintf("%d-%d-%d-",t.Year(),t.Month(),t.Day())
		filepath = date + filepath
	}
	file, err := os.OpenFile(filepath+".log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if nil != err{
		return nil,err
	}
	log := log.New(file, "",0)
	filelog = &FileLog{split:split,filename:filename,hmutex:new(sync.Mutex),file:file,exit:make(chan bool,1)}
	filelog.LogHandler = LogHandler{log:log,level:DEBUG,date:date}
	if true == split{
		go filelog.autoSplitLog()
	}
	return filelog,nil
}

func (filelog *FileLog)Init()(err error){
	if nil == filelog.log{
		return ILOGCLIENT_NIL
	}
	return nil
}

func (filelog *FileLog)DeInit()(err error){
	if nil != filelog.file{
		filelog.file.Close()
		filelog.file = nil
	}
	if true == filelog.split{
		filelog.exit <- true
	}
	if nil != filelog.log{
		filelog.log = nil
		return nil
	}
	return ILOGCLIENT_NIL
}


func (filelog *FileLog)autoSplitLog(){
	now := time.Now()
	next := now.Add(time.Hour * 24)
	next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
	var t *time.Timer = time.NewTimer(next.Sub(now))
	for{
		select{
			case <-t.C:{
				date := fmt.Sprintf("%d-%d-%d-",next.Year(),next.Month(),next.Day())
				filepath := date + filelog.filename
				newfile, _ := os.OpenFile(filepath+".log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
				if nil != filelog.log{
					filelog.log.SetOutput(newfile)
					filelog.hmutex.Lock()
					filelog.file.Close()
					filelog.file = newfile
					filelog.hmutex.Unlock()
					now = time.Now()
					next = now.Add(time.Hour * 24)
					next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
					t = time.NewTimer(next.Sub(now))
				}else{
					return
				}
			}
			case <-filelog.exit:{
				return
			}
		}
	}
}