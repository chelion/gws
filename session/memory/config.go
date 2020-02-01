package memory

// session memory config
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"github.com/chelion/gws/cache"
)

type MCacheConfig struct{
	Cache cache.Cache
	Addr string
	VirtualNodeNum int32
}

type Config struct {
	CacheConfig []*MCacheConfig
}

func (mc *Config) Name() string {
	return ProviderName
}