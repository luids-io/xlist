// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package config

import (
	cconfig "github.com/luids-io/common/config"
	iconfig "github.com/luids-io/xlist/internal/config"
	"github.com/luisguillenc/goconfig"
)

// Default returns the default configuration
func Default(program string) *goconfig.Config {
	cfg, err := goconfig.New(program,
		goconfig.Section{
			Name:     "api-check",
			Required: true,
			Short:    true,
			Data: &iconfig.APICheckCfg{
				RootListID: "root",
			},
		},
		goconfig.Section{
			Name:     "xlist",
			Required: true,
			Short:    true,
			Data:     &iconfig.XListCfg{},
		},
		goconfig.Section{
			Name:     "grpc-check",
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
