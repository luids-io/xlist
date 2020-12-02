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
	"github.com/luids-io/xlist/pkg/xlistd"
)

// WrapperClass registered.
const WrapperClass = "timeout"

// Wrapper implements an xlistd.List wrapper for include a timeout in check
// requests.
type Wrapper struct {
	timeout time.Duration
	list    xlistd.List
}

// New creates a Wrapper with timeout.
func New(list xlistd.List, timeout time.Duration) *Wrapper {
	return &Wrapper{
		timeout: timeout,
		list:    list,
	}
}

// ID implements xlistd.List interface.
func (w *Wrapper) ID() string {
	return w.list.ID()
}

// Class implements xlistd.List interface.
func (w *Wrapper) Class() string {
	return w.list.Class()
}

// Check implements xlist.Checker interface.
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	ctxChild, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	return w.list.Check(ctxChild, name, resource)
}

// Resources implements xlist.Checker interface.
func (w *Wrapper) Resources(ctx context.Context) ([]xlist.Resource, error) {
	return w.list.Resources(ctx)
}

// Ping implements xlistd.List interface.
func (w *Wrapper) Ping() error {
	return w.list.Ping()
}
