// Copyright 2018 Gin Core Team.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"github.com/chelion/gws/fasthttp"

	"github.com/golang/protobuf/proto"
)

// ProtoBuf contains the given interface object.
type ProtoBuf struct {
	Data interface{}
}

// Render (ProtoBuf) marshals the given interface object and writes data with custom ContentType.
func (r ProtoBuf) Render(ctx *fasthttp.RequestCtx) error {
	r.WriteContentType(ctx)
	bytes, err := proto.Marshal(r.Data.(proto.Message))
	if err != nil {
		return err
	}
	_, err = ctx.Write(bytes)
	return err
}

// WriteContentType (ProtoBuf) writes ProtoBuf ContentType.
func (r ProtoBuf) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, ProtoBufContentType)
}
