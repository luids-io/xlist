// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package geoip2xl

import (
	"errors"
	"fmt"
	"os"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		cfg := defaultCfg.Copy()
		if def.Source == "" {
			return nil, errors.New("'source' is required")
		}
		source := b.DataPath(def.Source)
		if !fileExists(source) {
			return nil, fmt.Errorf("geoip2 database file '%s' doesn't exists", source)
		}

		resources := xlist.ClearResourceDups(def.Resources, true)
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
		bl := New(def.ID, source, cfg, b.Logger())

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

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reason = reason
	}

	countries, ok, err := option.SliceString(opts, "countries")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Countries = make([]string, len(countries), len(countries))
		copy(dst.Countries, countries)
	}

	reverse, ok, err := option.Bool(opts, "reverse")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Reverse = reverse
	}

	return dst, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder(Config{}))
}
