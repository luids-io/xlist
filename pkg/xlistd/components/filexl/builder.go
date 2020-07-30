// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package filexl

import (
	"fmt"
	"os"
	"time"

	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/xlistd"
	"github.com/luids-io/xlist/pkg/xlistd/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "file"

// Builder returns a list builder function
func Builder(defaultCfg Config) builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlistd.List, error) {
		cfg := defaultCfg
		if def.Source == "" {
			def.Source = fmt.Sprintf("%s.xlist", def.ID)
		}
		source := b.SourcePath(def.Source)
		if !fileExists(source) {
			return nil, fmt.Errorf("file '%s' doesn't exists", source)
		}
		cfg.Logger = b.Logger()
		if def.Opts != nil {
			var err error
			cfg, err = parseOptions(cfg, def.Opts)
			if err != nil {
				return nil, err
			}
		}

		bl := New(def.ID, source, def.Resources, cfg)
		//register startup
		b.OnStartup(func() error {
			return bl.Open()
		})
		//register shutdown
		b.OnShutdown(func() error {
			bl.Close()
			return nil
		})

		return bl, nil
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil || os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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

	autoreload, ok, err := option.Bool(opts, "autoreload")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.Autoreload = autoreload
	}

	unsafereload, ok, err := option.Bool(opts, "unsafereload")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.UnsafeReload = unsafereload
	}

	reloadSecs, ok, err := option.Int(opts, "reloadseconds")
	if err != nil {
		return dst, err
	}
	if ok {
		dst.ReloadTime = time.Duration(reloadSecs) * time.Second
	}

	return dst, nil
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder(DefaultConfig()))
}
