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
	opts    options
	limiter Limiter
	checker xlist.Checker
}

// New creates a Wrapper with timeout
func New(limiter Limiter, checker xlist.Checker, opt ...Option) *Wrapper {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Wrapper{
		opts:    opts,
		limiter: limiter,
		checker: checker,
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
	return w.checker.Check(ctx, name, resource)
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	return w.checker.Resources()
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	return w.checker.Ping()
}
