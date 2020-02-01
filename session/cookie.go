package session
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import (
	"github.com/chelion/gws/fasthttp"
	"time"
)

type Cookie struct {}

func NewCookie() *Cookie {
	return &Cookie{}
}

func (c *Cookie) Get(ctx *fasthttp.RequestCtx, name string) (value string) {
	cookieByte := ctx.Request.Header.Cookie(name)
	if len(cookieByte) > 0 {
		value = string(cookieByte)
	}
	return
}

func (c *Cookie) Set(ctx *fasthttp.RequestCtx, name string, value string, domain string, expires time.Duration, secure bool) {

	cookie := fasthttp.AcquireCookie()
	
	cookie.SetKey(name)
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	cookie.SetDomain(domain)
	if expires >= 0 {
		if expires == 0 {
			// = 0 unlimited life
			cookie.SetExpire(fasthttp.CookieExpireUnlimited)
		} else {
			// > 0
			cookie.SetExpire(time.Now().Add(expires))
		}
	}
	if ctx.IsTLS() && secure {
		cookie.SetSecure(true)
	}

	cookie.SetValue(value)
	ctx.Response.Header.SetCookie(cookie)
	fasthttp.ReleaseCookie(cookie)
}

func (c *Cookie) Delete(ctx *fasthttp.RequestCtx, name string) {

	ctx.Response.Header.DelCookie(name)

	cookie := fasthttp.AcquireCookie()
	cookie.SetKey(name)
	cookie.SetValue("")
	cookie.SetPath("/")
	cookie.SetHTTPOnly(true)
	//RFC says 1 second, but let's do it 1 minute to make sure is working...
	exp := time.Now().Add(-time.Duration(1) * time.Minute)
	cookie.SetExpire(exp)
	ctx.Response.Header.SetCookie(cookie)

	// delete request's cookie also
	ctx.Request.Header.DelCookie(name)
	fasthttp.ReleaseCookie(cookie)
}


