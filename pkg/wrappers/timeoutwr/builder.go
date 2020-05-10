// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package timeoutwr

import (
	"time"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/builder"
)

// DefaultTimeout sets default timeout in construction
const DefaultTimeout = 1 * time.Second

// BuildClass defines class name for component builder
const BuildClass = "timeout"

// Builder returns a builder for the component
func Builder(timeout time.Duration) builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlist.List) (xlist.List, error) {
		if def.Opts != nil {
			v, ok, err := option.Int(def.Opts, "timeout")
			if err != nil {
				return nil, err
			}
			if ok {
				timeout = time.Duration(v) * time.Millisecond
			}
		}
		return New(list, timeout), nil
	}
}

func init() {
	builder.RegisterWrapperBuilder(BuildClass, Builder(DefaultTimeout))
}
