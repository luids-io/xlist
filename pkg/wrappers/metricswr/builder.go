// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package metricswr

import (
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "metrics"

// Builder returns a builder for the component
func Builder(opt ...Option) listbuilder.WrapperBuilder {
	return func(b *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.Checker) (xlist.Checker, error) {
		return New(listID, bl), nil
	}
}

func init() {
	listbuilder.RegisterWrapper(BuildClass, Builder())
}
