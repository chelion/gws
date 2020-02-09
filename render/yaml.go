// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package render

import (
	"gopkg.in/yaml.v2"
)

// YAML contains the given interface object.
type YAML struct {
	Data interface{}
}

// Render (YAML) marshals the given interface object and writes data with custom ContentType.
func (r YAML) Render(renderIO RenderIO) error {
	renderIO.SetContentType(YAMLContentType)
	bytes, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}
	_, err = renderIO.Write(bytes)
	return err
}
