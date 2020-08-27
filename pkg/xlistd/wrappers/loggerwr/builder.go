// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package loggerwr

import (
	"errors"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// Builder returns a builder function.
func Builder(defaultCfg Config) xlistd.BuildWrapperFn {
	return func(b *xlistd.Builder, def xlistd.WrapperDef, list xlistd.List) (xlistd.List, error) {
		cfg := defaultCfg
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		cfg.Prefix = list.ID()
		return New(list, b.Logger(), cfg), nil
	}
}

func parseOptions(src Config, opts map[string]interface{}) (Config, error) {
	dst := src
	showpeer, ok, err := option.Bool(opts, "showpeer")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.ShowPeer = showpeer
	}

	found, ok, err := option.String(opts, "found")
	if err != nil {
		return dst, err
	}
	if ok {
		lfound, err := StringToLevel(found)
		if err != nil {
			return dst, errors.New("invalid 'found'")
		}
		dst.Rules.Found = lfound
	}

	notfound, ok, err := option.String(opts, "notfound")
	if err != nil {
		return dst, err
	}
	if ok {
		lnotfound, err := StringToLevel(notfound)
		if err != nil {
			return dst, errors.New("invalid 'notfound'")
		}
		dst.Rules.NotFound = lnotfound
	}

	errorlevel, ok, err := option.String(opts, "error")
	if err != nil {
		return dst, err
	}
	if ok {
		lerror, err := StringToLevel(errorlevel)
		if err != nil {
			return dst, errors.New("invalid 'error'")
		}
		dst.Rules.Error = lerror
	}

	return dst, nil
}

func init() {
	xlistd.RegisterWrapperBuilder(WrapperClass, Builder(DefaultConfig()))
}
