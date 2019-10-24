// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	"github.com/luisguillenc/yalogi"

	cconfig "github.com/luids-io/common/config"
	cfactory "github.com/luids-io/common/factory"
	iconfig "github.com/luids-io/xlist/internal/config"
	ifactory "github.com/luids-io/xlist/internal/factory"
	"github.com/luids-io/xlist/pkg/xlget"
)

func createLogger(debug bool) (yalogi.Logger, error) {
	cfgLog := cfg.Data("log").(*cconfig.LoggerCfg)
	return cfactory.Logger(cfgLog, debug)
}

func createManager(logger yalogi.Logger) (*xlget.Manager, error) {
	cfgGet := cfg.Data("xlget").(*iconfig.XLGetCfg)
	return ifactory.XLGet(cfgGet, logger)
}
