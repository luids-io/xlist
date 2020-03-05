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
	"strings"

	"github.com/luids-io/core/xlist"
)

type options struct {
	forceValidation bool
	reason          string
}

var defaultOptions = options{}

// Option is used for common component configuration
type Option func(*options)

// ForceValidation forces components to ignore context and validate requests
func ForceValidation(b bool) Option {
	return func(o *options) {
		o.forceValidation = b
	}
}

// Reason sets a fixed reason for component
func Reason(s string) Option {
	return func(o *options) {
		o.reason = s
	}
}

// List implements a composite RBL that checks a whtelist before checking
// the blacklist. This means that if the checked resource exists in the
// whitelist, then it returns immediately with a negative result. If not
// in the whitelist, then returns the response of the blacklist.
type List struct {
	xlist.List

	opts options
	//resource types that list provides
	provides  []bool
	whitelist xlist.Checker
	blacklist xlist.Checker
}

// New constructs a new "white before" RBL, it receives the resource list that
// RBL supports.
func New(resources []xlist.Resource, opt ...Option) *List {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	w := &List{
		opts:     opts,
		provides: make([]bool, len(xlist.Resources), len(xlist.Resources)),
	}
	//set resource types that provides
	for _, r := range resources {
		if r.IsValid() {
			w.provides[int(r)] = true
		}
	}
	return w
}

// SetWhitelist sets the whitelist
func (w *List) SetWhitelist(c xlist.Checker) {
	w.whitelist = c
}

// SetBlacklist sets the blacklist
func (w *List) SetBlacklist(c xlist.Checker) {
	w.blacklist = c
}

// Check implements xlist.Checker interface
func (w *List) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if !w.checks(resource) {
		return xlist.Response{}, xlist.ErrNotImplemented
	}
	name, ctx, err := xlist.DoValidation(ctx, name, resource, w.opts.forceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	if w.whitelist != nil {
		resp, err := w.whitelist.Check(ctx, name, resource)
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
		if w.blacklist != nil {
			resp, err := w.blacklist.Check(ctx, name, resource)
			if err == nil && resp.Result && w.opts.reason != "" {
				resp.Reason = w.opts.reason
			}
			return resp, err
		}
		return xlist.Response{}, nil
	}
}

// Resources implements xlist.Checker interface
func (w *List) Resources() []xlist.Resource {
	resources := make([]xlist.Resource, 0, len(xlist.Resources))
	for _, r := range xlist.Resources {
		if w.provides[int(r)] {
			resources = append(resources, r)
		}
	}
	return resources
}

// Ping implements xlist.Checker interface
func (w *List) Ping() error {
	var errWhite, errBlack error
	if w.whitelist != nil {
		errWhite = w.whitelist.Ping()
	}
	if w.blacklist != nil {
		errBlack = w.blacklist.Ping()
	}
	if errWhite != nil || errBlack != nil {
		msgErr := make([]string, 0, 2)
		if errWhite != nil {
			msgErr = append(msgErr, fmt.Sprintf("wbefore[0]: %v", errWhite.Error()))
		}
		if errBlack != nil {
			msgErr = append(msgErr, fmt.Sprintf("wbefore[1]: %v", errBlack.Error()))
		}
		return errors.New(strings.Join(msgErr, ";"))
	}
	return nil
}

func (w *List) checks(r xlist.Resource) bool {
	if r.IsValid() {
		return w.provides[int(r)]
	}
	return false
}

// Append implements xlist.Writer interface
func (w *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (w *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (w *List) Clear(ctx context.Context) error {
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (w *List) ReadOnly() (bool, error) {
	return true, nil
}
