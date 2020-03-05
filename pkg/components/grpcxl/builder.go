// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE..

package grpcxl

import (
	"errors"
	"fmt"

	"github.com/luisguillenc/grpctls"

	"github.com/luids-io/api/xlist/check"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "grpc"

// Builder resturns a rpcxl builder
func Builder(opt ...check.ClientOption) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
		if def.Source == "" {
			return nil, errors.New("source is empty")
		}
		clientCfg := grpctls.ClientCfg{}
		if def.TLS != nil {
			clientCfg.CertFile = builder.CertPath(def.TLS.CertFile)
			clientCfg.KeyFile = builder.CertPath(def.TLS.KeyFile)
			clientCfg.ServerName = def.TLS.ServerName
			clientCfg.ServerCert = builder.CertPath(def.TLS.ServerCert)
			clientCfg.CACert = builder.CertPath(def.TLS.CACert)
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
		bl := check.NewClient(dial, def.Resources, opt...)
		if err != nil {
			return nil, fmt.Errorf("creating rpcxl: %v", err)
		}

		//register hooks
		builder.OnShutdown(func() error {
			builder.Logger().Debugf("shutting down grpc client '%s'", def.ID)
			return bl.Close()
		})
		return &grpclist{Checker: bl}, nil
	}
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder())
}
