// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package parallelxl

import (
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "parallel"

// Builder returns a builder for parallel List component
func Builder(opt ...Option) listbuilder.BuildCheckerFn {
	return func(builder *listbuilder.Builder, parents []string, list listbuilder.ListDef) (xlist.Checker, error) {
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
		for _, sublist := range list.Contains {
			if sublist.Disabled {
				continue
			}
			sl, err := builder.BuildChild(append(parents, list.ID), sublist)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", sublist.ID, err)
			}
			for _, r := range list.Resources {
				if !r.InArray(sl.Resources()) {
					return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", sublist.ID, r)
				}
			}
			bl.Append(sl)
		}
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
	skipErrors, ok, err := option.Bool(opts, "skiperrors")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, SkipErrors(skipErrors))
	}
	returnFirst, ok, err := option.Bool(opts, "first")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, FirstResponse(returnFirst))
	}
	return bopt, nil
}

func init() {
	listbuilder.RegisterCheckerBuilder(BuildClass, Builder())
}
