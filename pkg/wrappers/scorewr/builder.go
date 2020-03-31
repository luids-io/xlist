// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package scorewr

import (
	"fmt"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class name of component builder
const BuildClass = "score"

// Builder returns a builder for the component
func Builder(opt ...Option) listbuilder.BuildWrapperFn {
	return func(builder *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.List) (xlist.List, error) {
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		score := 0
		if def.Opts != nil {
			v, ok, err := option.Int(def.Opts, "value")
			if err != nil {
				return nil, err
			}
			if ok {
				score = v
			}
		}
		w := New(bl, score, bopt...)
		if def.Opts != nil {
			err := addExprFromOpts(w, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return w, nil
	}
}

func addExprFromOpts(w *Wrapper, opts map[string]interface{}) error {
	matches, ok, err := option.SliceHash(opts, "matches")
	if err != nil {
		return err
	}
	if ok {
		for _, match := range matches {
			expr, ok, err := option.String(match, "expr")
			if err != nil || !ok {
				return err
			}
			value, ok, err := option.Int(match, "value")
			if err != nil || !ok {
				return err
			}
			err = w.AddExpr(expr, value)
			if err != nil {
				return fmt.Errorf("invalid 'matches': invalid 'expr': %s %v", expr, err)
			}
		}
	}
	return nil
}

func init() {
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder())
}
