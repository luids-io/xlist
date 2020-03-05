// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package ratewr

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/luids-io/core/option"
	"github.com/luids-io/core/xlist"
	"github.com/luids-io/xlist/pkg/listbuilder"
)

// BuildClass defines class name for component builder
const BuildClass = "rate"

// Builder returns a builder
func Builder(mode string, opt ...Option) listbuilder.BuildWrapperFn {
	return func(builder *listbuilder.Builder, listID string, def listbuilder.WrapperDef, bl xlist.List) (xlist.List, error) {
		bopt := make([]Option, 0)
		bopt = append(bopt, opt...)

		var limiter Limiter
		// get limiter mode from opts
		m, ok, err := option.String(def.Opts, "mode")
		if err != nil {
			return nil, err
		}
		if ok {
			mode = m
		}
		switch strings.ToLower(mode) {
		case "naive":
			bucket, err := createNaiveFromOpts(def.Opts)
			if err != nil {
				return nil, fmt.Errorf("can't construct 'naive': %v", err)
			}
			builder.OnStartup(bucket.Start)
			builder.OnShutdown(bucket.Shutdown)

			limiter = bucket
		//TODO: implement build for limiter "golang.org/x/time/rate"
		default:
			return nil, errors.New("invalid 'mode'")
		}

		// get wrapper common settings
		if def.Opts != nil {
			var err error
			bopt, err = parseOptions(bopt, def.Opts)
			if err != nil {
				return nil, err
			}
		}
		return New(limiter, bl, bopt...), nil
	}
}

func createNaiveFromOpts(opts map[string]interface{}) (*NaiveBucket, error) {
	//get requests
	requests, ok, err := option.Int(opts, "requests")
	if err != nil {
		return nil, err
	}
	// by default capacity = requests
	capacity := requests
	c, ok, err := option.Int(opts, "capacity")
	if err != nil {
		return nil, err
	}
	if ok {
		capacity = c
	}
	// by default interval = 1 second
	interval := time.Second
	seconds, ok, err := option.Int(opts, "seconds")
	if err != nil {
		return nil, err
	}
	if ok {
		interval = time.Duration(seconds) * time.Second
	}

	return &NaiveBucket{
		Capacity:     capacity,
		DripInterval: interval,
		PerDrip:      requests,
	}, nil
}

func parseOptions(bopt []Option, opts map[string]interface{}) ([]Option, error) {
	wait, ok, err := option.Bool(opts, "wait")
	if err != nil {
		return bopt, err
	}
	if ok {
		bopt = append(bopt, Wait(wait))
	}
	return bopt, nil
}

func init() {
	listbuilder.RegisterWrapperBuilder(BuildClass, Builder("naive"))
}
