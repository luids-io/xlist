// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package timeoutwr

import (
	"time"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// DefaultTimeout sets default timeout in construction
const DefaultTimeout = 1 * time.Second

// BuildClass defines class name for component builder
const BuildClass = "timeout"

// Builder returns a builder for the component
func Builder(timeout time.Duration, opt ...Option) listbuilder.WrapperBuilder {
	return func(b *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.Checker) (xlist.Checker, error) {
		setTimeout := timeout
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			v, ok, err := option.Int(def.Opts, "timeout")
			if err != nil {
				return nil, err
			}
			if ok {
				setTimeout = time.Duration(v) * time.Millisecond
			}
		}
		return New(setTimeout, bl, bopt...), nil
	}
}

func init() {
	listbuilder.RegisterWrapper(BuildClass, Builder(DefaultTimeout))
}
