package gws
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	"github.com/chelion/gws/fasthttp"
)

var (
	defaultContentType = []byte("text/plain; charset=utf-8")
	questionMark       = []byte("?")
)

type Router struct{
	Server *GWS
	methodTrees map[string]*Node
	MiddlewareHandlers []ContextHandler
	NotFound ContextHandler
	MethodNotAllowed ContextHandler
}

func (r *Router) handle(method, path string, handler ContextHandler) {
	if r.methodTrees == nil {
		r.methodTrees = make(map[string]*Node)
	}
	nodeRoot := r.methodTrees[method]
	if nodeRoot == nil {
		nodeRoot = &Node{nodeType:NodeTypeRoot,handler:nil,wildcardParams:make([]string,0),childNodes:make([]*Node,0)}
		r.methodTrees[method] = nodeRoot
	}
	if "/" == path{
		nodeRoot.handler = handler
	}else{
		nodeRoot.Add(path, handler)
	}
}

func NewRouter(server *GWS) *Router {
	return &Router{
		Server:				 server,
		MethodNotAllowed: 		nil,
		NotFound:          		nil,
		MiddlewareHandlers:		nil,
	}
}

func (r *Router) GET(path string, handler ContextHandler) {
	r.handle("GET", path, handler)
}

func (r *Router) HEAD(path string, handler ContextHandler) {
	r.handle("HEAD", path, handler)
}

func (r *Router) OPTIONS(path string, handler ContextHandler) {
	r.handle("OPTIONS", path, handler)
}

func (r *Router) POST(path string, handler ContextHandler) {
	r.handle("POST", path, handler)
}

func (r *Router) PUT(path string, handler ContextHandler) {
	r.handle("PUT", path, handler)
}

func (r *Router) PATCH(path string, handler ContextHandler) {
	r.handle("PATCH", path, handler)
}

func (r *Router) DELETE(path string, handler ContextHandler) {
	r.handle("DELETE", path, handler)
}

func (r *Router) Handler(ctx *fasthttp.RequestCtx){
	var handler ContextHandler
	path := string(ctx.URI().Path())
	method := string(ctx.Method())
	if path == "/"{
		node := r.methodTrees[method]
		if nil != node{
			handler = node.handler
		}else{
			if nil != r.MethodNotAllowed{
				handler = r.MethodNotAllowed
			}else{
				defaultNotAllowedHandler(ctx)
				return
			}
		}
	}else{
		nodeRoot := r.methodTrees[method]
		if nil == nodeRoot{
			if nil != r.MethodNotAllowed{
				handler = r.MethodNotAllowed
			}else{
				defaultNotAllowedHandler(ctx)
				return
			}
		}else{
			h,_ := nodeRoot.Find(path,ctx.SetUserValue)
			if nil != h{
				handler = h
			}else{
				if nil != r.NotFound{
					handler = r.NotFound
				}else{
					defaultNotFoundHandler(ctx)
					return
				}
			}
		}
	}
	if nil != handler{
		context := Context{Rctx:ctx,Server:r.Server,index:-1,handlers:r.MiddlewareHandlers}
		context.handlers = append(context.handlers,handler)
		context.Next()
	}
}

func defaultNotAllowedHandler(ctx *fasthttp.RequestCtx){
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
	ctx.SetContentTypeBytes(defaultContentType)
	ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed))
}

func defaultNotFoundHandler(ctx *fasthttp.RequestCtx){
	ctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound),
	fasthttp.StatusNotFound)
}
