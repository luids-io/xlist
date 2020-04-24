// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package loggerwr

import (
	"errors"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines class name for component builder
const BuildClass = "logger"

// Builder returns a builder for the wrapper component with the logger passed
func Builder(cfg Config) builder.BuildWrapperFn {
	return func(b *builder.Builder, id string, def builder.WrapperDef, list xlist.List) (xlist.List, error) {
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		cfg.Prefix = id
		return New(list, b.Logger(), cfg), nil
	}
}

func parseOptions(cfg Config, opts map[string]interface{}) (Config, error) {
	rCfg := cfg
	showpeer, ok, err := option.Bool(opts, "showpeer")
	if err != nil {
		return rCfg, err
	}
	if ok {
		rCfg.ShowPeer = showpeer
	}

	found, ok, err := option.String(opts, "found")
	if err != nil {
		return rCfg, err
	}
	if ok {
		lfound, err := StringToLevel(found)
		if err != nil {
			return rCfg, errors.New("invalid 'found'")
		}
		rCfg.Rules.Found = lfound
	}

	notfound, ok, err := option.String(opts, "notfound")
	if err != nil {
		return rCfg, err
	}
	if ok {
		lnotfound, err := StringToLevel(notfound)
		if err != nil {
			return rCfg, errors.New("invalid 'notfound'")
		}
		rCfg.Rules.NotFound = lnotfound
	}

	errorlevel, ok, err := option.String(opts, "error")
	if err != nil {
		return rCfg, err
	}
	if ok {
		lerror, err := StringToLevel(errorlevel)
		if err != nil {
			return rCfg, errors.New("invalid 'error'")
		}
		rCfg.Rules.Error = lerror
	}

	return rCfg, nil
}

func init() {
	builder.RegisterWrapperBuilder(BuildClass, Builder(DefaultConfig()))
}
