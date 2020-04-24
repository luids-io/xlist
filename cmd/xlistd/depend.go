// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// dependency injection functions

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	checkapi "github.com/luids-io/api/xlist/check"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/utils/serverd"
	"github.com/luids-io/core/utils/yalogi"
	"github.com/luids-io/core/xlist"
	iconfig "github.com/luids-io/xlist/internal/config"
	ifactory "github.com/luids-io/xlist/internal/factory"
	"github.com/luids-io/xlist/pkg/builder"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
}

func createHealthSrv(msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgHealth := cfg.Data("health").(*cconfig.HealthCfg)
	if !cfgHealth.Empty() {
		hlis, health, err := cfactory.Health(cfgHealth, msrv, logger)
		if err != nil {
			return err
		}
		msrv.Register(serverd.Service{
			Name:     "health.server",
			Start:    func() error { go health.Serve(hlis); return nil },
			Shutdown: func() { health.Close() },
		})
	}
	return nil
}

func createLists(msrv *serverd.Manager, logger yalogi.Logger) (*builder.Builder, error) {
	cfgList := cfg.Data("xlist").(*iconfig.XListCfg)
	builder, err := ifactory.ListBuilder(cfgList, logger)
	if err != nil {
		return nil, err
	}
	//create lists
	err = ifactory.Lists(cfgList, builder, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "xlist-database.service",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createCheckAPI(gsrv *grpc.Server, finder xlist.ListFinder, msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgCheck := cfg.Data("api-check").(*iconfig.XListCheckAPICfg)
	gsvc, err := ifactory.XListCheckAPI(cfgCheck, finder, logger)
	if err != nil {
		return err
	}
	checkapi.RegisterServer(gsrv, gsvc)
	//get root list to monitor
	rootList, ok := finder.FindListByID(cfgCheck.RootListID)
	if !ok {
		return fmt.Errorf("rootlist '%s' not found", cfgCheck.RootListID)
	}
	msrv.Register(serverd.Service{
		Name: "xlist-check.service",
		Ping: rootList.Ping,
	})
	return nil
}

func createCheckSrv(msrv *serverd.Manager) (*grpc.Server, error) {
	cfgServer := cfg.Data("server-check").(*cconfig.ServerCfg)
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err != nil {
		return nil, err
	}
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	msrv.Register(serverd.Service{
		Name:     fmt.Sprintf("[%s].server", cfgServer.ListenURI),
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, nil
}
