// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"encoding/xml"
	"github.com/chelion/gws/fasthttp"
)

// XML contains the given interface object.
type XML struct {
	Data interface{}
}

// Render (XML) encodes the given interface object and writes data with custom ContentType.
func (r XML) Render(ctx *fasthttp.RequestCtx) error {
	r.WriteContentType(ctx)
	return xml.NewEncoder(ctx).Encode(r.Data)
}

// WriteContentType (XML) writes XML ContentType for response.
func (r XML) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, XMLContentType)
}
