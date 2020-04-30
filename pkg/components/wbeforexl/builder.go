// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package wbeforexl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "wbefore"

// Builder returns a builder for "white before" List component
func Builder(defaultCfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		cfg := defaultCfg
		if len(def.Contains) != 2 {
			return nil, errors.New("number of childs must be 2")
		}
		whitelist, err := b.BuildChild(append(parents, def.ID), def.Contains[0])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[0].ID, err)
		}
		for _, r := range def.Resources {
			if !r.InArray(whitelist.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[0].ID, r)
			}
		}
		blacklist, err := b.BuildChild(append(parents, def.ID), def.Contains[1])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[1].ID, err)
		}
		for _, r := range def.Resources {
			if !r.InArray(blacklist.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[1].ID, r)
			}
		}
		if def.Opts != nil {
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return New(whitelist, blacklist, def.Resources, cfg), nil
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
