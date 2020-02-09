package gws
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"net"
	"sync"
	"time"
	"errors"
	"crypto/tls"
	"html/template"
	"encoding/xml"
	"github.com/chelion/gws/fasthttp"
	"github.com/chelion/gws/log"
	"github.com/chelion/gws/configure"
	"github.com/chelion/gws/render"
	"github.com/chelion/gws/cache"
	"github.com/chelion/gws/session"
	"github.com/chelion/gws/session/memory"
)

var(
	PARAM_NIL = errors.New("param is nil")
)

type GWS struct{
	once 			sync.Once
	isDebug 		bool
	isHttps 		bool
	netWork 		string
	addr 			string
	listener 		net.Listener
	certFilePath 	string
	keyFilePath 	string
	delims			render.Delims
	secureJsonPrefix string
	Config 			configure.Configure
	Logger 			log.Log
	Session 		*session.Session
	router 			*Router
	RouterGroup 	*RouterGroup
	HTMLRender		render.HTMLRender
	FuncMap			template.FuncMap
	ServerName 		string
}

type M map[string]interface{}

// MarshalXML allows type M to be used with xml.Marshal.
func (v M) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range v {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func newHttpServer(config configure.Configure,logger log.Log,configSectionName string)(group *RouterGroup,err error){
	if nil == config || nil == logger{
		return nil,PARAM_NIL
	}
	httpServer := &GWS{netWork:"tcp",addr:":80",isDebug:true,
	isHttps:false,FuncMap:template.FuncMap{},delims:render.Delims{Left: "{{", Right: "}}"},
	secureJsonPrefix:"while(1);",ServerName:"GWS",Session:nil}
	httpServer.router = NewRouter(httpServer)
	httpServer.Config = config
	httpServer.Logger = logger
	httpServer.Config.SetSection(configSectionName)
	httpServer.netWork,err = httpServer.Config.GetString("netWork")
	if nil != err{
		httpServer.Logger.Error(err)
		return nil,err
	}
	httpServer.Logger.Info("http server netWork:"+httpServer.netWork)
	httpServer.addr,err = httpServer.Config.GetString("addr")
	if nil != err{
		httpServer.Logger.Error(err)
		return nil,err
	}
	httpServer.Logger.Info("http server addr:",httpServer.addr)
	httpServer.RouterGroup = &RouterGroup{
		basePath:"/",
		isRoot : true,
		server :httpServer,
	}
	return httpServer.RouterGroup,nil
}

func newHttpsServer(config configure.Configure,logger log.Log,configSectionName string)(group *RouterGroup,err error){
	if nil == config || nil == logger{
		return nil,PARAM_NIL
	}
	httpsServer := &GWS{netWork:"tcp",addr:":443",isDebug:true,isHttps:true,
	certFilePath:"https.crt",keyFilePath:"https.key",FuncMap:template.FuncMap{},delims:render.Delims{Left: "{{", Right: "}}"},
	secureJsonPrefix:"while(1);",ServerName:"CLIOT",Session:nil}
	httpsServer.router = NewRouter(httpsServer)
	httpsServer.Config = config
	httpsServer.Logger = logger
	httpsServer.Config.SetSection(configSectionName)
	httpsServer.netWork,err = httpsServer.Config.GetString("netWork")
	if nil != err{
		httpsServer.Logger.Error(err)
		return
	}
	httpsServer.Logger.Info("https server netWork:"+httpsServer.netWork)
	httpsServer.addr,err = httpsServer.Config.GetString("addr")
	if nil != err{
		httpsServer.Logger.Error(err)
		return
	}
	httpsServer.certFilePath,err = httpsServer.Config.GetString("certFilePath")
	if nil != err{
		httpsServer.Logger.Error(err)
		return
	}
	httpsServer.Logger.Info("https server cert file path addr:",httpsServer.certFilePath)
	httpsServer.keyFilePath,err = httpsServer.Config.GetString("keyFilePath")
	if nil != err{
		httpsServer.Logger.Error(err)
		return
	}
	httpsServer.Logger.Info("https server addr:",httpsServer.addr)
	httpsServer.RouterGroup = &RouterGroup{
		basePath:"/",
		isRoot : true,
		server :httpsServer,
	}
	return httpsServer.RouterGroup,nil
}

func startHttp(httpServer *GWS)(err error){
	s := &fasthttp.Server{
		Handler: httpServer.router.Handler,
		Logger:httpServer.Logger,
	}
	s.SetServerName(httpServer.ServerName)
	listener, err := net.Listen(httpServer.netWork,httpServer.addr)
	if err != nil {
		httpServer.Logger.Error(err)
		return
	}
	httpServer.listener = NewGracefulListener(listener,5*time.Second)
	httpServer.Logger.Info("http server starting...")
	err = s.Serve(httpServer.listener)
	if nil != err{
		if nil != httpServer.Logger{
			httpServer.Logger.Error(err)
		}
		return
	}
	return nil
}

func startHttps(httpsServer *GWS)(err error){
	s := &fasthttp.Server{
		Handler: httpsServer.router.Handler,
		Logger:httpsServer.Logger,
	}
	s.SetServerName(httpsServer.ServerName)
	ln, err := net.Listen(httpsServer.netWork,httpsServer.addr)
	if err != nil {
		httpsServer.Logger.Error(err)
		return
	}
	cert, err := tls.LoadX509KeyPair(httpsServer.certFilePath, httpsServer.keyFilePath)
	if err != nil {
		httpsServer.Logger.Error(err)
		return
	}
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
	}
	listener := tls.NewListener(ln, tlsConfig)
	httpsServer.Logger.Info("https server starting...")
	httpsServer.listener = NewGracefulListener(listener,5*time.Second)
	err = s.Serve(listener)
	if nil != err{
		if nil != httpsServer.Logger{
			httpsServer.Logger.Error(err)
		}
		return
	}
	return nil
}

