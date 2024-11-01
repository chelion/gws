package main

import (
	"fmt"
	"github.com/chelion/gws"
	"github.com/chelion/gws/configure"
	"github.com/chelion/gws/log"
	"runtime"
)

func CloudServer(ctx *gws.Context) {
	fmt.Println("here 1")
	fmt.Println(string(ctx.PostBody()))
	ctx.String(200, "{\"Status\":1}")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config, err := configure.NewIniConfigure("./http.ini")
	if nil != err {
		fmt.Println(err)
		return
	}
	config.Init()
	defer config.DeInit()
	logger, err := log.NewConsoleLog(true)
	if nil != err {
		fmt.Println(err)
		return
	}
	logger.Init()
	defer logger.DeInit()
	g, err := gws.New("http", "HttpServer", config, logger)
	if nil != err {
		fmt.Println(err)
		return
	}
	g.POST("/CloudServer", CloudServer)
	g.Start()
	defer g.Stop()
}
