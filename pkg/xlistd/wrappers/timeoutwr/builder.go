// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package timeoutwr

import (
	"time"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)

// DefaultTimeout sets default timeout in construction
const DefaultTimeout = 1 * time.Second

// Builder returns a builder for the component
func Builder(timeout time.Duration) builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlistd.List) (xlistd.List, error) {
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
	builder.RegisterWrapperBuilder(WrapperClass, Builder(DefaultTimeout))
}
