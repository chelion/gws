package render
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"github.com/chelion/gws/fasthttp"
)

type Data struct{
	ContentType string
	Data []byte
}

func (r Data)Render(ctx *fasthttp.RequestCtx)(err error){
	r.WriteContentType(ctx)
	_,err = ctx.Write(r.Data)
	return
}

func (r Data)WriteContentType(ctx *fasthttp.RequestCtx){
	writeContentType(ctx,r.ContentType)
}
