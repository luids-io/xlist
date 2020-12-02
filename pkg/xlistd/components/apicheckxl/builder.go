// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE..

package apicheckxl

import (
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder() xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		sname := def.Source
		if sname == "" {
			sname = def.ID
		}
		service, ok := b.APIService(sname)
		if !ok {
			return nil, fmt.Errorf("can't find service '%s'", sname)
		}
		list, ok := service.(xlist.Checker)
		if !ok {
			return nil, fmt.Errorf("service '%s' is not an xlist.Checker", sname)
		}
		return &apicheckList{
			id:        def.ID,
			checker:   list,
			resources: xlist.ClearResourceDups(def.Resources, true),
		}, nil
	}
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder())
}
