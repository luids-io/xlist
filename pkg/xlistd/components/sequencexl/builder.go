package sequencexl

import (
	"context"
	"fmt"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		cfg := defaultCfg
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		childs := make([]xlistd.List, 0, len(def.Contains))
		for _, childDef := range def.Contains {
			if childDef.Disabled {
				continue
			}
			child, err := b.BuildChild(append(parents, def.ID), childDef)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", childDef.ID, err)
			}
			childres, err := child.Resources(context.Background())
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", childDef.ID, err)
			}
			for _, r := range def.Resources {
				if !r.InArray(childres) {
					return nil, fmt.Errorf("child '%s' doesn't checks resource '%s': %v", childDef.ID, r, childDef)
				}
			}
			childs = append(childs, child)
		}
		return New(def.ID, childs, def.Resources, cfg), nil
	}
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reason = reason
	}
	skipErrors, ok, err := option.Bool(opts, "skiperrors")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.SkipErrors = skipErrors
	}
	returnFirst, ok, err := option.Bool(opts, "first")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.FirstResponse = returnFirst
	}
	return dst, nil
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder(Config{}))
}
