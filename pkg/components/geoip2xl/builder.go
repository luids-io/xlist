// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package geoip2xl

import (
	"errors"
	"fmt"
	"os"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "geoip2"

// Builder returns a list builder function
func Builder(opt ...Option) listbuilder.BuildListFn {
	return func(builder *listbuilder.Builder, parents []string, def listbuilder.ListDef) (xlist.List, error) {
		if def.Source == "" {
			return nil, errors.New("'source' is required")
		}
		source := builder.SourcePath(def.Source)
		if !fileExists(source) {
			return nil, fmt.Errorf("geoip2 database file '%s' doesn't exists", source)
		}
		resources := xlist.ClearResourceDups(def.Resources)
		if len(resources) != 1 || resources[0] != xlist.IPv4 {
			return nil, errors.New("invalid 'resources': geoip2 only supports ip4")
		}

		var rules Rules
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
			rules, err = getRulesFromOpts(def.Opts)
			if err != nil {
				return nil, err
			}
		}

		//create RBL list
		bl := New(source, rules, bopt...)

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
	return bopt, nil
}

func getRulesFromOpts(opts map[string]interface{}) (Rules, error) {
	rules := Rules{Countries: make([]string, 0)}
	//config rules
	countries, ok, err := option.SliceString(opts, "countries")
	if err != nil {
		return rules, err
	}
	if ok {
		for _, country := range countries {
			rules.Countries = append(rules.Countries, country)
		}
	}
	reverse, ok, err := option.Bool(opts, "reverse")
	if err != nil {
		return rules, err
	}
	if ok {
		rules.Reverse = reverse
	}
	return rules, nil
}

func init() {
	listbuilder.RegisterListBuilder(BuildClass, Builder())
}
