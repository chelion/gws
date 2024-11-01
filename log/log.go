package log

// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

type LevelEnum int

var (
	ILOGCLIENT_NIL = errors.New("log client is nil")
)

const (
	ALL    LevelEnum = iota //0
	DEBUG                   //1
	INFO                    //2
	WARN                    //3
	ERROR                   //4
	OFF                     //5
	UNKNOW                  //6
)

type Log interface {
	Init() (err error)   //初始化启动服务
	DeInit() (err error) //停用服务
	Printf(format string, v ...interface{})
	Print(v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	SetLevel(level LevelEnum)
	GetLevel() (err error, level LevelEnum)
}

type LogHandler struct {
	exit  bool
	msgs  chan string
	log   *log.Logger
	level LevelEnum
	date  bool
}

func (lh *LogHandler) getDate() (date string) {
	if true == lh.date {
		t := time.Now()
		date = fmt.Sprintf("\r\nDate:%s\r\n", t.Format("2006-01-02 15:04:05.999Z07:00"))
		return
	}
	return ""
}

func (lh *LogHandler) logout() {
	for {
		msg, ok := <-lh.msgs
		if ok {
			if nil != lh.log {
				lh.log.Output(2, msg)
			}
		}
		if true == lh.exit {
			return
		}
	}
}

func (lh *LogHandler) Init() (err error) {
	if nil == lh.log {
		return ILOGCLIENT_NIL
	}
	if nil == lh.msgs {
		lh.exit = false
		lh.msgs = make(chan string, 1024)
		go lh.logout()
	}
	return nil
}

func (lh *LogHandler) DeInit() (err error) {
	if nil != lh.log {
		lh.exit = true
		lh.log = nil
		return nil
	}
	return ILOGCLIENT_NIL
}

func (lh *LogHandler) Printf(format string, v ...interface{}) {
	if nil != lh.log {
		out := lh.getDate() + fmt.Sprintf(format,v...)
		lh.msgs <- out
	}
}

func (lh *LogHandler) Print(v ...interface{}) {
	if nil != lh.log {
		out := lh.getDate() + fmt.Sprint(v...)
		lh.msgs <- out
	}
}

func (lh *LogHandler) Println(v ...interface{}) {
	if nil != lh.log {
		out := lh.getDate() + fmt.Sprint(v...) + "\n"
		lh.msgs <- out
	}
}

func (lh *LogHandler) Fatal(v ...interface{}) {
	if nil != lh.log {
		out := lh.getDate() + fmt.Sprint(v...)
		lh.msgs <- out
		os.Exit(1)
	}
}

func (lh *LogHandler) Fatalf(format string, v ...interface{}) {
	if nil != lh.log {
		out := lh.getDate() + fmt.Sprintf(format,v...)
		lh.msgs <- out
		os.Exit(1)
	}
}

func (lh *LogHandler) Fatalln(v ...interface{}) {
	if nil != lh.log {
		out := lh.getDate() + fmt.Sprint(v...)
		lh.msgs <- out
		os.Exit(1)
	}
}

func (lh *LogHandler) Debug(v ...interface{}) {
	if lh.level <= DEBUG && nil != lh.log {
		out := lh.getDate() + "Debug->" + fmt.Sprint(v...)
		lh.msgs <- out
	}
}

func (lh *LogHandler) Info(v ...interface{}) {
	if lh.level <= INFO && nil != lh.log {
		out := lh.getDate() + "Info->" + fmt.Sprint(v...)
		lh.msgs <- out
	}
}

func (lh *LogHandler) Warn(v ...interface{}) {
	if lh.level <= WARN && nil != lh.log {
		out := lh.getDate() + "Warn->" + fmt.Sprint(v...)
		lh.msgs <- out
	}
}

func (lh *LogHandler) Error(v ...interface{}) {
	if lh.level <= ERROR && nil != lh.log {
		out := lh.getDate() + "Error->" + fmt.Sprint(v...)
		lh.msgs <- out
	}
}

func (lh *LogHandler) SetLevel(level LevelEnum) {
	lh.level = level
}

func (lh *LogHandler) GetLevel() (err error, level LevelEnum) {
	if nil != lh.log {
		return nil, lh.level
	}
	return ILOGCLIENT_NIL, UNKNOW
}
