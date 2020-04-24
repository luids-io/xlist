// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package parallelxl

import (
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "parallel"

// Builder returns a builder for parallel List component
func Builder(cfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		childs := make([]xlist.Checker, 0, len(def.Contains))
		for _, sublist := range def.Contains {
			if sublist.Disabled {
				continue
			}
			child, err := b.BuildChild(append(parents, def.ID), sublist)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", sublist.ID, err)
			}
			for _, r := range def.Resources {
				if !r.InArray(child.Resources()) {
					return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", sublist.ID, r)
				}
			}
			childs = append(childs, child)
		}
		return New(childs, def.Resources, cfg), nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Reason = reason
	}
	skipErrors, ok, err := option.Bool(opts, "skiperrors")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.SkipErrors = skipErrors
	}
	returnFirst, ok, err := option.Bool(opts, "first")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.FirstResponse = returnFirst
	}
	return rCfg, nil
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder(Config{}))
}
