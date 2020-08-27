// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// dependency injection functions

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	checkapi "github.com/luids-io/api/xlist/grpc/check"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/serverd"
	"github.com/luids-io/core/yalogi"
	iconfig "github.com/luids-io/xlist/internal/config"
	ifactory "github.com/luids-io/xlist/internal/factory"
	"github.com/luids-io/xlist/pkg/xlistd"
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
			Name:     fmt.Sprintf("health.[%s]", cfgHealth.ListenURI),
			Start:    func() error { go health.Serve(hlis); return nil },
			Shutdown: func() { health.Close() },
		})
	}
	return nil
}

func createAPIServices(msrv *serverd.Manager, logger yalogi.Logger) (apiservice.Discover, error) {
	cfgServices := cfg.Data("ids.api").(*cconfig.APIServicesCfg)
	registry, err := cfactory.APIAutoloader(cfgServices, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "ids.api",
		Ping:     registry.Ping,
		Shutdown: func() { registry.CloseAll() },
	})
	return registry, nil
}

func createLists(apisvc apiservice.Discover, msrv *serverd.Manager, logger yalogi.Logger) (*xlistd.Builder, error) {
	cfgList := cfg.Data("xlistd").(*iconfig.XListCfg)
	builder, err := ifactory.ListBuilder(cfgList, apisvc, logger)
	if err != nil {
		return nil, err
	}
	//setup plugins
	cfgDNSxl := cfg.Data("xlistd.plugin.dnsxl").(*iconfig.DNSxLCfg)
	if !cfgDNSxl.Empty() {
		err := ifactory.SetupDNSxL(cfgDNSxl)
		if err != nil {
			return nil, err
		}
	}
	//create lists
	err = ifactory.Lists(cfgList, builder, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "xlistd.service",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createCheckAPI(gsrv *grpc.Server, finder *xlistd.Builder, msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgCheck := cfg.Data("service.xlist.check").(*iconfig.XListCheckAPICfg)
	gsvc, err := ifactory.XListCheckAPI(cfgCheck, finder, logger)
	if err != nil {
		return err
	}
	checkapi.RegisterServer(gsrv, gsvc)
	//get root list to monitor
	rootList, ok := finder.List(cfgCheck.RootListID)
	if !ok {
		return fmt.Errorf("rootlist '%s' not found", cfgCheck.RootListID)
	}
	msrv.Register(serverd.Service{
		Name: "service.xlist.check",
		Ping: rootList.Ping,
	})
	return nil
}

func createServer(msrv *serverd.Manager) (*grpc.Server, error) {
	cfgServer := cfg.Data("server").(*cconfig.ServerCfg)
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err != nil {
		return nil, err
	}
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	msrv.Register(serverd.Service{
		Name:     fmt.Sprintf("server.[%s]", cfgServer.ListenURI),
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, nil
}
