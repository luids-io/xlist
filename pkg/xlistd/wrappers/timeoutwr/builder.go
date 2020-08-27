// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package timeoutwr

import (
	"time"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// DefaultTimeout sets default timeout in construction.
const DefaultTimeout = 1 * time.Second

// Builder returns a builder function.
func Builder(timeout time.Duration) xlistd.BuildWrapperFn {
	return func(b *xlistd.Builder, def xlistd.WrapperDef, list xlistd.List) (xlistd.List, error) {
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
	xlistd.RegisterWrapperBuilder(WrapperClass, Builder(DefaultTimeout))
}
