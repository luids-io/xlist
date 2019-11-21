// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package wbeforexl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "wbefore"

// Builder returns a builder for "white before" List component
func Builder(opt ...Option) listbuilder.BuildCheckerFn {
	return func(builder *listbuilder.Builder, parents []string, list listbuilder.ListDef) (xlist.Checker, error) {
		if len(list.Contains) != 2 {
			return nil, errors.New("number of childs must be 2")
		}
		// constructs childs
		whitelist, err := builder.BuildChild(append(parents, list.ID), list.Contains[0])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", list.Contains[0].ID, err)
		}
		for _, r := range list.Resources {
			if !r.InArray(whitelist.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", list.Contains[0].ID, r)
			}
		}
		blacklist, err := builder.BuildChild(append(parents, list.ID), list.Contains[1])
		if err != nil {
			return nil, fmt.Errorf("constructing child '%s': %v", list.Contains[1].ID, err)
		}
		for _, r := range list.Resources {
			if !r.InArray(blacklist.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", list.Contains[1].ID, r)
			}
		}

		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if list.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, list.Opts)
			if err != nil {
				return nil, err
			}
		}
		bl := New(list.Resources, bopt...)
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
	listbuilder.RegisterCheckerBuilder(BuildClass, Builder())
}
