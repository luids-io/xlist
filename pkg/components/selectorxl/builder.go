// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package selectorxl

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "selector"

// Builder returns a builder for selector List component
func Builder(opt ...Option) listbuilder.BuildCheckerFn {
	return func(builder *listbuilder.Builder, parents []string, list listbuilder.ListDef) (xlist.Checker, error) {
		if len(list.Resources) != len(list.Contains) {
			return nil, errors.New("number of resources doesn't match with members")
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
		bl := New(bopt...)
		for idx, sublist := range list.Contains {
			sl, err := builder.BuildChild(append(parents, list.ID), sublist)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", list.Contains[idx].ID, err)
			}
			resource := list.Resources[idx]
			if !resource.InArray(sl.Resources()) {
				return nil, fmt.Errorf("child '%s' doesn't checks resource '%s'", list.Contains[idx].ID, resource)
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
	listbuilder.RegisterCheckerBuilder(BuildClass, Builder())
}
