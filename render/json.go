// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"bytes"
	"fmt"
	"html/template"
	"encoding/json"
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
func (r JSON) Render(renderIO RenderIO) (err error) {
	renderIO.SetContentType(JSONContentType)
	encoder := json.NewEncoder(renderIO)
	err = encoder.Encode(r.Data)
	return err
}


// Render (IndentedJSON) marshals the given interface object and writes it with custom ContentType.
func (r IndentedJSON) Render(renderIO RenderIO) error {
	renderIO.SetContentType(JSONContentType)
	jsonBytes, err := json.MarshalIndent(r.Data, "", "    ")
	if err != nil {
		return err
	}
	_, err = renderIO.Write(jsonBytes)
	return err
}


// Render (SecureJSON) marshals the given interface object and writes it with custom ContentType.
func (r SecureJSON) Render(renderIO RenderIO) error {
	renderIO.SetContentType(JSONContentType)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	// if the jsonBytes is array values
	if bytes.HasPrefix(jsonBytes, []byte("[")) && bytes.HasSuffix(jsonBytes, []byte("]")) {
		_, err = renderIO.Write([]byte(r.Prefix))
		if err != nil {
			return err
		}
	}
	_, err = renderIO.Write(jsonBytes)
	return err
}

// Render (JsonpJSON) marshals the given interface object and writes it and its callback with custom ContentType.
func (r JsonpJSON) Render(renderIO RenderIO) (err error) {
	renderIO.SetContentType(JSONPContentType)
	ret, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	if r.Callback == "" {
		_, err = renderIO.Write(ret)
		return err
	}

	callback := template.JSEscapeString(r.Callback)
	_, err = renderIO.Write([]byte(callback))
	if err != nil {
		return err
	}
	_, err = renderIO.Write([]byte("("))
	if err != nil {
		return err
	}
	_, err = renderIO.Write(ret)
	if err != nil {
		return err
	}
	_, err = renderIO.Write([]byte(")"))
	if err != nil {
		return err
	}

	return nil
}


// Render (AsciiJSON) marshals the given interface object and writes it with custom ContentType.
func (r AsciiJSON) Render(renderIO RenderIO) (err error) {
	renderIO.SetContentType(JSONAsciiContentType)
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

	_, err = renderIO.Write(buffer.Bytes())
	return err
}


// Render (PureJSON) writes custom ContentType and encodes the given interface object.
func (r PureJSON) Render(renderIO RenderIO) error {
	renderIO.SetContentType(JSONContentType)
	encoder := json.NewEncoder(renderIO)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(r.Data)
}

