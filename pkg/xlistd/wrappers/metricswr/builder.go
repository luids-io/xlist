// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package metricswr

import (
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder() xlistd.BuildWrapperFn {
	return func(b *xlistd.Builder, def xlistd.WrapperDef, list xlistd.List) (xlistd.List, error) {
		return New(list), nil
	}
}

func init() {
	xlistd.RegisterWrapperBuilder(WrapperClass, Builder())
}
