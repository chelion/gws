package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"os"
	slog"log"
)

type ConsoleLog struct{
	LogHandler
}


func NewConsoleLog(date bool)(consolelog *ConsoleLog,err error){
	log := slog.New(os.Stderr, "",0)
	log.SetFlags(slog.Lshortfile)
	consolelog = &ConsoleLog{}
	consolelog.LogHandler = LogHandler{log:log,level:DEBUG,date:date}
	return consolelog,nil
}

func (consolelog *ConsoleLog)Init()(err error){
	if nil == consolelog.log{
		return ILOGCLIENT_NIL
	}
	return nil
}

func (consolelog *ConsoleLog)DeInit()(err error){
	if nil != consolelog.log{
		consolelog.log = nil
		return nil
	}
	return ILOGCLIENT_NIL
}