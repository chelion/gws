// Copyright 2017 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"github.com/chelion/gws/fasthttp"

	"github.com/ugorji/go/codec"
)

// MsgPack contains the given interface object.
type MsgPack struct {
	Data interface{}
}

// WriteContentType (MsgPack) writes MsgPack ContentType.
func (r MsgPack) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, MsgPackContentType)
}

// Render (MsgPack) encodes the given interface object and writes data with custom ContentType.
func (r MsgPack) Render(ctx *fasthttp.RequestCtx) error {
	return WriteMsgPack(ctx, r.Data)
}

// WriteMsgPack writes MsgPack ContentType and encodes the given interface object.
func WriteMsgPack(ctx *fasthttp.RequestCtx, obj interface{}) error {
	writeContentType(ctx, MsgPackContentType)
	var mh codec.MsgpackHandle
	return codec.NewEncoder(ctx, &mh).Encode(obj)
}
