// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package cachewr

import (
	"errors"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "cache"

// Builder returns a builder for the wrapper component
func Builder(opt ...Option) listbuilder.BuildWrapperFn {
	return func(builder *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.List) (xlist.List, error) {
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		w := New(bl, bopt...)
		return w, nil
	}
}

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return bopt, err
	}
	if ok {
		if ttl <= 0 {
			return bopt, errors.New("invalid 'ttl'")
		}
		bopt = append(bopt, TTL(ttl))
	}

	negativettl, ok, err := option.Int(opts, "negativettl")
	if err != nil {
		return bopt, err
	}
	if ok {
		if negativettl <= 0 && negativettl != xlist.NeverCache {
			return bopt, errors.New("invalid 'negativettl'")
		}
		bopt = append(bopt, NegativeTTL(negativettl))
	}

	return bopt, nil
}

func init() {
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder())
}
