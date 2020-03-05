// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package filexl

import (
	"fmt"
	"os"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines default class for component builder
const BuildClass = "file"

// Builder returns a list builder function
func Builder(opt ...Option) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
		if def.Source == "" {
			def.Source = fmt.Sprintf("%s.xlist", def.ID)
		}
		source := builder.SourcePath(def.Source)
		if !fileExists(source) {
			return nil, fmt.Errorf("file '%s' doesn't exists", source)
		}

		bopt := make([]Option, 0)
		bopt = append(bopt, SetLogger(builder.Logger()))
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		bl := New(source, def.Resources, bopt...)

		//register startup
		builder.OnStartup(func() error {
			builder.Logger().Debugf("starting '%s'", def.ID)
			return bl.Start()
		})

		//register shutdown
		builder.OnShutdown(func() error {
			builder.Logger().Debugf("shutting down '%s'", def.ID)
			bl.Shutdown()
			return nil
		})

		return bl, nil
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Reason(reason))
	}

	autoreload, ok, err := option.Bool(opts, "autoreload")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Autoreload(autoreload))
	}

	unsafereload, ok, err := option.Bool(opts, "unsafereload")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, UnsafeReload(unsafereload))
	}

	reloadSecs, ok, err := option.Int(opts, "reloadseconds")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, ReloadSeconds(reloadSecs))
	}

	return bopt, nil
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder())
}
