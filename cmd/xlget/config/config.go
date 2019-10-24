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
			Name:     "xlget",
			Required: true,
			Short:    true,
			Data:     &iconfig.XLGetCfg{},
		},
		goconfig.Section{
			Name:     "log",
			Required: true,
			Data: &cconfig.LoggerCfg{
				Level: "info",
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return cfg
}
