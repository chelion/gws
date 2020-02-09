// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"io"
)

// Reader contains the IO reader and its length, and custom ContentType and other headers.
type Reader struct {
	ContentType   string
	ContentLength int64
	Reader        io.Reader
	Headers       map[string]string
}

// Render (Reader) writes data with custom ContentType and headers.
func (r Reader) Render(renderIO RenderIO) (err error) {
	renderIO.SetContentType(r.ContentType)
	renderIO.SetContentLength(int(r.ContentLength))
	_, err = io.Copy(renderIO, r.Reader)
	return
}
