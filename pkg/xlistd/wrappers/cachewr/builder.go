// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package cachewr

import (
	"errors"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildWrapperFn {
	return func(b *xlistd.Builder, def xlistd.WrapperDef, list xlistd.List) (xlistd.List, error) {
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

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return dst, err
	}
	if ok {
		if ttl < xlist.NeverCache {
			return dst, errors.New("invalid 'ttl'")
		}
		dst.TTL = ttl
	}

	negativettl, ok, err := option.Int(opts, "negativettl")
	if err != nil {
		return dst, err
	}
	if ok {
		if negativettl < xlist.NeverCache {
			return dst, errors.New("invalid 'negativettl'")
		}
		dst.NegativeTTL = negativettl
	}

	minttl, ok, err := option.Int(opts, "minttl")
	if err != nil {
		return dst, err
	}
	if ok {
		if minttl < 0 {
			return dst, errors.New("invalid 'minttl'")
		}
		dst.MinTTL = minttl
	}

	maxttl, ok, err := option.Int(opts, "maxttl")
	if err != nil {
		return dst, err
	}
	if ok {
		if maxttl < 0 {
			return dst, errors.New("invalid 'maxttl'")
		}
		dst.MaxTTL = maxttl
	}

	return dst, nil
}

func init() {
	xlistd.RegisterWrapperBuilder(WrapperClass, Builder(DefaultConfig()))
}
