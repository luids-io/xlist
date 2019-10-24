// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package timeoutwr provides a wrapper for RBLs that sets a timeout in the
// check resquests.
//
// This package is a work in progress and makes no API stability promises.
package timeoutwr

import (
	"context"
	"time"

	"github.com/luids-io/core/xlist"
)

type options struct{}

// Option is used for component configuration
type Option func(*options)

// Wrapper implements an xlist.Checker wrapper for include a timeout in check
// requests
type Wrapper struct {
	timeout time.Duration
	checker xlist.Checker
}

// New creates a Wrapper with timeout
func New(timeout time.Duration, checker xlist.Checker, opt ...Option) *Wrapper {
	return &Wrapper{
		timeout: timeout,
		checker: checker,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	ctxChild, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	return w.checker.Check(ctxChild, name, resource)
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	return w.checker.Resources()
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	return w.checker.Ping()
}
