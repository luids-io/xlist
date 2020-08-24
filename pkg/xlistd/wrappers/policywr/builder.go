// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package policywr

import (
	"errors"
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/reason"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)

// Builder returns a builder for the component
func Builder(defaultCfg Config) builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlistd.List) (xlistd.List, error) {
		cfg := defaultCfg
		if def.Opts == nil {
			return nil, errors.New("'value' option is required")
		}
		//gets value policy
		value, ok, err := option.String(def.Opts, "value")
		if !ok || value == "" {
			return nil, errors.New("'value' option is required")
		}
		policy := reason.NewPolicy()
		err = policy.FromString(fmt.Sprintf("[policy]%s[/policy]", value))
		if err != nil {
			return nil, fmt.Errorf("'value' invalid: %v", err)
		}
		//gets config
		cfg, err = parseOptions(cfg, def.Opts)
		if err != nil {
			return nil, err
		}
		w := New(list, policy, cfg)
		return w, nil
	}
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	merge, ok, err := option.Bool(opts, "merge")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Merge = merge
	}

	threshold, ok, err := option.Int(opts, "threshold")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.UseThreshold = true
		dst.Score = threshold
	}
	return dst, nil
}

func init() {
	builder.RegisterWrapperBuilder(WrapperClass, Builder(Config{}))
}
