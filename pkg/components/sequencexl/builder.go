package sequencexl

import (
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "sequence"

// Builder returns a builder for sequence List component
func Builder(cfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		childs := make([]xlist.Checker, 0, len(def.Contains))
		for _, childDef := range def.Contains {
			if childDef.Disabled {
				continue
			}
			child, err := b.BuildChild(append(parents, def.ID), childDef)
			if err != nil {
				return nil, fmt.Errorf("constructing child '%s': %v", childDef.ID, err)
			}
			for _, r := range def.Resources {
				if !r.InArray(child.Resources()) {
					return nil, fmt.Errorf("child '%s' doesn't checks resource '%s': %v", childDef.ID, r, childDef)
				}
			}
			childs = append(childs, child)
		}
		return New(childs, cfg), nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Reason = reason
	}
	skipErrors, ok, err := option.Bool(opts, "skiperrors")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.SkipErrors = skipErrors
	}
	returnFirst, ok, err := option.Bool(opts, "first")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.FirstResponse = returnFirst
	}
	return rCfg, nil
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder(Config{}))
}
