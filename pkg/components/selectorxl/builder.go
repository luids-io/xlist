// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package selectorxl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "selector"

// Builder returns a builder for selector List component
func Builder(cfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		if len(def.Resources) != len(def.Contains) {
			return nil, errors.New("number of resources doesn't match with members")
		}
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		// create services
		services := make(map[xlist.Resource]xlist.List, len(def.Contains))
		for idx, childdef := range def.Contains {
			sl, err := b.BuildChild(append(parents, def.ID), childdef)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[idx].ID, err)
			}
			resource := def.Resources[idx]
			if !resource.InArray(sl.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[idx].ID, resource)
			}
			services[resource] = sl
		}
		return New(services, cfg), nil
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
	return rCfg, nil
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder(Config{}))
}
