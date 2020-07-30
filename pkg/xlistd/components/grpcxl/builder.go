// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE..

package grpcxl

import (
	"errors"
	"fmt"

	checkapi "github.com/luids-io/api/xlist/grpc/check"
	"github.com/luids-io/core/grpctls"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "grpc"

// Builder resturns a rpcxl builder
func Builder() builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlistd.List, error) {
		if def.Source == "" {
			return nil, errors.New("source is empty")
		}
		clientCfg := grpctls.ClientCfg{}
		if def.TLS != nil {
			clientCfg.CertFile = b.CertPath(def.TLS.CertFile)
			clientCfg.KeyFile = b.CertPath(def.TLS.KeyFile)
			clientCfg.ServerName = def.TLS.ServerName
			clientCfg.ServerCert = b.CertPath(def.TLS.ServerCert)
			clientCfg.CACert = b.CertPath(def.TLS.CACert)
			clientCfg.UseSystemCAs = def.TLS.UseSystemCAs
		}
		err := clientCfg.Validate()
		if err != nil {
			return nil, fmt.Errorf("bad TLS config: %v", err)
		}
		dial, err := grpctls.Dial(def.Source, clientCfg)
		if err != nil {
			return nil, fmt.Errorf("dialing: %v", err)
		}
		bl := checkapi.NewClient(dial, def.Resources, checkapi.SetLogger(b.Logger()))
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
	builder.RegisterListBuilder(BuildClass, Builder())
}
