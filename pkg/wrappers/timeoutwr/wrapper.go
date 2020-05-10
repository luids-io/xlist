// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package timeoutwr provides a wrapper for RBLs that sets a timeout in the
// check resquests.
//
// This package is a work in progress and makes no API stability promises.
package timeoutwr

import (
	"context"
	"time"

	"github.com/luids-io/api/xlist"
)

// Wrapper implements an xlist.Checker wrapper for include a timeout in check
// requests
type Wrapper struct {
	timeout time.Duration
	list    xlist.List
}

// New creates a Wrapper with timeout
func New(list xlist.List, timeout time.Duration) *Wrapper {
	return &Wrapper{
		timeout: timeout,
		list:    list,
	}
}

// ID implements xlist.List interface
func (w *Wrapper) ID() string {
	return w.list.ID()
}

// Class implements xlist.List interface
func (w *Wrapper) Class() string {
	return BuildClass
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

// ReadOnly implements xlist.Writer interface
func (w *Wrapper) ReadOnly() bool {
	return true
}
