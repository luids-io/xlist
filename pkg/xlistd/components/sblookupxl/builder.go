// Copyright 2020 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package sblookupxl

import (
	"errors"
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildListFn {
	return func(b *xlistd.Builder, parents []string, def xlistd.ListDef) (xlistd.List, error) {
		//get cache file
		dbname := fmt.Sprintf("%s.db", def.ID)
		if def.Source != "" {
			dbname = def.Source
		}
		cfg := defaultCfg
		cfg.Database = b.DataPath(dbname)
		//check resources
		resources := xlist.ClearResourceDups(def.Resources, true)
		if len(resources) != 1 || resources[0] != xlist.Domain {
			return nil, errors.New("invalid 'resources': sblookup only supports domain resource")
		}
		//get options
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		//create RBL list
		bl, err := New(def.ID, cfg)
		if err != nil {
			return nil, err
		}
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

	apikey, ok, err := option.String(opts, "apikey")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.APIKey = apikey
	}

	serverurl, ok, err := option.String(opts, "serverurl")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.ServerURL = serverurl
	}

	threats, ok, err := option.SliceString(opts, "threats")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Threats = make([]string, 0, len(threats))
		for _, t := range threats {
			dst.Threats = append(dst.Threats, t)
		}
	}
	return dst, nil
}

func init() {
	xlistd.RegisterListBuilder(ComponentClass, Builder(Config{}))
}
