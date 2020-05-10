// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package mockxl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/core/option"
	"github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "mock"

// Builder returns a list builder function that constructs mockups
func Builder() builder.BuildListFn {
	return func(b *builder.Builder, parents []string, def builder.ListDef) (xlist.List, error) {
		//create mockup and sets source
		bl := &List{ResourceList: xlist.ClearResourceDups(def.Resources)}
		if def.Source != "" {
			results, err := sourceToResults(def.Source)
			if err != nil {
				return nil, errors.New("invalid 'source'")
			}
			bl.Results = results
		}
		//config mockup
		if def.Opts != nil {
			err := configMockupFromOpts(bl, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return bl, nil
	}
}

func sourceToResults(s string) ([]bool, error) {
	if s == "" {
		return []bool{}, nil
	}
	tokens := strings.Split(s, ",")
	results := make([]bool, 0, len(tokens))
	for _, res := range tokens {
		res = strings.Trim(res, " ")
		switch res {
		case "true":
			results = append(results, true)
		case "false":
			results = append(results, false)
		default:
			return nil, fmt.Errorf("invalid value '%s' for result", res)
		}
	}
	return results, nil
}

func configMockupFromOpts(mockup *List, opts map[string]interface{}) error {
	fail, ok, err := option.Bool(opts, "fail")
	if err != nil {
		return err
	}
	if ok {
		mockup.Fail = fail
	}
	reason, ok, err := option.String(opts, "reason")
	if err != nil {
		return err
	}
	if ok {
		mockup.Reason = reason
	}
	ttl, ok, err := option.Int(opts, "ttl")
	if err != nil {
		return err
	}
	if ok {
		mockup.TTL = ttl
	}
	lazy, ok, err := option.Int(opts, "lazy")
	if err != nil {
		return err
	}
	if ok {
		mockup.Lazy = time.Duration(lazy) * time.Millisecond
	}
	return nil
}

func init() {
	builder.RegisterListBuilder(BuildClass, Builder())
}
