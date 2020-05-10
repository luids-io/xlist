// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

// dependency injection functions

import (
	"fmt"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	"github.com/luids-io/api/event"
	"github.com/luids-io/api/event/notifybuffer"
	"github.com/luids-io/api/xlist"
	checkapi "github.com/luids-io/api/xlist/grpc/check"
	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	"github.com/luids-io/core/apiservice"
	"github.com/luids-io/core/serverd"
	"github.com/luids-io/core/yalogi"
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

func createAPIServices(msrv *serverd.Manager, logger yalogi.Logger) (apiservice.Discover, error) {
	cfgServices := cfg.Data("ids.api").(*cconfig.APIServicesCfg)
	registry, err := cfactory.APIAutoloader(cfgServices, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "apiservices.service",
		Ping:     registry.Ping,
		Shutdown: func() { registry.CloseAll() },
	})
	return registry, nil
}

func setupEventNotify(registry apiservice.Discover, msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgEvent := cfg.Data("ids.event").(*cconfig.EventNotifyCfg)
	if !cfgEvent.Empty() {
		client, err := cfactory.EventNotify(cfgEvent, registry)
		if err != nil {
			return err
		}
		ebuffer := notifybuffer.New(client, cfgEvent.Buffer, notifybuffer.SetLogger(logger))
		msrv.Register(serverd.Service{
			Name:     "event-notify.service",
			Shutdown: func() { ebuffer.Close() },
		})
		event.SetBuffer(ebuffer)
	}
	return nil
}

func createLists(apisvc apiservice.Discover, msrv *serverd.Manager, logger yalogi.Logger) (*builder.Builder, error) {
	cfgList := cfg.Data("xlist").(*iconfig.XListCfg)
	builder, err := ifactory.ListBuilder(cfgList, apisvc, logger)
	if err != nil {
		return nil, err
	}
	//create lists
	err = ifactory.Lists(cfgList, builder, logger)
	if err != nil {
		return nil, err
	}
	msrv.Register(serverd.Service{
		Name:     "xlist",
		Start:    builder.Start,
		Shutdown: func() { builder.Shutdown() },
	})
	return builder, nil
}

func createCheckAPI(gsrv *grpc.Server, finder xlist.ListFinder, msrv *serverd.Manager, logger yalogi.Logger) error {
	cfgCheck := cfg.Data("xlist.api.check").(*iconfig.XListCheckAPICfg)
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
		Name: "xlist.api.check",
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
		Name:     fmt.Sprintf("[%s].server", cfgServer.ListenURI),
		Start:    func() error { go gsrv.Serve(glis); return nil },
		Shutdown: gsrv.GracefulStop,
		Stop:     gsrv.Stop,
	})
	return gsrv, nil
}
