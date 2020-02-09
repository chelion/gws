// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"encoding/xml"
)

// XML contains the given interface object.
type XML struct {
	Data interface{}
}

// Render (XML) encodes the given interface object and writes data with custom ContentType.
func (r XML) Render(renderIO RenderIO) error {
	renderIO.SetContentType(XMLContentType)
	return xml.NewEncoder(renderIO).Encode(r.Data)
}
