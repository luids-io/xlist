// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package geoip2xl

import (
	"errors"
	"fmt"
	"os"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "geoip2"

// Builder returns a list builder function
func Builder(cfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		if def.Source == "" {
			return nil, errors.New("'source' is required")
		}
		source := b.SourcePath(def.Source)
		if !fileExists(source) {
			return nil, fmt.Errorf("geoip2 database file '%s' doesn't exists", source)
		}

		resources := xlist.ClearResourceDups(def.Resources)
		if len(resources) != 1 || resources[0] != xlist.IPv4 {
			return nil, errors.New("invalid 'resources': geoip2 only supports ip4")
		}

		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}

		//create RBL list
		bl := New(source, cfg)

		//register startup
		b.OnStartup(func() error {
			b.Logger().Debugf("starting '%s'", def.ID)
			return bl.Open()
		})

		//register shutdown
		b.OnShutdown(func() error {
			b.Logger().Debugf("shutting down '%s'", def.ID)
			bl.Close()
			return nil
		})

		return bl, nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Reason = reason
	}

	countries, ok, err := option.SliceString(opts, "countries")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Countries = make([]string, len(countries), len(countries))
		copy(rCfg.Countries, countries)
	}

	reverse, ok, err := option.Bool(opts, "reverse")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.Reverse = reverse
	}

	return rCfg, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder(Config{}))
}
