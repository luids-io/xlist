// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE..

package grpcxl

import (
	"errors"
	"fmt"

	"github.com/luisguillenc/grpctls"

	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/check"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "grpc"

// Builder resturns a rpcxl builder
func Builder(opt ...check.ClientOption) listbuilder.ListBuilder {
	return func(builder *listbuilder.Builder, parents []string, list listbuilder.ListDef) (xlist.Checker, error) {
		if list.Source == "" {
			return nil, errors.New("source is empty")
		}
		clientCfg := grpctls.ClientCfg{}
		if list.TLS != nil {
			clientCfg.CertFile = builder.CertPath(list.TLS.CertFile)
			clientCfg.KeyFile = builder.CertPath(list.TLS.KeyFile)
			clientCfg.ServerName = list.TLS.ServerName
			clientCfg.ServerCert = builder.CertPath(list.TLS.ServerCert)
			clientCfg.CACert = builder.CertPath(list.TLS.CACert)
			clientCfg.UseSystemCAs = list.TLS.UseSystemCAs
		}
		err := clientCfg.Validate()
		if err != nil {
			return nil, fmt.Errorf("bad TLS config: %v", err)
		}
		dial, err := grpctls.Dial(list.Source, clientCfg)
		if err != nil {
			return nil, fmt.Errorf("dialing: %v", err)
		}
		bl := check.NewClient(dial, list.Resources, opt...)
		if err != nil {
			return nil, fmt.Errorf("creating rpcxl: %v", err)
		}

		//register hooks
		builder.OnShutdown(func() error {
			builder.Logger().Debugf("shutting down grpc client '%s'", list.ID)
			return bl.Close()
		})
		return bl, nil
	}
}

func init() {
	listbuilder.Register(BuildClass, Builder())
}
