// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"github.com/golang/protobuf/proto"
)

// ProtoBuf contains the given interface object.
type ProtoBuf struct {
	Data interface{}
}

// Render (ProtoBuf) marshals the given interface object and writes data with custom ContentType.
func (r ProtoBuf) Render(renderIO RenderIO) error {
	renderIO.SetContentType(ProtoBufContentType)
	bytes, err := proto.Marshal(r.Data.(proto.Message))
	if err != nil {
		return err
	}
	_, err = renderIO.Write(bytes)
	return err
}

