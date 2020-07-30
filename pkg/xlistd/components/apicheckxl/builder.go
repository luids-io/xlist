// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE..

package apicheckxl

import (
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "apicheck"

// Builder resturns a rpcxl builder
func Builder() builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlistd.List, error) {
		sname := def.Source
		if sname == "" {
			sname = def.ID
		}
		service, ok := b.APIService(sname)
		if !ok {
			return nil, fmt.Errorf("can't find")
		}
		if !ok {
			return nil, fmt.Errorf("can't find service '%s'", sname)
		}
		list, ok := service.(xlist.Checker)
		if !ok {
			return nil, fmt.Errorf("service '%s' is not an xlist.Checker", sname)
		}
		return &apicheckList{
			id:        def.ID,
			Checker:   list,
			resources: xlist.ClearResourceDups(def.Resources),
		}, nil
	}
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder())
}
