// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package gws

import (
	"sync"
	"github.com/chelion/gws/fasthttp"
)

var (
	defaultContentType = []byte("text/plain; charset=utf-8")
	questionMark       = []byte("?")
)

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes


type Router struct {
	trees map[string]*node

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	//middleware handler, use gw.Use(...),use you middleware
	MiddlewareHandlers []ContextHandler

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound ContextHandler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed ContextHandler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(*Context, interface{})

	Server *GWS
	contextPool       sync.Pool
}

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func NewRouter(server *GWS) *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		MiddlewareHandlers:		nil,
		Server:					server,
	}
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle ContextHandler) {
	r.Handle("GET", path, handle)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle ContextHandler) {
	r.Handle("HEAD", path, handle)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle ContextHandler) {
	r.Handle("OPTIONS", path, handle)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle ContextHandler) {
	r.Handle("POST", path, handle)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle ContextHandler) {
	r.Handle("PUT", path, handle)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle ContextHandler) {
	r.Handle("PATCH", path, handle)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle ContextHandler) {
	r.Handle("DELETE", path, handle)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handle ContextHandler) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root
	}

	root.addRoute(path, handle)
}


func (r *Router) recv(ctx *fasthttp.RequestCtx) {
	if rcv := recover(); rcv != nil {
		r.PanicHandler(&Context{Rctx:ctx,Server:r.Server}, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string, ctx *fasthttp.RequestCtx) (ContextHandler, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path, ctx)
	}
	return nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" || path == "/*" { // server-wide
		for method := range r.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _ := r.trees[method].getValue(path, nil)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

func (r *Router)acquireContext(rCtx *fasthttp.RequestCtx)(*Context){
	v := r.contextPool.Get()
	if v == nil {
		return &Context{Rctx:rCtx,Server:r.Server,index:-1,handlers:r.MiddlewareHandlers}
	}
	c := v.(*Context)
	c.Rctx = rCtx
	c.Server = r.Server
	c.index = -1
	c.handlers = r.MiddlewareHandlers
	return c
}

func (r *Router)releaseContext(c *Context) {
	r.contextPool.Put(c)
}

// Handler makes the router implement the fasthttp.ListenAndServe interface.
func (r *Router) Handler(rCtx *fasthttp.RequestCtx) {
	if r.PanicHandler != nil {
		defer r.recv(rCtx)
	}
	path := string(rCtx.URI().PathOriginal())
	method := string(rCtx.Method())
	if true == r.Server.isDebug{
		r.Server.Logger.Println(method,string(rCtx.URI().Path()))
	}
	if root := r.trees[method]; root != nil {
		if f, tsr := root.getValue(path, rCtx); f != nil {
			context := r.acquireContext(rCtx)
			context.handlers = append(context.handlers,f)
			context.Next()
			r.releaseContext(context)
			return
		} else if method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if method != "GET" {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && r.RedirectTrailingSlash {
				var uri string
				if len(path) > 1 && path[len(path)-1] == '/' {
					uri = path[:len(path)-1]
				} else {
					uri = path + "/"
				}
				
				if len(rCtx.URI().QueryString()) > 0 {
					uri += "?" + string(rCtx.QueryArgs().QueryString())
				}
				
				rCtx.Redirect(uri, code)
				return
			}

			// Try to fix the request path
			if r.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					string(rCtx.URI().Path()),
					r.RedirectTrailingSlash,
				)

				if found {
					queryBuf := rCtx.URI().QueryString()
					if len(queryBuf) > 0 {
						fixedPath = append(fixedPath, questionMark...)
						fixedPath = append(fixedPath, queryBuf...)
					}
					uri := string(fixedPath)
					rCtx.Redirect(uri, code)
					return
				}
			}
		}
	}

	if method == "OPTIONS" {
		// Handle OPTIONS requests
		if r.HandleOPTIONS {
			if allow := r.allowed(path, method); len(allow) > 0 {
				rCtx.Response.Header.Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if r.HandleMethodNotAllowed {
			if allow := r.allowed(path, method); len(allow) > 0 {
				rCtx.Response.Header.Set("Allow", allow)
				context := r.acquireContext(rCtx)
				if r.MethodNotAllowed != nil {
					context.handlers = append(context.handlers,r.MethodNotAllowed)
				} else {
					context.handlers = append(context.handlers,defaultNotAllowedHandler)
				}
				context.Next()
				r.releaseContext(context)
				return
			}
		}
	}

	// Handle 404
	context := r.acquireContext(rCtx)
	if r.NotFound != nil {
		context.handlers = append(context.handlers,r.NotFound)
	} else {
		context.handlers = append(context.handlers,defaultNotAllowedHandler)
	}
	context.Next()
	r.releaseContext(context)
}


func defaultNotAllowedHandler(ctx *Context){
	ctx.Rctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
	ctx.Rctx.SetContentTypeBytes(defaultContentType)
	ctx.Rctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusMethodNotAllowed))
}

func defaultNotFoundHandler(ctx *Context){
	ctx.Rctx.Error(fasthttp.StatusMessage(fasthttp.StatusNotFound),
	fasthttp.StatusNotFound)
}