// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package wbeforexl provides a simple xlist.Checker implementation that can
// be used to check on a white list before checking on a blacklist.
//
// This package is a work in progress and makes no API stability promises.
package wbeforexl

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/core/xlist"
)

// Config options
type Config struct {
	Resources       []xlist.Resource
	ForceValidation bool
	Reason          string
}

type options struct {
	forceValidation bool
	reason          string
}

// List implements a composite RBL that checks a whtelist before checking
// the blacklist. This means that if the checked resource exists in the
// whitelist, then it returns immediately with a negative result. If not
// in the whitelist, then returns the response of the blacklist.
type List struct {
	opts         options
	provides     []bool
	resources    []xlist.Resource
	white, black xlist.Checker
}

// New constructs a new "white before" RBL, it receives the resource list that
// RBL supports.
func New(white, black xlist.Checker, cfg Config) *List {
	l := &List{
		opts: options{
			forceValidation: cfg.ForceValidation,
			reason:          cfg.Reason,
		},
		white:     white,
		black:     black,
		resources: xlist.ClearResourceDups(cfg.Resources),
		provides:  make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range l.resources {
		l.provides[int(r)] = true
	}
	return l
}

// Check implements xlist.Checker interface
func (l *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !l.checks(resource) {
		return xlist.Response{}, xlist.ErrNotImplemented
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, l.opts.forceValidation)
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
		return xlist.Response{}, ctx.Err()
	default:
		if l.black != nil {
			resp, err := l.black.Check(ctx, name, resource)
			if err == nil && resp.Result && l.opts.reason != "" {
				resp.Reason = l.opts.reason
			}
			return resp, err
		}
		return xlist.Response{}, nil
	}
}

// Resources implements xlist.Checker interface
func (l *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, len(l.resources), len(l.resources))
	copy(resources, l.resources)
	return resources
}

// Ping implements xlist.Checker interface
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
			msgErr = fmt.Sprintf("wbefore[0]: %v", errWhite.Error())
		}
		if errBlack != nil {
			if msgErr != "" {
				msgErr = msgErr + ";"
			}
			msgErr = msgErr + fmt.Sprintf("wbefore[1]: %v", errBlack.Error())
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

// Append implements xlist.Writer interface
func (l *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *List) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *List) ReadOnly() (bool, error) {
	return true, nil
}
