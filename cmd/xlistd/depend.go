// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/luisguillenc/serverd"
	"github.com/luisguillenc/yalogi"

	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/core/xlist/check"
	iconfig "github.com/luids-io/xlist/internal/config"
	ifactory "github.com/luids-io/xlist/internal/factory"
	"github.com/luids-io/xlist/pkg/builder"

	//components
	_ "github.com/luids-io/xlist/pkg/components/dnsxl"
	_ "github.com/luids-io/xlist/pkg/components/filexl"
	_ "github.com/luids-io/xlist/pkg/components/geoip2xl"
	_ "github.com/luids-io/xlist/pkg/components/grpcxl"
	_ "github.com/luids-io/xlist/pkg/components/memxl"
	_ "github.com/luids-io/xlist/pkg/components/mockxl"
	_ "github.com/luids-io/xlist/pkg/components/parallelxl"
	_ "github.com/luids-io/xlist/pkg/components/selectorxl"
	_ "github.com/luids-io/xlist/pkg/components/sequencexl"
	_ "github.com/luids-io/xlist/pkg/components/wbeforexl"

	//wrappers
	_ "github.com/luids-io/xlist/pkg/wrappers/cachewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/loggerwr"
	_ "github.com/luids-io/xlist/pkg/wrappers/metricswr"
	_ "github.com/luids-io/xlist/pkg/wrappers/policywr"
	_ "github.com/luids-io/xlist/pkg/wrappers/ratewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/responsewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/scorewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/timeoutwr"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
}

func createBuilder(srv *serverd.Manager, logger yalogi.Logger) (*builder.Builder, error) {
	cfgBuild := cfg.Data("build").(*iconfig.BuilderCfg)
	builder, err := ifactory.Builder(cfgBuild, logger)
	if err != nil {
		return nil, err
	}
	srv.Register(serverd.Service{
		Name:     "xlist-builder.service",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createRootXList(builder *builder.Builder, srv *serverd.Manager) (xlist.Checker, error) {
	cfgXList := cfg.Data("xlist").(*iconfig.XListCfg)
	root, err := ifactory.RootXList(cfgXList, builder)
	if err != nil {
		return nil, err
	}
	srv.Register(serverd.Service{
		Name: "xlist-root.service",
		Ping: root.Ping,
	})
	return root, nil
}

func createXListSrv(rootList xlist.Checker, srv *serverd.Manager) error {
	//create server
	cfgServer := cfg.Data("grpc-check").(*cconfig.ServerCfg)
	glis, gsrv, err := cfactory.Server(cfgServer)
	if err != nil {
		return err
	}
	//create service
	cfgXList := cfg.Data("xlist").(*iconfig.XListCfg)
	service := check.NewService(rootList,
		check.DisclosureErrors(cfgXList.Disclosure),
		check.ExposePing(cfgXList.ExposePing),
	)
	// register service
	check.RegisterServer(gsrv, service)
	if cfgServer.Metrics {
		grpc_prometheus.Register(gsrv)
	}
	srv.Register(serverd.Service{
		Name:     "xlist.server",
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return nil
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
