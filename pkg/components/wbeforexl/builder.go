// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package wbeforexl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "wbefore"

// Builder returns a builder for "white before" List component
func Builder(opt ...Option) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
		if len(def.Contains) != 2 {
			return nil, errors.New("number of childs must be 2")
		}
		// constructs childs
		whitelist, err := builder.BuildChild(append(parents, def.ID), def.Contains[0])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[0].ID, err)
		}
		for _, r := range def.Resources {
			if !r.InArray(whitelist.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[0].ID, r)
			}
		}
		blacklist, err := builder.BuildChild(append(parents, def.ID), def.Contains[1])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[1].ID, err)
		}
		for _, r := range def.Resources {
			if !r.InArray(blacklist.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[1].ID, r)
			}
		}

		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		bl := New(def.Resources, bopt...)
		bl.SetWhitelist(whitelist)
		bl.SetBlacklist(blacklist)
		return bl, nil
	}
}

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Reason(reason))
	}
	return bopt, nil
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder())
}
