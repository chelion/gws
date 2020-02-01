// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"bytes"
	"fmt"
	"html/template"
	"github.com/chelion/gws/fasthttp"
	"github.com/json-iterator/go"
)

var(
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// JSON contains the given interface object.
type JSON struct {
	Data interface{}
}

// IndentedJSON contains the given interface object.
type IndentedJSON struct {
	Data interface{}
}

// SecureJSON contains the given interface object and its prefix.
type SecureJSON struct {
	Prefix string
	Data   interface{}
}

// JsonpJSON contains the given interface object its callback.
type JsonpJSON struct {
	Callback string
	Data     interface{}
}

// AsciiJSON contains the given interface object.
type AsciiJSON struct {
	Data interface{}
}

// SecureJSONPrefix is a string which represents SecureJSON prefix.
type SecureJSONPrefix string

// PureJSON contains the given interface object.
type PureJSON struct {
	Data interface{}
}


// Render (JSON) writes data with custom ContentType.
func (r JSON) Render(ctx *fasthttp.RequestCtx) (err error) {
	if err = WriteJSON(ctx, r.Data); err != nil {
		panic(err)
	}
	return
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSON) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, JSONContentType)
}

// WriteJSON marshals the given interface object and writes it with custom ContentType.
func WriteJSON(ctx *fasthttp.RequestCtx, obj interface{}) error {
	writeContentType(ctx, JSONContentType)
	encoder := json.NewEncoder(ctx)
	err := encoder.Encode(&obj)
	return err
}

// Render (IndentedJSON) marshals the given interface object and writes it with custom ContentType.
func (r IndentedJSON) Render(ctx *fasthttp.RequestCtx) error {
	r.WriteContentType(ctx)
	jsonBytes, err := json.MarshalIndent(r.Data, "", "    ")
	if err != nil {
		return err
	}
	_, err = ctx.Write(jsonBytes)
	return err
}

// WriteContentType (IndentedJSON) writes JSON ContentType.
func (r IndentedJSON) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, JSONContentType)
}

// Render (SecureJSON) marshals the given interface object and writes it with custom ContentType.
func (r SecureJSON) Render(ctx *fasthttp.RequestCtx) error {
	r.WriteContentType(ctx)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	// if the jsonBytes is array values
	if bytes.HasPrefix(jsonBytes, []byte("[")) && bytes.HasSuffix(jsonBytes, []byte("]")) {
		_, err = ctx.Write([]byte(r.Prefix))
		if err != nil {
			return err
		}
	}
	_, err = ctx.Write(jsonBytes)
	return err
}

// WriteContentType (SecureJSON) writes JSON ContentType.
func (r SecureJSON) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, JSONContentType)
}

// Render (JsonpJSON) marshals the given interface object and writes it and its callback with custom ContentType.
func (r JsonpJSON) Render(ctx *fasthttp.RequestCtx) (err error) {
	r.WriteContentType(ctx)
	ret, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	if r.Callback == "" {
		_, err = ctx.Write(ret)
		return err
	}

	callback := template.JSEscapeString(r.Callback)
	_, err = ctx.Write([]byte(callback))
	if err != nil {
		return err
	}
	_, err = ctx.Write([]byte("("))
	if err != nil {
		return err
	}
	_, err = ctx.Write(ret)
	if err != nil {
		return err
	}
	_, err = ctx.Write([]byte(")"))
	if err != nil {
		return err
	}

	return nil
}

// WriteContentType (JsonpJSON) writes Javascript ContentType.
func (r JsonpJSON) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, JSONPContentType)
}

// Render (AsciiJSON) marshals the given interface object and writes it with custom ContentType.
func (r AsciiJSON) Render(ctx *fasthttp.RequestCtx) (err error) {
	r.WriteContentType(ctx)
	ret, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	for _, r := range string(ret) {
		cvt := string(r)
		if r >= 128 {
			cvt = fmt.Sprintf("\\u%04x", int64(r))
		}
		buffer.WriteString(cvt)
	}

	_, err = ctx.Write(buffer.Bytes())
	return err
}

// WriteContentType (AsciiJSON) writes JSON ContentType.
func (r AsciiJSON) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, JSONAsciiContentType)
}

// Render (PureJSON) writes custom ContentType and encodes the given interface object.
func (r PureJSON) Render(ctx *fasthttp.RequestCtx) error {
	r.WriteContentType(ctx)
	encoder := json.NewEncoder(ctx)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(r.Data)
}

// WriteContentType (PureJSON) writes custom ContentType.
func (r PureJSON) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, JSONContentType)
}

