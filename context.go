package gws
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"math"
	"errors"
	"mime/multipart"
	"github.com/chelion/gws/fasthttp"
	"github.com/chelion/gws/render"
	"github.com/chelion/gws/session"
)

type ContextHandler func(ctx *Context)

type Context struct{
	Rctx 		*fasthttp.RequestCtx
	Server 		*GWS
	session 	session.SessionStore
	get 		*fasthttp.Args
	post 		*fasthttp.Args
	index    	int8
	handlers	[]ContextHandler
}

var(
	SESSION_NIL = errors.New("session is nil")
)
const(
	abortIndex int8 = math.MaxInt8 / 2
)

func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < int8(len(ctx.handlers)) {
		if nil != ctx.handlers[ctx.index]{
			ctx.handlers[ctx.index](ctx)
		}
		ctx.index++
	}
}

func (ctx *Context) IsAborted() bool {
	return ctx.index >= abortIndex
}

func (ctx *Context) Abort() {
	ctx.index = abortIndex
}

func(ctx *Context)SessionStart()error{
	var err error
	if nil != ctx.Server.Session{
		ctx.session,err = ctx.Server.Session.Start(ctx.Rctx)
		if nil != err{
			ctx.Server.Logger.Println(err)
		}
		return err
	}
	return SESSION_NIL
}

func (ctx *Context)SessionGetNum()int{
	if nil != ctx.Server.Session{
		return ctx.Server.Session.GetSessionNum()
	}
	return 0
}

func(ctx *Context)SessionDestory(){
	if nil != ctx.Server.Session{
		ctx.Server.Session.Destroy(ctx.Rctx)
		if nil != ctx.session{
			ctx.session = nil
		}
	}
}

func(ctx *Context)SessionSet(key string, value []byte){
	if nil != ctx.session{
		ctx.session.Set(key,value)
	}
}

func(ctx *Context)SessionGet(key string)([]byte){
	if nil != ctx.session{
		return ctx.session.Get(key)
	}
	return nil
}

func(ctx *Context)SessionDelete(key string){
	if nil != ctx.session{
		ctx.session.Delete(key)
	}
}

func (ctx *Context)SessionFlush(){
	if nil != ctx.session{
		ctx.session.Flush()
	}
}

func (ctx *Context)SessionGetId()(id string){
	if nil != ctx.session{
		return ctx.session.GetSessionId()
	}
	return ""
}

func (ctx *Context)SessionGetAll()(map[string][]byte){
	if nil != ctx.session{
		return ctx.session.GetAll()
	}
	return nil
}

func (ctx *Context)SessionSave(){
	if nil != ctx.session{
		err := ctx.session.Save()
		if nil != err{
			ctx.Server.Logger.Println(err)
		}
	}
}

func (ctx *Context)UserAgent()([]byte){
	return ctx.Rctx.UserAgent()
}

func (ctx *Context)RequestURI()([]byte){
	return ctx.Rctx.RequestURI()
}

func (ctx *Context)Referer()([]byte){
	if nil != ctx.Rctx{
		return ctx.Rctx.Referer()
	}
	return nil		
}

func (ctx *Context)Redirect(statusCode int,uri string){
	ctx.Rctx.Redirect(uri,statusCode)
}

func (ctx *Context)IsConnectClose()bool{
	return ctx.Rctx.IsConnectClose()
}

func (ctx *Context)Host()([]byte){
	return ctx.Rctx.Host()
}

func (ctx *Context)Method()([]byte){
	return ctx.Rctx.Method()
}

func (ctx *Context)Path()([]byte){
	return ctx.Rctx.Path()
}

func (ctx *Context)FormValue(key string)([]byte){
	mf, err := ctx.Rctx.MultipartForm()
	if err == nil && mf.Value != nil {
		vv := mf.Value[key]
		if len(vv) > 0 {
			return []byte(vv[0])
		}
	}
	return nil
}

func (ctx *Context)SendFile(path string){
	ctx.Rctx.SendFile(path)
}

func (ctx *Context)SendFileBytes(content []byte){
	ctx.Rctx.SendFileBytes(content)
}

func (ctx *Context)UserValue(key string)(interface{}){
	return ctx.Rctx.UserValue(key)
}

func (ctx *Context)Get(key string)([]byte){
	if nil == ctx.get{
		ctx.get = ctx.Rctx.QueryArgs()
	}
	return ctx.get.Peek(key)
}

func (ctx *Context)Post(key string)([]byte){
	if nil == ctx.post{
		ctx.post = ctx.Rctx.PostArgs()
	}
	return ctx.post.Peek(key)
}

func (ctx *Context)PostBody()([]byte){
	return ctx.Rctx.PostBody()
}

func (ctx *Context)MultipartForm()(*multipart.Form, error){
	return ctx.Rctx.MultipartForm()
}

func (ctx *Context)FormFile(key string) (*multipart.FileHeader, error) {
	return ctx.Rctx.FormFile(key)
}

func (ctx *Context)SetContentType(value string){
	ctx.Rctx.SetContentType(value)
}

func (ctx *Context) SetStatusCode(code int) {
	ctx.Rctx.Response.SetStatusCode(code)
}

func (ctx *Context)Render(code int, r render.Render) {
	ctx.Rctx.Response.SetStatusCode(code)
	if err := r.Render(ctx.Rctx); err != nil {
		ctx.Server.Logger.Println(err)
	}
}

func (ctx *Context) HTML(code int, name string, obj interface{}) {
	instance := ctx.Server.HTMLRender.Instance(name, obj)
	ctx.Render(code, instance)
}


func (ctx *Context) IndentedJSON(code int, obj interface{}) {
	ctx.Render(code, render.IndentedJSON{Data: obj})
}


func (ctx *Context) SecureJSON(code int, obj interface{}) {
	ctx.Render(code, render.SecureJSON{Prefix: ctx.Server.secureJsonPrefix, Data: obj})
}

func (ctx *Context) JSONP(code int, obj interface{}) {
	callback := make([]byte,0)
	if nil == ctx.get{
		ctx.get = ctx.Rctx.QueryArgs()
	}
	callback = ctx.get.Peek("callback")
	if callback == nil {
		ctx.Render(code, render.JSON{Data: obj})
		return
	}
	ctx.Render(code, render.JsonpJSON{Callback: string(callback), Data: obj})
}

func (ctx *Context) JSON(code int, obj interface{}) {
	ctx.Render(code, render.JSON{Data: obj})
}

func (ctx *Context) AsciiJSON(code int, obj interface{}) {
	ctx.Render(code, render.AsciiJSON{Data: obj})
}

func (ctx *Context) PureJSON(code int, obj interface{}) {
	ctx.Render(code, render.PureJSON{Data: obj})
}

func (ctx *Context) XML(code int, obj interface{}) {
	ctx.Render(code, render.XML{Data: obj})
}

func (ctx *Context) YAML(code int, obj interface{}) {
	ctx.Render(code, render.YAML{Data: obj})
}

func (ctx *Context) ProtoBuf(code int, obj interface{}) {
	ctx.Render(code, render.ProtoBuf{Data: obj})
}

func (ctx *Context) String(code int, format string, values ...interface{}) {
	ctx.Render(code, render.String{Format: format, Data: values})
}

func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{ContentType: contentType,Data:data})
}