// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package parallelxl

import (
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "parallel"

// Builder returns a builder for parallel List component
func Builder(opt ...Option) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
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
		for _, sublist := range def.Contains {
			if sublist.Disabled {
				continue
			}
			sl, err := builder.BuildChild(append(parents, def.ID), sublist)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", sublist.ID, err)
			}
			for _, r := range def.Resources {
				if !r.InArray(sl.Resources()) {
					return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", sublist.ID, r)
				}
			}
			bl.AddChecker(sl)
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
	listbuilder.RegisterListBuilder(BuildClass, Builder())
}
