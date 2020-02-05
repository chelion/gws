package main

import(
	"fmt"
	"runtime"
	"github.com/chelion/gws"
	"github.com/chelion/gws/utils"
	"github.com/chelion/gws/configure"
	"github.com/chelion/gws/log"
)

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	config,err := configure.NewIniConfigure("./httpserver.ini")
	if nil != err{
		fmt.Println(err)
		return
	}
	config.Init()
	defer config.DeInit()
	logger,err := log.NewConsoleLog(true)
	if nil != err{
		fmt.Println(err)
		return
	}
	logger.Init()
	defer logger.DeInit()
	g,err := gws.New("http","HttpServer",config,logger)
	if nil != err{
		fmt.Println(err)
		return
	}
	g.GET("/",func(ctx *gws.Context){
		ctx.Data(200,"text/html; charset=utf-8",utils.String2Bytes("Hello, World!"))
	})
	g.UseMiddleWare(func (ctx *gws.Context){
		fmt.Println(string(ctx.Path()))
	})
	group1 := g.Group("/v1")
	{
		group1.StaticFile("/favicon.ico","./fav.ico")
		group1.StaticFS("/filefs","./filefs")
		group1.Static("/image","./image")
		group1.GET("/",func(ctx *gws.Context){
			ctx.Data(200,"text/html; charset=utf-8",utils.String2Bytes("Group1,Hello, World!"))
		})
	}
	
	group2 := g.Group("/v2")
	{
		group2.GET("/",func(ctx *gws.Context){
			ctx.Data(200,"text/html; charset=utf-8",utils.String2Bytes("Group2,Hello, World!"))
		})
	}
	
	g.Start()
	defer g.Stop()
}