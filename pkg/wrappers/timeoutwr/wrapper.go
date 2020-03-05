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
	xlist.List

	timeout time.Duration
	list    xlist.List
}

// New creates a Wrapper with timeout
func New(timeout time.Duration, list xlist.List, opt ...Option) *Wrapper {
	return &Wrapper{
		timeout: timeout,
		list:    list,
	}
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	ctxChild, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	return w.list.Check(ctxChild, name, resource)
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
