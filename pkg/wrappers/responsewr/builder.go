// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package responsewr

import (
	"errors"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "response"

// Builder returns a builder for the component
func Builder(opt ...Option) listbuilder.BuildWrapperFn {
	return func(builder *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.List) (xlist.List, error) {
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts, listID)
			if err != nil {
				return nil, err
			}
		}
		w := New(bl, bopt...)
		return w, nil
	}
}

func parseOptions(bopt []Option, opts map[string]interface{}, listID string) ([]Option, error) {
	clean, ok, err := option.Bool(opts, "clean")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Clean(clean))
	}

	aggregate, ok, err := option.Bool(opts, "aggregate")
	if err != nil {
		return bopt, err
	}
	if ok {
		if clean && aggregate {
			return bopt, errors.New("'clean' and 'aggregate' fields are incompatible")
		}
		bopt = append(bopt, Aggregate(aggregate))
	}

	negate, ok, err := option.Bool(opts, "negate")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Negate(negate))
	}

	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, TTL(ttl))
	}

	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Reason(reason))
	}

	preffixID, ok, err := option.Bool(opts, "preffixid")
	if err != nil {
		return bopt, err
	}
	if ok && preffixID {
		bopt = append(bopt, PreffixReason(listID))
	}

	preffix, ok, err := option.String(opts, "preffix")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, PreffixReason(preffix))
	}

	threshold, ok, err := option.Int(opts, "threshold")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Threshold(threshold))
	}

	return bopt, nil
}

func init() {
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder())
}
