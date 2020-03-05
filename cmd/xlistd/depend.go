// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// dependency injection functions

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/luisguillenc/serverd"
	"github.com/luisguillenc/yalogi"
	"google.golang.org/grpc"

	checkapi "github.com/luids-io/api/xlist/check"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/xlist"
	iconfig "github.com/luids-io/xlist/internal/config"
	ifactory "github.com/luids-io/xlist/internal/factory"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
}

func createHealthSrv(srv *serverd.Manager, logger yalogi.Logger) error {
	cfgHealth := cfg.Data("health").(*cconfig.HealthCfg)
	if !cfgHealth.Empty() {
		hlis, health, err := cfactory.Health(cfgHealth, srv, logger)
		if err != nil {
			logger.Fatalf("creating health server: %v", err)
		}
		srv.Register(serverd.Service{
			Name:     "health.server",
			Start:    func() error { go health.Serve(hlis); return nil },
			Shutdown: func() { health.Close() },
		})
	}
	return nil
}

func createCheckSrv(srv *serverd.Manager, logger yalogi.Logger) (*grpc.Server, error) {
	cfgServer := cfg.Data("grpc-check").(*cconfig.ServerCfg)
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err != nil {
		return nil, err
	}
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	srv.Register(serverd.Service{
		Name:     "grpc-check.server",
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, nil
}

func createLists(srv *serverd.Manager, logger yalogi.Logger) (*listbuilder.Builder, error) {
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
	srv.Register(serverd.Service{
		Name:     "lists.service",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createCheckAPIService(gsrv *grpc.Server, finder xlist.ListFinder, srv *serverd.Manager, logger yalogi.Logger) error {
	cfgCheck := cfg.Data("api-check").(*iconfig.APICheckCfg)
	gsvc, err := ifactory.CheckAPIService(cfgCheck, finder, logger)
	if err != nil {
		return fmt.Errorf("creating checkapi service: %v", err)
	}
	checkapi.RegisterServer(gsrv, gsvc)
	//get root list to monitor
	rootList, ok := finder.FindListByID(cfgCheck.RootListID)
	if !ok {
		return fmt.Errorf("rootlist '%s' not found", cfgCheck.RootListID)
	}
	srv.Register(serverd.Service{
		Name: "checkapi.service",
		Ping: rootList.Ping,
	})
	return nil
}
