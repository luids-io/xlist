// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package policywr

import (
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/reason"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "policy"

// Builder returns a builder for the component
func Builder(cfg Config) listbuilder.BuildWrapperFn {
	return func(builder *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.List) (xlist.List, error) {
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
		w := New(bl, policy, cfg)
		return w, nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	merge, ok, err := option.Bool(opts, "merge")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Merge = merge
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
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder(Config{}))
}
