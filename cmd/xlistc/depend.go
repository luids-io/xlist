// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"github.com/luids-io/api/xlist/grpc/check"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/yalogi"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
}

func createClient(logger yalogi.Logger) (*check.Client, error) {
	//create dial
	cfgDial := cfg.Data("client").(*cconfig.ClientCfg)
	dial, err := cfactory.ClientConn(cfgDial)
	if err != nil {
		return nil, err
	}
	//create grpc client
	client := check.NewClient(dial, check.SetLogger(logger))
	return client, nil
}
