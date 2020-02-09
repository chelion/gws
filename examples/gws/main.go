package main

import(
	"fmt"
	"runtime"
	"github.com/chelion/gws/configure"
	"github.com/chelion/gws/log"
	"github.com/chelion/gws"
)

func main(){
	var requestcnt uint64
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
	g.UseDefaultSession()
	g.GET("/",func(ctx *gws.Context){
		ctx.Rctx.Write([]byte("gws is ok!"))
	})
	g.LoadHTMLGlob("html/**/*")
	g.UseMiddleWare(func (ctx *gws.Context){
		requestcnt++
	})
	g.NotFound(func (ctx *gws.Context){
		ctx.String(404,"404 not find,is my 404")
	})
	g.NotAllowed(func(ctx *gws.Context){
		ctx.String(403,"403 not allow,is my 403")
	})
	g.GET("/statistics",func(ctx *gws.Context){
		ctx.String(200,"request count %d",requestcnt)
	})
	g.GET("/sessionnum",func(ctx *gws.Context){
		ctx.String(200,"session number %d",ctx.SessionGetNum())
	})
	g.GET("/setsession",func(ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			if nil == ctx.SessionStart(){
				ctx.SessionSet("Name",name)
				ctx.SessionSet("Age",[]byte("20"))
				ctx.SessionSave()
				ctx.String(200,"set session success!")
				return
			}
		}
		ctx.String(200,"set session fail!")
	})
	g.GET("/getsession",func(ctx *gws.Context){
			ctx.SessionStart()
			data := ctx.SessionGetAll()
			if nil != data{
				ctx.String(200,"get session success,=>"+string(data["Name"]))
			}else{
				ctx.String(200,"get session fail")
			}
			ctx.SessionSave()
	})
	g.GET("/text",func (ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			ctx.String(200,"hello,%s!",name)
		}else{
			ctx.String(200,"hello")
		}
	})
	g.GET("/zh",func (ctx *gws.Context){
		ctx.HTML(200, "zh/index.tmpl",nil)
	})
	g.GET("/en",func (ctx *gws.Context){
		ctx.HTML(200, "en/index.tmpl",nil)
	})
	g.GET("/json",func (ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			ctx.JSON(200,gws.M{
				"Name": string(name),
			})
		}else{
			ctx.JSON(200,gws.M{
				"Name": "Unknow",
			})	
		}
	})
	g.GET("/purejson",func (ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			ctx.PureJSON(200,gws.M{
				"Name": string(name),
			})
		}else{
			ctx.PureJSON(200,gws.M{
				"Name": "Unknow",
			})	
		}
	})
	g.GET("/xml",func (ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			ctx.XML(200,gws.M{
				"Name": string(name),
			})
		}else{
			ctx.XML(200,gws.M{
				"Name": "Unknow",
			})	
		}
	})
	g.GET("/yaml",func (ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			ctx.YAML(200,gws.M{
				"Name": string(name),
			})
		}else{
			ctx.YAML(200,gws.M{
				"Name": "Unknow",
			})	
		}
	})
	g.GET("/securejson",func (ctx *gws.Context){
		if nil != ctx.Get("name"){
			name := ctx.Get("name")
			ctx.SecureJSON(200,gws.M{
				"Name": string(name),
			})
		}else{
			ctx.SecureJSON(200,gws.M{
				"Name": "Unknow",
			})	
		}
	})
	g.StaticFile("/favicon.ico","./fav.ico")
	g.StaticFS("/filefs","./filefs")
	g.Static("/image","./image")
	g.Start()
	defer g.Stop()
}