package gws
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"strings"
	"github.com/chelion/gws/fasthttp"
)

type RouterGroup struct{
	basePath string
	isRoot bool
	server *GWS
}

func (group *RouterGroup)Group(basePath string)(* RouterGroup){
	if false == group.isRoot{
		panic("need use group by New")
	}
	return &RouterGroup{
		basePath: basePath,
		isRoot:false,
		server:group.server}
}

func (group *RouterGroup)GET(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.GET(path,handle)
}

func (group *RouterGroup)HEAD(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.HEAD(path,handle)
}

func (group *RouterGroup)OPTIONS(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.OPTIONS(path,handle)
}

func (group *RouterGroup)POST(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.POST(path,handle)
}

func (group *RouterGroup)PUT(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.PUT(path,handle)
}

func (group *RouterGroup)PATCH(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.PATCH(path,handle)
}

func (group *RouterGroup)DELETE(path string, handle ContextHandler) {
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.DELETE(path,handle)
}

func (group *RouterGroup)StaticFS(path string,rootpath string){
	if path == ""{
		panic("StaticFS path can not be empty")
	}
	plen := len(path)
	if false == group.isRoot{
		path = group.basePath+path
	}
	if path[plen-1] == '/'{
		path += "*filepath"
	}else{
		path += "/*filepath"
	}
	plen = len(path)
	if plen < 10 || path[plen-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}
	prefix := path[:plen-10]
	fileHandler := fasthttp.FSHandler(rootpath, strings.Count(prefix, "/"))
	group.server.router.GET(path, func(ctx *Context) {
		fileHandler(ctx.Rctx)
	})
}

func (group *RouterGroup)Static(path string,rootpath string){
	if path == ""{
		panic("Static path can not be empty")
	}
	plen := len(path)
	if false == group.isRoot{
		path = group.basePath+path
	}
	if path[plen-1] == '/'{
		path += "*filepath"
	}else{
		path += "/*filepath"
	}
	plen = len(path)
	if plen < 10 || path[plen-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}
	prefix := path[:plen-10]
	fs := &fasthttp.FS{
		Root:               rootpath,
		IndexNames:         []string{"index.html"},
		GenerateIndexPages: false,
		AcceptByteRange:    true,
	}
	stripSlashes := strings.Count(prefix, "/")
	if stripSlashes > 0 {
		fs.PathRewrite = fasthttp.NewPathSlashesStripper(stripSlashes)
	}
	fileHandler := fs.NewRequestHandler()
	group.server.router.GET(path, func(ctx *Context) {
		fileHandler(ctx.Rctx)
	})
}

func (group *RouterGroup)StaticFile(path string,rootpath string){
	if false == group.isRoot{
		path = group.basePath+path
	}
	group.server.router.GET(path,func(ctx *Context){
		fasthttp.ServeFile(ctx.Rctx,rootpath)
	})
}