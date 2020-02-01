// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package render

import (
	"github.com/chelion/gws/fasthttp"

	"gopkg.in/yaml.v2"
)

// YAML contains the given interface object.
type YAML struct {
	Data interface{}
}

// Render (YAML) marshals the given interface object and writes data with custom ContentType.
func (r YAML) Render(ctx *fasthttp.RequestCtx) error {
	r.WriteContentType(ctx)
	bytes, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = ctx.Write(bytes)
	return err
}

// WriteContentType (YAML) writes YAML ContentType for response.
func (r YAML) WriteContentType(ctx *fasthttp.RequestCtx) {
	writeContentType(ctx, YAMLContentType)
}
