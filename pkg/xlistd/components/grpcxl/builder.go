// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE..

package grpcxl

import (
	"errors"
	"fmt"

	checkapi "github.com/luids-io/api/xlist/grpc/check"
	"github.com/luids-io/core/grpctls"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder() xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		if def.Source == "" {
			return nil, errors.New("source is empty")
		}
		clientCfg := def.ClientCfg()
		err := clientCfg.Validate()
		if err != nil {
			return nil, fmt.Errorf("bad TLS config: %v", err)
		}
		dial, err := grpctls.Dial(def.Source, clientCfg)
		if err != nil {
			return nil, fmt.Errorf("dialing: %v", err)
		}
		bl := checkapi.NewClient(dial, checkapi.SetLogger(b.Logger()))
		if err != nil {
			return nil, fmt.Errorf("creating rpcxl: %v", err)
		}

		//register hooks
		b.OnShutdown(func() error {
			return bl.Close()
		})
		return &grpclist{id: def.ID, Checker: bl}, nil
	}
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder())
}
