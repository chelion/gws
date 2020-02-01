package log
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"os"
	"errors"
	"log"
	"time"
	"fmt"
)

type LevelEnum int

var(
	ILOGCLIENT_NIL = errors.New("log client is nil")
)

const (
	ALL LevelEnum = iota//0
	DEBUG				//1
	INFO				//2
	WARN				//3
	ERROR				//4
	OFF					//5
	UNKNOW				//6
)

type Log interface{
	Init()(err error)//初始化启动服务
	DeInit()(err error)//停用服务
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
	GetLevel()(err error,level LevelEnum)
}


type LogHandler struct{
	log *log.Logger
	level LevelEnum
	date bool
}

func (lh *LogHandler)getDate()(date string){
	if true == lh.date{
		t := time.Now()
		date = fmt.Sprintf("\r\nDate:%s\r\n",t.Format("2006-01-02 15:04:05.999Z07:00"))
		return
	}
	return ""
}

func (lh *LogHandler)Init()(err error){
	if nil == lh.log{
		return ILOGCLIENT_NIL
	}
	return nil
}

func (lh *LogHandler)DeInit()(err error){
	if nil != lh.log{
		lh.log = nil
		return nil
	}
	return ILOGCLIENT_NIL
}

func (lh *LogHandler)Printf(format string, v ...interface{}){
	if nil != lh.log{
		lh.log.Print(lh.getDate())
		lh.log.Printf(format,v...)
	}
}

func (lh *LogHandler)Print(v ...interface{}){
	if nil != lh.log{
		lh.log.Print(lh.getDate())
		lh.log.Print(v...)
	}
}

func (lh *LogHandler)Println(v ...interface{}){
	if nil != lh.log{
		lh.log.Print(lh.getDate())
		lh.log.Println(v...)
	}
}

func (lh *LogHandler)Fatal(v ...interface{}){
	if nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprint(v...))
		os.Exit(1)
	}
}

func (lh *LogHandler)Fatalf(format string, v ...interface{}){
	if nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprintf(format,v...))
		os.Exit(1)
	}
}

func (lh *LogHandler)Fatalln(v ...interface{}){
	if nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprintln(v...))
		os.Exit(1)
	}
}

func (lh *LogHandler)Debug(v ...interface{}){
	if lh.level <= DEBUG && nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprintln("Debug->", v))
	}
}

func (lh *LogHandler)Info(v ...interface{}){
	if lh.level <= INFO && nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprintln("Info->", v))
	}
}

func (lh *LogHandler)Warn(v ...interface{}){
	if lh.level <= WARN && nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprintln("Warn->", v))
	}
}

func (lh *LogHandler)Error(v ...interface{}){
	if lh.level <= ERROR && nil != lh.log{
		lh.log.Output(2,lh.getDate())
		lh.log.Output(2, fmt.Sprintln("Error->", v))
	}
}

func (lh *LogHandler)SetLevel(level LevelEnum){
	lh.level = level
}

func (lh *LogHandler)GetLevel()(err error,level LevelEnum){
	if nil != lh.log{
		return nil,lh.level
	}
	return ILOGCLIENT_NIL,UNKNOW
}