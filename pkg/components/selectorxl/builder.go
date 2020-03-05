// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package selectorxl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "selector"

// Builder returns a builder for selector List component
func Builder(opt ...Option) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
		if len(def.Resources) != len(def.Contains) {
			return nil, errors.New("number of resources doesn't match with members")
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
		bl := New(bopt...)
		for idx, sublist := range def.Contains {
			sl, err := builder.BuildChild(append(parents, def.ID), sublist)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", def.Contains[idx].ID, err)
			}
			resource := def.Resources[idx]
			if !resource.InArray(sl.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", def.Contains[idx].ID, resource)
			}
			bl.SetService(resource, sl)
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
	return bopt, nil
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder())
}
