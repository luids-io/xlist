// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package loggerwr

import (
	"errors"

	"github.com/luids-io/core/utils/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "logger"

// Builder returns a builder for the wrapper component with the logger passed
func Builder(opt ...Option) listbuilder.BuildWrapperFn {
	return func(builder *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.List) (xlist.List, error) {
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)

		rules := DefaultRules()
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
			r, err := getRulesFromOpts(def.Opts)
			if err != nil {
				return nil, err
			}
			rules = r
		}
		return New(listID, bl, builder.Logger(), rules, bopt...), nil
	}
}

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	showpeer, ok, err := option.Bool(opts, "showpeer")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, ShowPeer(showpeer))
	}
	return bopt, nil
}

func getRulesFromOpts(opts map[string]interface{}) (Rules, error) {
	rules := DefaultRules()

	found, ok, err := option.String(opts, "found")
	if err != nil {
		return rules, err
	}
	if ok {
		lfound, err := StringToLevel(found)
		if err != nil {
			return rules, errors.New("invalid 'found'")
		}
		rules.Found = lfound
	}

	notfound, ok, err := option.String(opts, "notfound")
	if err != nil {
		return rules, err
	}
	if ok {
		lnotfound, err := StringToLevel(notfound)
		if err != nil {
			return rules, errors.New("invalid 'notfound'")
		}
		rules.NotFound = lnotfound
	}

	errorlevel, ok, err := option.String(opts, "error")
	if err != nil {
		return rules, err
	}
	if ok {
		lerror, err := StringToLevel(errorlevel)
		if err != nil {
			return rules, errors.New("invalid 'error'")
		}
		rules.Error = lerror
	}

	return rules, nil
}

func init() {
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder())
}
