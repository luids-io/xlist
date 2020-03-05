// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package mockxl provides a simple xlist.Checker implementation that can
// be used to test other components or test configurations.
//
// This package is a work in progress and makes no API stability promises.
package mockxl

import (
	"context"
	"sync"
	"time"

	"github.com/luids-io/core/xlist"
)

var defaultReason = "The resource is on the mockup list"

// List is a mockup list that implements xlist.Checker, see examples.
type List struct {
	xlist.List

	// ResourceList that this list checks
	ResourceList []xlist.Resource
	// Results is the sequence of results that the list returns on checks
	Results []bool
	// Fail setup a list that fails
	Fail bool
	// Lazy sets a delay on checks, having a cancellation
	Lazy time.Duration
	// Sleep sets a delay on checks, having NO cancellation
	Sleep time.Duration
	// ForceValidation forces validation in each check
	ForceValidation bool
	// Reason changes the default reason
	Reason string
	// TTL sets a ttl value on checks
	TTL int

	mu   sync.Mutex
	next int
}

// Check implements xlist.Checker
func (l *List) Check(ctx context.Context, name string, res xlist.Resource) (xlist.Response, error) {
	if l.Fail {
		return xlist.Response{}, xlist.ErrNotAvailable
	}
	name, ctx, err := xlist.DoValidation(ctx, name, res, l.ForceValidation)
	if err != nil {
		return xlist.Response{}, err
	}
	if !res.InArray(l.ResourceList) {
		return xlist.Response{}, xlist.ErrNotImplemented
	}
	if l.Lazy > 0 {
		select {
		case <-time.After(l.Lazy):
			break
		case <-ctx.Done():
			return xlist.Response{}, ctx.Err()
		}
	} else if l.Sleep > 0 {
		time.Sleep(l.Sleep)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	resp := xlist.Response{}
	if len(l.Results) > 0 {
		result := l.Results[l.next]
		if result {
			resp.Result = true
			resp.Reason = defaultReason
			if l.Reason != "" {
				resp.Reason = l.Reason
			}
		}
		l.next++
		if l.next >= len(l.Results) {
			l.next = 0
		}
	}
	if l.TTL > 0 {
		resp.TTL = l.TTL
	}
	return resp, nil
}

// Ping implements xlist.Checker
func (l *List) Ping() error {
	if l.Fail {
		return xlist.ErrNotAvailable
	}
	return nil
}

// Resources implements xlist.Checker
func (l *List) Resources() []xlist.Resource {
	ret := make([]xlist.Resource, len(l.ResourceList), len(l.ResourceList))
	copy(ret, l.ResourceList)
	return ret
}

// Append implements xlist.Writer interface
func (l *List) Append(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	if l.Fail {
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// Remove implements xlist.Writer interface
func (l *List) Remove(ctx context.Context, name string, r xlist.Resource, f xlist.Format) error {
	if l.Fail {
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// Clear implements xlist.Writer interface
func (l *List) Clear(ctx context.Context) error {
	if l.Fail {
		return xlist.ErrNotAvailable
	}
	return xlist.ErrReadOnlyMode
}

// ReadOnly implements xlist.Writer interface
func (l *List) ReadOnly() (bool, error) {
	if l.Fail {
		return false, xlist.ErrNotAvailable
	}
	return true, nil
}
