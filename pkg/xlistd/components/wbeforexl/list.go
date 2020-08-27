// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package wbeforexl provides a simple xlistd.List implementation that can
// be used to check on a white list before checking on a blacklist.
//
// This package is a work in progress and makes no API stability promises.
package wbeforexl

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/api/xlist"
	"github.com/luids-io/xlist/pkg/xlistd"
)

// ComponentClass registered.
const ComponentClass = "wbefore"

// Config options.
type Config struct {
	ForceValidation bool
	Reason          string
}

// List implements a composite RBL that checks a whtelist before checking
// the blacklist. This means that if the checked resource exists in the
// whitelist, then it returns immediately with a negative result. If not
// in the whitelist, then returns the response of the blacklist.
type List struct {
	id           string
	cfg          Config
	provides     []bool
	resources    []xlist.Resource
	white, black xlistd.List
}

// New constructs a new "white before" RBL, it receives the resource list that
// RBL supports.
func New(id string, white, black xlistd.List, resources []xlist.Resource, cfg Config) *List {
	l := &List{
		id:        id,
		cfg:       cfg,
		white:     white,
		black:     black,
		resources: xlist.ClearResourceDups(resources, true),
		provides:  make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range l.resources {
		l.provides[int(r)] = true
	}
	return l
}

// ID implements xlistd.List interface.
func (l *List) ID() string {
	return l.id
}

// Class implements xlistd.List interface.
func (l *List) Class() string {
	return ComponentClass
}

// Check implements xlist.Checker interface.
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotSupported
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, l.cfg.ForceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	if l.white != nil {
		resp, err := l.white.Check(ctx, name, resource)
		if err != nil {
			return xlist.Response{}, err
		}
		if resp.Result {
			return xlist.Response{}, nil
		}
	}
	select {
	case <-ctx.Done():
		return xlist.Response{}, xlist.ErrCanceledRequest
	default:
		if l.black != nil {
			resp, err := l.black.Check(ctx, name, resource)
			if err == nil && resp.Result && l.cfg.Reason != "" {
				resp.Reason = l.cfg.Reason
			}
			return resp, err
		}
		return xlist.Response{}, nil
	}
}

// Resources implements xlist.Checker interface.
func (l *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(resources, l.resources)
	return resources
}

// Ping implements xlist.Checker interface.
func (l *List) Ping() error {
	var errWhite, errBlack error
	if l.white != nil {
		errWhite = l.white.Ping()
	}
	if l.black != nil {
		errBlack = l.black.Ping()
	}
	if errWhite != nil || errBlack != nil {
		var msgErr string
		if errWhite != nil {
			msgErr = fmt.Sprintf("%s: %v", l.white.ID(), errWhite.Error())
		}
		if errBlack != nil {
			if msgErr != "" {
				msgErr = msgErr + ";"
			}
			msgErr = msgErr + fmt.Sprintf("%s: %v", l.black.ID(), errBlack.Error())
		}
		return errors.New(msgErr)
	}
	return nil
}

func (l *List) checks(r xlist.Resource) bool {
	if r.IsValid() {
		return l.provides[int(r)]
	}
	return false
}
