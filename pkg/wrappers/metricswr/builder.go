// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package metricswr

import (
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "metrics"

// Builder returns a builder for the component
func Builder() listbuilder.BuildWrapperFn {
	return func(b *listbuilder.Builder, id string, def listbuilder.WrapperDef, list xlist.List) (xlist.List, error) {
		return New(list, id), nil
	}
}

func init() {
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder())
}
