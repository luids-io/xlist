// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package timeoutwr

import (
	"time"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// DefaultTimeout sets default timeout in construction
const DefaultTimeout = 1 * time.Second

// BuildClass defines class name for component builder
const BuildClass = "timeout"

// Builder returns a builder for the component
func Builder(timeout time.Duration) listbuilder.BuildWrapperFn {
	return func(b *listbuilder.Builder, id string, def listbuilder.WrapperDef, list xlist.List) (xlist.List, error) {
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
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder(DefaultTimeout))
}
