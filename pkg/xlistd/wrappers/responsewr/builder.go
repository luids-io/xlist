// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package responsewr

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
			cfg, err = parseOptions(cfg, def.Opts, list.ID())
			if err != nil {
				return nil, err
			}
		}
		return New(list, cfg), nil
	}
}

func parseOptions(src Config, opts map[string]interface{}, listID string) (Config, error) {
	dst := src
	clean, ok, err := option.Bool(opts, "clean")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Clean = clean
	}

	aggregate, ok, err := option.Bool(opts, "aggregate")
	if err != nil {
		return dst, err
	}
	if ok {
		if clean && aggregate {
			return dst, errors.New("'clean' and 'aggregate' fields are incompatible")
		}
		dst.Aggregate = aggregate
	}

	negate, ok, err := option.Bool(opts, "negate")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Negate = negate
	}

	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return dst, err
	}
	if ok && ttl >= xlist.NeverCache {
		dst.TTL = ttl
	}

	negativettl, ok, err := option.Int(opts, "negativettl")
	if err != nil {
		return dst, err
	}
	if ok && negativettl >= xlist.NeverCache {
		dst.NegativeTTL = negativettl
	}

	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reason = reason
	}

	preffixID, ok, err := option.Bool(opts, "preffixid")
	if err != nil {
		return dst, err
	}
	if ok && preffixID {
		dst.Preffix = listID
	}

	preffix, ok, err := option.String(opts, "preffix")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Preffix = preffix
	}

	threshold, ok, err := option.Int(opts, "threshold")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.UseThreshold = true
		dst.Score = threshold
	}

	return dst, nil
}

func init() {
	xlistd.RegisterWrapperBuilder(WrapperClass, Builder(Config{}))
}
