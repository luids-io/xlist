// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package ratewr provides a wrapper for RBLs that sets a rate using golang
// standard time/rate package.
//
// This package is a work in progress and makes no API stability promises.
package ratewr

import (
	"context"
	"errors"
	"fmt"

	"github.com/luids-io/core/xlist"
)

type options struct {
	buffer int
	wait   bool
}

var defaultOptions = options{
	buffer: 1,
}

// Option is used for component configuration
type Option func(*options)

// Buffer defines the token buffer for the burst
func Buffer(i int) Option {
	return func(o *options) {
		if i > 0 {
			o.buffer = i
		}
	}
}

// Wait defines if the request must wait for a token
func Wait(b bool) Option {
	return func(o *options) {
		o.wait = b
	}
}

// Wrapper implements an xlist.Checker wrapper for include a rate checking
// requests
type Wrapper struct {
	xlist.List

	opts    options
	limiter Limiter
	list    xlist.List
}

// New creates a Wrapper with timeout
func New(limiter Limiter, list xlist.List, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Wrapper{
		opts:    opts,
		limiter: limiter,
		list:    list,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	if w.opts.wait {
		err := w.limiter.Wait(ctx)
		if err != nil {
			return xlist.Response{}, fmt.Errorf("rate limit: %v", err)
		}
	} else {
		if !w.limiter.Allow() {
			return xlist.Response{}, errors.New("rate limit")
		}
	}
	return w.list.Check(ctx, name, resource)
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	return w.list.Resources()
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	return w.list.Ping()
}

// Append implements xlist.Writer interface
func (w *Wrapper) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return w.list.Append(ctx, name, r, f)
}

// Remove implements xlist.Writer interface
func (w *Wrapper) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	return w.list.Remove(ctx, name, r, f)
}

// Clear implements xlist.Writer interface
func (w *Wrapper) Clear(ctx context.Context) error {
	return w.list.Clear(ctx)
}

// ReadOnly implements xlist.Writer interface
func (w *Wrapper) ReadOnly() (bool, error) {
	return w.list.ReadOnly()
}
