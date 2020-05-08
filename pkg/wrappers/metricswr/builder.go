// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package metricswr

import (
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "metrics"

// Builder returns a builder for the component
func Builder() builder.BuildWrapperFn {
	return func(b *builder.Builder, def builder.WrapperDef, list xlist.List) (xlist.List, error) {
		return New(list), nil
	}
}

func init() {
	builder.RegisterWrapperBuilder(BuildClass, Builder())
}
