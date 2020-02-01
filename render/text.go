// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"fmt"
	"io"
	"github.com/chelion/gws/fasthttp"
)

// String contains the given interface object slice and its format.
type String struct {
	Format string
	Data   []interface{}
}

// Render (String) writes data with custom ContentType.
func (r String) Render(ctx *fasthttp.RequestCtx) error {
	return WriteString(ctx, r.Format, r.Data)
}

// WriteContentType (String) writes Plain ContentType.
func (r String) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, PlainContentType)
}

// WriteString writes data according to its format and write custom ContentType.
func WriteString(ctx *fasthttp.RequestCtx, format string, data []interface{}) (err error) {
	writeContentType(ctx, PlainContentType)
	if len(data) > 0 {
		_, err = fmt.Fprintf(ctx, format, data...)
		return
	}
	_, err = io.WriteString(ctx, format)
	return
}
