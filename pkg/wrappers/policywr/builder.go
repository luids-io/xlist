// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package policywr

import (
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/reason"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "policy"

// Builder returns a builder for the component
func Builder(opt ...Option) listbuilder.WrapperBuilder {
	return func(b *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.Checker) (xlist.Checker, error) {
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		policy := reason.NewPolicy()
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
			for k, v := range def.Opts {
				if k == "merge" || k == "threshold" {
					continue
				}
				err := policy.Set(k, fmt.Sprintf("%v", v))
				if err != nil {
					return nil, fmt.Errorf("invalid policy: %v", err)
				}
			}
		}
		blc := New(bl, policy, bopt...)
		return blc, nil
	}
}

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	merge, ok, err := option.Bool(opts, "merge")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Merge(merge))
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
	listbuilder.RegisterWrapper(BuildClass, Builder())
}
