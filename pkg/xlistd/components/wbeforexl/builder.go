// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package wbeforexl

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		cfg := defaultCfg
		if len(def.Contains) != 2 {
			return nil, errors.New("number of childs must be 2")
		}
		whitelist, err := b.BuildChild(append(parents, def.ID), def.Contains[0])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[0].ID, err)
		}
		whiteres, err := whitelist.Resources(context.Background())
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[0].ID, err)
		}
		for _, r := range def.Resources {
			if !r.InArray(whiteres) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[0].ID, r)
			}
		}
		blacklist, err := b.BuildChild(append(parents, def.ID), def.Contains[1])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[1].ID, err)
		}
		blackres, err := blacklist.Resources(context.Background())
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[1].ID, err)
		}
		for _, r := range def.Resources {
			if !r.InArray(blackres) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[1].ID, r)
			}
		}
		if def.Opts != nil {
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return New(def.ID, whitelist, blacklist, def.Resources, cfg), nil
	}
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reason = reason
	}
	return dst, nil
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder(Config{}))
}