func New(protocol,configSectionName string,config configure.Configure,logger log.Log)(group *RouterGroup,err error){
	if protocol == "https"{
		return newHttpsServer(config,logger,configSectionName)
	}else{
		return newHttpServer(config,logger,configSectionName)
	}
}

func (group *RouterGroup)NotFound(handle ContextHandler){
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.router.NotFound = handle
}

func (group *RouterGroup)NotAllowed(handle ContextHandler){
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.router.MethodNotAllowed = handle
}

func (group *RouterGroup)UseMiddleWare(middleware ...ContextHandler)*RouterGroup {
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.router.MiddlewareHandlers = append(group.server.router.MiddlewareHandlers,middleware...)
	return group
}

func (group *RouterGroup)Start()(err error){
	if false == group.isRoot{
		panic("need use group by New")
	}
	if group.server.isHttps{
		return startHttps(group.server)
	}else{
		return startHttp(group.server)
	}
}

func (group *RouterGroup)Stop()(err error){
	if false == group.isRoot{
		panic("need use group by New")
	}
	if nil != group.server.listener{
		return group.server.listener.Close()
	}
	return PARAM_NIL
}

func (group *RouterGroup)EnableDebug(debug bool){
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.isDebug= debug
}

func (group *RouterGroup)SetFuncMap(funcMap template.FuncMap){
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.FuncMap= funcMap
}

func (group *RouterGroup)GetServerName(servername string)string{
	if false == group.isRoot{
		panic("need use group by New")
	}
	return group.server.ServerName
}

func (group *RouterGroup)UseDefaultSession()*RouterGroup{
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.once.Do(func(){
		localCache,_ := cache.NewLocalCache(&cache.LocalCacheConfig{Cachesize:512*1024*1024})
		group.server.Session = session.NewSession(session.NewDefaultConfig())
		memoryConfig := &memory.Config{CacheConfig:make([]*memory.MCacheConfig,0)}
		memoryConfig.CacheConfig = append(memoryConfig.CacheConfig,&memory.MCacheConfig{Cache:localCache,Addr:"localcache",VirtualNodeNum:16})
		group.server.Session.SetProvider("memory",memoryConfig)
	})
	return group
}

func (group *RouterGroup)SetServerSession(session *session.Session)*RouterGroup{
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.Session = session
	return group
}

func (group *RouterGroup)SetServerName(servername string)*RouterGroup{
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.ServerName = servername
	return group
}

func (group *RouterGroup)Delims(left, right string)*RouterGroup {
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.delims = render.Delims{Left: left, Right: right}
	return group
}

func (group *RouterGroup)SecureJsonPrefix(prefix string)*RouterGroup {
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.secureJsonPrefix = prefix
	return group
}

func (group *RouterGroup) LoadHTMLGlob(pattern string) {
	if false == group.isRoot{
		panic("need use group by New")
	}
	if group.server.isDebug {
		group.server.HTMLRender = render.HTMLDebug{Glob: pattern, FuncMap: group.server.FuncMap, Delims: group.server.delims}
		return
	}
	left := group.server.delims.Left
	right := group.server.delims.Right
	templ := template.Must(template.New("").Delims(left, right).Funcs(group.server.FuncMap).ParseGlob(pattern))
	group.SetHTMLTemplate(templ)
}

func (group *RouterGroup)LoadHTMLFiles(files ...string) {
	if false == group.isRoot{
		panic("need use group by New")
	}
	if group.server.isDebug {
		group.server.HTMLRender = render.HTMLDebug{Files: files, FuncMap: group.server.FuncMap, Delims: group.server.delims}
		return
	}
	templ := template.Must(template.New("").Delims(group.server.delims.Left, group.server.delims.Right).Funcs(group.server.FuncMap).ParseFiles(files...))
	group.SetHTMLTemplate(templ)
}

func (group *RouterGroup)SetHTMLTemplate(templ *template.Template) {
	if false == group.isRoot{
		panic("need use group by New")
	}
	group.server.HTMLRender = render.HTMLProduction{Template: templ.Funcs(group.server.FuncMap)}
}