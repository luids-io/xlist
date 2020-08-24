// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package metricswr

import (
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)


// Builder returns a builder for the component
func Builder() builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlistd.List) (xlistd.List, error) {
		return New(list), nil
	}
}

func init() {
	builder.RegisterWrapperBuilder(WrapperClass, Builder())
}
