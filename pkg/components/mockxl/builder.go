// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package mockxl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	listbuilder "github.com/luids-io/xlist/pkg/builder"
)

// BuildClass defines default class for component builder
const BuildClass = "mock"

// Builder returns a list builder function that constructs mockups
func Builder() listbuilder.ListBuilder {
	return func(builder *listbuilder.Builder, parents []string, list listbuilder.ListDef) (xlist.Checker, error) {
		//create mockup and sets source
		mockup := &List{ResourceList: xlist.ClearResourceDups(list.Resources)}
		if list.Source != "" {
			results, err := sourceToResults(list.Source)
			if err != nil {
				return nil, errors.New("invalid 'source'")
			}
			mockup.Results = results
		}
		//config mockup
		if list.Opts != nil {
			err := configMockupFromOpts(mockup, list.Opts)
			if err != nil {
				return nil, err
			}
		}
		return mockup, nil
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
	listbuilder.Register(BuildClass, Builder())
}
