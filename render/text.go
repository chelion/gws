// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"fmt"
	"io"
)

// String contains the given interface object slice and its format.
type String struct {
	Format string
	Data   []interface{}
}

// Render (String) writes data with custom ContentType.
func (r String) Render(renderIO RenderIO)(err error) {
	renderIO.SetContentType(PlainContentType)
	if len(r.Data) > 0 {
		_, err = fmt.Fprintf(renderIO, r.Format, r.Data...)
		return
	}
	_, err = io.WriteString(renderIO, r.Format)
	return
}

