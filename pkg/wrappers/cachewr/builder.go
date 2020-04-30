// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package cachewr

import (
	"errors"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "cache"

// Builder returns a builder for the wrapper component
func Builder(defaultCfg Config) builder.BuildWrapperFn {
	return func(b *builder.Builder, id string, def builder.WrapperDef, list xlist.List) (xlist.List, error) {
		cfg := defaultCfg
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return New(list, cfg), nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return rCfg, err
	}
	if ok {
		if ttl <= 0 {
			return rCfg, errors.New("invalid 'ttl'")
		}
		rCfg.TTL = ttl
	}

	negativettl, ok, err := option.Int(opts, "negativettl")
	if err != nil {
		return rCfg, err
	}
	if ok {
		if negativettl <= 0 && negativettl != xlist.NeverCache {
			return rCfg, errors.New("invalid 'negativettl'")
		}
		rCfg.NegativeTTL = negativettl
	}

	return rCfg, nil
}

func init() {
	builder.RegisterWrapperBuilder(BuildClass, Builder(DefaultConfig()))
}
