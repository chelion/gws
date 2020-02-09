package render
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

type Data struct{
	ContentType string
	Data []byte
}

func (r Data)Render(renderIO RenderIO)(err error){
	renderIO.SetContentType(r.ContentType)
	_,err = renderIO.Write(r.Data)
	return
}
