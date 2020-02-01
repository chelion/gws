// Copyright 2018 Gin Core Team.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"io"
	"github.com/chelion/gws/fasthttp"
)

// Reader contains the IO reader and its length, and custom ContentType and other headers.
type Reader struct {
	ContentType   string
	ContentLength int64
	Reader        io.Reader
	Headers       map[string]string
}

// Render (Reader) writes data with custom ContentType and headers.
func (r Reader) Render(ctx *fasthttp.RequestCtx) (err error) {
	r.WriteContentType(ctx)
	r.writeHeaders(ctx,r.ContentLength)
	_, err = io.Copy(ctx, r.Reader)
	return
}

// WriteContentType (Reader) writes custom ContentType.
func (r Reader) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx,r.ContentType)
}

// writeHeaders writes custom Header.
func (r Reader) writeHeaders(ctx *fasthttp.RequestCtx,len int64) {
	ctx.Response.Header.SetContentLength(int(len))
}
