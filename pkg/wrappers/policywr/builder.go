// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package policywr

import (
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/core/reason"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "policy"

// Builder returns a builder for the component
func Builder(defaultCfg Config) builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlist.List) (xlist.List, error) {
		cfg := defaultCfg
		policy := reason.NewPolicy()
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
			for k, v := range def.Opts {
				if k == "merge" || k == "threshold" {
					continue
				}
				err := policy.Set(k, fmt.Sprintf("%v", v))
				if err != nil {
					return nil, fmt.Errorf("invalid policy: %v", err)
				}
			}
		}
		w := New(list, policy, cfg)
		return w, nil
	}
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	merge, ok, err := option.Bool(opts, "merge")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Merge = merge
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
	builder.RegisterWrapperBuilder(BuildClass, Builder(Config{}))
}
