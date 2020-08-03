// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	cconfig "github.com/luids-io/common/config"
	"github.com/luids-io/core/goconfig"
	iconfig "github.com/luids-io/xlist/internal/config"
)

// Default returns the default configuration
func Default(program string) *goconfig.Config {
	cfg, err := goconfig.New(program,
		goconfig.Section{
			Name:     "xlist",
			Required: true,
			Short:    true,
			Data:     &iconfig.XListCfg{},
		},
		goconfig.Section{
			Name:     "xlist.plugin.dnsxl",
			Required: false,
			Short:    false,
			Data:     &iconfig.DNSxLCfg{},
		},
		goconfig.Section{
			Name:     "api.xlist.check",
			Required: true,
			Short:    true,
			Data: &iconfig.XListCheckAPICfg{
				Enable:     true,
				Log:        true,
				RootListID: "root",
			},
		},
		goconfig.Section{
			Name:     "apiservice",
			Required: false,
			Data:     &cconfig.APIServicesCfg{},
		},
		goconfig.Section{
			Name:     "apiservice.event",
			Required: false,
			Data: &cconfig.EventNotifyCfg{
				Buffer: 100,
			},
		},
		goconfig.Section{
			Name:     "server",
			Required: true,
			Short:    true,
			Data: &cconfig.ServerCfg{
				ListenURI: "tcp://127.0.0.1:5801",
			},
		},
		goconfig.Section{
			Name:     "log",
			Required: true,
			Data: &cconfig.LoggerCfg{
				Level: "info",
			},
		},
		goconfig.Section{
			Name:     "health",
			Required: false,
			Data:     &cconfig.HealthCfg{},
		},
	)
	if err != nil {
		panic(err)
	}
	return cfg
}
