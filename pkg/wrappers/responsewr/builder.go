// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package responsewr

import (
	"errors"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "response"

// Builder returns a builder for the component
func Builder(cfg Config) builder.BuildWrapperFn {
	return func(b *builder.Builder, listID string, def builder.WrapperDef, bl xlist.List) (xlist.List, error) {
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts, listID)
			if err != nil {
				return nil, err
			}
		}
		return New(bl, cfg), nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}, listID string) (Config, error) {
	rCfg := cfg
	clean, ok, err := option.Bool(opts, "clean")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Clean = clean
	}

	aggregate, ok, err := option.Bool(opts, "aggregate")
	if err != nil {
		return rCfg, err
	}
	if ok {
		if clean && aggregate {
			return rCfg, errors.New("'clean' and 'aggregate' fields are incompatible")
		}
		rCfg.Aggregate = aggregate
	}

	negate, ok, err := option.Bool(opts, "negate")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Negate = negate
	}

	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return rCfg, err
	}
	if ok && ttl >= xlist.NeverCache {
		rCfg.TTL = ttl
	}

	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Reason = reason
	}

	preffixID, ok, err := option.Bool(opts, "preffixid")
	if err != nil {
		return rCfg, err
	}
	if ok && preffixID {
		rCfg.Preffix = listID
	}

	preffix, ok, err := option.String(opts, "preffix")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Preffix = preffix
	}

	threshold, ok, err := option.Int(opts, "threshold")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.UseThreshold = true
		rCfg.Score = threshold
	}

	return rCfg, nil
}

func init() {
	builder.RegisterWrapperBuilder(BuildClass, Builder(Config{}))
}
