// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"github.com/vmihailenco/msgpack"
)

// MsgPack contains the given interface object.
type MsgPack struct {
	Data interface{}
}

// Render (MsgPack) encodes the given interface object and writes data with custom ContentType.
func (r MsgPack) Render(renderIO RenderIO) error {
	renderIO.SetContentType(MsgPackContentType)
	msgPackBytes, err := msgpack.Marshal(r.Data)
	if err != nil {
		panic(err)
	}
	_, err = renderIO.Write(msgPackBytes)
	return err
}
